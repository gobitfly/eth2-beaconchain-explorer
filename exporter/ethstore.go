package exporter

import (
	"context"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	ethstore "github.com/gobitfly/eth.store"
	"github.com/jmoiron/sqlx"
)

type EthStoreExporter struct {
	DB                *sqlx.DB
	ExecutionEndpoint string
	ConsensusEndpoint string
	UpdateInverval    time.Duration
	ErrorInterval     time.Duration
	Sleep             time.Duration
}

// start exporting of eth.store into db
func ethStoreExporter() {
	logger.Info("starting eth.store exporter")
	ese := &EthStoreExporter{
		DB:                db.WriterDb,
		ExecutionEndpoint: utils.Config.EthStoreExporter.ExecutionEndpoint,
		ConsensusEndpoint: utils.Config.EthStoreExporter.ConsensusEndpoint,
		UpdateInverval:    utils.Config.EthStoreExporter.UpdateInterval,
		ErrorInterval:     utils.Config.EthStoreExporter.ErrorInterval,
		Sleep:             utils.Config.EthStoreExporter.Sleep,
	}
	// set sane defaults if config is not set
	if len(ese.ExecutionEndpoint) == 0 {
		ese.ExecutionEndpoint = utils.Config.Indexer.Eth1Endpoint
	}
	if len(ese.ConsensusEndpoint) == 0 {
		ese.ConsensusEndpoint = fmt.Sprintf("http://%s:%s", utils.Config.Indexer.Node.Host, utils.Config.Indexer.Node.Port)
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

	// export json file into db if config is set
	var err error
	if utils.Config.EthStoreExporter.JSONExportEnabled {
		// check if json path is set in config, otherwise get path through chain network name
		if len(utils.Config.EthStoreExporter.JSONPath) > 0 {
			err = ese.ExportJSON(utils.Config.EthStoreExporter.JSONPath)
		} else {
			err = ese.ExportJSON(fmt.Sprintf("exporter/ethstore.%s.json", utils.Config.Chain.Name))
		}
	}
	if err != nil {
		logger.WithError(err).Error("eth.store exporter: failed exporting json into db, continuing to export normally")
	}

	ese.Run()
}

func (ese *EthStoreExporter) ExportJSON(path string) error {
	// check if filepath exists
	logger.Info(path)
	_, err := os.Stat(path)
	if err != nil {
		return err
	}

	// parse json file
	jsonDays := []*ethstore.Day{}

	jsonDaysBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonDaysBytes, &jsonDays)
	if err != nil {
		return err
	}

	// build insert string
	nArgs := 6

	valueStringsArr := make([]string, nArgs)
	for i := range valueStringsArr {
		valueStringsArr[i] = "$%d"
	}
	valueStringsTpl := "(" + strings.Join(valueStringsArr, ",") + ")"
	valueStringsArgs := make([]interface{}, nArgs)
	valueStrings := make([]string, 0, len(jsonDays))
	valueArgs := make([]interface{}, 0, len(jsonDays)*nArgs)

	for dayIndex, day := range jsonDays {
		for argIndex := 0; argIndex < nArgs; argIndex++ {
			valueStringsArgs[argIndex] = dayIndex*nArgs + argIndex + 1
		}
		valueStrings = append(valueStrings, fmt.Sprintf(valueStringsTpl, valueStringsArgs...))
		valueArgs = append(valueArgs, day.Day)
		valueArgs = append(valueArgs, day.EffectiveBalanceGwei)
		valueArgs = append(valueArgs, day.StartBalanceGwei)
		valueArgs = append(valueArgs, day.EndBalanceGwei)
		valueArgs = append(valueArgs, day.DepositsSumGwei)
		valueArgs = append(valueArgs, day.TxFeesSumWei.String())

	}
	stmt := fmt.Sprintf(`
	INSERT INTO eth_store_stats (day, effective_balances_sum, start_balances_sum, end_balances_sum, deposits_sum, tx_fees_sum)
	VALUES %s on conflict (day) do nothing`, strings.Join(valueStrings, ","))

	// export days into db
	_, err = ese.DB.Exec(stmt, valueArgs...)
	if err != nil {
		return err
	}

	logger.Infof("eth.store exporter: exported %d days from json into db", len(jsonDays))

	return nil
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
	return ethstore.Calculate(context.Background(), ese.ConsensusEndpoint, ese.ExecutionEndpoint, day)
}

func (ese *EthStoreExporter) Run() {
	t := time.NewTicker(ese.UpdateInverval)
	defer t.Stop()
DBCHECK:
	for {
		// get latest eth.store day
		latest, err := ese.getStoreDay("finalized")
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

		if ethStoreDayCount <= latest.Day.BigInt().Uint64() {
			// db is incomplete
			// init export map, set every day to true
			daysToExport := make(map[uint64]bool)
			for i := uint64(0); i <= latest.Day.BigInt().Uint64(); i++ {
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
					if ethStoreDayCount < latest.Day.BigInt().Uint64() {
						// more than 1 day is being exported, sleep for duration specified in config
						time.Sleep(ese.Sleep)
					}
				}
			}
		}
		<-t.C
	}
}
