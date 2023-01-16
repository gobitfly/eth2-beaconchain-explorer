package handlers

import (
	"eth2-exporter/templates"
	"net/http"
)

// Faq will return the data from the frequently asked questions (FAQ) using a go template
func UnitConverter(w http.ResponseWriter, r *http.Request) {
	var unitConverterTemplate = templates.GetTemplate("layout.html", "unitConverter.html")

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "unitConverter", "/unitConerter", "Unit Converter")

	if handleTemplateError(w, r, unitConverterTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}
