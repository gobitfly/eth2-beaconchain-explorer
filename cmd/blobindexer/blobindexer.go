package main

import (
	"flag"
	"fmt"

	"github.com/gobitfly/eth2-beaconchain-explorer/exporter"
	"github.com/gobitfly/eth2-beaconchain-explorer/metrics"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
	"github.com/gobitfly/eth2-beaconchain-explorer/version"

	"github.com/sirupsen/logrus"
)

func main() {
	configFlag := flag.String("config", "config.yml", "path to config")
	versionFlag := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *versionFlag {
		fmt.Println(version.Version)
		return
	}
	utils.Config = &types.Config{}
	err := utils.ReadConfig(utils.Config, *configFlag)
	if err != nil {
		logrus.Fatal(err)
	}
	if utils.Config.Metrics.Enabled {
		go func(addr string) {
			logrus.Infof("serving metrics on %v", addr)
			if err := metrics.Serve(addr); err != nil {
				logrus.WithError(err).Fatal("Error serving metrics")
			}
		}(utils.Config.Metrics.Address)
	}
	blobIndexer, err := exporter.NewBlobIndexer()
	if err != nil {
		logrus.Fatal(err)
	}
	go blobIndexer.Start()
	utils.WaitForCtrlC()
}
