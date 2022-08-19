package main

import (
	"eth2-exporter/db"
	"eth2-exporter/rpc"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"flag"
	"fmt"
	"math/big"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func main() {

	project := flag.String("project", "", "GCP project name")
	instance := flag.String("instance", "", "Bigtable instance name")
	chainId := flag.Uint64("chainId", 1, "Chain ID")

	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")

	start := flag.Int("start", 1, "Start epoch")
	end := flag.Int("end", 1, "End epoch")

	flag.Parse()

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg

	bt, err := db.InitBigtable(*project, *instance, fmt.Sprintf("%d", *chainId))
	if err != nil {
		logrus.Fatalf("error connecting to bigtable: %v", err)
	}
	defer bt.Close()

	chainIDBig := new(big.Int).SetUint64(*chainId)

	rpcClient, err := rpc.NewLighthouseClient("http://"+cfg.Indexer.Node.Host+":"+cfg.Indexer.Node.Port, chainIDBig)
	if err != nil {
		logrus.Fatal(err)
	}

	for i := *start; i <= *end; i++ {
		data, err := rpcClient.GetEpochData(uint64(i))
		if err != nil {
			logrus.Fatal(err)
		}

		g := new(errgroup.Group)

		g.Go(func() error {
			return bt.SaveValidatorBalances(data.Epoch, data.Validators)
		})

		g.Go(func() error {
			return bt.SaveAttestationAssignments(data.Epoch, data.ValidatorAssignmentes.AttestorAssignments)
		})

		g.Go(func() error {
			return bt.SaveProposalAssignments(data.Epoch, data.ValidatorAssignmentes.ProposerAssignments)
		})

		g.Go(func() error {
			return bt.SaveAttestations(data.Blocks)
		})

		g.Go(func() error {
			return bt.SaveProposals(data.Blocks)
		})

		err = g.Wait()

		if err != nil {
			logrus.Fatal(err)
		}
	}
}
