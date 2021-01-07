package handlers

import (
	"encoding/json"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"html/template"
	"net/http"
	"strings"
)

var indexTemplate = template.Must(template.New("index").Funcs(utils.GetTemplateFuncs()).ParseFiles(
	"templates/layout.html",
	"templates/index/index.html",
	"templates/index/depositProgress.html",
	"templates/index/depositChart.html",
	"templates/index/genesis.html",
	"templates/index/hero.html",
	"templates/index/networkStats.html",
	"templates/index/participationWarning.html",
	"templates/index/postGenesis.html",
	"templates/index/preGenesis.html",
	"templates/index/recentBlocks.html",
	"templates/index/recentEpochs.html",
	"templates/index/genesisCountdown.html",
	"templates/index/depositDistribution.html",
	"templates/components/banner.html",
	"templates/svg/bricks.html",
	"templates/svg/professor.html",
	"templates/svg/timeline.html",
	"templates/components/rocket.html",
))

// Index will return the main "index" page using a go template
func Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	data := InitPageData(w, r, "index", "", "Index")
	data.Data = services.LatestIndexPageData()

	data.Data.(*types.IndexPageData).ShowSyncingMessage = data.ShowSyncingMessage

	acceptedLangs := strings.Split(r.Header.Get("Accept-Language"), ",")

	if len(acceptedLangs) > 0 {
		if strings.Contains(acceptedLangs[0], "ru") || strings.Contains(acceptedLangs[0], "RU") {
			data.Data.(*types.IndexPageData).Lang = "ru-RU"
		}
	}

	for _, v := range r.Cookies() {
		if v.Name == "language" {
			data.Data.(*types.IndexPageData).Lang = v.Value
			break
		}
	}
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
