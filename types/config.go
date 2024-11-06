package types

import (
	"html/template"
	"time"

	"github.com/ethereum/go-ethereum/params"
)

// Config is a struct to hold the configuration data
type Config struct {
	ReaderDatabase struct {
		Username     string `yaml:"user" envconfig:"READER_DB_USERNAME"`
		Password     string `yaml:"password" envconfig:"READER_DB_PASSWORD"`
		Name         string `yaml:"name" envconfig:"READER_DB_NAME"`
		Host         string `yaml:"host" envconfig:"READER_DB_HOST"`
		Port         string `yaml:"port" envconfig:"READER_DB_PORT"`
		MaxOpenConns int    `yaml:"maxOpenConns" envconfig:"READER_DB_MAX_OPEN_CONNS"`
		MaxIdleConns int    `yaml:"maxIdleConns" envconfig:"READER_DB_MAX_IDLE_CONNS"`
		SSL          bool   `yaml:"ssl" envconfig:"READER_DB_SSL"`
	} `yaml:"readerDatabase"`
	WriterDatabase struct {
		Username     string `yaml:"user" envconfig:"WRITER_DB_USERNAME"`
		Password     string `yaml:"password" envconfig:"WRITER_DB_PASSWORD"`
		Name         string `yaml:"name" envconfig:"WRITER_DB_NAME"`
		Host         string `yaml:"host" envconfig:"WRITER_DB_HOST"`
		Port         string `yaml:"port" envconfig:"WRITER_DB_PORT"`
		MaxOpenConns int    `yaml:"maxOpenConns" envconfig:"WRITER_DB_MAX_OPEN_CONNS"`
		MaxIdleConns int    `yaml:"maxIdleConns" envconfig:"WRITER_DB_MAX_IDLE_CONNS"`
		SSL          bool   `yaml:"ssl" envconfig:"WRITER_DB_SSL"`
	} `yaml:"writerDatabase"`
	Bigtable struct {
		Project             string `yaml:"project" envconfig:"BIGTABLE_PROJECT"`
		Instance            string `yaml:"instance" envconfig:"BIGTABLE_INSTANCE"`
		Emulator            bool   `yaml:"emulator" envconfig:"BIGTABLE_EMULATOR"`
		EmulatorPort        int    `yaml:"emulatorPort" envconfig:"BIGTABLE_EMULATOR_PORT"`
		EmulatorHost        string `yaml:"emulatorHost" envconfig:"BIGTABLE_EMULATOR_HOST"`
		V2SchemaCutOffEpoch uint64 `yaml:"v2SchemaCutOffEpoch" envconfig:"BIGTABLE_V2_SCHEMA_CUTT_OFF_EPOCH"`
	} `yaml:"bigtable"`
	BlobIndexer struct {
		S3 struct {
			Endpoint        string `yaml:"endpoint" envconfig:"BLOB_INDEXER_S3_ENDPOINT"`
			Bucket          string `yaml:"bucket" envconfig:"BLOB_INDEXER_S3_BUCKET"`
			AccessKeyId     string `yaml:"accessKeyId" envconfig:"BLOB_INDEXER_S3_ACCESS_KEY_ID"`
			AccessKeySecret string `yaml:"accessKeySecret" envconfig:"BLOB_INDEXER_S3_ACCESS_KEY_SECRET"`
		} `yaml:"s3"`
	} `yaml:"blobIndexer"`
	Chain struct {
		Name                       string `yaml:"name" envconfig:"CHAIN_NAME"`
		Id                         uint64 `yaml:"id" envconfig:"CHAIN_ID"`
		GenesisTimestamp           uint64 `yaml:"genesisTimestamp" envconfig:"CHAIN_GENESIS_TIMESTAMP"`
		GenesisValidatorsRoot      string `yaml:"genesisValidatorsRoot" envconfig:"CHAIN_GENESIS_VALIDATORS_ROOT"`
		DomainBLSToExecutionChange string `yaml:"domainBLSToExecutionChange" envconfig:"CHAIN_DOMAIN_BLS_TO_EXECUTION_CHANGE"`
		DomainVoluntaryExit        string `yaml:"domainVoluntaryExit" envconfig:"CHAIN_DOMAIN_VOLUNTARY_EXIT"`
		ClConfigPath               string `yaml:"clConfigPath" envconfig:"CHAIN_CL_CONFIG_PATH"`
		ElConfigPath               string `yaml:"elConfigPath" envconfig:"CHAIN_EL_CONFIG_PATH"`
		ClConfig                   ClChainConfig
		ElConfig                   *params.ChainConfig
	} `yaml:"chain"`
	Eth1ErigonEndpoint        string `yaml:"eth1ErigonEndpoint" envconfig:"ETH1_ERIGON_ENDPOINT"`
	Eth1GethEndpoint          string `yaml:"eth1GethEndpoint" envconfig:"ETH1_GETH_ENDPOINT"`
	EtherscanAPIKey           string `yaml:"etherscanApiKey" envconfig:"ETHERSCAN_API_KEY"`
	EtherscanAPIBaseURL       string `yaml:"etherscanApiBaseUrl" envconfig:"ETHERSCAN_API_BASEURL"`
	RedisCacheEndpoint        string `yaml:"redisCacheEndpoint" envconfig:"REDIS_CACHE_ENDPOINT"`
	RedisSessionStoreEndpoint string `yaml:"redisSessionStoreEndpoint" envconfig:"REDIS_SESSION_STORE_ENDPOINT"`
	TieredCacheProvider       string `yaml:"tieredCacheProvider" envconfig:"CACHE_PROVIDER"`
	ReportServiceStatus       bool   `yaml:"reportServiceStatus" envconfig:"REPORT_SERVICE_STATUS"`
	ClickHouse                struct {
		ReaderDatabase struct {
			Username     string `yaml:"user" envconfig:"CLICKHOUSE_READER_DB_USERNAME"`
			Password     string `yaml:"password" envconfig:"CLICKHOUSE_READER_DB_PASSWORD"`
			Name         string `yaml:"name" envconfig:"CLICKHOUSE_READER_DB_NAME"`
			Host         string `yaml:"host" envconfig:"CLICKHOUSE_READER_DB_HOST"`
			Port         string `yaml:"port" envconfig:"CLICKHOUSE_READER_DB_PORT"`
			MaxOpenConns int    `yaml:"maxOpenConns" envconfig:"CLICKHOUSE_READER_DB_MAX_OPEN_CONNS"`
			MaxIdleConns int    `yaml:"maxIdleConns" envconfig:"CLICKHOUSE_READER_DB_MAX_IDLE_CONNS"`
		} `yaml:"readerDatabase"`
	} `yaml:"clickhouse"`
	ClickHouseEnabled bool          `yaml:"clickHouseEnabled" envconfig:"CLICKHOUSE_ENABLED"`
	ClickhouseDelay   time.Duration `yaml:"clickhouseDelay" envconfig:"CLICKHOUSE_DELAY"`
	Indexer           struct {
		Enabled bool `yaml:"enabled" envconfig:"INDEXER_ENABLED"`
		Node    struct {
			Port     string `yaml:"port" envconfig:"INDEXER_NODE_PORT"`
			Host     string `yaml:"host" envconfig:"INDEXER_NODE_HOST"`
			Type     string `yaml:"type" envconfig:"INDEXER_NODE_TYPE"`
			PageSize int32  `yaml:"pageSize" envconfig:"INDEXER_NODE_PAGE_SIZE"`
		} `yaml:"node"`
		Eth1DepositContractFirstBlock uint64 `yaml:"eth1DepositContractFirstBlock" envconfig:"INDEXER_ETH1_DEPOSIT_CONTRACT_FIRST_BLOCK"`
		PubKeyTagsExporter            struct {
			Enabled bool `yaml:"enabled" envconfig:"PUBKEY_TAGS_EXPORTER_ENABLED"`
		} `yaml:"pubkeyTagsExporter"`
		EnsTransformer struct {
			ValidRegistrarContracts []string `yaml:"validRegistrarContracts" envconfig:"ENS_VALID_REGISTRAR_CONTRACTS"`
		} `yaml:"ensTransformer"`
	} `yaml:"indexer"`
	Frontend struct {
		Debug                          bool   `yaml:"debug" envconfig:"FRONTEND_DEBUG"`
		BeaconchainETHPoolBridgeSecret string `yaml:"beaconchainETHPoolBridgeSecret" envconfig:"FRONTEND_BEACONCHAIN_ETHPOOL_BRIDGE_SECRET"`
		Kong                           string `yaml:"kong" envconfig:"FRONTEND_KONG"`
		OnlyAPI                        bool   `yaml:"onlyAPI" envconfig:"FRONTEND_ONLY_API"`
		CsrfAuthKey                    string `yaml:"csrfAuthKey" envconfig:"FRONTEND_CSRF_AUTHKEY"`
		CsrfInsecure                   bool   `yaml:"csrfInsecure" envconfig:"FRONTEND_CSRF_INSECURE"`
		DisableCharts                  bool   `yaml:"disableCharts" envconfig:"disableCharts"`
		RecaptchaSiteKey               string `yaml:"recaptchaSiteKey" envconfig:"FRONTEND_RECAPTCHA_SITEKEY"`
		RecaptchaSecretKey             string `yaml:"recaptchaSecretKey" envconfig:"FRONTEND_RECAPTCHA_SECRETKEY"`
		Enabled                        bool   `yaml:"enabled" envconfig:"FRONTEND_ENABLED"`
		BlobProviderUrl                string `yaml:"blobProviderUrl" envconfig:"FRONTEND_BLOB_PROVIDER_URL"`
		SiteBrand                      string `yaml:"siteBrand" envconfig:"FRONTEND_SITE_BRAND"`
		Keywords                       string `yaml:"keywords" envconfig:"FRONTEND_KEYWORDS"`
		// Imprint is deprdecated place imprint file into the legal directory
		Imprint string `yaml:"imprint" envconfig:"FRONTEND_IMPRINT"`
		Legal   struct {
			TermsOfServiceUrl string `yaml:"termsOfServiceUrl" envconfig:"FRONTEND_LEGAL_TERMS_OF_SERVICE_URL"`
			PrivacyPolicyUrl  string `yaml:"privacyPolicyUrl" envconfig:"FRONTEND_LEGAL_PRIVACY_POLICY_URL"`
			ImprintTemplate   string `yaml:"imprintTemplate" envconfig:"FRONTEND_LEGAL_IMPRINT_TEMPLATE"`
		} `yaml:"legal"`
		SiteDomain   string `yaml:"siteDomain" envconfig:"FRONTEND_SITE_DOMAIN"`
		SiteName     string `yaml:"siteName" envconfig:"FRONTEND_SITE_NAME"`
		SiteTitle    string `yaml:"siteTitle" envconfig:"FRONTEND_SITE_TITLE"`
		SiteSubtitle string `yaml:"siteSubtitle" envconfig:"FRONTEND_SITE_SUBTITLE"`
		Server       struct {
			Port string `yaml:"port" envconfig:"FRONTEND_SERVER_PORT"`
			Host string `yaml:"host" envconfig:"FRONTEND_SERVER_HOST"`
		} `yaml:"server"`
		ReaderDatabase struct {
			Username     string `yaml:"user" envconfig:"FRONTEND_READER_DB_USERNAME"`
			Password     string `yaml:"password" envconfig:"FRONTEND_READER_DB_PASSWORD"`
			Name         string `yaml:"name" envconfig:"FRONTEND_READER_DB_NAME"`
			Host         string `yaml:"host" envconfig:"FRONTEND_READER_DB_HOST"`
			Port         string `yaml:"port" envconfig:"FRONTEND_READER_DB_PORT"`
			MaxOpenConns int    `yaml:"maxOpenConns" envconfig:"FRONTEND_READER_DB_MAX_OPEN_CONNS"`
			MaxIdleConns int    `yaml:"maxIdleConns" envconfig:"FRONTEND_READER_DB_MAX_IDLE_CONNS"`
			SSL          bool   `yaml:"ssl" envconfig:"FRONTEND_READER_DB_SSL"`
		} `yaml:"readerDatabase"`
		WriterDatabase struct {
			Username     string `yaml:"user" envconfig:"FRONTEND_WRITER_DB_USERNAME"`
			Password     string `yaml:"password" envconfig:"FRONTEND_WRITER_DB_PASSWORD"`
			Name         string `yaml:"name" envconfig:"FRONTEND_WRITER_DB_NAME"`
			Host         string `yaml:"host" envconfig:"FRONTEND_WRITER_DB_HOST"`
			Port         string `yaml:"port" envconfig:"FRONTEND_WRITER_DB_PORT"`
			MaxOpenConns int    `yaml:"maxOpenConns" envconfig:"FRONTEND_WRITER_DB_MAX_OPEN_CONNS"`
			MaxIdleConns int    `yaml:"maxIdleConns" envconfig:"FRONTEND_WRITER_DB_MAX_IDLE_CONNS"`
			SSL          bool   `yaml:"ssl" envconfig:"FRONTEND_WRITER_DB_SSL"`
		} `yaml:"writerDatabase"`
		OldProductsDeadlineUnix int64 `yaml:"oldProductsDeadline" envconfig:"FRONTEND_OLD_PRODUCTS_DEADLINE_UNIX"`
		Stripe                  struct {
			Webhook   string `yaml:"webhook" envconfig:"FRONTEND_STRIPE_WEBHOOK"`
			SecretKey string `yaml:"secretKey" envconfig:"FRONTEND_STRIPE_SECRET_KEY"`
			PublicKey string `yaml:"publicKey" envconfig:"FRONTEND_STRIPE_PUBLIC_KEY"`

			Sapphire string `yaml:"sapphire" envconfig:"FRONTEND_STRIPE_SAPPHIRE"`
			Emerald  string `yaml:"emerald" envconfig:"FRONTEND_STRIPE_EMERALD"`
			Diamond  string `yaml:"diamond" envconfig:"FRONTEND_STRIPE_DIAMOND"`
			Whale    string `yaml:"whale" envconfig:"FRONTEND_STRIPE_WHALE"`
			Goldfish string `yaml:"goldfish" envconfig:"FRONTEND_STRIPE_GOLDFISH"`
			Plankton string `yaml:"plankton" envconfig:"FRONTEND_STRIPE_PLANKTON"`

			Iron         string `yaml:"iron" envconfig:"FRONTEND_STRIPE_IRON"`
			IronYearly   string `yaml:"ironYearly" envconfig:"FRONTEND_STRIPE_IRON_YEARLY"`
			Silver       string `yaml:"silver" envconfig:"FRONTEND_STRIPE_SILVER"`
			SilverYearly string `yaml:"silverYearly" envconfig:"FRONTEND_STRIPE_SILVER_YEARLY"`
			Gold         string `yaml:"gold" envconfig:"FRONTEND_STRIPE_GOLD"`
			GoldYearly   string `yaml:"goldYearly" envconfig:"FRONTEND_STRIPE_GOLD_YEARLY"`

			Guppy         string `yaml:"guppy" envconfig:"FRONTEND_STRIPE_GUPPY"`
			GuppyYearly   string `yaml:"guppyYearly" envconfig:"FRONTEND_STRIPE_GUPPY_YEARLY"`
			Dolphin       string `yaml:"dolphin" envconfig:"FRONTEND_STRIPE_DOLPHIN"`
			DolphinYearly string `yaml:"dolphinYearly" envconfig:"FRONTEND_STRIPE_DOLPHIN_YEARLY"`
			Orca          string `yaml:"orca" envconfig:"FRONTEND_STRIPE_ORCA"`
			OrcaYearly    string `yaml:"orcaYearly" envconfig:"FRONTEND_STRIPE_ORCA_YEARLY"`

			VdbAddon1k        string `yaml:"vdbAddon1k" envconfig:"FRONTEND_STRIPE_VDB_ADDON_1K"`
			VdbAddon1kYearly  string `yaml:"vdbAddon1kYearly" envconfig:"FRONTEND_STRIPE_VDB_ADDON_1K_YEARLY"`
			VdbAddon10k       string `yaml:"vdbAddon10k" envconfig:"FRONTEND_STRIPE_VDB_ADDON_10K"`
			VdbAddon10kYearly string `yaml:"vdbAddon10kYearly" envconfig:"FRONTEND_STRIPE_VDB_ADDON_10K_YEARLY"`
		}
		RatelimitUpdateInterval              time.Duration `yaml:"ratelimitUpdateInterval" envconfig:"FRONTEND_RATELIMIT_UPDATE_INTERVAL"`
		SessionSameSiteNone                  bool          `yaml:"sessionSameSiteNone" envconfig:"FRONTEND_SESSION_SAMESITE_NONE"`
		SessionSecret                        string        `yaml:"sessionSecret" envconfig:"FRONTEND_SESSION_SECRET"`
		SessionCookieDomain                  string        `yaml:"sessionCookieDomain" envconfig:"FRONTEND_SESSION_COOKIE_DOMAIN"`
		SessionCookieDeriveDomainFromRequest bool          `yaml:"sessionCookieDeriveDomainFromRequest" envconfig:"FRONTEND_SESSION_COOKIE_DERIVE_DOMAIN_FROM_REQUEST"`
		JwtSigningSecret                     string        `yaml:"jwtSigningSecret" envconfig:"FRONTEND_JWT_SECRET"`
		JwtIssuer                            string        `yaml:"jwtIssuer" envconfig:"FRONTEND_JWT_ISSUER"`
		JwtValidityInMinutes                 int           `yaml:"jwtValidityInMinutes" envconfig:"FRONTEND_JWT_VALIDITY_INMINUTES"`
		MaxMailsPerEmailPerDay               int           `yaml:"maxMailsPerEmailPerDay" envconfig:"FRONTEND_MAX_MAIL_PER_EMAIL_PER_DAY"`
		Mail                                 struct {
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
			Contact struct {
				SupportEmail string `yaml:"supportEmail" envconfig:"FRONTEND_MAIL_CONTACT_SUPPORT_EMAIL"`
				InquiryEmail string `yaml:"inquiryEmail" envconfig:"FRONTEND_MAIL_CONTACT_INQUIRY_EMAIL"`
			} `yaml:"contact"`
		} `yaml:"mail"`
		GATag         string `yaml:"gatag" envconfig:"GATAG"`
		VerifyAppSubs bool   `yaml:"verifyAppSubscriptions" envconfig:"FRONTEND_VERIFY_APP_SUBSCRIPTIONS"`
		Apple         struct {
			LegacyAppSubsAppleSecret string `yaml:"appSubsAppleSecret" envconfig:"FRONTEND_APP_SUBS_APPLE_SECRET"`
			KeyID                    string `yaml:"keyID" envconfig:"FRONTEND_APPLE_APP_KEY_ID"`
			IssueID                  string `yaml:"issueID" envconfig:"FRONTEND_APPLE_ISSUE_ID"`
			Certificate              string `yaml:"certificate" envconfig:"FRONTEND_APPLE_CERTIFICATE"`
		} `yaml:"apple"`
		AppSubsGoogleJSONPath string `yaml:"appSubsGoogleJsonPath" envconfig:"FRONTEND_APP_SUBS_GOOGLE_JSON_PATH"`
		DisableStatsInserts   bool   `yaml:"disableStatsInserts" envconfig:"FRONTEND_DISABLE_STATS_INSERTS"`
		ShowDonors            struct {
			Enabled bool   `yaml:"enabled" envconfig:"FRONTEND_SHOW_DONORS_ENABLED"`
			URL     string `yaml:"gitcoinURL" envconfig:"FRONTEND_GITCOIN_URL"`
		} `yaml:"showDonors"`
		Countdown struct {
			Enabled   bool          `yaml:"enabled" envconfig:"FRONTEND_COUNTDOWN_ENABLED"`
			Title     template.HTML `yaml:"title" envconfig:"FRONTEND_COUNTDOWN_TITLE"`
			Timestamp uint64        `yaml:"timestamp" envconfig:"FRONTEND_COUNTDOWN_TIMESTAMP"`
			Info      string        `yaml:"info" envconfig:"FRONTEND_COUNTDOWN_INFO"`
		} `yaml:"countdown"`
		HttpReadTimeout    time.Duration `yaml:"httpReadTimeout" envconfig:"FRONTEND_HTTP_READ_TIMEOUT"`
		HttpWriteTimeout   time.Duration `yaml:"httpWriteTimeout" envconfig:"FRONTEND_HTTP_WRITE_TIMEOUT"`
		HttpIdleTimeout    time.Duration `yaml:"httpIdleTimeout" envconfig:"FRONTEND_HTTP_IDLE_TIMEOUT"`
		ClCurrency         string        `yaml:"clCurrency" envconfig:"FRONTEND_CL_CURRENCY"`
		ClCurrencyDivisor  int64         `yaml:"clCurrencyDivisor" envconfig:"FRONTEND_CL_CURRENCY_DIVISOR"`
		ClCurrencyDecimals int64         `yaml:"clCurrencyDecimals" envconfig:"FRONTEND_CL_CURRENCY_DECIMALS"`
		ElCurrency         string        `yaml:"elCurrency" envconfig:"FRONTEND_EL_CURRENCY"`
		ElCurrencyDivisor  int64         `yaml:"elCurrencyDivisor" envconfig:"FRONTEND_EL_CURRENCY_DIVISOR"`
		ElCurrencyDecimals int64         `yaml:"elCurrencyDecimals" envconfig:"FRONTEND_EL_CURRENCY_DECIMALS"`
		MainCurrency       string        `yaml:"mainCurrency" envconfig:"FRONTEND_MAIN_CURRENCY"`
	} `yaml:"frontend"`
	Metrics struct {
		Enabled bool   `yaml:"enabled" envconfig:"METRICS_ENABLED"`
		Address string `yaml:"address" envconfig:"METRICS_ADDRESS"`
		Pprof   bool   `yaml:"pprof" envconfig:"METRICS_PPROF"`
	} `yaml:"metrics"`
	Notifications struct {
		UserDBNotifications                           bool    `yaml:"userDbNotifications" envconfig:"USERDB_NOTIFICATIONS_ENABLED"`
		FirebaseCredentialsPath                       string  `yaml:"firebaseCredentialsPath" envconfig:"NOTIFICATIONS_FIREBASE_CRED_PATH"`
		ValidatorBalanceDecreasedNotificationsEnabled bool    `yaml:"validatorBalanceDecreasedNotificationsEnabled" envconfig:"VALIDATOR_BALANCE_DECREASED_NOTIFICATIONS_ENABLED"`
		PubkeyCachePath                               string  `yaml:"pubkeyCachePath" envconfig:"NOTIFICATIONS_PUBKEY_CACHE_PATH"`
		OnlineDetectionLimit                          int     `yaml:"onlineDetectionLimit" envconfig:"ONLINE_DETECTION_LIMIT"`
		OfflineDetectionLimit                         int     `yaml:"offlineDetectionLimit" envconfig:"OFFLINE_DETECTION_LIMIT"`
		MachineEventThreshold                         uint64  `yaml:"machineEventThreshold" envconfig:"MACHINE_EVENT_THRESHOLD"`
		MachineEventFirstRatioThreshold               float64 `yaml:"machineEventFirstRatioThreshold" envconfig:"MACHINE_EVENT_FIRST_RATIO_THRESHOLD"`
		MachineEventSecondRatioThreshold              float64 `yaml:"machineEventSecondRatioThreshold" envconfig:"MACHINE_EVENT_SECOND_RATIO_THRESHOLD"`
	} `yaml:"notifications"`
	RatelimitUpdater struct {
		Enabled        bool          `yaml:"enabled" envconfig:"RATELIMIT_UPDATER_ENABLED"`
		UpdateInterval time.Duration `yaml:"updateInterval" envconfig:"RATELIMIT_UPDATER_UPDATE_INTERVAL"`
	} `yaml:"ratelimitUpdater"`
	SSVExporter struct {
		Enabled bool   `yaml:"enabled" envconfig:"SSV_EXPORTER_ENABLED"`
		Address string `yaml:"address" envconfig:"SSV_EXPORTER_ADDRESS"`
	} `yaml:"SSVExporter"`
	RocketpoolExporter struct {
		Enabled bool `yaml:"enabled" envconfig:"ROCKETPOOL_EXPORTER_ENABLED"`
	} `yaml:"rocketpoolExporter"`
	MevBoostRelayExporter struct {
		Enabled bool `yaml:"enabled" envconfig:"MEVBOOSTRELAY_EXPORTER_ENABLED"`
	} `yaml:"mevBoostRelayExporter"`
	Pprof struct {
		Enabled bool   `yaml:"enabled" envconfig:"PPROF_ENABLED"`
		Port    string `yaml:"port" envconfig:"PPROF_PORT"`
	} `yaml:"pprof"`
	NodeJobsProcessor struct {
		ElEndpoint string `yaml:"elEndpoint" envconfig:"NODE_JOBS_PROCESSOR_EL_ENDPOINT"`
		ClEndpoint string `yaml:"clEndpoint" envconfig:"NODE_JOBS_PROCESSOR_CL_ENDPOINT"`
	} `yaml:"nodeJobsProcessor"`
	Monitoring struct {
		ApiKey                          string                           `yaml:"apiKey" envconfig:"MONITORING_API_KEY"`
		ServiceMonitoringConfigurations []ServiceMonitoringConfiguration `yaml:"serviceMonitoringConfigurations" envconfig:"SERVICE_MONITORING_CONFIGURATIONS"`
	} `yaml:"monitoring"`
	GithubApiHost string `yaml:"githubApiHost" envconfig:"GITHUB_API_HOST"`
}

type DatabaseConfig struct {
	Username     string
	Password     string
	Name         string
	Host         string
	Port         string
	MaxOpenConns int
	MaxIdleConns int
	SSL          bool
}

type ServiceMonitoringConfiguration struct {
	Name     string        `yaml:"name" envconfig:"NAME"`
	Duration time.Duration `yaml:"duration" envconfig:"DURATION"`
}

type ConfigJsonResponse struct {
	Data struct {
		ConfigName                              string `json:"CONFIG_NAME"`
		PresetBase                              string `json:"PRESET_BASE"`
		TerminalTotalDifficulty                 string `json:"TERMINAL_TOTAL_DIFFICULTY"`
		TerminalBlockHash                       string `json:"TERMINAL_BLOCK_HASH"`
		TerminalBlockHashActivationEpoch        string `json:"TERMINAL_BLOCK_HASH_ACTIVATION_EPOCH"`
		SafeSlotsToImportOptimistically         string `json:"SAFE_SLOTS_TO_IMPORT_OPTIMISTICALLY"`
		MinGenesisActiveValidatorCount          string `json:"MIN_GENESIS_ACTIVE_VALIDATOR_COUNT"`
		MinGenesisTime                          string `json:"MIN_GENESIS_TIME"`
		GenesisForkVersion                      string `json:"GENESIS_FORK_VERSION"`
		GenesisDelay                            string `json:"GENESIS_DELAY"`
		AltairForkVersion                       string `json:"ALTAIR_FORK_VERSION"`
		AltairForkEpoch                         string `json:"ALTAIR_FORK_EPOCH"`
		BellatrixForkVersion                    string `json:"BELLATRIX_FORK_VERSION"`
		BellatrixForkEpoch                      string `json:"BELLATRIX_FORK_EPOCH"`
		CapellaForkVersion                      string `json:"CAPELLA_FORK_VERSION"`
		CapellaForkEpoch                        string `json:"CAPELLA_FORK_EPOCH"`
		DenebForkVersion                        string `json:"DENEB_FORK_VERSION"`
		DenebForkEpoch                          string `json:"DENEB_FORK_EPOCH"`
		SecondsPerSlot                          string `json:"SECONDS_PER_SLOT"`
		SecondsPerEth1Block                     string `json:"SECONDS_PER_ETH1_BLOCK"`
		MinValidatorWithdrawabilityDelay        string `json:"MIN_VALIDATOR_WITHDRAWABILITY_DELAY"`
		ShardCommitteePeriod                    string `json:"SHARD_COMMITTEE_PERIOD"`
		Eth1FollowDistance                      string `json:"ETH1_FOLLOW_DISTANCE"`
		SubnetsPerNode                          string `json:"SUBNETS_PER_NODE"`
		InactivityScoreBias                     string `json:"INACTIVITY_SCORE_BIAS"`
		InactivityScoreRecoveryRate             string `json:"INACTIVITY_SCORE_RECOVERY_RATE"`
		EjectionBalance                         string `json:"EJECTION_BALANCE"`
		MinPerEpochChurnLimit                   string `json:"MIN_PER_EPOCH_CHURN_LIMIT"`
		ChurnLimitQuotient                      string `json:"CHURN_LIMIT_QUOTIENT"`
		MaxPerEpochActivationChurnLimit         string `json:"MAX_PER_EPOCH_ACTIVATION_CHURN_LIMIT"`
		ProposerScoreBoost                      string `json:"PROPOSER_SCORE_BOOST"`
		DepositChainID                          string `json:"DEPOSIT_CHAIN_ID"`
		DepositNetworkID                        string `json:"DEPOSIT_NETWORK_ID"`
		DepositContractAddress                  string `json:"DEPOSIT_CONTRACT_ADDRESS"`
		MaxCommitteesPerSlot                    string `json:"MAX_COMMITTEES_PER_SLOT"`
		TargetCommitteeSize                     string `json:"TARGET_COMMITTEE_SIZE"`
		MaxValidatorsPerCommittee               string `json:"MAX_VALIDATORS_PER_COMMITTEE"`
		ShuffleRoundCount                       string `json:"SHUFFLE_ROUND_COUNT"`
		HysteresisQuotient                      string `json:"HYSTERESIS_QUOTIENT"`
		HysteresisDownwardMultiplier            string `json:"HYSTERESIS_DOWNWARD_MULTIPLIER"`
		HysteresisUpwardMultiplier              string `json:"HYSTERESIS_UPWARD_MULTIPLIER"`
		SafeSlotsToUpdateJustified              string `json:"SAFE_SLOTS_TO_UPDATE_JUSTIFIED"`
		MinDepositAmount                        string `json:"MIN_DEPOSIT_AMOUNT"`
		MaxEffectiveBalance                     string `json:"MAX_EFFECTIVE_BALANCE"`
		EffectiveBalanceIncrement               string `json:"EFFECTIVE_BALANCE_INCREMENT"`
		MinAttestationInclusionDelay            string `json:"MIN_ATTESTATION_INCLUSION_DELAY"`
		SlotsPerEpoch                           string `json:"SLOTS_PER_EPOCH"`
		MinSeedLookahead                        string `json:"MIN_SEED_LOOKAHEAD"`
		MaxSeedLookahead                        string `json:"MAX_SEED_LOOKAHEAD"`
		EpochsPerEth1VotingPeriod               string `json:"EPOCHS_PER_ETH1_VOTING_PERIOD"`
		SlotsPerHistoricalRoot                  string `json:"SLOTS_PER_HISTORICAL_ROOT"`
		MinEpochsToInactivityPenalty            string `json:"MIN_EPOCHS_TO_INACTIVITY_PENALTY"`
		EpochsPerHistoricalVector               string `json:"EPOCHS_PER_HISTORICAL_VECTOR"`
		EpochsPerSlashingsVector                string `json:"EPOCHS_PER_SLASHINGS_VECTOR"`
		HistoricalRootsLimit                    string `json:"HISTORICAL_ROOTS_LIMIT"`
		ValidatorRegistryLimit                  string `json:"VALIDATOR_REGISTRY_LIMIT"`
		BaseRewardFactor                        string `json:"BASE_REWARD_FACTOR"`
		WhistleblowerRewardQuotient             string `json:"WHISTLEBLOWER_REWARD_QUOTIENT"`
		ProposerRewardQuotient                  string `json:"PROPOSER_REWARD_QUOTIENT"`
		InactivityPenaltyQuotient               string `json:"INACTIVITY_PENALTY_QUOTIENT"`
		MinSlashingPenaltyQuotient              string `json:"MIN_SLASHING_PENALTY_QUOTIENT"`
		ProportionalSlashingMultiplier          string `json:"PROPORTIONAL_SLASHING_MULTIPLIER"`
		MaxProposerSlashings                    string `json:"MAX_PROPOSER_SLASHINGS"`
		MaxAttesterSlashings                    string `json:"MAX_ATTESTER_SLASHINGS"`
		MaxAttestations                         string `json:"MAX_ATTESTATIONS"`
		MaxDeposits                             string `json:"MAX_DEPOSITS"`
		MaxVoluntaryExits                       string `json:"MAX_VOLUNTARY_EXITS"`
		InactivityPenaltyQuotientAltair         string `json:"INACTIVITY_PENALTY_QUOTIENT_ALTAIR"`
		MinSlashingPenaltyQuotientAltair        string `json:"MIN_SLASHING_PENALTY_QUOTIENT_ALTAIR"`
		ProportionalSlashingMultiplierAltair    string `json:"PROPORTIONAL_SLASHING_MULTIPLIER_ALTAIR"`
		SyncCommitteeSize                       string `json:"SYNC_COMMITTEE_SIZE"`
		EpochsPerSyncCommitteePeriod            string `json:"EPOCHS_PER_SYNC_COMMITTEE_PERIOD"`
		MinSyncCommitteeParticipants            string `json:"MIN_SYNC_COMMITTEE_PARTICIPANTS"`
		InactivityPenaltyQuotientBellatrix      string `json:"INACTIVITY_PENALTY_QUOTIENT_BELLATRIX"`
		MinSlashingPenaltyQuotientBellatrix     string `json:"MIN_SLASHING_PENALTY_QUOTIENT_BELLATRIX"`
		ProportionalSlashingMultiplierBellatrix string `json:"PROPORTIONAL_SLASHING_MULTIPLIER_BELLATRIX"`
		MaxBytesPerTransaction                  string `json:"MAX_BYTES_PER_TRANSACTION"`
		MaxTransactionsPerPayload               string `json:"MAX_TRANSACTIONS_PER_PAYLOAD"`
		BytesPerLogsBloom                       string `json:"BYTES_PER_LOGS_BLOOM"`
		MaxExtraDataBytes                       string `json:"MAX_EXTRA_DATA_BYTES"`
		MaxBlsToExecutionChanges                string `json:"MAX_BLS_TO_EXECUTION_CHANGES"`
		MaxWithdrawalsPerPayload                string `json:"MAX_WITHDRAWALS_PER_PAYLOAD"`
		MaxValidatorsPerWithdrawalsSweep        string `json:"MAX_VALIDATORS_PER_WITHDRAWALS_SWEEP"`
		DomainAggregateAndProof                 string `json:"DOMAIN_AGGREGATE_AND_PROOF"`
		TargetAggregatorsPerSyncSubcommittee    string `json:"TARGET_AGGREGATORS_PER_SYNC_SUBCOMMITTEE"`
		SyncCommitteeSubnetCount                string `json:"SYNC_COMMITTEE_SUBNET_COUNT"`
		BlsWithdrawalPrefix                     string `json:"BLS_WITHDRAWAL_PREFIX"`
		DomainRandao                            string `json:"DOMAIN_RANDAO"`
		DomainVoluntaryExit                     string `json:"DOMAIN_VOLUNTARY_EXIT"`
		DomainSyncCommitteeSelectionProof       string `json:"DOMAIN_SYNC_COMMITTEE_SELECTION_PROOF"`
		DomainBeaconAttester                    string `json:"DOMAIN_BEACON_ATTESTER"`
		DomainBeaconProposer                    string `json:"DOMAIN_BEACON_PROPOSER"`
		DomainDeposit                           string `json:"DOMAIN_DEPOSIT"`
		DomainSelectionProof                    string `json:"DOMAIN_SELECTION_PROOF"`
		DomainSyncCommittee                     string `json:"DOMAIN_SYNC_COMMITTEE"`
		TargetAggregatorsPerCommittee           string `json:"TARGET_AGGREGATORS_PER_COMMITTEE"`
		DomainContributionAndProof              string `json:"DOMAIN_CONTRIBUTION_AND_PROOF"`
		DomainApplicationMask                   string `json:"DOMAIN_APPLICATION_MASK"`
	} `json:"data"`
}
