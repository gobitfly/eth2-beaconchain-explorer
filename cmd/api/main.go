package main

import (
	"eth2-exporter/db"
	"eth2-exporter/handlers"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/stripe/stripe-go"
	httpSwagger "github.com/swaggo/http-swagger"
	"gopkg.in/yaml.v2"

	"github.com/sirupsen/logrus"

	_ "eth2-exporter/docs"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/urfave/negroni"
	"github.com/zesik/proxyaddr"
)

func initStripe(http *mux.Router) error {
	if utils.Config == nil {
		return fmt.Errorf("error no config found")
	}
	stripe.Key = utils.Config.Frontend.Stripe.SecretKey //os.Getenv("STRIPE_SECRET_KEY")
	// http.HandleFunc("/stripe/setup", handlers.StripeSetup)
	http.HandleFunc("/stripe/create-checkout-session", handlers.StripeCreateCheckoutSession).Methods("POST")
	http.HandleFunc("/stripe/checkout-session", handlers.StripeCheckoutSession).Methods("GET")
	http.HandleFunc("/stripe/customer-portal", handlers.StripeCustomerPortal).Methods("POST")
	http.HandleFunc("/stripe/webhook", handlers.StripeWebhook).Methods("POST")
	http.HandleFunc("/stripe/success", handlers.PricingSuccess).Methods("GET")
	http.HandleFunc("/stripe/cancled", handlers.PricingCancled).Methods("GET")
	return nil
}

func main() {
	configPath := flag.String("config", "config.yml", "Path to the config file")
	flag.Parse()

	logrus.Printf("config file path: %v", *configPath)
	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg

	// decode phase0 config
	if len(utils.Config.Chain.Phase0Path) > 0 {
		phase0 := &types.Phase0{}
		f, err := os.Open(utils.Config.Chain.Phase0Path)
		if err != nil {
			fmt.Errorf("error opening Phase0 Config file %v: %v", utils.Config.Chain.Phase0Path, err)
		} else {
			decoder := yaml.NewDecoder(f)
			err = decoder.Decode(phase0)
			if err != nil {
				fmt.Errorf("error decoding Phase0 Config file %v: %v", utils.Config.Chain.Phase0Path, err)
			} else {
				utils.Config.Chain.Phase0 = *phase0
			}
		}
	}

	db.MustInitDB(cfg.Database.Username, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	defer db.DB.Close()

	logrus.Infof("database connection established")
	if utils.Config.Chain.SlotsPerEpoch == 0 || utils.Config.Chain.SecondsPerSlot == 0 {
		logrus.Fatal("invalid chain configuration specified, you must specify the slots per epoch, seconds per slot and genesis timestamp in the config file")
	}

	router := mux.NewRouter()

	apiV1Router := mux.NewRouter().PathPrefix("/api/v1").Subrouter()
	router.PathPrefix("/api/v1/docs/").Handler(httpSwagger.WrapHandler)
	apiV1Router.HandleFunc("/epoch/{epoch}", handlers.ApiEpoch).Methods("GET", "OPTIONS")
	apiV1Router.HandleFunc("/epoch/{epoch}/blocks", handlers.ApiEpochBlocks).Methods("GET", "OPTIONS")
	apiV1Router.HandleFunc("/block/{slotOrHash}", handlers.ApiBlock).Methods("GET", "OPTIONS")
	apiV1Router.HandleFunc("/block/{slot}/attestations", handlers.ApiBlockAttestations).Methods("GET", "OPTIONS")
	apiV1Router.HandleFunc("/block/{slot}/deposits", handlers.ApiBlockDeposits).Methods("GET", "OPTIONS")
	apiV1Router.HandleFunc("/block/{slot}/attesterslashings", handlers.ApiBlockAttesterSlashings).Methods("GET", "OPTIONS")
	apiV1Router.HandleFunc("/block/{slot}/proposerslashings", handlers.ApiBlockProposerSlashings).Methods("GET", "OPTIONS")
	apiV1Router.HandleFunc("/block/{slot}/voluntaryexits", handlers.ApiBlockVoluntaryExits).Methods("GET", "OPTIONS")
	apiV1Router.HandleFunc("/eth1deposit/{txhash}", handlers.ApiEth1Deposit).Methods("GET", "OPTIONS")
	apiV1Router.HandleFunc("/validator/leaderboard", handlers.ApiValidatorLeaderboard).Methods("GET", "OPTIONS")
	apiV1Router.HandleFunc("/validator/{indexOrPubkey}", handlers.ApiValidator).Methods("GET", "OPTIONS")
	apiV1Router.HandleFunc("/validator/{indexOrPubkey}/balancehistory", handlers.ApiValidatorBalanceHistory).Methods("GET", "OPTIONS")
	apiV1Router.HandleFunc("/validator/{indexOrPubkey}/performance", handlers.ApiValidatorPerformance).Methods("GET", "OPTIONS")
	apiV1Router.HandleFunc("/validator/{indexOrPubkey}/attestations", handlers.ApiValidatorAttestations).Methods("GET", "OPTIONS")
	apiV1Router.HandleFunc("/validator/{indexOrPubkey}/proposals", handlers.ApiValidatorProposals).Methods("GET", "OPTIONS")
	apiV1Router.HandleFunc("/validator/{indexOrPubkey}/deposits", handlers.ApiValidatorDeposits).Methods("GET", "OPTIONS")
	apiV1Router.HandleFunc("/validator/eth1/{address}", handlers.ApiValidatorByEth1Address).Methods("GET", "OPTIONS")
	apiV1Router.HandleFunc("/chart/{chart}", handlers.ApiChart).Methods("GET", "OPTIONS")
	apiV1Router.Use(utils.CORSMiddleware)
	router.PathPrefix("/api/v1").Handler(apiV1Router)

	router.HandleFunc("/api/healthz", handlers.ApiHealthz).Methods("GET", "HEAD")

	n := negroni.New(negroni.NewRecovery())

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

	utils.WaitForCtrlC()

	logrus.Println("exiting...")
}
