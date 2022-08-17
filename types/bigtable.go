package types

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

type ValidatorEffectiveness struct {
	Validatorindex        uint64  `json:"validatorindex"`
	AttestationEfficiency float64 `json:"attestation_efficiency"`
}
