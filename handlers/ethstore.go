package handlers

import (
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"net/http"
)

func EthStore(w http.ResponseWriter, r *http.Request) {
	ethStoreTemplate := templates.GetTemplate(
		append(layoutTemplateFiles, []string{"ethstore.html",
			"svg/barChart.html"}...)...,
	)
	w.Header().Set("Content-Type", "text/html")
	data := InitPageData(w, r, "services", "/ethstore", "ETH.STORE Statistics")
	data.Data = services.LatestEthStoreStatistics()

	if handleTemplateError(w, r, "ethstore.go", "EthStore", "", ethStoreTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}
