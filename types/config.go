package types

import (
	"html/template"
)

// Config is a struct to hold the configuration data
type Config struct {
	ReaderDatabase struct {
		Username string `yaml:"user" envconfig:"READER_DB_USERNAME"`
		Password string `yaml:"password" envconfig:"READER_DB_PASSWORD"`
		Name     string `yaml:"name" envconfig:"READER_DB_NAME"`
		Host     string `yaml:"host" envconfig:"READER_DB_HOST"`
		Port     string `yaml:"port" envconfig:"READER_DB_PORT"`
	} `yaml:"readerDatabase"`
	WriterDatabase struct {
		Username string `yaml:"user" envconfig:"WRITER_DB_USERNAME"`
		Password string `yaml:"password" envconfig:"WRITER_DB_PASSWORD"`
		Name     string `yaml:"name" envconfig:"WRITER_DB_NAME"`
		Host     string `yaml:"host" envconfig:"WRITER_DB_HOST"`
		Port     string `yaml:"port" envconfig:"WRITER_DB_PORT"`
	} `yaml:"writerDatabase"`
	Chain struct {
		Name             string `yaml:"name" envconfig:"CHAIN_NAME"`
		GenesisTimestamp uint64 `yaml:"genesisTimestamp" envconfig:"CHAIN_GENESIS_TIMESTAMP"`
		ConfigPath       string `yaml:"configPath" envconfig:"CHAIN_CONFIG_PATH"`
		Config           ChainConfig
	} `yaml:"chain"`
	Indexer struct {
		Enabled                     bool `yaml:"enabled" envconfig:"INDEXER_ENABLED"`
		FixCanonOnStartup           bool `yaml:"fixCanonOnStartup" envconfig:"INDEXER_FIX_CANON_ON_STARTUP"`
		FullIndexOnStartup          bool `yaml:"fullIndexOnStartup" envconfig:"INDEXER_FULL_INDEX_ON_STARTUP"`
		IndexMissingEpochsOnStartup bool `yaml:"indexMissingEpochsOnStartup" envconfig:"INDEXER_MISSING_INDEX_ON_STARTUP"`
		CheckAllBlocksOnStartup     bool `yaml:"checkAllBlocksOnStartup" envconfig:"INDEXER_CHECK_ALL_BLOCKS_ON_STARTUP"`
		UpdateAllEpochStatistics    bool `yaml:"updateAllEpochStatistics" envconfig:"INDEXER_UPDATE_ALL_EPOCH_STATISTICS"`
		Node                        struct {
			Port     string `yaml:"port" envconfig:"INDEXER_NODE_PORT"`
			Host     string `yaml:"host" envconfig:"INDEXER_NODE_HOST"`
			Type     string `yaml:"type" envconfig:"INDEXER_NODE_TYPE"`
			PageSize int32  `yaml:"pageSize" envconfig:"INDEXER_NODE_PAGE_SIZE"`
		} `yaml:"node"`
		Eth1Endpoint string `yaml:"eth1Endpoint" envconfig:"INDEXER_ETH1_ENDPOINT"`
		// Deprecated Please use Phase0 config DEPOSIT_CONTRACT_ADDRESS
		Eth1DepositContractAddress    string `yaml:"eth1DepositContractAddress" envconfig:"INDEXER_ETH1_DEPOSIT_CONTRACT_ADDRESS"`
		Eth1DepositContractFirstBlock uint64 `yaml:"eth1DepositContractFirstBlock" envconfig:"INDEXER_ETH1_DEPOSIT_CONTRACT_FIRST_BLOCK"`
		OneTimeExport                 struct {
			Enabled    bool     `yaml:"enabled" envconfig:"INDEXER_ONETIMEEXPORT_ENABLED"`
			StartEpoch uint64   `yaml:"startEpoch" envconfig:"INDEXER_ONETIMEEXPORT_START_EPOCH"`
			EndEpoch   uint64   `yaml:"endEpoch" envconfig:"INDEXER_ONETIMEEXPORT_END_EPOCH"`
			Epochs     []uint64 `yaml:"epochs" envconfig:"INDEXER_ONETIMEEXPORT_EPOCHS"`
		} `yaml:"onetimeexport"`
		PubKeyTagsExporter struct {
			Enabled bool `yaml:"enabled" envconfig:"PUBKEY_TAGS_EXPORTER_ENABLED"`
		} `yaml:"pubkeyTagsExporter"`
	} `yaml:"indexer"`
	Frontend struct {
		BeaconchainETHPoolBridgeSecret string `yaml:"beaconchainETHPoolBridgeSecret" envconfig:"FRONTEND_BEACONCHAIN_ETHPOOL_BRIDGE_SECRET"`
		Kong                           string `yaml:"kong" envconfig:"FRONTEND_KONG"`
		OnlyAPI                        bool   `yaml:"onlyAPI" envconfig:"FRONTEND_ONLY_API"`
		CsrfAuthKey                    string `yaml:"csrfAuthKey" envconfig:"FRONTEND_CSRF_AUTHKEY"`
		CsrfInsecure                   bool   `yaml:"csrfInsecure" envconfig:"FRONTEND_CSRF_INSECURE"`
		DisableCharts                  bool   `yaml:"disableCharts" envconfig:"disableCharts"`
		RecaptchaSiteKey               string `yaml:"recaptchaSiteKey" envconfig:"FRONTEND_RECAPTCHA_SITEKEY"`
		RecaptchaSecretKey             string `yaml:"recaptchaSecretKey" envconfig:"FRONTEND_RECAPTCHA_SECRETKEY"`
		Enabled                        bool   `yaml:"enabled" envconfig:"FRONTEND_ENABLED"`
		// Imprint is deprecated place imprint file into the legal directory
		Imprint      string `yaml:"imprint" envconfig:"FRONTEND_IMPRINT"`
		LegalDir     string `yaml:"legalDir" envconfig:"FRONTEND_LEGAL"`
		SiteDomain   string `yaml:"siteDomain" envconfig:"FRONTEND_SITE_DOMAIN"`
		SiteName     string `yaml:"siteName" envconfig:"FRONTEND_SITE_NAME"`
		SiteSubtitle string `yaml:"siteSubtitle" envconfig:"FRONTEND_SITE_SUBTITLE"`
		Server       struct {
			Port string `yaml:"port" envconfig:"FRONTEND_SERVER_PORT"`
			Host string `yaml:"host" envconfig:"FRONTEND_SERVER_HOST"`
		} `yaml:"server"`
		ReaderDatabase struct {
			Username string `yaml:"user" envconfig:"FRONTEND_READER_DB_USERNAME"`
			Password string `yaml:"password" envconfig:"FRONTEND_READER_DB_PASSWORD"`
			Name     string `yaml:"name" envconfig:"FRONTEND_READER_DB_NAME"`
			Host     string `yaml:"host" envconfig:"FRONTEND_READER_DB_HOST"`
			Port     string `yaml:"port" envconfig:"FRONTEND_READER_DB_PORT"`
		} `yaml:"readerDatabase"`
		WriterDatabase struct {
			Username string `yaml:"user" envconfig:"FRONTEND_WRITER_DB_USERNAME"`
			Password string `yaml:"password" envconfig:"FRONTEND_WRITER_DB_PASSWORD"`
			Name     string `yaml:"name" envconfig:"FRONTEND_WRITER_DB_NAME"`
			Host     string `yaml:"host" envconfig:"FRONTEND_WRITER_DB_HOST"`
			Port     string `yaml:"port" envconfig:"FRONTEND_WRITER_DB_PORT"`
		} `yaml:"writerDatabase"`
		Stripe struct {
			SecretKey string `yaml:"secretKey" envconfig:"FRONTEND_STRIPE_SECRET_KEY"`
			PublicKey string `yaml:"publicKey" envconfig:"FRONTEND_STRIPE_PUBLIC_KEY"`
			Sapphire  string `yaml:"sapphire" envconfig:"FRONTEND_STRIPE_SAPPHIRE"`
			Emerald   string `yaml:"emerald" envconfig:"FRONTEND_STRIPE_EMERALD"`
			Diamond   string `yaml:"diamond" envconfig:"FRONTEND_STRIPE_DIAMOND"`
			Whale     string `yaml:"whale" envconfig:"FRONTEND_STRIPE_WHALE"`
			Goldfish  string `yaml:"goldfish" envconfig:"FRONTEND_STRIPE_GOLDFISH"`
			Plankton  string `yaml:"plankton" envconfig:"FRONTEND_STRIPE_PLANKTON"`
			Webhook   string `yaml:"webhook" envconfig:"FRONTEND_STRIPE_WEBHOOK"`
		}
		SessionSecret          string `yaml:"sessionSecret" envconfig:"FRONTEND_SESSION_SECRET"`
		JwtSigningSecret       string `yaml:"jwtSigningSecret" envconfig:"FRONTEND_JWT_SECRET"`
		JwtIssuer              string `yaml:"jwtIssuer" envconfig:"FRONTEND_JWT_ISSUER"`
		JwtValidityInMinutes   int    `yaml:"jwtValidityInMinutes" envconfig:"FRONTEND_JWT_VALIDITY_INMINUTES"`
		MaxMailsPerEmailPerDay int    `yaml:"maxMailsPerEmailPerDay" envconfig:"FRONTEND_MAX_MAIL_PER_EMAIL_PER_DAY"`
		Mail                   struct {
			SMTP struct {
				Server   string `yaml:"server" envconfig:"FRONTEND_MAIL_SMTP_SERVER"`
				Host     string `yaml:"host" envconfig:"FRONTEND_MAIL_SMTP_HOST"`
				User     string `yaml:"user" envconfig:"FRONTEND_MAIL_SMTP_USER"`
				Password string `yaml:"password" envconfig:"FRONTEND_MAIL_SMTP_PASSWORD"`
			} `yaml:"smtp"`
			Mailgun struct {
				Domain     string `yaml:"domain" envconfig:"FRONTEND_MAIL_MAILGUN_DOMAIN"`
				PrivateKey string `yaml:"privateKey" envconfig:"FRONTEND_MAIL_MAILGUN_PRIVATE_KEY"`
				Sender     string `yaml:"sender" envconfig:"FRONTEND_MAIL_MAILGUN_SENDER"`
			} `yaml:"mailgun"`
		} `yaml:"mail"`
		GATag                  string `yaml:"gatag" envconfig:"GATAG"`
		VerifyAppSubs          bool   `yaml:"verifyAppSubscriptions" envconfig:"FRONTEND_VERIFY_APP_SUBSCRIPTIONS"`
		AppSubsAppleSecret     string `yaml:"appSubsAppleSecret" envconfig:"FRONTEND_APP_SUBS_APPLE_SECRET"`
		AppSubsGoogleJSONPath  string `yaml:"appSubsGoogleJsonPath" envconfig:"FRONTEND_APP_SUBS_GOOGLE_JSON_PATH"`
		CleanupOldMachineStats bool   `yaml:"cleanupOldMachineStats" envconfig:"FRONTEND_CLEANUP_OLD_MACHINE_STATS"`
		DisableStatsInserts    bool   `yaml:"disableStatsInserts" envconfig:"FRONTEND_DISABLE_STATS_INSERTS"`
		ShowDonors             struct {
			Enabled bool   `yaml:"enabled" envconfig:"FRONTEND_SHOW_DONORS_ENABLED"`
			URL     string `yaml:"gitcoinURL" envconfig:"FRONTEND_GITCOIN_URL"`
		} `yaml:"showDonors"`
		Countdown struct {
			Enabled   bool          `yaml:"enabled" envconfig:"FRONTEND_COUNTDOWN_ENABLED"`
			Title     template.HTML `yaml:"title" envconfig:"FRONTEND_COUNTDOWN_TITLE"`
			Timestamp uint64        `yaml:"timestamp" envconfig:"FRONTEND_COUNTDOWN_TIMESTAMP"`
			Info      string        `yaml:"info" envconfig:"FRONTEND_COUNTDOWN_INFO"`
		} `yaml:"countdown"`
		PoolsUpdater struct {
			Enabled bool `yaml:"enabled" envconfig:"FRONTEND_POOLS_UPDATER"`
		} `yaml:"poolsUpdater"`
	} `yaml:"frontend"`
	Metrics struct {
		Enabled bool   `yaml:"enabled" envconfig:"METRICS_ENABLED"`
		Address string `yaml:"address" envconfig:"METRICS_ADDRESS"`
	} `yaml:"metrics"`
	Notifications struct {
		Enabled                                       bool   `yaml:"enabled" envconfig:"FRONTEND_NOTIFICATIONS_ENABLED"`
		Sender                                        bool   `yaml:"sender" envconfig:"FRONTEND_NOTIFICATIONS_ENABLED"`
		UserDBNotifications                           bool   `yaml:"userDbNotifications" envconfig:"FRONTEND_USERDB_NOTIFICATIONS_ENABLED"`
		FirebaseCredentialsPath                       string `yaml:"firebaseCredentialsPath" envconfig:"FRONTEND_NOTIFICATIONS_FIREBASE_CRED_PATH"`
		ValidatorBalanceDecreasedNotificationsEnabled bool   `yaml:"validatorBalanceDecreasedNotificationsEnabled" envconfig:"FRONTEND_VALIDATOR_BALANCE_DECREASED_NOTIFICATIONS_ENABLED"`
	} `yaml:"notifications"`
	SSVExporter struct {
		Enabled bool   `yaml:"enabled" envconfig:"SSV_EXPORTER_ENABLED"`
		Address string `yaml:"address" envconfig:"SSV_EXPORTER_ADDRESS"`
	} `yaml:"SSVExporter"`
	RocketpoolExporter struct {
		Enabled                   bool   `yaml:"enabled" envconfig:"ROCKETPOOL_EXPORTER_ENABLED"`
		StorageContractAddress    string `yaml:"storageContractAddress" envconfig:"ROCKETPOOL_EXPORTER_STORAGE_CONTRACT_ADDRESS"`
		StorageContractFirstBlock uint64 `yaml:"storageContractFirstBlock" envconfig:"ROCKETPOOL_EXPORTER_STORAGE_CONTRACT_FIRST_BLOCK"`
	} `yaml:"rocketpoolExporter"`
}

type DatabaseConfig struct {
	Username string
	Password string
	Name     string
	Host     string
	Port     string
}
