package handlers

import (
	"encoding/json"
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

const CHART_PREVIEW_POINTS = 100

// Charts uses a go template for presenting the page to show charts
func Charts(w http.ResponseWriter, r *http.Request) {

	var chartsTemplate = templates.GetTemplate("layout.html", "charts.html")
	var chartsUnavailableTemplate = templates.GetTemplate("layout.html", "chartsunavailable.html")

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "stats", "/charts", "Charts")

	chartsPageData := services.LatestChartsPageData()

	if chartsPageData == nil {
		if handleTemplateError(w, r, "charts.go", "Charts", "LatestChartsPageData", chartsUnavailableTemplate.ExecuteTemplate(w, "layout", data)) != nil {
			return // an error has occurred and was processed
		}
		return
	}

	cpd := make([]types.ChartsPageDataChart, 0, len(chartsPageData))
	for i := 0; i < len(chartsPageData); i++ {
		chartData := *chartsPageData[i]
		data := *(*chartsPageData[i]).Data
		chartData.Data = &data
		cpd = append(cpd, chartData)
	}

	for _, chart := range cpd {
		chart.Data.Series = nil
	}

	data.Data = cpd
	if handleTemplateError(w, r, "charts.go", "Charts", "Done", chartsTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

// Chart renders a single chart
func Chart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chartVar := vars["chart"]
	switch chartVar {
	case "slotviz":
		SlotViz(w, r)
	default:
		GenericChart(w, r)
	}
}

// GenericChart uses a go template for presenting the page of a generic chart
func GenericChart(w http.ResponseWriter, r *http.Request) {

	var genericChartTemplate = templates.GetTemplate("layout.html", "genericchart.html")
	var chartsUnavailableTemplate = templates.GetTemplate("layout.html", "chartsunavailable.html")

	vars := mux.Vars(r)
	chartVar := vars["chart"]

	w.Header().Set("Content-Type", "text/html")
	data := InitPageData(w, r, "stats", "/charts", "Chart")

	chartsPageData := services.LatestChartsPageData()
	if chartsPageData == nil {
		if handleTemplateError(w, r, "charts.go", "GenericChart", "LatestChartsPageData", chartsUnavailableTemplate.ExecuteTemplate(w, "layout", data)) != nil {
			return // an error has occurred and was processed
		}
		return
	}

	var chartData *types.GenericChartData
	for _, d := range chartsPageData {
		if d.Path == chartVar {
			chartData = d.Data
			break
		}
	}

	if chartData == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	SetPageDataTitle(data, fmt.Sprintf("%v Chart", chartData.Title))
	data.Meta.Path = "/charts/" + chartVar
	data.Data = chartData

	if handleTemplateError(w, r, "charts.go", "GenericChart", "Done", genericChartTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

func GenericChartData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chartVar := vars["chart"]

	w.Header().Set("Content-Type", "application/json")

	chartsPageData := services.LatestChartsPageData()
	if chartsPageData == nil {
		utils.LogError(nil, "error getting chart page data", 0)
		SendErrorResponse(w, r.URL.String(), "error getting chart page data")
		return
	}

	var chartData *types.GenericChartData
	for _, d := range chartsPageData {
		if fmt.Sprintf("chart-holder-%d", d.Order) == chartVar {
			chartData = d.Data
			break
		}
	}

	if chartData == nil {
		SendErrorResponse(w, r.URL.String(), fmt.Sprintf("error the chart you requested is not available. Chart: %v", chartVar))
		return
	}

	sendOKResponse(json.NewEncoder(w), r.URL.String(), []interface{}{chartData.Series})
}

// SlotViz renders a single page with a d3 slot (block) visualisation
func SlotViz(w http.ResponseWriter, r *http.Request) {
	var slotVizTemplate = templates.GetTemplate("layout.html", "slotViz.html", "slotVizPage.html")

	w.Header().Set("Content-Type", "text/html")
	data := InitPageData(w, r, "stats", "/charts", "Charts")

	slotVizData := types.SlotVizPageData{
		Selector: "checklist",
		Epochs:   services.LatestSlotVizMetrics(),
	}
	data.Data = slotVizData
	if handleTemplateError(w, r, "charts.go", "SlotViz", "", slotVizTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}
