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
		err := db.DB.Select(blocks, "SELECT slot, ENCODE(blockroot::bytea, 'hex') AS blockroot FROM blocks WHERE CAST(slot AS text) LIKE $1 OR ENCODE(blockroot::bytea, 'hex') LIKE $1 ORDER BY slot LIMIT 10", search+"%")
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
		err := db.DB.Select(graffiti, "SELECT slot, ENCODE(blockroot::bytea, 'hex') AS blockroot, graffiti FROM blocks WHERE LOWER(ENCODE(graffiti , 'escape')) LIKE LOWER($1) LIMIT 10", "%"+search+"%")
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
			WHERE ENCODE(pubkey::bytea, 'hex') LIKE $1 
				OR CAST(validatorindex AS text) LIKE $1
			UNION
			SELECT 'deposited' AS index, ENCODE(publickey::bytea, 'hex') as pubkey 
			FROM eth1_deposits 
			LEFT JOIN validators ON eth1_deposits.publickey = validators.pubkey
			WHERE validators.pubkey IS NULL AND 
				(
					ENCODE(publickey::bytea, 'hex') LIKE $1
					OR ENCODE(from_address::bytea, 'hex') LIKE $1
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
	case "eth1deposits":
		eth1 := &types.SearchAheadEth1Result{}
		err := db.DB.Select(eth1, `
		SELECT DISTINCT
			ENCODE(from_address::bytea, 'hex') as from_address
		FROM
		 eth1_deposits
		WHERE
		ENCODE(from_address::bytea, 'hex') LIKE $1
		LIMIT 10`, search+"%")
		if err != nil {
			logger.WithError(err).Error("error doing search-query")
			http.Error(w, "Internal server error", 503)
			return
		}
		err = json.NewEncoder(w).Encode(eth1)
		if err != nil {
			logger.WithError(err).Error("error encoding searchAhead-blocks-result")
			http.Error(w, "Internal server error", 503)
		}
	default:
		http.Error(w, "Not found", 404)
	}
}
