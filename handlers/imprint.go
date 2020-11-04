package handlers

import (
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

// Imprint will show the imprint data using a go template
func Imprint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	imprintTemplate, err := template.ParseFiles("templates/layout.html", utils.Config.Frontend.Imprint)

	if err != nil {
		logger.Errorf("error parsing imprint page template: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	data := &types.PageData{
		HeaderAd: true,
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - Imprint - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/imprint",
			GATag:       utils.Config.Frontend.GATag,
		},
		ShowSyncingMessage:    services.IsSyncing(),
		Active:                "imprint",
		Data:                  nil,
		User:                  getUser(w, r),
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
		CurrentEpoch:          services.LatestEpoch(),
		CurrentSlot:           services.LatestSlot(),
		FinalizationDelay:     services.FinalizationDelay(),
		Mainnet:               utils.Config.Chain.Mainnet,
		DepositContract:       utils.Config.Indexer.Eth1DepositContractAddress,
	}

	err = imprintTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
