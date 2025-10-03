package services

import (
	"context"

	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
)

// RunStepImport runs the primary import step (from local CSV or Dune if configured).
func RunStepImport(ctx context.Context) (err error) {
	finish := withStepMetrics("validator_tagger_import")
	defer func() { finish(err == nil) }()

	cfg := utils.Config.ValidatorTagger
	if (cfg.Dune.ApiKey == "" || cfg.Dune.QueryID == 0) && utils.Config.ValidatorTagger.LocalCSVPath == "" {
		validatorTaggerLogger.Warn("no local csv path or dune settings configured; skipping import step")
		return nil
	}
	err = refreshAndLoadValidatorNames(ctx, cfg)
	return err
}

// RunStepLido enriches Lido validators via the existing exporter/indexer.
func RunStepLido(ctx context.Context) (err error) {
	finish := withStepMetrics("validator_tagger_lido")
	defer func() { finish(err == nil) }()
	return indexLidoValidators()
}

// RunStepLidoCSM enriches validators from Lido CSM module.
func RunStepLidoCSM(ctx context.Context) (err error) {
	finish := withStepMetrics("validator_tagger_lido_csm")
	defer func() { finish(err == nil) }()
	return indexLidoCSMValidators()
}

// RunStepLidoSimpleDVT enriches validators from Lido Simple DVT module.
func RunStepLidoSimpleDVT(ctx context.Context) (err error) {
	finish := withStepMetrics("validator_tagger_lido_simple_dvt")
	defer func() { finish(err == nil) }()
	return indexLidoSimpleDVTValidators()
}

// RunStepRocketPool sets Rocket Pool sub_entity based on node address.
func RunStepRocketPool(ctx context.Context) (err error) {
	finish := withStepMetrics("validator_tagger_rocketpool")
	defer func() { finish(err == nil) }()
	return enrichRocketPoolSubEntity(ctx)
}

// RunStepWithdrawalTagging tags untagged validators based on withdrawal credentials and returns balances map for reuse.
func RunStepWithdrawalTagging(ctx context.Context) (_ map[uint64]int64, err error) {
	finish := withStepMetrics("validator_tagger_withdrawal_tagging")
	defer func() { finish(err == nil) }()
	return autoTagUntaggedValidatorsWithdrawal(ctx)
}

// RunStepDepositTagging tags remaining untagged validators by earliest deposit from_address.
// If balanceByIndex is nil, it will compute balances from withdrawal tagging logic first.
func RunStepDepositTagging(ctx context.Context, balanceByIndex map[uint64]int64) (err error) {
	finish := withStepMetrics("validator_tagger_deposit_tagging")
	defer func() { finish(err == nil) }()

	if balanceByIndex == nil {
		balanceByIndex, err = autoTagUntaggedValidatorsWithdrawal(ctx)
		if err != nil {
			return err
		}
	}
	return autoTagUntaggedValidatorsByDeposit(ctx, balanceByIndex)
}

// RunStepPrecompute runs the precompute aggregation for entities across periods.
func RunStepPrecompute(ctx context.Context) (err error) {
	finish := withStepMetrics("validator_tagger_precompute")
	defer func() { finish(err == nil) }()
	return precomputeEntityData(ctx)
}

// RunStepPopulateValidatorNames populates validator_names based on validator_entities.
func RunStepPopulateValidatorNames(ctx context.Context) (err error) {
	finish := withStepMetrics("validator_tagger_populate_names")
	defer func() { finish(err == nil) }()
	return populateValidatorNamesTable(ctx)
}
