package handlers

import (
	"encoding/json"
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"html/template"
	"net/http"
)

var indexTemplateFiles = append(layoutTemplateFiles,
	"index/index.html",
	"index/depositProgress.html",
	"index/depositChart.html",
	"index/genesis.html",
	"index/hero.html",
	"index/networkStats.html",
	"index/participationWarning.html",
	"index/postGenesis.html",
	"index/preGenesis.html",
	"index/recentBlocks.html",
	"index/recentEpochs.html",
	"index/genesisCountdown.html",
	"index/depositDistribution.html",
	"components/banner.html",
	"svg/bricks.html",
	"svg/professor.html",
	"svg/timeline.html",
	"components/rocket.html",
	"slotViz.html",
)

var indexTemplate = template.Must(template.New("index").Funcs(utils.GetTemplateFuncs()).ParseFS(templates.Files,
	indexTemplateFiles...,
))

// Index will return the main "index" page using a go template
func Index(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html")
	data := InitPageData(w, r, "index", "", "")
	data.Data = services.LatestIndexPageData()

	// data.Data.(*types.IndexPageData).ShowSyncingMessage = data.ShowSyncingMessage
	data.Data.(*types.IndexPageData).Countdown = utils.Config.Frontend.Countdown

	if utils.Config.Frontend.SlotViz.Enabled {
		data.Data.(*types.IndexPageData).SlotVizData = struct {
			Epochs        []*types.SlotVizEpochs
			Selector      string
			HardforkEpoch uint64
		}{
			Epochs:        services.LatestSlotVizMetrics(),
			Selector:      "slotsViz",
			HardforkEpoch: utils.Config.Frontend.SlotViz.HardforkEpoch,
		}
	}

	if handleTemplateError(w, r, "index.go", "Index", "", indexTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

// IndexPageData will show the main "index" page in json format
func IndexPageData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(services.LatestIndexPageData())

	if err != nil {
		logger.Errorf("error sending latest index page data: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}
