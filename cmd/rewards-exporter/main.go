package main

import (
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"flag"
	"fmt"
	"time"

	eth_rewards "github.com/gobitfly/eth-rewards"
	"github.com/gobitfly/eth-rewards/beacon"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/sirupsen/logrus"
)

func main() {
	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")
	bnAddress := flag.String("beacon-node-address", "", "Url of the beacon node api")
	enAddress := flag.String("execution-node-address", "", "Url of the execution node api")
	epoch := flag.Int64("epoch", -1, "epoch to export (use -1 to export latest finalized epoch)")

	epochStart := flag.Uint64("epoch-start", 0, "start epoch to export")
	epochEnd := flag.Uint64("epoch-end", 0, "end epoch to export")

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

	client := beacon.NewClient(*bnAddress, time.Minute*5)

	bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.Config.DepositChainID))
	if err != nil {
		logrus.Fatalf("error connecting to bigtable: %v", err)
	}
	defer bt.Close()

	if *epochEnd != 0 {
		for i := *epochStart; i <= *epochEnd; i++ {
			err := export(uint64(i), bt, client, enAddress)
			if err != nil {
				logrus.Fatal(err)
			}
		}
		return
	}
	if *epoch == -1 {
		for {

			notExportedEpochs := []uint64{}
			err = db.WriterDb.Select(&notExportedEpochs, "SELECT epoch FROM epochs WHERE finalized AND NOT rewards_exported ORDER BY epoch")
			if err != nil {
				utils.LogFatal(err, "getting chain head from lighthouse error", 0)
			}
			for _, e := range notExportedEpochs {
				err := export(e, bt, client, enAddress)

				if err != nil {
					logrus.Error(err)
					continue
				}

				_, err = db.WriterDb.Exec("UPDATE epochs SET rewards_exported = true WHERE epoch = $1", e)

				if err != nil {
					logrus.Errorf("error marking rewards_exported as true for epoch %v: %v", e, err)
				}
				services.ReportStatus("rewardsExporter", "Running", nil)
			}

			services.ReportStatus("rewardsExporter", "Running", nil)
			time.Sleep(time.Minute)

		}
	}

	err = export(uint64(*epoch), bt, client, enAddress)
	if err != nil {
		logrus.Fatal(err)
	}
}

func export(epoch uint64, bt *db.Bigtable, client *beacon.Client, elClient *string) error {
	start := time.Now()
	logrus.Infof("retrieving rewards details for epoch %v", epoch)

	rewards, err := eth_rewards.GetRewardsForEpoch(epoch, client, *elClient)

	if err != nil {
		return fmt.Errorf("error retrieving reward details for epoch %v: %v", epoch, err)
	} else {
		logrus.Infof("retrieved %v reward details for epoch %v in %v", len(rewards), epoch, time.Since(start))
	}

	err = bt.SaveValidatorIncomeDetails(uint64(epoch), rewards)
	if err != nil {
		return fmt.Errorf("error saving reward details to bigtable: %v", err)
	}
	return nil
}
