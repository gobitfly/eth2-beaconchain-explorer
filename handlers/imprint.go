package handlers

import (
	"eth2-exporter/templates"
	"eth2-exporter/utils"
	"html/template"
	"net/http"
)

// Imprint will show the imprint data using a go template
func Imprint(w http.ResponseWriter, r *http.Request) {

	imprintTemplate := templates.GetTemplate(layoutTemplateFiles...)
	imprintTemplate = template.Must(imprintTemplate.Parse(utils.Config.Frontend.Legal.ImprintTemplate))
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "imprint", "/imprint", "Imprint", layoutTemplateFiles)

	if handleTemplateError(w, r, "imprint.go", "Imprint", "", imprintTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}
