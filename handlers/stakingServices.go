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

var stakingServicesTemplate = template.Must(template.New("stakingServices").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/stakingServices.html"))

func StakingServices(w http.ResponseWriter, r *http.Request) {
	var err error
	stakingServicesTemplate = template.Must(template.New("stakingServices").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/stakingServices.html"))

	w.Header().Set("Content-Type", "text/html")
	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - Ethereum 2.0 Staking Services Overview - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/stakingServices",
			GATag:       utils.Config.Frontend.GATag,
		},
		ShowSyncingMessage:    services.IsSyncing(),
		Active:                "stakingServices",
		User:                  getUser(w, r),
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
		CurrentEpoch:          services.LatestEpoch(),
		CurrentSlot:           services.LatestSlot(),
		FinalizationDelay:     services.FinalizationDelay(),
	}

	err = stakingServicesTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
