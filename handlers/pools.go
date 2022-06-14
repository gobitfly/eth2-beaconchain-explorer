package handlers

import (
	// "eth2-exporter/db"

	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"html/template"
	"net/http"
	// "strings"
)

var poolsServicesTemplate = template.Must(template.New("poolsServices").Funcs(utils.GetTemplateFuncs()).ParseFiles(
	"templates/layout.html",
	"templates/poolsServices.html",
	"templates/bannerPoolsServices.html",
	"templates/index/poolsDistribution.html"))

func Pools(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "services", "/pools", "Staking Pools Services Overview")

	chartData, err := services.ChartHandlers["pools_distribution"].DataFunc()
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	var poolData types.PoolsResp

	err = db.ReaderDb.Select(&poolData.PoolInfos, "select coalesce(pool, 'Unknown') as name, count(*) as count, avg(performance31d)::integer as avg_performance_31d, avg(performance7d)::integer as avg_performance_7d, avg(performance1d)::integer as avg_performance_1d from validators left outer join validator_pool on validators.pubkey = validator_pool.publickey left outer join validator_performance on validators.validatorindex = validator_performance.validatorindex where validators.status in ('active_online', 'active_offline') group by name order by count(*) desc;")
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

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
