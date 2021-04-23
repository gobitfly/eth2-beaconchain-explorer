package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

var searchNotFoundTemplate = template.Must(template.New("searchnotfound").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/searchnotfound.html"))

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

	logger := logger.WithField("searchType", searchType)

	var err error
	var result interface{}

	switch searchType {
	case "blocks":
		result = &types.SearchAheadBlocksResult{}
		err = db.DB.Select(result, `
			SELECT slot, ENCODE(blockroot::bytea, 'hex') AS blockroot 
			FROM blocks 
			WHERE CAST(slot AS text) LIKE $1 OR ENCODE(blockroot::bytea, 'hex') LIKE $1
			ORDER BY slot LIMIT 10`, search+"%")
	case "graffiti":
		graffiti := &types.SearchAheadGraffitiResult{}
		err = db.DB.Select(graffiti, `
			SELECT graffiti, count(*)
			FROM blocks
			WHERE 
				LOWER(ENCODE(graffiti , 'escape')) LIKE LOWER($1)
				OR ENCODE(graffiti, 'hex') LIKE ($2)
			GROUP BY graffiti
			ORDER BY count desc
			LIMIT 10`, "%"+search+"%", fmt.Sprintf("%%%x%%", search))
		if err == nil {
			for i := range *graffiti {
				(*graffiti)[i].Graffiti = utils.FormatGraffitiString((*graffiti)[i].Graffiti)
			}
		}
		result = graffiti
	case "epochs":
		result = &types.SearchAheadEpochsResult{}
		err = db.DB.Select(result, "SELECT epoch FROM epochs WHERE CAST(epoch AS text) LIKE $1 ORDER BY epoch LIMIT 10", search+"%")
	case "validators":
		// find all validators that have a index, publickey or name like the search-query
		result = &types.SearchAheadValidatorsResult{}
		err = db.DB.Select(result, `
			SELECT
				validatorindex AS index,
				pubkeyhex AS pubkey
			FROM validators
			LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
			WHERE CAST(validatorindex AS text) LIKE $1
				OR pubkeyhex LIKE LOWER($1)
				OR LOWER(validator_names.name) LIKE LOWER($2)
			ORDER BY index LIMIT 10`, search+"%", "%"+search+"%")
	case "eth1_addresses":
		result = &types.SearchAheadEth1Result{}
		err = db.DB.Select(result, `
			SELECT DISTINCT ENCODE(from_address::bytea, 'hex') as from_address
			FROM eth1_deposits
			WHERE ENCODE(from_address::bytea, 'hex') LIKE LOWER($1)
			LIMIT 10`, search+"%")
	case "indexed_validators":
		// find all validators that have a publickey or index like the search-query
		result = &types.SearchAheadValidatorsResult{}
		err = db.DB.Select(result, `
			SELECT validatorindex AS index, pubkeyhex AS pubkey
			FROM validators
			LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
			WHERE CAST(validatorindex AS text) LIKE $1
				OR pubkeyhex LIKE LOWER($1)
				OR LOWER(validator_names.name) LIKE LOWER($2)
			ORDER BY index LIMIT 10`, search+"%", "%"+search+"%")
	case "indexed_validators_by_eth1_addresses":
		// find validators per eth1-address (limit result by N addresses and M validators per address)
		result = &[]struct {
			Eth1Address      string        `db:"from_address" json:"eth1_address"`
			ValidatorIndices pq.Int64Array `db:"validatorindices" json:"validator_indices"`
			Count            uint64        `db:"count" json:"-"`
		}{}
		err = db.DB.Select(result, `
			SELECT from_address, COUNT(*), ARRAY_AGG(validatorindex) validatorindices FROM (
				SELECT 
					DISTINCT ON(validatorindex) validatorindex,
					ENCODE(from_address::bytea, 'hex') as from_address,
					ROW_NUMBER() OVER (PARTITION BY from_address ORDER BY validatorindex) AS validatorrow,
					DENSE_RANK() OVER (ORDER BY from_address) AS addressrow
				FROM eth1_deposits
				INNER JOIN validators ON validators.pubkey = eth1_deposits.publickey
				WHERE ENCODE(from_address::bytea, 'hex') LIKE LOWER($1) 
			) a 
			WHERE validatorrow <= 301 AND addressrow <= 10
			GROUP BY from_address
			ORDER BY count DESC`, search+"%")
	case "indexed_validators_by_graffiti":
		// find validators per graffiti (limit result by N graffities and M validators per graffiti)
		res := []struct {
			Graffiti         string        `db:"graffiti" json:"graffiti"`
			ValidatorIndices pq.Int64Array `db:"validatorindices" json:"validator_indices"`
			Count            uint64        `db:"count" json:"-"`
		}{}
		err = db.DB.Select(&res, `
			SELECT graffiti, COUNT(*), ARRAY_AGG(validatorindex) validatorindices FROM (
				SELECT 
					DISTINCT ON(validatorindex) validatorindex,
					graffiti,
					DENSE_RANK() OVER(PARTITION BY graffiti ORDER BY validatorindex) AS validatorrow,
					DENSE_RANK() OVER(ORDER BY graffiti) AS graffitirow
				FROM blocks 
				LEFT JOIN validators ON blocks.proposer = validators.validatorindex
				WHERE 
					LOWER(ENCODE(graffiti , 'escape')) LIKE LOWER($1)
					OR ENCODE(graffiti, 'hex') LIKE ($2)
			) a 
			WHERE validatorrow <= 301 AND graffitirow <= 10
			GROUP BY graffiti
			ORDER BY count DESC`, "%"+search+"%", fmt.Sprintf("%%%x%%", search))
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
		err = db.DB.Select(&res, `
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
			WHERE validatorrow <= 301 AND namerow <= 10
			GROUP BY name
			ORDER BY count DESC, name DESC`, "%"+search+"%")
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
