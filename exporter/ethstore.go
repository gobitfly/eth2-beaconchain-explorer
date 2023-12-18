package exporter

import (
	"context"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	ethstore "github.com/gobitfly/eth.store"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type EthStoreExporter struct {
	DB             *sqlx.DB
	BNAddress      string
	ENAddress      string
	UpdateInverval time.Duration
	ErrorInterval  time.Duration
	Sleep          time.Duration
}

// start exporting of eth.store into db
func StartEthStoreExporter(bnAddress string, enAddress string, updateInterval, errorInterval, sleepInterval time.Duration, startDayReexport, endDayReexport int64) {
	logger.Info("starting eth.store exporter")
	ese := &EthStoreExporter{
		DB:             db.WriterDb,
		BNAddress:      bnAddress,
		ENAddress:      enAddress,
		UpdateInverval: updateInterval,
		ErrorInterval:  errorInterval,
		Sleep:          sleepInterval,
	}
	// set sane defaults if config is not set
	if ese.UpdateInverval == 0 {
		ese.UpdateInverval = time.Minute
	}
	if ese.ErrorInterval == 0 {
		ese.ErrorInterval = time.Second * 10
	}
	if ese.Sleep == 0 {
		ese.Sleep = time.Minute
	}

	// Reexport days if specified
	if startDayReexport != -1 && endDayReexport != -1 {
		for day := startDayReexport; day <= endDayReexport; day++ {
			err := ese.reexportDay(strconv.FormatInt(day, 10))
			if err != nil {
				utils.LogError(err, fmt.Sprintf("error reexporting eth.store day %d in database", day), 0)
				return
			}
		}
		return
	}

	ese.Run()
}

func (ese *EthStoreExporter) reexportDay(day string) error {
	tx, err := ese.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ese.prepareClearDayTx(tx, day)
	if err != nil {
		return err
	}

	ese.prepareExportDayTx(tx, day)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (ese *EthStoreExporter) exportDay(day string) error {
	tx, err := ese.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = ese.prepareExportDayTx(tx, day)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (ese *EthStoreExporter) prepareClearDayTx(tx *sqlx.Tx, day string) error {
	dayInt, err := strconv.ParseInt(day, 10, 64)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`DELETE FROM eth_store_stats WHERE day = $1`, dayInt)
	return err
}

func (ese *EthStoreExporter) prepareExportDayTx(tx *sqlx.Tx, day string) error {
	ethStoreDay, validators, err := ese.getStoreDay(day)
	if err != nil {
		return err
	}

	numArgs := 10
	batchSize := 65535 / numArgs // max 65535 params per batch, since postgres uses int16 for binding input params
	valueArgs := make([]interface{}, 0, batchSize*numArgs)
	valueStrings := make([]string, 0, batchSize)
	valueStringArr := make([]string, numArgs)
	batchIdx, allIdx := 0, 0
	for index, day := range validators {
		for u := 0; u < numArgs; u++ {
			valueStringArr[u] = fmt.Sprintf("$%d", batchIdx*numArgs+1+u)
		}
		valueStrings = append(valueStrings, "("+strings.Join(valueStringArr, ",")+")")
		valueArgs = append(valueArgs, day.Day)
		valueArgs = append(valueArgs, index)
		valueArgs = append(valueArgs, day.EffectiveBalanceGwei.Mul(decimal.NewFromInt(1e9)))
		valueArgs = append(valueArgs, day.StartBalanceGwei.Mul(decimal.NewFromInt(1e9)))
		valueArgs = append(valueArgs, day.EndBalanceGwei.Mul(decimal.NewFromInt(1e9)))
		valueArgs = append(valueArgs, day.DepositsSumGwei.Mul(decimal.NewFromInt(1e9)))
		valueArgs = append(valueArgs, day.TxFeesSumWei)
		valueArgs = append(valueArgs, day.ConsensusRewardsGwei.Mul(decimal.NewFromInt(1e9)))
		valueArgs = append(valueArgs, day.TotalRewardsWei)
		valueArgs = append(valueArgs, day.Apr)
		batchIdx++
		allIdx++
		if batchIdx >= batchSize || allIdx >= len(validators) {
			stmt := fmt.Sprintf(`INSERT INTO eth_store_stats (day, validator, effective_balances_sum_wei, start_balances_sum_wei, end_balances_sum_wei, deposits_sum_wei, tx_fees_sum_wei, consensus_rewards_sum_wei, total_rewards_wei, apr) VALUES %s`, strings.Join(valueStrings, ","))
			_, err := tx.Exec(stmt, valueArgs...)
			if err != nil {
				return err
			}
			batchIdx = 0
			valueArgs = valueArgs[:0]
			valueStrings = valueStrings[:0]
		}
	}

	stmt, err := tx.Prepare(`
	INSERT INTO eth_store_stats (day, validator, effective_balances_sum_wei, start_balances_sum_wei, end_balances_sum_wei, deposits_sum_wei, tx_fees_sum_wei, consensus_rewards_sum_wei, total_rewards_wei, apr)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		ethStoreDay.Day,
		-1,
		ethStoreDay.EffectiveBalanceGwei.Mul(decimal.NewFromInt(1e9)),
		ethStoreDay.StartBalanceGwei.Mul(decimal.NewFromInt(1e9)),
		ethStoreDay.EndBalanceGwei.Mul(decimal.NewFromInt(1e9)),
		ethStoreDay.DepositsSumGwei.Mul(decimal.NewFromInt(1e9)),
		ethStoreDay.TxFeesSumWei,
		ethStoreDay.ConsensusRewardsGwei.Mul(decimal.NewFromInt(1e9)),
		ethStoreDay.TotalRewardsWei,
		ethStoreDay.Apr,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		insert into historical_pool_performance 
		select 
			eth_store_stats.day, 
			coalesce(validator_pool.pool, 'Unknown'), 
			count(*) as validators,
			sum(effective_balances_sum_wei) as effective_balances_sum_wei, 
			sum(start_balances_sum_wei) as start_balances_sum_wei, 
			sum(end_balances_sum_wei) as end_balances_sum_wei, 
			sum(deposits_sum_wei) as deposits_sum_wei, 
			sum(tx_fees_sum_wei) as tx_fees_sum_wei, 
			sum(consensus_rewards_sum_wei) as consensus_rewards_sum_wei, 
			sum(total_rewards_wei) as total_rewards_wei, 
			avg(eth_store_stats.apr) as apr
		from validators 
		left join validator_pool on validators.pubkey = validator_pool.publickey 
		inner join eth_store_stats on validators.validatorindex = eth_store_stats.validator 
		where day = $1 
		group by validator_pool.pool, eth_store_stats.day
		on conflict (day, pool) do update set
			day                         = excluded.day,
			pool                        = excluded.pool,
			validators                  = excluded.validators,
			effective_balances_sum_wei  = excluded.effective_balances_sum_wei,
			start_balances_sum_wei      = excluded.start_balances_sum_wei,
			end_balances_sum_wei        = excluded.end_balances_sum_wei,
			deposits_sum_wei            = excluded.deposits_sum_wei,
			tx_fees_sum_wei             = excluded.tx_fees_sum_wei,
			consensus_rewards_sum_wei   = excluded.consensus_rewards_sum_wei,
			total_rewards_wei           = excluded.total_rewards_wei,
			apr                         = excluded.apr`,
		ethStoreDay.Day)

	return err
}

func (ese *EthStoreExporter) getStoreDay(day string) (*ethstore.Day, map[uint64]*ethstore.Day, error) {
	logger.Infof("retrieving eth.store for day %v", day)
	return ethstore.Calculate(context.Background(), ese.BNAddress, ese.ENAddress, day, 1, ethstore.RECEIPTS_MODE_SINGLE)
}

func (ese *EthStoreExporter) Run() {
	t := time.NewTicker(ese.UpdateInverval)
	defer t.Stop()
DBCHECK:
	for {
		// get latest eth.store day
		latestFinalizedEpoch, err := db.GetLatestFinalizedEpoch()
		if err != nil {
			utils.LogError(err, "error retrieving latest finalized epoch from db", 0)
			time.Sleep(ese.ErrorInterval)
			continue
		}

		if latestFinalizedEpoch == 0 {
			utils.LogError(err, "error retrieved 0 as latest finalized epoch from the db", 0)
			time.Sleep(ese.ErrorInterval)
			continue
		}
		latestDay := utils.DayOfSlot(latestFinalizedEpoch*utils.Config.Chain.ClConfig.SlotsPerEpoch) - 1

		logger.Infof("latest day is %v", latestDay)
		// count rows of eth.store days in db
		var ethStoreDayCount uint64
		err = ese.DB.Get(&ethStoreDayCount, `
				SELECT COUNT(*)
				FROM eth_store_stats WHERE validator = -1`)
		if err != nil {
			utils.LogError(err, "error retrieving eth.store days count from db", 0)
			time.Sleep(ese.ErrorInterval)
			continue
		}

		logger.Infof("ethStoreDayCount is %v", ethStoreDayCount)

		if ethStoreDayCount <= latestDay {
			// db is incomplete
			// init export map, set every day to true
			daysToExport := make(map[uint64]bool)
			for i := uint64(0); i <= latestDay; i++ {
				daysToExport[i] = true
			}

			// set every existing day in db to false in export map
			if ethStoreDayCount > 0 {
				var ethStoreDays []types.EthStoreDay
				err = ese.DB.Select(&ethStoreDays, `
						SELECT day 
						FROM eth_store_stats WHERE validator = -1`)
				if err != nil {
					utils.LogError(err, "error retrieving eth.store days from db", 0)
					time.Sleep(ese.ErrorInterval)
					continue
				}
				for _, ethStoreDay := range ethStoreDays {
					daysToExport[ethStoreDay.Day] = false
				}
			}
			daysToExportArray := make([]uint64, 0, len(daysToExport))
			for dayToExport, shouldExport := range daysToExport {
				if shouldExport {
					daysToExportArray = append(daysToExportArray, dayToExport)
				}
			}

			sort.Slice(daysToExportArray, func(i, j int) bool {
				return daysToExportArray[i] > daysToExportArray[j]
			})
			// export missing days
			for _, dayToExport := range daysToExportArray {
				err = ese.exportDay(strconv.FormatUint(dayToExport, 10))
				if err != nil {
					utils.LogError(err, fmt.Sprintf("error exporting eth.store day %d into database", dayToExport), 0)
					time.Sleep(ese.ErrorInterval)
					continue DBCHECK
				}
				logger.Infof("exported eth.store day %d into db", dayToExport)
				if ethStoreDayCount < latestDay {
					// more than 1 day is being exported, sleep for duration specified in config
					time.Sleep(ese.Sleep)
				}
			}
		}

		services.ReportStatus("ethstoreExporter", "Running", nil)
		<-t.C
	}
}
