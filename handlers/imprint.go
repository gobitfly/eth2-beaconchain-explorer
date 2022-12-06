package handlers

import (
	"eth2-exporter/templates"
	"eth2-exporter/utils"
	"net/http"
	"path"
)

// Imprint will show the imprint data using a go template
func Imprint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	var imprintTemplate = templates.GetTemplate("layout.html")

	if utils.Config.Frontend.LegalDir == "" {
		imprintTemplate = templates.AddTemplateFile(imprintTemplate, utils.Config.Frontend.Imprint)
	} else {
		imprintTemplate = templates.AddTemplateFile(imprintTemplate, path.Join(utils.Config.Frontend.LegalDir, "index.html"))
	}

	data := InitPageData(w, r, "imprint", "/imprint", "Imprint")
	data.HeaderAd = true

	err := imprintTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}
