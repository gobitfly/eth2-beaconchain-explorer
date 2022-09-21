package handlers

import (
	"eth2-exporter/templates"
	"eth2-exporter/utils"
	"html/template"
	"net/http"
)

var faqTemplate = template.Must(template.New("faq").Funcs(utils.GetTemplateFuncs()).ParseFS(templates.Files, "layout.html", "faq.html"))

// Faq will return the data from the frequently asked questions (FAQ) using a go template
func Faq(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "faq", "/faq", "FAQ")
	data.HeaderAd = true

	err := faqTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}
