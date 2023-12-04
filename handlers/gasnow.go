package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/price"
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/utils"
	"fmt"
	"net/http"
	"sort"
	"time"
)

// Will return the gas now page
func GasNow(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "gasnow.html")
	var gasNowTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "gasnow", "/gasnow", fmt.Sprintf("%v Gwei", 34), templateFiles)

	now := time.Now().Truncate(time.Minute)
	lastWeek := time.Now().Truncate(time.Minute).Add(-utils.Week)

	history, err := db.BigtableClient.GetGasNowHistory(now, lastWeek)
	if err != nil {
		logger.Errorf("error retrieving gas price histors: %v", err)
		return
	}

	group := make(map[int64]float64, 0)
	for i := 0; i < len(history); i++ {
		_, ok := group[history[i].Ts.Truncate(time.Hour).Unix()]
		if !ok {
			group[history[i].Ts.Truncate(time.Hour).Unix()] = float64(history[i].Fast.Int64())
		} else {
			group[history[i].Ts.Truncate(time.Hour).Unix()] = (group[history[i].Ts.Truncate(time.Hour).Unix()] + float64(history[i].Fast.Int64())) / 2
		}
	}

	resRet := []*struct {
		Ts      int64   `json:"ts"`
		AvgFast float64 `json:"fast"`
	}{}

	for ts, fast := range group {
		resRet = append(resRet, &struct {
			Ts      int64   `json:"ts"`
			AvgFast float64 `json:"fast"`
		}{
			Ts:      ts,
			AvgFast: fast,
		})
	}

	sort.SliceStable(resRet, func(i int, j int) bool {
		return resRet[i].Ts > resRet[j].Ts
	})

	data.Data = resRet

	if handleTemplateError(w, r, "gasnow.go", "GasNow", "", gasNowTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

func GasNowData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	gasnowData := services.LatestGasNowData()
	if gasnowData == nil {
		utils.LogError(nil, "error obtaining latest gas now data 'nil'", 0)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	currency := GetCurrency(r)
	if currency == utils.Config.Frontend.ElCurrency {
		currency = "USD"
	}
	gasnowData.Data.Price = price.GetPrice(utils.Config.Frontend.ElCurrency, currency)
	gasnowData.Data.Currency = currency

	err := json.NewEncoder(w).Encode(gasnowData)
	if err != nil {
		logger.Errorf("error serializing json data for API %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
