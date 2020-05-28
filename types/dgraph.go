package types

type EpochNode struct {
	Epoch                   uint64
	Validators              []*Validator
	ValidatorIndices        map[string]uint64
	ValidatorAssignmentes   *EpochAssignments
	BeaconCommittees        map[uint64][]*BeaconCommitteItem
	Blocks                  map[uint64]map[string]*Block
	EpochParticipationStats *ValidatorParticipation
}

type BlockNode struct {
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

type ValidatorNode struct {
	Index                      uint64
	PublicKey                  []byte
	Balance                    uint64
	EffectiveBalance           uint64
	Slashed                    bool
	ActivationEligibilityEpoch uint64
	ActivationEpoch            uint64
	ExitEpoch                  uint64
	WithdrawableEpoch          uint64
	WithdrawalCredentials      []byte
}
