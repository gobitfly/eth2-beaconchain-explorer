package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/metrics"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
	"github.com/gobitfly/eth2-beaconchain-explorer/version"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/sirupsen/logrus"
)

func main() {
	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version.Version)
		fmt.Println(version.GoVersion)
		return
	}

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg
	logrus.WithField("config", *configPath).WithField("version", version.Version).WithField("chainName", utils.Config.Chain.ClConfig.ConfigName).Printf("starting")

	if utils.Config.Metrics.Enabled {
		go func(addr string) {
			logrus.Infof("serving metrics on %v", addr)
			if err := metrics.Serve(addr); err != nil {
				logrus.WithError(err).Fatal("Error serving metrics")
			}
		}(utils.Config.Metrics.Address)
	}

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

	nrp := NewNodeJobsProcessor(utils.Config.NodeJobsProcessor.ClEndpoint, utils.Config.NodeJobsProcessor.ElEndpoint)
	go nrp.Run()

	utils.WaitForCtrlC()
	logrus.Println("exiting …")
}

type NodeJobsProcessor struct {
	ELAddr string
	CLAddr string
	logger *logrus.Entry
}

func NewNodeJobsProcessor(clAddr, elAddr string) *NodeJobsProcessor {
	logger := logrus.New().WithField("module", "node-jobs-processor")
	njp := &NodeJobsProcessor{
		CLAddr: clAddr,
		ELAddr: elAddr,
		logger: logger,
	}
	return njp
}

func (njp *NodeJobsProcessor) Run() {
	for {
		err := njp.Process()
		if err != nil {
			njp.logger.WithError(err).Errorf("error processing node-jobs")
		}
		time.Sleep(time.Second * 10)
	}
}

func (njp *NodeJobsProcessor) Process() error {
	err := db.UpdateNodeJobs()
	if err != nil {
		return fmt.Errorf("error updating job: %w", err)
	}
	err = db.SubmitNodeJobs()
	if err != nil {
		return fmt.Errorf("error submitting job: %w", err)
	}
	return nil
}
