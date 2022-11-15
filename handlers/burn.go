package handlers

import (
	"encoding/json"
	"etherchain-web-v2/services"
	"etherchain-web-v2/types"
	"etherchain-web-v2/utils"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

var burnTemplate = template.Must(template.New("burn").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/burn.html", "templates/components.html"))

func Burn(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html")

	data := &types.PageData{
		Meta: &types.Meta{
			Image:       "https://etherchain.org/img/burn-512x512.png",
			Title:       fmt.Sprintf("Ethereum (ETH) Blockchain Explorer - etherchain.org - %v", time.Now().Year()),
			Description: "etherchain.org makes the Ethereum block chain accessible to non-technical end users.",
			Path:        "",
			Tlabel1:     "Total Burned",
			Tlabel2:     "Burn Rate 24h",
		},
		ShowSyncingMessage: false,
		CurrentBlock:       services.LatestBlock(),
		GPO:                services.LatestGasNowData(),
		Active:             "statistics",
		Data:               services.LatestBurnPageData(),
		DepositContract:    utils.Config.Frontend.DepositContract,
	}
	data.Meta.Tdata1 = utils.FormatUSD((data.Data.(*types.BurnPageData).TotalBurned / 1e18) * data.Data.(*types.BurnPageData).Price)
	data.Meta.Tdata2 = utils.FormatUSD(data.Data.(*types.BurnPageData).BurnRate24h/1e18) + " ETH/min"
	data.Meta.Description = "The current ethereum burn rate is " + data.Meta.Tdata2 + ". A total of " + utils.FormatUSD(data.Data.(*types.BurnPageData).TotalBurned/1e18) + "ETH with a market value of $" + data.Meta.Tdata1 + " has been burned. " + data.Meta.Description

	err := burnTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		return
	}
}

func BurnPageData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(services.LatestBurnPageData())
	if err != nil {
		logger.Errorf("error sending latest burn page data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
