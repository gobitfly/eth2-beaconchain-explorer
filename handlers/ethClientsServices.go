package handlers

import (
	ethclients "eth2-exporter/ethClients"
	"eth2-exporter/utils"
	"html/template"
	"net/http"
)

var ethClientsServicesTemplate = template.Must(template.New("ethClientsServices").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/ethClientsServices.html"))

func EthClientsServices(w http.ResponseWriter, r *http.Request) {
	var err error

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "services", "/ethClientsServices", "Ethereum Clients Services Overview")

	pageData := ethclients.GetEthClientData()
	// pageData.Banner = ethclients.GetBannerClients()

	if err != nil {
		logger.Errorf("error retrieving flashes for advertisewithusform %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	data.Data = pageData

	err = ethClientsServicesTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
