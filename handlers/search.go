package handlers

import (
	"encoding/hex"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

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
	search = strings.Replace(search, "0x", "", -1)
	search = strings.Replace(search, "0X", "", -1)
	var err error
	logger := logger.WithField("searchType", searchType)
	var result interface{}

	switch searchType {
	case "slots":
		if len(search) <= 1 {
			break
		}
		result = &types.SearchAheadSlotsResult{}
		if searchLikeRE.MatchString(search) {
			if _, convertErr := strconv.Atoi(search); convertErr == nil {
				err = db.ReaderDb.Select(result, `
				SELECT slot, ENCODE(blockroot, 'hex') AS blockroot 
				FROM blocks 
				WHERE slot = $1
				ORDER BY slot LIMIT 10`, search)
			} else if len(search) == 64 {
				blockHash, err := hex.DecodeString(search)
				if err != nil {
					logger.Errorf("error parsing blockHash to int: %v", err)
					http.Error(w, "Internal server error", http.StatusServiceUnavailable)
					return
				}
				err = db.ReaderDb.Select(result, `
				SELECT slot, ENCODE(blockroot, 'hex') AS blockroot 
				FROM blocks 
				WHERE blockroot = $1 OR
					stateroot = $1
				ORDER BY slot LIMIT 10`, blockHash)
				if err != nil {
					logger.Errorf("error reading block root: %v", err)
					http.Error(w, "Internal server error", http.StatusServiceUnavailable)
					return
				}
			}
		}
	case "blocks":
		number, err := strconv.ParseUint(search, 10, 64)
		if err == nil {
			block, err := db.BigtableClient.GetBlockFromBlocksTable(number)
			if err == nil {
				result = &types.SearchAheadBlocksResult{{
					Block: block.Number,
					Hash:  fmt.Sprintf("%#x", block.Hash),
				}}
			}
		}
	case "graffiti":
		graffiti := &types.SearchAheadGraffitiResult{}
		err = db.ReaderDb.Select(graffiti, `
			SELECT graffiti, count(*)
			FROM blocks
			WHERE graffiti_text ILIKE LOWER($1)
			GROUP BY graffiti
			ORDER BY count desc
			LIMIT 10`, "%"+search+"%")
		if err == nil {
			for i := range *graffiti {
				(*graffiti)[i].Graffiti = utils.FormatGraffitiString((*graffiti)[i].Graffiti)
			}
		}
		result = graffiti
	case "transactions":
		result = &types.SearchAheadTransactionsResult{}
		if transactionLikeRE.MatchString(strings.ToLower(strings.Replace(search, "0x", "", -1))) {
			txHash, txHashErr := hex.DecodeString(strings.ToLower(strings.Replace(search, "0x", "", -1)))
			if txHashErr != nil {
				logger.Errorf("error parsing txHash %v: %v", search, txHashErr)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			var tx *types.Eth1TransactionIndexed
			tx, err = db.BigtableClient.GetIndexedEth1Transaction(txHash)
			if err == nil && tx != nil {
				result = &types.SearchAheadTransactionsResult{{TxHash: fmt.Sprintf("%x", tx.Hash)}}
			}
		}
	case "epochs":
		result = &types.SearchAheadEpochsResult{}
		err = db.ReaderDb.Select(result, "SELECT epoch FROM epochs WHERE CAST(epoch AS text) LIKE $1 ORDER BY epoch LIMIT 10", search+"%")
	case "validators":
		// find all validators that have a index, publickey or name like the search-query
		result = &types.SearchAheadValidatorsResult{}
		indexNumeric, errParse := strconv.ParseInt(search, 10, 64)
		if errParse == nil { // search the validator by its index
			err = db.ReaderDb.Select(result, `SELECT validatorindex AS index, pubkeyhex as pubkey FROM validators WHERE validatorindex = $1`, indexNumeric)
			if err != nil {
				logger.Errorf("error reading result data: %v", err)
				http.Error(w, "Internal server error", http.StatusServiceUnavailable)
				return
			}
		} else if thresholdHexLikeRE.MatchString(search) {
			err = db.ReaderDb.Select(result, `SELECT validatorindex AS index, pubkeyhex as pubkey FROM validators WHERE pubkeyhex LIKE LOWER($1 || '%')`, search)
			if err != nil {
				logger.Errorf("error reading result data: %v", err)
				http.Error(w, "Internal server error", http.StatusServiceUnavailable)
				return
			}
		} else {
			err = db.ReaderDb.Select(result, `
			SELECT validatorindex AS index, pubkeyhex AS pubkey
			FROM validators
			LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
			WHERE LOWER(validator_names.name) LIKE LOWER($1)
			ORDER BY index LIMIT 10`, search+"%")
			if err != nil {
				logger.Errorf("error reading result data: %v", err)
				http.Error(w, "Internal server error", http.StatusServiceUnavailable)
				return
			}
		}
	case "eth1_addresses":

		var ensData *types.EnsDomainResponse
		if utils.IsValidEnsDomain(search) {
			ensData, _ = GetEnsDomain(search)
			if len(ensData.Address) > 0 {
				search = strings.Replace(ensData.Address, "0x", "", -1)
			}
		}
		// start := time.Now()
		if len(search) <= 1 {
			break
		}
		if len(search)%2 != 0 {
			search = search[:len(search)-1]
		}
		if searchLikeRE.MatchString(search) {
			eth1AddressHash, err := hex.DecodeString(search)
			if err != nil {
				logger.Errorf("error parsing eth1AddressHash to hash: %v", err)
				http.Error(w, "Internal server error", http.StatusServiceUnavailable)
				return
			}
			result, err = db.BigtableClient.SearchForAddress(eth1AddressHash, 10)
			if err != nil {
				logger.Errorf("error searching for eth1AddressHash: %v", err)
				http.Error(w, "Internal server error", http.StatusServiceUnavailable)
				return
			} else if ensData != nil && len(ensData.Domain) > 0 {
				cast, ok := result.([]*types.Eth1AddressSearchItem)
				if ok && len(cast) > 0 {
					cast[0].Name = ensData.Domain
				}
			}
		} else {
			result = []*types.Eth1AddressSearchItem{}
		}
		// logger.WithFields(logrus.Fields{
		// 	"duration": time.Since(start),
		// }).Infof("finished searching for eth1_addresses")
	case "indexed_validators":
		// find all validators that have a publickey or index like the search-query
		result = &types.SearchAheadValidatorsResult{}
		indexNumeric, errParse := strconv.ParseInt(search, 10, 64)
		if errParse == nil { // search the validator by its index
			err = db.ReaderDb.Select(result, `SELECT validatorindex AS index, pubkeyhex as pubkey FROM validators WHERE validatorindex = $1`, indexNumeric)
			if err != nil {
				logger.Errorf("error reading result data: %v", err)
				http.Error(w, "Internal server error", http.StatusServiceUnavailable)
				return
			}
		} else if thresholdHexLikeRE.MatchString(search) {
			err = db.ReaderDb.Select(result, `SELECT validatorindex AS index, pubkeyhex as pubkey FROM validators WHERE pubkeyhex LIKE LOWER($1 || '%')`, search)
			if err != nil {
				logger.Errorf("error reading result data: %v", err)
				http.Error(w, "Internal server error", http.StatusServiceUnavailable)
				return
			}
		} else {
			err = db.ReaderDb.Select(result, `
			SELECT validatorindex AS index, pubkeyhex AS pubkey
			FROM validators
			LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
			WHERE LOWER(validator_names.name) LIKE LOWER($1)
			ORDER BY index LIMIT 10`, search+"%")
			if err != nil {
				logger.Errorf("error reading result data: %v", err)
				http.Error(w, "Internal server error", http.StatusServiceUnavailable)
				return
			}
		}
	case "validators_by_pubkey":
		result = &types.SearchAheadPubkeyResult{}
		if thresholdHexLikeRE.MatchString(search) {
			// Find the validators that have made a deposit but have no index yet and therefore are not in the validators table
			err = db.ReaderDb.Select(result, `
				SELECT DISTINCT
					ENCODE(eth1_deposits.publickey, 'hex') AS pubkey
					FROM eth1_deposits
					LEFT JOIN validators ON validators.pubkey = eth1_deposits.publickey
					WHERE validators.pubkey IS NULL AND ENCODE(eth1_deposits.publickey, 'hex') LIKE ($1 || '%')`, search)
			if err != nil {
				logger.Errorf("error reading result data: %v", err)
				http.Error(w, "Internal server error", http.StatusServiceUnavailable)
				return
			}
		}
	case "indexed_validators_by_eth1_addresses":
		result, err = FindValidatorIndicesByEth1Address(search)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		}
	case "count_indexed_validators_by_eth1_address":
		if len(search) <= 1 {
			break
		}
		var ensData *types.EnsDomainResponse
		if utils.IsValidEnsDomain(search) {
			ensData, _ = GetEnsDomain(search)
			if len(ensData.Address) > 0 {
				search = strings.Replace(ensData.Address, "0x", "", -1)
			}
		}
		if len(search)%2 != 0 {
			search = search[:len(search)-1]
		}
		if searchLikeRE.MatchString(search) {
			// find validators per eth1-address (limit result by N addresses and M validators per address)
			result = &[]struct {
				Eth1Address string `db:"from_address" json:"eth1_address"`
				Count       uint64 `db:"count" json:"count"`
			}{}
			eth1AddressHash, err := hex.DecodeString(search)
			if err != nil {
				logger.Errorf("error parsing eth1AddressHash to hex: %v", err)
				http.Error(w, "Internal server error", http.StatusServiceUnavailable)
				return
			}
			err = db.ReaderDb.Select(result, `
			SELECT from_address, COUNT(*) FROM (
				SELECT 
					DISTINCT ON(validatorindex) validatorindex,
					ENCODE(from_address::bytea, 'hex') as from_address
				FROM eth1_deposits
				INNER JOIN validators ON validators.pubkey = eth1_deposits.publickey
				WHERE from_address LIKE $1 || '%'::bytea
			) a 
			GROUP BY from_address`, eth1AddressHash)
			if err != nil {
				logger.Errorf("error retrieving count of indexed validators by address data: %v", err)
				http.Error(w, "Internal server error", http.StatusServiceUnavailable)
				return
			}
		}
	case "indexed_validators_by_graffiti":
		// find validators per graffiti (limit result by N graffities and M validators per graffiti)
		res := []struct {
			Graffiti         string        `db:"graffiti" json:"graffiti"`
			ValidatorIndices pq.Int64Array `db:"validatorindices" json:"validator_indices"`
			Count            uint64        `db:"count" json:"-"`
		}{}
		err = db.ReaderDb.Select(&res, `
			SELECT graffiti, COUNT(*), ARRAY_AGG(validatorindex) validatorindices FROM (
				SELECT 
					DISTINCT ON(validatorindex) validatorindex,
					graffiti,
					DENSE_RANK() OVER(PARTITION BY graffiti ORDER BY validatorindex) AS validatorrow,
					DENSE_RANK() OVER(ORDER BY graffiti) AS graffitirow
				FROM blocks 
				LEFT JOIN validators ON blocks.proposer = validators.validatorindex
				WHERE graffiti_text ILIKE LOWER($1)
			) a 
			WHERE validatorrow <= $2 AND graffitirow <= 10
			GROUP BY graffiti
			ORDER BY count DESC`, "%"+search+"%", searchValidatorsResultLimit)
		if err == nil {
			for i := range res {
				res[i].Graffiti = utils.FormatGraffitiString(res[i].Graffiti)
			}
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
		if err == nil {
			for i := range res {
				res[i].Name = string(utils.FormatValidatorName(res[i].Name))
			}
		}
		result = &res
	case "ens":
		data, err := GetEnsDomain(search)
		if err == nil {
			result = &data
		}

	default:
		http.Error(w, "Not found", 404)
		return
	}

	if err != nil {
		logger.WithError(err).WithField("searchType", searchType).Error("error doing query for searchAhead")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		logger.WithError(err).Error("error encoding searchAhead")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
	}
}

// search can ether be a valid ETH address or an ENS name mapping to one
func FindValidatorIndicesByEth1Address(search string) (types.SearchValidatorsByEth1Result, error) {
	result := &[]struct {
		Eth1Address      string        `db:"from_address" json:"eth1_address"`
		ValidatorIndices pq.Int64Array `db:"validatorindices" json:"validator_indices"`
		Count            uint64        `db:"count" json:"-"`
	}{}

	search = strings.Replace(ReplaceEnsNameWithAddress(search), "0x", "", -1)
	if len(search)%2 != 0 {
		search = search[:len(search)-1]
	}
	if len(search) <= 1 || len(search) > 40 {
		return nil, fmt.Errorf("not a valid Eth1 address: %v", search)
	}
	// find validators per eth1-address (limit result by N addresses and M validators per address)

	eth1AddressHash, err := hex.DecodeString(search)
	if err != nil {
		return nil, fmt.Errorf("error parsing eth1AddressHash to hex: %v", err)
	}
	if len(eth1AddressHash) == 20 {
		// if it is an eth1-address just search for exact match
		err = db.ReaderDb.Select(result, `
			SELECT from_address, COUNT(*), ARRAY_AGG(validatorindex) validatorindices FROM (
				SELECT 
					DISTINCT ON(validatorindex) validatorindex,
					ENCODE(from_address::bytea, 'hex') as from_address,
					DENSE_RANK() OVER (PARTITION BY from_address ORDER BY validatorindex) AS validatorrow,
					DENSE_RANK() OVER (ORDER BY from_address) AS addressrow
				FROM eth1_deposits
				INNER JOIN validators ON validators.pubkey = eth1_deposits.publickey
				WHERE from_address = $1
			) a 
			WHERE validatorrow <= $2 AND addressrow <= 10
			GROUP BY from_address
			ORDER BY count DESC`, eth1AddressHash, searchValidatorsResultLimit)
	} else if len(eth1AddressHash) <= 16 {
		// if the lenght is less then 32 use byte-wise comparison
		err = db.ReaderDb.Select(result, `
			SELECT from_address, COUNT(*), ARRAY_AGG(validatorindex) validatorindices FROM (
				SELECT 
					DISTINCT ON(validatorindex) validatorindex,
					ENCODE(from_address::bytea, 'hex') as from_address,
					DENSE_RANK() OVER (PARTITION BY from_address ORDER BY validatorindex) AS validatorrow,
					DENSE_RANK() OVER (ORDER BY from_address) AS addressrow
				FROM eth1_deposits
				INNER JOIN validators ON validators.pubkey = eth1_deposits.publickey
				WHERE from_address LIKE $1 || '%'::bytea
			) a 
			WHERE validatorrow <= $2 AND addressrow <= 10
			GROUP BY from_address
			ORDER BY count DESC`, eth1AddressHash, searchValidatorsResultLimit)
	} else {
		// otherwise use hex (see BIDS-1570)
		err = db.ReaderDb.Select(result, `
			SELECT from_address, COUNT(*), ARRAY_AGG(validatorindex) validatorindices FROM (
				SELECT 
					DISTINCT ON(validatorindex) validatorindex,
					ENCODE(from_address::bytea, 'hex') as from_address,
					DENSE_RANK() OVER (PARTITION BY from_address ORDER BY validatorindex) AS validatorrow,
					DENSE_RANK() OVER (ORDER BY from_address) AS addressrow
				FROM eth1_deposits
				INNER JOIN validators ON validators.pubkey = eth1_deposits.publickey
				WHERE encode(from_address,'hex') LIKE $1
			) a 
			WHERE validatorrow <= $2 AND addressrow <= 10
			GROUP BY from_address
			ORDER BY count DESC`, search+"%", searchValidatorsResultLimit)
	}
	if err != nil {
		utils.LogError(err, "error getting validators for eth1 address from db", 0)
		return nil, fmt.Errorf("error reading result data: %v", err)
	}
	return *result, nil
}
