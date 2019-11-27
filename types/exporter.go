package types

import (
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
)

type EpochData struct {
	Epoch                   uint64
	Validators              map[string]*ethpb.Validator
	ValidatorAssignmentes   []*ethpb.ValidatorAssignments_CommitteeAssignment
	BeaconCommittees        []*ethpb.BeaconCommittees_CommitteeItem
	ValidatorBalances       map[string]*ethpb.ValidatorBalances_Balance
	Blocks                  map[uint64]*BlockContainer
	EpochParticipationStats *ethpb.ValidatorParticipationResponse
}

type BlockContainer struct {
	Status   string
	Proposer []byte
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
