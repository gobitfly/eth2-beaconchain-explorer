package main

import (
	"encoding/hex"
	"eth2-exporter/db"
	ethclients "eth2-exporter/ethClients"
	"eth2-exporter/exporter"
	"eth2-exporter/handlers"
	"eth2-exporter/metrics"
	"eth2-exporter/price"
	"eth2-exporter/rpc"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"flag"
	"fmt"
	"math/big"
	"net/http"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/sirupsen/logrus"

	_ "eth2-exporter/docs"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/stripe/stripe-go/v72"
	"github.com/urfave/negroni"
	"github.com/zesik/proxyaddr"
)

func initStripe(http *mux.Router) error {
	if utils.Config == nil {
		return fmt.Errorf("error no config found")
	}
	stripe.Key = utils.Config.Frontend.Stripe.SecretKey
	http.HandleFunc("/stripe/create-checkout-session", handlers.StripeCreateCheckoutSession).Methods("POST")
	http.HandleFunc("/stripe/customer-portal", handlers.StripeCustomerPortal).Methods("POST")
	return nil
}

func main() {
	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")
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
	}, cfg.Frontend.SessionSecret)
	defer db.FrontendReaderDB.Close()
	defer db.FrontendWriterDB.Close()

	if utils.Config.Metrics.Enabled {
		go metrics.MonitorDB(db.WriterDb)
		DBStr := fmt.Sprintf("%v-%v-%v-%v-%v", cfg.WriterDatabase.Username, cfg.WriterDatabase.Password, cfg.WriterDatabase.Host, cfg.WriterDatabase.Port, cfg.WriterDatabase.Name)
		frontendDBStr := fmt.Sprintf("%v-%v-%v-%v-%v", cfg.Frontend.WriterDatabase.Username, cfg.Frontend.WriterDatabase.Password, cfg.Frontend.WriterDatabase.Host, cfg.Frontend.WriterDatabase.Port, cfg.Frontend.WriterDatabase.Name)
		if DBStr != frontendDBStr {
			go metrics.MonitorDB(db.FrontendWriterDB)
		}
	}

	logrus.Infof("database connection established")
	if utils.Config.Chain.Config.SlotsPerEpoch == 0 || utils.Config.Chain.Config.SecondsPerSlot == 0 {
		logrus.Fatal("invalid chain configuration specified, you must specify the slots per epoch, seconds per slot and genesis timestamp in the config file")
	}

	if utils.Config.Indexer.Enabled {
		var rpcClient rpc.Client

		chainID := new(big.Int).SetUint64(utils.Config.Chain.Config.DepositChainID)
		if utils.Config.Indexer.Node.Type == "prysm" {
			if utils.Config.Indexer.Node.PageSize == 0 {
				logrus.Printf("setting default rpc page size to 500")
				utils.Config.Indexer.Node.PageSize = 500
			}
			rpcClient, err = rpc.NewPrysmClient(cfg.Indexer.Node.Host+":"+cfg.Indexer.Node.Port, chainID)
			if err != nil {
				logrus.Fatal(err)
			}
		} else if utils.Config.Indexer.Node.Type == "lighthouse" {
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

		router := mux.NewRouter()

		apiV1Router := router.PathPrefix("/api/v1").Subrouter()
		router.PathPrefix("/api/v1/docs/").Handler(httpSwagger.WrapHandler)
		apiV1Router.HandleFunc("/epoch/{epoch}", handlers.ApiEpoch).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/epoch/{epoch}/blocks", handlers.ApiEpochBlocks).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/block/{slotOrHash}", handlers.ApiBlock).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/block/{slot}/attestations", handlers.ApiBlockAttestations).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/block/{slot}/deposits", handlers.ApiBlockDeposits).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/block/{slot}/attesterslashings", handlers.ApiBlockAttesterSlashings).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/block/{slot}/proposerslashings", handlers.ApiBlockProposerSlashings).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/block/{slot}/voluntaryexits", handlers.ApiBlockVoluntaryExits).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/sync_committee/{period}", handlers.ApiSyncCommittee).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/eth1deposit/{txhash}", handlers.ApiEth1Deposit).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/validator/leaderboard", handlers.ApiValidatorLeaderboard).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/validator/{indexOrPubkey}", handlers.ApiValidator).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/validator/{indexOrPubkey}/balancehistory", handlers.ApiValidatorBalanceHistory).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/validator/{indexOrPubkey}/performance", handlers.ApiValidatorPerformance).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/validator/{indexOrPubkey}/attestations", handlers.ApiValidatorAttestations).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/validator/{indexOrPubkey}/proposals", handlers.ApiValidatorProposals).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/validator/{indexOrPubkey}/deposits", handlers.ApiValidatorDeposits).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/validator/{indexOrPubkey}/attestationefficiency", handlers.ApiValidatorAttestationEfficiency).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/validator/{indexOrPubkey}/attestationeffectiveness", handlers.ApiValidatorAttestationEffectiveness).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/validator/stats/{index}", handlers.ApiValidatorDailyStats).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/validator/eth1/{address}", handlers.ApiValidatorByEth1Address).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/validators/queue", handlers.ApiValidatorQueue).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/graffitiwall", handlers.ApiGraffitiwall).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/chart/{chart}", handlers.ApiChart).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/user/token", handlers.APIGetToken).Methods("POST", "OPTIONS")
		apiV1Router.HandleFunc("/dashboard/data/balances", handlers.DashboardDataBalance).Methods("GET", "OPTIONS")   // new app versions
		apiV1Router.HandleFunc("/dashboard/data/balance", handlers.APIDashboardDataBalance).Methods("GET", "OPTIONS") // old app versions
		apiV1Router.HandleFunc("/dashboard/data/proposals", handlers.DashboardDataProposals).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/stripe/webhook", handlers.StripeWebhook).Methods("POST")
		apiV1Router.HandleFunc("/stats/{apiKey}/{machine}", handlers.ClientStatsPostOld).Methods("POST", "OPTIONS")
		apiV1Router.HandleFunc("/stats/{apiKey}", handlers.ClientStatsPostOld).Methods("POST", "OPTIONS")
		apiV1Router.HandleFunc("/client/metrics", handlers.ClientStatsPostNew).Methods("POST", "OPTIONS")
		apiV1Router.HandleFunc("/app/dashboard", handlers.ApiDashboard).Methods("POST", "OPTIONS")
		apiV1Router.HandleFunc("/rocketpool/stats", handlers.ApiRocketpoolStats).Methods("GET", "OPTIONS")
		apiV1Router.HandleFunc("/rocketpool/validator/{indexOrPubkey}", handlers.ApiRocketpoolValidators).Methods("GET", "OPTIONS")

		apiV1Router.HandleFunc("/validator/{indexOrPubkey}/widget", handlers.GetMobileWidgetStats).Methods("GET")
		apiV1Router.Use(utils.CORSMiddleware)

		apiV1AuthRouter := apiV1Router.PathPrefix("/user").Subrouter()
		apiV1AuthRouter.HandleFunc("/mobile/notify/register", handlers.MobileNotificationUpdatePOST).Methods("POST", "OPTIONS")
		apiV1AuthRouter.HandleFunc("/mobile/settings", handlers.MobileDeviceSettings).Methods("GET", "OPTIONS")
		apiV1AuthRouter.HandleFunc("/mobile/settings", handlers.MobileDeviceSettingsPOST).Methods("POST", "OPTIONS")
		apiV1AuthRouter.HandleFunc("/validator/saved", handlers.MobileTagedValidators).Methods("GET", "OPTIONS")
		apiV1AuthRouter.HandleFunc("/subscription/register", handlers.RegisterMobileSubscriptions).Methods("POST", "OPTIONS")

		apiV1AuthRouter.HandleFunc("/validator/{pubkey}/add", handlers.UserValidatorWatchlistAdd).Methods("POST", "OPTIONS")
		apiV1AuthRouter.HandleFunc("/validator/{pubkey}/remove", handlers.UserValidatorWatchlistRemove).Methods("POST", "OPTIONS")
		apiV1AuthRouter.HandleFunc("/dashboard/save", handlers.UserDashboardWatchlistAdd).Methods("POST", "OPTIONS")
		apiV1AuthRouter.HandleFunc("/notifications/bundled/subscribe", handlers.MultipleUsersNotificationsSubscribe).Methods("POST", "OPTIONS")
		apiV1AuthRouter.HandleFunc("/notifications/bundled/unsubscribe", handlers.MultipleUsersNotificationsUnsubscribe).Methods("POST", "OPTIONS")
		apiV1AuthRouter.HandleFunc("/notifications/subscribe", handlers.UserNotificationsSubscribe).Methods("POST", "OPTIONS")
		apiV1AuthRouter.HandleFunc("/notifications/unsubscribe", handlers.UserNotificationsUnsubscribe).Methods("POST", "OPTIONS")
		apiV1AuthRouter.HandleFunc("/notifications", handlers.UserNotificationsSubscribed).Methods("POST", "GET", "OPTIONS")
		apiV1AuthRouter.HandleFunc("/stats", handlers.ClientStats).Methods("GET", "OPTIONS")
		apiV1AuthRouter.HandleFunc("/stats/{offset}/{limit}", handlers.ClientStats).Methods("GET", "OPTIONS")
		apiV1AuthRouter.HandleFunc("/ethpool", handlers.RegisterEthpoolSubscription).Methods("POST", "OPTIONS")
		apiV1AuthRouter.Use(utils.CORSMiddleware)
		apiV1AuthRouter.Use(utils.AuthorizedAPIMiddleware)

		router.HandleFunc("/api/healthz", handlers.ApiHealthz).Methods("GET", "HEAD")
		router.HandleFunc("/api/healthz-loadbalancer", handlers.ApiHealthzLoadbalancer).Methods("GET", "HEAD")

		services.Init() // Init frontend services
		price.Init()
		ethclients.Init()

		logrus.Infof("frontend services initiated")

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

			router.HandleFunc("/", handlers.Index).Methods("GET")
			router.HandleFunc("/latestState", handlers.LatestState).Methods("GET")
			router.HandleFunc("/launchMetrics", handlers.LaunchMetricsData).Methods("GET")
			router.HandleFunc("/index/data", handlers.IndexPageData).Methods("GET")
			router.HandleFunc("/block/{slotOrHash}", handlers.Block).Methods("GET")
			router.HandleFunc("/block/{slotOrHash}/deposits", handlers.BlockDepositData).Methods("GET")
			router.HandleFunc("/block/{slotOrHash}/votes", handlers.BlockVoteData).Methods("GET")
			router.HandleFunc("/blocks", handlers.Blocks).Methods("GET")
			router.HandleFunc("/blocks/data", handlers.BlocksData).Methods("GET")
			router.HandleFunc("/vis", handlers.Vis).Methods("GET")
			router.HandleFunc("/charts", handlers.Charts).Methods("GET")
			router.HandleFunc("/charts/{chart}", handlers.Chart).Methods("GET")
			router.HandleFunc("/vis/blocks", handlers.VisBlocks).Methods("GET")
			router.HandleFunc("/vis/votes", handlers.VisVotes).Methods("GET")
			router.HandleFunc("/epoch/{epoch}", handlers.Epoch).Methods("GET")
			router.HandleFunc("/epochs", handlers.Epochs).Methods("GET")
			router.HandleFunc("/epochs/data", handlers.EpochsData).Methods("GET")

			router.HandleFunc("/validator/{index}", handlers.Validator).Methods("GET")
			router.HandleFunc("/validator/{index}/proposedblocks", handlers.ValidatorProposedBlocks).Methods("GET")
			router.HandleFunc("/validator/{index}/attestations", handlers.ValidatorAttestations).Methods("GET")
			router.HandleFunc("/validator/{index}/sync", handlers.ValidatorSync).Methods("GET")
			router.HandleFunc("/validator/{index}/history", handlers.ValidatorHistory).Methods("GET")
			router.HandleFunc("/validator/{pubkey}/deposits", handlers.ValidatorDeposits).Methods("GET")
			router.HandleFunc("/validator/{index}/slashings", handlers.ValidatorSlashings).Methods("GET")
			router.HandleFunc("/validator/{index}/effectiveness", handlers.ValidatorAttestationInclusionEffectiveness).Methods("GET")
			router.HandleFunc("/validator/{pubkey}/save", handlers.ValidatorSave).Methods("POST")
			router.HandleFunc("/validator/{pubkey}/add", handlers.UserValidatorWatchlistAdd).Methods("POST")
			router.HandleFunc("/validator/{pubkey}/remove", handlers.UserValidatorWatchlistRemove).Methods("POST")
			router.HandleFunc("/validator/{index}/stats", handlers.ValidatorStatsTable).Methods("GET")
			router.HandleFunc("/validators", handlers.Validators).Methods("GET")
			router.HandleFunc("/validators/data", handlers.ValidatorsData).Methods("GET")
			router.HandleFunc("/validators/slashings", handlers.ValidatorsSlashings).Methods("GET")
			router.HandleFunc("/validators/slashings/data", handlers.ValidatorsSlashingsData).Methods("GET")
			router.HandleFunc("/validators/leaderboard", handlers.ValidatorsLeaderboard).Methods("GET")
			router.HandleFunc("/validators/leaderboard/data", handlers.ValidatorsLeaderboardData).Methods("GET")
			router.HandleFunc("/validators/streakleaderboard", handlers.ValidatorsStreakLeaderboard).Methods("GET")
			router.HandleFunc("/validators/streakleaderboard/data", handlers.ValidatorsStreakLeaderboardData).Methods("GET")
			router.HandleFunc("/validators/eth1deposits", handlers.Eth1Deposits).Methods("GET")
			router.HandleFunc("/validators/eth1deposits/data", handlers.Eth1DepositsData).Methods("GET")
			router.HandleFunc("/validators/eth1leaderboard", handlers.Eth1DepositsLeaderboard).Methods("GET")
			router.HandleFunc("/validators/eth1leaderboard/data", handlers.Eth1DepositsLeaderboardData).Methods("GET")
			router.HandleFunc("/validators/eth2deposits", handlers.Eth2Deposits).Methods("GET")
			router.HandleFunc("/validators/eth2deposits/data", handlers.Eth2DepositsData).Methods("GET")

			router.HandleFunc("/dashboard", handlers.Dashboard).Methods("GET")
			router.HandleFunc("/dashboard/save", handlers.UserDashboardWatchlistAdd).Methods("POST")

			router.HandleFunc("/dashboard/data/balance", handlers.DashboardDataBalance).Methods("GET")
			router.HandleFunc("/dashboard/data/proposals", handlers.DashboardDataProposals).Methods("GET")
			router.HandleFunc("/dashboard/data/proposalshistory", handlers.DashboardDataProposalsHistory).Methods("GET")
			router.HandleFunc("/dashboard/data/validators", handlers.DashboardDataValidators).Methods("GET")
			router.HandleFunc("/dashboard/data/effectiveness", handlers.DashboardDataEffectiveness).Methods("GET")
			router.HandleFunc("/dashboard/data/earnings", handlers.DashboardDataEarnings).Methods("GET")
			router.HandleFunc("/graffitiwall", handlers.Graffitiwall).Methods("GET")
			router.HandleFunc("/calculator", handlers.StakingCalculator).Methods("GET")
			router.HandleFunc("/search", handlers.Search).Methods("POST")
			router.HandleFunc("/search/{type}/{search}", handlers.SearchAhead).Methods("GET")
			router.HandleFunc("/faq", handlers.Faq).Methods("GET")
			router.HandleFunc("/imprint", handlers.Imprint).Methods("GET")
			router.HandleFunc("/poap", handlers.Poap).Methods("GET")
			router.HandleFunc("/poap/data", handlers.PoapData).Methods("GET")
			router.HandleFunc("/mobile", handlers.MobilePage).Methods("GET")
			router.HandleFunc("/mobile", handlers.MobilePagePost).Methods("POST")

			router.HandleFunc("/stakingServices", handlers.StakingServices).Methods("GET")
			router.HandleFunc("/stakingServices", handlers.AddStakingServicePost).Methods("POST")

			router.HandleFunc("/education", handlers.EducationServices).Methods("GET")
			router.HandleFunc("/ethClients", handlers.EthClientsServices).Methods("GET")
			if utils.Config.Frontend.PoolsUpdater.Enabled {
				router.HandleFunc("/pools", handlers.Pools).Methods("GET")
				// router.HandleFunc("/pools/streak/current", handlers.GetAvgCurrentStreak).Methods("GET")
				// router.HandleFunc("/pools/chart/income_per_eth", handlers.GetIncomePerEthChart).Methods("GET")
			}
			router.HandleFunc("/pools/rocketpool", handlers.PoolsRocketpool).Methods("GET")
			router.HandleFunc("/pools/rocketpool/data/minipools", handlers.PoolsRocketpoolDataMinipools).Methods("GET")
			router.HandleFunc("/pools/rocketpool/data/nodes", handlers.PoolsRocketpoolDataNodes).Methods("GET")
			router.HandleFunc("/pools/rocketpool/data/dao_proposals", handlers.PoolsRocketpoolDataDAOProposals).Methods("GET")
			router.HandleFunc("/pools/rocketpool/data/dao_members", handlers.PoolsRocketpoolDataDAOMembers).Methods("GET")

			router.HandleFunc("/advertisewithus", handlers.AdvertiseWithUs).Methods("GET")
			router.HandleFunc("/advertisewithus", handlers.AdvertiseWithUsPost).Methods("POST")

			// confirming the email update should not require auth
			router.HandleFunc("/settings/email/{hash}", handlers.UserConfirmUpdateEmail).Methods("GET")
			router.HandleFunc("/gitcoinfeed", handlers.GitcoinFeed).Methods("GET")
			router.HandleFunc("/rewards", handlers.ValidatorRewards).Methods("GET")
			router.HandleFunc("/rewards/hist", handlers.RewardsHistoricalData).Methods("GET")
			router.HandleFunc("/rewards/hist/download", handlers.DownloadRewardsHistoricalData).Methods("GET")

			router.HandleFunc("/notifications/unsubscribe", handlers.UserNotificationsUnsubscribeByHash).Methods("GET")

			// router.HandleFunc("/user/validators", handlers.UserValidators).Methods("GET")

			signUpRouter := router.PathPrefix("/").Subrouter()
			signUpRouter.HandleFunc("/login", handlers.Login).Methods("GET")
			signUpRouter.HandleFunc("/login", handlers.LoginPost).Methods("POST")
			signUpRouter.HandleFunc("/logout", handlers.Logout).Methods("GET")
			signUpRouter.HandleFunc("/register", handlers.Register).Methods("GET")
			signUpRouter.HandleFunc("/register", handlers.RegisterPost).Methods("POST")
			signUpRouter.HandleFunc("/resend", handlers.ResendConfirmation).Methods("GET")
			signUpRouter.HandleFunc("/resend", handlers.ResendConfirmationPost).Methods("POST")
			signUpRouter.HandleFunc("/requestReset", handlers.RequestResetPassword).Methods("GET")
			signUpRouter.HandleFunc("/requestReset", handlers.RequestResetPasswordPost).Methods("POST")
			signUpRouter.HandleFunc("/reset", handlers.ResetPasswordPost).Methods("POST")
			signUpRouter.HandleFunc("/reset/{hash}", handlers.ResetPassword).Methods("GET")
			signUpRouter.HandleFunc("/confirm/{hash}", handlers.ConfirmEmail).Methods("GET")
			signUpRouter.HandleFunc("/confirmation", handlers.Confirmation).Methods("GET")
			signUpRouter.HandleFunc("/pricing", handlers.Pricing).Methods("GET")
			signUpRouter.HandleFunc("/pricing", handlers.PricingPost).Methods("POST")
			signUpRouter.HandleFunc("/premium", handlers.MobilePricing).Methods("GET")
			signUpRouter.Use(csrfHandler)

			oauthRouter := router.PathPrefix("/user").Subrouter()
			oauthRouter.HandleFunc("/authorize", handlers.UserAuthorizeConfirm).Methods("GET")
			oauthRouter.HandleFunc("/cancel", handlers.UserAuthorizationCancel).Methods("GET")
			oauthRouter.Use(csrfHandler)

			authRouter := router.PathPrefix("/user").Subrouter()
			authRouter.HandleFunc("/mobile/settings", handlers.MobileDeviceSettingsPOST).Methods("POST")
			authRouter.HandleFunc("/mobile/delete", handlers.MobileDeviceDeletePOST).Methods("POST", "OPTIONS")
			authRouter.HandleFunc("/authorize", handlers.UserAuthorizeConfirmPost).Methods("POST")
			authRouter.HandleFunc("/settings", handlers.UserSettings).Methods("GET")
			authRouter.HandleFunc("/settings/password", handlers.UserUpdatePasswordPost).Methods("POST")
			authRouter.HandleFunc("/settings/flags", handlers.UserUpdateFlagsPost).Methods("POST")
			authRouter.HandleFunc("/settings/delete", handlers.UserDeletePost).Methods("POST")
			authRouter.HandleFunc("/settings/email", handlers.UserUpdateEmailPost).Methods("POST")
			authRouter.HandleFunc("/notifications", handlers.UserNotificationsCenter).Methods("GET")
			authRouter.HandleFunc("/notifications/channels", handlers.UsersNotificationChannels).Methods("POST")
			authRouter.HandleFunc("/notifications/data", handlers.UserNotificationsData).Methods("GET")
			authRouter.HandleFunc("/notifications/subscribe", handlers.UserNotificationsSubscribe).Methods("POST")
			authRouter.HandleFunc("/notifications/unsubscribe", handlers.UserNotificationsUnsubscribe).Methods("POST")
			authRouter.HandleFunc("/notifications/bundled/subscribe", handlers.MultipleUsersNotificationsSubscribeWeb).Methods("POST", "OPTIONS")
			authRouter.HandleFunc("/notifications-center", handlers.UserNotificationsCenter).Methods("GET")
			authRouter.HandleFunc("/notifications-center/removeall", handlers.RemoveAllValidatorsAndUnsubscribe).Methods("POST")
			authRouter.HandleFunc("/notifications-center/validatorsub", handlers.AddValidatorsAndSubscribe).Methods("POST")
			authRouter.HandleFunc("/notifications-center/updatesubs", handlers.UserUpdateSubscriptions).Methods("POST")
			// authRouter.HandleFunc("/notifications-center/monitoring/updatesubs", handlers.UserUpdateMonitoringSubscriptions).Methods("POST")
			authRouter.HandleFunc("/subscriptions/data", handlers.UserSubscriptionsData).Methods("GET")
			authRouter.HandleFunc("/generateKey", handlers.GenerateAPIKey).Methods("POST")
			authRouter.HandleFunc("/ethClients", handlers.EthClientsServices).Methods("GET")
			authRouter.HandleFunc("/rewards", handlers.ValidatorRewards).Methods("GET")
			authRouter.HandleFunc("/rewards/subscribe", handlers.RewardNotificationSubscribe).Methods("POST")
			authRouter.HandleFunc("/rewards/unsubscribe", handlers.RewardNotificationUnsubscribe).Methods("POST")
			authRouter.HandleFunc("/rewards/subscriptions/data", handlers.RewardGetUserSubscriptions).Methods("POST")
			authRouter.HandleFunc("/webhooks", handlers.NotificationWebhookPage).Methods("GET")
			authRouter.HandleFunc("/webhooks/add", handlers.UsersAddWebhook).Methods("POST")
			authRouter.HandleFunc("/webhooks/{webhookID}/update", handlers.UsersEditWebhook).Methods("POST")
			authRouter.HandleFunc("/webhooks/{webhookID}/delete", handlers.UsersDeleteWebhook).Methods("POST")

			err = initStripe(authRouter)
			if err != nil {
				logrus.Errorf("error could not init stripe, %v", err)
			}

			authRouter.Use(handlers.UserAuthMiddleware)
			authRouter.Use(csrfHandler)

			legalFs := http.FileServer(http.Dir(utils.Config.Frontend.LegalDir))
			router.PathPrefix("/legal").Handler(http.StripPrefix("/legal/", legalFs))
			router.PathPrefix("/").Handler(http.FileServer(http.Dir("static")))

		}

		if utils.Config.Metrics.Enabled {
			router.Use(metrics.HttpMiddleware)
		}

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
				logrus.WithError(err).Fatal("Error serving frontend")
			}
		}()
	}
	if utils.Config.Notifications.Enabled {
		services.InitNotifications()
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

	if utils.Config.Frontend.PoolsUpdater.Enabled {
		// services.InitPools() // making sure the website is available before updating
	}

	utils.WaitForCtrlC()

	logrus.Println("exiting...")
}
