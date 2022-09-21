package handlers

import (
	"eth2-exporter/utils"
	"html/template"
	"net/http"
	"path"
)

// Imprint will show the imprint data using a go template
func Imprint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	var imprintTemplate *template.Template
	var err error

	if utils.Config.Frontend.LegalDir == "" {
		imprintTemplate, err = template.New("imprint").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", utils.Config.Frontend.Imprint)
	} else {
		imprintTemplate, err = template.New("imprint").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", path.Join(utils.Config.Frontend.LegalDir, "index.html"))
	}

	if err != nil {
		logger.Errorf("error parsing imprint page template: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	data := InitPageData(w, r, "imprint", "/imprint", "Imprint")
	data.HeaderAd = true

	err = imprintTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}
