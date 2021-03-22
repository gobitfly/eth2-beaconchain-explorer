package handlers

import (
	// "eth2-exporter/db"

	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	types "eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"sync"
	"time"
	// "strings"
)

var poolsServicesTemplate = template.Must(template.New("poolsServices").Funcs(utils.GetTemplateFuncs()).ParseFiles(
	"templates/layout.html",
	"templates/poolsServices.html",
	"templates/bannerPoolsServices.html",
	"templates/index/depositDistribution.html"))

type pools struct {
	Address  string  `db:"address" json:"address"`
	Name     string  `db:"name" json:"name"`
	Deposit  int64   `db:"deposit" json:"deposit"`
	Category *string `db:"category" json:"category"`
}

type poolInfo struct {
	Status         string `db:"status" json:"status"`
	ValidatorIndex uint64 `db:"validatorindex" json:"validatorindex"`
	Balance31d     uint64 `db:"balance31d" json:"balance31d"`
}

type poolStatsData struct {
	PoolInfo []poolInfo
	Address  string
}

type chart struct {
	DepositDistribution types.ChartsPageDataChart
	StakedEther         string
	PoolInfo            []respData
	EthSupply           interface{}
}

type respData struct {
	Address    string                   `json:"address"`
	Name       string                   `json:"name"`
	Deposit    int64                    `json:"deposit"`
	Category   *string                  `json:"category"`
	PoolInfo   []poolInfo               `json:"poolInfo"`
	PoolIncome *types.ValidatorEarnings `json:"poolIncome"`
}

var poolInfoTemp []respData
var poolInfoTempTime time.Time
var ethSupply interface{}
var updateMux = &sync.RWMutex{}

func Pools(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "services", "/pools", "Stacking Pools Services Overview")

	chartData, err := services.ChartHandlers["deposits_distribution"].DataFunc()
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}

	var pieChart chart

	indexStats := services.LatestIndexPageData()

	pieChart.DepositDistribution.Data = chartData
	pieChart.DepositDistribution.Height = 500
	pieChart.DepositDistribution.Path = "deposits_distribution"
	pieChart.StakedEther = indexStats.StakedEther

	updateMux.Lock()
	defer updateMux.Unlock()
	if time.Now().Sub(poolInfoTempTime).Hours() > 1 { // query db every 1 hour
		poolInfoTemp = getPoolInfo()
		ethSupply = getEthSupply()
		pieChart.PoolInfo = poolInfoTemp
		pieChart.EthSupply = ethSupply
		poolInfoTempTime = time.Now()
	} else {
		pieChart.PoolInfo = poolInfoTemp
		pieChart.EthSupply = ethSupply
	}

	data.Data = pieChart

	err = poolsServicesTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

func getPoolInfo() []respData {
	var resp []respData

	var stakePools []pools
	err := db.DB.Select(&stakePools, "select address, name, deposit, category from stake_pools_stats;")
	if err != nil {
		logger.Errorf("error retrieving stake pools stats %v ", err)
	}

	stats := getPoolStats(stakePools)

	for i, pool := range stakePools {
		state := []poolInfo{}
		if len(stats) > i {
			if pool.Address == stats[i].Address {
				state = stats[i].PoolInfo
				// get income
				income, err := getPoolIncome(state)
				if err != nil {
					income = &types.ValidatorEarnings{}
				}
				resp = append(resp, respData{
					Address:    pool.Address,
					Category:   pool.Category,
					Deposit:    pool.Deposit,
					Name:       pool.Name,
					PoolInfo:   state,
					PoolIncome: income,
				})
			}
		}
	}

	return resp
}

func getPoolStats(pools []pools) []poolStatsData {
	var result []poolStatsData
	for _, pool := range pools {
		var states []poolInfo
		err := db.DB.Select(&states,
			`SELECT status, validatorindex, balance31d
			 FROM validators 
			 WHERE pubkey = ANY(
								SELECT publickey 
								FROM eth1_deposits 
								WHERE ENCODE(from_address::bytea, 'hex') LIKE LOWER($1)
							)
			 ORDER BY balance31d DESC`, pool.Address)
		if err != nil {
			logger.Errorf("error encoding:'%s', %v", pool.Address, err)
			continue
		}
		result = append(result, poolStatsData{PoolInfo: states, Address: pool.Address})
	}

	return result
}

func getPoolIncome(pools []poolInfo) (*types.ValidatorEarnings, error) {
	var indexes = make([]uint64, len(pools))
	for i, pools := range pools {
		indexes[i] = pools.ValidatorIndex
	}

	return GetValidatorEarnings(indexes)
}

func getEthSupply() interface{} {
	var respjson interface{}
	resp, err := http.Get("https://api.etherscan.io/api?module=stats&action=ethsupply&apikey=")

	if err != nil {
		logger.Errorf("error retrieving ETH Supply Data: %v", err)
		return nil
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&respjson)

	if err != nil {
		logger.Errorf("error decoding ETH Supply json response to interface: %v", err)
		return nil
	}

	return respjson
}

func GetSumLongestStreak(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	pool := strings.Replace(q.Get("pool"), "0x", "", -1)
	if len(pool) > 128 {
		pool = pool[:128]
	}

	// var sqlData []struct {
	// 	Long    *string
	// 	Current *string
	// }
	var sqlData []string

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
						validatorindex, start, length, rank() over (order by length desc),
						rank() over (partition by validatorindex order by length desc) as vrank
					from validator_attestation_streaks
					where status = 1
				)
			select  
				SUM(ls.length)
			from longeststreaks ls
			inner join matched_validators v on ls.validatorindex = v.validatorindex
			left join (select count(*) from longeststreaks) cnt(totalcount) on true
			where vrank = 1
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
