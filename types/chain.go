package types

type ForkVersion struct {
	Epoch           uint64
	CurrentVersion  []byte
	PreviousVersion []byte
}

// https://github.com/ethereum/consensus-specs/blob/dev/configs/mainnet.yaml
type ChainConfig struct {
	PresetBase                       string `yaml:"PRESET_BASE"`
	ConfigName                       string `yaml:"CONFIG_NAME"`
	TerminalTotalDifficulty          string `yaml:"TERMINAL_TOTAL_DIFFICULTY"`
	TerminalBlockHash                string `yaml:"TERMINAL_BLOCK_HASH"`
	TerminalBlockHashActivationEpoch uint64 `yaml:"TERMINAL_BLOCK_HASH_ACTIVATION_EPOCH"`
	MinGenesisActiveValidatorCount   uint64 `yaml:"MIN_GENESIS_ACTIVE_VALIDATOR_COUNT"`
	MinGenesisTime                   int64  `yaml:"MIN_GENESIS_TIME"`
	GenesisForkVersion               string `yaml:"GENESIS_FORK_VERSION"`
	GenesisDelay                     uint64 `yaml:"GENESIS_DELAY"`
	AltairForkVersion                string `yaml:"ALTAIR_FORK_VERSION"`
	AltairForkEpoch                  uint64 `yaml:"ALTAIR_FORK_EPOCH"`
	BellatrixForkVersion             string `yaml:"BELLATRIX_FORK_VERSION"`
	BellatrixForkEpoch               uint64 `yaml:"BELLATRIX_FORK_EPOCH"`
	CappellaForkVersion              string `yaml:"CAPELLA_FORK_VERSION"`
	CappellaForkEpoch                uint64 `yaml:"CAPELLA_FORK_EPOCH"`
	DenebForkVersion                 string `yaml:"DENEB_FORK_VERSION"`
	DenebForkEpoch                   uint64 `yaml:"DENEB_FORK_EPOCH"`
	ShardingForkVersion              string `yaml:"SHARDING_FORK_VERSION"`
	ShardingForkEpoch                uint64 `yaml:"SHARDING_FORK_EPOCH"`
	SecondsPerSlot                   uint64 `yaml:"SECONDS_PER_SLOT"`
	SecondsPerEth1Block              uint64 `yaml:"SECONDS_PER_ETH1_BLOCK"`
	MinValidatorWithdrawabilityDelay uint64 `yaml:"MIN_VALIDATOR_WITHDRAWABILITY_DELAY"`
	ShardCommitteePeriod             uint64 `yaml:"SHARD_COMMITTEE_PERIOD"`
	Eth1FollowDistance               uint64 `yaml:"ETH1_FOLLOW_DISTANCE"`
	InactivityScoreBias              uint64 `yaml:"INACTIVITY_SCORE_BIAS"`
	InactivityScoreRecoveryRate      uint64 `yaml:"INACTIVITY_SCORE_RECOVERY_RATE"`
	EjectionBalance                  uint64 `yaml:"EJECTION_BALANCE"`
	MinPerEpochChurnLimit            uint64 `yaml:"MIN_PER_EPOCH_CHURN_LIMIT"`
	ChurnLimitQuotient               uint64 `yaml:"CHURN_LIMIT_QUOTIENT"`
	ProposerScoreBoost               uint64 `yaml:"PROPOSER_SCORE_BOOST"`
	DepositChainID                   uint64 `yaml:"DEPOSIT_CHAIN_ID"`
	DepositNetworkID                 uint64 `yaml:"DEPOSIT_NETWORK_ID"`
	DepositContractAddress           string `yaml:"DEPOSIT_CONTRACT_ADDRESS"`

	// phase0
	// https://github.com/ethereum/consensus-specs/blob/dev/presets/mainnet/phase0.yaml
	MaxCommitteesPerSlot           uint64 `yaml:"MAX_COMMITTEES_PER_SLOT"`
	TargetCommitteeSize            uint64 `yaml:"TARGET_COMMITTEE_SIZE"`
	MaxValidatorsPerCommittee      uint64 `yaml:"MAX_VALIDATORS_PER_COMMITTEE"`
	ShuffleRoundCount              uint64 `yaml:"SHUFFLE_ROUND_COUNT"`
	HysteresisQuotient             uint64 `yaml:"HYSTERESIS_QUOTIENT"`
	HysteresisDownwardMultiplier   uint64 `yaml:"HYSTERESIS_DOWNWARD_MULTIPLIER"`
	HysteresisUpwardMultiplier     uint64 `yaml:"HYSTERESIS_UPWARD_MULTIPLIER"`
	SafeSlotsToUpdateJustified     uint64 `yaml:"SAFE_SLOTS_TO_UPDATE_JUSTIFIED"`
	MinDepositAmount               uint64 `yaml:"MIN_DEPOSIT_AMOUNT"`
	MaxEffectiveBalance            uint64 `yaml:"MAX_EFFECTIVE_BALANCE"`
	EffectiveBalanceIncrement      uint64 `yaml:"EFFECTIVE_BALANCE_INCREMENT"`
	MinAttestationInclusionDelay   uint64 `yaml:"MIN_ATTESTATION_INCLUSION_DELAY"`
	SlotsPerEpoch                  uint64 `yaml:"SLOTS_PER_EPOCH"`
	MinSeedLookahead               uint64 `yaml:"MIN_SEED_LOOKAHEAD"`
	MaxSeedLookahead               uint64 `yaml:"MAX_SEED_LOOKAHEAD"`
	EpochsPerEth1VotingPeriod      uint64 `yaml:"EPOCHS_PER_ETH1_VOTING_PERIOD"`
	SlotsPerHistoricalRoot         uint64 `yaml:"SLOTS_PER_HISTORICAL_ROOT"`
	MinEpochsToInactivityPenalty   uint64 `yaml:"MIN_EPOCHS_TO_INACTIVITY_PENALTY"`
	EpochsPerHistoricalVector      uint64 `yaml:"EPOCHS_PER_HISTORICAL_VECTOR"`
	EpochsPerSlashingsVector       uint64 `yaml:"EPOCHS_PER_SLASHINGS_VECTOR"`
	HistoricalRootsLimit           uint64 `yaml:"HISTORICAL_ROOTS_LIMIT"`
	ValidatorRegistryLimit         uint64 `yaml:"VALIDATOR_REGISTRY_LIMIT"`
	BaseRewardFactor               uint64 `yaml:"BASE_REWARD_FACTOR"`
	WhistleblowerRewardQuotient    uint64 `yaml:"WHISTLEBLOWER_REWARD_QUOTIENT"`
	ProposerRewardQuotient         uint64 `yaml:"PROPOSER_REWARD_QUOTIENT"`
	InactivityPenaltyQuotient      uint64 `yaml:"INACTIVITY_PENALTY_QUOTIENT"`
	MinSlashingPenaltyQuotient     uint64 `yaml:"MIN_SLASHING_PENALTY_QUOTIENT"`
	ProportionalSlashingMultiplier uint64 `yaml:"PROPORTIONAL_SLASHING_MULTIPLIER"`
	MaxProposerSlashings           uint64 `yaml:"MAX_PROPOSER_SLASHINGS"`
	MaxAttesterSlashings           uint64 `yaml:"MAX_ATTESTER_SLASHINGS"`
	MaxAttestations                uint64 `yaml:"MAX_ATTESTATIONS"`
	MaxDeposits                    uint64 `yaml:"MAX_DEPOSITS"`
	MaxVoluntaryExits              uint64 `yaml:"MAX_VOLUNTARY_EXITS"`

	// altair
	// https://github.com/ethereum/consensus-specs/blob/dev/presets/mainnet/altair.yaml
	InvactivityPenaltyQuotientAltair     uint64 `yaml:"INACTIVITY_PENALTY_QUOTIENT_ALTAIR"`
	MinSlashingPenaltyQuotientAltair     uint64 `yaml:"MIN_SLASHING_PENALTY_QUOTIENT_ALTAIR"`
	ProportionalSlashingMultiplierAltair uint64 `yaml:"PROPORTIONAL_SLASHING_MULTIPLIER_ALTAIR"`
	SyncCommitteeSize                    uint64 `yaml:"SYNC_COMMITTEE_SIZE"`
	EpochsPerSyncCommitteePeriod         uint64 `yaml:"EPOCHS_PER_SYNC_COMMITTEE_PERIOD"`
	MinSyncCommitteeParticipants         uint64 `yaml:"MIN_SYNC_COMMITTEE_PARTICIPANTS"`

	// bellatrix
	// https://github.com/ethereum/consensus-specs/blob/dev/presets/mainnet/bellatrix.yaml
	InvactivityPenaltyQuotientBellatrix     uint64 `yaml:"INACTIVITY_PENALTY_QUOTIENT_BELLATRIX"`
	MinSlashingPenaltyQuotientBellatrix     uint64 `yaml:"MIN_SLASHING_PENALTY_QUOTIENT_BELLATRIX"`
	ProportionalSlashingMultiplierBellatrix uint64 `yaml:"PROPORTIONAL_SLASHING_MULTIPLIER_BELLATRIX"`
	MaxBytesPerTransaction                  uint64 `yaml:"MAX_BYTES_PER_TRANSACTION"`
	MaxTransactionsPerPayload               uint64 `yaml:"MAX_TRANSACTIONS_PER_PAYLOAD"`
	BytesPerLogsBloom                       uint64 `yaml:"BYTES_PER_LOGS_BLOOM"`
	MaxExtraDataBytes                       uint64 `yaml:"MAX_EXTRA_DATA_BYTES"`

	// capella
	// https://github.com/ethereum/consensus-specs/blob/dev/presets/minimal/capella.yaml
	MaxWithdrawalsPerPayload        uint64 `yaml:"MAX_WITHDRAWALS_PER_PAYLOAD"`
	MaxValidatorsPerWithdrawalSweep uint64 `yaml:"MAX_VALIDATORS_PER_WITHDRAWALS_SWEEP"`
	MaxBlsToExecutionChange         uint64 `yaml:"MAX_BLS_TO_EXECUTION_CHANGES"`
}
