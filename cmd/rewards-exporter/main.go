package main

import (
	"eth2-exporter/cache"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"flag"
	"fmt"
	"sync/atomic"
	"time"

	eth_rewards "github.com/gobitfly/eth-rewards"
	"github.com/gobitfly/eth-rewards/beacon"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func main() {
	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")
	bnAddress := flag.String("beacon-node-address", "", "Url of the beacon node api")
	enAddress := flag.String("execution-node-address", "", "Url of the execution node api")
	epoch := flag.Int64("epoch", -1, "epoch to export (use -1 to export latest finalized epoch)")
	batchConcurrency := flag.Int("batch-concurrency", 5, "epoch to export at the same time (only for historic)")

	epochStart := flag.Uint64("epoch-start", 0, "start epoch to export")
	epochEnd := flag.Uint64("epoch-end", 0, "end epoch to export")
	sleepDuration := flag.Duration("sleep", time.Minute, "duration to sleep between export runs")

	versionFlag := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version.Version)
		fmt.Println(version.GoVersion)
		return
	}

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg
	logrus.WithField("config", *configPath).WithField("version", version.Version).WithField("chainName", utils.Config.Chain.ClConfig.ConfigName).Printf("starting")

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

	if bnAddress == nil || *bnAddress == "" {
		if utils.Config.Indexer.Node.Host == "" {
			utils.LogFatal(nil, "no beacon node url provided", 0)
		} else {
			logrus.Info("applying becon node endpoint from config")
			*bnAddress = fmt.Sprintf("http://%s:%s", utils.Config.Indexer.Node.Host, utils.Config.Indexer.Node.Port)
		}
	}

	if enAddress == nil || *enAddress == "" {
		if utils.Config.Eth1ErigonEndpoint == "" {
			utils.LogFatal(nil, "no execution node url provided", 0)
		} else {
			logrus.Info("applying execution node endpoint from config")
			*enAddress = utils.Config.Eth1ErigonEndpoint
		}
	}

	client := beacon.NewClient(*bnAddress, time.Minute*5)

	bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.ClConfig.DepositChainID), utils.Config.RedisCacheEndpoint)
	if err != nil {
		logrus.Fatalf("error connecting to bigtable: %v", err)
	}
	defer bt.Close()

	cache.MustInitTieredCache(utils.Config.RedisCacheEndpoint)
	logrus.Infof("tiered Cache initialized, latest finalized epoch: %v", services.LatestFinalizedEpoch())

	if *epochEnd != 0 {
		latestFinalizedEpoch := services.LatestFinalizedEpoch()
		if *epochEnd > latestFinalizedEpoch {
			logrus.Errorf("error epochEnd [%v] is greater then latestFinalizedEpoch [%v]", epochEnd, latestFinalizedEpoch)
			return
		}
		g := errgroup.Group{}
		g.SetLimit(*batchConcurrency)

		start := time.Now()
		epochsCompleted := int64(0)
		notExportedEpochs := []uint64{}
		err = db.WriterDb.Select(&notExportedEpochs, "SELECT epoch FROM epochs WHERE NOT rewards_exported AND epoch >= $1 AND epoch <= $2 ORDER BY epoch DESC", *epochStart, *epochEnd)
		if err != nil {
			logrus.Fatal(err)
		}
		epochsToExport := int64(len(notExportedEpochs))

		go func() {
			for {
				c := atomic.LoadInt64(&epochsCompleted)

				if c == 0 {
					time.Sleep(time.Second)
					continue
				}

				epochsRemaining := epochsToExport - c

				elapsed := time.Since(start)
				remaining := time.Duration(epochsRemaining * int64(time.Since(start).Nanoseconds()) / c)
				epochDuration := time.Duration(elapsed.Nanoseconds() / c)

				logrus.Infof("exported %v of %v epochs in %v (%v/epoch), estimated time remaining: %vs", c, epochsToExport, elapsed, epochDuration, remaining)
				time.Sleep(time.Second * 10)
			}
		}()

		for _, e := range notExportedEpochs {
			e := e
			g.Go(func() error {

				var err error
				for i := 0; i < 10; i++ {
					err = export(e, bt, client, enAddress)

					if err != nil {
						logrus.Error(err)
					} else {
						break
					}
				}
				if err != nil {
					logrus.Error(err)
					return nil
				}

				_, err = db.WriterDb.Exec("UPDATE epochs SET rewards_exported = true WHERE epoch = $1", e)

				if err != nil {
					logrus.Errorf("error marking rewards_exported as true for epoch %v: %v", e, err)
				}

				atomic.AddInt64(&epochsCompleted, 1)
				return nil
			})
		}
		g.Wait()
		return
	}

	if *epoch == -1 {
		lastExportedEpoch := uint64(0)
		for {
			latestFinalizedEpoch := services.LatestFinalizedEpoch()
			notExportedEpochs := []uint64{}
			err = db.WriterDb.Select(&notExportedEpochs, "SELECT epoch FROM epochs WHERE NOT rewards_exported AND epoch > $1 AND epoch <= $2 ORDER BY epoch desc LIMIT 10", lastExportedEpoch, latestFinalizedEpoch)
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

				if e > lastExportedEpoch {
					lastExportedEpoch = e
				}
			}

			services.ReportStatus("rewardsExporter", "Running", nil)
			time.Sleep(*sleepDuration)

		}
	}

	latestFinalizedEpoch := services.LatestFinalizedEpoch()
	if *epoch > int64(latestFinalizedEpoch) {
		logrus.Errorf("error epoch [%v] is greater then latestFinalizedEpoch [%v]", epoch, latestFinalizedEpoch)
		return
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
