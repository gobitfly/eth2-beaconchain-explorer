package types

import (
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
)

type EpochData struct {
	Epoch                   uint64
	Validators              []*ethpb.Validator
	ValidatorAssignmentes   *EpochAssignments
	BeaconCommittees        []*ethpb.BeaconCommittees_CommitteeItem
	ValidatorBalances       []*ethpb.ValidatorBalances_Balance
	Blocks                  map[uint64]*BlockContainer
	EpochParticipationStats *ethpb.ValidatorParticipationResponse
}

type BlockContainer struct {
	Status   uint64
	Proposer uint64
	Block    *ethpb.BeaconBlockContainer
}

type MinimalBlock struct {
	Epoch    uint64
	Slot     uint64
	BockRoot []byte `db:"blockroot"`
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
