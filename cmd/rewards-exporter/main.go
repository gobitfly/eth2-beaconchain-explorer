package main

import (
	"eth2-exporter/db"
	"eth2-exporter/rpc"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"flag"
	"fmt"
	"math/big"
	"time"

	geth_rpc "github.com/ethereum/go-ethereum/rpc"

	eth_rewards "github.com/gobitfly/eth-rewards"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/prysmaticlabs/prysm/v3/api/client/beacon"
	"github.com/sirupsen/logrus"
)

func main() {

	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")
	bnAddress := flag.String("beacon-node-address", "", "Url of the beacon node api")
	enAddress := flag.String("execution-node-address", "", "Url of the execution node api")
	epoch := flag.Int64("epoch", -1, "epoch to export (use -1 to export latest finalized epoch)")
	network := flag.String("network", "", "Config to use (can be mainnet, prater or sepolia")

	flag.Parse()

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg
	logrus.WithField("config", *configPath).WithField("version", version.Version).WithField("chainName", utils.Config.Chain.Config.ConfigName).Printf("starting")

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

	client, err := beacon.NewClient(*bnAddress)
	if err != nil {
		logrus.Fatal(err)
	}

	lc, err := rpc.NewLighthouseClient(*bnAddress, big.NewInt(5))
	if err != nil {
		logrus.Fatal(err)
	}

	elClient, err := geth_rpc.Dial(*enAddress)
	if err != nil {
		logrus.Fatal(err)
	}

	bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.Config.DepositChainID))
	if err != nil {
		logrus.Fatalf("error connecting to bigtable: %v", err)
	}
	defer bt.Close()

	if *epoch == -1 {
		for {
			head, err := lc.GetChainHead()
			if err != nil {
				logrus.Fatal(err)
			}
			if int64(head.FinalizedEpoch) <= *epoch {

				services.ReportStatus("rewardsExporter", "Running", nil)
				time.Sleep(time.Second * 12)
				continue
			}

			*epoch = int64(head.FinalizedEpoch)

			start := time.Now()
			logrus.Infof("retrieving rewards details for epoch %v", *epoch)

			rewards, err := eth_rewards.GetRewardsForEpoch(int(*epoch), client, elClient, *network)

			if err != nil {
				logrus.Fatalf("error retrieving reward details for epoch %v: %v", *epoch, err)
			} else {
				logrus.Infof("retrieved %v reward details for epoch %v in %v", len(rewards), *epoch, time.Since(start))
			}

			err = bt.SaveValidatorIncomeDetails(uint64(*epoch), rewards)
			if err != nil {
				logrus.Fatalf("error saving reward details to bigtable: %v", err)
			}
		}
	}

	start := time.Now()
	logrus.Infof("retrieving rewards details for epoch %v", *epoch)

	rewards, err := eth_rewards.GetRewardsForEpoch(int(*epoch), client, elClient, *network)

	if err != nil {
		logrus.Fatalf("error retrieving reward details for epoch %v: %v", *epoch, err)
	} else {
		logrus.Infof("retrieved %v reward details for epoch %v in %v", len(rewards), *epoch, time.Since(start))
	}

	err = bt.SaveValidatorIncomeDetails(uint64(*epoch), rewards)
	if err != nil {
		logrus.Fatalf("error saving reward details to bigtable: %v", err)
	}

}
