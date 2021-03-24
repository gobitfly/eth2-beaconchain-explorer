package services

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/metrics"
	"eth2-exporter/types"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/lib/pq"
)

type Pools struct {
	Address  string  `db:"address" json:"address"`
	Name     string  `db:"name" json:"name"`
	Deposit  int64   `db:"deposit" json:"deposit"`
	Category *string `db:"category" json:"category"`
}

type PoolInfo struct {
	Status         string `db:"status" json:"status"`
	ValidatorIndex uint64 `db:"validatorindex" json:"validatorindex"`
	Balance31d     uint64 `db:"balance31d" json:"balance31d"`
}

type PoolStatsData struct {
	PoolInfo []PoolInfo
	Address  string
}

type Chart struct {
	DepositDistribution types.ChartsPageDataChart
	StakedEther         string
	PoolInfo            []RespData
	EthSupply           interface{}
	LastUpdate          int64
}

type RespData struct {
	Address    string                   `json:"address"`
	Name       string                   `json:"name"`
	Deposit    int64                    `json:"deposit"`
	Category   *string                  `json:"category"`
	PoolInfo   []PoolInfo               `json:"poolInfo"`
	PoolIncome *types.ValidatorEarnings `json:"poolIncome"`
}

var poolInfoTemp []RespData
var poolInfoTempTime time.Time
var ethSupply interface{}
var updateMux = &sync.RWMutex{}
var firstReq = true

func updatePoolInfo() {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("service_pools_updater").Observe(time.Since(start).Seconds())
	}()
	updateMux.Lock()
	defer updateMux.Unlock()
	if time.Now().Sub(poolInfoTempTime).Hours() > 6 { // query db every 6 hour
		poolInfoTemp = getPoolInfo()
		ethSupply = getEthSupply()
		poolInfoTempTime = time.Now()
		logger.Infoln("Updated Pool Info")
	}

}

func InitPools() {
	updatePoolInfo()
	go func() {
		for true {
			time.Sleep(time.Minute * 10)
			updatePoolInfo()
		}
	}()
}

func GetPoolsData() ([]RespData, interface{}, int64) {
	updateMux.Lock()
	defer updateMux.Unlock()
	return poolInfoTemp, ethSupply, poolInfoTempTime.Unix()
}

func getPoolInfo() []RespData {
	var resp []RespData

	var stakePools []Pools
	err := db.DB.Select(&stakePools, "select address, name, deposit, category from stake_pools_stats;")
	if err != nil {
		logger.Errorf("error retrieving stake pools stats %v ", err)
	}
	// logger.Errorln("pool stats", time.Now())
	stats := getPoolStats(stakePools)
	// logger.Errorln("pool stats after", time.Now())
	for i, pool := range stakePools {
		state := []PoolInfo{}
		if len(stats) > i {
			if pool.Address == stats[i].Address {
				state = stats[i].PoolInfo
				// get income
				income, err := getPoolIncome(state)
				if err != nil {
					income = &types.ValidatorEarnings{}
				}
				resp = append(resp, RespData{
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

func getPoolStats(pools []Pools) []PoolStatsData {
	var result []PoolStatsData
	for _, pool := range pools { // needs optimisation takes 10 sec. to run
		var states []PoolInfo
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
		result = append(result, PoolStatsData{PoolInfo: states, Address: pool.Address})
	}

	return result
}

func getPoolIncome(pools []PoolInfo) (*types.ValidatorEarnings, error) {
	var indexes = make([]uint64, len(pools))
	for i, pool := range pools {
		indexes[i] = pool.ValidatorIndex
	}

	return getValidatorEarnings(indexes)
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

func getValidatorEarnings(validators []uint64) (*types.ValidatorEarnings, error) {
	validatorsPQArray := pq.Array(validators)
	latestEpoch := int64(LatestEpoch())
	lastDayEpoch := latestEpoch - 225
	lastWeekEpoch := latestEpoch - 225*7
	lastMonthEpoch := latestEpoch - 225*31

	if lastDayEpoch < 0 {
		lastDayEpoch = 0
	}
	if lastWeekEpoch < 0 {
		lastWeekEpoch = 0
	}
	if lastMonthEpoch < 0 {
		lastMonthEpoch = 0
	}

	balances := []*types.Validator{}

	err := db.DB.Select(&balances, `SELECT 
			   COALESCE(balance, 0) AS balance, 
			   COALESCE(balanceactivation, 0) AS balanceactivation, 
			   COALESCE(balance1d, 0) AS balance1d, 
			   COALESCE(balance7d, 0) AS balance7d, 
			   COALESCE(balance31d , 0) AS balance31d,
       			activationepoch,
       			pubkey
		FROM validators WHERE validatorindex = ANY($1)`, validatorsPQArray)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	deposits := []struct {
		Epoch     int64
		Amount    int64
		Publickey []byte
	}{}

	err = db.DB.Select(&deposits, "SELECT block_slot / 32 AS epoch, amount, publickey FROM blocks_deposits WHERE publickey IN (SELECT pubkey FROM validators WHERE validatorindex = ANY($1))", validatorsPQArray)
	if err != nil {
		return nil, err
	}

	depositsMap := make(map[string]map[int64]int64)
	for _, d := range deposits {
		if _, exists := depositsMap[fmt.Sprintf("%x", d.Publickey)]; !exists {
			depositsMap[fmt.Sprintf("%x", d.Publickey)] = make(map[int64]int64)
		}
		depositsMap[fmt.Sprintf("%x", d.Publickey)][d.Epoch] += d.Amount
	}

	var earningsTotal int64
	var earningsLastDay int64
	var earningsLastWeek int64
	var earningsLastMonth int64
	var totalDeposits int64

	for _, balance := range balances {

		if int64(balance.ActivationEpoch) > latestEpoch {
			continue
		}
		for epoch, deposit := range depositsMap[fmt.Sprintf("%x", balance.PublicKey)] {
			totalDeposits += deposit
			if epoch > int64(balance.ActivationEpoch) {
				earningsTotal -= deposit
			}
			if epoch > lastDayEpoch && epoch > int64(balance.ActivationEpoch) {
				earningsLastDay -= deposit
			}
			if epoch > lastWeekEpoch && epoch > int64(balance.ActivationEpoch) {
				earningsLastWeek -= deposit
			}
			if epoch > lastMonthEpoch && epoch > int64(balance.ActivationEpoch) {
				earningsLastMonth -= deposit
			}
		}

		if int64(balance.ActivationEpoch) > lastDayEpoch {
			balance.Balance1d = balance.BalanceActivation
		}
		if int64(balance.ActivationEpoch) > lastWeekEpoch {
			balance.Balance7d = balance.BalanceActivation
		}
		if int64(balance.ActivationEpoch) > lastMonthEpoch {
			balance.Balance31d = balance.BalanceActivation
		}
		earningsTotal += int64(balance.Balance) - int64(balance.BalanceActivation)
		earningsLastDay += int64(balance.Balance) - int64(balance.Balance1d)
		earningsLastWeek += int64(balance.Balance) - int64(balance.Balance7d)
		earningsLastMonth += int64(balance.Balance) - int64(balance.Balance31d)
	}

	return &types.ValidatorEarnings{
		Total:         earningsTotal,
		LastDay:       earningsLastDay,
		LastWeek:      earningsLastWeek,
		LastMonth:     earningsLastMonth,
		TotalDeposits: totalDeposits,
	}, nil
}
