package types

// Config is a struct to hold the configuration data
type Config struct {
	Database struct {
		Username string `yaml:"user" envconfig:"DB_USERNAME"`
		Password string `yaml:"password" envconfig:"DB_PASSWORD"`
		Name     string `yaml:"name" envconfig:"DB_NAME"`
		Host     string `yaml:"host" envconfig:"DB_HOST"`
		Port     string `yaml:"port" envconfig:"DB_PORT"`
	} `yaml:"database"`
	Chain struct {
		Network                        string `yaml:"network" envconfig:"CHAIN_NETWORK"`
		SlotsPerEpoch                  uint64 `yaml:"slotsPerEpoch" envconfig:"CHAIN_SLOTS_PER_EPOCH"`
		SecondsPerSlot                 uint64 `yaml:"secondsPerSlot" envconfig:"CHAIN_SECONDS_PER_SLOT"`
		GenesisTimestamp               uint64 `yaml:"genesisTimestamp" envconfig:"CHAIN_GENESIS_TIMESTAMP"`
		MinGenesisActiveValidatorCount uint64 `yaml:"minGenesisActiveValidatorCount" envconfig:"CHAIN_MIN_GENESIS_ACTIVE_VALIDATOR_COUNT"`
		GenesisDelay                   uint64 `yaml:"genesisDelay" envconfig:"CHAIN_GENESIS_DELAY"`
	} `yaml:"chain"`
	Indexer struct {
		Enabled                     bool `yaml:"enabled" envconfig:"INDEXER_ENABLED"`
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
		Eth1Endpoint                  string `yaml:"eth1Endpoint" envconfig:"INDEXER_ETH1_ENDPOINT"`
		Eth1DepositContractAddress    string `yaml:"eth1DepositContractAddress" envconfig:"INDEXER_ETH1_DEPOSIT_CONTRACT_ADDRESS"`
		Eth1DepositContractFirstBlock uint64 `yaml:"eth1DepositContractFirstBlock" envconfig:"INDEXER_ETH1_DEPOSIT_CONTRACT_FIRST_BLOCK"`
	} `yaml:"indexer"`
	OneTimeExport struct {
		Enabled    bool   `yaml:"enabled"`
		StartEpoch uint64 `yaml:"startEpoch"`
		EndEpoch   uint64 `yaml:"endEpoch"`
	} `yaml:"onetimeexport"`
	Frontend struct {
		Enabled      bool   `yaml:"enabled" envconfig:"FRONTEND_ENABLED"`
		Imprint      string `yaml:"imprint" envconfig:"FRONTEND_IMPRINT"`
		SiteDomain   string `yaml:"siteDomain" envconfig:"FRONTEND_SITE_DOMAIN"`
		SiteName     string `yaml:"siteName" envconfig:"FRONTEND_SITE_NAME"`
		SiteSubtitle string `yaml:"siteSubtitle" envconfig:"FRONTEND_SITE_SUBTITLE"`
		Server       struct {
			Port string `yaml:"port" envconfig:"FRONTEND_SERVER_PORT"`
			Host string `yaml:"host" envconfig:"FRONTEND_SERVER_HOST"`
		} `yaml:"server"`
		Database struct {
			Username string `yaml:"user" envconfig:"FRONTEND_DB_USERNAME"`
			Password string `yaml:"password" envconfig:"FRONTEND_DB_PASSWORD"`
			Name     string `yaml:"name" envconfig:"FRONTEND_DB_NAME"`
			Host     string `yaml:"host" envconfig:"FRONTEND_DB_HOST"`
			Port     string `yaml:"port" envconfig:"FRONTEND_DB_PORT"`
		} `yaml:"database"`
		Notifications struct {
			Enabled bool `yaml:"enabled" envconfig:"FRONTEND_NOTIFICATIONS_ENABLED"`
		} `yaml:"notifications"`
		SessionSecret          string `yaml:"sessionSecret" envconfig:"FRONTEND_SESSION_SECRET"`
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
		GATag string `yaml:"gatag"  envconfig:"GATAG"`
	} `yaml:"frontend"`
}
