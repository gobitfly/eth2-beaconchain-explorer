package handlers

import (
	"encoding/json"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

var indexTemplate = template.Must(template.New("index").Funcs(utils.GetTemplateFuncs()).ParseFiles(
	"templates/layout.html",
	"templates/index/index.html",
	"templates/index/depositProgress.html",
	"templates/index/genesis.html",
	"templates/index/hero.html",
	"templates/index/networkStats.html",
	"templates/index/participationWarning.html",
	"templates/index/postGenesis.html",
	"templates/index/preGenesis.html",
	"templates/index/recentBlocks.html",
	"templates/index/recentEpochs.html",
	"templates/index/genesisCountdown.html",
	"templates/components/banner.html",
	"templates/svg/bricks.html",
	"templates/svg/professor.html",
	"templates/svg/timeline.html",
))

// Index will return the main "index" page using a go template
func Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	// indexTemplate = template.Must(template.New("index").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/index.html"))
	user := getUser(w, r)
	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - Index - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "",
			GATag:       utils.Config.Frontend.GATag,
		},
		ShowSyncingMessage:    services.IsSyncing(),
		Active:                "index",
		User:                  user,
		Data:                  services.LatestIndexPageData(),
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
		CurrentEpoch:          services.LatestEpoch(),
		CurrentSlot:           services.LatestSlot(),
		FinalizationDelay:     services.FinalizationDelay(),
	}

	data.Data.(*types.IndexPageData).ShowSyncingMessage = data.ShowSyncingMessage

	err := indexTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

// IndexPageData will show the main "index" page in json format
func IndexPageData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(services.LatestIndexPageData())

	if err != nil {
		logger.Errorf("error sending latest index page data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
