package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

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
		http.Error(w, "Not found", 404)
	}
}

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
			logger.WithError(err).Error("Failed doing search-query")
			http.Error(w, "Internal server error", 503)
			return
		}
		err = json.NewEncoder(w).Encode(blocks)
		if err != nil {
			logger.WithError(err).Error("Failed encoding searchAhead-blocks-result")
			http.Error(w, "Internal server error", 503)
		}
	case "epochs":
		epochs := &types.SearchAheadEpochsResult{}
		err := db.DB.Select(epochs, "SELECT epoch FROM epochs WHERE CAST(epoch AS text) LIKE $1 ORDER BY epoch LIMIT 10", search+"%")
		if err != nil {
			logger.WithError(err).Error("Failed doing search-query")
			http.Error(w, "Internal server error", 503)
			return
		}
		err = json.NewEncoder(w).Encode(epochs)
		if err != nil {
			logger.WithError(err).Error("Failed encoding searchAhead-epochs-result")
			http.Error(w, "Internal server error", 503)
		}
	case "validators":
		validators := &types.SearchAheadValidatorsResult{}
		err := db.DB.Select(validators, "SELECT validatorindex AS index, ENCODE(pubkey::bytea, 'hex') AS pubkey FROM validators WHERE CAST(validatorindex AS text) LIKE $1 OR ENCODE(pubkey::bytea, 'hex') LIKE $1 ORDER BY index LIMIT 10", search+"%")
		if err != nil {
			logger.WithError(err).Error("Failed doing search-query")
			http.Error(w, "Internal server error", 503)
			return
		}
		err = json.NewEncoder(w).Encode(validators)
		if err != nil {
			logger.WithError(err).Error("Failed encoding searchAhead-validators-result")
			http.Error(w, "Internal server error", 503)
		}
	default:
		http.Error(w, "Not found", 404)
	}
}
