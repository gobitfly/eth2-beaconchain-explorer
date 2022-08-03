package exporter

import (
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type HistoricalPoolPerformanceExporter struct {
	DB             *sqlx.DB
	UpdateInverval time.Duration
	ErrorInterval  time.Duration
	Sleep          time.Duration
}

func CHANGE_BEFORE_PR_LALALALALAL() {
	logger.Info("starting historicalpoolperformance exporter")
	hpp := HistoricalPoolPerformanceExporter{
		DB:             db.WriterDb,
		UpdateInverval: utils.Config.HistoricalPoolPerformanceExporter.UpdateInterval,
		ErrorInterval:  utils.Config.HistoricalPoolPerformanceExporter.ErrorInterval,
		Sleep:          utils.Config.HistoricalPoolPerformanceExporter.Sleep,
	}
	if hpp.UpdateInverval == 0 {
		hpp.UpdateInverval = time.Minute
	}
	if hpp.ErrorInterval == 0 {
		hpp.ErrorInterval = time.Second * 10
	}
	if hpp.Sleep == 0 {
		hpp.Sleep = time.Second * 5
	}

	hpp.Run()
}

func (hpp *HistoricalPoolPerformanceExporter) Run() {
	t := time.NewTicker(hpp.UpdateInverval)
	defer t.Stop()
DBCHECK:
	for {
		// get list of missing days in db
		var missingDays []types.PerformanceDay
		err := hpp.DB.Select(&missingDays, `
			SELECT	vss.day as day
			FROM 	validator_stats_status vss
					LEFT JOIN historical_pool_performance hpp
					ON vss.day = hpp.day
			WHERE 
					hpp.day IS NULL
					AND vss.status = true`)
		if err != nil {
			logger.WithError(err).Error("historicalpoolperformance exporter: error retrieving list of missing days in db")
			time.Sleep(hpp.ErrorInterval)
			continue
		}

		// export missing days
		for _, day := range missingDays {
			// retrieve day data from db
			poolPerfDay, err := hpp.GetPoolPerformanceDay(day.Day)
			if err != nil {
				logger.WithError(err).Errorf("historicalpoolperformance exporter: error retrieving data for day %d from db", day.Day)
				time.Sleep(hpp.ErrorInterval)
				continue DBCHECK
			}
			// export day data into db
			err = hpp.ExportPoolPerformanceDay(poolPerfDay)
			if err != nil {
				logger.WithError(err).Errorf("historicalpoolperformance exporter: error exporting day %d into db", day.Day)
				time.Sleep(hpp.ErrorInterval)
				continue DBCHECK
			}
			logger.Infof("historicalpoolperformance exporter: exported day %d into db", poolPerfDay[0].Day)
			time.Sleep(hpp.Sleep)
		}
		<-t.C
	}
}

func (hpp *HistoricalPoolPerformanceExporter) ExportPoolPerformanceDay(poolPerfDay []types.PerformanceDay) error {
	// build insert string
	nArgs := 6

	valueStringsArr := make([]string, nArgs)
	for i := range valueStringsArr {
		valueStringsArr[i] = "$%d"
	}
	valueStringsTpl := "(" + strings.Join(valueStringsArr, ",") + ")"
	valueStringsArgs := make([]interface{}, nArgs)
	valueStrings := make([]string, 0, len(poolPerfDay))
	valueArgs := make([]interface{}, 0, len(poolPerfDay)*nArgs)

	for poolIndex, pool := range poolPerfDay {
		for argIndex := 0; argIndex < nArgs; argIndex++ {
			valueStringsArgs[argIndex] = poolIndex*nArgs + argIndex + 1
		}
		valueStrings = append(valueStrings, fmt.Sprintf(valueStringsTpl, valueStringsArgs...))
		valueArgs = append(valueArgs, pool.Pool)
		valueArgs = append(valueArgs, pool.Day)
		valueArgs = append(valueArgs, pool.EffectiveBalancesSum)
		valueArgs = append(valueArgs, pool.StartBalancesSum)
		valueArgs = append(valueArgs, pool.EndBalancesSum)
		valueArgs = append(valueArgs, pool.DepositsSum)

	}
	stmt := fmt.Sprintf(`
	INSERT INTO historical_pool_performance (pool, day, effective_balances_sum, start_balances_sum, end_balances_sum, deposits_sum) 
	VALUES %s`, strings.Join(valueStrings, ","))

	// export
	_, err := hpp.DB.Exec(stmt, valueArgs...)
	if err != nil {
		return err
	}

	return nil
}

func (hpp *HistoricalPoolPerformanceExporter) GetPoolPerformanceDay(day uint64) ([]types.PerformanceDay, error) {
	var poolPerf []types.PerformanceDay

	err := hpp.DB.Select(&poolPerf, `
	SELECT pool,
		day,
		SUM(start_balance)                AS start_balances_sum,
		SUM(end_balance)                  AS end_balances_sum,
		SUM(start_effective_balance)      AS effective_balances_sum,
		COALESCE(SUM(deposits_amount), 0) AS deposits_sum
	FROM   validator_stats
		INNER JOIN validators
				ON validators.validatorindex = validator_stats.validatorindex
		INNER JOIN validator_pool
				ON validators.pubkey = validator_pool.publickey
	WHERE  day = $1
	GROUP  BY pool,
		   day
	HAVING SUM(start_effective_balance) > 0  `, day)
	if err != nil {
		return nil, err
	}
	return poolPerf, nil
}
