package main

import (
	"context"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"eth2-exporter/cache"
	"eth2-exporter/db"
	ethclients "eth2-exporter/ethClients"
	"eth2-exporter/exporter"
	"eth2-exporter/handlers"
	"eth2-exporter/metrics"
	"eth2-exporter/price"
	"eth2-exporter/rpc"
	"eth2-exporter/services"
	"eth2-exporter/static"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"flag"
	"fmt"
	"math/big"
	"net/http"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/sirupsen/logrus"

	_ "eth2-exporter/docs"
	_ "net/http/pprof"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/stripe/stripe-go/v72"
	"github.com/urfave/negroni"
	"github.com/zesik/proxyaddr"
)

func initStripe(router *mux.Router) error {
	if utils.Config == nil {
		return fmt.Errorf("error no config found")
	}
	stripe.Key = utils.Config.Frontend.Stripe.SecretKey
	router.Handle("/stripe/create-checkout-session", recoverPanicWrap(http.HandlerFunc(handlers.StripeCreateCheckoutSession))).Methods("POST")
	router.Handle("/stripe/customer-portal", recoverPanicWrap(http.HandlerFunc(handlers.StripeCustomerPortal))).Methods("POST")
	return nil
}

func init() {
	gob.Register(types.DataTableSaveState{})
}

func main() {
	defer recoverPanic()
	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")

	flag.Parse()

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg
	logrus.WithField("config", *configPath).WithField("version", version.Version).WithField("chainName", utils.Config.Chain.Config.ConfigName).Printf("starting")

	if utils.Config.Chain.Config.SlotsPerEpoch == 0 || utils.Config.Chain.Config.SecondsPerSlot == 0 {
		logrus.Fatal("invalid chain configuration specified, you must specify the slots per epoch, seconds per slot and genesis timestamp in the config file")
	}

	err = handlers.CheckAndPreloadImprint()
	if err != nil {
		logrus.Fatalf("error check / preload imprint: %v", err)
	}

	if utils.Config.Pprof.Enabled {
		go func() {
			logrus.Infof("starting pprof http server on port %s", utils.Config.Pprof.Port)
			logrus.Info(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", utils.Config.Pprof.Port), nil))
		}()
	}

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
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
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		db.MustInitFrontendDB(&types.DatabaseConfig{
			Username: cfg.Frontend.WriterDatabase.Username,
			Password: cfg.Frontend.WriterDatabase.Password,
			Name:     cfg.Frontend.WriterDatabase.Name,
			Host:     cfg.Frontend.WriterDatabase.Host,
			Port:     cfg.Frontend.WriterDatabase.Port,
		}, &types.DatabaseConfig{
			Username: cfg.Frontend.ReaderDatabase.Username,
			Password: cfg.Frontend.ReaderDatabase.Password,
			Name:     cfg.Frontend.ReaderDatabase.Name,
			Host:     cfg.Frontend.ReaderDatabase.Host,
			Port:     cfg.Frontend.ReaderDatabase.Port,
		})
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		rpc.CurrentErigonClient, err = rpc.NewErigonClient(utils.Config.Eth1ErigonEndpoint)
		if err != nil {
			logrus.Fatalf("error initializing erigon client: %v", err)
		}

		erigonChainId, err := rpc.CurrentErigonClient.GetNativeClient().ChainID(ctx)
		if err != nil {
			logrus.Fatalf("error retrieving erigon chain id: %v", err)
		}

		rpc.CurrentGethClient, err = rpc.NewGethClient(utils.Config.Eth1GethEndpoint)
		if err != nil {
			logrus.Fatalf("error initializing geth client: %v", err)
		}

		gethChainId, err := rpc.CurrentGethClient.GetNativeClient().ChainID(ctx)
		if err != nil {
			logrus.Fatalf("error retrieving geth chain id: %v", err)
		}

		if !(erigonChainId.String() == gethChainId.String() && erigonChainId.String() == fmt.Sprintf("%d", utils.Config.Chain.Config.DepositChainID)) {
			logrus.Fatalf("chain id missmatch: erigon chain id %v, geth chain id %v, requested chain id %v", erigonChainId.String(), erigonChainId.String(), fmt.Sprintf("%d", utils.Config.Chain.Config.DepositChainID))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.Config.DepositChainID)) //
		if err != nil {
			logrus.Fatalf("error connecting to bigtable: %v", err)
		}
		db.BigtableClient = bt
	}()

	if utils.Config.TieredCacheProvider == "redis" || len(utils.Config.RedisCacheEndpoint) != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.MustInitTieredCache(utils.Config.RedisCacheEndpoint)
			logrus.Infof("Tiered Cache initialized. Latest finalized epoch: %v", services.LatestFinalizedEpoch())

		}()
	}

	wg.Wait()
	if utils.Config.TieredCacheProvider == "bigtable" && len(utils.Config.RedisCacheEndpoint) == 0 {
		cache.MustInitTieredCacheBigtable(db.BigtableClient.GetClient(), fmt.Sprintf("%d", utils.Config.Chain.Config.DepositChainID))
		logrus.Infof("Tiered Cache initialized. Latest finalized epoch: %v", services.LatestFinalizedEpoch())
	}

	if utils.Config.TieredCacheProvider != "bigtable" && utils.Config.TieredCacheProvider != "redis" {
		logrus.Fatalf("No cache provider set. Please set TierdCacheProvider (example redis, bigtable)")
	}

	defer db.ReaderDb.Close()
	defer db.WriterDb.Close()
	defer db.FrontendReaderDB.Close()
	defer db.FrontendWriterDB.Close()
	defer db.BigtableClient.Close()

	if utils.Config.Metrics.Enabled {
		go metrics.MonitorDB(db.WriterDb)
		DBStr := fmt.Sprintf("%v-%v-%v-%v-%v", cfg.WriterDatabase.Username, cfg.WriterDatabase.Password, cfg.WriterDatabase.Host, cfg.WriterDatabase.Port, cfg.WriterDatabase.Name)
		frontendDBStr := fmt.Sprintf("%v-%v-%v-%v-%v", cfg.Frontend.WriterDatabase.Username, cfg.Frontend.WriterDatabase.Password, cfg.Frontend.WriterDatabase.Host, cfg.Frontend.WriterDatabase.Port, cfg.Frontend.WriterDatabase.Name)
		if DBStr != frontendDBStr {
			go metrics.MonitorDB(db.FrontendWriterDB)
		}
	}

	logrus.Infof("database connection established")

	if utils.Config.Indexer.Enabled {

		err = services.InitLastAttestationCache(utils.Config.LastAttestationCachePath)

		if err != nil {
			logrus.Fatalf("error initializing last attesation cache: %v", err)
		}

		var rpcClient rpc.Client

		chainID := new(big.Int).SetUint64(utils.Config.Chain.Config.DepositChainID)
		if utils.Config.Indexer.Node.Type == "lighthouse" {
			rpcClient, err = rpc.NewLighthouseClient("http://"+cfg.Indexer.Node.Host+":"+cfg.Indexer.Node.Port, chainID)
			if err != nil {
				logrus.Fatal(err)
			}
		} else {
			logrus.Fatalf("invalid note type %v specified. supported node types are prysm and lighthouse", utils.Config.Indexer.Node.Type)
		}

		if utils.Config.Indexer.OneTimeExport.Enabled {
			if len(utils.Config.Indexer.OneTimeExport.Epochs) > 0 {
				logrus.Infof("onetimeexport epochs: %+v", utils.Config.Indexer.OneTimeExport.Epochs)
				for _, epoch := range utils.Config.Indexer.OneTimeExport.Epochs {
					err := exporter.ExportEpoch(epoch, rpcClient)
					if err != nil {
						logrus.Fatal(err)
					}
				}
			} else {
				logrus.Infof("onetimeexport epochs: %v-%v", utils.Config.Indexer.OneTimeExport.StartEpoch, utils.Config.Indexer.OneTimeExport.EndEpoch)
				for epoch := utils.Config.Indexer.OneTimeExport.StartEpoch; epoch <= utils.Config.Indexer.OneTimeExport.EndEpoch; epoch++ {
					err := exporter.ExportEpoch(epoch, rpcClient)
					if err != nil {
						logrus.Fatal(err)
					}
				}
			}
			return
		}

		go services.StartHistoricPriceService()
		go exporter.Start(rpcClient)
	}

	if cfg.Frontend.Enabled {

		if cfg.Frontend.OnlyAPI {
			services.ReportStatus("api", "Running", nil)
		} else {
			services.ReportStatus("frontend", "Running", nil)
		}

		router := mux.NewRouter()

		apiV1Router := router.PathPrefix("/api/v1").Subrouter()
		router.PathPrefix("/api/v1/docs/").Handler(recoverPanicWrap(http.HandlerFunc(httpSwagger.WrapHandler)))
		apiV1Router.Handle("/epoch/{epoch}", recoverPanicWrap(http.HandlerFunc(handlers.ApiEpoch))).Methods("GET", "OPTIONS")

		apiV1Router.Handle("/epoch/{epoch}/blocks", recoverPanicWrap(http.HandlerFunc(handlers.ApiEpochSlots))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/epoch/{epoch}/slots", recoverPanicWrap(http.HandlerFunc(handlers.ApiEpochSlots))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/slot/{slotOrHash}", recoverPanicWrap(http.HandlerFunc(handlers.ApiSlots))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/slot/{slot}/attestations", recoverPanicWrap(http.HandlerFunc(handlers.ApiSlotAttestations))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/slot/{slot}/deposits", recoverPanicWrap(http.HandlerFunc(handlers.ApiSlotDeposits))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/slot/{slot}/attesterslashings", recoverPanicWrap(http.HandlerFunc(handlers.ApiSlotAttesterSlashings))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/slot/{slot}/proposerslashings", recoverPanicWrap(http.HandlerFunc(handlers.ApiSlotProposerSlashings))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/slot/{slot}/voluntaryexits", recoverPanicWrap(http.HandlerFunc(handlers.ApiSlotVoluntaryExits))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/slot/{slot}/withdrawals", recoverPanicWrap(http.HandlerFunc(handlers.ApiSlotWithdrawals))).Methods("GET", "OPTIONS")

		// deprecated, use slot equivalents
		apiV1Router.Handle("/block/{slotOrHash}", recoverPanicWrap(http.HandlerFunc(handlers.ApiSlots))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/block/{slot}/attestations", recoverPanicWrap(http.HandlerFunc(handlers.ApiSlotAttestations))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/block/{slot}/deposits", recoverPanicWrap(http.HandlerFunc(handlers.ApiSlotDeposits))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/block/{slot}/attesterslashings", recoverPanicWrap(http.HandlerFunc(handlers.ApiSlotAttesterSlashings))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/block/{slot}/proposerslashings", recoverPanicWrap(http.HandlerFunc(handlers.ApiSlotProposerSlashings))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/block/{slot}/voluntaryexits", recoverPanicWrap(http.HandlerFunc(handlers.ApiSlotVoluntaryExits))).Methods("GET", "OPTIONS")

		apiV1Router.Handle("/sync_committee/{period}", recoverPanicWrap(http.HandlerFunc(handlers.ApiSyncCommittee))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/eth1deposit/{txhash}", recoverPanicWrap(http.HandlerFunc(handlers.ApiEth1Deposit))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/validator/leaderboard", recoverPanicWrap(http.HandlerFunc(handlers.ApiValidatorLeaderboard))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/validator/{indexOrPubkey}", recoverPanicWrap(http.HandlerFunc(handlers.ApiValidator))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/validator/{indexOrPubkey}/withdrawals", recoverPanicWrap(http.HandlerFunc(handlers.ApiValidatorWithdrawals))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/validator/{indexOrPubkey}/blsChange", recoverPanicWrap(http.HandlerFunc(handlers.ApiValidatorBlsChange))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/validator/{indexOrPubkey}/balancehistory", recoverPanicWrap(http.HandlerFunc(handlers.ApiValidatorBalanceHistory))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/validator/{indexOrPubkey}/incomedetailhistory", recoverPanicWrap(http.HandlerFunc(handlers.ApiValidatorIncomeDetailsHistory))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/validator/{indexOrPubkey}/performance", recoverPanicWrap(http.HandlerFunc(handlers.ApiValidatorPerformance))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/validator/{indexOrPubkey}/execution/performance", recoverPanicWrap(http.HandlerFunc(handlers.ApiValidatorExecutionPerformance))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/validator/{indexOrPubkey}/attestations", recoverPanicWrap(http.HandlerFunc(handlers.ApiValidatorAttestations))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/validator/{indexOrPubkey}/proposals", recoverPanicWrap(http.HandlerFunc(handlers.ApiValidatorProposals))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/validator/{indexOrPubkey}/deposits", recoverPanicWrap(http.HandlerFunc(handlers.ApiValidatorDeposits))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/validator/{indexOrPubkey}/attestationefficiency", recoverPanicWrap(http.HandlerFunc(handlers.ApiValidatorAttestationEfficiency))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/validator/{indexOrPubkey}/attestationeffectiveness", recoverPanicWrap(http.HandlerFunc(handlers.ApiValidatorAttestationEffectiveness))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/validator/stats/{index}", recoverPanicWrap(http.HandlerFunc(handlers.ApiValidatorDailyStats))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/validator/eth1/{address}", recoverPanicWrap(http.HandlerFunc(handlers.ApiValidatorByEth1Address))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/validators/queue", recoverPanicWrap(http.HandlerFunc(handlers.ApiValidatorQueue))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/graffitiwall", recoverPanicWrap(http.HandlerFunc(handlers.ApiGraffitiwall))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/chart/{chart}", recoverPanicWrap(http.HandlerFunc(handlers.ApiChart))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/user/token", recoverPanicWrap(http.HandlerFunc(handlers.APIGetToken))).Methods("POST", "OPTIONS")
		apiV1Router.Handle("/dashboard/data/allbalances", recoverPanicWrap(http.HandlerFunc(handlers.DashboardDataBalanceCombined))).Methods("GET", "OPTIONS") // consensus & execution
		apiV1Router.Handle("/dashboard/data/balances", recoverPanicWrap(http.HandlerFunc(handlers.DashboardDataBalance))).Methods("GET", "OPTIONS")            // new app versions
		apiV1Router.Handle("/dashboard/data/balance", recoverPanicWrap(http.HandlerFunc(handlers.APIDashboardDataBalance))).Methods("GET", "OPTIONS")          // old app versions
		apiV1Router.Handle("/dashboard/data/proposals", recoverPanicWrap(http.HandlerFunc(handlers.DashboardDataProposals))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/stripe/webhook", recoverPanicWrap(http.HandlerFunc(handlers.StripeWebhook))).Methods("POST")
		apiV1Router.Handle("/stats/{apiKey}/{machine}", recoverPanicWrap(http.HandlerFunc(handlers.ClientStatsPostOld))).Methods("POST", "OPTIONS")
		apiV1Router.Handle("/stats/{apiKey}", recoverPanicWrap(http.HandlerFunc(handlers.ClientStatsPostOld))).Methods("POST", "OPTIONS")
		apiV1Router.Handle("/client/metrics", recoverPanicWrap(http.HandlerFunc(handlers.ClientStatsPostNew))).Methods("POST", "OPTIONS")
		apiV1Router.Handle("/app/dashboard", recoverPanicWrap(http.HandlerFunc(handlers.ApiDashboard))).Methods("POST", "OPTIONS")
		apiV1Router.Handle("/rocketpool/stats", recoverPanicWrap(http.HandlerFunc(handlers.ApiRocketpoolStats))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/rocketpool/validator/{indexOrPubkey}", recoverPanicWrap(http.HandlerFunc(handlers.ApiRocketpoolValidators))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/ethstore/{day}", recoverPanicWrap(http.HandlerFunc(handlers.ApiEthStoreDay))).Methods("GET", "OPTIONS")

		apiV1Router.Handle("/execution/gasnow", recoverPanicWrap(http.HandlerFunc(handlers.ApiEth1GasNowData))).Methods("GET", "OPTIONS")
		// query params: token
		apiV1Router.Handle("/execution/block/{blockNumber}", recoverPanicWrap(http.HandlerFunc(handlers.ApiETH1ExecBlocks))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/execution/{addressIndexOrPubkey}/produced", recoverPanicWrap(http.HandlerFunc(handlers.ApiETH1AccountProducedBlocks))).Methods("GET", "OPTIONS")

		apiV1Router.Handle("/execution/address/{address}", recoverPanicWrap(http.HandlerFunc(handlers.ApiEth1Address))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/execution/address/{address}/transactions", recoverPanicWrap(http.HandlerFunc(handlers.ApiEth1AddressTx))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/execution/address/{address}/internalTx", recoverPanicWrap(http.HandlerFunc(handlers.ApiEth1AddressItx))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/execution/address/{address}/blocks", recoverPanicWrap(http.HandlerFunc(handlers.ApiEth1AddressBlocks))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/execution/address/{address}/uncles", recoverPanicWrap(http.HandlerFunc(handlers.ApiEth1AddressUncles))).Methods("GET", "OPTIONS")
		apiV1Router.Handle("/execution/address/{address}/tokens", recoverPanicWrap(http.HandlerFunc(handlers.ApiEth1AddressTokens))).Methods("GET", "OPTIONS")
		// // query params: type={erc20,erc721,erc1155}, address

		// apiV1Router.Handle("/execution/transactions", recoverPanicWrap(http.HandlerFunc(handlers.ApiEth1Tx))).Methods("GET", "OPTIONS")
		// apiV1Router.Handle("/execution/transaction/{txhash}/itx", recoverPanicWrap(http.HandlerFunc(handlers.ApiEth1TxItx))).Methods("GET", "OPTIONS")
		// apiV1Router.Handle("/execution/transaction/{txhash}/status", recoverPanicWrap(http.HandlerFunc(handlers.ApiEth1TxStatus))).Methods("GET", "OPTIONS")
		// apiV1Router.Handle("/execution/token/{token}", recoverPanicWrap(http.HandlerFunc(handlers.ApiEth1))).Methods("GET", "OPTIONS")
		// apiV1Router.Handle("/stats/overall/epoch/{epoch}/rewards", recoverPanicWrap(http.HandlerFunc(handlers.ApiEth1))).Methods("GET", "OPTIONS")
		// apiV1Router.Handle("/stats/overall/daily/eth-price?offset={timestamp}&limit={limit}&order={order}", recoverPanicWrap(http.HandlerFunc(handlers.ApiEth1))).Methods("GET", "OPTIONS")
		// apiV1Router.Handle("/stats/execution/blocksize?offset={timestamp}&limit={limit}&order={order}", recoverPanicWrap(http.HandlerFunc(handlers.ApiEth1))).Methods("GET", "OPTIONS")
		// apiV1Router.Handle("/stats/execution/daily/avg-gas-limit?offset={timestamp}&limit={limit}&order={order} OR ?timestamp={timestamp}", recoverPanicWrap(http.HandlerFunc(handlers.ApiEth1))).Methods("GET", "OPTIONS")
		// apiV1Router.Handle("/stats/execution/daily/gas-used?offset={timestamp}&limit={limit}&order={order} OR ?timestamp={timestamp}", recoverPanicWrap(http.HandlerFunc(handlers.ApiEth1))).Methods("GET", "OPTIONS")
		// apiV1Router.Handle("/stats/execution/gas-orcale", recoverPanicWrap(http.HandlerFunc(handlers.ApiEth1))).Methods("GET", "OPTIONS")
		// apiV1Router.Handle("/stats/token/{token}/supply?block={block}", recoverPanicWrap(http.HandlerFunc(handlers.ApiEth1))).Methods("GET", "OPTIONS")
		// apiV1Router.Handle("/utils/execution/publish-txn?raw={txndata}", recoverPanicWrap(http.HandlerFunc(handlers.ApiEth1))).Methods("GET", "OPTIONS")

		// apiV1Router.Handle("/execution/block/{blockNumber}", recoverPanicWrap(http.HandlerFunc(handlers.APIETH1))).Methods("GET", "OPTIONS")

		apiV1Router.Handle("/validator/{indexOrPubkey}/widget", recoverPanicWrap(http.HandlerFunc(handlers.GetMobileWidgetStatsGet))).Methods("GET")
		apiV1Router.Handle("/dashboard/widget", recoverPanicWrap(http.HandlerFunc(handlers.GetMobileWidgetStatsPost))).Methods("POST")
		apiV1Router.Use(utils.CORSMiddleware)

		apiV1AuthRouter := apiV1Router.PathPrefix("/user").Subrouter()
		apiV1AuthRouter.Handle("/mobile/notify/register", recoverPanicWrap(http.HandlerFunc(handlers.MobileNotificationUpdatePOST))).Methods("POST", "OPTIONS")
		apiV1AuthRouter.Handle("/mobile/settings", recoverPanicWrap(http.HandlerFunc(handlers.MobileDeviceSettings))).Methods("GET", "OPTIONS")
		apiV1AuthRouter.Handle("/mobile/settings", recoverPanicWrap(http.HandlerFunc(handlers.MobileDeviceSettingsPOST))).Methods("POST", "OPTIONS")
		apiV1AuthRouter.Handle("/validator/saved", recoverPanicWrap(http.HandlerFunc(handlers.MobileTagedValidators))).Methods("GET", "OPTIONS")
		apiV1AuthRouter.Handle("/subscription/register", recoverPanicWrap(http.HandlerFunc(handlers.RegisterMobileSubscriptions))).Methods("POST", "OPTIONS")

		apiV1AuthRouter.Handle("/validator/{pubkey}/add", recoverPanicWrap(http.HandlerFunc(handlers.UserValidatorWatchlistAdd))).Methods("POST", "OPTIONS")
		apiV1AuthRouter.Handle("/validator/{pubkey}/remove", recoverPanicWrap(http.HandlerFunc(handlers.UserValidatorWatchlistRemove))).Methods("POST", "OPTIONS")
		apiV1AuthRouter.Handle("/dashboard/save", recoverPanicWrap(http.HandlerFunc(handlers.UserDashboardWatchlistAdd))).Methods("POST", "OPTIONS")
		apiV1AuthRouter.Handle("/notifications/bundled/subscribe", recoverPanicWrap(http.HandlerFunc(handlers.MultipleUsersNotificationsSubscribe))).Methods("POST", "OPTIONS")
		apiV1AuthRouter.Handle("/notifications/bundled/unsubscribe", recoverPanicWrap(http.HandlerFunc(handlers.MultipleUsersNotificationsUnsubscribe))).Methods("POST", "OPTIONS")
		apiV1AuthRouter.Handle("/notifications/subscribe", recoverPanicWrap(http.HandlerFunc(handlers.UserNotificationsSubscribe))).Methods("POST", "OPTIONS")
		apiV1AuthRouter.Handle("/notifications/unsubscribe", recoverPanicWrap(http.HandlerFunc(handlers.UserNotificationsUnsubscribe))).Methods("POST", "OPTIONS")
		apiV1AuthRouter.Handle("/notifications", recoverPanicWrap(http.HandlerFunc(handlers.UserNotificationsSubscribed))).Methods("POST", "GET", "OPTIONS")
		apiV1AuthRouter.Handle("/stats", recoverPanicWrap(http.HandlerFunc(handlers.ClientStats))).Methods("GET", "OPTIONS")
		apiV1AuthRouter.Handle("/stats/{offset}/{limit}", recoverPanicWrap(http.HandlerFunc(handlers.ClientStats))).Methods("GET", "OPTIONS")
		apiV1AuthRouter.Handle("/ethpool", recoverPanicWrap(http.HandlerFunc(handlers.RegisterEthpoolSubscription))).Methods("POST", "OPTIONS")

		apiV1AuthRouter.Use(utils.CORSMiddleware)
		apiV1AuthRouter.Use(utils.AuthorizedAPIMiddleware)

		router.Handle("/api/healthz", recoverPanicWrap(http.HandlerFunc(handlers.ApiHealthz))).Methods("GET", "HEAD")
		router.Handle("/api/healthz-loadbalancer", recoverPanicWrap(http.HandlerFunc(handlers.ApiHealthzLoadbalancer))).Methods("GET", "HEAD")

		// logrus.Infof("initializing frontend services")
		// services.Init() // Init frontend services
		// logrus.Infof("frontend services initiated")

		logrus.Infof("initializing prices")
		price.Init(utils.Config.Chain.Config.DepositChainID)
		logrus.Infof("prices initialized")
		if !utils.Config.Frontend.Debug {
			logrus.Infof("initializing ethclients")
			ethclients.Init()
			logrus.Infof("ethclients initialized")
		}

		if !utils.Config.Frontend.OnlyAPI {
			if utils.Config.Frontend.SiteDomain == "" {
				utils.Config.Frontend.SiteDomain = "beaconcha.in"
			}

			logrus.Infof("frontend database connection established")

			utils.InitSessionStore(cfg.Frontend.SessionSecret)

			csrfBytes, err := hex.DecodeString(cfg.Frontend.CsrfAuthKey)
			if err != nil {
				logrus.WithError(err).Error("error decoding csrf auth key falling back to empty csrf key")
			}
			csrfHandler := csrf.Protect(
				csrfBytes,
				csrf.FieldName("CsrfField"),
				csrf.Secure(!cfg.Frontend.CsrfInsecure),
				csrf.Path("/"),
			)

			router.Handle("/", recoverPanicWrap(http.HandlerFunc(handlers.Index))).Methods("GET")
			router.Handle("/latestState", recoverPanicWrap(http.HandlerFunc(handlers.LatestState))).Methods("GET")
			router.Handle("/launchMetrics", recoverPanicWrap(http.HandlerFunc(handlers.SlotVizMetrics))).Methods("GET")
			router.Handle("/index/data", recoverPanicWrap(http.HandlerFunc(handlers.IndexPageData))).Methods("GET")
			router.Handle("/slot/{slotOrHash}", recoverPanicWrap(http.HandlerFunc(handlers.Slot))).Methods("GET")
			router.Handle("/slot/{slotOrHash}/deposits", recoverPanicWrap(http.HandlerFunc(handlers.SlotDepositData))).Methods("GET")
			router.Handle("/slot/{slotOrHash}/votes", recoverPanicWrap(http.HandlerFunc(handlers.SlotVoteData))).Methods("GET")
			router.Handle("/slot/{slot}/attestations", recoverPanicWrap(http.HandlerFunc(handlers.SlotAttestationsData))).Methods("GET")
			router.Handle("/slot/{slot}/withdrawals", recoverPanicWrap(http.HandlerFunc(handlers.SlotWithdrawalData))).Methods("GET")
			router.Handle("/slot/{slot}/blsChange", recoverPanicWrap(http.HandlerFunc(handlers.SlotBlsChangeData))).Methods("GET")
			router.Handle("/slots", recoverPanicWrap(http.HandlerFunc(handlers.Slots))).Methods("GET")
			router.Handle("/slots/data", recoverPanicWrap(http.HandlerFunc(handlers.SlotsData))).Methods("GET")
			router.Handle("/blocks", recoverPanicWrap(http.HandlerFunc(handlers.Eth1Blocks))).Methods("GET")
			router.Handle("/blocks/data", recoverPanicWrap(http.HandlerFunc(handlers.Eth1BlocksData))).Methods("GET")
			router.Handle("/blocks/highest", recoverPanicWrap(http.HandlerFunc(handlers.Eth1BlocksHighest))).Methods("GET")
			router.Handle("/address/{address}", recoverPanicWrap(http.HandlerFunc(handlers.Eth1Address))).Methods("GET")
			router.Handle("/address/{address}/blocks", recoverPanicWrap(http.HandlerFunc(handlers.Eth1AddressBlocksMined))).Methods("GET")
			router.Handle("/address/{address}/uncles", recoverPanicWrap(http.HandlerFunc(handlers.Eth1AddressUnclesMined))).Methods("GET")
			router.Handle("/address/{address}/withdrawals", recoverPanicWrap(http.HandlerFunc(handlers.Eth1AddressWithdrawals))).Methods("GET")
			router.Handle("/address/{address}/transactions", recoverPanicWrap(http.HandlerFunc(handlers.Eth1AddressTransactions))).Methods("GET")
			router.Handle("/address/{address}/internalTxns", recoverPanicWrap(http.HandlerFunc(handlers.Eth1AddressInternalTransactions))).Methods("GET")
			router.Handle("/address/{address}/erc20", recoverPanicWrap(http.HandlerFunc(handlers.Eth1AddressErc20Transactions))).Methods("GET")
			router.Handle("/address/{address}/erc721", recoverPanicWrap(http.HandlerFunc(handlers.Eth1AddressErc721Transactions))).Methods("GET")
			router.Handle("/address/{address}/erc1155", recoverPanicWrap(http.HandlerFunc(handlers.Eth1AddressErc1155Transactions))).Methods("GET")
			router.Handle("/token/{token}", recoverPanicWrap(http.HandlerFunc(handlers.Eth1Token))).Methods("GET")
			router.Handle("/token/{token}/transfers", recoverPanicWrap(http.HandlerFunc(handlers.Eth1TokenTransfers))).Methods("GET")
			router.Handle("/transactions", recoverPanicWrap(http.HandlerFunc(handlers.Eth1Transactions))).Methods("GET")
			router.Handle("/transactions/data", recoverPanicWrap(http.HandlerFunc(handlers.Eth1TransactionsData))).Methods("GET")
			router.Handle("/block/{block}", recoverPanicWrap(http.HandlerFunc(handlers.Eth1Block))).Methods("GET")
			router.Handle("/block/{block}/transactions", recoverPanicWrap(http.HandlerFunc(handlers.BlockTransactionsData))).Methods("GET")
			router.Handle("/tx/{hash}", recoverPanicWrap(http.HandlerFunc(handlers.Eth1TransactionTx))).Methods("GET")
			router.Handle("/mempool", recoverPanicWrap(http.HandlerFunc(handlers.MempoolView))).Methods("GET")
			router.Handle("/burn", recoverPanicWrap(http.HandlerFunc(handlers.Burn))).Methods("GET")
			router.Handle("/burn/data", recoverPanicWrap(http.HandlerFunc(handlers.BurnPageData))).Methods("GET")
			router.Handle("/gasnow", recoverPanicWrap(http.HandlerFunc(handlers.GasNow))).Methods("GET")
			router.Handle("/gasnow/data", recoverPanicWrap(http.HandlerFunc(handlers.GasNowData))).Methods("GET")
			router.Handle("/correlations", recoverPanicWrap(http.HandlerFunc(handlers.Correlations))).Methods("GET")
			router.Handle("/correlations/data", recoverPanicWrap(http.HandlerFunc(handlers.CorrelationsData))).Methods("POST")

			router.Handle("/vis", recoverPanicWrap(http.HandlerFunc(handlers.Vis))).Methods("GET")
			router.Handle("/charts", recoverPanicWrap(http.HandlerFunc(handlers.Charts))).Methods("GET")
			router.Handle("/charts/{chart}", recoverPanicWrap(http.HandlerFunc(handlers.Chart))).Methods("GET")
			router.Handle("/charts/{chart}/data", recoverPanicWrap(http.HandlerFunc(handlers.GenericChartData))).Methods("GET")
			router.Handle("/vis/blocks", recoverPanicWrap(http.HandlerFunc(handlers.VisBlocks))).Methods("GET")
			router.Handle("/vis/votes", recoverPanicWrap(http.HandlerFunc(handlers.VisVotes))).Methods("GET")
			router.Handle("/epoch/{epoch}", recoverPanicWrap(http.HandlerFunc(handlers.Epoch))).Methods("GET")
			router.Handle("/epochs", recoverPanicWrap(http.HandlerFunc(handlers.Epochs))).Methods("GET")
			router.Handle("/epochs/data", recoverPanicWrap(http.HandlerFunc(handlers.EpochsData))).Methods("GET")

			router.Handle("/validator/{index}", recoverPanicWrap(http.HandlerFunc(handlers.Validator))).Methods("GET")
			router.Handle("/validator/{index}/proposedblocks", recoverPanicWrap(http.HandlerFunc(handlers.ValidatorProposedBlocks))).Methods("GET")
			router.Handle("/validator/{index}/attestations", recoverPanicWrap(http.HandlerFunc(handlers.ValidatorAttestations))).Methods("GET")
			router.Handle("/validator/{index}/withdrawals", recoverPanicWrap(http.HandlerFunc(handlers.ValidatorWithdrawals))).Methods("GET")
			router.Handle("/validator/{index}/sync", recoverPanicWrap(http.HandlerFunc(handlers.ValidatorSync))).Methods("GET")
			router.Handle("/validator/{index}/history", recoverPanicWrap(http.HandlerFunc(handlers.ValidatorHistory))).Methods("GET")
			router.Handle("/validator/{pubkey}/deposits", recoverPanicWrap(http.HandlerFunc(handlers.ValidatorDeposits))).Methods("GET")
			router.Handle("/validator/{index}/slashings", recoverPanicWrap(http.HandlerFunc(handlers.ValidatorSlashings))).Methods("GET")
			router.Handle("/validator/{index}/effectiveness", recoverPanicWrap(http.HandlerFunc(handlers.ValidatorAttestationInclusionEffectiveness))).Methods("GET")
			router.Handle("/validator/{pubkey}/save", recoverPanicWrap(http.HandlerFunc(handlers.ValidatorSave))).Methods("POST")
			router.Handle("/watchlist/add", recoverPanicWrap(http.HandlerFunc(handlers.UsersModalAddValidator))).Methods("POST")
			router.Handle("/validator/{pubkey}/remove", recoverPanicWrap(http.HandlerFunc(handlers.UserValidatorWatchlistRemove))).Methods("POST")
			router.Handle("/validator/{index}/stats", recoverPanicWrap(http.HandlerFunc(handlers.ValidatorStatsTable))).Methods("GET")
			router.Handle("/validators", recoverPanicWrap(http.HandlerFunc(handlers.Validators))).Methods("GET")
			router.Handle("/validators/data", recoverPanicWrap(http.HandlerFunc(handlers.ValidatorsData))).Methods("GET")
			router.Handle("/validators/slashings", recoverPanicWrap(http.HandlerFunc(handlers.ValidatorsSlashings))).Methods("GET")
			router.Handle("/validators/slashings/data", recoverPanicWrap(http.HandlerFunc(handlers.ValidatorsSlashingsData))).Methods("GET")
			router.Handle("/validators/leaderboard", recoverPanicWrap(http.HandlerFunc(handlers.ValidatorsLeaderboard))).Methods("GET")
			router.Handle("/validators/leaderboard/data", recoverPanicWrap(http.HandlerFunc(handlers.ValidatorsLeaderboardData))).Methods("GET")
			router.Handle("/validators/streakleaderboard", recoverPanicWrap(http.HandlerFunc(handlers.ValidatorsStreakLeaderboard))).Methods("GET")
			router.Handle("/validators/streakleaderboard/data", recoverPanicWrap(http.HandlerFunc(handlers.ValidatorsStreakLeaderboardData))).Methods("GET")
			router.Handle("/validators/withdrawals", recoverPanicWrap(http.HandlerFunc(handlers.Withdrawals))).Methods("GET")
			router.Handle("/validators/withdrawals/data", recoverPanicWrap(http.HandlerFunc(handlers.WithdrawalsData))).Methods("GET")
			router.Handle("/validators/withdrawals/bls", recoverPanicWrap(http.HandlerFunc(handlers.BLSChangeData))).Methods("GET")
			router.Handle("/validators/deposits", recoverPanicWrap(http.HandlerFunc(handlers.Deposits))).Methods("GET")
			router.Handle("/validators/initiated-deposits", recoverPanicWrap(http.HandlerFunc(handlers.Eth1Deposits))).Methods("GET")
			router.Handle("/validators/initiated-deposits/data", recoverPanicWrap(http.HandlerFunc(handlers.Eth1DepositsData))).Methods("GET")
			router.Handle("/validators/deposit-leaderboard", recoverPanicWrap(http.HandlerFunc(handlers.Eth1DepositsLeaderboard))).Methods("GET")
			router.Handle("/validators/deposit-leaderboard/data", recoverPanicWrap(http.HandlerFunc(handlers.Eth1DepositsLeaderboardData))).Methods("GET")
			router.Handle("/validators/included-deposits", recoverPanicWrap(http.HandlerFunc(handlers.Eth2Deposits))).Methods("GET")
			router.Handle("/validators/included-deposits/data", recoverPanicWrap(http.HandlerFunc(handlers.Eth2DepositsData))).Methods("GET")

			router.Handle("/heatmap", recoverPanicWrap(http.HandlerFunc(handlers.Heatmap))).Methods("GET")

			router.Handle("/dashboard", recoverPanicWrap(http.HandlerFunc(handlers.Dashboard))).Methods("GET")
			router.Handle("/dashboard/save", recoverPanicWrap(http.HandlerFunc(handlers.UserDashboardWatchlistAdd))).Methods("POST")

			router.Handle("/dashboard/data/allbalances", recoverPanicWrap(http.HandlerFunc(handlers.DashboardDataBalanceCombined))).Methods("GET")
			router.Handle("/dashboard/data/balance", recoverPanicWrap(http.HandlerFunc(handlers.DashboardDataBalance))).Methods("GET")
			router.Handle("/dashboard/data/proposals", recoverPanicWrap(http.HandlerFunc(handlers.DashboardDataProposals))).Methods("GET")
			router.Handle("/dashboard/data/proposalshistory", recoverPanicWrap(http.HandlerFunc(handlers.DashboardDataProposalsHistory))).Methods("GET")
			router.Handle("/dashboard/data/validators", recoverPanicWrap(http.HandlerFunc(handlers.DashboardDataValidators))).Methods("GET")
			router.Handle("/dashboard/data/effectiveness", recoverPanicWrap(http.HandlerFunc(handlers.DashboardDataEffectiveness))).Methods("GET")
			router.Handle("/dashboard/data/earnings", recoverPanicWrap(http.HandlerFunc(handlers.DashboardDataEarnings))).Methods("GET")
			router.Handle("/graffitiwall", recoverPanicWrap(http.HandlerFunc(handlers.Graffitiwall))).Methods("GET")
			router.Handle("/calculator", recoverPanicWrap(http.HandlerFunc(handlers.StakingCalculator))).Methods("GET")
			router.Handle("/search", recoverPanicWrap(http.HandlerFunc(handlers.Search))).Methods("POST")
			router.Handle("/search/{type}/{search}", recoverPanicWrap(http.HandlerFunc(handlers.SearchAhead))).Methods("GET")
			router.Handle("/faq", recoverPanicWrap(http.HandlerFunc(handlers.Faq))).Methods("GET")
			router.Handle("/imprint", recoverPanicWrap(http.HandlerFunc(handlers.Imprint))).Methods("GET")
			router.Handle("/poap", recoverPanicWrap(http.HandlerFunc(handlers.Poap))).Methods("GET")
			router.Handle("/poap/data", recoverPanicWrap(http.HandlerFunc(handlers.PoapData))).Methods("GET")
			router.Handle("/mobile", recoverPanicWrap(http.HandlerFunc(handlers.MobilePage))).Methods("GET")
			router.Handle("/mobile", recoverPanicWrap(http.HandlerFunc(handlers.MobilePagePost))).Methods("POST")
			router.Handle("/tools/unitConverter", recoverPanicWrap(http.HandlerFunc(handlers.UnitConverter))).Methods("GET")
			router.Handle("/tools/broadcast", recoverPanicWrap(http.HandlerFunc(handlers.Broadcast))).Methods("GET")
			router.Handle("/tools/broadcast", recoverPanicWrap(http.HandlerFunc(handlers.BroadcastPost))).Methods("POST")
			router.Handle("/tools/broadcast/status/{jobID}", recoverPanicWrap(http.HandlerFunc(handlers.BroadcastStatus))).Methods("GET")

			router.Handle("/tables/state", recoverPanicWrap(http.HandlerFunc(handlers.DataTableStateChanges))).Methods("POST")

			router.Handle("/ethstore", recoverPanicWrap(http.HandlerFunc(handlers.EthStore))).Methods("GET")

			router.Handle("/stakingServices", recoverPanicWrap(http.HandlerFunc(handlers.StakingServices))).Methods("GET")
			router.Handle("/stakingServices", recoverPanicWrap(http.HandlerFunc(handlers.AddStakingServicePost))).Methods("POST")

			router.Handle("/education", recoverPanicWrap(http.HandlerFunc(handlers.EducationServices))).Methods("GET")
			router.Handle("/ethClients", recoverPanicWrap(http.HandlerFunc(handlers.EthClientsServices))).Methods("GET")
			if utils.Config.Frontend.PoolsUpdater.Enabled {
				router.Handle("/pools", recoverPanicWrap(http.HandlerFunc(handlers.Pools))).Methods("GET")
				// router.Handle("/pools/streak/current", recoverPanicWrap(http.HandlerFunc(handlers.GetAvgCurrentStreak))).Methods("GET")
				// router.Handle("/pools/chart/income_per_eth", recoverPanicWrap(http.HandlerFunc(handlers.GetIncomePerEthChart))).Methods("GET")
			}
			router.Handle("/relays", recoverPanicWrap(http.HandlerFunc(handlers.Relays))).Methods("GET")
			router.Handle("/pools/rocketpool", recoverPanicWrap(http.HandlerFunc(handlers.PoolsRocketpool))).Methods("GET")
			router.Handle("/pools/rocketpool/data/minipools", recoverPanicWrap(http.HandlerFunc(handlers.PoolsRocketpoolDataMinipools))).Methods("GET")
			router.Handle("/pools/rocketpool/data/nodes", recoverPanicWrap(http.HandlerFunc(handlers.PoolsRocketpoolDataNodes))).Methods("GET")
			router.Handle("/pools/rocketpool/data/dao_proposals", recoverPanicWrap(http.HandlerFunc(handlers.PoolsRocketpoolDataDAOProposals))).Methods("GET")
			router.Handle("/pools/rocketpool/data/dao_members", recoverPanicWrap(http.HandlerFunc(handlers.PoolsRocketpoolDataDAOMembers))).Methods("GET")

			router.Handle("/advertisewithus", recoverPanicWrap(http.HandlerFunc(handlers.AdvertiseWithUs))).Methods("GET")
			router.Handle("/advertisewithus", recoverPanicWrap(http.HandlerFunc(handlers.AdvertiseWithUsPost))).Methods("POST")

			// confirming the email update should not require auth
			router.Handle("/settings/email/{hash}", recoverPanicWrap(http.HandlerFunc(handlers.UserConfirmUpdateEmail))).Methods("GET")
			router.Handle("/gitcoinfeed", recoverPanicWrap(http.HandlerFunc(handlers.GitcoinFeed))).Methods("GET")
			router.Handle("/rewards", recoverPanicWrap(http.HandlerFunc(handlers.ValidatorRewards))).Methods("GET")
			router.Handle("/rewards/hist", recoverPanicWrap(http.HandlerFunc(handlers.RewardsHistoricalData))).Methods("GET")
			router.Handle("/rewards/hist/download", recoverPanicWrap(http.HandlerFunc(handlers.DownloadRewardsHistoricalData))).Methods("GET")

			router.Handle("/notifications/unsubscribe", recoverPanicWrap(http.HandlerFunc(handlers.UserNotificationsUnsubscribeByHash))).Methods("GET")

			router.Handle("/monitoring/{module}", recoverPanicWrap(http.HandlerFunc(handlers.Monitoring))).Methods("GET", "OPTIONS")

			// router.Handle("/user/validators", recoverPanicWrap(http.HandlerFunc(handlers.UserValidators))).Methods("GET")

			signUpRouter := router.PathPrefix("/").Subrouter()
			signUpRouter.Handle("/login", recoverPanicWrap(http.HandlerFunc(handlers.Login))).Methods("GET")
			signUpRouter.Handle("/login", recoverPanicWrap(http.HandlerFunc(handlers.LoginPost))).Methods("POST")
			signUpRouter.Handle("/logout", recoverPanicWrap(http.HandlerFunc(handlers.Logout))).Methods("GET")
			signUpRouter.Handle("/register", recoverPanicWrap(http.HandlerFunc(handlers.Register))).Methods("GET")
			signUpRouter.Handle("/register", recoverPanicWrap(http.HandlerFunc(handlers.RegisterPost))).Methods("POST")
			signUpRouter.Handle("/resend", recoverPanicWrap(http.HandlerFunc(handlers.ResendConfirmation))).Methods("GET")
			signUpRouter.Handle("/resend", recoverPanicWrap(http.HandlerFunc(handlers.ResendConfirmationPost))).Methods("POST")
			signUpRouter.Handle("/requestReset", recoverPanicWrap(http.HandlerFunc(handlers.RequestResetPassword))).Methods("GET")
			signUpRouter.Handle("/requestReset", recoverPanicWrap(http.HandlerFunc(handlers.RequestResetPasswordPost))).Methods("POST")
			signUpRouter.Handle("/reset", recoverPanicWrap(http.HandlerFunc(handlers.ResetPasswordPost))).Methods("POST")
			signUpRouter.Handle("/reset/{hash}", recoverPanicWrap(http.HandlerFunc(handlers.ResetPassword))).Methods("GET")
			signUpRouter.Handle("/confirm/{hash}", recoverPanicWrap(http.HandlerFunc(handlers.ConfirmEmail))).Methods("GET")
			signUpRouter.Handle("/confirmation", recoverPanicWrap(http.HandlerFunc(handlers.Confirmation))).Methods("GET")
			signUpRouter.Handle("/pricing", recoverPanicWrap(http.HandlerFunc(handlers.Pricing))).Methods("GET")
			signUpRouter.Handle("/pricing", recoverPanicWrap(http.HandlerFunc(handlers.PricingPost))).Methods("POST")
			signUpRouter.Handle("/premium", recoverPanicWrap(http.HandlerFunc(handlers.MobilePricing))).Methods("GET")
			signUpRouter.Use(csrfHandler)

			oauthRouter := router.PathPrefix("/user").Subrouter()
			oauthRouter.Handle("/authorize", recoverPanicWrap(http.HandlerFunc(handlers.UserAuthorizeConfirm))).Methods("GET")
			oauthRouter.Handle("/cancel", recoverPanicWrap(http.HandlerFunc(handlers.UserAuthorizationCancel))).Methods("GET")
			oauthRouter.Use(csrfHandler)

			authRouter := router.PathPrefix("/user").Subrouter()
			authRouter.Handle("/mobile/settings", recoverPanicWrap(http.HandlerFunc(handlers.MobileDeviceSettingsPOST))).Methods("POST")
			authRouter.Handle("/mobile/delete", recoverPanicWrap(http.HandlerFunc(handlers.MobileDeviceDeletePOST))).Methods("POST", "OPTIONS")
			authRouter.Handle("/authorize", recoverPanicWrap(http.HandlerFunc(handlers.UserAuthorizeConfirmPost))).Methods("POST")
			authRouter.Handle("/settings", recoverPanicWrap(http.HandlerFunc(handlers.UserSettings))).Methods("GET")
			authRouter.Handle("/settings/password", recoverPanicWrap(http.HandlerFunc(handlers.UserUpdatePasswordPost))).Methods("POST")
			authRouter.Handle("/settings/flags", recoverPanicWrap(http.HandlerFunc(handlers.UserUpdateFlagsPost))).Methods("POST")
			authRouter.Handle("/settings/delete", recoverPanicWrap(http.HandlerFunc(handlers.UserDeletePost))).Methods("POST")
			authRouter.Handle("/settings/email", recoverPanicWrap(http.HandlerFunc(handlers.UserUpdateEmailPost))).Methods("POST")
			authRouter.Handle("/notifications", recoverPanicWrap(http.HandlerFunc(handlers.UserNotificationsCenter))).Methods("GET")
			authRouter.Handle("/notifications/channels", recoverPanicWrap(http.HandlerFunc(handlers.UsersNotificationChannels))).Methods("POST")
			authRouter.Handle("/notifications/data", recoverPanicWrap(http.HandlerFunc(handlers.UserNotificationsData))).Methods("GET")
			authRouter.Handle("/notifications/subscribe", recoverPanicWrap(http.HandlerFunc(handlers.UserNotificationsSubscribe))).Methods("POST")
			authRouter.Handle("/notifications/network/update", recoverPanicWrap(http.HandlerFunc(handlers.UserModalAddNetworkEvent))).Methods("POST")
			authRouter.Handle("/watchlist/add", recoverPanicWrap(http.HandlerFunc(handlers.UsersModalAddValidator))).Methods("POST")
			authRouter.Handle("/watchlist/remove", recoverPanicWrap(http.HandlerFunc(handlers.UserModalRemoveSelectedValidator))).Methods("POST")
			authRouter.Handle("/watchlist/update", recoverPanicWrap(http.HandlerFunc(handlers.UserModalManageNotificationModal))).Methods("POST")
			authRouter.Handle("/notifications/unsubscribe", recoverPanicWrap(http.HandlerFunc(handlers.UserNotificationsUnsubscribe))).Methods("POST")
			authRouter.Handle("/notifications/bundled/subscribe", recoverPanicWrap(http.HandlerFunc(handlers.MultipleUsersNotificationsSubscribeWeb))).Methods("POST", "OPTIONS")
			authRouter.Handle("/global_notification", recoverPanicWrap(http.HandlerFunc(handlers.UserGlobalNotification))).Methods("GET")
			authRouter.Handle("/global_notification", recoverPanicWrap(http.HandlerFunc(handlers.UserGlobalNotificationPost))).Methods("POST")

			authRouter.Handle("/notifications-center", recoverPanicWrap(http.HandlerFunc(handlers.UserNotificationsCenter))).Methods("GET")
			authRouter.Handle("/notifications-center/removeall", recoverPanicWrap(http.HandlerFunc(handlers.RemoveAllValidatorsAndUnsubscribe))).Methods("POST")
			authRouter.Handle("/notifications-center/validatorsub", recoverPanicWrap(http.HandlerFunc(handlers.AddValidatorsAndSubscribe))).Methods("POST")
			authRouter.Handle("/notifications-center/updatesubs", recoverPanicWrap(http.HandlerFunc(handlers.UserUpdateSubscriptions))).Methods("POST")
			// authRouter.Handle("/notifications-center/monitoring/updatesubs", recoverPanicWrap(http.HandlerFunc(handlers.UserUpdateMonitoringSubscriptions))).Methods("POST")

			authRouter.Handle("/subscriptions/data", recoverPanicWrap(http.HandlerFunc(handlers.UserSubscriptionsData))).Methods("GET")
			authRouter.Handle("/generateKey", recoverPanicWrap(http.HandlerFunc(handlers.GenerateAPIKey))).Methods("POST")
			authRouter.Handle("/ethClients", recoverPanicWrap(http.HandlerFunc(handlers.EthClientsServices))).Methods("GET")
			authRouter.Handle("/rewards", recoverPanicWrap(http.HandlerFunc(handlers.ValidatorRewards))).Methods("GET")
			authRouter.Handle("/rewards/subscribe", recoverPanicWrap(http.HandlerFunc(handlers.RewardNotificationSubscribe))).Methods("POST")
			authRouter.Handle("/rewards/unsubscribe", recoverPanicWrap(http.HandlerFunc(handlers.RewardNotificationUnsubscribe))).Methods("POST")
			authRouter.Handle("/rewards/subscriptions/data", recoverPanicWrap(http.HandlerFunc(handlers.RewardGetUserSubscriptions))).Methods("POST")
			authRouter.Handle("/webhooks", recoverPanicWrap(http.HandlerFunc(handlers.NotificationWebhookPage))).Methods("GET")
			authRouter.Handle("/webhooks/add", recoverPanicWrap(http.HandlerFunc(handlers.UsersAddWebhook))).Methods("POST")
			authRouter.Handle("/webhooks/{webhookID}/update", recoverPanicWrap(http.HandlerFunc(handlers.UsersEditWebhook))).Methods("POST")
			authRouter.Handle("/webhooks/{webhookID}/delete", recoverPanicWrap(http.HandlerFunc(handlers.UsersDeleteWebhook))).Methods("POST")

			err = initStripe(authRouter)
			if err != nil {
				logrus.Errorf("error could not init stripe, %v", err)
			}

			authRouter.Use(handlers.UserAuthMiddleware)
			authRouter.Use(csrfHandler)

			if utils.Config.Frontend.Debug {
				// serve files from local directory when debugging, instead of from go embed file
				templatesHandler := http.FileServer(http.Dir("templates"))
				router.PathPrefix("/templates").Handler(recoverPanicWrap(http.StripPrefix("/templates/", templatesHandler)))

				cssHandler := http.FileServer(http.Dir("static/css"))
				router.PathPrefix("/css").Handler(recoverPanicWrap(http.StripPrefix("/css/", cssHandler)))

				jsHandler := http.FileServer(http.Dir("static/js"))
				router.PathPrefix("/js").Handler(recoverPanicWrap(http.StripPrefix("/js/", jsHandler)))
			}
			legalFs := http.Dir(utils.Config.Frontend.LegalDir)
			//router.PathPrefix("/legal").Handler(recoverPanicWrap(http.StripPrefix("/legal/", http.FileServer(legalFs))))
			router.PathPrefix("/legal").Handler(recoverPanicWrap(http.StripPrefix("/legal/", handlers.CustomFileServer(http.FileServer(legalFs), legalFs, handlers.NotFound))))
			//router.PathPrefix("/").Handler(recoverPanicWrap(http.FileServer(http.FS(static.Files))))
			fileSys := http.FS(static.Files)
			router.PathPrefix("/").Handler(recoverPanicWrap(handlers.CustomFileServer(http.FileServer(fileSys), fileSys, handlers.NotFound)))

		}

		if utils.Config.Metrics.Enabled {
			router.Use(metrics.HttpMiddleware)
		}

		// l := negroni.NewLogger()
		// l.SetFormat(`{{.Request.Header.Get "X-Forwarded-For"}}, {{.Request.RemoteAddr}} | {{.StartTime}} | {{.Status}} | {{.Duration}} | {{.Hostname}} | {{.Method}} {{.Path}}{{if ne .Request.URL.RawQuery ""}}?{{.Request.URL.RawQuery}}{{end}}`)

		n := negroni.New(negroni.NewRecovery()) //, l

		// Customize the logging middleware to include a proper module entry for the frontend
		//frontendLogger := negronilogrus.NewMiddleware()
		//frontendLogger.Before = func(entry *logrus.Entry, request *http.Request, s string) *logrus.Entry {
		//	entry = negronilogrus.DefaultBefore(entry, request, s)
		//	return entry.WithField("module", "frontend")
		//}
		//frontendLogger.After = func(entry *logrus.Entry, writer negroni.ResponseWriter, duration time.Duration, s string) *logrus.Entry {
		//	entry = negronilogrus.DefaultAfter(entry, writer, duration, s)
		//	return entry.WithField("module", "frontend")
		//}
		//n.Use(frontendLogger)

		n.Use(gzip.Gzip(gzip.DefaultCompression))

		pa := &proxyaddr.ProxyAddr{}
		pa.Init(proxyaddr.CIDRLoopback)
		n.Use(pa)

		n.UseHandler(recoverPanicWrap(router))

		if utils.Config.Frontend.HttpWriteTimeout == 0 {
			utils.Config.Frontend.HttpIdleTimeout = time.Second * 15
		}
		if utils.Config.Frontend.HttpReadTimeout == 0 {
			utils.Config.Frontend.HttpIdleTimeout = time.Second * 15
		}
		if utils.Config.Frontend.HttpIdleTimeout == 0 {
			utils.Config.Frontend.HttpIdleTimeout = time.Second * 60
		}
		srv := &http.Server{
			Addr:         cfg.Frontend.Server.Host + ":" + cfg.Frontend.Server.Port,
			WriteTimeout: utils.Config.Frontend.HttpWriteTimeout,
			ReadTimeout:  utils.Config.Frontend.HttpReadTimeout,
			IdleTimeout:  utils.Config.Frontend.HttpIdleTimeout,
			Handler:      n,
		}

		logrus.Printf("http server listening on %v", srv.Addr)
		go func() {
			if err := srv.ListenAndServe(); err != nil {
				logrus.WithError(err).Fatal("Error serving frontend")
			}
		}()
	}
	if utils.Config.Notifications.Enabled {
		services.InitNotifications(utils.Config.Notifications.PubkeyCachePath)
	}

	if utils.Config.Metrics.Enabled {
		go func(addr string) {
			logrus.Infof("Serving metrics on %v", addr)
			if err := metrics.Serve(addr); err != nil {
				logrus.WithError(err).Fatal("Error serving metrics")
			}
		}(utils.Config.Metrics.Address)
	}

	if utils.Config.Frontend.ShowDonors.Enabled {
		services.InitGitCoinFeed()
	}

	// if utils.Config.Frontend.PoolsUpdater.Enabled {
	// services.InitPools() // making sure the website is available before updating
	// }

	utils.WaitForCtrlC()

	logrus.Println("exiting...")
}

func recoverPanic() {
	err := getRecoverError(recover())
	if err != nil {
		handleRecoverError(err, "panic/fatal", 1)
		debug.PrintStack()
	}
}

func recoverPanicWrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := getRecoverError(recover())
			if err != nil {
				handleRecoverError(err, "Recovered from panic/fatal", 2)
				debug.PrintStack()
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, r)
	})
}

func handleRecoverError(err error, text string, skip int) {
	pc, fullFilePath, line, ok := runtime.Caller(skip) // TODO: switch to utils global error handling (#BIDS-1297)
	if ok {
		logrus.WithFields(logrus.Fields{
			"file":       filepath.Base(fullFilePath),
			"line":       strconv.Itoa(line),
			"function":   runtime.FuncForPC(pc).Name(),
			"error type": fmt.Sprintf("%T", err),
		}).WithError(err).Error(text)
	} else {
		logrus.WithFields(logrus.Fields{
			"location":   "Cannot read callstack",
			"error type": fmt.Sprintf("%T", err),
		}).WithError(err).Error(text)
	}
}

func getRecoverError(r any) error {
	if r != nil {
		switch t := r.(type) {
		case string:
			return errors.New(t)
		case error:
			return t
		default:
			return errors.New("unknown error")
		}
	}
	return nil
}
