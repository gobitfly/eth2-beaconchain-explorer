package handlers

import (
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"net/http"
)

func Relays(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	templateFiles := append(layoutTemplateFiles, "relays.html")
	var relaysServicesTemplate = templates.GetTemplate(templateFiles...)

	data := InitPageData(w, r, "services", "/relays", "Relay Overview", templateFiles)

	relayData := services.LatestRelaysPageData()

	data.Data = relayData

	if handleTemplateError(w, r, "relays.go", "Relays", "", relaysServicesTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}
