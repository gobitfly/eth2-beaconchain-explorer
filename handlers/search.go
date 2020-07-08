package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

var searchNotFoundTemplate = template.Must(template.New("searchnotfound").ParseFiles("templates/layout.html", "templates/searchnotfound.html"))

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
		data := &types.PageData{
			Meta: &types.Meta{
				Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			},
			ShowSyncingMessage:    services.IsSyncing(),
			Active:                "search",
			Data:                  nil,
			Version:               version.Version,
			ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
			ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
			ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
			CurrentEpoch:          services.LatestEpoch(),
			CurrentSlot:           services.LatestSlot(),
		}
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

	switch searchType {
	case "blocks":
		blocks := &types.SearchAheadBlocksResult{}
		err := db.DB.Select(blocks, `
			SELECT slot, ENCODE(blockroot::bytea, 'hex') AS blockroot 
			FROM blocks 
			WHERE CAST(slot AS text) LIKE $1 OR ENCODE(blockroot::bytea, 'hex') LIKE $1
			ORDER BY slot LIMIT 10`, search+"%")
		if err != nil {
			logger.WithError(err).Error("error doing search-query for blocks")
			http.Error(w, "Internal server error", 503)
			return
		}
		err = json.NewEncoder(w).Encode(blocks)
		if err != nil {
			logger.WithError(err).Error("error encoding searchAhead-blocks-result")
			http.Error(w, "Internal server error", 503)
		}
	case "graffiti":
		graffiti := &types.SearchAheadGraffitiResult{}
		err := db.DB.Select(graffiti, `
			SELECT graffiti, count(*)
			FROM blocks
			WHERE LOWER(ENCODE(graffiti , 'escape')) LIKE LOWER($1)
			GROUP BY graffiti
			ORDER BY count desc
			LIMIT 10`, "%"+search+"%")
		if err != nil {
			logger.WithError(err).Error("error doing search-query for graffiti")
			http.Error(w, "Internal server error", 503)
			return
		}
		for i := range *graffiti {
			(*graffiti)[i].Graffiti = utils.FormatGraffitiString((*graffiti)[i].Graffiti)
		}
		err = json.NewEncoder(w).Encode(graffiti)
		if err != nil {
			logger.WithError(err).Error("error encoding searchAhead-graffiti-result")
			http.Error(w, "Internal server error", 503)
		}
	case "epochs":
		epochs := &types.SearchAheadEpochsResult{}
		err := db.DB.Select(epochs, "SELECT epoch FROM epochs WHERE CAST(epoch AS text) LIKE $1 ORDER BY epoch LIMIT 10", search+"%")
		if err != nil {
			logger.WithError(err).Error("error doing search-query for epochs")
			http.Error(w, "Internal server error", 503)
			return
		}
		err = json.NewEncoder(w).Encode(epochs)
		if err != nil {
			logger.WithError(err).Error("error encoding searchAhead-epochs-result")
			http.Error(w, "Internal server error", 503)
		}
	case "validators":
		// find all validators that have a publickey or index like the search-query
		// or validators that have deposited to the eth1-deposit-contract but did not get included into the beaconchain yet
		validators := &types.SearchAheadValidatorsResult{}
		err := db.DB.Select(validators, `
			SELECT CAST(validatorindex AS text) AS index, ENCODE(pubkey::bytea, 'hex') AS pubkey
			FROM validators
			WHERE ENCODE(pubkey::bytea, 'hex') LIKE LOWER($1)
				OR CAST(validatorindex AS text) LIKE $1
			UNION
			SELECT 'deposited' AS index, ENCODE(publickey::bytea, 'hex') as pubkey 
			FROM eth1_deposits 
			LEFT JOIN validators ON eth1_deposits.publickey = validators.pubkey
			WHERE validators.pubkey IS NULL AND 
				(
					ENCODE(publickey::bytea, 'hex') LIKE LOWER($1)
					OR ENCODE(from_address::bytea, 'hex') LIKE LOWER($1)
				)
			ORDER BY index LIMIT 10`, search+"%")
		if err != nil {
			logger.WithError(err).Error("error doing search-query for validators")
			http.Error(w, "Internal server error", 503)
			return
		}
		err = json.NewEncoder(w).Encode(validators)
		if err != nil {
			logger.WithError(err).Error("error encoding searchAhead-validators-result")
			http.Error(w, "Internal server error", 503)
		}
	case "eth1_addresses":
		eth1 := &types.SearchAheadEth1Result{}
		err := db.DB.Select(eth1, `
			SELECT DISTINCT ENCODE(from_address::bytea, 'hex') as from_address
			FROM eth1_deposits
			WHERE ENCODE(from_address::bytea, 'hex') LIKE LOWER($1)
			LIMIT 10`, search+"%")
		if err != nil {
			logger.WithError(err).Error("error doing search-query for eth1_addresses")
			http.Error(w, "Internal server error", 503)
			return
		}
		err = json.NewEncoder(w).Encode(eth1)
		if err != nil {
			logger.WithError(err).Error("error encoding searchAhead-blocks-result")
			http.Error(w, "Internal server error", 503)
		}
	case "indexed_validators":
		// find all validators that have a publickey or index like the search-query
		validators := &types.SearchAheadValidatorsResult{}
		err := db.DB.Select(validators, `
			SELECT DISTINCT CAST(validatorindex AS text) AS index, ENCODE(pubkey::bytea, 'hex') AS pubkey
			FROM validators
			LEFT JOIN eth1_deposits ON eth1_deposits.publickey = validators.pubkey
			WHERE ENCODE(pubkey::bytea, 'hex') LIKE LOWER($1)
				OR CAST(validatorindex AS text) LIKE $1
				OR ENCODE(from_address::bytea, 'hex') LIKE LOWER($1)
			ORDER BY index LIMIT 10`, search+"%")
		if err != nil {
			logger.WithError(err).Error("error doing search-query for indexed_validators")
			http.Error(w, "Internal server error", 503)
			return
		}
		err = json.NewEncoder(w).Encode(validators)
		if err != nil {
			logger.WithError(err).Error("error encoding searchAhead-indexedvalidators-result")
			http.Error(w, "Internal server error", 503)
		}
	case "indexed_validators_by_eth1_addresses":
		result := []struct {
			Eth1Address      string        `db:"from_address" json:"eth1_address"`
			ValidatorIndices pq.Int64Array `db:"validatorindices" json:"validator_indices"`
			Count            uint64        `db:"count" json:"-"`
		}{}
		// find validators per eth1-address (limit result by 10 addresses and 100 validators per address)
		err := db.DB.Select(&result, `
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
			WHERE validatorrow <= 101 AND addressrow <= 10
			GROUP BY from_address
			ORDER BY count DESC`, search+"%")
		if err != nil {
			logger.WithError(err).Error("error doing search-query for indexed_validators_by_eth1_addresses")
			http.Error(w, "Internal server error", 503)
			return
		}
		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			logger.WithError(err).Error("error encoding searchAhead-blocks-result")
			http.Error(w, "Internal server error", 503)
		}
	case "indexed_validators_by_graffiti":
		result := []struct {
			Graffiti         string        `db:"graffiti" json:"graffiti"`
			ValidatorIndices pq.Int64Array `db:"validatorindices" json:"validator_indices"`
			Count            uint64        `db:"count" json:"-"`
		}{}
		// find validators per graffiti (limit result by 10 graffities and 100 validators per graffiti)
		err := db.DB.Select(&result, `
			SELECT graffiti, COUNT(*), ARRAY_AGG(validatorindex) validatorindices FROM (
				SELECT 
					DISTINCT ON(validatorindex) validatorindex,
					graffiti,
					DENSE_RANK() OVER(PARTITION BY graffiti ORDER BY validatorindex) AS validatorrow,
					DENSE_RANK() OVER(ORDER BY graffiti) AS graffitirow
				FROM blocks 
				LEFT JOIN validators ON blocks.proposer = validators.validatorindex
				WHERE LOWER(ENCODE(graffiti , 'escape')) LIKE LOWER($1)
			) a 
			WHERE validatorrow <= 101 AND graffitirow <= 10
			GROUP BY graffiti
			ORDER BY count DESC`, "%"+search+"%")
		if err != nil {
			logger.WithError(err).Error("error doing search-query for indexed_validators_by_graffiti")
			http.Error(w, "Internal server error", 503)
			return
		}
		for i := range result {
			result[i].Graffiti = utils.FormatGraffitiString(result[i].Graffiti)
		}
		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			logger.WithError(err).Error("error encoding searchAhead-blocks-result")
			http.Error(w, "Internal server error", 503)
		}
	default:
		http.Error(w, "Not found", 404)
	}
}
