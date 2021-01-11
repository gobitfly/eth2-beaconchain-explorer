package main

import (
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"flag"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/sirupsen/logrus"
	"time"
)

func main() {
	configPath := flag.String("config", "", "Path to the config file")
	dayToExport := flag.Int64("day", -1, "Day to export (will export the day independent if it has been already exported or not")

	flag.Parse()

	logrus.Printf("config file path: %v", *configPath)
	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)

	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg

	db.MustInitDB(cfg.Database.Username, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	defer db.DB.Close()

	if *dayToExport >= 0 {
		_, err := db.DB.Exec("delete from validator_stats_status where day = $1", *dayToExport)
		if err != nil {
			logrus.Fatalf("error resetting status for day %v: %v", *dayToExport, err)
		}

		err = db.WriteStatisticsForDay(uint64(*dayToExport))
		if err != nil {
			logrus.Errorf("error exporting stats for day %v: %v", dayToExport, err)
		}
		return
	}

	for true {
		latestEpoch, err := db.GetLatestEpoch()
		if err != nil {
			logrus.Errorf("error retreiving latest epoch from the db: %v", err)
		}
		currentDay := latestEpoch / ((24 * 60 * 60) / utils.Config.Chain.SlotsPerEpoch / utils.Config.Chain.SecondsPerSlot)
		previousDay := currentDay - 1

		var lastExportedDay uint64
		err = db.DB.Get(&lastExportedDay, "select COALESCE(max(day), 0) from validator_stats_status where status")
		if err != nil {
			logrus.Errorf("error retreiving latest exported day from the db: %v", err)
		}
		if lastExportedDay != 0 {
			lastExportedDay++
		}

		logrus.Infof("previous day is %v, last exported day is %v", previousDay, lastExportedDay)
		if lastExportedDay <= previousDay || lastExportedDay == 0 {
			for day := lastExportedDay; day <= previousDay; day++ {
				err := db.WriteStatisticsForDay(day)
				if err != nil {
					logrus.Errorf("error exporting stats for day %v: %v", day, err)
				}
			}
		}
		time.Sleep(time.Minute)
	}
}
