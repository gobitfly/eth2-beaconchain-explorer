package main

import (
	"eth2-exporter/db"
	"eth2-exporter/rpc"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"flag"
	"fmt"
	"math/big"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func main() {

	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")

	start := flag.Uint64("start", 1, "Start epoch")
	end := flag.Uint64("end", 1, "End epoch")

	versionFlag := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version.Version)
		return
	}

	if *start == 1 && *end == 1 {
		monitor(*configPath)
	}

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg

	bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.Config.DepositChainID), utils.Config.RedisCacheEndpoint)
	if err != nil {
		logrus.Fatalf("error connecting to bigtable: %v", err)
	}
	defer bt.Close()

	chainIDBig := new(big.Int).SetUint64(utils.Config.Chain.Config.DepositChainID)

	rpcClient, err := rpc.NewLighthouseClient("http://"+cfg.Indexer.Node.Host+":"+cfg.Indexer.Node.Port, chainIDBig)
	if err != nil {
		utils.LogFatal(err, "new bigtable lighthouse client error", 0)
	}

	for i := *start; i <= *end; i++ {
		i := i

		logrus.Infof("exporting epoch %v", i)

		logrus.Infof("deleting existing epoch data")
		err := bt.DeleteEpoch(i)
		if err != nil {
			utils.LogFatal(err, "deleting epoch error", 0)
		}

		data, err := rpcClient.GetEpochData(uint64(i), true)
		if err != nil {
			utils.LogFatal(err, "getting epoch data error", 0)
		}

		g := new(errgroup.Group)

		g.Go(func() error {
			err = db.BigtableClient.SaveProposalAssignments(i, data.ValidatorAssignmentes.ProposerAssignments)
			if err != nil {
				return fmt.Errorf("error exporting proposal assignments to bigtable: %v", err)
			}
			return nil
		})
		g.Go(func() error {
			err = db.BigtableClient.SaveAttestationDuties(data.AttestationDuties)
			if err != nil {
				return fmt.Errorf("error exporting attestations to bigtable: %v", err)
			}
			return nil
		})
		g.Go(func() error {
			err = db.BigtableClient.SaveProposals(data.Blocks)
			if err != nil {
				return fmt.Errorf("error exporting proposals to bigtable: %v", err)
			}
			return nil
		})
		g.Go(func() error {
			err = db.BigtableClient.SaveSyncComitteeDuties(data.SyncDuties)
			if err != nil {
				return fmt.Errorf("error exporting sync committee duties to bigtable: %v", err)
			}
			return nil
		})

		err = g.Wait()

		if err != nil {
			utils.LogFatal(err, "wait group error", 0)
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

	bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.Config.DepositChainID), utils.Config.RedisCacheEndpoint)
	if err != nil {
		logrus.Fatalf("error connecting to bigtable: %v", err)
	}
	defer bt.Close()

	chainIDBig := new(big.Int).SetUint64(utils.Config.Chain.Config.DepositChainID)

	rpcClient, err := rpc.NewLighthouseClient("http://"+cfg.Indexer.Node.Host+":"+cfg.Indexer.Node.Port, chainIDBig)
	if err != nil {
		utils.LogFatal(err, "new bigtable lighthouse client in monitor error", 0)
	}
	current := uint64(0)

	for {
		head, err := rpcClient.GetChainHead()
		if err != nil {
			utils.LogFatal(err, "getting chain head from lighthouse in monitor error", 0)
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
				utils.LogFatal(err, "getting epoch data error", 0)
			}

			g := new(errgroup.Group)

			g.Go(func() error {
				err = db.BigtableClient.SaveProposalAssignments(i, data.ValidatorAssignmentes.ProposerAssignments)
				if err != nil {
					return fmt.Errorf("error exporting proposal assignments to bigtable: %v", err)
				}
				return nil
			})
			g.Go(func() error {
				err = db.BigtableClient.SaveAttestationDuties(data.AttestationDuties)
				if err != nil {
					return fmt.Errorf("error exporting attestations to bigtable: %v", err)
				}
				return nil
			})
			g.Go(func() error {
				err = db.BigtableClient.SaveProposals(data.Blocks)
				if err != nil {
					return fmt.Errorf("error exporting proposals to bigtable: %v", err)
				}
				return nil
			})
			g.Go(func() error {
				err = db.BigtableClient.SaveSyncComitteeDuties(data.SyncDuties)
				if err != nil {
					return fmt.Errorf("error exporting sync committee duties to bigtable: %v", err)
				}
				return nil
			})

			err = g.Wait()

			if err != nil {
				utils.LogFatal(err, "wait group error", 0)
			}
		}
		current = head.HeadEpoch
	}
}
