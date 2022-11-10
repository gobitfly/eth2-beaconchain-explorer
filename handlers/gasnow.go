package handlers

import (
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

// Blocks will return information about blocks using a go template
func GasNow(w http.ResponseWriter, r *http.Request) {
	var gasNowTemplate = template.Must(template.New("gasnow").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/gasnow.html", "templates/components.html"))

	w.Header().Set("Content-Type", "text/html")

	// data := &types.PageData{
	// 	HeaderAd: true,
	// 	Meta: &types.Meta{
	// 		Title:       fmt.Sprintf("%v - GasNow - etherchain.org - %v", utils.Config.Frontend.SiteName, time.Now().Year()),
	// 		Description: "etherchain.org makes the Ethereum block chain accessible to non-technical end users",
	// 		Path:        "/tools/gasnow",
	// 		GATag:       utils.Config.Frontend.GATag,
	// 	},
	// 	Active:       "tools",
	// 	Data:         nil,
	// 	CurrentBlock: services.LatestBlock(),
	// 	GPO:          services.LatestGasNowData(),
	// }

	data := InitPageData(w, r, "gasnow", "/gasnow", fmt.Sprintf("fast: %v | Gas Price Oracle", "33 gwei"))

	res := []*struct {
		Ts      time.Time
		AvgFast float64
	}{}

	resRet := []*struct {
		Ts      int64   `json:"ts"`
		AvgFast float64 `json:"fast"`
	}{}

	for _, r := range res {
		resRet = append(resRet, &struct {
			Ts      int64   `json:"ts"`
			AvgFast float64 `json:"fast"`
		}{
			Ts:      r.Ts.Unix(),
			AvgFast: r.AvgFast,
		})
	}
	data.Data = resRet

	err := gasNowTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		return
	}
}
