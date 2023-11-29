package main

import (
	"eth2-exporter/cache"
	"eth2-exporter/db"
	"eth2-exporter/price"
	"eth2-exporter/rpc"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"flag"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sirupsen/logrus"
)

type options struct {
	configPath                string
	statisticsDayToExport     int64
	statisticsDaysToExport    string
	statisticsValidatorToggle bool
	statisticsChartToggle     bool
	statisticsGraffitiToggle  bool
	resetStatus               bool
}

var opt = &options{}

func main() {
	flag.StringVar(&opt.configPath, "config", "", "Path to the config file")
	flag.Int64Var(&opt.statisticsDayToExport, "statistics.day", -1, "Day to export statistics (will export the day independent if it has been already exported or not")
	flag.StringVar(&opt.statisticsDaysToExport, "statistics.days", "", "Days to export statistics (will export the day independent if it has been already exported or not")
	flag.BoolVar(&opt.statisticsValidatorToggle, "validators.enabled", false, "Toggle exporting validator statistics")
	flag.BoolVar(&opt.statisticsChartToggle, "charts.enabled", false, "Toggle exporting chart series")
	flag.BoolVar(&opt.statisticsGraffitiToggle, "graffiti.enabled", false, "Toggle exporting graffiti statistics")
	flag.BoolVar(&opt.resetStatus, "validators.reset", false, "Export stats independet if they have already been exported previously")

	versionFlag := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version.Version)
		fmt.Println(version.GoVersion)
		return
	}

	logrus.Printf("version: %v, config file path: %v", version.Version, opt.configPath)
	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, opt.configPath)

	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg

	if utils.Config.Chain.ClConfig.SlotsPerEpoch == 0 || utils.Config.Chain.ClConfig.SecondsPerSlot == 0 {
		utils.LogFatal(fmt.Errorf("error ether SlotsPerEpoch [%v] or SecondsPerSlot [%v] are not set", utils.Config.Chain.ClConfig.SlotsPerEpoch, utils.Config.Chain.ClConfig.SecondsPerSlot), "", 0)
		return
	} else {
		logrus.Infof("Writing statistic with: SlotsPerEpoch [%v] or SecondsPerSlot [%v]", utils.Config.Chain.ClConfig.SlotsPerEpoch, utils.Config.Chain.ClConfig.SecondsPerSlot)
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

	_, err = db.InitBigtable(cfg.Bigtable.Project, cfg.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.ClConfig.DepositChainID), utils.Config.RedisCacheEndpoint)
	if err != nil {
		logrus.Fatalf("error connecting to bigtable: %v", err)
	}

	price.Init(utils.Config.Chain.ClConfig.DepositChainID, utils.Config.Eth1ErigonEndpoint, utils.Config.Frontend.ClCurrency, utils.Config.Frontend.ElCurrency)

	if utils.Config.TieredCacheProvider != "redis" {
		logrus.Fatalf("No cache provider set. Please set TierdCacheProvider (example redis)")
	}

	if utils.Config.TieredCacheProvider == "redis" || len(utils.Config.RedisCacheEndpoint) != 0 {
		cache.MustInitTieredCache(utils.Config.RedisCacheEndpoint)
	}

	var rpcClient rpc.Client

	chainID := new(big.Int).SetUint64(utils.Config.Chain.ClConfig.DepositChainID)
	if utils.Config.Indexer.Node.Type == "lighthouse" {
		rpcClient, err = rpc.NewLighthouseClient("http://"+cfg.Indexer.Node.Host+":"+cfg.Indexer.Node.Port, chainID)
		if err != nil {
			utils.LogFatal(err, "new explorer lighthouse client error", 0)
		}
	} else {
		logrus.Fatalf("invalid note type %v specified. supported node types are prysm and lighthouse", utils.Config.Indexer.Node.Type)
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

				if opt.resetStatus {
					clearStatsStatusTable(d)
				}

				err = db.WriteValidatorStatisticsForDay(uint64(d), rpcClient)
				if err != nil {
					utils.LogError(err, fmt.Errorf("error exporting stats for day %v", d), 0)
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
			if opt.resetStatus {
				clearStatsStatusTable(uint64(opt.statisticsDayToExport))
			}

			err = db.WriteValidatorStatisticsForDay(uint64(opt.statisticsDayToExport), rpcClient)
			if err != nil {
				utils.LogError(err, fmt.Errorf("error exporting stats for day %v", opt.statisticsDayToExport), 0)
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

	go statisticsLoop(rpcClient)

	utils.WaitForCtrlC()

	logrus.Println("exiting...")
}

func statisticsLoop(client rpc.Client) {
	for {

		var loopError error
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

		logrus.Infof("Performing statisticsLoop with currentDay %v and previousDay %v", currentDay, previousDay)
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
					err := db.WriteValidatorStatisticsForDay(day, client)
					if err != nil {
						utils.LogError(err, fmt.Errorf("error exporting stats for day %v", day), 0)
						loopError = err
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
						loopError = err
						break
					}
				}
			}
		}

		if opt.statisticsGraffitiToggle {
			graffitiStatsStatus := []struct {
				Day    uint64
				Status bool
			}{}
			err := db.WriterDb.Select(&graffitiStatsStatus, "select day, status from graffiti_stats_status")
			if err != nil {
				logrus.Errorf("error retrieving graffitiStatsStatus: %v", err)
			} else {
				graffitiStatsStatusMap := map[uint64]bool{}
				for _, s := range graffitiStatsStatus {
					graffitiStatsStatusMap[s.Day] = s.Status
				}
				for day := uint64(0); day <= currentDay; day++ {
					if !graffitiStatsStatusMap[day] {
						logrus.Infof("exporting graffiti-stats for day %v", day)
						err = db.WriteGraffitiStatisticsForDay(int64(day))
						if err != nil {
							logrus.Errorf("error exporting graffiti-stats for day %v: %v", day, err)
							loopError = err
							break
						}
					}
				}
			}
		}

		if loopError == nil {
			services.ReportStatus("statistics", "Running", nil)
		} else {
			services.ReportStatus("statistics", loopError.Error(), nil)
		}
		time.Sleep(time.Minute)
	}
}

func clearStatsStatusTable(day uint64) {
	logrus.Infof("deleting validator_stats_status for day %v", day)
	_, err := db.WriterDb.Exec("DELETE FROM validator_stats_status WHERE day = $1", day)
	if err != nil {
		logrus.Fatalf("error resetting status for day %v: %v", day, err)
	}
}
