package handlers

import (
	"encoding/json"
	"eth2-exporter/services"
	"eth2-exporter/utils"
	"html/template"
	"net/http"
)

var burnTemplate = template.Must(template.New("burn").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/burn.html", "templates/components.html"))

func Burn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	data := InitPageData(w, r, "burn", "/burn", "Eth Burned")

	// data.Meta.Tdata1 = utils.FormatAmount((data.Data.(*types.BurnPageData).TotalBurned / 1e18) * data.Data.(*types.BurnPageData).Price)
	// data.Meta.Tdata2 = utils.FormatAmount(data.Data.(*types.BurnPageData).BurnRate24h/1e18) + " ETH/min"
	// data.Meta.Description = "The current ethereum burn rate is " + data.Meta.Tdata2 + ". A total of " + utils.FormatUSD(data.Data.(*types.BurnPageData).TotalBurned/1e18) + "ETH with a market value of $" + data.Meta.Tdata1 + " has been burned. " + data.Meta.Description

	err := burnTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		return
	}
}

func BurnPageData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(services.LatestBurnData())
	if err != nil {
		logger.Errorf("error sending latest burn page data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
