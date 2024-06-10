package handlers

import (
	"encoding/json"
	"eth2-exporter/price"
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/utils"
	"net/http"
)

func Burn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	templateFiles := append(layoutTemplateFiles, "burn.html")
	data := InitPageData(w, r, "burn", "/burn", "Eth Burned", templateFiles)

	var burnTemplate = templates.GetTemplate(templateFiles...)

	// data.Meta.Tdata1 = utils.FormatAmount((data.Data.(*types.BurnPageData).TotalBurned / 1e18) * data.Data.(*types.BurnPageData).Price)
	// data.Meta.Tdata2 = utils.FormatAmount(data.Data.(*types.BurnPageData).BurnRate24h/1e18) + " ETH/min"
	// data.Meta.Description = "The current ethereum burn rate is " + data.Meta.Tdata2 + ". A total of " + utils.FormatUSD(data.Data.(*types.BurnPageData).TotalBurned/1e18) + "ETH with a market value of $" + data.Meta.Tdata1 + " has been burned. " + data.Meta.Description

	latestBurn := services.LatestBurnData()

	currency := GetCurrency(r)

	if currency == utils.Config.Frontend.ElCurrency {
		currency = "USD"
	}

	latestBurn.Price = price.GetPrice(utils.Config.Frontend.ElCurrency, currency)
	latestBurn.Currency = currency

	data.Data = latestBurn
	if handleTemplateError(w, r, "burn.go", "Burn", "", burnTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

func BurnPageData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	latestBurn := services.LatestBurnData()

	currency := GetCurrency(r)

	if currency == "ETH" {
		currency = "USD"
	}

	latestBurn.Price = price.GetPrice(utils.Config.Frontend.ElCurrency, currency)
	latestBurn.Currency = currency

	err := json.NewEncoder(w).Encode(latestBurn)
	if err != nil {
		logger.Errorf("error sending latest burn page data: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
