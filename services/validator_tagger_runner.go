package services

import (
	"context"
	"strings"
	"time"

	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
	"github.com/sirupsen/logrus"
)

// RunValidatorTagger runs the validator tagger loop in the current goroutine.
func RunValidatorTagger() {
	cfg := utils.Config.ValidatorTagger
	if !cfg.Enabled {
		validatorTaggerLogger.Info("validator tagger disabled; exiting")
		return
	}

	if strings.TrimSpace(cfg.LocalCSVPath) == "" && (cfg.Dune.ApiKey == "" || cfg.Dune.QueryID == 0) {
		validatorTaggerLogger.Warn("no local csv path or dune settings configured; skipping primary import; will run tagging only")
	}

	interval := cfg.Interval
	if interval == 0 {
		interval = 24 * time.Hour
	}

	validatorTaggerLogger.WithFields(logrus.Fields{
		"interval": interval.String(),
		"query_id": cfg.Dune.QueryID,
	}).Info("starting validator tagger service")

	importValidatorTags := func() {
		start := time.Now()
		name := "validator_tagger"
		ReportStatus(name, "running", nil)

		stepEnabled := func(b *bool) bool { return b == nil || *b }

		// 1) Primary import from either local CSV (if provided) or Dune
		if stepEnabled(cfg.Steps.ImportEnabled) {
			if strings.TrimSpace(cfg.LocalCSVPath) != "" || (cfg.Dune.ApiKey != "" && cfg.Dune.QueryID != 0) {
				if err := refreshAndLoadValidatorNames(context.Background(), cfg); err != nil {
					utils.LogError(err, "validator_tagger primary import failed", 0, map[string]interface{}{"duration": time.Since(start).String()})
					// Continue further tagging even if the primary import fails
				}
			} else {
				validatorTaggerLogger.Warn("no local csv path or dune settings configured; skipping primary import step")
			}
		} else {
			validatorTaggerLogger.Info("step disabled: import")
		}

		// 2) Lido exporter step: enrich Lido data between import and grouping steps
		if stepEnabled(cfg.Steps.LidoEnabled) {
			if err := indexLidoValidators(); err != nil {
				utils.LogError(err, "validator_tagger lido exporter step failed", 0, map[string]interface{}{"duration": time.Since(start).String()})
				// Continue with next steps even if the Lido step fails
			}
		} else {
			validatorTaggerLogger.Info("step disabled: lido")
		}

		// 2a) Lido CSM enrichment stage: tag validators from Lido CSM module
		if stepEnabled(cfg.Steps.LidoCSMEnabled) {
			if err := indexLidoCSMValidators(); err != nil {
				utils.LogError(err, "validator_tagger lido CSM enrichment step failed", 0, map[string]interface{}{"duration": time.Since(start).String()})
				// Continue with next steps even if the CSM step fails
			}
		} else {
			validatorTaggerLogger.Info("step disabled: lidoCSM")
		}

		// 2b) Rocket Pool enrichment stage: set sub_entity to node_address for 'Rocket Pool'
		if stepEnabled(cfg.Steps.RocketPoolSubEntityEnabled) {
			if err := enrichRocketPoolSubEntity(context.Background()); err != nil {
				utils.LogError(err, "validator_tagger rocketpool enrichment step failed", 0, map[string]interface{}{"duration": time.Since(start).String()})
				// Continue with generic/deposit tagging even if Rocket Pool step fails
			}
		} else {
			validatorTaggerLogger.Info("step disabled: rocketpoolSubEntity")
		}

		// 3) Generic tagging for remaining untagged validators (by withdrawal credentials)
		var balanceByIndex map[uint64]int64
		if stepEnabled(cfg.Steps.WithdrawalTaggingEnabled) {
			var err error
			balanceByIndex, err = autoTagUntaggedValidatorsWithdrawal(context.Background())
			if err != nil {
				utils.LogError(err, "validator_tagger generic tagging failed", 0, map[string]interface{}{"duration": time.Since(start).String()})
				ReportStatus(name, "error", nil)
				return
			}
		} else {
			validatorTaggerLogger.Info("step disabled: withdrawalTagging")
		}

		// 3) Deposit-based tagging for remaining untagged validators (by earliest from_address)
		if stepEnabled(cfg.Steps.DepositTaggingEnabled) {
			if err := autoTagUntaggedValidatorsByDeposit(context.Background(), balanceByIndex); err != nil {
				utils.LogError(err, "validator_tagger deposit-based tagging failed", 0, map[string]interface{}{"duration": time.Since(start).String()})
				ReportStatus(name, "error", nil)
				return
			}
		} else {
			validatorTaggerLogger.Info("step disabled: depositTagging")
		}

		// 4) Populate validator_names from validator_entities
		if stepEnabled(cfg.Steps.PopulateNamesEnabled) {
			if err := populateValidatorNamesTable(context.Background()); err != nil {
				utils.LogError(err, "validator_tagger populateValidatorNamesTable failed", 0, map[string]interface{}{"duration": time.Since(start).String()})
				// Continue to precompute even if this fails
			}
		} else {
			validatorTaggerLogger.Info("step disabled: populateNames")
		}

		// 5) Final step: precompute aggregated entity data (prints to console)
		if stepEnabled(cfg.Steps.PrecomputeEnabled) {
			if err := precomputeEntityData(context.Background()); err != nil {
				utils.LogError(err, "validator_tagger precomputeEntityData failed", 0, map[string]interface{}{"duration": time.Since(start).String()})
				// Do not mark the whole run as error; continue and report ok but with warning in logs
			}
		} else {
			validatorTaggerLogger.Info("step disabled: precompute")
		}

		validatorTaggerLogger.WithField("duration", time.Since(start)).Info("validator tagger run completed")
		ReportStatus(name, "ok", nil)
	}

	importValidatorTags()
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		<-t.C
		importValidatorTags()
	}
}
