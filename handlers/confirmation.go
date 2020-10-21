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

var confirmationTemplate = template.Must(template.New("confirmation").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/confirmation.html"))

// Blocks will return information about blocks using a go template
func Confirmation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	type confirmationPageData struct {
		Flashes []interface{}
	}

	// _, session, err := getUserSession(w, r)
	// if err != nil {
	// 	logger.Errorf("error retrieving session: %v", err)
	// 	http.Error(w, "Internal server error", http.StatusInternalServerError)
	// 	return
	// }
	// session.AddFlash("this is something")
	// session.AddFlash("Error: this is something else")
	// session.Save(r, w)

	pageData := confirmationPageData{}
	pageData.Flashes = utils.GetFlashes(w, r, authSessionName)

	if len(pageData.Flashes) <= 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - Blocks - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/blocks",
			GATag:       utils.Config.Frontend.GATag,
		},
		ShowSyncingMessage:    services.IsSyncing(),
		Active:                "confirmation",
		Data:                  pageData,
		User:                  getUser(w, r),
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
		CurrentEpoch:          services.LatestEpoch(),
		CurrentSlot:           services.LatestSlot(),
		FinalizationDelay:     services.FinalizationDelay(),
	}

	err := confirmationTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
