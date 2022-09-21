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

	err := faqTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}
