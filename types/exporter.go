package types

import (
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
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

// EpochData is a struct to hold epoch data
type EpochData struct {
	Epoch                   uint64
	Validators              []*Validator
	ValidatorAssignmentes   *EpochAssignments
	Blocks                  map[uint64]map[string]*Block
	EpochParticipationStats *ValidatorParticipation
}

// ValidatorParticipation is a struct to hold validator participation data
type ValidatorParticipation struct {
	Epoch                   uint64
	Finalized               bool
	GlobalParticipationRate float32
	VotedEther              uint64
	EligibleEther           uint64
}

// BeaconCommitteItem is a struct to hold beacon committee data
type BeaconCommitteItem struct {
	ValidatorIndices []uint64
}

// Validator is a struct to hold validator data
type Validator struct {
	Index                      uint64 `db:"validatorindex"`
	PublicKey                  []byte `db:"pubkey"`
	Balance                    uint64 `db:"balance"`
	EffectiveBalance           uint64 `db:"effectivebalance"`
	Slashed                    bool   `db:"slashed"`
	ActivationEligibilityEpoch uint64 `db:"activationeligibilityepoch"`
	ActivationEpoch            uint64 `db:"activationepoch"`
	ExitEpoch                  uint64 `db:"exitepoch"`
	WithdrawableEpoch          uint64 `db:"withdrawableepoch"`
	WithdrawalCredentials      []byte `db:"withdrawalcredentials"`

	BalanceActivation uint64 `db:"balanceactivation"`
	Balance1d         uint64 `db:"balance1d"`
	Balance7d         uint64 `db:"balance7d"`
	Balance31d        uint64 `db:"balance31d"`
	Status            string `db:"status"`

	LastAttestationSlot uint64 `db:"lastattestationslot"`
	LastProposalSlot    uint64 `db:"lastproposalslot"`
}

// ValidatorQueue is a struct to hold validator queue data
type ValidatorQueue struct {
	Activating uint64
	Exititing  uint64
}

type SyncAggregate struct {
	SyncCommitteeValidators    []uint64
	SyncCommitteeBits          []byte
	SyncCommitteeSignature     []byte
	SyncAggregateParticipation float64
}

// Block is a struct to hold block data
type Block struct {
	Status            uint64
	Proposer          uint64
	BlockRoot         []byte
	Slot              uint64
	ParentRoot        []byte
	StateRoot         []byte
	Signature         []byte
	RandaoReveal      []byte
	Graffiti          []byte
	Eth1Data          *Eth1Data
	BodyRoot          []byte
	ProposerSlashings []*ProposerSlashing
	AttesterSlashings []*AttesterSlashing
	Attestations      []*Attestation
	Deposits          []*Deposit
	VoluntaryExits    []*VoluntaryExit
	SyncAggregate     *SyncAggregate    // warning: sync aggregate may be nil, for phase0 blocks
	ExecutionPayload  *ExecutionPayload // warning: payload may be nil, for phase0/altair blocks
	Canonical         bool
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

// BlockContainer is a struct to hold block container data
type BlockContainer struct {
	Status   uint64
	Proposer uint64

	Block *ethpb.BeaconBlockContainer
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

// BlockComparisonContainer is a struct to hold block comparison data
type BlockComparisonContainer struct {
	Epoch uint64
	Db    *MinimalBlock
	Node  *MinimalBlock
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
