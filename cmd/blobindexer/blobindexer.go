package main

import (
	"flag"
	"fmt"

	"github.com/gobitfly/eth2-beaconchain-explorer/exporter"
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
	blobIndexer, err := exporter.NewBlobIndexer()
	if err != nil {
		logrus.Fatal(err)
	}
	go blobIndexer.Start()
	utils.WaitForCtrlC()
}
