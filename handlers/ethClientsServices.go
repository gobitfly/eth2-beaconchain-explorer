package handlers

import (
	ethclients "eth2-exporter/ethClients"
	"eth2-exporter/utils"
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"
)

var ethClientsServicesTemplate = template.Must(template.New("ethClientsServices").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/ethClientsServices.html"))

func EthClientsServices(w http.ResponseWriter, r *http.Request) {
	var err error

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "services", "/ethClientsServices", "Ethereum Clients Services Overview")

	pageData := ethclients.GetEthClientData()
	pageData.CsrfField = csrf.TemplateField(r)
	// pageData.Banner = ethclients.GetBannerClients()

	data.Data = pageData

	err = ethClientsServicesTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
