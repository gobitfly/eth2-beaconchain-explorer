package types

import (
	"database/sql/driver"
	"encoding/json"
	"math/big"
	"time"

	"github.com/pkg/errors"
)

type ApiResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

type StatsSystem struct {
	CPUCores                      uint64 `mapstructure:"cpu_cores"`
	CPUThreads                    uint64 `mapstructure:"cpu_threads"`
	CPUNodeSystemSecondsTotal     uint64 `mapstructure:"cpu_node_system_seconds_total"`
	CPUNodeUserSecondsTotal       uint64 `mapstructure:"cpu_node_user_seconds_total"`
	CPUNodeIowaitSecondsTotal     uint64 `mapstructure:"cpu_node_iowait_seconds_total"`
	CPUNodeIdleSecondsTotal       uint64 `mapstructure:"cpu_node_idle_seconds_total"`
	MemoryNodeBytesTotal          uint64 `mapstructure:"memory_node_bytes_total"`
	MemoryNodeBytesFree           uint64 `mapstructure:"memory_node_bytes_free"`
	MemoryNodeBytesCached         uint64 `mapstructure:"memory_node_bytes_cached"`
	MemoryNodeBytesBuffers        uint64 `mapstructure:"memory_node_bytes_buffers"`
	DiskNodeBytesTotal            uint64 `mapstructure:"disk_node_bytes_total"`
	DiskNodeBytesFree             uint64 `mapstructure:"disk_node_bytes_free"`
	DiskNodeIoSeconds             uint64 `mapstructure:"disk_node_io_seconds"`
	DiskNodeReadsTotal            uint64 `mapstructure:"disk_node_reads_total"`
	DiskNodeWritesTotal           uint64 `mapstructure:"disk_node_writes_total"`
	NetworkNodeBytesTotalReceive  uint64 `mapstructure:"network_node_bytes_total_receive"`
	NetworkNodeBytesTotalTransmit uint64 `mapstructure:"network_node_bytes_total_transmit"`
	MiscNodeBootTsSeconds         uint64 `mapstructure:"misc_node_boot_ts_seconds"`
	MiscOS                        string `mapstructure:"misc_os"`
}

type StatsProcess struct {
	CPUProcessSecondsTotal     uint64 `mapstructure:"cpu_process_seconds_total"`
	MemoryProcessBytes         uint64 `mapstructure:"memory_process_bytes"`
	ClientName                 string `mapstructure:"client_name"`
	ClientVersion              string `mapstructure:"client_version"`
	ClientBuild                uint64 `mapstructure:"client_build"`
	SyncEth2FallbackConfigured bool   `mapstructure:"sync_eth2_fallback_configured"`
	SyncEth2FallbackConnected  bool   `mapstructure:"sync_eth2_fallback_connected"`
}

type StatsAdditionalsValidator struct {
	ValidatorTotal  uint64 `mapstructure:"validator_total"`
	ValidatorActive uint64 `mapstructure:"validator_active"`
}

type StatsAdditionalsBeaconnode struct {
	DiskBeaconchainBytesTotal       uint64 `mapstructure:"disk_beaconchain_bytes_total"`
	NetworkLibp2pBytesTotalReceive  uint64 `mapstructure:"network_libp2p_bytes_total_receive"`
	NetworkLibp2pBytesTotalTransmit uint64 `mapstructure:"network_libp2p_bytes_total_transmit"`
	NetworkPeersConnected           uint64 `mapstructure:"network_peers_connected"`
	SyncEth1Connected               bool   `mapstructure:"sync_eth1_connected"`
	SyncEth2Synced                  bool   `mapstructure:"sync_eth2_synced"`
	SyncBeaconHeadSlot              uint64 `mapstructure:"sync_beacon_head_slot"`
	SyncEth1FallbackConfigured      bool   `mapstructure:"sync_eth1_fallback_configured"`
	SyncEth1FallbackConnected       bool   `mapstructure:"sync_eth1_fallback_connected"`
}

type StatsMeta struct {
	Version         uint64 `mapstructure:"version"`
	Timestamp       uint64 `mapstructure:"timestamp"`
	Process         string `mapstructure:"process"`
	Machine         string
	ExporterVersion string `mapstructure:"exporter_version"`
}

type StatsDataStruct struct {
	Validator interface{} `json:"validator"`
	Node      interface{} `json:"node"`
	System    interface{} `json:"system"`
}

type WidgetResponse struct {
	Eff             any   `json:"efficiency"`
	Validator       any   `json:"validator"`
	Epoch           int64 `json:"epoch"`
	RocketpoolStats any   `json:"rocketpool_network_stats"`
}

type UsersNotificationsRequest struct {
	EventNames    []string `json:"event_names"`
	EventFilters  []string `json:"event_filters"`
	Search        string   `json:"search"`
	Limit         uint64   `json:"limit"`
	Offset        uint64   `json:"offset"`
	JoinValidator bool     `json:"join_validator"`
}

type DashboardRequest struct {
	IndicesOrPubKey string `json:"indicesOrPubkey"`
}

type DiscordEmbed struct {
	Color       string              `json:"color,omitempty"`
	Description string              `json:"description,omitempty"`
	Fields      []DiscordEmbedField `json:"fields,omitempty"`
	Title       string              `json:"title,omitempty"`
	Type        string              `json:"type,omitempty"`
}

type DiscordEmbedField struct {
	Inline bool   `json:"inline"`
	Name   string `json:"name"`
	Value  string `json:"value"`
}

type DiscordComponent struct {
	Type       uint64                   `json:"type"`
	Components []DiscordComponentButton `json:"components"`
}

type DiscordComponentButton struct {
	Style    uint64 `json:"style"`
	CustomID string `json:"custom_id"`
	Label    string `json:"label"`
	URL      string `json:"url"`
	Disabled bool   `json:"disabled"`
	Type     uint64 `json:"type"`
}

type DiscordReq struct {
	Content         string             `json:"content,omitempty"`
	Username        string             `json:"username,omitempty"`
	Avatar_url      string             `json:"avatar_url,omitempty"`
	Tts             bool               `json:"tts,omitempty"`
	Embeds          []DiscordEmbed     `json:"embeds,omitempty"`
	AllowedMentions []interface{}      `json:"allowedMentions,omitempty"`
	Components      []DiscordComponent `json:"components,omitempty"`
	Files           interface{}        `json:"files,omitempty"`
	Payload         string             `json:"payload,omitempty"`
	Attachments     interface{}        `json:"attachments,omitempty"`
	Flags           int                `json:"flags,omitempty"`
}

type ExecutionPerformanceResponse struct {
	Performance1d  *big.Int `json:"performance1d"`
	Performance7d  *big.Int `json:"performance7d"`
	Performance31d *big.Int `json:"performance31d"`
	ValidatorIndex uint64   `json:"validatorindex"`
}

type ExecutionBlockApiResponse struct {
	Hash               string                `json:"blockHash"`
	BlockNumber        uint64                `json:"blockNumber"`
	Timestamp          uint64                `json:"timestamp"`
	BlockReward        *big.Int              `json:"blockReward"`
	BlockMevReward     *big.Int              `json:"blockMevReward"`
	FeeRecipientReward *big.Int              `json:"producerReward"`
	FeeRecipient       string                `json:"feeRecipient"`
	GasLimit           uint64                `json:"gasLimit"`
	GasUsed            uint64                `json:"gasUsed"`
	BaseFee            *big.Int              `json:"baseFee"`
	TxCount            uint64                `json:"txCount"`
	InternalTxCount    uint64                `json:"internalTxCount"`
	UncleCount         uint64                `json:"uncleCount"`
	ParentHash         string                `json:"parentHash"`
	UncleHash          string                `json:"uncleHash"`
	Difficulty         *big.Int              `json:"difficulty"`
	PoSData            *ExecBlockProposer    `json:"posConsensus"`
	RelayData          *RelayDataApiResponse `json:"relay"`
	ConsensusAlgorithm string                `json:"consensusAlgorithm"`
}

type RelayDataApiResponse struct {
	TagID                string `json:"tag"`
	BuilderPubKey        string `json:"builderPubkey"`
	ProposerFeeRecipient string `json:"producerFeeRecipient"`
}

type AddressIndexOrPubkey struct {
	Address []byte
	Index   uint64
	Pubkey  []byte
}

type RelaysData struct {
	MevRecipient  []byte    `db:"proposer_fee_recipient"`
	MevBribe      WeiString `db:"value"`
	ExecBlockHash []byte    `db:"exec_block_hash"`
	TagID         string    `db:"tag_id"`
	BuilderPubKey []byte    `db:"builder_pubkey"`
}

type ExecBlockProposer struct {
	ExecBlock uint64 `db:"exec_block_number" json:"executionBlockNumber"`
	Proposer  uint64 `db:"proposer" json:"proposerIndex"`
	Slot      uint64 `db:"slot" json:"slot"`
	Epoch     uint64 `db:"epoch" json:"epoch"`
	Finalized bool   `json:"finalized"`
}

func (e *DiscordReq) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &e)
}

func (a DiscordReq) Value() (driver.Value, error) {
	return json.Marshal(a)
}

type ApiEth1AddressResponse struct {
	Address string `json:"address"`
	Ether   string `json:"ether"`
	Tokens  []struct {
		Address  string  `json:"address"`
		Balance  string  `json:"balance"`
		Symbol   string  `json:"symbol"`
		Decimals string  `json:"decimals,omitempty"`
		Price    float64 `json:"price,omitempty"`
		Currency string  `json:"currency,omitempty"`
	} `json:"tokens"`
}

type APIEth1AddressTxResponse struct {
	Transactions []Eth1TransactionParsed `json:"transactions"`
	Page         string                  `json:"page"`
}

type Eth1TransactionParsed struct {
	Hash               string    `json:"hash,omitempty"`
	BlockNumber        uint64    `json:"block,omitempty"`
	Time               time.Time `json:"time,omitempty"`
	MethodId           string    `json:"method,omitempty"`
	From               string    `json:"from,omitempty"`
	To                 string    `json:"to,omitempty"`
	Value              string    `json:"value,omitempty"`
	TxFee              string    `json:"fee,omitempty"`
	GasPrice           string    `json:"gasPrice,omitempty"`
	IsContractCreation bool      `json:"is_contract_creation,omitempty"`
	InvokesContract    bool      `json:"invokes_contract,omitempty"`
}

type APIEth1AddressItxResponse struct {
	InternalTransactions []Eth1InternalTransactionParsed `json:"internal_transactions"`
	Page                 string                          `json:"page"`
}

type Eth1InternalTransactionParsed struct {
	ParentHash  string    `json:"parent"`
	BlockNumber uint64    `json:"block"`
	Type        string    `json:"type"`
	Time        time.Time `json:"time"`
	From        string    `json:"from"`
	To          string    `json:"to"`
	Value       string    `json:"value"`
}

type APIEth1AddressBlockResponse struct {
	ProducedBlocks []Eth1BlockParsed `json:"blocks"`
	Page           string            `json:"page"`
}

type Eth1BlockParsed struct {
	Hash                     string    `json:"hash,omitempty"`
	ParentHash               string    `json:"parent_hash,omitempty"`
	UncleHash                string    `json:"uncle_hash,omitempty"`
	Coinbase                 string    `json:"coinbase,omitempty"`
	TxReward                 string    `json:"tx_reward,omitempty"`
	Difficulty               string    `json:"difficulty,omitempty"`
	Number                   uint64    `json:"number,omitempty"`
	GasLimit                 uint64    `json:"gas_limit,omitempty"`
	GasUsed                  uint64    `json:"gas_used,omitempty"`
	Time                     time.Time `json:"time,omitempty"`
	BaseFee                  string    `json:"base_fee,omitempty"`
	UncleCount               uint64    `json:"uncle_count,omitempty"`
	TransactionCount         uint64    `json:"transaction_count,omitempty"`
	InternalTransactionCount uint64    `json:"internal_transaction_count,omitempty"`
	Mev                      string    `json:"mev,omitempty"`
	LowestGasPrice           string    `json:"lowest_gas_price,omitempty"`
	HighestGasPrice          string    `json:"highest_gas_price,omitempty"`
	// Duration uint64 `json:"duration,omitempty"`
	UncleReward string `json:"uncle_reward,omitempty"`
	// BaseFeeChange string `json:"base_fee_change,omitempty"`
	// BlockUtilizationChange string `json:"block_utilization_change,omitempty"`
}

type APIEth1AddressUncleResponse struct {
	ProducedUncles []Eth1UncleParsed `json:"uncles"`
	Page           string            `json:"page"`
}

type Eth1UncleParsed struct {
	BlockNumber uint64    `json:"block,omitempty"`
	Number      uint64    `json:"number,omitempty"`
	GasLimit    uint64    `json:"gas_limit,omitempty"`
	GasUsed     uint64    `json:"gas_used,omitempty"`
	BaseFee     string    `json:"base_fee,omitempty"`
	Difficulty  string    `json:"difficulty,omitempty"`
	Time        time.Time `json:"time,omitempty"`
	Reward      string    `json:"reward,omitempty"`
}

type APIEth1TokenResponse struct {
	TokenTxs []*Eth1TokenTxParsed `json:"transactions"`
	Page     string               `json:"page"`
}

type Eth1TokenTxParsed struct {
	ParentHash   string    `json:"transaction,omitempty"`
	BlockNumber  uint64    `json:"block,omitempty"`
	TokenAddress string    `json:"token_address,omitempty"`
	Time         time.Time `json:"time,omitempty"`
	From         string    `json:"from,omitempty"`
	To           string    `json:"to,omitempty"`
	Value        string    `json:"value,omitempty"`
	TokenId      string    `json:"token_id,omitempty"`
	Operator     string    `json:"operator,omitempty"`
}
type ApiWithdrawalCredentialsResponse struct {
	Publickey      string `json:"publickey"`
	ValidatorIndex uint64 `json:"validatorindex"`
}

type APIEpochResponse struct {
	Epoch                   uint64 `json:"epoch"`
	Ts                      uint64 `json:"ts"`
	AttestationsCount       uint64 `json:"attestationscount"`
	AttesterSlashingsCount  uint64 `json:"attesterslashingscount"`
	AverageValidatorBalance uint64 `json:"averagevalidatorbalance"`
	BlocksCount             uint64 `json:"blockscount"`
	DepositsCount           uint64 `json:"depositscount"`
	EligibleEther           uint64 `json:"eligibleether"`
	Finalized               bool   `json:"finalized"`
	GlobalParticipationRate uint64 `json:"globalparticipationrate"`
	MissedBlocks            uint64 `json:"missedblocks"`
	OrphanedBlocks          uint64 `json:"orphanedblocks"`
	ProposedBlocks          uint64 `json:"proposedblocks"`
	ProposerSlashingsCount  uint64 `json:"proposerslashingscount"`
	ScheduledBlocks         uint64 `json:"scheduledblocks"`
	TotalValidatorBalance   uint64 `json:"totalvalidatorbalance"`
	ValidatorsCount         uint64 `json:"validatorscount"`
	VoluntaryExitsCount     uint64 `json:"voluntaryexitscount"`
	VotedEther              uint64 `json:"votedether"`
	RewardsExported         uint64 `json:"rewards_exported"`
	WithdrawalCount         uint64 `json:"withdrawalcount"`
}

type APISlotResponse struct {
	Attestationscount          uint64  `json:"attestationscount"`
	Attesterslashingscount     uint64  `json:"attesterslashingscount"`
	Blockroot                  string  `json:"blockroot"`
	Depositscount              uint64  `json:"depositscount"`
	Epoch                      uint64  `json:"epoch"`
	Eth1dataBlockhash          string  `json:"eth1data_blockhash"`
	Eth1dataDepositcount       uint64  `json:"eth1data_depositcount"`
	Eth1dataDepositroot        string  `json:"eth1data_depositroot"`
	ExecBaseFeePerGas          uint64  `json:"exec_base_fee_per_gas" extensions:"x-nullable"`
	ExecBlockHash              string  `json:"exec_block_hash" extensions:"x-nullable"`
	ExecBlockNumber            uint64  `json:"exec_block_number" extensions:"x-nullable"`
	ExecExtraData              string  `json:"exec_extra_data" extensions:"x-nullable"`
	ExecFeeRecipient           string  `json:"exec_fee_recipient" extensions:"x-nullable"`
	ExecGasLimit               uint64  `json:"exec_gas_limit" extensions:"x-nullable"`
	ExecGasUsed                uint64  `json:"exec_gas_used" extensions:"x-nullable"`
	ExecLogsBloom              string  `json:"exec_logs_bloom" extensions:"x-nullable"`
	ExecParentHash             string  `json:"exec_parent_hash" extensions:"x-nullable"`
	ExecRandom                 string  `json:"exec_random" extensions:"x-nullable"`
	ExecReceiptsRoot           string  `json:"exec_receipts_root" extensions:"x-nullable"`
	ExecStateRoot              string  `json:"exec_state_root" extensions:"x-nullable"`
	ExecTimestamp              uint64  `json:"exec_timestamp" extensions:"x-nullable"`
	ExecTransactionsCount      uint64  `json:"exec_transactions_count" extensions:"x-nullable"`
	Graffiti                   string  `json:"graffiti"`
	GraffitiText               string  `json:"graffiti_text"`
	Parentroot                 string  `json:"parentroot"`
	Proposer                   uint64  `json:"proposer"`
	Proposerslashingscount     uint64  `json:"proposerslashingscount"`
	Randaoreveal               string  `json:"randaoreveal"`
	Signature                  string  `json:"signature"`
	Slot                       uint64  `json:"slot"`
	Stateroot                  string  `json:"stateroot"`
	Status                     string  `json:"status"`
	SyncaggregateBits          string  `json:"syncaggregate_bits"`
	SyncaggregateParticipation float64 `json:"syncaggregate_participation"`
	SyncaggregateSignature     string  `json:"syncaggregate_signature"`
	Voluntaryexitscount        uint64  `json:"voluntaryexitscount"`
	WithdrawalCount            uint64  `json:"withdrawalcount"`
}

type APIAttestationResponse struct {
	Aggregationbits string  `json:"aggregationbits"`
	Beaconblockroot string  `json:"beaconblockroot"`
	BlockIndex      int64   `json:"block_index"`
	BlockRoot       string  `json:"block_root"`
	BlockSlot       int64   `json:"block_slot"`
	Committeeindex  int64   `json:"committeeindex"`
	Signature       string  `json:"signature"`
	Slot            int64   `json:"slot"`
	SourceEpoch     int64   `json:"source_epoch"`
	SourceRoot      string  `json:"source_root"`
	TargetEpoch     int64   `json:"target_epoch"`
	TargetRoot      string  `json:"target_root"`
	Validators      []int64 `json:"validators"`
}

type APIDepositResponse struct {
	Amount                uint64 `json:"amount"`
	BlockIndex            uint64 `json:"block_index"`
	BlockRoot             string `json:"block_root"`
	BlockSlot             uint64 `json:"block_slot"`
	Proof                 string `json:"proof"`
	Publickey             string `json:"publickey"`
	Signature             string `json:"signature"`
	Withdrawalcredentials string `json:"withdrawalcredentials"`
}

type APIAttesterSlashingResponse struct {
	Attestation1_beaconblockroot string   `json:"attestation1_beaconblockroot"`
	Attestation1_index           uint64   `json:"attestation1_index"`
	Attestation1_indices         []uint64 `json:"attestation1_indices"`
	Attestation1_signature       string   `json:"attestation1_signature"`
	Attestation1_slot            uint64   `json:"attestation1_slot"`
	Attestation1_source_epoch    uint64   `json:"attestation1_source_epoch"`
	Attestation1_source_root     string   `json:"attestation1_source_root"`
	Attestation1_target_epoch    uint64   `json:"attestation1_target_epoch"`
	Attestation1_target_root     string   `json:"attestation1_target_root"`
	Attestation2_beaconblockroot string   `json:"attestation2_beaconblockroot"`
	Attestation2_index           uint64   `json:"attestation2_index"`
	Attestation2_indices         []uint64 `json:"attestation2_indices"`
	Attestation2_signature       string   `json:"attestation2_signature"`
	Attestation2_slot            uint64   `json:"attestation2_slot"`
	Attestation2_source_epoch    uint64   `json:"attestation2_source_epoch"`
	Attestation2_source_root     string   `json:"attestation2_source_root"`
	Attestation2_target_epoch    uint64   `json:"attestation2_target_epoch"`
	Attestation2_target_root     string   `json:"attestation2_target_root"`
	BlockIndex                   uint64   `json:"block_index"`
	BlockRoot                    string   `json:"block_root"`
	BlockSlot                    uint64   `json:"block_slot"`
}

type APIProposerSlashingResponse struct {
	BlockIndex        uint64 `json:"block_index"`
	BlockRoot         string `json:"block_root"`
	BlockSlot         uint64 `json:"block_slot"`
	Header1Bodyroot   string `json:"header1_bodyroot"`
	Header1Parentroot string `json:"header1_parentroot"`
	Header1Signature  string `json:"header1_signature"`
	Header1Slot       uint64 `json:"header1_slot"`
	Header1Stateroot  string `json:"header1_stateroot"`
	Header2Bodyroot   string `json:"header2_bodyroot"`
	Header2Parentroot string `json:"header2_parentroot"`
	Header2Signature  string `json:"header2_signature"`
	Header2Slot       uint64 `json:"header2_slot"`
	Header2Stateroot  string `json:"header2_stateroot"`
	ProposerIndex     uint64 `json:"proposerindex"`
}

type APIVoluntaryExitResponse struct {
	BlockIndex     uint64 `json:"block_index"`
	BlockRoot      string `json:"block_root"`
	BlockSlot      uint64 `json:"block_slot"`
	Epoch          uint64 `json:"epoch"`
	Signature      string `json:"signature"`
	ValidatorIndex uint64 `json:"validatorindex"`
}

type APISyncCommitteeResponse struct {
	EndEpoch   uint64   `json:"end_epoch"`
	Period     uint64   `json:"period"`
	StartEpoch uint64   `json:"start_epoch"`
	Validators []uint64 `json:"validators"`
}

type APIRocketpoolStatsResponse struct {
	ClaimIntervalTime      string  `json:"claim_interval_time"`
	ClaimIntervalTimeStart int64   `json:"claim_interval_time_start"`
	CurrentNodeDemand      float64 `json:"current_node_demand"`
	CurrentNodeFee         float64 `json:"current_node_fee"`
	EffectiveRPLStaked     float64 `json:"effective_rpl_staked"`
	MinipoolCount          int64   `json:"minipool_count"`
	NodeCount              int64   `json:"node_count"`
	NodeOperatorRewards    float64 `json:"node_operator_rewards"`
	OdaoMemberCount        int64   `json:"odao_member_count"`
	RethApr                float64 `json:"reth_apr"`
	RethExchangeRate       float64 `json:"reth_exchange_rate"`
	RethSupply             float64 `json:"reth_supply"`
	RplPrice               int64   `json:"rpl_price"`
	TotalEthBalance        float64 `json:"total_eth_balance"`
	TotalEthStaking        float64 `json:"total_eth_staking"`
}

type ApiRocketpoolValidatorResponse struct {
	ClaimedSmoothingPool   float64 `json:"claimed_smoothing_pool"`
	Index                  uint64  `json:"index"`
	MinipoolAddress        string  `json:"minipool_address"`
	MinipoolDepositType    string  `json:"minipool_deposit_type"`
	MinipoolNodeFee        float64 `json:"minipool_node_fee"`
	MinipoolStatus         string  `json:"minipool_status"`
	MinipoolStatusTime     uint64  `json:"minipool_status_time"`
	NodeAddress            string  `json:"node_address"`
	NodeMaxRplStake        float64 `json:"node_max_rpl_stake"`
	NodeMinRplStake        float64 `json:"node_min_rpl_stake"`
	NodeRplStake           float64 `json:"node_rpl_stake"`
	NodeTimezoneLocation   string  `json:"node_timezone_location"`
	PenaltyCount           uint64  `json:"penalty_count"`
	RplCumulativeRewards   float64 `json:"rpl_cumulative_rewards"`
	SmoothingPoolOptedIn   bool    `json:"smoothing_pool_opted_in"`
	UnclaimedRplRewards    float64 `json:"unclaimed_rpl_rewards"`
	UnclaimedSmoothingPool float64 `json:"unclaimed_smoothing_pool"`
}

type ApiValidatorQueueResponse struct {
	BeaconchainEntering uint64 `json:"beaconchain_entering"`
	BeaconchainExiting  uint64 `json:"beaconchain_exiting"`
	ValidatorsCount     uint64 `json:"validators_count"`
}

type APIValidatorResponse struct {
	ActivationEligibilityEpoch uint64 `json:"activation_eligibility_epoch"`
	ActivationEpoch            uint64 `json:"activation_epoch"`
	Balance                    uint64 `json:"balance"`
	EffectiveBalance           uint64 `json:"effective_balance"`
	ExitEpoch                  uint64 `json:"exit_epoch"`
	LastAttestationSlot        uint64 `json:"last_attestation_slot"`
	Name                       string `json:"name"`
	Pubkey                     string `json:"pubkey"`
	Slashed                    bool   `json:"slashed"`
	Status                     string `json:"status"`
	ValidatorIndex             uint64 `json:"validator_index"`
	WithdrawableEpoch          uint64 `json:"withdrawable_epoch"`
	WithdrawalCredentials      string `json:"withdrawal_credentials"`
}

type ApiValidatorDailyStatsResponse struct {
	ValidatorIndex        uint64    `json:"validatorindex"`
	AttesterSlashings     uint64    `json:"attester_slashings"`
	Day                   uint64    `json:"day"`
	DayStart              time.Time `json:"day_start"`
	DayEnd                time.Time `json:"day_end"`
	Deposits              uint64    `json:"deposits"`
	DepositsAmount        uint64    `json:"deposits_amount"`
	Withdrawals           uint64    `json:"withdrawals"`
	WithdrawalsAmount     uint64    `json:"withdrawals_amount"`
	EndBalance            uint64    `json:"end_balance"`
	EndEffectiveBalance   uint64    `json:"end_effective_balance"`
	MaxBalance            uint64    `json:"max_balance"`
	MaxEffectiveBalance   uint64    `json:"max_effective_balance"`
	MinBalance            uint64    `json:"min_balance"`
	MinEffectiveBalance   uint64    `json:"min_effective_balance"`
	MissedAttestations    uint64    `json:"missed_attestations"`
	MissedBlocks          uint64    `json:"missed_blocks"`
	MissedSync            uint64    `json:"missed_sync"`
	OrphanedAttestations  uint64    `json:"orphaned_attestations"`
	OrphanedBlocks        uint64    `json:"orphaned_blocks"`
	OrphanedSync          uint64    `json:"orphaned_sync"`
	ParticipatedSync      uint64    `json:"participated_sync"`
	ProposedBlocks        uint64    `json:"proposed_blocks"`
	ProposerSlashings     uint64    `json:"proposer_slashings"`
	StartBalance          uint64    `json:"start_balance"`
	StartEffectiveBalance uint64    `json:"start_effective_balance"`
}

type ApiValidatorEth1Response struct {
	PublicKey      string `json:"public_key"`
	ValidSignature bool   `json:"valid_signature"`
	ValidatorIndex uint64 `json:"validator_index"`
}

type ApiValidatorIncomeHistoryResponse struct {
	Income struct {
		AttestationSourceReward uint64 `json:"attestation_source_reward"`
		AttestationTargetReward uint64 `json:"attestation_target_reward"`
		AttestationHeadReward   uint64 `json:"attestation_head_reward"`
	} `json:"income"`
	Epoch          uint64    `json:"epoch"`
	ValidatorIndex uint64    `json:"validatorindex"`
	Week           uint64    `json:"week"`
	WeekStart      time.Time `json:"week_start"`
	WeekEnd        time.Time `json:"week_end"`
}

type ApiValidatorBalanceHistoryResponse struct {
	Balance          uint64    `json:"balance"`
	EffectiveBalance uint64    `json:"effectivebalance"`
	Epoch            uint64    `json:"epoch"`
	Validatorindex   uint64    `json:"validatorindex"`
	Week             uint64    `json:"week"`
	WeekStart        time.Time `json:"week_start"`
	WeekEnd          time.Time `json:"week_end"`
}

type ApiValidatorWithdrawalResponse struct {
	Epoch          uint64 `json:"epoch,omitempty"`
	Slot           uint64 `json:"slot,omitempty"`
	BlockRoot      string `json:"blockroot,omitempty"`
	Index          uint64 `json:"withdrawalindex"`
	ValidatorIndex uint64 `json:"validatorindex"`
	Address        string `json:"address"`
	Amount         uint64 `json:"amount"`
}

type ApiValidatorBlsChangeResponse struct {
	Epoch                    uint64 `db:"epoch" json:"epoch,omitempty"`
	Slot                     uint64 `db:"slot" json:"slot,omitempty"`
	BlockRoot                string `db:"block_rot" json:"blockroot,omitempty"`
	Validatorindex           uint64 `db:"validatorindex" json:"validatorindex,omitempty"`
	BlsPubkey                string `db:"pubkey" json:"bls_pubkey,omitempty"`
	Signature                string `db:"signature" json:"bls_signature,omitempty"`
	Address                  string `db:"address" json:"address,omitempty"`
	WithdrawalCredentialsOld string `db:"withdrawalcredentials_0x00" json:"withdrawalcredentials_0x00,omitempty"`
	WithdrawalCredentialsNew string `db:"withdrawalcredentials_0x01" json:"withdrawalcredentials_0x01,omitempty"`
}

type ApiValidatorPerformanceResponse struct {
	Balance         uint64 `json:"balance"`
	Performance1d   uint64 `json:"performance1d"`
	Performance31d  uint64 `json:"performance31d"`
	Performance365d uint64 `json:"performance365d"`
	Performance7d   uint64 `json:"performance7d"`
	Rank7d          uint64 `json:"rank7d"`
	Validatorindex  uint64 `json:"validatorindex"`
}

type ApiValidatorExecutionPerformanceResponse struct {
	Performance1d  uint64 `json:"performance1d"`
	Performance7d  uint64 `json:"performance7d"`
	Performance31d uint64 `json:"performance31d"`
	Validatorindex uint64 `json:"validatorindex"`
}

type ApiValidatorDepositsResponse struct {
	Amount                uint64 `json:"amount"`
	BlockNumber           uint64 `json:"block_number"`
	BlockTs               uint64 `json:"block_ts"`
	FromAddress           string `json:"from_address"`
	MerkletreeIndex       string `json:"merkletree_index"`
	Publickey             string `json:"publickey"`
	Removed               bool   `json:"removed"`
	Signature             string `json:"signature"`
	TxHash                string `json:"tx_hash"`
	TxIndex               uint64 `json:"tx_index"`
	TxInput               string `json:"tx_input"`
	ValidSignature        bool   `json:"valid_signature"`
	WithdrawalCredentials string `json:"withdrawal_credentials"`
}

type ApiValidatorAttestationsResponse struct {
	AttesterSlot   uint64    `json:"attesterslot"`
	CommitteeIndex uint64    `json:"committeeindex"`
	Epoch          uint64    `json:"epoch"`
	InclusionSlot  uint64    `json:"inclusionslot"`
	Status         uint64    `json:"status"`
	ValidatorIndex uint64    `json:"validatorindex"`
	Week           uint64    `json:"week"`
	WeekStart      time.Time `json:"week_start"`
	WeekEnd        time.Time `json:"week_end"`
}

// convert this json object to a golang struct called ApiValidatorProposalsResponse
type ApiValidatorProposalsResponse struct {
	Attestationscount          uint64  `db:"attestationscount" json:"attestationscount"`
	Attesterslashingscount     uint64  `db:"attesterslashingscount" json:"attesterslashingscount"`
	Blockroot                  string  `db:"blockroot" json:"blockroot"`
	Depositscount              uint64  `db:"depositscount" json:"depositscount"`
	Epoch                      uint64  `db:"epoch" json:"epoch"`
	Eth1dataBlockhash          string  `db:"eth1data_blockhash" json:"eth1data_blockhash"`
	Eth1dataDepositcount       uint64  `db:"eth1data_depositcount" json:"eth1data_depositcount"`
	Eth1dataDepositroot        string  `db:"eth1data_depositroot" json:"eth1data_depositroot"`
	ExecBaseFeePerGas          *uint64 `db:"exec_base_fee_per_gas" json:"exec_base_fee_per_gas,omitempty"`
	ExecBlockHash              *string `db:"exec_block_hash" json:"exec_block_hash,omitempty"`
	ExecBlockNumber            *uint64 `db:"exec_block_number" json:"exec_block_number,omitempty"`
	ExecExtra_data             *string `db:"exec_extra_data" json:"exec_extra_data,omitempty"`
	ExecFeeRecipient           *string `db:"exec_fee_recipient" json:"exec_fee_recipient,omitempty"`
	ExecGasLimit               *uint64 `db:"exec_gas_limit" json:"exec_gas_limit,omitempty"`
	ExecGasUsed                *uint64 `db:"exec_gas_used" json:"exec_gas_used,omitempty"`
	ExecLogsBloom              *string `db:"exec_logs_bloom" json:"exec_logs_bloom,omitempty"`
	ExecParentHash             *string `db:"exec_parent_hash" json:"exec_parent_hash,omitempty"`
	ExecRandom                 *string `db:"exec_random" json:"exec_random,omitempty"`
	ExecReceiptsRoot           *string `db:"exec_receipts_root" json:"exec_receipts_root,omitempty"`
	ExecStateRoot              *string `db:"exec_state_root" json:"exec_state_root,omitempty"`
	ExecTimestamp              *uint64 `db:"exec_timestamp" json:"exec_timestamp,omitempty"`
	ExecTransactionsCount      *uint64 `db:"exec_transactions_count" json:"exec_transactions_count,omitempty"`
	Graffiti                   string  `db:"graffiti" json:"graffiti"`
	GraffitiText               string  `db:"graffiti_text" json:"graffiti_text"`
	Parentroot                 string  `db:"parentroot" json:"parentroot"`
	Proposer                   uint64  `db:"proposer" json:"proposer"`
	Proposerslashingscount     uint64  `db:"proposerslashingscount" json:"proposerslashingscount"`
	Randaoreveal               string  `db:"randaoreveal" json:"randaoreveal"`
	Signature                  string  `db:"signature" json:"signature"`
	Slot                       uint64  `db:"slot" json:"slot"`
	Stateroot                  string  `db:"stateroot" json:"stateroot"`
	Status                     string  `db:"status" json:"status"`
	SyncaggregateBits          string  `db:"syncaggregate_bits" json:"syncaggregate_bits"`
	SyncaggregateParticipation float64 `db:"syncaggregate_participation" json:"syncaggregate_participation"`
	SyncaggregateSignature     string  `db:"syncaggregate_signature" json:"syncaggregate_signature"`
	Voluntaryexitscount        uint64  `db:"voluntaryexitscount" json:"voluntaryexitscount"`
}
