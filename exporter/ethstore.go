package exporter

import (
	"context"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"sort"
	"strconv"
	"time"

	ethstore "github.com/gobitfly/eth.store"
	"github.com/jmoiron/sqlx"
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
func StartEthStoreExporter(bnAddress string, enAddress string, updateInterval, errorInterval, sleepInterval time.Duration) {
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

	ese.Run()

}

func (ese *EthStoreExporter) ExportDay(day string) error {
	ethStoreDay, err := ese.getStoreDay(day)
	if err != nil {
		return err
	}
	_, err = ese.DB.Exec(`
		INSERT INTO eth_store_stats (day, effective_balances_sum, start_balances_sum, end_balances_sum, deposits_sum, tx_fees_sum)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		ethStoreDay.Day,
		ethStoreDay.EffectiveBalanceGwei,
		ethStoreDay.StartBalanceGwei,
		ethStoreDay.EndBalanceGwei,
		ethStoreDay.DepositsSumGwei,
		ethStoreDay.TxFeesSumWei.String())
	if err != nil {
		return err
	}
	return nil
}

func (ese *EthStoreExporter) getStoreDay(day string) (*ethstore.Day, error) {
	logger.Infof("retrieving eth.store for day %v", day)
	ethstore.SetDebugLevel(1)
	return ethstore.Calculate(context.Background(), ese.BNAddress, ese.ENAddress, day)
}

func (ese *EthStoreExporter) Run() {
	t := time.NewTicker(ese.UpdateInverval)
	defer t.Stop()
DBCHECK:
	for {
		// get latest eth.store day
		var latestFinalizedEpoch uint64
		err := db.WriterDb.Get(&latestFinalizedEpoch, "SELECT COALESCE(MAX(epoch), 0) FROM epochs where finalized is true")
		if err != nil {
			logger.WithError(err).Error("error retrieving latest finalized epoch from db")
			time.Sleep(ese.ErrorInterval)
			continue
		}
		latestDay := utils.DayOfSlot(latestFinalizedEpoch * utils.Config.Chain.Config.SlotsPerEpoch)

		logger.Infof("latest day is %v", latestDay)
		// count rows of eth.store days in db
		var ethStoreDayCount uint64
		err = ese.DB.Get(&ethStoreDayCount, `
				SELECT COUNT(*)
				FROM eth_store_stats`)
		if err != nil {
			logger.WithError(err).Error("error retrieving eth.store days count from db")
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
				var ethStoreDays []types.PerformanceDay
				err = ese.DB.Select(&ethStoreDays, `
						SELECT day 
						FROM eth_store_stats`)
				if err != nil {
					logger.WithError(err).Error("error retrieving eth.store days from db")
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
				return daysToExportArray[i] < daysToExportArray[j]
			})
			// export missing days
			for _, dayToExport := range daysToExportArray {
				err = ese.ExportDay(strconv.FormatUint(dayToExport, 10))
				if err != nil {
					logger.WithError(err).Errorf("error exporting eth.store day %d into database", dayToExport)
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
		<-t.C
	}
}
