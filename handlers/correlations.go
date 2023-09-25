package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/templates"
	"eth2-exporter/types"

	"net/http"
	"time"
)

// Blocks will return information about blocks using a go template
func Correlations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	templateFiles := append(layoutTemplateFiles, "correlations.html")
	data := InitPageData(w, r, "correlations", "/correlations", "Correlations", templateFiles)

	var indicators []string
	err := db.ReaderDb.Select(&indicators, "SELECT DISTINCT(indicator) AS indicator FROM chart_series WHERE time > NOW() - INTERVAL '1 week' ORDER BY indicator;")

	if err != nil {
		logger.Errorf("error retrieving correlation indicators: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data.Data = indicators

	var correlationsTemplate = templates.GetTemplate(templateFiles...)

	if handleTemplateError(w, r, "correlations.go", "Correlations", "", correlationsTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

func CorrelationsData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)

	x := r.FormValue("x")
	y := r.FormValue("y")
	startDate, err := time.Parse("2006-01-02", r.FormValue("startDate"))

	if err != nil {
		logger.Infof("invalid correlation start date %v provided: %v", startDate, err)
		enc.Encode(&types.CorrelationDataResponse{Status: "error", Message: "Invalid or missing parameters"})
		return
	}
	endDate, err := time.Parse("2006-01-02", r.FormValue("endDate"))
	if err != nil {
		logger.Infof("invalid correlation end date %v provided: %v", endDate, err)
		enc.Encode(&types.CorrelationDataResponse{Status: "error", Message: "Invalid or missing parameters"})
		return
	}

	var data []*types.CorrelationData
	err = db.ReaderDb.Select(&data, `
		SELECT indicator, 
		       value, 
		       EXTRACT(epoch from date_trunc('day', time)) as time 
		FROM chart_series 
		WHERE (indicator = $1 OR indicator = $2) AND time >= $3 AND time <= $4`,
		x, y, startDate, endDate)
	if err != nil {
		logger.Infof("error querying correlation data: %v", err)
		enc.Encode(&types.CorrelationDataResponse{Status: "error", Message: "Data error"})
		return
	}

	err = enc.Encode(&types.CorrelationDataResponse{Status: "ok", Data: data})
	if err != nil {
		logger.Errorf("error serializing json data for %v route: %v", r.URL.String(), err)
	}
}
