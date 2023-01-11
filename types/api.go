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
