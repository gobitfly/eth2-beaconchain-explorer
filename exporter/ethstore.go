package exporter

import (
	"context"
	"eth2-exporter/db"
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

type EthStoreDay struct {
	Day                  uint64 `db:"day"`
	EffectiveBalancesSum uint64 `db:"effective_balances_sum"`
	StartBalancesSum     uint64 `db:"start_balances_sum"`
	EndBalancesSum       uint64 `db:"end_balances_sum"`
	DepositsSum          uint64 `db:"deposits_sum"`
}

// start exporting of eth.store into db
func ethStoreExporter() {
	logger.Info("starting eth.store exporter")
	ese := &EthStoreExporter{
		DB:             db.WriterDb,
		NodeHost:       utils.Config.Indexer.Node.Host,
		NodePort:       utils.Config.Indexer.Node.Port,
		UpdateInverval: time.Minute,
		ErrorInterval:  time.Second * 10,
		Sleep:          utils.Config.EthStoreExporter.Sleep,
	}

	ese.Run()

}

func (ese *EthStoreExporter) ExportDay(day string) error {
	ethStoreDay, err := ese.getStoreDay(day)
	if err != nil {
		return err
	}
	_, err = ese.DB.Exec(`
		INSERT INTO eth_store_stats (day, effective_balances_sum, start_balances_sum, end_balances_sum, deposits_sum)
		VALUES ($1, $2, $3, $4, $5)`,
		ethStoreDay.Day, ethStoreDay.EffectiveBalance, ethStoreDay.StartBalance, ethStoreDay.EndBalance, ethStoreDay.DepositsSum)
	if err != nil {
		return err
	}
	logger.Infof("exported eth.store day %s into db", day)
	return nil
}

func (ese *EthStoreExporter) getStoreDay(day string) (*ethstore.Day, error) {
	return ethstore.Calculate(context.Background(), fmt.Sprintf("http://%s:%s", ese.NodeHost, ese.NodePort), day, nil)
}

func (ese *EthStoreExporter) Run() {
	t := time.NewTicker(ese.UpdateInverval)
	defer t.Stop()
OUTER:
	for {
		// get latest eth.store day
		latest, err := ese.getStoreDay("latest")
		if err != nil {
			logger.WithError(err).Errorf("error retreiving eth.store data")
			time.Sleep(ese.ErrorInterval)
			continue
		}

		// count rows of eth.store days in db
		var ethStoreDayCount uint64
		err = ese.DB.Get(&ethStoreDayCount, `
				SELECT COUNT(*)
				FROM eth_store_stats`)
		if err != nil {
			logger.WithError(err).Error("error retreiving eth.store days count from db")
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
				var ethStoreDays []EthStoreDay
				err = ese.DB.Select(&ethStoreDays, `
						SELECT day 
						FROM eth_store_stats`)
				if err != nil {
					logger.WithError(err).Error("error retreiving eth.store days from db")
					time.Sleep(ese.ErrorInterval)
					continue
				}
				for _, ethStoreDay := range ethStoreDays {
					daysToExport[ethStoreDay.Day] = false
				}
			}

			// export missing days
			for k, v := range daysToExport {
				if v {
					err = ese.ExportDay(strconv.FormatUint(k, 10))
					if err != nil {
						logger.WithError(err).Errorf("error exporting eth.store day %d into database", k)
						time.Sleep(ese.ErrorInterval)
						continue OUTER
					}
				}
				if ethStoreDayCount < latest.Day {
					// more than 1 day is being exported, sleep for duration specified in config
					time.Sleep(ese.Sleep)
				}
			}
		}
		<-t.C
	}
}
