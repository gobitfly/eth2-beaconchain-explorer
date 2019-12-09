package handlers

import (
	"encoding/json"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

var indexTemplate = template.Must(template.New("index").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/index.html"))

func Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - Index - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "",
		},
		ShowSyncingMessage: services.IsSyncing(),
		Active:             "index",
		Data:               services.LatestIndexPageData(),
	}

	err := indexTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Fatalf("error executing template for %v route: %v", r.URL.String(), err)
	}
}

func IndexPageData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(services.LatestIndexPageData())

	if err != nil {
		logger.Fatalf("error sending latest index page data: %v", err)
	}
}
