package exporter

import (
	"context"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"strconv"
	"time"

	ethstore "github.com/gobitfly/eth.store"
	"github.com/jmoiron/sqlx"
)

type EthStoreExporter struct {
	DB             *sqlx.DB
	NodeHost       string
	NodePort       string
	UpdateInverval time.Duration
	ErrorInterval  time.Duration
	Sleep          time.Duration
}

// start exporting of eth.store into db
func ethStoreExporter() {
	logger.Info("starting eth.store exporter")
	ese := &EthStoreExporter{
		DB:             db.WriterDb,
		NodeHost:       utils.Config.EthStoreExporter.Node.Host,
		NodePort:       utils.Config.EthStoreExporter.Node.Port,
		UpdateInverval: utils.Config.EthStoreExporter.UpdateInterval,
		ErrorInterval:  utils.Config.EthStoreExporter.ErrorInterval,
		Sleep:          utils.Config.EthStoreExporter.Sleep,
	}
	// set sane defaults if config is not set
	if len(ese.NodeHost) == 0 {
		ese.NodeHost = utils.Config.Indexer.Node.Host
	}
	if len(ese.NodePort) == 0 {
		ese.NodePort = utils.Config.Indexer.Node.Port
	}
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
	return ethstore.Calculate(context.Background(), fmt.Sprintf("http://%s:%s", ese.NodeHost, ese.NodePort), day)
}

func (ese *EthStoreExporter) Run() {
	t := time.NewTicker(ese.UpdateInverval)
	defer t.Stop()
DBCHECK:
	for {
		// get latest eth.store day
		latest, err := ese.getStoreDay("latest")
		if err != nil {
			logger.WithError(err).Errorf("error retrieving eth.store data")
			time.Sleep(ese.ErrorInterval)
			continue
		}

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

		if ethStoreDayCount <= latest.Day {
			// db is incomplete
			// init export map, set every day to true
			daysToExport := make(map[uint64]bool)
			for i := uint64(0); i <= latest.Day; i++ {
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

			// export missing days
			for dayToExport, shouldExport := range daysToExport {
				if shouldExport {
					err = ese.ExportDay(strconv.FormatUint(dayToExport, 10))
					if err != nil {
						logger.WithError(err).Errorf("error exporting eth.store day %d into database", dayToExport)
						time.Sleep(ese.ErrorInterval)
						continue DBCHECK
					}
					logger.Infof("exported eth.store day %d into db", dayToExport)
					if ethStoreDayCount < latest.Day {
						// more than 1 day is being exported, sleep for duration specified in config
						time.Sleep(ese.Sleep)
					}
				}
			}
		}
		<-t.C
	}
}
