package handlers

import (
	// "eth2-exporter/db"
	"eth2-exporter/services"
	types "eth2-exporter/types"
	"eth2-exporter/utils"
	"html/template"
	"net/http"
	// "strings"
)

var poolsServicesTemplate = template.Must(template.New("poolsServices").Funcs(utils.GetTemplateFuncs()).ParseFiles(
	"templates/layout.html",
	"templates/poolsServices.html",
	"templates/index/depositDistribution.html"))

func Pools(w http.ResponseWriter, r *http.Request) {
	// var err error

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "services", "/pools", "Stacking Pools Services Overview")

	chartData, err := services.ChartHandlers["deposits_distribution"].DataFunc()
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
	}

	// type pools struct {
	// 	Address string
	// 	Name    string
	// }

	// var stakePools []pools
	// err = db.DB.Select(&stakePools, "select address, name from stake_pools_stats;")
	// if err != nil {
	// 	logger.Errorf("error retrieving stake pools stats: %v", err)
	// 	http.Error(w, "Internal server error", 503)
	// 	return
	// }
	// // logger.Errorln(stakePools)
	// // logger.Errorf("%T", chartData.Series[0].Data.([]types.SeriesDataItem))
	// for _, pool := range stakePools{
	// 	for i, slice := range chartData.Series[0].Data.([]types.SeriesDataItem){
	// 		logger.Errorln(i, slice.Name, "0x"+pool.Address, (slice.Name == "0x"+pool.Address))
	// 		if strings.ToLower(slice.Name) == strings.ToLower("0x"+pool.Address) {
	// 			chartData.Series[0].Data.([]types.SeriesDataItem)[i].Name = pool.Name
	// 			// break
	// 		}
	// 	}
	// 	logger.Errorln("")
	// }
	type chart struct {
		DepositDistribution types.ChartsPageDataChart
	}
	var pieChart chart
	pieChart.DepositDistribution.Data = chartData
	pieChart.DepositDistribution.Height = 500
	pieChart.DepositDistribution.Path = "deposits_distribution"

	data.Data = pieChart

	// pageData.CsrfField = csrf.TemplateField(r)
	// data.Data = pageData
	err = poolsServicesTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
