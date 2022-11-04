package handlers

import (
	"eth2-exporter/templates"
	"eth2-exporter/utils"
	"net/http"
	"path"
)

// Imprint will show the imprint data using a go template
func Imprint(w http.ResponseWriter, r *http.Request) {
	templatePath := utils.Config.Frontend.Imprint
	if utils.Config.Frontend.LegalDir != "" {
		templatePath = path.Join(utils.Config.Frontend.LegalDir, "index.html")
	}

	imprintTemplate := templates.GetTemplate("layout.html", templatePath)

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "imprint", "/imprint", "Imprint")
	data.HeaderAd = true

	err := imprintTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}
