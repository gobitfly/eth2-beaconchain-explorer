package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"net/http"
)

func Graffitiwall(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "graffitiwall.html")
	var graffitiwallTemplate = templates.GetTemplate(templateFiles...)

	var err error

	w.Header().Set("Content-Type", "text/html")

	var graffitiwallData []*types.GraffitiwallData

	// only fetch latest entry for each pixel
	err = db.ReaderDb.Select(&graffitiwallData, "SELECT DISTINCT ON (x, y) x, y, color, slot, validator from graffitiwall ORDER BY x, y, slot DESC")

	if err != nil {
		logger.Errorf("error retrieving block tree data: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	data := InitPageData(w, r, "more", "/graffitiwall", "Graffitiwall", templateFiles)
	data.Data = graffitiwallData

	if handleTemplateError(w, r, "graffitiwall.go", "Graffitiwall", "", graffitiwallTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}
