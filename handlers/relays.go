package handlers

import (
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"net/http"
)

func Relays(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	var relaysServicesTemplate = templates.GetTemplate("layout.html", "relays.html")

	data := InitPageData(w, r, "services", "/relays", "MEV-Boost Relay Overview")

	relayData := services.LatestRelaysPageData()

	data.Data = relayData

	err := relaysServicesTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}
