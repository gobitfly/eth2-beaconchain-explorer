package state

import (
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"sync"
)

var validatorState *types.ValidatorState
var validatorStateMu = &sync.RWMutex{}

func InitValidatorsSate() error {
	validatorState = &types.ValidatorState{
		Validators: make(map[uint64]*types.Validator),
		Epoch:      0,
	}
	return nil
}

func LoadValidatorState() error {
	return nil
}

func UpdateValidatorState(data *types.EpochData) error {
	validatorStateMu.Lock()
	defer validatorStateMu.Unlock()

	if data.Epoch != validatorState.Epoch+1 {
		return fmt.Errorf("error cannot write non-continous validator state %v (expected epoch %v)", data.Epoch, validatorState.Epoch+1)
	}

	farFutureEpoch := uint64(18446744073709551615)
	// farFutureEpochSQL := uint64(9223372036854775807)

	// Update the balance for all validators
	for _, validator := range data.Validators {
		if validatorState.Validators[validator.Index] == nil {
			validatorState.Validators[validator.Index] = validator
		}

		validatorState.Validators[validator.Index].Balance = validator.Balance
		validatorState.Validators[validator.Index].EffectiveBalance = validator.EffectiveBalance
		validatorState.Validators[validator.Index].Slashed = validator.Slashed
		validatorState.Validators[validator.Index].ActivationEligibilityEpoch = validator.ActivationEligibilityEpoch
		validatorState.Validators[validator.Index].ActivationEpoch = validator.ActivationEpoch
		validatorState.Validators[validator.Index].ExitEpoch = validator.ExitEpoch
		validatorState.Validators[validator.Index].WithdrawableEpoch = validator.WithdrawableEpoch
		validatorState.Validators[validator.Index].WithdrawalCredentials = validator.WithdrawalCredentials

		validatorState.Validators[validator.Index].Balance1d = validator.Balance1d
		validatorState.Validators[validator.Index].Balance7d = validator.Balance7d
		validatorState.Validators[validator.Index].Balance31d = validator.Balance31d

		if data.Epoch == validator.ActivationEpoch || (data.Epoch == 1 && validator.ActivationEpoch == 0) {
			validatorState.Validators[validator.Index].BalanceActivation = validator.Balance
		}
	}

	for _, blocks := range data.Blocks {
		for _, block := range blocks {
			for _, attestation := range block.Attestations {
				for _, attester := range attestation.Attesters {
					validatorState.Validators[attester].LastAttestationSlot = attestation.Data.Slot
				}
			}

			validatorState.Validators[block.Proposer].LastProposalSlot = block.Slot
		}
	}

	thresholdSlot := data.Epoch*utils.Config.Chain.SlotsPerEpoch - 96
	if data.Epoch < 2 {
		thresholdSlot = 0
	}

	for _, validator := range data.Validators {
		if validator.ExitEpoch <= data.Epoch && validator.Slashed {
			validator.Status = "slashed"
		} else if validator.ExitEpoch <= data.Epoch {
			validator.Status = "exited"
		} else if validator.ActivationEligibilityEpoch == farFutureEpoch {
			validator.Status = "deposited"
		} else if validator.ActivationEpoch > data.Epoch {
			validator.Status = "pending"
		} else if validator.Slashed && validator.ActivationEpoch < data.Epoch && (validator.LastAttestationSlot < thresholdSlot || validator.LastAttestationSlot == 0) {
			validator.Status = "slashing_offline"
		} else if validator.Slashed {
			validator.Status = "slashing_online"
		} else if validator.ExitEpoch < farFutureEpoch && (validator.LastAttestationSlot < thresholdSlot || validator.LastAttestationSlot == 0) {
			validator.Status = "exiting_offline"
		} else if validator.ExitEpoch < farFutureEpoch {
			validator.Status = "exiting_online"
		} else if validator.ActivationEpoch < data.Epoch && (validator.LastAttestationSlot < thresholdSlot || validator.LastAttestationSlot == 0) {
			validator.Status = "active_offline"
		} else {
			validator.Status = "active_online"
		}
	}

	validatorState.Epoch = data.Epoch

	return nil
}
