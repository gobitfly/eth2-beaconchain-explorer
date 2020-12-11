package handlers

import (
	"eth2-exporter/utils"
	"html/template"
	"net/http"
)

var ethClientsServicesTemplate = template.Must(template.New("ethClientsServices").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/ethClientsServices.html"))

func EthClientsServices(w http.ResponseWriter, r *http.Request) {
	var err error

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "ethClientsServices", "/ethClientsServices", "Ethereum Clients Services Overview")

	err = ethClientsServicesTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
