package handlers

import (
	"eth2-exporter/types"
	"html/template"
	"net/http"
)

var faqTemplate = template.Must(template.ParseFiles("templates/layout.html", "templates/faq.html"))

func Faq(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       "FAQ - beaconcha.in",
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/faq",
		},
		Active: "faq",
		Data:   nil,
	}

	err := faqTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Fatalf("Error executing template for %v route: %v", r.URL.String(),  err)
	}
}
