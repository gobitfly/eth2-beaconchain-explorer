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
		Network          string `yaml:"network" envconfig:"CHAIN_NETWORK"`
		SlotsPerEpoch    uint64 `yaml:"slotsPerEpoch" envconfig:"CHAIN_SLOTS_PER_EPOCH"`
		SecondsPerSlot   uint64 `yaml:"secondsPerSlot" envconfig:"CHAIN_SECONDS_PER_SLOT"`
		GenesisTimestamp uint64 `yaml:"genesisTimestamp" envconfig:"CHAIN_GENESIS_TIMESTAMP"`
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
	Frontend struct {
		Enabled      bool   `yaml:"enabled" envconfig:"FRONTEND_ENABLED"`
		Imprint      string `yaml:"imprint" envconfig:"FRONTEND_IMPRINT"`
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
		SessionSecret string `yaml:"sessionSecret" envconfig:"FRONTEND_SESSION_SECRET"`
		Mail          struct {
			SMTP struct {
				Server   string `yaml:"server" envconfig:"FRONTEND_MAIL_SMTP_SERVER"`
				Host     string `yaml:"host" envconfig:"FRONTEND_MAIL_SMTP_HOST"`
				User     string `yaml:"user" envconfig:"FRONTEND_MAIL_SMTP_USER"`
				Password string `yaml:"password" envconfig:"FRONTEND_MAIL_SMTP_PASSWORD"`
			} `yaml:"smtp"`
			Gunmail struct {
				Domain     string `yaml:"domain" envconfig:"FRONTEND_MAIL_GUNMAIL_DOMAIN"`
				PrivateKey string `yaml:"privateKey" envconfig:"FRONTEND_MAIL_GUNMAIL_PRIVATE_KEY"`
				Sender     string `yaml:"sender" envconfig:"FRONTEND_MAIL_GUNMAIL_SENDER"`
			} `yaml:"gunmail"`
		} `yaml:"mail"`
		FlashSecret string `yaml:"flashSecret" envconfig:"FRONTEND_FLASH_SECRET"`
	} `yaml:"frontend"`
}
