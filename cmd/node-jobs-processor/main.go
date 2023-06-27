package main

import (
	"eth2-exporter/db"
	"eth2-exporter/metrics"
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
	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")
	metricsAddr := flag.String("metrics.address", "localhost:9090", "serve metrics on that addr")
	metricsEnabled := flag.Bool("metrics.enabled", false, "enable serving metrics")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version.Version)
		return
	}

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg
	logrus.WithField("config", *configPath).WithField("version", version.Version).WithField("chainName", utils.Config.Chain.Config.ConfigName).Printf("starting")

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

	nrp := NewNodeJobsProcessor(utils.Config.NodeJobsProcessor.ClEndpoint, utils.Config.NodeJobsProcessor.ElEndpoint)
	go nrp.Run()

	if *metricsEnabled {
		go func() {
			logrus.WithFields(logrus.Fields{"addr": *metricsAddr}).Infof("Serving metrics")
			if err := metrics.Serve(*metricsAddr); err != nil {
				logrus.WithError(err).Fatal("Error serving metrics")
			}
		}()
	}

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
