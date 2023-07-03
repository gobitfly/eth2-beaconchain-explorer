package types

import (
	"math/big"
	"time"
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
