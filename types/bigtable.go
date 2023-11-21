package types

import (
	"math/big"
	"time"

	gcp_bigtable "cloud.google.com/go/bigtable"
)

type ValidatorBalanceStatistic struct {
	Index                 uint64
	MinEffectiveBalance   uint64
	MaxEffectiveBalance   uint64
	MinBalance            uint64
	MaxBalance            uint64
	StartEffectiveBalance uint64
	EndEffectiveBalance   uint64
	StartBalance          uint64
	EndBalance            uint64
}

type ValidatorMissedAttestationsStatistic struct {
	Index              uint64
	MissedAttestations uint64
}

type ValidatorSyncDutiesStatistic struct {
	Index            uint64
	ParticipatedSync uint64
	MissedSync       uint64
	OrphanedSync     uint64
}

type ValidatorWithdrawal struct {
	Index  uint64
	Epoch  uint64
	Slot   uint64
	Amount uint64
}

type ValidatorProposal struct {
	Index  uint64
	Slot   uint64
	Status uint64
}

type ValidatorEffectiveness struct {
	Validatorindex        uint64  `json:"validatorindex"`
	AttestationEfficiency float64 `json:"attestation_efficiency"`
}

type GasNowHistory struct {
	Ts       time.Time
	Slow     *big.Int
	Standard *big.Int
	Fast     *big.Int
	Rapid    *big.Int
}

type BulkMutations struct {
	Keys []string
	Muts []*gcp_bigtable.Mutation
}

func NewBulkMutations(length int) *BulkMutations {
	return &BulkMutations{
		Keys: make([]string, 0, length),
		Muts: make([]*gcp_bigtable.Mutation, 0, length),
	}
}

func (bulkMutations *BulkMutations) Add(key string, mut *gcp_bigtable.Mutation) {
	bulkMutations.Keys = append(bulkMutations.Keys, key)
	bulkMutations.Muts = append(bulkMutations.Muts, mut)
}

func (bulkMutations *BulkMutations) Len() int {
	return len(bulkMutations.Keys)
}

func (bulkMutations *BulkMutations) Less(i, j int) bool {
	return bulkMutations.Keys[i] < bulkMutations.Keys[j]
}

func (bulkMutations *BulkMutations) Swap(i, j int) {
	bulkMutations.Keys[i], bulkMutations.Keys[j] = bulkMutations.Keys[j], bulkMutations.Keys[i]
	bulkMutations.Muts[i], bulkMutations.Muts[j] = bulkMutations.Muts[j], bulkMutations.Muts[i]
}

type BulkMutation struct {
	Key string
	Mut *gcp_bigtable.Mutation
}
