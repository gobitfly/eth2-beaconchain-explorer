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
	"sync"
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
	Status         string `db:"status" json:"status"`
	ValidatorIndex uint64 `db:"validatorindex" json:"validatorindex"`
	Balance31d     uint64 `db:"balance31d" json:"balance31d"`
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
	EthSupply           interface{}
	LastUpdate          int64
	IdEthSeries         idEthSeriesDrill
	TotalValidators     uint64
	IsMainnet           bool
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

var poolInfoTemp []PoolsInfo
var poolInfoTempTime time.Time
var ethSupply interface{}
var updateMux = &sync.RWMutex{}
var firstReq = true
var idEthSeriesTemp = idEthSeriesDrill{}

func updatePoolInfo() {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("service_pools_updater").Observe(time.Since(start).Seconds())
	}()
	updateMux.Lock()
	lastUpdateTime := poolInfoTempTime
	updateMux.Unlock()

	if time.Now().Sub(lastUpdateTime).Hours() > 3 { // query db every 3 hour
		// deleteOldChartEntries()
		poolInfoTempLocal := getPoolInfo()
		ethSupplyLocal := getEthSupply()
		idEthSeriesTempLocal := getIDEthChartSeries()

		updateMux.Lock()
		defer updateMux.Unlock()
		poolInfoTemp = poolInfoTempLocal
		ethSupply = ethSupplyLocal
		idEthSeriesTemp = idEthSeriesTempLocal
		poolInfoTempTime = time.Now()
		logger.Infoln("Updated Pool Info")
	}

}

func InitPools() {
	// updatePoolInfo()
	go func() {
		for true {
			updatePoolInfo()
			time.Sleep(time.Minute * 10)
		}
	}()
}

func GetPoolsData() ([]PoolsInfo, interface{}, int64, idEthSeriesDrill) {
	updateMux.Lock()
	defer updateMux.Unlock()
	return poolInfoTemp, ethSupply, poolInfoTempTime.Unix(), idEthSeriesTemp
}

func getPoolInfo() []PoolsInfo {
	var resp []PoolsInfo

	var stakePools []Pools
	addrName := map[string]Pools{}

	if utils.Config.Chain.Network == "mainnet" {
		var stakePoolsNames []Pools
		err := db.DB.Select(&stakePoolsNames, "select address, name, deposit, category from stake_pools_stats;") // deposit is a placeholder the actual value is not used on frontend
		if err != nil {
			logger.Errorf("error retrieving stake pools stats %v ", err)
		}

		for _, pool := range stakePoolsNames {
			if _, exist := addrName[pool.Address]; !exist {
				addrName[pool.Address] = pool
			}
		}
	}

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

	loopstart := time.Now()
	// logger.Errorln("pool stats", loopstart)

	for _, pool := range stakePools {
		// li := time.Now()
		var stats []PoolStatsData
		err := db.DB.Select(&stats,
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
		// st := time.Now().Sub(li).Seconds()
		if len(stats) > 0 {
			pName := ""
			if utils.Config.Chain.Network == "mainnet" {
				nPool, exist := addrName[pool.Address]
				if exist {
					pName = nPool.Name
					pool.Name = nPool.Name
					pool.Category = nPool.Category
				}
			}
			income, err := getPoolIncome(pool.Address, pName)
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

func getPoolIncome(poolAddress string, poolName string) (*types.ValidatorEarnings, error) {
	// var indexes = make([]uint64, len(pool))
	// for i, validator := range pool {
	// 	indexes[i] = validator.ValidatorIndex
	// }
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

	return getValidatorEarnings(indexes, poolName)
}

func getValidatorEarnings(validators []uint64, poolName string) (*types.ValidatorEarnings, error) {
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

	// go func() {
	// 	updateChartDB(poolName, lastWeekEpoch, earningsInPeriod, earningsInPeriodBalance)
	// }()

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

// func updateChartDB(poolName string, epoch int64, income int64, balance int64) {
// 	if poolName == "" {
// 		return
// 	}
// 	_, err := db.DB.Exec(`
// 		INSERT INTO staking_pools_chart
// 		(epoch, name, income, balance)
// 		VALUES
// 		($1, $2, $3, $4)
// 	`, epoch, poolName, income, balance)
// 	if err != nil {
// 		logger.Errorf("error inserting staking pool chart data (if 'duplicate key' error not critical): %v", err)
// 	}
// }

// func deleteOldChartEntries() {
// 	latestEpoch := int64(LatestEpoch())
// 	sixMonthsOld := latestEpoch - 225*31*6
// 	_, err := db.DB.Exec(`
// 		DELETE FROM staking_pools_chart
// 		WHERE epoch <= $1
// 	`, sixMonthsOld)
// 	if err != nil {
// 		logger.Errorf("error removing old staking pool chart data: %v", err)
// 	}
// }

func getIDEthChartSeries() idEthSeriesDrill {

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
		FROM staking_pools_chart order by epoch asc`)
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
