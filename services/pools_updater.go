package services

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/metrics"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/lib/pq"
)

type Pools struct {
	Address  string  `db:"address" json:"address"`
	Name     string  `db:"name" json:"name"`
	Deposit  int64   `db:"deposit" json:"deposit"`
	Category *string `db:"category" json:"category"`
	ValCount int64   `db:"vcount"`
}

type PoolStatsData struct {
	Status          string `db:"status" json:"status"`
	ValidatorIndex  uint64 `db:"validatorindex" json:"validatorindex"`
	Balance31d      uint64 `db:"balance31d" json:"balance31d"`
	Activationepoch uint64 `db:"activationepoch" json:"activationepoch"`
	Exitepoch       uint64 `db:"exitepoch" json:"exitepoch"`
}

type PoolStats struct {
	PoolInfo []PoolStatsData
	Address  string
}

type idEthSeries struct {
	Name   string       `json:"name"`
	Data   [][2]float64 `json:"data"`
	Marker struct {
		Enabled bool `json:"enabled"`
	} `json:"marker"`
}

type PoolsResp struct {
	DepositDistribution types.ChartsPageDataChart
	StakedEther         string
	PoolInfo            []PoolsInfo
	EthSupply           ethPriceResp
	LastUpdate          int64
	IdEthSeries         idEthSeriesDrill
	TotalValidators     uint64
	IsMainnet           bool
	NoAds               bool
}

type PoolsInfo struct {
	Address    string                   `json:"address"`
	Name       string                   `json:"name"`
	Deposit    int64                    `json:"deposit"`
	Category   *string                  `json:"category"`
	PoolInfo   []PoolStatsData          `json:"poolInfo"`
	PoolIncome *types.ValidatorEarnings `json:"poolIncome"`
}

type idEthSeriesDrill struct {
	MainSeries  []idEthSeries `json:"mainSeries"`
	DrillSeries []idEthSeries `json:"drillSeries"`
}

type ethPriceResp struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}

var poolInfoTemp atomic.Value     //[]PoolsInfo
var poolInfoTempTime atomic.Value //time.Time
var ethSupply atomic.Value        //interface{}
var idEthSeriesTemp atomic.Value  //= idEthSeriesDrill{}

func updatePoolInfo() {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("service_pools_updater").Observe(time.Since(start).Seconds())
	}()

	lastUpdateTime := poolInfoTempTime.Load().(time.Time)

	if time.Now().Sub(lastUpdateTime).Hours() > 3 { // query db every 3 hour
		// deleteOldChartEntries()
		poolInfoTempLocal := getPoolInfo()
		ethSupplyLocal := getEthSupply()
		idEthSeriesTempLocal := getIDEthChartSeries()

		poolInfoTemp.Store(poolInfoTempLocal)
		if ethSupplyLocal != nil {
			ethSupply.Store(*ethSupplyLocal)
		}
		poolInfoTempTime.Store(time.Now())

		idEthSeriesTemp.Store(idEthSeriesTempLocal)

		logger.Infoln("Updated Pool Info")
	}

}

func InitPools() {
	// updatePoolInfo()
	poolInfoTemp.Store([]PoolsInfo{})
	poolInfoTempTime.Store(time.Time{})
	ethSupply.Store(ethPriceResp{})
	idEthSeriesTemp.Store(idEthSeriesDrill{})
	go func() {
		for true {
			updatePoolInfo()
			time.Sleep(time.Minute * 10)
		}
	}()
}

func GetPoolsData() ([]PoolsInfo, ethPriceResp, int64) {
	// updateMux.Lock()
	// defer updateMux.Unlock()
	unix := poolInfoTempTime.Load().(time.Time).Unix()
	return poolInfoTemp.Load().([]PoolsInfo), ethSupply.Load().(ethPriceResp), unix
}

func GetIncomePerDepositedETHChart() idEthSeriesDrill {
	// idEthMux.Lock()
	// defer idEthMux.Unlock()
	return idEthSeriesTemp.Load().(idEthSeriesDrill)
}

func getPoolInfo() []PoolsInfo {
	var resp []PoolsInfo

	var stakePools []Pools
	// addrName := map[string]Pools{}

	if utils.Config.Chain.Network == "mainnet" || utils.Config.Chain.Network == "prater" {
		err := db.DB.Select(&stakePools, `
		select sps.address, sps.name, sps.category, sps.deposit, b.vcount
		from (select ENCODE(from_address::bytea, 'hex') as address, count(*) as vcount
			from (
				select publickey, from_address
				from eth1_deposits
				where valid_signature = true
				group by publickey, from_address
				having sum(amount) >= 32e9
			) a 
			group by from_address) b
		inner join stake_pools_stats as sps on b.address=sps.address
		order by vcount desc 
		`)
		if err != nil {
			logger.Errorf("error getting eth1-deposits-distribution for stake pools mainnet: %w", err)
		}
	} else {
		err := db.DB.Select(&stakePools, `
			select ENCODE(from_address::bytea, 'hex') as address, count(*) as vcount
			from (
				select publickey, from_address
				from eth1_deposits
				where valid_signature = true
				group by publickey, from_address
				having sum(amount) >= 32e9
			) a
			group by from_address 
			order by vcount desc limit 100`) // total at this point is 7k+, the limit is important
		if err != nil {
			logger.Errorf("error getting eth1-deposits-distribution for stake pools: %w", err)
		}
	}

	loopstart := time.Now()
	// logger.Errorln("pool stats", loopstart)

	for _, pool := range stakePools {
		// li := time.Now()
		var stats []PoolStatsData
		err := db.DB.Select(&stats,
			`SELECT status, validatorindex, balance31d, activationepoch, exitepoch
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
		// st := time.Now().Sub(li).Seconds()
		if len(stats) > 0 {
			income, err := getPoolIncome(pool.Address)
			// logger.Errorf("\n %s\nst %f\ngp %f\n", pName, st, time.Now().Sub(li).Seconds()-st)
			if err != nil {
				income = &types.ValidatorEarnings{}
			}
			resp = append(resp, PoolsInfo{
				Address:    pool.Address,
				Category:   pool.Category,
				Deposit:    pool.Deposit,
				Name:       pool.Name,
				PoolInfo:   stats,
				PoolIncome: income,
			})
		}
	}
	logger.Infof("pool update for loop took %f seconds", time.Now().Sub(loopstart).Seconds())
	return resp
}

func getPoolIncome(poolAddress string) (*types.ValidatorEarnings, error) {
	var indexes []uint64
	err := db.DB.Select(&indexes,
		`SELECT validatorindex
		 FROM validators 
		 WHERE pubkey = ANY(
							SELECT publickey 
							FROM eth1_deposits 
							WHERE ENCODE(from_address::bytea, 'hex') LIKE LOWER($1)
						)`, poolAddress)
	if err != nil {
		logger.Errorf("error selecting validator indexes:'%s', %v", poolAddress, err)
	}

	return getValidatorEarnings(indexes)
}

func getValidatorEarnings(validators []uint64) (*types.ValidatorEarnings, error) {
	validatorsPQArray := pq.Array(validators)
	latestEpoch := int64(LatestEpoch())
	lastDayEpoch := latestEpoch - 225
	lastWeekEpoch := latestEpoch - 225*7
	lastMonthEpoch := latestEpoch - 225*31
	twoWeeksBeforeEpoch := latestEpoch - 255*14
	threeWeeksBeforeEpoch := latestEpoch - 255*21

	if lastDayEpoch < 0 {
		lastDayEpoch = 0
	}
	if lastWeekEpoch < 0 {
		lastWeekEpoch = 0
	}
	if lastMonthEpoch < 0 {
		lastMonthEpoch = 0
	}
	if twoWeeksBeforeEpoch < 0 {
		twoWeeksBeforeEpoch = 0
	}
	if threeWeeksBeforeEpoch < 0 {
		threeWeeksBeforeEpoch = 0
	}

	balances := []*types.Validator{}

	err := db.DB.Select(&balances, `SELECT 
			   validatorindex,
			   COALESCE(balance, 0) AS balance, 
			   COALESCE(balanceactivation, 0) AS balanceactivation, 
			   COALESCE(balance1d, 0) AS balance1d, 
			   COALESCE(balance7d, 0) AS balance7d, 
			   COALESCE(balance31d , 0) AS balance31d,
       			activationepoch,
       			pubkey,
				status
		FROM validators WHERE validatorindex = ANY($1)`, validatorsPQArray)
	if err != nil {
		logger.Error("error selecting balances from validators: %v", err)
		return nil, err
	}

	deposits := []struct {
		Epoch     int64
		Amount    int64
		Publickey []byte
	}{}

	err = db.DB.Select(&deposits, `
	SELECT block_slot / 32 AS epoch, amount, publickey 
	FROM blocks_deposits 
	WHERE publickey IN (
		SELECT pubkey 
		FROM validators 
		WHERE validatorindex = ANY($1)
	)`, validatorsPQArray)
	if err != nil {
		logger.Error("error selecting deposits from blocks_deposits: %v", err)
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
	var earningsInPeriod int64
	var earningsInPeriodBalance int64

	for _, balance := range balances {

		if int64(balance.ActivationEpoch) > latestEpoch {
			continue
		}

		for epoch, deposit := range depositsMap[fmt.Sprintf("%x", balance.PublicKey)] {
			totalDeposits += deposit

			if epoch >= threeWeeksBeforeEpoch && epoch <= lastWeekEpoch &&
				epoch > int64(balance.ActivationEpoch) {
				earningsInPeriod -= deposit
			}

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

		if int64(balance.ActivationEpoch) <= lastMonthEpoch && balance.Status == "active_online" {
			earningsInPeriod += (int64(balance.Balance) - int64(balance.Balance31d)) - (int64(balance.Balance) - int64(balance.Balance7d))
			earningsInPeriodBalance += int64(balance.BalanceActivation)
		}
	}

	return &types.ValidatorEarnings{
		Total:                   earningsTotal,
		LastDay:                 earningsLastDay,
		LastWeek:                earningsLastWeek,
		LastMonth:               earningsLastMonth,
		TotalDeposits:           totalDeposits,
		EarningsInPeriodBalance: earningsInPeriodBalance,
		EarningsInPeriod:        earningsInPeriod,
		EpochStart:              lastMonthEpoch,
		EpochEnd:                lastWeekEpoch,
	}, nil
}

func getIDEthChartSeries() idEthSeriesDrill {
	epoch := int64(LatestEpoch())
	lastMonthEpoch := epoch - 225*31
	if lastMonthEpoch < 0 {
		lastMonthEpoch = 0
	}
	type idEthSeriesData struct {
		Epoch   int64
		Name    string
		Income  int64
		Balance int64
	}

	dbData := []idEthSeriesData{}
	err := db.DB.Select(&dbData, `SELECT 
			   epoch,
			   name, 
			   income,
			   (CASE balance WHEN 0 THEN 1 ELSE balance END) as balance
		FROM staking_pools_chart WHERE epoch >= $1 order by epoch asc`, lastMonthEpoch)
	if err != nil {
		logger.Error("error selecting balances from validators: %v", err)
		return idEthSeriesDrill{}
	}

	seriesMap := map[string]idEthSeries{}
	for _, item := range dbData {
		elem, exist := seriesMap[item.Name]
		if !exist {
			seriesMap[item.Name] = idEthSeries{
				Name: item.Name,
				Data: [][2]float64{{float64(item.Epoch), float64(item.Income) / float64(item.Balance)}},
			}
			continue
		}
		elem.Data = append(elem.Data, [2]float64{float64(item.Epoch), float64(item.Income) / float64(item.Balance)})
		seriesMap[item.Name] = elem
	}

	mainSeriesSlice := []idEthSeries{}
	subSeriesSlice := []idEthSeries{}

	for key, item := range seriesMap {
		poolName := strings.Split(key, "-")
		if len(poolName) == 1 {
			mainSeriesSlice = append(mainSeriesSlice, item)
		} else if len(poolName) == 2 {
			i, err := strconv.ParseInt(strings.ReplaceAll(poolName[1], " ", ""), 10, 64)
			if err != nil || i == 1 {
				item.Name = poolName[0]
				mainSeriesSlice = append(mainSeriesSlice, item)
			}
		}

		subSeriesSlice = append(subSeriesSlice, item)
	}

	return idEthSeriesDrill{MainSeries: mainSeriesSlice, DrillSeries: subSeriesSlice}

}

func GetTotalValidators() uint64 {
	epoch := int64(LatestEpoch())
	var limit int64 = 0
	if epoch > 0 {
		limit = epoch - 1
	}
	var activeValidators uint64

	err := db.DB.Get(&activeValidators, `
			SELECT validatorscount 
			FROM epochs 
			WHERE epoch = $1`, limit)
	if err != nil {
		logger.Errorf("error retrieving staked ether history: %v", err)
	}

	return activeValidators
}

func getEthSupply() *ethPriceResp {
	var respjson *ethPriceResp
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
