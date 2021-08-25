package services

import (
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"flag"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	configPath := flag.String("config", "config.yml", "Path to the config file")
	flag.Parse()
	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg

	if cfg.Database.Password != "xxx" {
		logrus.Fatal("error do not run these tests in production")
	}

	db.MustInitDB(cfg.Database.Username, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	defer db.DB.Close()

	logger.Infof("connected to db:          %+v", cfg.Database)

	db.MustInitFrontendDB(cfg.Frontend.Database.Username, cfg.Frontend.Database.Password, cfg.Frontend.Database.Host, cfg.Frontend.Database.Port, cfg.Frontend.Database.Name, cfg.Frontend.SessionSecret)
	defer db.FrontendDB.Close()

	logger.Infof("connected to FrontendDB:  %+v", cfg.Frontend.Database)

	var epoch uint64
	err = db.DB.Get(&epoch, "SELECT COALESCE(MAX(epoch), 0) FROM epochs")
	if err != nil {
		logger.Errorf("error retrieving latest epoch from the database: %v", err)
		return
	}
	atomic.StoreUint64(&latestEpoch, epoch)

	var slot uint64
	err = db.DB.Get(&slot, "SELECT COALESCE(MAX(slot), 0) FROM blocks where slot < $1", utils.TimeToSlot(uint64(time.Now().Add(time.Second*10).Unix())))
	if err != nil {
		logger.Errorf("error retrieving latest slot from the database: %v", err)
		return
	}
	atomic.StoreUint64(&latestSlot, slot)

	m.Run()
}
