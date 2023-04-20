package main

import (
	"eth2-exporter/cache"
	"eth2-exporter/db"
	"eth2-exporter/price"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/sirupsen/logrus"
)

type options struct {
	configPath                string
	statisticsDayToExport     int64
	statisticsDaysToExport    string
	statisticsValidatorToggle bool
	statisticsChartToggle     bool
}

var opt *options

func main() {
	configPath := flag.String("config", "", "Path to the config file")
	statisticsDayToExport := flag.Int64("statistics.day", -1, "Day to export statistics (will export the day independent if it has been already exported or not")
	statisticsDaysToExport := flag.String("statistics.days", "", "Days to export statistics (will export the day independent if it has been already exported or not")
	statisticsValidatorToggle := flag.Bool("validators.enabled", false, "Toggle exporting validator statistics")
	statisticsChartToggle := flag.Bool("charts.enabled", false, "Toggle exporting chart series")

	flag.Parse()

	opt = &options{
		configPath:                *configPath,
		statisticsDayToExport:     *statisticsDayToExport,
		statisticsDaysToExport:    *statisticsDaysToExport,
		statisticsChartToggle:     *statisticsChartToggle,
		statisticsValidatorToggle: *statisticsValidatorToggle,
	}

	logrus.Printf("version: %v, config file path: %v", version.Version, *configPath)
	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)

	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg

	if *statisticsChartToggle && utils.Config.Chain.Config.DepositChainID != 1 {
		logrus.Infof("Execution charts are currently only available for mainnet")
	}

	db.MustInitDB(&types.DatabaseConfig{
		Username: cfg.WriterDatabase.Username,
		Password: cfg.WriterDatabase.Password,
		Name:     cfg.WriterDatabase.Name,
		Host:     cfg.WriterDatabase.Host,
		Port:     cfg.WriterDatabase.Port,
	}, &types.DatabaseConfig{
		Username: cfg.ReaderDatabase.Username,
		Password: cfg.ReaderDatabase.Password,
		Name:     cfg.ReaderDatabase.Name,
		Host:     cfg.ReaderDatabase.Host,
		Port:     cfg.ReaderDatabase.Port,
	})
	defer db.ReaderDb.Close()
	defer db.WriterDb.Close()

	db.MustInitFrontendDB(&types.DatabaseConfig{
		Username: cfg.Frontend.WriterDatabase.Username,
		Password: cfg.Frontend.WriterDatabase.Password,
		Name:     cfg.Frontend.WriterDatabase.Name,
		Host:     cfg.Frontend.WriterDatabase.Host,
		Port:     cfg.Frontend.WriterDatabase.Port,
	}, &types.DatabaseConfig{
		Username: cfg.Frontend.ReaderDatabase.Username,
		Password: cfg.Frontend.ReaderDatabase.Password,
		Name:     cfg.Frontend.ReaderDatabase.Name,
		Host:     cfg.Frontend.ReaderDatabase.Host,
		Port:     cfg.Frontend.ReaderDatabase.Port,
	})
	defer db.FrontendReaderDB.Close()
	defer db.FrontendWriterDB.Close()

	_, err = db.InitBigtable(cfg.Bigtable.Project, cfg.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.Config.DepositChainID))
	if err != nil {
		logrus.Fatalf("error connecting to bigtable: %v", err)
	}

	price.Init(utils.Config.Chain.Config.DepositChainID, utils.Config.Eth1ErigonEndpoint)

	if utils.Config.TieredCacheProvider == "redis" || len(utils.Config.RedisCacheEndpoint) != 0 {
		cache.MustInitTieredCache(utils.Config.RedisCacheEndpoint)
	} else if utils.Config.TieredCacheProvider == "bigtable" && len(utils.Config.RedisCacheEndpoint) == 0 {
		cache.MustInitTieredCacheBigtable(db.BigtableClient.GetClient(), fmt.Sprintf("%d", utils.Config.Chain.Config.DepositChainID))
	}

	if utils.Config.TieredCacheProvider != "bigtable" && utils.Config.TieredCacheProvider != "redis" {
		logrus.Fatalf("No cache provider set. Please set TierdCacheProvider (example redis, bigtable)")
	}

	if *statisticsDaysToExport != "" {
		s := strings.Split(*statisticsDaysToExport, "-")
		if len(s) < 2 {
			logrus.Fatalf("invalid arg")
		}
		firstDay, err := strconv.ParseUint(s[0], 10, 64)
		if err != nil {
			utils.LogFatal(err, "error parsing first day of statisticsDaysToExport flag to uint", 0)
		}
		lastDay, err := strconv.ParseUint(s[1], 10, 64)
		if err != nil {
			utils.LogFatal(err, "error parsing last day of statisticsDaysToExport flag to uint", 0)
		}

		if *statisticsValidatorToggle {
			logrus.Infof("exporting validator statistics for days %v-%v", firstDay, lastDay)
			for d := firstDay; d <= lastDay; d++ {
				_, err := db.WriterDb.Exec("delete from validator_stats_status where day = $1", d)
				if err != nil {
					logrus.Fatalf("error resetting status for day %v: %v", d, err)
				}

				err = db.WriteValidatorStatisticsForDay(uint64(d))
				if err != nil {
					logrus.Errorf("error exporting stats for day %v: %v", d, err)
				}
			}
		}

		if *statisticsChartToggle && utils.Config.Chain.Config.DepositChainID == 1 {
			logrus.Infof("exporting chart series for days %v-%v", firstDay, lastDay)
			for d := firstDay; d <= lastDay; d++ {
				_, err = db.WriterDb.Exec("delete from chart_series_status where day = $1", d)
				if err != nil {
					logrus.Fatalf("error resetting status for chart series status for day %v: %v", d, err)
				}

				err = db.WriteChartSeriesForDay(int64(d))
				if err != nil {
					logrus.Errorf("error exporting chart series from day %v: %v", d, err)
				}
			}
		}

		return
	} else if *statisticsDayToExport >= 0 {

		if *statisticsValidatorToggle {
			_, err := db.WriterDb.Exec("delete from validator_stats_status where day = $1", *statisticsDayToExport)
			if err != nil {
				logrus.Fatalf("error resetting status for day %v: %v", *statisticsDayToExport, err)
			}

			err = db.WriteValidatorStatisticsForDay(uint64(*statisticsDayToExport))
			if err != nil {
				logrus.Errorf("error exporting stats for day %v: %v", *statisticsDayToExport, err)
			}
		}

		if *statisticsChartToggle && utils.Config.Chain.Config.DepositChainID == 1 {
			_, err = db.WriterDb.Exec("delete from chart_series_status where day = $1", *statisticsDayToExport)
			if err != nil {
				logrus.Fatalf("error resetting status for chart series status for day %v: %v", *statisticsDayToExport, err)
			}

			err = db.WriteChartSeriesForDay(int64(*statisticsDayToExport))
			if err != nil {
				logrus.Errorf("error exporting chart series from day %v: %v", *statisticsDayToExport, err)
			}
		}
		return
	}

	go statisticsLoop()

	utils.WaitForCtrlC()

	logrus.Println("exiting...")
}

func statisticsLoop() {
	for {

		latestEpoch := services.LatestFinalizedEpoch()
		if latestEpoch == 0 {
			logrus.Errorf("error retreiving latest finalized epoch from cache")
			time.Sleep(time.Minute)
			continue
		}

		epochsPerDay := utils.EpochsPerDay()
		if latestEpoch < epochsPerDay {
			logrus.Infof("skipping exporting stats, first day has not been indexed yet")
			time.Sleep(time.Minute)
			continue
		}
		currentDay := latestEpoch / epochsPerDay
		previousDay := currentDay - 1

		if previousDay > currentDay {
			previousDay = currentDay
		}

		if opt.statisticsValidatorToggle {
			var lastExportedDayValidator uint64
			err := db.WriterDb.Get(&lastExportedDayValidator, "select COALESCE(max(day), 0) from validator_stats_status where status")
			if err != nil {
				logrus.Errorf("error retreiving latest exported day from the db: %v", err)
			}
			if lastExportedDayValidator != 0 {
				lastExportedDayValidator++
			}

			logrus.Infof("Validator Statistics: Latest epoch is %v, previous day is %v, last exported day is %v", latestEpoch, previousDay, lastExportedDayValidator)
			if lastExportedDayValidator <= previousDay || lastExportedDayValidator == 0 {
				for day := lastExportedDayValidator; day <= previousDay; day++ {
					err := db.WriteValidatorStatisticsForDay(day)
					if err != nil {
						logrus.Errorf("error exporting stats for day %v: %v", day, err)
					}
				}
			}

		}

		if opt.statisticsChartToggle {
			var lastExportedDayChart uint64
			err := db.WriterDb.Get(&lastExportedDayChart, "select COALESCE(max(day), 0) from chart_series_status where status")
			if err != nil {
				logrus.Errorf("error retreiving latest exported day from the db: %v", err)
			}
			if lastExportedDayChart != 0 {
				lastExportedDayChart++
			}
			if utils.Config.Chain.Config.DepositChainID == 1 {
				logrus.Infof("Chart statistics: latest epoch is %v, previous day is %v, last exported day is %v", latestEpoch, previousDay, lastExportedDayChart)
				if lastExportedDayChart <= previousDay || lastExportedDayChart == 0 {
					for day := lastExportedDayChart; day <= previousDay; day++ {
						err = db.WriteChartSeriesForDay(int64(day))
						if err != nil {
							logrus.Errorf("error exporting chart series from day %v: %v", day, err)
						}
					}
				}
			}
		}

		services.ReportStatus("statistics", "Running", nil)
		time.Sleep(time.Minute)
	}
}
