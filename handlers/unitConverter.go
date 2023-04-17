package handlers

import (
	"eth2-exporter/templates"
	"net/http"
)

// Faq will return the data from the frequently asked questions (FAQ) using a go template
func UnitConverter(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "unitConverter.html")
	var unitConverterTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "unitConverter", "/unitConerter", "Unit Converter", templateFiles)

	if handleTemplateError(w, r, "unitConverter.go", "UnitConverter", "", unitConverterTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}
