package main

import (
	"eth2-exporter/db"
	"eth2-exporter/price"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"flag"
	"fmt"
	"strconv"
	"strings"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/sirupsen/logrus"
)

func main() {
	configPath := flag.String("config", "", "Path to the config file")
	statisticsDaysToExport := flag.String("statistics.days", "", "Days to export statistics (will export the day independent if it has been already exported or not")

	flag.Parse()

	logrus.Printf("version: %v, config file path: %v", version.Version, *configPath)

	if *statisticsDaysToExport != "" {
		cfg := &types.Config{}
		err := utils.ReadConfig(cfg, *configPath)

		if err != nil {
			logrus.Fatalf("error reading config file: %v", err)
		}
		utils.Config = cfg

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

		logrus.Infof("exporting validator statistics for days %v-%v", firstDay, lastDay)
		for d := firstDay; d <= lastDay; d++ {
			err = db.WriteValidatorProposerStatisticsForDay(uint64(d))
			if err != nil {
				logrus.Errorf("error exporting stats for day %v: %v", d, err)
			}
		}
		logrus.Println("finished updating proposer reward statistics, exiting...")
	} else {
		logrus.Println("no days to export specified, exiting...")
	}
}
