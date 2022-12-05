package handlers

import (
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/types"
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
		err := chartsUnavailableTemplate.ExecuteTemplate(w, "layout", data)
		if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", http.StatusServiceUnavailable)
			return
		}
		return
	}

	// only display the most recent N entries as a preview
	// for i, ch := range *chartsPageData {
	// 	if ch != nil && ch.Data != nil {
	// 		for j, series := range ch.Data.Series {
	// 			switch series.Data.(type) {
	// 			case []interface{}:
	// 				l := len(series.Data.([]interface{}))
	// 				if l > CHART_PREVIEW_POINTS*2 {
	// 					(*chartsPageData)[i].Data.Series[j].Data = series.Data.([]interface{})[l-CHART_PREVIEW_POINTS:]
	// 				}
	// 			default:
	// 				logger.Infof("unknown type: %v for chart: %v", reflect.TypeOf(series.Data), ch.Data.Title)
	// 			}
	// 		}
	// 	}
	// }

	data.Data = chartsPageData

	err := chartsTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
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
		err := chartsUnavailableTemplate.ExecuteTemplate(w, "layout", data)
		if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", http.StatusServiceUnavailable)
			return
		}
		return
	}

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

	SetPageDataTitle(data, fmt.Sprintf("%v Chart", chartData.Title))
	data.Meta.Path = "/charts/" + chartVar
	data.Data = chartData

	err := genericChartTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
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
	err := slotVizTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
