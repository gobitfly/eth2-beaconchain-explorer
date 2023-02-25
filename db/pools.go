package db

import (
	"database/sql"
	"eth2-exporter/metrics"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
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
}

var lastUpdateTimePools time.Time

func UpdatePoolInfo() {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("service_pools_updater").Observe(time.Since(start).Seconds())
	}()

	if time.Since(lastUpdateTimePools).Hours() > 3 { // query db every 3 hour
		// tx, err := DB.Begin()
		// if err != nil {
		// 	logger.Errorf("error connecting to db %v", err)
		// 	return
		// }
		// defer tx.Rollback()

		deleteOldChartEntries()
		getPoolInfo()

		lastUpdateTimePools = time.Now()
		logger.Infoln("Updated Pool Info")
	}

}

func getPoolInfo() {
	var stakePools []Pools
	addrName := map[string]Pools{}

	if utils.Config.Chain.Config.ConfigName == "mainnet" || utils.Config.Chain.Config.ConfigName == "prater" {
		var stakePoolsNames []Pools
		err := ReaderDb.Select(&stakePoolsNames, "select address, name, deposit, category from stake_pools_stats;") // deposit is a placeholder the actual value is not used on frontend
		if err != nil {
			logger.Errorf("error retrieving stake pools stats %v ", err)
		}

		for _, pool := range stakePoolsNames {
			if _, exist := addrName[pool.Address]; !exist {
				addrName[pool.Address] = pool
			}
		}
	}

	err := ReaderDb.Select(&stakePools, `
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
		logger.Errorf("error getting eth1-deposits-distribution for stake pools: %v", err)
	}

	loopstart := time.Now()
	// logger.Errorln("pool stats", loopstart)

	for _, pool := range stakePools {
		// li := time.Now()
		var stats []PoolStatsData
		err := ReaderDb.Select(&stats,
			`SELECT status, validatorindex
			 FROM validators 
			 WHERE pubkey = ANY(
								SELECT publickey 
								FROM eth1_deposits 
								WHERE ENCODE(from_address::bytea, 'hex') LIKE LOWER($1)
							)`, pool.Address)
		if err != nil {
			logger.Errorf("error encoding:'%s', %v", pool.Address, err)
			continue
		}
		// st := time.Now().Sub(li).Seconds()
		if len(stats) > 0 {
			pName := ""
			if utils.Config.Chain.Config.ConfigName == "mainnet" || utils.Config.Chain.Config.ConfigName == "prater" {
				nPool, exist := addrName[pool.Address]
				if exist {
					pName = nPool.Name
					pool.Name = nPool.Name
					pool.Category = nPool.Category
				}
			}
			getPoolIncome(pool.Address, pName)
			// logger.Errorf("\n %s\nst %f\ngp %f\n", pName, st, time.Now().Sub(li).Seconds()-st)

		}
	}
	logger.Infof("pool update for loop took %f seconds", time.Since(loopstart).Seconds())
}

func getPoolIncome(poolAddress string, poolName string) {
	// var indexes = make([]uint64, len(pool))
	// for i, validator := range pool {
	// 	indexes[i] = validator.ValidatorIndex
	// }
	var indexes []uint64
	err := ReaderDb.Select(&indexes,
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

	getValidatorEarnings(indexes, poolName)
}

func getValidatorEarnings(validators []uint64, poolName string) {
	validatorsPQArray := pq.Array(validators)

	latestEpochUint, err := GetLatestEpoch()
	if err != nil {
		logger.Errorf("error getting latest epoch %v", err)
		return
	}

	latestEpoch := int64(latestEpochUint)
	lastDayEpoch := latestEpoch - int64(utils.EpochsPerDay())
	lastWeekEpoch := latestEpoch - int64(utils.EpochsPerDay())*7
	lastMonthEpoch := latestEpoch - int64(utils.EpochsPerDay())*31
	twoWeeksBeforeEpoch := latestEpoch - int64(utils.EpochsPerDay())*14
	threeWeeksBeforeEpoch := latestEpoch - int64(utils.EpochsPerDay())*21

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

	err = ReaderDb.Select(&balances, `SELECT 
			    validatorindex,
			    COALESCE(balanceactivation, 0) AS balanceactivation, 
       			activationepoch,
       			pubkey,
				status
		FROM validators WHERE validatorindex = ANY($1)`, validatorsPQArray)
	if err != nil {
		logger.Errorf("error selecting balances from validators: %v", err)
		return
	}

	balancesMap := make(map[uint64]*types.Validator, len(balances))

	for _, balance := range balances {
		balancesMap[balance.Index] = balance
	}

	latestBalances, err := BigtableClient.GetValidatorBalanceHistory(validators, uint64(latestEpoch), uint64(latestEpoch))
	if err != nil {
		logger.Errorf("error getting validator balance data in getValidatorEarnings: %v", err)
		return
	}
	for balanceIndex, balance := range latestBalances {
		if len(balance) == 0 || balancesMap[balanceIndex] == nil {
			continue
		}
		balancesMap[balanceIndex].Balance = balance[0].Balance
	}

	balances1d, err := BigtableClient.GetValidatorBalanceHistory(validators, uint64(lastDayEpoch), uint64(lastDayEpoch))
	if err != nil {
		logger.Errorf("error getting validator Balance1d data in getValidatorEarnings: %v", err)
		return
	}
	for balanceIndex, balance := range balances1d {
		if len(balance) == 0 || balancesMap[balanceIndex] == nil {
			continue
		}
		balancesMap[balanceIndex].Balance1d = sql.NullInt64{
			Int64: int64(balance[0].Balance),
			Valid: true,
		}
	}

	balances7d, err := BigtableClient.GetValidatorBalanceHistory(validators, uint64(lastWeekEpoch), uint64(lastWeekEpoch))
	if err != nil {
		logger.Errorf("error getting validator Balance7d data in getValidatorEarnings: %v", err)
		return
	}
	for balanceIndex, balance := range balances7d {
		if len(balance) == 0 || balancesMap[balanceIndex] == nil {
			continue
		}
		balancesMap[balanceIndex].Balance7d = sql.NullInt64{
			Int64: int64(balance[0].Balance),
			Valid: true,
		}
	}

	balances31d, err := BigtableClient.GetValidatorBalanceHistory(validators, uint64(lastMonthEpoch), uint64(lastMonthEpoch))
	if err != nil {
		logger.Errorf("error getting validator Balance31d data in getValidatorEarnings: %v", err)
		return
	}
	for balanceIndex, balance := range balances31d {
		if len(balance) == 0 || balancesMap[balanceIndex] == nil {
			continue
		}
		balancesMap[balanceIndex].Balance31d = sql.NullInt64{
			Int64: int64(balance[0].Balance),
			Valid: true,
		}
	}

	deposits := []struct {
		Epoch     int64
		Amount    int64
		Publickey []byte
	}{}

	err = ReaderDb.Select(&deposits, `
	SELECT block_slot / $2 AS epoch, amount, publickey 
	FROM blocks_deposits 
	WHERE publickey IN (
		SELECT pubkey 
		FROM validators 
		WHERE validatorindex = ANY($1)
	)`, validatorsPQArray, utils.Config.Chain.Config.SlotsPerEpoch)
	if err != nil {
		logger.Errorf("error selecting deposits from blocks_deposits: %v", err)
		return
	}

	depositsMap := make(map[string]map[int64]int64)
	for _, d := range deposits {
		if _, exists := depositsMap[fmt.Sprintf("%x", d.Publickey)]; !exists {
			depositsMap[fmt.Sprintf("%x", d.Publickey)] = make(map[int64]int64)
		}
		depositsMap[fmt.Sprintf("%x", d.Publickey)][d.Epoch] += d.Amount
	}

	var totalDeposits int64
	var earningsInPeriod int64
	var earningsInPeriodBalance int64

	for _, balance := range balancesMap {

		if int64(balance.ActivationEpoch) > latestEpoch {
			continue
		}

		for epoch, deposit := range depositsMap[fmt.Sprintf("%x", balance.PublicKey)] {
			totalDeposits += deposit

			if epoch >= threeWeeksBeforeEpoch && epoch <= lastWeekEpoch &&
				epoch > int64(balance.ActivationEpoch) {
				earningsInPeriod -= deposit
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

		if int64(balance.ActivationEpoch) <= lastMonthEpoch && balance.Status == "active_online" {
			earningsInPeriod += (int64(balance.Balance) - balance.Balance31d.Int64) - (int64(balance.Balance) - balance.Balance7d.Int64)
			earningsInPeriodBalance += balance.BalanceActivation.Int64
		}
	}

	updateChartDB(poolName, lastWeekEpoch, earningsInPeriod, earningsInPeriodBalance)
}

func updateChartDB(poolName string, epoch int64, income int64, balance int64) {
	if poolName == "" {
		return
	}
	_, err := WriterDb.Exec(`
		INSERT INTO staking_pools_chart
		(epoch, name, income, balance)
		VALUES
		($1, $2, $3, $4)
	`, epoch, poolName, income, balance)
	if err != nil {
		logger.Errorf("error inserting staking pool chart data (if 'duplicate key' error not critical): %v", err)
	}
}

func deleteOldChartEntries() {

	latestEpochUint, err := GetLatestEpoch()
	if err != nil {
		logger.Errorf("error getting latest epoch %v", err)
		return
	}

	latestEpoch := int64(latestEpochUint)
	sixMonthsOld := latestEpoch - int64(utils.EpochsPerDay())*31*6
	_, err = WriterDb.Exec(`
		DELETE FROM staking_pools_chart
		WHERE epoch <= $1
	`, sixMonthsOld)
	if err != nil {
		logger.Errorf("error removing old staking pool chart data: %v", err)
	}
}
