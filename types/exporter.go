package types

import (
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
)

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

type ValidatorParticipation struct {
	Epoch                   uint64
	Finalized               bool
	GlobalParticipationRate float32
	VotedEther              uint64
	EligibleEther           uint64
}

type ValidatorBalance struct {
	PublicKey []byte
	Index     uint64
	Balance   uint64
}

type BeaconCommitteItem struct {
	ValidatorIndices []uint64
}

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

type ValidatorQueue struct {
	ChurnLimit           uint64
	ActivationPublicKeys [][]byte
	ExitPublicKeys       [][]byte
}

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

type Eth1Data struct {
	DepositRoot  []byte
	DepositCount uint64
	BlockHash    []byte
}

type ProposerSlashing struct {
	ProposerIndex uint64
	Header_1      *Block
	Header_2      *Block
}

type AttesterSlashing struct {
	Attestation_1 *IndexedAttestation
	Attestation_2 *IndexedAttestation
}

type IndexedAttestation struct {
	CustodyBit_0Indices []uint64
	CustodyBit_1Indices []uint64
	Data                *AttestationData
	Signature           []byte
}

type Attestation struct {
	AggregationBits []byte
	Attesters       []uint64
	Data            *AttestationData
	CustodyBits     []byte
	Signature       []byte
}

type AttestationData struct {
	Slot            uint64
	CommitteeIndex  uint64
	BeaconBlockRoot []byte
	Source          *Checkpoint
	Target          *Checkpoint
}

type Checkpoint struct {
	Epoch uint64
	Root  []byte
}

type Deposit struct {
	Proof                 [][]byte
	PublicKey             []byte
	WithdrawalCredentials []byte
	Amount                uint64
	Signature             []byte
}

type VoluntaryExit struct {
	Epoch          uint64
	ValidatorIndex uint64
	Signature      []byte
}

type BlockContainer struct {
	Status   uint64
	Proposer uint64

	Block *ethpb.BeaconBlockContainer
}

type MinimalBlock struct {
	Epoch      uint64 `db:"epoch"`
	Slot       uint64 `db:"slot"`
	BlockRoot  []byte `db:"blockroot"`
	ParentRoot []byte `db:"parentroot"`
}

type BlockComparisonContainer struct {
	Epoch uint64
	Db    *MinimalBlock
	Node  *MinimalBlock
}

type EpochAssignments struct {
	ProposerAssignments map[uint64]uint64
	AttestorAssignments map[string]uint64
}
