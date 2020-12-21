package handlers

import (
	"html/template"
	"net/http"
)

var mobileTemplate = template.Must(template.ParseFiles("templates/layout.html", "templates/mobilepage.html"))

func MobilePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "more", "/mobile", "Beaconchain Dashboard")
	data.HeaderAd = true

	err := mobileTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
