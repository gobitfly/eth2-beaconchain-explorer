package types

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// ChainHead is a struct to hold chain head data
type ChainHead struct {
	HeadSlot                   uint64
	HeadEpoch                  uint64
	HeadBlockRoot              []byte
	FinalizedSlot              uint64
	FinalizedEpoch             uint64
	FinalizedBlockRoot         []byte
	JustifiedSlot              uint64
	JustifiedEpoch             uint64
	JustifiedBlockRoot         []byte
	PreviousJustifiedSlot      uint64
	PreviousJustifiedEpoch     uint64
	PreviousJustifiedBlockRoot []byte
}

type FinalityCheckpoints struct {
	PreviousJustified struct {
		Epoch uint64 `json:"epoch"`
		Root  string `json:"root"`
	} `json:"previous_justified"`
	CurrentJustified struct {
		Epoch uint64 `json:"epoch"`
		Root  string `json:"root"`
	} `json:"current_justified"`
	Finalized struct {
		Epoch uint64 `json:"epoch"`
		Root  string `json:"root"`
	} `json:"finalized"`
}

type Slot uint64
type Epoch uint64
type ValidatorIndex uint64
type SyncCommitteePeriod uint64
type CommitteeIndex uint64

// EpochData is a struct to hold epoch data
type EpochData struct {
	Epoch                   uint64
	Validators              []*Validator
	ValidatorAssignmentes   *EpochAssignments
	Blocks                  map[uint64]map[string]*Block
	FutureBlocks            map[uint64]map[string]*Block
	EpochParticipationStats *ValidatorParticipation
	AttestationDuties       map[Slot]map[ValidatorIndex][]Slot
	SyncDuties              map[Slot]map[ValidatorIndex]bool
	Finalized               bool
}

// ValidatorParticipation is a struct to hold validator participation data
type ValidatorParticipation struct {
	Epoch                   uint64
	GlobalParticipationRate float32
	VotedEther              uint64
	EligibleEther           uint64
	Finalized               bool
}

// Validator is a struct to hold validator data
type Validator struct {
	Index                      uint64 `db:"validatorindex"`
	PublicKey                  []byte `db:"pubkey"`
	PublicKeyHex               string `db:"pubkeyhex"`
	Balance                    uint64 `db:"balance"`
	EffectiveBalance           uint64 `db:"effectivebalance"`
	Slashed                    bool   `db:"slashed"`
	ActivationEligibilityEpoch uint64 `db:"activationeligibilityepoch"`
	ActivationEpoch            uint64 `db:"activationepoch"`
	ExitEpoch                  uint64 `db:"exitepoch"`
	WithdrawableEpoch          uint64 `db:"withdrawableepoch"`
	WithdrawalCredentials      []byte `db:"withdrawalcredentials"`

	BalanceActivation sql.NullInt64 `db:"balanceactivation"`
	Status            string        `db:"status"`

	LastAttestationSlot sql.NullInt64 `db:"lastattestationslot"`
}

// ValidatorQueue is a struct to hold validator queue data
type ValidatorQueue struct {
	Activating uint64
	Exiting    uint64
}

type SyncAggregate struct {
	SyncCommitteeValidators    []uint64
	SyncCommitteeBits          []byte
	SyncCommitteeSignature     []byte
	SyncAggregateParticipation float64
}

// Block is a struct to hold block data
type Block struct {
	Status                     uint64
	Proposer                   uint64
	BlockRoot                  []byte
	Slot                       uint64
	ParentRoot                 []byte
	StateRoot                  []byte
	Signature                  []byte
	RandaoReveal               []byte
	Graffiti                   []byte
	Eth1Data                   *Eth1Data
	BodyRoot                   []byte
	ProposerSlashings          []*ProposerSlashing
	AttesterSlashings          []*AttesterSlashing
	Attestations               []*Attestation
	Deposits                   []*Deposit
	VoluntaryExits             []*VoluntaryExit
	SyncAggregate              *SyncAggregate    // warning: sync aggregate may be nil, for phase0 blocks
	ExecutionPayload           *ExecutionPayload // warning: payload may be nil, for phase0/altair blocks
	SignedBLSToExecutionChange []*SignedBLSToExecutionChange
	BlobGasUsed                uint64
	ExcessBlobGas              uint64
	BlobKZGCommitments         [][]byte
	BlobKZGProofs              [][]byte
	AttestationDuties          map[ValidatorIndex][]Slot
	SyncDuties                 map[ValidatorIndex]bool
	Finalized                  bool
	EpochAssignments           *EpochAssignments
	Validators                 []*Validator
}

type SignedBLSToExecutionChange struct {
	Message   BLSToExecutionChange
	Signature []byte
}

type BLSToExecutionChange struct {
	Validatorindex uint64
	BlsPubkey      []byte
	Address        []byte
}

type Transaction struct {
	Raw []byte
	// Note: below values may be nil/0 if Raw fails to decode into a valid transaction
	TxHash       []byte
	AccountNonce uint64
	// big endian
	Price     []byte
	GasLimit  uint64
	Sender    []byte
	Recipient []byte
	// big endian
	Amount  []byte
	Payload []byte

	MaxPriorityFeePerGas uint64
	MaxFeePerGas         uint64

	MaxFeePerBlobGas    uint64
	BlobVersionedHashes [][]byte
}

type ExecutionPayload struct {
	ParentHash    []byte
	FeeRecipient  []byte
	StateRoot     []byte
	ReceiptsRoot  []byte
	LogsBloom     []byte
	Random        []byte
	BlockNumber   uint64
	GasLimit      uint64
	GasUsed       uint64
	Timestamp     uint64
	ExtraData     []byte
	BaseFeePerGas uint64
	BlockHash     []byte
	Transactions  []*Transaction
	Withdrawals   []*Withdrawals
	BlobGasUsed   uint64
	ExcessBlobGas uint64
}

type Withdrawals struct {
	Slot           uint64 `json:"slot,omitempty"`
	BlockRoot      []byte `json:"blockroot,omitempty"`
	Index          uint64 `json:"index"`
	ValidatorIndex uint64 `json:"validatorindex"`
	Address        []byte `json:"address"`
	Amount         uint64 `json:"amount"`
}

type WithdrawalsByEpoch struct {
	Epoch          uint64
	ValidatorIndex uint64
	Amount         uint64
}

type WithdrawalsNotification struct {
	Slot           uint64 `json:"slot,omitempty"`
	Index          uint64 `json:"index"`
	ValidatorIndex uint64 `json:"validatorindex"`
	Address        []byte `json:"address"`
	Amount         uint64 `json:"amount"`
	Pubkey         []byte `json:"pubkey"`
}

// Eth1Data is a struct to hold the ETH1 data
type Eth1Data struct {
	DepositRoot  []byte
	DepositCount uint64
	BlockHash    []byte
}

// ProposerSlashing is a struct to hold proposer slashing data
type ProposerSlashing struct {
	ProposerIndex uint64
	Header1       *Block
	Header2       *Block
}

// AttesterSlashing is a struct to hold attester slashing
type AttesterSlashing struct {
	Attestation1 *IndexedAttestation
	Attestation2 *IndexedAttestation
}

// IndexedAttestation is a struct to hold indexed attestation data
type IndexedAttestation struct {
	Data             *AttestationData
	AttestingIndices []uint64
	Signature        []byte
}

// Attestation is a struct to hold attestation header data
type Attestation struct {
	AggregationBits []byte
	Attesters       []uint64
	Data            *AttestationData
	Signature       []byte
}

// AttestationData to hold attestation detail data
type AttestationData struct {
	Slot            uint64
	CommitteeIndex  uint64
	BeaconBlockRoot []byte
	Source          *Checkpoint
	Target          *Checkpoint
}

// Checkpoint is a struct to hold checkpoint data
type Checkpoint struct {
	Epoch uint64
	Root  []byte
}

// Deposit is a struct to hold deposit data
type Deposit struct {
	Proof                 [][]byte
	PublicKey             []byte
	WithdrawalCredentials []byte
	Amount                uint64
	Signature             []byte
}

// VoluntaryExit is a struct to hold voluntary exit data
type VoluntaryExit struct {
	Epoch          uint64
	ValidatorIndex uint64
	Signature      []byte
}

// MinimalBlock is a struct to hold minimal block data
type MinimalBlock struct {
	Epoch      uint64 `db:"epoch"`
	Slot       uint64 `db:"slot"`
	BlockRoot  []byte `db:"blockroot"`
	ParentRoot []byte `db:"parentroot"`
	Canonical  bool   `db:"-"`
}

// CanonBlock is a struct to hold canon block data
type CanonBlock struct {
	BlockRoot []byte `db:"blockroot"`
	Slot      uint64 `db:"slot"`
	Canonical bool   `db:"-"`
}

// EpochAssignments is a struct to hold epoch assignment data
type EpochAssignments struct {
	ProposerAssignments map[uint64]uint64
	AttestorAssignments map[string]uint64
	SyncAssignments     []uint64
}

// Eth1Deposit is a struct to hold eth1-deposit data
type Eth1Deposit struct {
	TxHash                []byte `db:"tx_hash"`
	TxInput               []byte `db:"tx_input"`
	TxIndex               uint64 `db:"tx_index"`
	BlockNumber           uint64 `db:"block_number"`
	BlockTs               int64  `db:"block_ts"`
	FromAddress           []byte `db:"from_address"`
	FromName              string
	PublicKey             []byte `db:"publickey"`
	WithdrawalCredentials []byte `db:"withdrawal_credentials"`
	Amount                uint64 `db:"amount"`
	Signature             []byte `db:"signature"`
	MerkletreeIndex       []byte `db:"merkletree_index"`
	Removed               bool   `db:"removed"`
	ValidSignature        bool   `db:"valid_signature"`
}

// Eth2Deposit is a struct to hold eth2-deposit data
type Eth2Deposit struct {
	BlockSlot             uint64 `db:"block_slot"`
	BlockIndex            uint64 `db:"block_index"`
	BlockRoot             []byte `db:"block_root"`
	Proof                 []byte `db:"proof"`
	Publickey             []byte `db:"publickey"`
	Withdrawalcredentials []byte `db:"withdrawalcredentials"`
	Amount                uint64 `db:"amount"`
	Signature             []byte `db:"signature"`
}

// EthStoreDay is a struct to hold performance data for a specific beaconchain-day.
// All fields use Gwei unless specified otherwise by the field name
type EthStoreDay struct {
	Pool                   string          `db:"pool"`
	Day                    uint64          `db:"day"`
	EffectiveBalancesSum   decimal.Decimal `db:"effective_balances_sum_wei"`
	StartBalancesSum       decimal.Decimal `db:"start_balances_sum_wei"`
	EndBalancesSum         decimal.Decimal `db:"end_balances_sum_wei"`
	DepositsSum            decimal.Decimal `db:"deposits_sum_wei"`
	TxFeesSumWei           decimal.Decimal `db:"tx_fees_sum_wei"`
	ConsensusRewardsSumWei decimal.Decimal `db:"consensus_rewards_sum_wei"`
	TotalRewardsWei        decimal.Decimal `db:"total_rewards_wei"`
	APR                    decimal.Decimal `db:"apr"`
}

type HistoricEthPrice struct {
	MarketData struct {
		CurrentPrice struct {
			Aed float64 `json:"aed"`
			Ars float64 `json:"ars"`
			Aud float64 `json:"aud"`
			Bdt float64 `json:"bdt"`
			Bhd float64 `json:"bhd"`
			Bmd float64 `json:"bmd"`
			Brl float64 `json:"brl"`
			Btc float64 `json:"btc"`
			Cad float64 `json:"cad"`
			Chf float64 `json:"chf"`
			Clp float64 `json:"clp"`
			Cny float64 `json:"cny"`
			Czk float64 `json:"czk"`
			Dkk float64 `json:"dkk"`
			Eth float64 `json:"eth"`
			Eur float64 `json:"eur"`
			Gbp float64 `json:"gbp"`
			Hkd float64 `json:"hkd"`
			Huf float64 `json:"huf"`
			Idr float64 `json:"idr"`
			Ils float64 `json:"ils"`
			Inr float64 `json:"inr"`
			Jpy float64 `json:"jpy"`
			Krw float64 `json:"krw"`
			Kwd float64 `json:"kwd"`
			Lkr float64 `json:"lkr"`
			Ltc float64 `json:"ltc"`
			Mmk float64 `json:"mmk"`
			Mxn float64 `json:"mxn"`
			Myr float64 `json:"myr"`
			Ngn float64 `json:"ngn"`
			Nok float64 `json:"nok"`
			Nzd float64 `json:"nzd"`
			Php float64 `json:"php"`
			Pkr float64 `json:"pkr"`
			Pln float64 `json:"pln"`
			Rub float64 `json:"rub"`
			Sar float64 `json:"sar"`
			Sek float64 `json:"sek"`
			Sgd float64 `json:"sgd"`
			Thb float64 `json:"thb"`
			Try float64 `json:"try"`
			Twd float64 `json:"twd"`
			Uah float64 `json:"uah"`
			Usd float64 `json:"usd"`
			Vef float64 `json:"vef"`
			Vnd float64 `json:"vnd"`
			Xag float64 `json:"xag"`
			Xau float64 `json:"xau"`
			Xdr float64 `json:"xdr"`
			Zar float64 `json:"zar"`
		} `json:"current_price"`
		MarketCap struct {
			Aed float64 `json:"aed"`
			Ars float64 `json:"ars"`
			Aud float64 `json:"aud"`
			Bdt float64 `json:"bdt"`
			Bhd float64 `json:"bhd"`
			Bmd float64 `json:"bmd"`
			Brl float64 `json:"brl"`
			Btc float64 `json:"btc"`
			Cad float64 `json:"cad"`
			Chf float64 `json:"chf"`
			Clp float64 `json:"clp"`
			Cny float64 `json:"cny"`
			Czk float64 `json:"czk"`
			Dkk float64 `json:"dkk"`
			Eth float64 `json:"eth"`
			Eur float64 `json:"eur"`
			Gbp float64 `json:"gbp"`
			Hkd float64 `json:"hkd"`
			Huf float64 `json:"huf"`
			Idr float64 `json:"idr"`
			Ils float64 `json:"ils"`
			Inr float64 `json:"inr"`
			Jpy float64 `json:"jpy"`
			Krw float64 `json:"krw"`
			Kwd float64 `json:"kwd"`
			Lkr float64 `json:"lkr"`
			Ltc float64 `json:"ltc"`
			Mmk float64 `json:"mmk"`
			Mxn float64 `json:"mxn"`
			Myr float64 `json:"myr"`
			Ngn float64 `json:"ngn"`
			Nok float64 `json:"nok"`
			Nzd float64 `json:"nzd"`
			Php float64 `json:"php"`
			Pkr float64 `json:"pkr"`
			Pln float64 `json:"pln"`
			Rub float64 `json:"rub"`
			Sar float64 `json:"sar"`
			Sek float64 `json:"sek"`
			Sgd float64 `json:"sgd"`
			Thb float64 `json:"thb"`
			Try float64 `json:"try"`
			Twd float64 `json:"twd"`
			Uah float64 `json:"uah"`
			Usd float64 `json:"usd"`
			Vef float64 `json:"vef"`
			Vnd float64 `json:"vnd"`
			Xag float64 `json:"xag"`
			Xau float64 `json:"xau"`
			Xdr float64 `json:"xdr"`
			Zar float64 `json:"zar"`
		} `json:"market_cap"`
		TotalVolume struct {
			Aed float64 `json:"aed"`
			Ars float64 `json:"ars"`
			Aud float64 `json:"aud"`
			Bdt float64 `json:"bdt"`
			Bhd float64 `json:"bhd"`
			Bmd float64 `json:"bmd"`
			Brl float64 `json:"brl"`
			Btc float64 `json:"btc"`
			Cad float64 `json:"cad"`
			Chf float64 `json:"chf"`
			Clp float64 `json:"clp"`
			Cny float64 `json:"cny"`
			Czk float64 `json:"czk"`
			Dkk float64 `json:"dkk"`
			Eth float64 `json:"eth"`
			Eur float64 `json:"eur"`
			Gbp float64 `json:"gbp"`
			Hkd float64 `json:"hkd"`
			Huf float64 `json:"huf"`
			Idr float64 `json:"idr"`
			Ils float64 `json:"ils"`
			Inr float64 `json:"inr"`
			Jpy float64 `json:"jpy"`
			Krw float64 `json:"krw"`
			Kwd float64 `json:"kwd"`
			Lkr float64 `json:"lkr"`
			Ltc float64 `json:"ltc"`
			Mmk float64 `json:"mmk"`
			Mxn float64 `json:"mxn"`
			Myr float64 `json:"myr"`
			Ngn float64 `json:"ngn"`
			Nok float64 `json:"nok"`
			Nzd float64 `json:"nzd"`
			Php float64 `json:"php"`
			Pkr float64 `json:"pkr"`
			Pln float64 `json:"pln"`
			Rub float64 `json:"rub"`
			Sar float64 `json:"sar"`
			Sek float64 `json:"sek"`
			Sgd float64 `json:"sgd"`
			Thb float64 `json:"thb"`
			Try float64 `json:"try"`
			Twd float64 `json:"twd"`
			Uah float64 `json:"uah"`
			Usd float64 `json:"usd"`
			Vef float64 `json:"vef"`
			Vnd float64 `json:"vnd"`
			Xag float64 `json:"xag"`
			Xau float64 `json:"xau"`
			Xdr float64 `json:"xdr"`
			Zar float64 `json:"zar"`
		} `json:"total_volume"`
	} `json:"market_data"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}

type Relay struct {
	ID                  string         `db:"tag_id"`
	Endpoint            string         `db:"endpoint"`
	Link                sql.NullString `db:"public_link"`
	IsCensoring         sql.NullBool   `db:"is_censoring"`
	IsEthical           sql.NullBool   `db:"is_ethical"`
	ExportFailureCount  uint64         `db:"export_failure_count"`
	LastExportTryTs     time.Time      `db:"last_export_try_ts"`
	LastExportSuccessTs time.Time      `db:"last_export_success_ts"`
	Logger              logrus.Entry
}

type RelayBlock struct {
	ID                   string `db:"tag_id" json:"tag_id"`
	BlockSlot            uint64 `db:"block_slot" json:"block_slot"`
	BlockRoot            string `db:"block_root" json:"block_root"`
	ExecBlockHash        string `db:"exec_block_hash"`
	Value                uint64 `db:"value" json:"value"`
	BuilderPubkey        string `db:"builder_pubkey" json:"builder_pubkey"`
	ProposerPubkey       string `db:"proposer_pubkey" json:"proposer_pubkey"`
	ProposerFeeRecipient string `db:"proposer_fee_recipient" json:"proposer_fee_recipient"`
}

type BlockTag struct {
	ID        string `db:"tag_id"`
	BlockSlot uint64 `db:"slot"`
	BlockRoot string `db:"blockroot"`
}

type TagMetadata struct {
	Name        string `json:"name"`
	Summary     string `json:"summary"`
	PublicLink  string `json:"public_url"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

type TagMetadataSlice []TagMetadata

func (s *TagMetadataSlice) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		err := json.Unmarshal(v, s)
		if err != nil {
			return err
		}
		// if no tags were found we will get back an empty TagMetadata, we don't want that
		if len(*s) == 1 && (*s)[0].Name == "" {
			*s = nil
		}
		return nil
	case string:
		return json.Unmarshal([]byte(v), s)
	}
	return errors.New("type assertion failed")
}

type WeiString struct {
	pgtype.Numeric
}

func (b WeiString) MarshalJSON() ([]byte, error) {
	return []byte("\"" + b.BigInt().String() + "\""), nil
}

func (b *WeiString) UnmarshalJSON(p []byte) error {
	if string(p) == "null" {
		return nil
	}
	if p[0] == byte('"') {
		p = p[1 : len(p)-1]
	}
	err := b.Set(string(p))
	if err != nil {
		return fmt.Errorf("failed to unmarshal WeiString: %v", err)
	}
	return nil
}

func (b *WeiString) BigInt() *big.Int {
	// this is stupid

	if b.Exp == 0 {
		return b.Int
	}
	if b.Exp < 0 {
		return b.Int
	}

	num := &big.Int{}
	num.Set(b.Int)
	mul := &big.Int{}
	mul.Exp(big.NewInt(10), big.NewInt(int64(b.Exp)), nil)
	num.Mul(num, mul)

	return num
}

type ValidatorStatsTableDbRow struct {
	ValidatorIndex uint64 `db:"validatorindex"`
	Day            int64  `db:"day"`

	StartBalance          int64 `db:"start_balance"`
	EndBalance            int64 `db:"end_balance"`
	MinBalance            int64 `db:"min_balance"`
	MaxBalance            int64 `db:"max_balance"`
	StartEffectiveBalance int64 `db:"start_effective_balance"`
	EndEffectiveBalance   int64 `db:"end_effective_balance"`
	MinEffectiveBalance   int64 `db:"min_effective_balance"`
	MaxEffectiveBalance   int64 `db:"max_effective_balance"`

	MissedAttestations      int64 `db:"missed_attestations"`
	MissedAttestationsTotal int64 `db:"missed_attestations_total"`
	OrphanedAttestations    int64 `db:"orphaned_attestations"`

	ParticipatedSync      int64 `db:"participated_sync"`
	ParticipatedSyncTotal int64 `db:"participated_sync_total"`
	MissedSync            int64 `db:"missed_sync"`
	MissedSyncTotal       int64 `db:"missed_sync_total"`
	OrphanedSync          int64 `db:"orphaned_sync"`
	OrphanedSyncTotal     int64 `db:"orphaned_sync_total"`

	ProposedBlocks int64 `db:"proposed_blocks"`
	MissedBlocks   int64 `db:"missed_blocks"`
	OrphanedBlocks int64 `db:"orphaned_blocks"`

	AttesterSlashings int64 `db:"attester_slashings"`
	ProposerSlashing  int64 `db:"proposer_slashings"`

	Deposits            int64 `db:"deposits"`
	DepositsTotal       int64 `db:"deposits_total"`
	DepositsAmount      int64 `db:"deposits_amount"`
	DepositsAmountTotal int64 `db:"deposits_amount_total"`

	Withdrawals            int64 `db:"withdrawals"`
	WithdrawalsTotal       int64 `db:"withdrawals_total"`
	WithdrawalsAmount      int64 `db:"withdrawals_amount"`
	WithdrawalsAmountTotal int64 `db:"withdrawals_amount_total"`

	ClRewardsGWei      int64 `db:"cl_rewards_gwei"`
	ClRewardsGWeiTotal int64 `db:"cl_rewards_gwei_total"`

	ClPerformance1d   int64 `db:"-"`
	ClPerformance7d   int64 `db:"-"`
	ClPerformance31d  int64 `db:"-"`
	ClPerformance365d int64 `db:"-"`

	ElRewardsWei      decimal.Decimal `db:"el_rewards_wei"`
	ElRewardsWeiTotal decimal.Decimal `db:"el_rewards_wei_total"`

	ElPerformance1d   decimal.Decimal `db:"-"`
	ElPerformance7d   decimal.Decimal `db:"-"`
	ElPerformance31d  decimal.Decimal `db:"-"`
	ElPerformance365d decimal.Decimal `db:"-"`

	MEVRewardsWei      decimal.Decimal `db:"mev_rewards_wei"`
	MEVRewardsWeiTotal decimal.Decimal `db:"mev_rewards_wei_total"`

	MEVPerformance1d   decimal.Decimal `db:"-"`
	MEVPerformance7d   decimal.Decimal `db:"-"`
	MEVPerformance31d  decimal.Decimal `db:"-"`
	MEVPerformance365d decimal.Decimal `db:"-"`
}
