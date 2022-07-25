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
	ese := &EthStoreExporter{
		DB:             db.WriterDb,
		NodeHost:       utils.Config.Indexer.Node.Host,
		NodePort:       utils.Config.Indexer.Node.Port,
		UpdateInverval: time.Minute,
		ErrorInterval:  time.Second * 10,
	}

	ese.Run()

}

func (ese *EthStoreExporter) ExportDay(tx *sqlx.Tx, day string) error {
	ethStoreDay, err := ese.getStoreDay(day)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
		INSERT INTO eth_store_stats (day, effective_balances_sum, start_balances_sum, end_balances_sum, deposits_sum)
		VALUES ($1, $2, $3, $4, $5)`,
		ethStoreDay.Day, ethStoreDay.EffectiveBalance, ethStoreDay.StartBalance, ethStoreDay.EndBalance, ethStoreDay.DepositsSum)
	if err != nil {
		return err
	}
	return nil
}

func (ese *EthStoreExporter) getStoreDay(day string) (*ethstore.Day, error) {
	return ethstore.Calculate(context.Background(), fmt.Sprintf("http://%s:%s", ese.NodeHost, ese.NodePort), day, nil)
}

func (ese *EthStoreExporter) Run() {
	t := time.NewTicker(ese.UpdateInverval)
	defer t.Stop()
	for ; true; <-t.C {
		// get latest eth.store day
		latest, err := ese.getStoreDay("latest")
		if err != nil {
			logger.WithError(err).Errorf("error retreiving eth.store data")
			t.Reset(ese.ErrorInterval)
			continue
		}

		// count rows of eth.store days in db
		var ethStoreDayCount uint64
		err = db.WriterDb.Get(&ethStoreDayCount, `
			SELECT COUNT(*) 
			FROM eth_store_stats`)
		if err != nil {
			logger.WithError(err).Error("error retreiving db data")
			t.Reset(ese.ErrorInterval)
			continue
		}

		if ethStoreDayCount <= latest.Day {
			// db is incomplete
			// init export map, set every day to true
			daysToExport := make(map[uint64]bool)
			for i := uint64(0); i <= latest.Day; i++ {
				daysToExport[i] = true
			}

			//init db txs
			tx, err := ese.DB.Beginx()
			if err != nil {
				logger.WithError(err).Errorf("error starting db transactions")
				t.Reset(ese.ErrorInterval)
				continue
			}
			defer tx.Rollback()

			// set every existing day in db to false in export map
			if ethStoreDayCount > 0 {
				var ethStoreDays []EthStoreDay
				err = tx.Select(&ethStoreDays, `
					SELECT day 
					FROM eth_store_stats 
					ORDER BY day DESC`)
				if err != nil {
					logger.WithError(err).Error("error retreiving db data")
					t.Reset(ese.ErrorInterval)
					continue
				}
				for _, ethStoreDay := range ethStoreDays {
					daysToExport[ethStoreDay.Day] = false
				}
			}

			// export missing days
			for k, v := range daysToExport {
				if v {
					err = ese.ExportDay(tx, strconv.FormatUint(k, 10))
					if err != nil {
						logger.WithError(err).Errorf("error exporting day $d into database", k)
						t.Reset(ese.ErrorInterval)
						continue
					}
				}
			}
			tx.Commit()
			logger.Infof("exported missing eth.store days")
		}
	}
}
