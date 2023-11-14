package handlers

import (
	"encoding/json"
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

const CHART_PREVIEW_POINTS = 100

// Charts uses a go template for presenting the page to show charts
func Charts(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "charts.html")
	var chartsTemplate = templates.GetTemplate(templateFiles...)
	var chartsUnavailableTemplate = templates.GetTemplate(append(layoutTemplateFiles, "chartsunavailable.html")...)

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "stats", "/charts", "Charts", templateFiles)

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

	disclaimer := ""
	for _, chart := range cpd {
		chart.Data.Series = nil

		// If at least one chart shows info about ETH.STORE, then show the disclaimer
		if disclaimer == "" && strings.Contains(chart.Data.Subtitle, "ETH.STORE®") {
			disclaimer = services.EthStoreDisclaimer()
		}
	}

	data.Data = &types.ChartsPageData{ChartsPageDataCharts: cpd, Disclaimer: disclaimer}

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
	templateFiles := append(layoutTemplateFiles, "genericchart.html")
	var genericChartTemplate = templates.GetTemplate(templateFiles...)
	var chartsUnavailableTemplate = templates.GetTemplate(append(layoutTemplateFiles, "chartsunavailable.html")...)

	vars := mux.Vars(r)
	chartVar := vars["chart"]

	w.Header().Set("Content-Type", "text/html")
	data := InitPageData(w, r, "stats", "/charts", "Chart", templateFiles)

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
		NotFound(w, r)
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
		SendBadRequestResponse(w, r.URL.String(), "error getting chart page data")
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
		SendBadRequestResponse(w, r.URL.String(), fmt.Sprintf("error the chart you requested is not available. Chart: %v", chartVar))
		return
	}

	SendOKResponse(json.NewEncoder(w), r.URL.String(), []interface{}{chartData.Series})
}

// SlotViz renders a single page with a d3 slot (block) visualisation
func SlotViz(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "slotViz.html", "slotVizPage.html")
	var slotVizTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")
	data := InitPageData(w, r, "stats", "/charts", "Charts", templateFiles)

	slotVizData := types.SlotVizPageData{
		Selector: "checklist",
		Epochs:   services.LatestSlotVizMetrics(),
	}
	// The following struct is needed so that we can handle the SlotVizPageData same as in the index.go page.
	data.Data = struct {
		SlotVizData types.SlotVizPageData
	}{
		SlotVizData: slotVizData,
	}
	if handleTemplateError(w, r, "charts.go", "SlotViz", "", slotVizTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}
