package main

import (
	"eth2-exporter/db"
	"eth2-exporter/exporter"
	"eth2-exporter/handlers"
	"eth2-exporter/rpc"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"flag"
	"net/http"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/sirupsen/logrus"

	_ "eth2-exporter/docs"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/urfave/negroni"
	"github.com/zesik/proxyaddr"
)

func main() {
	configPath := flag.String("config", "", "Path to the config file")
	flag.Parse()

	logrus.Printf("config file path: %v", *configPath)
	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)

	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg

	db.MustInitDB(cfg.Database.Username, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	defer db.DB.Close()

	logrus.Infof("database connection established")
	if utils.Config.Chain.SlotsPerEpoch == 0 || utils.Config.Chain.SecondsPerSlot == 0 || utils.Config.Chain.GenesisTimestamp == 0 {
		logrus.Fatal("invalid chain configuration specified, you must specify the slots per epoch, seconds per slot and genesis timestamp in the config file")
	}

	if cfg.OneTimeExport.Enabled {
		var rpcClient rpc.Client

		if utils.Config.Indexer.Node.Type == "prysm" {
			if utils.Config.Indexer.Node.PageSize == 0 {
				logrus.Printf("setting default rpc page size to 500")
				utils.Config.Indexer.Node.PageSize = 500
			}
			rpcClient, err = rpc.NewPrysmClient(cfg.Indexer.Node.Host + ":" + cfg.Indexer.Node.Port)
			if err != nil {
				logrus.Fatal(err)
			}
		} else if utils.Config.Indexer.Node.Type == "lighthouse" {
			rpcClient, err = rpc.NewLighthouseClient(cfg.Indexer.Node.Host + ":" + cfg.Indexer.Node.Port)
			if err != nil {
				logrus.Fatal(err)
			}
		} else {
			logrus.Fatalf("invalid note type %v specified. supported node types are prysm and lighthouse", utils.Config.Indexer.Node.Type)
		}
		err := exporter.ExportEpoch(cfg.OneTimeExport.Epoch, rpcClient)

		if err != nil {
			logrus.Fatal(err)
		}
		return
	}

	if cfg.Indexer.Enabled {
		var rpcClient rpc.Client

		if utils.Config.Indexer.Node.Type == "prysm" {
			if utils.Config.Indexer.Node.PageSize == 0 {
				logrus.Printf("setting default rpc page size to 500")
				utils.Config.Indexer.Node.PageSize = 500
			}
			rpcClient, err = rpc.NewPrysmClient(cfg.Indexer.Node.Host + ":" + cfg.Indexer.Node.Port)
			if err != nil {
				logrus.Fatal(err)
			}
		} else if utils.Config.Indexer.Node.Type == "lighthouse" {
			rpcClient, err = rpc.NewLighthouseClient(cfg.Indexer.Node.Host + ":" + cfg.Indexer.Node.Port)
			if err != nil {
				logrus.Fatal(err)
			}
		} else {
			logrus.Fatalf("invalid note type %v specified. supported node types are prysm and lighthouse", utils.Config.Indexer.Node.Type)
		}
		go exporter.Start(rpcClient)
	}

	if cfg.Frontend.Enabled {
		if utils.Config.Frontend.SiteDomain == "" {
			utils.Config.Frontend.SiteDomain = "beaconcha.in"
		}
		db.MustInitFrontendDB(cfg.Frontend.Database.Username, cfg.Frontend.Database.Password, cfg.Frontend.Database.Host, cfg.Frontend.Database.Port, cfg.Frontend.Database.Name, cfg.Frontend.SessionSecret)
		defer db.FrontendDB.Close()

		logrus.Infof("frontend database connection established")
		services.Init() // Init frontend services

		logrus.Infof("frontend services initiated")
		utils.InitSessionStore(cfg.Frontend.SessionSecret)

		router := mux.NewRouter()
		router.HandleFunc("/", handlers.Index).Methods("GET")
		router.HandleFunc("/latestState", handlers.LatestState).Methods("GET")
		router.HandleFunc("/launchMetrics", handlers.LaunchMetricsData).Methods("GET")
		router.HandleFunc("/index/data", handlers.IndexPageData).Methods("GET")
		router.HandleFunc("/block/{slotOrHash}", handlers.Block).Methods("GET")
		router.HandleFunc("/blocks", handlers.Blocks).Methods("GET")
		router.HandleFunc("/blocks/data", handlers.BlocksData).Methods("GET")
		router.HandleFunc("/vis", handlers.Vis).Methods("GET")
		router.HandleFunc("/charts", handlers.Charts).Methods("GET")
		router.HandleFunc("/charts/{chart}", handlers.GenericChart).Methods("GET")
		router.HandleFunc("/vis/blocks", handlers.VisBlocks).Methods("GET")
		router.HandleFunc("/vis/votes", handlers.VisVotes).Methods("GET")
		router.HandleFunc("/epoch/{epoch}", handlers.Epoch).Methods("GET")
		router.HandleFunc("/epochs", handlers.Epochs).Methods("GET")
		router.HandleFunc("/epochs/data", handlers.EpochsData).Methods("GET")

		router.HandleFunc("/validator/{index}", handlers.Validator).Methods("GET")
		router.HandleFunc("/validator/{pubkey}/add", handlers.UserValidatorWatchlistAdd).Methods("POST")
		router.HandleFunc("/validator/{pubkey}/remove", handlers.UserValidatorWatchlistRemove).Methods("POST")
		router.HandleFunc("/validator/{index}/proposedblocks", handlers.ValidatorProposedBlocks).Methods("GET")
		router.HandleFunc("/validator/{index}/attestations", handlers.ValidatorAttestations).Methods("GET")
		router.HandleFunc("/validator/{pubkey}/deposits", handlers.ValidatorDeposits).Methods("GET")
		router.HandleFunc("/validator/{index}/slashings", handlers.ValidatorSlashings).Methods("GET")
		router.HandleFunc("/validator/{pubkey}/save", handlers.ValidatorSave).Methods("POST")
		router.HandleFunc("/validators", handlers.Validators).Methods("GET")
		router.HandleFunc("/validators/data", handlers.ValidatorsData).Methods("GET")
		router.HandleFunc("/validators/slashings", handlers.ValidatorsSlashings).Methods("GET")
		router.HandleFunc("/validators/slashings/data", handlers.ValidatorsSlashingsData).Methods("GET")
		router.HandleFunc("/validators/leaderboard", handlers.ValidatorsLeaderboard).Methods("GET")
		router.HandleFunc("/validators/leaderboard/data", handlers.ValidatorsLeaderboardData).Methods("GET")
		router.HandleFunc("/validators/eth1deposits", handlers.Eth1Deposits).Methods("GET")
		router.HandleFunc("/validators/eth1deposits/data", handlers.Eth1DepositsData).Methods("GET")
		router.HandleFunc("/validators/eth1leaderboard", handlers.Eth1DepositsLeaderboard).Methods("GET")
		router.HandleFunc("/validators/eth1leaderboard/data", handlers.Eth1DepositsLeaderboardData).Methods("GET")
		router.HandleFunc("/validators/eth2deposits", handlers.Eth2Deposits).Methods("GET")
		router.HandleFunc("/validators/eth2deposits/data", handlers.Eth2DepositsData).Methods("GET")

		router.HandleFunc("/dashboard", handlers.Dashboard).Methods("GET")
		router.HandleFunc("/dashboard/data/balance", handlers.DashboardDataBalance).Methods("GET")
		router.HandleFunc("/dashboard/data/proposals", handlers.DashboardDataProposals).Methods("GET")
		router.HandleFunc("/dashboard/data/validators", handlers.DashboardDataValidators).Methods("GET")
		router.HandleFunc("/dashboard/data/earnings", handlers.DashboardDataEarnings).Methods("GET")
		router.HandleFunc("/graffitiwall", handlers.Graffitiwall).Methods("GET")
		router.HandleFunc("/calculator", handlers.StakingCalculator).Methods("GET")
		router.HandleFunc("/search", handlers.Search).Methods("POST")
		router.HandleFunc("/search/{type}/{search}", handlers.SearchAhead).Methods("GET")
		router.HandleFunc("/faq", handlers.Faq).Methods("GET")
		router.HandleFunc("/imprint", handlers.Imprint).Methods("GET")
		router.HandleFunc("/poap", handlers.Poap).Methods("GET")
		router.HandleFunc("/poap/data", handlers.PoapData).Methods("GET")

		router.HandleFunc("/login", handlers.Login).Methods("GET")
		router.HandleFunc("/login", handlers.LoginPost).Methods("POST")
		router.HandleFunc("/logout", handlers.Logout).Methods("GET")
		router.HandleFunc("/register", handlers.Register).Methods("GET")
		router.HandleFunc("/register", handlers.RegisterPost).Methods("POST")
		router.HandleFunc("/resend", handlers.ResendConfirmation).Methods("GET")
		router.HandleFunc("/resend", handlers.ResendConfirmationPost).Methods("POST")
		router.HandleFunc("/requestReset", handlers.RequestResetPassword).Methods("GET")
		router.HandleFunc("/requestReset", handlers.RequestResetPasswordPost).Methods("POST")
		router.HandleFunc("/confirm/{hash}", handlers.ConfirmEmail).Methods("GET")
		router.HandleFunc("/reset/{hash}", handlers.ResetPassword).Methods("GET")
		router.HandleFunc("/reset", handlers.ResetPasswordPost).Methods("POST")

		router.HandleFunc("/stakingServices", handlers.StakingServices).Methods("GET")
		router.HandleFunc("/stakingServices", handlers.AddStakingServicePost).Methods("Post")
		router.HandleFunc("/advertisewithus", handlers.AdvertiseWithUs).Methods("GET")
		router.HandleFunc("/advertisewithus", handlers.AdvertiseWithUsPost).Methods("POST")
		router.HandleFunc("/api/healthz", handlers.ApiHealthz).Methods("GET")

		apiV1Router := mux.NewRouter().PathPrefix("/api/v1").Subrouter()
		router.PathPrefix("/api/v1/docs/").Handler(httpSwagger.WrapHandler)
		apiV1Router.HandleFunc("/epoch/{epoch}", handlers.ApiEpoch).Methods("GET")
		apiV1Router.HandleFunc("/epoch/{epoch}/blocks", handlers.ApiEpochBlocks).Methods("GET")
		apiV1Router.HandleFunc("/block/{slotOrHash}", handlers.ApiBlock).Methods("GET")
		apiV1Router.HandleFunc("/block/{slot}/attestations", handlers.ApiBlockAttestations).Methods("GET")
		apiV1Router.HandleFunc("/block/{slot}/deposits", handlers.ApiBlockDeposits).Methods("GET")
		apiV1Router.HandleFunc("/block/{slot}/attesterslashings", handlers.ApiBlockAttesterSlashings).Methods("GET")
		apiV1Router.HandleFunc("/block/{slot}/proposerslashings", handlers.ApiBlockProposerSlashings).Methods("GET")
		apiV1Router.HandleFunc("/block/{slot}/voluntaryexits", handlers.ApiBlockVoluntaryExits).Methods("GET")
		apiV1Router.HandleFunc("/eth1deposit/{txhash}", handlers.ApiEth1Deposit).Methods("GET")
		apiV1Router.HandleFunc("/validator/leaderboard", handlers.ApiValidatorLeaderboard).Methods("GET")
		apiV1Router.HandleFunc("/validator/{index}", handlers.ApiValidator).Methods("GET")
		apiV1Router.HandleFunc("/validator/{index}/balancehistory", handlers.ApiValidatorBalanceHistory).Methods("GET")
		apiV1Router.HandleFunc("/validator/{index}/performance", handlers.ApiValidatorPerformance).Methods("GET")
		apiV1Router.HandleFunc("/validator/{index}/attestations", handlers.ApiValidatorAttestations).Methods("GET")
		apiV1Router.HandleFunc("/validator/{index}/proposals", handlers.ApiValidatorProposals).Methods("GET")
		apiV1Router.HandleFunc("/validator/{index}/deposits", handlers.ApiValidatorDeposits).Methods("GET")
		apiV1Router.HandleFunc("/validator/eth1/{address}", handlers.ApiValidatorByEth1Address).Methods("GET")
		apiV1Router.HandleFunc("/chart/{chart}", handlers.ApiChart).Methods("GET")
		router.PathPrefix("/api/v1").Handler(apiV1Router)

		// confirming the email update should not require auth
		router.HandleFunc("/settings/email/{hash}", handlers.UserConfirmUpdateEmail).Methods("GET")

		authRouter := mux.NewRouter().PathPrefix("/user").Subrouter()
		authRouter.HandleFunc("/settings", handlers.UserSettings).Methods("GET")
		authRouter.HandleFunc("/settings/password", handlers.UserUpdatePasswordPost).Methods("POST")
		authRouter.HandleFunc("/settings/delete", handlers.UserDeletePost).Methods("POST")
		authRouter.HandleFunc("/settings/email", handlers.UserUpdateEmailPost).Methods("POST")
		authRouter.HandleFunc("/notifications", handlers.UserNotifications).Methods("GET")
		authRouter.HandleFunc("/notifications/data", handlers.UserNotificationsData).Methods("GET")
		authRouter.HandleFunc("/notifications/subscribe", handlers.UserNotificationsSubscribe).Methods("POST")
		authRouter.HandleFunc("/notifications/unsubscribe", handlers.UserNotificationsUnsubscribe).Methods("POST")
		authRouter.HandleFunc("/subscriptions/data", handlers.UserSubscriptionsData).Methods("GET")

		authRouter.HandleFunc("/dashboard/save", handlers.UserDashboardWatchlistAdd).Methods("POST")

		router.PathPrefix("/user").Handler(
			negroni.New(
				negroni.HandlerFunc(handlers.UserAuthMiddleware),
				negroni.Wrap(authRouter),
			),
		)

		router.HandleFunc("/confirmation", handlers.Confirmation).Methods("GET")

		// router.HandleFunc("/user/validators", handlers.UserValidators).Methods("GET")

		router.PathPrefix("/").Handler(http.FileServer(http.Dir("static")))

		n := negroni.New(negroni.NewRecovery())

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

		n.UseHandler(router)

		srv := &http.Server{
			Addr:         cfg.Frontend.Server.Host + ":" + cfg.Frontend.Server.Port,
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
			IdleTimeout:  time.Second * 60,
			Handler:      n,
		}

		logrus.Printf("http server listening on %v", srv.Addr)
		go func() {
			if err := srv.ListenAndServe(); err != nil {
				logrus.Println(err)
			}
		}()
	}

	utils.WaitForCtrlC()

	logrus.Println("exiting...")
}
