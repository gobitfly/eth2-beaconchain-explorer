package main

import (
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"flag"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/sirupsen/logrus"
)

func main() {
	var fromDay = 0
	var toDay = 0
	var batchLimit = 0
	var sleepInBetween = 0

	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")
	flag.IntVar(&fromDay, "day.from", 0, "start day to migrate")
	flag.IntVar(&toDay, "day.to", 9999999, "end day to migrate")
	flag.IntVar(&batchLimit, "batch-size", 20000, "batch size")
	flag.IntVar(&sleepInBetween, "sleep", 5, "sleep in between batches in seconds")

	flag.Parse()

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg
	logrus.WithField("config", *configPath).WithField("version", version.Version).WithField("chainName", utils.Config.Chain.Config.ConfigName).Printf("starting")

	_, err = db.InitBigtable(cfg.Bigtable.Project, cfg.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.Config.DepositChainID))
	if err != nil {
		logrus.Fatalf("error initializing bigtable %v", err)
	}

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

	logrus.Infof("[%v, %v], batchSize: %v, sleepInBetween: %v", fromDay, toDay, batchLimit, sleepInBetween)
	time.Sleep(5 * time.Second)
	err = db.BigtableClient.MigrateMachineStatsFromDBToBigtable(fromDay, toDay, batchLimit, sleepInBetween)
	if err != nil {
		logrus.Errorf("Can not migrate data: %v", err)
	}
}
