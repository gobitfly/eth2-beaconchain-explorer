package handlers

import (
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var chartsTemplate = template.Must(template.New("charts").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/charts.html"))
var genericChartTemplate = template.Must(template.New("chart").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/genericchart.html"))
var chartsUnavailableTemplate = template.Must(template.New("chart").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/chartsunavailable.html"))

// Charts uses a go template for presenting the page to show charts
func Charts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "stats", "/charts", "Charts")

	chartsPageData := services.LatestChartsPageData()
	if chartsPageData == nil {
		err := chartsUnavailableTemplate.ExecuteTemplate(w, "layout", data)
		if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", 503)
			return
		}
		return
	}

	data.Data = chartsPageData

	chartsTemplate = template.Must(template.New("charts").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/charts.html"))
	err := chartsTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GenericChart uses a go template for presenting the page of a generic chart
func GenericChart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "stats", "/charts", "Chart")

	chartsPageData := services.LatestChartsPageData()
	if chartsPageData == nil {
		err := chartsUnavailableTemplate.ExecuteTemplate(w, "layout", data)
		if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", 503)
			return
		}
		return
	}

	vars := mux.Vars(r)
	chartVar := vars["chart"]
	var chartData *types.GenericChartData
	for _, d := range *chartsPageData {
		if d.Path == chartVar {
			chartData = d.Data
			break
		}
	}

	if chartData == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	data.Meta.Title = fmt.Sprintf("%v - %v Chart - beaconcha.in - %v", chartData.Title, utils.Config.Frontend.SiteName, time.Now().Year())
	data.Meta.Path = "/charts/" + chartVar
	data.Data = chartData

	genericChartTemplate = template.Must(template.New("chart").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/genericchart.html"))
	err := genericChartTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
