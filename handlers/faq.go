package handlers

import (
	"eth2-exporter/templates"
	"net/http"
)

// Faq will return the data from the frequently asked questions (FAQ) using a go template
func Faq(w http.ResponseWriter, r *http.Request) {
	var faqTemplate = templates.GetTemplate("layout.html", "faq.html")

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "faq", "/faq", "FAQ")
	data.HeaderAd = true

	if handleTemplateError(w, r, faqTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}
