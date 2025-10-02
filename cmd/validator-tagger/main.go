package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/metrics"
	"github.com/gobitfly/eth2-beaconchain-explorer/rpc"
	"github.com/gobitfly/eth2-beaconchain-explorer/services"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
	"github.com/gobitfly/eth2-beaconchain-explorer/version"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sirupsen/logrus"
)

func main() {
	var err error

	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")
	versionFlag := flag.Bool("version", false, "Show version and exit")
	scheduleFlag := flag.Bool("schedule", false, "Start scheduler loop (daily 10:00 UTC and hourly precompute)")
	runFlag := flag.String("run", "", "Comma-separated steps to run on demand (import,lido,lido_csm,rocketpool,withdrawal_tagging,deposit_tagging,populate_validator_names,precompute,all)")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version.Version)
		fmt.Println(version.GoVersion)
		return
	}

	cfg := &types.Config{}
	if err := utils.ReadConfig(cfg, *configPath); err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg
	logrus.WithFields(logrus.Fields{
		"config":    *configPath,
		"version":   version.Version,
		"chainName": utils.Config.Chain.ClConfig.ConfigName,
	}).Info("starting validator-tagger")

	// Validate critical chain params (common pattern across binaries)
	if utils.Config.Chain.ClConfig.SlotsPerEpoch == 0 || utils.Config.Chain.ClConfig.SecondsPerSlot == 0 || utils.Config.Chain.GenesisTimestamp == 0 {
		logrus.Fatal("invalid chain configuration: missing SlotsPerEpoch, SecondsPerSlot, or GenesisTimestamp")
	}

	// Initialize primary databases (writer/reader)
	db.MustInitDB(&types.DatabaseConfig{
		Username:     cfg.WriterDatabase.Username,
		Password:     cfg.WriterDatabase.Password,
		Name:         cfg.WriterDatabase.Name,
		Host:         cfg.WriterDatabase.Host,
		Port:         cfg.WriterDatabase.Port,
		MaxOpenConns: cfg.WriterDatabase.MaxOpenConns,
		MaxIdleConns: cfg.WriterDatabase.MaxIdleConns,
		SSL:          cfg.WriterDatabase.SSL,
	}, &types.DatabaseConfig{
		Username:     cfg.ReaderDatabase.Username,
		Password:     cfg.ReaderDatabase.Password,
		Name:         cfg.ReaderDatabase.Name,
		Host:         cfg.ReaderDatabase.Host,
		Port:         cfg.ReaderDatabase.Port,
		MaxOpenConns: cfg.ReaderDatabase.MaxOpenConns,
		MaxIdleConns: cfg.ReaderDatabase.MaxIdleConns,
		SSL:          cfg.ReaderDatabase.SSL,
	}, "pgx", "postgres")
	defer db.ReaderDb.Close()
	defer db.WriterDb.Close()

	// If ClickHouse HTTP is enabled, fetch the large result as zstd-compressed Parquet and parse with parquet-go
	db.MustInitClickhouseDB(nil, &types.DatabaseConfig{
		Username:     cfg.ClickHouse.ReaderDatabase.Username,
		Password:     cfg.ClickHouse.ReaderDatabase.Password,
		Name:         cfg.ClickHouse.ReaderDatabase.Name,
		Host:         cfg.ClickHouse.ReaderDatabase.Host,
		Port:         cfg.ClickHouse.ReaderDatabase.Port,
		MaxOpenConns: cfg.ClickHouse.ReaderDatabase.MaxOpenConns,
		MaxIdleConns: cfg.ClickHouse.ReaderDatabase.MaxIdleConns,
		SSL:          true,
	}, "clickhouse", "clickhouse")

	// Initialize Erigon RPC client (required for Lido exporter step)
	rpc.CurrentErigonClient, err = rpc.NewErigonClient(utils.Config.Eth1ErigonEndpoint)
	if err != nil {
		logrus.Fatalf("error initializing erigon client: %v", err)
	}

	if utils.Config.Metrics.Enabled {
		go func(addr string) {
			logrus.Infof("serving metrics on %v", addr)
			if err := metrics.Serve(addr); err != nil {
				logrus.WithError(err).Fatal("Error serving metrics")
			}
		}(utils.Config.Metrics.Address)
	}

	// On-demand single/multi-step run
	if *runFlag != "" {
		if err := services.RunValidatorTaggerOnDemand(context.Background(), *runFlag); err != nil {
			logrus.Fatalf("on-demand run failed: %v", err)
		}
		logrus.Info("on-demand run completed; exiting")
		return
	}

	// Scheduler mode (enabled only if flag set or config toggled)
	if *scheduleFlag || utils.Config.ValidatorTagger.SchedulerEnabled {
		services.RunValidatorTaggerScheduler()
		return
	}
}
