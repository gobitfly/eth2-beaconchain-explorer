package main

import (
	"eth2-exporter/db"
	"eth2-exporter/rpc"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"flag"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func main() {

	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")

	start := flag.Uint64("start", 1, "Start epoch")
	end := flag.Uint64("end", 1, "End epoch")

	flag.Parse()

	if *start == 1 && *end == 1 {
		monitor(*configPath)
	}

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg

	bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.Config.DepositChainID))
	if err != nil {
		logrus.Fatalf("error connecting to bigtable: %v", err)
	}
	defer bt.Close()

	chainIDBig := new(big.Int).SetUint64(utils.Config.Chain.Config.DepositChainID)

	rpcClient, err := rpc.NewLighthouseClient("http://"+cfg.Indexer.Node.Host+":"+cfg.Indexer.Node.Port, chainIDBig)
	if err != nil {
		logrus.Fatal(err)
	}

	for i := *start; i <= *end; i++ {
		i := i

		logrus.Infof("exporting epoch %v", i)

		logrus.Infof("deleting existing epoch data")
		err := bt.DeleteEpoch(i)
		if err != nil {
			logrus.Fatal(err)
		}

		firstSlot := i * utils.Config.Chain.Config.SlotsPerEpoch
		lastSlot := (i+1)*utils.Config.Chain.Config.SlotsPerEpoch - 1

		c, err := rpcClient.GetSyncCommittee(fmt.Sprintf("%d", firstSlot), i)
		if err != nil {
			logrus.Fatal(err)
		}

		validatorsU64 := make([]uint64, len(c.Validators))
		for i, idxStr := range c.Validators {
			idxU64, err := strconv.ParseUint(idxStr, 10, 64)
			if err != nil {
				logrus.Fatal(err)
			}
			validatorsU64[i] = idxU64
		}

		logrus.Infof("saving sync assignments for %v validators", len(validatorsU64))

		err = db.BigtableClient.SaveSyncCommitteesAssignments(firstSlot, lastSlot, validatorsU64)
		if err != nil {
			logrus.Fatalf("error saving sync committee assignments: %v", err)
		}
		logrus.Infof("exported sync committee assignments for epoch %v to bigtable in %v", i)

		data, err := rpcClient.GetEpochData(uint64(i), true)
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

		g.Go(func() error {
			return bt.SaveSyncComitteeDuties(data.Blocks)
		})

		err = g.Wait()

		if err != nil {
			logrus.Fatal(err)
		}
	}

}

func monitor(configPath string) {

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg

	bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.Config.DepositChainID))
	if err != nil {
		logrus.Fatalf("error connecting to bigtable: %v", err)
	}
	defer bt.Close()

	chainIDBig := new(big.Int).SetUint64(utils.Config.Chain.Config.DepositChainID)

	rpcClient, err := rpc.NewLighthouseClient("http://"+cfg.Indexer.Node.Host+":"+cfg.Indexer.Node.Port, chainIDBig)
	if err != nil {
		logrus.Fatal(err)
	}
	current := uint64(0)

	for {
		head, err := rpcClient.GetChainHead()
		if err != nil {
			logrus.Fatal(err)
		}

		logrus.Infof("current is %v, head is %v, finalized is %v", current, head.HeadEpoch, head.FinalizedEpoch)

		if current == head.HeadEpoch {
			time.Sleep(time.Second * 12)
			continue
		}

		for i := head.FinalizedEpoch; i <= head.HeadEpoch; i++ {
			logrus.Infof("exporting epoch %v", i)
			data, err := rpcClient.GetEpochData(i, true)
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

			g.Go(func() error {
				return bt.SaveSyncComitteeDuties(data.Blocks)
			})

			err = g.Wait()

			if err != nil {
				logrus.Fatal(err)
			}
		}
		current = head.HeadEpoch
	}
}
