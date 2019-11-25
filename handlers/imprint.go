package handlers

import (
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"html/template"
	"net/http"
)

func Imprint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	imprintTemplate, err := template.ParseFiles("templates/layout.html", utils.Config.Frontend.Imprint)

	if err != nil {
		logger.Printf("Error parsing imprint page template: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       "Imprint - beaconcha.in",
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/imprint",
		},
		Active: "imprint",
		Data:   nil,
	}

	err = imprintTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Fatalf("Error executing template for %v route: %v", r.URL.String(), err)
	}
}
