package types

import (
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
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

// EpochData is a struct to hold epoch data
type EpochData struct {
	Epoch                   uint64
	Validators              []*Validator
	ValidatorIndices        map[string]uint64
	ValidatorAssignmentes   *EpochAssignments
	BeaconCommittees        map[uint64][]*BeaconCommitteItem
	ValidatorBalances       []*ValidatorBalance
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

// ValidatorBalance is a struct to hold validator balance data
type ValidatorBalance struct {
	PublicKey []byte
	Index     uint64
	Balance   uint64
}

// BeaconCommitteItem is a struct to hold beacon committee data
type BeaconCommitteItem struct {
	ValidatorIndices []uint64
}

// Validator is a struct to hold validator data
type Validator struct {
	PublicKey                  []byte
	WithdrawalCredentials      []byte
	EffectiveBalance           uint64
	Slashed                    bool
	ActivationEligibilityEpoch uint64
	ActivationEpoch            uint64
	ExitEpoch                  uint64
	WithdrawableEpoch          uint64
}

// ValidatorQueue is a struct to hold validator queue data
type ValidatorQueue struct {
	ChurnLimit           uint64
	ActivationPublicKeys [][]byte
	ExitPublicKeys       [][]byte
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
	Custodybit0indices []uint64
	Custodybit1indices []uint64
	Data               *AttestationData
	Signature          []byte
}

// Attestation is a struct to hold attestation header data
type Attestation struct {
	AggregationBits []byte
	Attesters       []uint64
	Data            *AttestationData
	CustodyBits     []byte
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
}
