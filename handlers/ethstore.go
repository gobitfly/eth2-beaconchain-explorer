package handlers

import (
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"net/http"
)

func EthStore(w http.ResponseWriter, r *http.Request) {
	ethStoreTemplate := templates.GetTemplate(
		"layout.html",
		"ethstore.html",
		"svg/barChart.html")
	w.Header().Set("Content-Type", "text/html")
	data := InitPageData(w, r, "services", "/ethstore", "ETH.STORE Statistics")
	data.Data = services.LatestEthStoreStatistics()

	err := ethStoreTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
