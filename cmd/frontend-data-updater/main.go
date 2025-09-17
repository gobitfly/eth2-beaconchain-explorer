package main

import (
	"flag"
	"fmt"
	"math/big"
	"sync"

	"github.com/gobitfly/eth2-beaconchain-explorer/cache"
	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/metrics"
	"github.com/gobitfly/eth2-beaconchain-explorer/price"
	"github.com/gobitfly/eth2-beaconchain-explorer/rpc"
	"github.com/gobitfly/eth2-beaconchain-explorer/services"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
	"github.com/gobitfly/eth2-beaconchain-explorer/version"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sirupsen/logrus"

	"net/http"
	_ "net/http/pprof"
)

func main() {
	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")
	versionFlag := flag.Bool("version", false, "Print version and exit")
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

	// enable pprof endpoint if requested
	if utils.Config.Pprof.Enabled {
		go func() {
			logrus.Infof("starting pprof http server on port %s", utils.Config.Pprof.Port)
			logrus.Info(http.ListenAndServe(fmt.Sprintf("localhost:%s", utils.Config.Pprof.Port), nil))
		}()
	}

	if utils.Config.Metrics.Enabled {
		go func(addr string) {
			logrus.Infof("serving metrics on %v", addr)
			if err := metrics.Serve(addr); err != nil {
				logrus.WithError(err).Fatal("Error serving metrics")
			}
		}(utils.Config.Metrics.Address)
	}

	_, err = db.InitBigtable(cfg.Bigtable.Project, cfg.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.ClConfig.DepositChainID), utils.Config.RedisCacheEndpoint)
	if err != nil {
		logrus.Fatalf("error initializing bigtable %v", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		db.MustInitFrontendDB(&types.DatabaseConfig{
			Username:     cfg.Frontend.WriterDatabase.Username,
			Password:     cfg.Frontend.WriterDatabase.Password,
			Name:         cfg.Frontend.WriterDatabase.Name,
			Host:         cfg.Frontend.WriterDatabase.Host,
			Port:         cfg.Frontend.WriterDatabase.Port,
			MaxOpenConns: cfg.Frontend.WriterDatabase.MaxOpenConns,
			MaxIdleConns: cfg.Frontend.WriterDatabase.MaxIdleConns,
			SSL:          cfg.Frontend.WriterDatabase.SSL,
		}, &types.DatabaseConfig{
			Username:     cfg.Frontend.ReaderDatabase.Username,
			Password:     cfg.Frontend.ReaderDatabase.Password,
			Name:         cfg.Frontend.ReaderDatabase.Name,
			Host:         cfg.Frontend.ReaderDatabase.Host,
			Port:         cfg.Frontend.ReaderDatabase.Port,
			MaxOpenConns: cfg.Frontend.ReaderDatabase.MaxOpenConns,
			MaxIdleConns: cfg.Frontend.ReaderDatabase.MaxIdleConns,
			SSL:          cfg.Frontend.ReaderDatabase.SSL,
		}, "pgx", "postgres")
	}()
	wg.Wait()
	defer db.FrontendReaderDB.Close()
	defer db.FrontendWriterDB.Close()

	db.MustInitDB(&types.DatabaseConfig{
		Username:     cfg.WriterDatabase.Username,
		Password:     cfg.WriterDatabase.Password,
		Name:         cfg.WriterDatabase.Name,
		Host:         cfg.WriterDatabase.Host,
		Port:         cfg.WriterDatabase.Port,
		MaxOpenConns: cfg.WriterDatabase.MaxOpenConns,
		MaxIdleConns: cfg.WriterDatabase.MaxIdleConns,
		SSL:          cfg.WriterDatabase.SSL,
	}, &types.DatabaseConfig{
		Username:     cfg.ReaderDatabase.Username,
		Password:     cfg.ReaderDatabase.Password,
		Name:         cfg.ReaderDatabase.Name,
		Host:         cfg.ReaderDatabase.Host,
		Port:         cfg.ReaderDatabase.Port,
		MaxOpenConns: cfg.ReaderDatabase.MaxOpenConns,
		MaxIdleConns: cfg.ReaderDatabase.MaxIdleConns,
		SSL:          cfg.ReaderDatabase.SSL,
	}, "pgx", "postgres")
	defer db.ReaderDb.Close()
	defer db.WriterDb.Close()

	if utils.Config.TieredCacheProvider == "redis" || len(utils.Config.RedisCacheEndpoint) != 0 {
		cache.MustInitTieredCache(utils.Config.RedisCacheEndpoint)
	}

	if utils.Config.TieredCacheProvider != "redis" {
		logrus.Fatalf("No cache provider set. Please set TierdCacheProvider (example redis, bigtable)")
	}

	logrus.Infof("initializing prices")
	price.Init(utils.Config.Chain.ClConfig.DepositChainID, utils.Config.Eth1ErigonEndpoint, utils.Config.Frontend.ClCurrency, utils.Config.Frontend.ElCurrency)

	chainID := new(big.Int).SetUint64(utils.Config.Chain.ClConfig.DepositChainID)
	rpcClient, err := rpc.NewLighthouseClient("http://"+cfg.Indexer.Node.Host+":"+cfg.Indexer.Node.Port, chainID)
	if err != nil {
		utils.LogFatal(err, "new explorer lighthouse client error", 0)
	}
	rpc.CurrentClient = rpcClient

	logrus.Infof("initializing frontend services")
	services.Init() // Init frontend services
	logrus.Infof("frontend services initiated")

	utils.WaitForCtrlC()

	logrus.Println("exiting...")
}
