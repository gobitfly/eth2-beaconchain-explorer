package main

import (
	"eth2-exporter/exporter2"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"flag"

	"github.com/sirupsen/logrus"
)

var clientURL string
var configPath string

func main() {
	flag.StringVar(&clientURL, "url", "", "url")
	flag.StringVar(&configPath, "config", "config-example.yml", "Path to the config file")
	flag.Parse()
	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg
	exporter2.Start(clientURL)
}
