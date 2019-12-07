package types

import (
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
)

type EpochData struct {
	Epoch                   uint64
	Validators              []*ethpb.Validator
	ValidatorAssignmentes   *EpochAssignments
	BeaconCommittees        map[uint64][]*ethpb.BeaconCommittees_CommitteeItem
	ValidatorBalances       []*ethpb.ValidatorBalances_Balance
	Blocks                  map[uint64]map[string]*BlockContainer
	EpochParticipationStats *ethpb.ValidatorParticipationResponse
}

type BlockContainer struct {
	Status   uint64
	Proposer uint64
	Block    *ethpb.BeaconBlockContainer
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
