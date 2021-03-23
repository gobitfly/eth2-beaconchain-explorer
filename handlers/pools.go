package handlers

import (
	// "eth2-exporter/db"

	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	// "strings"
)

var poolsServicesTemplate = template.Must(template.New("poolsServices").Funcs(utils.GetTemplateFuncs()).ParseFiles(
	"templates/layout.html",
	"templates/poolsServices.html",
	"templates/bannerPoolsServices.html",
	"templates/index/depositDistribution.html"))

func Pools(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "services", "/pools", "Stacking Pools Services Overview")

	chartData, err := services.ChartHandlers["deposits_distribution"].DataFunc()
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}

	var pieChart services.Chart

	indexStats := services.LatestIndexPageData()

	pieChart.DepositDistribution.Data = chartData
	pieChart.DepositDistribution.Height = 500
	pieChart.DepositDistribution.Path = "deposits_distribution"
	pieChart.StakedEther = indexStats.StakedEther

	pieChart.PoolInfo, pieChart.EthSupply, pieChart.LastUpdate = services.GetPoolsData()

	data.Data = pieChart

	err = poolsServicesTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

func GetAvgCurrentStreak(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	pool := strings.Replace(q.Get("pool"), "0x", "", -1)
	if len(pool) > 128 {
		pool = pool[:128]
	}

	var sqlData []*string

	err := db.DB.Select(&sqlData, `
			with 
				matched_validators as (
					SELECT v.validatorindex  
					FROM validators v 
					LEFT JOIN eth1_deposits e ON e.publickey = v.pubkey
					WHERE ENCODE(e.from_address::bytea, 'hex') LIKE LOWER($1)

				),
				longeststreaks as (
					select 
						validatorindex, start, length
					from validator_attestation_streaks
					where status = 1
				),
				currentstreaks as (
					select validatorindex, start, length
					from validator_attestation_streaks
					where status = 1 and start+length = (select max(start+length) from validator_attestation_streaks)
				)
			select  
				AVG(coalesce(cs.length,0))
			from longeststreaks ls
			inner join matched_validators v on ls.validatorindex = v.validatorindex
			left join currentstreaks cs on cs.validatorindex = ls.validatorindex
			`, pool)

	if err != nil {
		http.Error(w, fmt.Sprintf("Internal server error: %v", err), 503)
		return
	}

	err = json.NewEncoder(w).Encode(sqlData)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}

}
