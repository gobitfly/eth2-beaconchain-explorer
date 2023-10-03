package main

import (
	"eth2-exporter/db"
	"eth2-exporter/exporter"
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
	concurrency := flag.Int("concurrency", 1, "Number of parallel epoch exports")

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
	db.MustInitFrontendDB(&types.DatabaseConfig{
		Username:     cfg.Frontend.WriterDatabase.Username,
		Password:     cfg.Frontend.WriterDatabase.Password,
		Name:         cfg.Frontend.WriterDatabase.Name,
		Host:         cfg.Frontend.WriterDatabase.Host,
		Port:         cfg.Frontend.WriterDatabase.Port,
		MaxOpenConns: cfg.Frontend.WriterDatabase.MaxOpenConns,
		MaxIdleConns: cfg.Frontend.WriterDatabase.MaxIdleConns,
	}, &types.DatabaseConfig{
		Username:     cfg.Frontend.ReaderDatabase.Username,
		Password:     cfg.Frontend.ReaderDatabase.Password,
		Name:         cfg.Frontend.ReaderDatabase.Name,
		Host:         cfg.Frontend.ReaderDatabase.Host,
		Port:         cfg.Frontend.ReaderDatabase.Port,
		MaxOpenConns: cfg.Frontend.ReaderDatabase.MaxOpenConns,
		MaxIdleConns: cfg.Frontend.ReaderDatabase.MaxIdleConns,
	})
	defer db.FrontendReaderDB.Close()
	defer db.FrontendWriterDB.Close()

	chainIDBig := new(big.Int).SetUint64(utils.Config.Chain.Config.DepositChainID)

	rpcClient, err := rpc.NewLighthouseClient("http://"+cfg.Indexer.Node.Host+":"+cfg.Indexer.Node.Port, chainIDBig)
	if err != nil {
		utils.LogFatal(err, "new bigtable lighthouse client error", 0)
	}

	gOuter := errgroup.Group{}
	gOuter.SetLimit(*concurrency)
	for epoch := *start; epoch <= *end; epoch++ {
		epoch := epoch
		gOuter.Go(func() error {
			logrus.Infof("exporting epoch %v", epoch)
			start := time.Now()

			startGetEpochData := time.Now()
			logrus.Printf("retrieving data for epoch %v", epoch)

			data, err := rpcClient.GetEpochData(epoch, false)
			if err != nil {
				logrus.Fatalf("error retrieving epoch data: %v", err)
			}
			logrus.WithFields(logrus.Fields{"duration": time.Since(startGetEpochData), "epoch": epoch}).Info("completed getting epoch-data")
			logrus.Printf("data for epoch %v retrieved, took %v", epoch, time.Since(start))

			if len(data.Validators) == 0 {
				logrus.Fatal("error retrieving epoch data: no validators received for epoch")
			}

			// export epoch data to bigtable
			g := new(errgroup.Group)
			g.SetLimit(6)
			g.Go(func() error {
				err = db.BigtableClient.SaveValidatorBalances(epoch, data.Validators)
				if err != nil {
					return fmt.Errorf("error exporting validator balances to bigtable: %v", err)
				}
				return nil
			})
			g.Go(func() error {
				err = db.BigtableClient.SaveProposalAssignments(epoch, data.ValidatorAssignmentes.ProposerAssignments)
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
			g.Go(func() error {
				err = db.BigtableClient.MigrateIncomeDataV1V2Schema(epoch)
				if err != nil {
					return fmt.Errorf("error exporting sync committee duties to bigtable: %v", err)
				}
				return nil
			})

			err = g.Wait()
			if err != nil {
				return fmt.Errorf("error during bigtable export: %w", err)
			}
			logrus.WithFields(logrus.Fields{"duration": time.Since(start), "epoch": epoch}).Info("completed exporting epoch")
			return nil
		})
	}

	err = gOuter.Wait()
	if err != nil {
		logrus.Fatalf("error during bigtable export: %v", err)
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
			exporter.ExportEpoch(i, rpcClient)
		}
		current = head.HeadEpoch
	}
}
