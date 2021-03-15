package handlers

import (
	// "eth2-exporter/db"

	"eth2-exporter/db"
	"eth2-exporter/services"
	types "eth2-exporter/types"
	"eth2-exporter/utils"
	"html/template"
	"net/http"
	"sync"
	"time"
	// "strings"
)

var poolsServicesTemplate = template.Must(template.New("poolsServices").Funcs(utils.GetTemplateFuncs()).ParseFiles(
	"templates/layout.html",
	"templates/poolsServices.html",
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
}

type respData struct {
	Address  string     `json:"address"`
	Name     string     `json:"name"`
	Deposit  int64      `json:"deposit"`
	Category *string    `json:"category"`
	PoolInfo []poolInfo `json:"poolInfo"`
}

var poolInfoTemp []respData
var poolInfoTempTime time.Time
var poolInfoTempMux = &sync.RWMutex{}

func Pools(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "services", "/pools", "Stacking Pools Services Overview")

	chartData, err := services.ChartHandlers["deposits_distribution"].DataFunc()
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
	}

	var pieChart chart

	indexStats := services.LatestIndexPageData()

	pieChart.DepositDistribution.Data = chartData
	pieChart.DepositDistribution.Height = 500
	pieChart.DepositDistribution.Path = "deposits_distribution"
	pieChart.StakedEther = indexStats.StakedEther

	poolInfoTempMux.Lock()
	defer poolInfoTempMux.Unlock()
	if time.Now().Sub(poolInfoTempTime).Minutes() > 5 { // query db every 5 min
		poolInfoTemp = getPoolInfo()
		pieChart.PoolInfo = poolInfoTemp
		poolInfoTempTime = time.Now()
	} else {
		pieChart.PoolInfo = poolInfoTemp
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
			}
		}

		resp = append(resp, respData{
			Address:  pool.Address,
			Category: pool.Category,
			Deposit:  pool.Deposit,
			Name:     pool.Name,
			PoolInfo: state,
		})
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
