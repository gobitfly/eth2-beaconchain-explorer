package handlers

import (
	"encoding/hex"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"html/template"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

const searchValidatorsResultLimit = 300

var searchNotFoundTemplate = template.Must(template.New("searchnotfound").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/searchnotfound.html"))

var searchLikeRE = regexp.MustCompile(`^[0-9a-fA-F]{0,96}$`)
var thresholdHexLikeRE = regexp.MustCompile(`^[0-9a-fA-F]{5,96}$`)

// Search handles search requests
func Search(w http.ResponseWriter, r *http.Request) {

	search := r.FormValue("search")

	_, err := strconv.Atoi(search)

	if err == nil {
		http.Redirect(w, r, "/block/"+search, 301)
		return
	}

	search = strings.Replace(search, "0x", "", -1)

	if len(search) == 64 {
		http.Redirect(w, r, "/block/"+search, 301)
	} else if len(search) == 96 {
		http.Redirect(w, r, "/validator/"+search, 301)
	} else if utils.IsValidEth1Address(search) {
		http.Redirect(w, r, "/validators/eth1deposits?q="+search, 301)
	} else {
		w.Header().Set("Content-Type", "text/html")
		data := InitPageData(w, r, "search", "/search", "")
		data.HeaderAd = true

		err := searchNotFoundTemplate.ExecuteTemplate(w, "layout", data)
		if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", 503)
			return
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
	case "blocks":
		if len(search) <= 1 {
			break
		}
		result = &types.SearchAheadBlocksResult{}
		if len(search)%2 != 0 {
			search = search[:len(search)-1]
		}
		if searchLikeRE.MatchString(search) {
			if len(search) < 64 {
				err = db.ReaderDb.Select(result, `
				SELECT slot, ENCODE(blockroot, 'hex') AS blockroot 
				FROM blocks 
				WHERE CAST(slot AS text) LIKE LOWER($1)
				ORDER BY slot LIMIT 10`, search+"%")
			} else if len(search) == 64 {
				blockHash, err := hex.DecodeString(search)
				if err != nil {
					logger.Errorf("error parsing blockHash to int: %v", err)
					http.Error(w, "Internal server error", 503)
					return
				}
				err = db.ReaderDb.Select(result, `
				SELECT slot, ENCODE(blockroot, 'hex') AS blockroot 
				FROM blocks 
				WHERE blockroot = $1 OR
					stateroot = $1
				ORDER BY slot LIMIT 10`, blockHash)
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
		err = db.ReaderDb.Select(result, `
			SELECT block_slot as slot, ENCODE(txhash::bytea, 'hex') AS txhash
			FROM blocks_transactions
			WHERE ENCODE(txhash::bytea, 'hex') LIKE LOWER($1)
			ORDER BY block_slot LIMIT 10`, search+"%")
	case "epochs":
		result = &types.SearchAheadEpochsResult{}
		err = db.ReaderDb.Select(result, "SELECT epoch FROM epochs WHERE CAST(epoch AS text) LIKE $1 ORDER BY epoch LIMIT 10", search+"%")
	case "validators":
		// find all validators that have a index, publickey or name like the search-query
		result = &types.SearchAheadValidatorsResult{}
		query := `
			SELECT
				validatorindex AS index,
				pubkeyhex AS pubkey
			FROM validators
			WHERE CAST(validatorindex AS text) LIKE $1 || '%' OR pubkeyhex LIKE LOWER($1 || '%')
			UNION
			SELECT
				validators.validatorindex AS index,
				validators.pubkeyhex AS pubkey
			FROM validator_names LEFT JOIN validators ON validator_names.name LIKE '%' || $1 || '%' AND validators.pubkey = validator_names.publickey
			ORDER BY index
			LIMIT 10
		`

		// its too slow to search for names
		if thresholdHexLikeRE.MatchString(search) {
			query = `
				SELECT
					validatorindex AS index,
					pubkeyhex AS pubkey
				FROM validators
				WHERE CAST(validatorindex AS text) LIKE $1 || '%'
					OR pubkeyhex LIKE LOWER($1 || '%')
				ORDER BY index
				LIMIT 10
		`
		}

		err = db.ReaderDb.Select(result, query, search)
	case "eth1_addresses":
		// start := time.Now()
		if len(search) <= 1 {
			break
		}
		result = &types.SearchAheadEth1Result{}
		if len(search)%2 != 0 {
			search = search[:len(search)-1]
		}
		if searchLikeRE.MatchString(search) {
			eth1AddressHash, err := hex.DecodeString(search)
			if err != nil {
				logger.Errorf("error parsing eth1AddressHash to hash: %v", err)
				http.Error(w, "Internal server error", 503)
				return
			}
			err = db.ReaderDb.Select(result, `
				SELECT DISTINCT ENCODE(from_address::bytea, 'hex') as from_address
				FROM eth1_deposits
				WHERE from_address LIKE $1 || '%'::bytea 
				LIMIT 10`, eth1AddressHash)
		}
		// logger.WithFields(logrus.Fields{
		// 	"duration": time.Since(start),
		// }).Infof("finished searching for eth1_addresses")
	case "indexed_validators":
		// find all validators that have a publickey or index like the search-query
		result = &types.SearchAheadValidatorsResult{}
		err = db.ReaderDb.Select(result, `
			SELECT validatorindex AS index, pubkeyhex AS pubkey
			FROM validators
			LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
			WHERE CAST(validatorindex AS text) LIKE $1
				OR pubkeyhex LIKE LOWER($1)
				OR LOWER(validator_names.name) LIKE LOWER($2)
			ORDER BY index LIMIT 10`, search+"%", "%"+search+"%")
	case "indexed_validators_by_eth1_addresses":
		if len(search) <= 1 {
			break
		}
		if len(search)%2 != 0 {
			search = search[:len(search)-1]
		}
		if searchLikeRE.MatchString(search) {
			// find validators per eth1-address (limit result by N addresses and M validators per address)
			result = &[]struct {
				Eth1Address      string        `db:"from_address" json:"eth1_address"`
				ValidatorIndices pq.Int64Array `db:"validatorindices" json:"validator_indices"`
				Count            uint64        `db:"count" json:"-"`
			}{}
			eth1AddressHash, err := hex.DecodeString(search)
			if err != nil {
				logger.Errorf("error parsing eth1AddressHash to hex: %v", err)
				http.Error(w, "Internal server error", 503)
				return
			}
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
	default:
		http.Error(w, "Not found", 404)
		return
	}

	if err != nil {
		logger.WithError(err).Error("error doing query for searchAhead")
		http.Error(w, "Internal server error", 503)
		return
	}
	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		logger.WithError(err).Error("error encoding searchAhead")
		http.Error(w, "Internal server error", 503)
	}
}
