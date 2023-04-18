package handlers

import (
	"encoding/json"
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
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
	"index/postGenesis.html",
	"index/preGenesis.html",
	"index/recentBlocks.html",
	"index/recentEpochs.html",
	"index/genesisCountdown.html",
	"index/depositDistribution.html",
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
	data := InitPageData(w, r, "index", "", "", indexTemplateFiles)
	pageData := services.LatestIndexPageData()

	// data.Data.(*types.IndexPageData).ShowSyncingMessage = data.ShowSyncingMessage
	pageData.Countdown = utils.Config.Frontend.Countdown

	pageData.SlotVizData = getSlotVizData(data.CurrentEpoch)

	data.Data = pageData

	if handleTemplateError(w, r, "index.go", "Index", "", indexTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

// IndexPageData will show the main "index" page in json format
func IndexPageData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", utils.Config.Chain.Config.SecondsPerSlot)) // set local cache to the seconds per slot interval

	err := json.NewEncoder(w).Encode(services.LatestIndexPageData())

	if err != nil {
		logger.Errorf("error sending latest index page data: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func getSlotVizData(currentEpoch uint64) *types.SlotVizPageData {
	var visiblFrom uint64
	var visibleTo uint64
	configuration, err := services.GetExplorerConfigurationsWithDefaults()
	if err != nil {
		utils.LogError(err, "Could not load SlotViz configuration for index page", 0)
		return nil
	}
	visiblFrom, err = configuration.GetUInt64Value(services.ConfigurationCategorySlotViz, services.ConfigurationKeyVisibleFromEpoch)
	if err != nil {
		utils.LogError(err, "Could not get visbleFrom for SlotViz on index page", 0)
		return nil
	}
	visibleTo, err = configuration.GetUInt64Value(services.ConfigurationCategorySlotViz, services.ConfigurationKeyVisibleToEpoch)
	if err != nil {
		utils.LogError(err, "Could not get visibleTo for SlotViz on index page", 0)
		return nil
	}
	if visiblFrom <= currentEpoch && visibleTo >= currentEpoch {
		return &types.SlotVizPageData{
			Epochs:   services.LatestSlotVizMetrics(),
			Selector: "slotsViz",
			Config:   configuration[services.ConfigurationCategorySlotViz]}

	}
	return nil
}
