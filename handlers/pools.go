package handlers

import (
	// "eth2-exporter/db"

	"eth2-exporter/services"
	"eth2-exporter/utils"
	"html/template"
	"net/http"
	// "strings"
)

var poolsServicesTemplate = template.Must(template.New("poolsServices").Funcs(utils.GetTemplateFuncs()).ParseFiles(
	"templates/layout.html",
	"templates/pools.html",
	"templates/bannerPools.html"))

func Pools(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "services", "/pools", "Staking Pools Services Overview")

	chartData, err := services.ChartHandlers["pools_distribution"].DataFunc()
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	poolData := services.LatestPoolsPageData()

	poolData.PoolsDistribution.Data = chartData
	poolData.PoolsDistribution.Height = 500
	poolData.PoolsDistribution.Path = "pools_distribution"
	data.Data = poolData

	err = poolsServicesTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}
