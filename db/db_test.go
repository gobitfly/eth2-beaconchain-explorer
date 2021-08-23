package db

import (
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"flag"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
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
	if len(utils.Config.Chain.Phase0Path) > 0 {
		logrus.Infof("parsing phase 0 config path: %v", utils.Config.Chain.Phase0Path)
		phase0 := &types.Phase0{}
		f, err := os.Open(utils.Config.Chain.Phase0Path)
		if err != nil {
			wd, err := os.Getwd()
			if err != nil {
				logrus.WithError(err).Error("error getting current working directory")
			}
			logrus.Errorf("error opening Phase0 Config file %v in directory: %v: %v", utils.Config.Chain.Phase0Path, wd, err)
		} else {
			decoder := yaml.NewDecoder(f)
			err = decoder.Decode(phase0)
			if err != nil {
				logrus.Errorf("error decoding Phase0 Config file %v: %v", utils.Config.Chain.Phase0Path, err)
			} else {
				utils.Config.Chain.Phase0 = *phase0
			}
		}
	}

	if cfg.Database.Password != "xxx" {
		logrus.Fatal("error do not run these tests in production")
	}

	logrus.Infof("Running test for network: %+v", utils.Config.Chain.Phase0.ConfigName)

	MustInitDB(cfg.Database.Username, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	defer DB.Close()

	logger.Infof("connected to db:          %+v", cfg.Database)

	MustInitFrontendDB(cfg.Frontend.Database.Username, cfg.Frontend.Database.Password, cfg.Frontend.Database.Host, cfg.Frontend.Database.Port, cfg.Frontend.Database.Name, cfg.Frontend.SessionSecret)
	logger.Infof("connected to FrontendDB:  %+v", cfg.Frontend.Database)

	defer FrontendDB.Close()
	m.Run()
}
