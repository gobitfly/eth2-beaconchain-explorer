package handlers

import (
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"net/http"
)

func Relays(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	var relaysServicesTemplate = templates.GetTemplate("layout.html", "relays.html")

	data := InitPageData(w, r, "services", "/relays", "Relay Overview", "relays.html")

	relayData := services.LatestRelaysPageData()

	data.Data = relayData

	if handleTemplateError(w, r, "relays.go", "Relays", "", relaysServicesTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}
