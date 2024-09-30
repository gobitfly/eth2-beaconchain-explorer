package handlers

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/templates"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

const searchValidatorsResultLimit = 300

var transactionLikeRE = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)
var searchLikeRE = regexp.MustCompile(`^[0-9a-fA-F]{0,96}$`)
var thresholdHexLikeRE = regexp.MustCompile(`^[0-9a-fA-F]{5,96}$`)

// Search handles search requests
func Search(w http.ResponseWriter, r *http.Request) {

	search := r.FormValue("search")

	_, err := strconv.Atoi(search)

	if err == nil {
		http.Redirect(w, r, "/block/"+search, http.StatusMovedPermanently)
		return
	}

	var ensData *types.EnsDomainResponse
	if utils.IsValidEnsDomain(search) {
		ensData, _ = GetEnsDomain(search)
	}
	search = strings.Replace(search, "0x", "", -1)
	if ensData != nil && len(ensData.Address) > 0 {
		http.Redirect(w, r, "/address/"+ensData.Domain, http.StatusMovedPermanently)
	} else if utils.IsValidWithdrawalCredentials(search) {
		http.Redirect(w, r, "/validators/deposits?q="+search, http.StatusMovedPermanently)
	} else if utils.IsValidEth1Tx(search) {
		http.Redirect(w, r, "/tx/"+search, http.StatusMovedPermanently)
	} else if len(search) == 96 {
		http.Redirect(w, r, "/validator/"+search, http.StatusMovedPermanently)
	} else if utils.IsValidEth1Address(search) {
		http.Redirect(w, r, "/address/"+search, http.StatusMovedPermanently)
	} else {
		w.Header().Set("Content-Type", "text/html")
		templateFiles := append(layoutTemplateFiles, "searchnotfound.html")
		var searchNotFoundTemplate = templates.GetTemplate(templateFiles...)
		data := InitPageData(w, r, "search", "/search", "", templateFiles)

		if handleTemplateError(w, r, "search.go", "Search", "", searchNotFoundTemplate.ExecuteTemplate(w, "layout", data)) != nil {
			return // an error has occurred and was processed
		}
	}
}

// SearchAhead handles responses for the frontend search boxes
func SearchAhead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	searchType := vars["type"]
	search := vars["search"]
	var err error
	logger := logger.WithField("searchType", searchType)
	var result interface{}

	strippedSearch := strings.Replace(search, "0x", "", -1)
	lowerStrippedSearch := strings.ToLower(strippedSearch)

	if len(strippedSearch) == 0 {
		_ = json.NewEncoder(w).Encode(result)
		return
	}

	switch searchType {
	case "slots":
		if !searchLikeRE.MatchString(strippedSearch) {
			break
		}
		result = &types.SearchAheadSlotsResult{}
		if _, parseErr := strconv.ParseInt(search, 10, 32); parseErr == nil {
			err = db.ReaderDb.Select(result, `
			SELECT slot, ENCODE(blockroot, 'hex') AS blockroot 
			FROM blocks 
			WHERE slot = $1
			ORDER BY slot LIMIT 10`, search)
		} else if len(strippedSearch) == 64 {
			var blockHash []byte
			blockHash, err = hex.DecodeString(strippedSearch)
			if err != nil {
				err = fmt.Errorf("error parsing blockHash to int: %v", err)
				break
			}
			err = db.ReaderDb.Select(result, `
			SELECT slot, ENCODE(blockroot, 'hex') AS blockroot 
			FROM blocks 
			WHERE blockroot = $1 OR
				stateroot = $1
			ORDER BY slot LIMIT 10`, blockHash)
		}
	case "blocks":
		number, parseErr := strconv.ParseUint(search, 10, 64)
		if parseErr != nil {
			break
		}
		block, blockErr := db.BigtableClient.GetBlockFromBlocksTable(number)
		if blockErr != nil {
			if blockErr != db.ErrBlockNotFound {
				err = blockErr
			}
			break
		}
		result = &types.SearchAheadBlocksResult{{
			Block: fmt.Sprintf("%v", block.Number),
			Hash:  fmt.Sprintf("%#x", block.Hash),
		}}
	case "graffiti":
		graffiti := &types.SearchAheadGraffitiResult{}
		err = db.ReaderDb.Select(graffiti, `
			select graffiti, sum(count) as count 
			from graffiti_stats 
			where graffiti_text ilike lower($1) 
			group by graffiti 
			order by count desc 
			limit 10`, "%"+search+"%")
		if err != nil {
			break
		}
		for i := range *graffiti {
			(*graffiti)[i].Graffiti = utils.FormatGraffitiString((*graffiti)[i].Graffiti)
		}
		result = graffiti
	case "transactions":
		if !transactionLikeRE.MatchString(lowerStrippedSearch) {
			break
		}
		result = &types.SearchAheadTransactionsResult{}
		var txHash []byte
		txHash, err = hex.DecodeString(lowerStrippedSearch)
		if err != nil {
			err = fmt.Errorf("error parsing txHash %v: %v", search, err)
			break
		}
		var tx *types.Eth1TransactionIndexed
		tx, err = db.BigtableClient.GetIndexedEth1Transaction(txHash)
		if err != nil || tx == nil {
			break
		}
		result = &types.SearchAheadTransactionsResult{{TxHash: fmt.Sprintf("%x", tx.Hash)}}
	case "epochs":
		epochNumber, parseErr := strconv.ParseUint(search, 10, 32)
		if parseErr != nil {
			break
		}
		result = &types.SearchAheadEpochsResult{}
		err = db.ReaderDb.Select(result, "SELECT epoch FROM epochs WHERE epoch = $1 ORDER BY epoch LIMIT 10", epochNumber)
	case "validators":
		// find all validators that have a index, publickey or name like the search-query
		result = &types.SearchAheadValidatorsResult{}
		indexNumeric, parseErr := strconv.ParseInt(search, 10, 32)
		if parseErr == nil { // search the validator by its index
			err = db.ReaderDb.Select(result, `SELECT validatorindex AS index, pubkeyhex as pubkey FROM validators WHERE validatorindex = $1`, indexNumeric)
		} else if thresholdHexLikeRE.MatchString(lowerStrippedSearch) {
			err = db.ReaderDb.Select(result, `SELECT validatorindex AS index, pubkeyhex as pubkey FROM validators WHERE pubkeyhex LIKE ($1 || '%')`, lowerStrippedSearch)
		} else {
			err = db.ReaderDb.Select(result, `
			SELECT validatorindex AS index, pubkeyhex AS pubkey
			FROM validators WHERE pubkey IN 
				(SELECT publickey FROM validator_names WHERE LOWER(validator_names.name) LIKE LOWER($1) LIMIT 10)`, search+"%")
		}
	case "eth1_addresses":
		if utils.IsValidEnsDomain(search) {
			ensData, _ := GetEnsDomain(search)
			if len(ensData.Address) > 0 {
				result = []*types.Eth1AddressSearchItem{{
					Address: ensData.Address,
					Name:    ensData.Domain,
					Token:   "",
				}}
				break
			}
		}
		if !searchLikeRE.MatchString(strippedSearch) {
			break
		}
		if len(strippedSearch)%2 != 0 { // pad with 0 if uneven
			strippedSearch = strippedSearch + "0"
		}
		eth1AddressHash, decodeErr := hex.DecodeString(strippedSearch)
		if decodeErr != nil {
			break
		}
		result, err = db.BigtableClient.SearchForAddress(eth1AddressHash, 10)
		if err != nil {
			err = fmt.Errorf("error searching for eth1AddressHash: %v", err)
		}
	case "indexed_validators":
		// find all validators that have a publickey or index like the search-query
		result = &types.SearchAheadValidatorsResult{}
		indexNumeric, errParse := strconv.ParseInt(search, 10, 32)
		if errParse == nil { // search the validator by its index
			err = db.ReaderDb.Select(result, `SELECT validatorindex AS index, pubkeyhex as pubkey FROM validators WHERE validatorindex = $1`, indexNumeric)
		} else if thresholdHexLikeRE.MatchString(lowerStrippedSearch) {
			err = db.ReaderDb.Select(result, `SELECT validatorindex AS index, pubkeyhex as pubkey FROM validators WHERE pubkeyhex LIKE ($1 || '%')`, lowerStrippedSearch)
		} else {
			err = db.ReaderDb.Select(result, `SELECT validatorindex AS index, pubkeyhex AS pubkey
			FROM validators WHERE pubkey IN 
				(SELECT publickey FROM validator_names WHERE LOWER(validator_names.name) LIKE LOWER($1) LIMIT 10)`, search+"%")
		}
	case "validators_by_pubkey":
		if !thresholdHexLikeRE.MatchString(lowerStrippedSearch) {
			break
		}
		result = &types.SearchAheadPubkeyResult{}
		// Find the validators that have made a deposit but have no index yet and therefore are not in the validators table
		err = db.ReaderDb.Select(result, `
		SELECT DISTINCT
			ENCODE(eth1_deposits.publickey, 'hex') AS pubkey
			FROM eth1_deposits
			LEFT JOIN validators ON validators.pubkey = eth1_deposits.publickey
			WHERE validators.pubkey IS NULL AND ENCODE(eth1_deposits.publickey, 'hex') LIKE ($1 || '%')`, lowerStrippedSearch)
	case "indexed_validators_by_eth1_addresses":
		search = ReplaceEnsNameWithAddress(search)
		if !utils.IsEth1Address(search) {
			break
		}
		result, err = FindValidatorIndicesByEth1Address(strings.ToLower(search))
	case "count_indexed_validators_by_eth1_address":
		var ensData *types.EnsDomainResponse
		if utils.IsValidEnsDomain(search) {
			ensData, _ = GetEnsDomain(search)
			if len(ensData.Address) > 0 {
				lowerStrippedSearch = strings.ToLower(strings.Replace(ensData.Address, "0x", "", -1))
			}
		}
		if !searchLikeRE.MatchString(lowerStrippedSearch) {
			break
		}
		// find validators per eth1-address
		result = &[]struct {
			Eth1Address string `db:"from_address_text" json:"eth1_address"`
			Count       uint64 `db:"count" json:"count"`
		}{}

		err = db.ReaderDb.Select(result, `
			SELECT from_address_text, COUNT(*) FROM (
				SELECT 
					DISTINCT ON(validatorindex) validatorindex,					
					from_address_text
				FROM eth1_deposits
				INNER JOIN validators ON validators.pubkey = eth1_deposits.publickey
				WHERE from_address_text LIKE $1 || '%'
			) a 
			GROUP BY from_address_text`, lowerStrippedSearch)
	case "count_indexed_validators_by_withdrawal_credential":
		var ensData *types.EnsDomainResponse
		if utils.IsValidEnsDomain(search) {
			ensData, _ = GetEnsDomain(search)
			if len(ensData.Address) > 0 {
				lowerStrippedSearch = strings.ToLower(strings.Replace(ensData.Address, "0x", "", -1))
			}
		}
		if len(lowerStrippedSearch) == 40 {
			// when the user gives an address (that validators might withdraw to) we transform the address into credentials
			lowerStrippedSearch = utils.BeginningOfSetWithdrawalCredentials + lowerStrippedSearch
		}
		if !utils.IsValidWithdrawalCredentials(lowerStrippedSearch) {
			break
		}
		decodedCredential, decodeErr := hex.DecodeString(lowerStrippedSearch)
		if decodeErr != nil {
			break
		}
		// find validators per withdrawal credential
		dbFinding := []struct {
			DecodedCredential []byte `db:"withdrawalcredentials"`
			Count             uint64 `db:"count"`
		}{}
		err = db.ReaderDb.Select(&dbFinding, `
			SELECT withdrawalcredentials, COUNT(*) FROM validators
			WHERE withdrawalcredentials = $1
			GROUP BY withdrawalcredentials`, decodedCredential)
		if err == nil {
			res := make([]struct {
				EncodedCredential string `json:"withdrawalcredentials"`
				Count             uint64 `json:"count"`
			},
				len(dbFinding))
			for i := range dbFinding {
				res[i].EncodedCredential = fmt.Sprintf("%x", dbFinding[i].DecodedCredential)
				res[i].Count = dbFinding[i].Count
			}
			result = &res
		}
	case "indexed_validators_by_graffiti":
		// find validators per graffiti (limit result by N graffities and M validators per graffiti)
		res := []struct {
			Graffiti         string        `db:"graffiti" json:"graffiti"`
			ValidatorIndices pq.Int64Array `db:"validatorindices" json:"validator_indices"`
			Count            uint64        `db:"count" json:"-"`
		}{}
		err = db.ReaderDb.Select(&res, `
			WITH 
				graffiti_days AS (SELECT day FROM (SELECT day FROM graffiti_stats WHERE graffiti_text ILIKE LOWER($1) ORDER BY DAY DESC LIMIT $2) a GROUP BY day),
				graffiti_blocks AS (SELECT proposer, graffiti FROM blocks INNER JOIN graffiti_days ON blocks.epoch >= graffiti_days.day*$3 AND blocks.epoch < graffiti_days.day*$3+$3 WHERE blocks.graffiti_text ILIKE LOWER($1) ORDER BY slot DESC LIMIT $2)
			SELECT graffiti, COUNT(*), ARRAY_AGG(validatorindex) validatorindices FROM (
				SELECT 
					DISTINCT ON(validatorindex) validatorindex,
					graffiti,
					DENSE_RANK() OVER(PARTITION BY graffiti ORDER BY validatorindex) AS validatorrow,
					DENSE_RANK() OVER(ORDER BY graffiti) AS graffitirow
				FROM graffiti_blocks 
				LEFT JOIN validators ON graffiti_blocks.proposer = validators.validatorindex
			) a 
			WHERE validatorrow <= $2 AND graffitirow <= 10
			GROUP BY graffiti
			ORDER BY count DESC`, "%"+search+"%", searchValidatorsResultLimit, utils.EpochsPerDay())
		if err != nil {
			break
		}
		for i := range res {
			res[i].Graffiti = utils.FormatGraffitiString(res[i].Graffiti)
		}
		result = &res
	case "indexed_validators_by_name":
		// find validators per name (limit result by N names and N validators per name)
		res := []struct {
			Name             string        `db:"name" json:"name"`
			ValidatorIndices pq.Int64Array `db:"validatorindices" json:"validator_indices"`
			Count            uint64        `db:"count" json:"-"`
		}{}
		err = db.ReaderDb.Select(&res, `
			SELECT name, COUNT(*), ARRAY_AGG(validatorindex) validatorindices FROM (
				SELECT
					validatorindex,
					validator_names.name,
					DENSE_RANK() OVER(PARTITION BY validator_names.name ORDER BY validatorindex) AS validatorrow,
					DENSE_RANK() OVER(PARTITION BY validator_names.name) AS namerow
				FROM validators
				LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
				WHERE LOWER(validator_names.name) LIKE LOWER($1)
			) a
			WHERE validatorrow <= $2 AND namerow <= 10
			GROUP BY name
			ORDER BY count DESC, name DESC`, "%"+search+"%", searchValidatorsResultLimit)
		if err != nil {
			break
		}
		for i := range res {
			res[i].Name = string(utils.FormatValidatorName(res[i].Name))
		}
		result = &res
	case "ens":
		if !utils.IsValidEnsDomain(search) {
			break
		}
		data, ensErr := GetEnsDomain(search)
		if ensErr != nil {
			if ensErr != sql.ErrNoRows {
				err = ensErr
			}
			break
		}
		result = &data
	default:
		http.Error(w, "Not found", 404)
		return
	}

	if err != nil {
		logger.WithError(err).WithField("searchType", searchType).Error("error doing query for searchAhead")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		logger.WithError(err).Error("error encoding searchAhead")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// search can either be a valid ETH address or an ENS name mapping to one
func FindValidatorIndicesByEth1Address(search string) (types.SearchValidatorsByEth1Result, error) {
	search = strings.ToLower(strings.Replace(ReplaceEnsNameWithAddress(search), "0x", "", -1))
	if !utils.IsValidEth1Address(search) {
		return nil, fmt.Errorf("not a valid Eth1 address: %v", search)
	}
	// find validators per eth1-address (limit result by N addresses and M validators per address)

	result := &[]struct {
		Eth1Address      string        `db:"from_address_text" json:"eth1_address"`
		ValidatorIndices pq.Int64Array `db:"validatorindices" json:"validator_indices"`
		Count            uint64        `db:"count" json:"-"`
	}{}
	// just search for exact match (substring matches turned out to be too heavy for the db!)
	err := db.ReaderDb.Select(result, `
		SELECT from_address_text, COUNT(*), ARRAY_AGG(validatorindex) validatorindices FROM (
			SELECT 
				DISTINCT ON(validatorindex) validatorindex,
				from_address_text,
				DENSE_RANK() OVER (PARTITION BY from_address_text ORDER BY validatorindex) AS validatorrow,
				DENSE_RANK() OVER (ORDER BY from_address_text) AS addressrow
			FROM eth1_deposits
			INNER JOIN validators ON validators.pubkey = eth1_deposits.publickey
			WHERE from_address_text = $1
		) a 
		WHERE validatorrow <= $2 AND addressrow <= 10
		GROUP BY from_address_text
		ORDER BY count DESC`, search, searchValidatorsResultLimit)
	if err != nil {
		utils.LogError(err, "error getting validators for eth1 address from db", 0)
		return nil, fmt.Errorf("error reading result data: %v", err)
	}
	return *result, nil
}
