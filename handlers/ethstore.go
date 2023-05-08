package handlers

import (
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"net/http"
)

func EthStore(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "ethstore.html", "svg/barChart.html")
	ethStoreTemplate := templates.GetTemplate(templateFiles...)
	w.Header().Set("Content-Type", "text/html")
	data := InitPageData(w, r, "services", "/ethstore", "ETH.STORE Statistics", templateFiles)
	data.Data = ethStoreData()

	if handleTemplateError(w, r, "ethstore.go", "EthStore", "", ethStoreTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

func ethStoreData() *types.EthStoreData {
	data := &types.EthStoreData{}
	data.EthStoreStatistics = services.LatestEthStoreStatistics()
	data.Disclaimer = services.EthStoreDisclaimer()

	return data
}
