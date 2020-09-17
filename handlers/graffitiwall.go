package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

var graffitiwallTemplate = template.Must(template.New("vis").ParseFiles("templates/layout.html", "templates/graffitiwall.html"))

func Graffitiwall(w http.ResponseWriter, r *http.Request) {

	var err error

	w.Header().Set("Content-Type", "text/html")

	var graffitiwallData []*types.GraffitiwallData

	err = db.DB.Select(&graffitiwallData, "select x, y, color, slot, validator from graffitiwall")

	if err != nil {
		logger.Errorf("error retrieving block tree data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	data := &types.PageData{
		HeaderAd: true,
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - Graffitiwall - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/graffitiwall",
			GATag:       utils.Config.Frontend.GATag,
		},
		ShowSyncingMessage:    services.IsSyncing(),
		Active:                "more",
		Data:                  graffitiwallData,
		User:                  getUser(w, r),
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
		CurrentEpoch:          services.LatestEpoch(),
		CurrentSlot:           services.LatestSlot(),
		FinalizationDelay:     services.FinalizationDelay(),
	}

	err = graffitiwallTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
