package handlers

import (
	"eth2-exporter/services"
	types "eth2-exporter/types"
	"eth2-exporter/utils"
	"html/template"
	"net/http"
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
