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
	configPath                    string
	statisticsDayToExport         int64
	statisticsDaysToExport        string
	statisticsValidatorToggle     bool
	statisticsResetColumns        string
	statisticsChartToggle         bool
	statisticsGraffitiToggle      bool
	concurrencyTotal              uint64
	concurrencyCl                 uint64
	concurrencyFailedAttestations uint64
}

var opt = &options{}

func main() {
	flag.StringVar(&opt.configPath, "config", "", "Path to the config file")
	flag.Int64Var(&opt.statisticsDayToExport, "statistics.day", -1, "Day to export statistics (will export the day independent if it has been already exported or not")
	flag.StringVar(&opt.statisticsDaysToExport, "statistics.days", "", "Days to export statistics (will export the day independent if it has been already exported or not")
	flag.BoolVar(&opt.statisticsValidatorToggle, "validators.enabled", false, "Toggle exporting validator statistics")
	flag.StringVar(&opt.statisticsResetColumns, "validators.reset", "", "validator_stats_status columns to reset. Comma separated. Use 'all' for complete resync.")
	flag.BoolVar(&opt.statisticsChartToggle, "charts.enabled", false, "Toggle exporting chart series")
	flag.BoolVar(&opt.statisticsGraffitiToggle, "graffiti.enabled", false, "Toggle exporting graffiti statistics")
	flag.Uint64Var(&opt.concurrencyTotal, "concurrency.total", 10, "Concurrency to use when writing total rewards/performance postgres queries")
	flag.Uint64Var(&opt.concurrencyCl, "concurrency.cl", 50, "Concurrency to use when writing cl postgres queries")
	flag.Uint64Var(&opt.concurrencyFailedAttestations, "concurrency.fa", 10, "Concurrency to use when fetching failed attestaations from bt")

	versionFlag := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version.Version)
		return
	}

	logrus.Printf("version: %v, config file path: %v", version.Version, opt.configPath)
	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, opt.configPath)

	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg

	if utils.Config.Chain.Config.SlotsPerEpoch == 0 || utils.Config.Chain.Config.SecondsPerSlot == 0 {
		utils.LogFatal(fmt.Errorf("error ether SlotsPerEpoch [%v] or SecondsPerSlot [%v] are not set", utils.Config.Chain.Config.SlotsPerEpoch, utils.Config.Chain.Config.SecondsPerSlot), "", 0)
		return
	} else {
		logrus.Infof("Writing statistic with: SlotsPerEpoch [%v] or SecondsPerSlot [%v]", utils.Config.Chain.Config.SlotsPerEpoch, utils.Config.Chain.Config.SecondsPerSlot)
	}

	db.MustInitDB(&types.DatabaseConfig{
		Username:     cfg.WriterDatabase.Username,
		Password:     cfg.WriterDatabase.Password,
		Name:         cfg.WriterDatabase.Name,
		Host:         cfg.WriterDatabase.Host,
		Port:         cfg.WriterDatabase.Port,
		MaxOpenConns: cfg.WriterDatabase.MaxOpenConns,
		MaxIdleConns: cfg.WriterDatabase.MaxIdleConns,
	}, &types.DatabaseConfig{
		Username:     cfg.ReaderDatabase.Username,
		Password:     cfg.ReaderDatabase.Password,
		Name:         cfg.ReaderDatabase.Name,
		Host:         cfg.ReaderDatabase.Host,
		Port:         cfg.ReaderDatabase.Port,
		MaxOpenConns: cfg.ReaderDatabase.MaxOpenConns,
		MaxIdleConns: cfg.ReaderDatabase.MaxIdleConns,
	})
	defer db.ReaderDb.Close()
	defer db.WriterDb.Close()

	db.MustInitFrontendDB(&types.DatabaseConfig{
		Username:     cfg.Frontend.WriterDatabase.Username,
		Password:     cfg.Frontend.WriterDatabase.Password,
		Name:         cfg.Frontend.WriterDatabase.Name,
		Host:         cfg.Frontend.WriterDatabase.Host,
		Port:         cfg.Frontend.WriterDatabase.Port,
		MaxOpenConns: cfg.Frontend.WriterDatabase.MaxOpenConns,
		MaxIdleConns: cfg.Frontend.WriterDatabase.MaxIdleConns,
	}, &types.DatabaseConfig{
		Username:     cfg.Frontend.ReaderDatabase.Username,
		Password:     cfg.Frontend.ReaderDatabase.Password,
		Name:         cfg.Frontend.ReaderDatabase.Name,
		Host:         cfg.Frontend.ReaderDatabase.Host,
		Port:         cfg.Frontend.ReaderDatabase.Port,
		MaxOpenConns: cfg.Frontend.ReaderDatabase.MaxOpenConns,
		MaxIdleConns: cfg.Frontend.ReaderDatabase.MaxIdleConns,
	})
	defer db.FrontendReaderDB.Close()
	defer db.FrontendWriterDB.Close()

	_, err = db.InitBigtable(cfg.Bigtable.Project, cfg.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.Config.DepositChainID), utils.Config.RedisCacheEndpoint)
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

	if opt.statisticsDaysToExport != "" {
		s := strings.Split(opt.statisticsDaysToExport, "-")
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

		if opt.statisticsValidatorToggle {
			logrus.Infof("exporting validator statistics for days %v-%v", firstDay, lastDay)
			for d := firstDay; d <= lastDay; d++ {

				clearStatsStatusTable(d, opt.statisticsResetColumns)

				err = db.WriteValidatorStatisticsForDay(uint64(d), opt.concurrencyTotal, opt.concurrencyCl, opt.concurrencyFailedAttestations)
				if err != nil {
					logrus.Errorf("error exporting stats for day %v: %v", d, err)
					break
				}
			}
		}

		if opt.statisticsChartToggle {
			logrus.Infof("exporting chart series for days %v-%v", firstDay, lastDay)
			for d := firstDay; d <= lastDay; d++ {
				_, err = db.WriterDb.Exec("delete from chart_series_status where day = $1", d)
				if err != nil {
					logrus.Fatalf("error resetting status for chart series status for day %v: %v", d, err)
				}

				err = db.WriteChartSeriesForDay(int64(d))
				if err != nil {
					logrus.Errorf("error exporting chart series from day %v: %v", d, err)
					break
				}
			}
		}

		if opt.statisticsGraffitiToggle {
			for d := firstDay; d <= lastDay; d++ {
				err = db.WriteGraffitiStatisticsForDay(int64(d))
				if err != nil {
					logrus.Errorf("error exporting graffiti-stats from day %v: %v", opt.statisticsDayToExport, err)
					break
				}
			}
		}

		return
	} else if opt.statisticsDayToExport >= 0 {

		if opt.statisticsValidatorToggle {
			clearStatsStatusTable(uint64(opt.statisticsDayToExport), opt.statisticsResetColumns)

			err = db.WriteValidatorStatisticsForDay(uint64(opt.statisticsDayToExport), opt.concurrencyTotal, opt.concurrencyCl, opt.concurrencyFailedAttestations)
			if err != nil {
				logrus.Errorf("error exporting stats for day %v: %v", opt.statisticsDayToExport, err)
			}
		}

		if opt.statisticsChartToggle {
			_, err = db.WriterDb.Exec("delete from chart_series_status where day = $1", opt.statisticsDayToExport)
			if err != nil {
				logrus.Fatalf("error resetting status for chart series status for day %v: %v", opt.statisticsDayToExport, err)
			}

			err = db.WriteChartSeriesForDay(int64(opt.statisticsDayToExport))
			if err != nil {
				logrus.Errorf("error exporting chart series from day %v: %v", opt.statisticsDayToExport, err)
			}
		}

		if opt.statisticsGraffitiToggle {
			err = db.WriteGraffitiStatisticsForDay(int64(opt.statisticsDayToExport))
			if err != nil {
				logrus.Errorf("error exporting chart series from day %v: %v", opt.statisticsDayToExport, err)
			}
		}
		return
	}

	go statisticsLoop(opt.concurrencyTotal, opt.concurrencyCl, opt.concurrencyFailedAttestations)

	utils.WaitForCtrlC()

	logrus.Println("exiting...")
}

func statisticsLoop(concurrencyTotal uint64, concurrencyCl uint64, concurrencyFailedAttestations uint64) {
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
			lastExportedDayValidator, err := db.GetLastExportedStatisticDay()
			if err != nil {
				logrus.Errorf("error retreiving latest exported day from the db: %v", err)
			}

			logrus.Infof("Validator Statistics: Latest epoch is %v, previous day is %v, last exported day is %v", latestEpoch, previousDay, lastExportedDayValidator)
			if lastExportedDayValidator != 0 {
				lastExportedDayValidator++
			}
			if lastExportedDayValidator <= previousDay || lastExportedDayValidator == 0 {
				for day := lastExportedDayValidator; day <= previousDay; day++ {
					err := db.WriteValidatorStatisticsForDay(day, concurrencyTotal, concurrencyCl, concurrencyFailedAttestations)
					if err != nil {
						logrus.Errorf("error exporting stats for day %v: %v", day, err)
						break
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

			logrus.Infof("Chart statistics: latest epoch is %v, previous day is %v, last exported day is %v", latestEpoch, previousDay, lastExportedDayChart)
			if lastExportedDayChart != 0 {
				lastExportedDayChart++
			}
			if lastExportedDayChart <= previousDay || lastExportedDayChart == 0 {
				for day := lastExportedDayChart; day <= previousDay; day++ {
					err = db.WriteChartSeriesForDay(int64(day))
					if err != nil {
						logrus.Errorf("error exporting chart series from day %v: %v", day, err)
						break
					}
				}
			}
		}

		if opt.statisticsGraffitiToggle {
			var lastDay int64
			err := db.WriterDb.Get(&lastDay, "select COALESCE(max(day), 0) from graffiti_stats")
			if err != nil {
				logrus.Errorf("error retreiving latest exported day from graffiti_stats: %v", err)
			} else {
				nextDay := lastDay + 1
				err = db.WriteGraffitiStatisticsForDay(nextDay)
				if err != nil {
					logrus.Errorf("error exporting graffiti-stats for day %v: %v", nextDay, err)
				}
			}
		}

		services.ReportStatus("statistics", "Running", nil)
		time.Sleep(time.Minute)
	}
}

func clearStatsStatusTable(day uint64, columns string) {
	if columns == "all" {
		logrus.Infof("Delete validator_stats_status for day %v", day)
		_, err := db.WriterDb.Exec("DELETE FROM validator_stats_status WHERE day = $1", day)
		if err != nil {
			logrus.Fatalf("error resetting status for day %v: %v", day, err)
		}
	} else if len(columns) > 0 {
		logrus.Infof("Resetting columns %v of validator_stats_status for day %v ", columns, day)
		cols := strings.Join(strings.Split(columns, ","), " = false,")
		_, err := db.WriterDb.Exec(fmt.Sprintf(`
			UPDATE validator_stats_status
			SET %v = false
			WHERE day = $1
		`, cols), day)
		if err != nil {
			logrus.Fatalf("error resetting status for day %v: %v", day, err)
		}
	}
}
