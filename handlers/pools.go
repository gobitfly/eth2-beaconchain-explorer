package handlers

import (
	// "eth2-exporter/db"

	"eth2-exporter/services"
	"eth2-exporter/templates"
	"net/http"
	// "strings"
)

func Pools(w http.ResponseWriter, r *http.Request) {
	var poolsServicesTemplate = templates.GetTemplate(
		"layout.html",
		"pools/pools.html",
		"pools/loadingSvg.html",
		"pools/charts.html",
		"bannerPools.html")

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "services", "/pools", "Staking Pools Services Overview")

	distributionData, err := services.ChartHandlers["pools_distribution"].DataFunc()
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	performanceData, err := services.ChartHandlers["historic_pool_performance"].DataFunc()
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	poolData := services.LatestPoolsPageData()

	poolData.PoolsDistribution.Data = distributionData
	poolData.PoolsDistribution.Height = 500
	poolData.PoolsDistribution.Path = "pools_distribution"

	poolData.HistoricPoolPerformance.Data = performanceData
	poolData.HistoricPoolPerformance.Height = 500
	poolData.HistoricPoolPerformance.Path = "historic_pool_performance"

	data.Data = poolData

	if handleTemplateError(w, r, poolsServicesTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}
