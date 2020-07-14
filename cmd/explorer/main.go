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
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/urfave/negroni"
	"github.com/zesik/proxyaddr"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	configPath := flag.String("config", "", "Path to the config file")
	flag.Parse()

	log.Printf("config file path: %v", *configPath)
	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)

	if err != nil {
		log.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg

	db.MustInitDB(cfg.Database.Username, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	defer db.DB.Close()

	if utils.Config.Chain.SlotsPerEpoch == 0 || utils.Config.Chain.SecondsPerSlot == 0 || utils.Config.Chain.GenesisTimestamp == 0 {
		log.Fatal("invalid chain configuration specified, you must specify the slots per epoch, seconds per slot and genesis timestamp in the config file")
	}

	if cfg.Indexer.Enabled {
		var rpcClient rpc.Client

		if utils.Config.Indexer.Node.Type == "prysm" {
			if utils.Config.Indexer.Node.PageSize == 0 {
				log.Printf("setting default rpc page size to 500")
				utils.Config.Indexer.Node.PageSize = 500
			}
			rpcClient, err = rpc.NewPrysmClient(cfg.Indexer.Node.Host + ":" + cfg.Indexer.Node.Port)
			if err != nil {
				log.Fatal(err)
			}
		} else if utils.Config.Indexer.Node.Type == "lighthouse" {
			rpcClient, err = rpc.NewLighthouseClient(cfg.Indexer.Node.Host + ":" + cfg.Indexer.Node.Port)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatalf("invalid note type %v specified. supported node types are prysm and lighthouse", utils.Config.Indexer.Node.Type)
		}

		go exporter.Start(rpcClient)
	}

	if cfg.Frontend.Enabled {
		db.MustInitFrontendDB(cfg.Frontend.Database.Username, cfg.Frontend.Database.Password, cfg.Frontend.Database.Host, cfg.Frontend.Database.Port, cfg.Frontend.Database.Name, cfg.Frontend.SessionSecret)
		defer db.FrontendDB.Close()

		services.Init() // Init frontend services
		utils.InitFlash(cfg.Frontend.FlashSecret)
		utils.InitSession(cfg.Frontend.SessionSecret)

		router := mux.NewRouter()
		router.HandleFunc("/", handlers.Index).Methods("GET")
		router.HandleFunc("/latestState", handlers.LatestState).Methods("GET")
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

		router.HandleFunc("/login", handlers.Login).Methods("GET")
		router.HandleFunc("/login", handlers.LoginPost).Methods("POST")
		router.HandleFunc("/logout", handlers.Logout).Methods("GET")
		router.HandleFunc("/confirm/{hash}", handlers.ConfirmEmail).Methods("GET")
		router.HandleFunc("/register", handlers.Register).Methods("GET")
		router.HandleFunc("/register", handlers.RegisterPost).Methods("POST")
		router.HandleFunc("/reset", handlers.ResetPasswordPost).Methods("POST")
		router.HandleFunc("/reset", handlers.ResetPassword).Methods("GET")
		router.HandleFunc("/resend", handlers.ResendConfirmation).Methods("GET")
		router.HandleFunc("/resend", handlers.ResendConfirmationPost).Methods("POST")
		router.HandleFunc("/requestReset", handlers.RequestResetPassword).Methods("GET")
		router.HandleFunc("/requestReset", handlers.RequestResetPassword).Methods("POST")
		router.HandleFunc("/user", handlers.User).Methods("GET")

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

		log.Printf("http server listening on %v", srv.Addr)
		go func() {
			if err := srv.ListenAndServe(); err != nil {
				log.Println(err)
			}
		}()
	}

	utils.WaitForCtrlC()

	log.Println("exiting...")
}
