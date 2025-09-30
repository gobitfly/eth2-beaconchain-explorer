package services

import (
	"context"

	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
)

// RunStepImport runs the primary import step (from local CSV or Dune if configured).
func RunStepImport(ctx context.Context) error {
	cfg := utils.Config.ValidatorTagger
	if (cfg.Dune.ApiKey == "" || cfg.Dune.QueryID == 0) && utils.Config.ValidatorTagger.LocalCSVPath == "" {
		validatorTaggerLogger.Warn("no local csv path or dune settings configured; skipping import step")
		return nil
	}
	return refreshAndLoadValidatorNames(ctx, cfg)
}

// RunStepLido enriches Lido validators via the existing exporter/indexer.
func RunStepLido(ctx context.Context) error {
	return indexLidoValidators()
}

// RunStepLidoCSM enriches validators from Lido CSM module.
func RunStepLidoCSM(ctx context.Context) error {
	return indexLidoCSMValidators()
}

// RunStepRocketPool sets Rocket Pool sub_entity based on node address.
func RunStepRocketPool(ctx context.Context) error {
	return enrichRocketPoolSubEntity(ctx)
}

// RunStepWithdrawalTagging tags untagged validators based on withdrawal credentials and returns balances map for reuse.
func RunStepWithdrawalTagging(ctx context.Context) (map[uint64]int64, error) {
	return autoTagUntaggedValidatorsWithdrawal(ctx)
}

// RunStepDepositTagging tags remaining untagged validators by earliest deposit from_address.
// If balanceByIndex is nil, it will compute balances from withdrawal tagging logic first.
func RunStepDepositTagging(ctx context.Context, balanceByIndex map[uint64]int64) error {
	var err error
	if balanceByIndex == nil {
		balanceByIndex, err = autoTagUntaggedValidatorsWithdrawal(ctx)
		if err != nil {
			return err
		}
	}
	return autoTagUntaggedValidatorsByDeposit(ctx, balanceByIndex)
}

// RunStepPrecompute runs the precompute aggregation for entities across periods.
func RunStepPrecompute(ctx context.Context) error {
	return precomputeEntityData(ctx)
}

// RunStepPopulateValidatorNames populates validator_names based on validator_entities.
func RunStepPopulateValidatorNames(ctx context.Context) error {
	return populateValidatorNamesTable(ctx)
}
