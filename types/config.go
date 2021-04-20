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
		// Deprecated Use Phase0 config CONFIG_NAME
		Network string `yaml:"network" envconfig:"CHAIN_NETWORK"`
		// Deprecated Use Phase0 config SLOTS_PER_EPOCH
		SlotsPerEpoch uint64 `yaml:"slotsPerEpoch" envconfig:"CHAIN_SLOTS_PER_EPOCH"`
		// Deprecated Use Phase0 config SECONDS_PER_SLOT
		SecondsPerSlot uint64 `yaml:"secondsPerSlot" envconfig:"CHAIN_SECONDS_PER_SLOT"`
		// Deprecated Use Phase0 config GENESIS_TIMESTMAP
		GenesisTimestamp uint64 `yaml:"genesisTimestamp" envconfig:"CHAIN_GENESIS_TIMESTAMP"`
		// Deprecated Use Phase0 config MIN_GENESIS_ACTIVE_VALIDATOR_COUNT
		MinGenesisActiveValidatorCount uint64 `yaml:"minGenesisActiveValidatorCount" envconfig:"CHAIN_MIN_GENESIS_ACTIVE_VALIDATOR_COUNT"`
		// Deprecated Use Phase0 config GENESIS_DELAY
		GenesisDelay uint64 `yaml:"genesisDelay" envconfig:"CHAIN_GENESIS_DELAY"`
		// Deprecated Use Phase0 config CONFIG_NAME == "mainnet"
		Mainnet    bool   `yaml:"mainnet" envconfig:"CHAIN_MAINNET"`
		Phase0Path string `yaml:"phase0path" envconfig:"CHAIN_PHASE0_PATH"`
		Phase0
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
	} `yaml:"indexer"`
	Frontend struct {
		Kong               string `yaml:"kong" envconfig:"FRONTEND_KONG"`
		OnlyAPI            bool   `yaml:"onlyAPI" envconfig:"FRONTEND_ONLY_API"`
		CsrfAuthKey        string `yaml:"csrfAuthKey" envconfig:"FRONTEND_CSRF_AUTHKEY`
		CsrfInsecure       bool   `yaml:"csrfInsecure" envconfig:"FRONTEND_CSRF_INSECURE"`
		DisableCharts      bool   `yaml:"disableCharts" envconfig:"disableCharts"`
		RecaptchaSiteKey   string `yaml:"recaptchaSiteKey" envconfig:"FRONTEND_RECAPTCHA_SITEKEY"`
		RecaptchaSecretKey string `yaml:"recaptchaSecretKey" envconfig:"FRONTEND_RECAPTCHA_SECRETKEY"`
		Enabled            bool   `yaml:"enabled" envconfig:"FRONTEND_ENABLED"`
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
		Database struct {
			Username string `yaml:"user" envconfig:"FRONTEND_DB_USERNAME"`
			Password string `yaml:"password" envconfig:"FRONTEND_DB_PASSWORD"`
			Name     string `yaml:"name" envconfig:"FRONTEND_DB_NAME"`
			Host     string `yaml:"host" envconfig:"FRONTEND_DB_HOST"`
			Port     string `yaml:"port" envconfig:"FRONTEND_DB_PORT"`
		} `yaml:"database"`
		Stripe struct {
			SecretKey string `yaml:"secretKey" envconfig:"FRONTEND_STRIPE_SECRET_KEY"`
			PublicKey string `yaml:"publicKey" envconfig:"FRONTEND_STRIPE_PUBLIC_KEY"`
			Sapphire  string `yaml:"sapphire" envconfig:"FRONTEND_STRIPE_SAPPHIRE"`
			Emerald   string `yaml:"emerald" envconfig:"FRONTEND_STRIPE_EMERALD"`
			Diamond   string `yaml:"diamond" envconfig:"FRONTEND_STRIPE_DIAMOND"`
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
		GATag      string `yaml:"gatag"  envconfig:"GATAG"`
		ShowDonors struct {
			Enabled bool   `yaml:"enabled" envconfig:"FRONTEND_SHOW_DONORS_ENABLED"`
			URL     string `yaml:"gitcoinURL" envconfig:"FRONTEND_GITCOIN_URL"`
		} `yaml:"showDonors"`
	} `yaml:"frontend"`
	Metrics struct {
		Enabled bool   `yaml:"enabled" envconfig:"METRICS_ENABLED"`
		Address string `yaml:"address" envconfig:"METRICS_ADDRESS"`
	} `yaml:"metrics"`
	Notifications struct {
		Enabled                 bool   `yaml:"enabled" envconfig:"FRONTEND_NOTIFICATIONS_ENABLED"`
		FirebaseCredentialsPath string `yaml:"firebaseCredentialsPath" envconfig:"FRONTEND_NOTIFICATIONS_FIREBASE_CRED_PATH"`
	} `yaml:"notifications"`
}

// Phase0 is the config for beacon chain phase0
type Phase0 struct {
	ConfigName string `yaml:"CONFIG_NAME"` // the name of the configuration e.g. mainnet

	// Misc constants.
	MaxCommitteesPerSlot           uint64 `yaml:"MAX_COMMITTEES_PER_SLOT"`            // MaxCommitteesPerSlot defines the max amount of committee in a single slot.
	TargetCommitteeSize            uint64 `yaml:"TARGET_COMMITTEE_SIZE"`              // TargetCommitteeSize is the number of validators in a committee when the chain is healthy.
	MaxValidatorsPerCommittee      uint64 `yaml:"MAX_VALIDATORS_PER_COMMITTEE"`       // MaxValidatorsPerCommittee defines the upper bound of the size of a committee.
	MinPerEpochChurnLimit          uint64 `yaml:"MIN_PER_EPOCH_CHURN_LIMIT"`          // MinPerEpochChurnLimit is the minimum amount of churn allotted for validator rotations.
	ChurnLimitQuotient             uint64 `yaml:"CHURN_LIMIT_QUOTIENT"`               // ChurnLimitQuotient is used to determine the limit of how many validators can rotate per epoch.
	ShuffleRoundCount              uint64 `yaml:"SHUFFLE_ROUND_COUNT"`                // ShuffleRoundCount is used for retrieving the permuted index.
	MinGenesisActiveValidatorCount uint64 `yaml:"MIN_GENESIS_ACTIVE_VALIDATOR_COUNT"` // MinGenesisActiveValidatorCount defines how many validator deposits needed to kick off beacon chain.
	MinGenesisTime                 uint64 `yaml:"MIN_GENESIS_TIME"`                   // MinGenesisTime is the time that needed to pass before kicking off beacon chain.
	HysteresisQuotient             uint64 `yaml:"HYSTERESIS_QUOTIENT"`                // HysteresisQuotient defines the hysteresis quotient for effective balance calculations.
	HysteresisDownwardMultiplier   uint64 `yaml:"HYSTERESIS_DOWNWARD_MULTIPLIER"`     // HysteresisDownwardMultiplier defines the hysteresis downward multiplier for effective balance calculations.
	HysteresisUpwardMultiplier     uint64 `yaml:"HYSTERESIS_UPWARD_MULTIPLIER"`       // HysteresisUpwardMultiplier defines the hysteresis upward multiplier for effective balance calculations.

	// Fork Choice
	SafeSlotsToUpdateJustified uint64 `yaml:"SAFE_SLOTS_TO_UPDATE_JUSTIFIED"` // SafeSlotsToUpdateJustified is the minimal slots needed to update justified check point.

	// Validator
	Eth1FollowDistance                uint64 `yaml:"ETH1_FOLLOW_DISTANCE"`             // Eth1FollowDistance is the number of eth1.0 blocks to wait before considering a new deposit for voting. This only applies after the chain as been started.
	TargetAggregatorsPerCommittee     uint64 `yaml:"TARGET_AGGREGATORS_PER_COMMITTEE"` // TargetAggregatorsPerCommittee defines the number of aggregators inside one committee.
	RandomSubnetsPerValidator         uint64 `yaml:"RANDOM_SUBNETS_PER_VALIDATOR"`
	EpochsPerRandomSubnetSubscription uint64 `yaml:"EPOCHS_PER_RANDOM_SUBNET_SUBSCRIPTION"`
	SecondsPerETH1Block               uint64 `yaml:"SECONDS_PER_ETH1_BLOCK"` // SecondsPerETH1Block is the approximate time for a single eth1 block to be produced.

	// Deposit Contract
	DepositChainID         uint64 `yaml:"DEPOSIT_CHAIN_ID"`
	DepositNetworkID       uint64 `yaml:"DEPOSIT_NETWORK_ID"`
	DepositContractAddress string `yaml:"DEPOSIT_CONTRACT_ADDRESS"`

	// Gwei value constants.
	MinDepositAmount          uint64 `yaml:"MIN_DEPOSIT_AMOUNT"`          // MinDepositAmount is the minimum amount of Gwei a validator can send to the deposit contract at once (lower amounts will be reverted).
	MaxEffectiveBalance       uint64 `yaml:"MAX_EFFECTIVE_BALANCE"`       // MaxEffectiveBalance is the maximal amount of Gwei that is effective for staking.
	EjectionBalance           uint64 `yaml:"EJECTION_BALANCE"`            // EjectionBalance is the minimal GWei a validator needs to have before ejected.
	EffectiveBalanceIncrement uint64 `yaml:"EFFECTIVE_BALANCE_INCREMENT"` // EffectiveBalanceIncrement is used for converting the high balance into the low balance for validators.

	// Initial values
	// GenesisForkVersion
	// BLSWithdrawalPrefix

	// Time parameters constants.
	GenesisDelay                     uint64 `yaml:"GENESIS_DELAY"`                       // GenesisDelay is the minimum number of seconds to delay starting the ETH2 genesis. Must be at least 1 second.
	SecondsPerSlot                   uint64 `yaml:"SECONDS_PER_SLOT"`                    // SecondsPerSlot is how many seconds are in a single slot.
	MinAttestationInclusionDelay     uint64 `yaml:"MIN_ATTESTATION_INCLUSION_DELAY"`     // MinAttestationInclusionDelay defines how many slots validator has to wait to include attestation for beacon block.
	SlotsPerEpoch                    uint64 `yaml:"SLOTS_PER_EPOCH"`                     // SlotsPerEpoch is the number of slots in an epoch.
	MinSeedLookahead                 uint64 `yaml:"MIN_SEED_LOOKAHEAD"`                  // MinSeedLookahead is the duration of randao look ahead seed.
	MaxSeedLookahead                 uint64 `yaml:"MAX_SEED_LOOKAHEAD"`                  // MaxSeedLookahead is the duration a validator has to wait for entry and exit in epoch.
	EpochsPerEth1VotingPeriod        uint64 `yaml:"EPOCHS_PER_ETH1_VOTING_PERIOD"`       // EpochsPerEth1VotingPeriod defines how often the merkle root of deposit receipts get updated in beacon node on per epoch basis.
	SlotsPerHistoricalRoot           uint64 `yaml:"SLOTS_PER_HISTORICAL_ROOT"`           // SlotsPerHistoricalRoot defines how often the historical root is saved.
	MinValidatorWithdrawabilityDelay uint64 `yaml:"MIN_VALIDATOR_WITHDRAWABILITY_DELAY"` // MinValidatorWithdrawabilityDelay is the shortest amount of time a validator has to wait to withdraw.
	ShardCommitteePeriod             uint64 `yaml:"SHARD_COMMITTEE_PERIOD"`              // ShardCommitteePeriod is the minimum amount of epochs a validator must participate before exiting.
	MinEpochsToInactivityPenalty     uint64 `yaml:"MIN_EPOCHS_TO_INACTIVITY_PENALTY"`    // MinEpochsToInactivityPenalty defines the minimum amount of epochs since finality to begin penalizing inactivity.

	// State vector lengths
	EpochsPerHistoricalVector uint64 `yaml:"EPOCHS_PER_HISTORICAL_VECTOR"`
	EpochsPerSlashingsVector  uint64 `yaml:"EPOCHS_PER_SLASHINGS_VECTOR"`
	HistoricalRootsLimit      uint64 `yaml:"HISTORICAL_ROOTS_LIMIT"`
	ValidatorRegistryLimit    uint64 `yaml:"VALIDATOR_REGISTRY_LIMIT"`

	// Reward and penalty quotients
	BaseRewardFactor               uint64 `yaml:"BASE_REWARD_FACTOR"`
	WhistleblowerRewardQuotient    uint64 `yaml:"WHISTLEBLOWER_REWARD_QUOTIENT"`
	ProposerRewardQuotient         uint64 `yaml:"PROPOSER_REWARD_QUOTIENT"`
	InactivityPenaltyQuotient      uint64 `yaml:"INACTIVITY_PENALTY_QUOTIENT"`
	MinSlashingPenaltyQuotient     uint64 `yaml:"MIN_SLASHING_PENALTY_QUOTIENT"`
	PorportionalSlashingMultiplier uint64 `yaml:"PROPORTIONAL_SLASHING_MULTIPLIER"`

	// Max Operations per blockconst
	MaxProposerSlashings uint64 `yaml:"MAX_PROPOSER_SLASHINGS"`
	MaxAttesterSlashings uint64 `yaml:"MAX_ATTESTER_SLASHINGS"`
	MaxAttestations      uint64 `yaml:"MAX_ATTESTATIONS"`
	MaxDeposits          uint64 `yaml:"MAX_DEPOSITS"`
	MaxVoluntaryExits    uint64 `yaml:"MAX_VOLUNTARY_EXITS"`

	// Signature domains
	// DomainBeaconProposer
	// DomainBeaconAttester
	// DomainRandao
	// DomainDeposit
	// DomainVoluntaryExit
	// DomainSelectionProof
	// DomainAggregateAndProof
}
