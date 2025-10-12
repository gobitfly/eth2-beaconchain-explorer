package services

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RunValidatorTaggerScheduler starts the scheduler loop:
// - Daily at 10:00 UTC: runs all daily steps in order, then precompute.
// - Hourly at HH:00 UTC (other than 10:00): runs precompute only.
// Uses validator_tagger_job_runs to track scheduling, status, and errors.
func RunValidatorTaggerScheduler() {
	cfg := utils.Config.ValidatorTagger
	if !cfg.Enabled {
		validatorSchedulerLogger.Info("validator tagger disabled; scheduler exiting")
		return
	}

	validatorSchedulerLogger.WithField("chainName", utils.Config.Chain.ClConfig.ConfigName).Info("starting validator tagger scheduler")

	ctx := context.Background()
	for {
		now := time.Now().UTC()
		nextHourly := now.Truncate(time.Hour).Add(time.Hour)
		nextDaily := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, time.UTC)
		if !nextDaily.After(now) {
			nextDaily = nextDaily.Add(24 * time.Hour)
		}
		next := nextHourly
		isDaily := false
		if nextDaily.Before(next) || nextDaily.Equal(next) {
			next = nextDaily
			isDaily = true
		}

		sleepFor := time.Until(next)
		if sleepFor > 0 {
			validatorSchedulerLogger.WithFields(logrus.Fields{"sleep": sleepFor.String(), "wakeupAt": next}).Info("scheduler sleeping until next tick")
			time.Sleep(sleepFor)
		}

		// Exactly at the chosen boundary
		boundary := next
		if isDaily {
			// Run the daily chain, then precompute
			runGroup := uuid.New().String()
			validatorSchedulerLogger.WithFields(logrus.Fields{"boundary": boundary, "group": runGroup}).Info("daily run starting")
			runDailyChain(ctx, boundary, runGroup)
			// After the daily step completes (success or partial), run the precompute step bound to the same group id
			if stepEnabled := func(b *bool) bool { return b == nil || *b }; stepEnabled(utils.Config.ValidatorTagger.Steps.PrecomputeEnabled) {
				if err := runScheduledJob(ctx, "precompute", boundary, &runGroup, "scheduler", func() error { return RunStepPrecompute(ctx) }); err != nil {
					validatorSchedulerLogger.WithError(err).Error("precompute job error (10:00)")
				}
			} else {
				validatorSchedulerLogger.Info("step disabled: precompute")
			}
		} else {
			// hourly precompute only (not 10:00)
			if boundary.Hour() != 10 {
				if stepEnabled := func(b *bool) bool { return b == nil || *b }; stepEnabled(utils.Config.ValidatorTagger.Steps.PrecomputeEnabled) {
					if err := runScheduledJob(ctx, "precompute", boundary, nil, "scheduler", func() error { return RunStepPrecompute(ctx) }); err != nil {
						validatorSchedulerLogger.WithError(err).Error("precompute job error (hourly)")
					}
				} else {
					validatorSchedulerLogger.Info("step disabled: precompute")
				}
			}
		}
	}
}

func runDailyChain(ctx context.Context, scheduledAt time.Time, runGroupID string) {
	cfg := utils.Config.ValidatorTagger
	// Determine enabled steps for the daily chain (all except the precompute step)
	type namedStep struct {
		name string
		fn   func(context.Context) error
	}
	steps := make([]namedStep, 0, 6)
	stepEnabled := func(b *bool) bool { return b == nil || *b }
	if stepEnabled(cfg.Steps.ImportEnabled) {
		steps = append(steps, namedStep{"import", func(ctx context.Context) error { return RunStepImport(ctx) }})
	}
	if stepEnabled(cfg.Steps.LidoEnabled) {
		steps = append(steps, namedStep{"lido", func(ctx context.Context) error { return RunStepLido(ctx) }})
	}
	if stepEnabled(cfg.Steps.LidoCSMEnabled) {
		steps = append(steps, namedStep{"lido_csm", func(ctx context.Context) error { return RunStepLidoCSM(ctx) }})
	}
	if stepEnabled(cfg.Steps.RocketPoolSubEntityEnabled) {
		steps = append(steps, namedStep{"rocketpool", func(ctx context.Context) error { return RunStepRocketPool(ctx) }})
	}
	if stepEnabled(cfg.Steps.WithdrawalTaggingEnabled) {
		steps = append(steps, namedStep{"withdrawal_tagging", func(ctx context.Context) error { _, err := RunStepWithdrawalTagging(ctx); return err }})
	}
	if stepEnabled(cfg.Steps.DepositTaggingEnabled) {
		// Deposit tagging benefits from the prior balance map, but the wrapper recomputes if needed
		steps = append(steps, namedStep{"deposit_tagging", func(ctx context.Context) error { return RunStepDepositTagging(ctx, nil) }})
	}
	if stepEnabled(cfg.Steps.PopulateNamesEnabled) {
		steps = append(steps, namedStep{"populate_validator_names", func(ctx context.Context) error { return RunStepPopulateValidatorNames(ctx) }})
	}

	for _, s := range steps {
		if err := runScheduledJob(ctx, s.name, scheduledAt, &runGroupID, "scheduler", func() error { return s.fn(ctx) }); err != nil {
			validatorSchedulerLogger.WithFields(logrus.Fields{"step": s.name}).WithError(err).Error("daily step failed")
			// Continue with the next steps as per requirement; failures are reported in the tracking table
		}
	}
}

// runScheduledJob ensures a row exists, attempts to claim it, runs the fn, and records status.
func runScheduledJob(ctx context.Context, jobName string, scheduledAt time.Time, runGroupID *string, triggeredBy string, fn func() error) error {
	if err := ensureJobRow(ctx, jobName, scheduledAt, runGroupID, triggeredBy); err != nil {
		return err
	}
	id, ok, err := tryStartJob(ctx, jobName, scheduledAt)
	if err != nil {
		return err
	}
	if !ok {
		validatorSchedulerLogger.WithFields(logrus.Fields{"job": jobName, "scheduled_at": scheduledAt}).Info("job already running or finished; skipping")
		return nil
	}
	start := time.Now()
	var runErr error
	defer func() {
		status := JobStatusOK
		var errText *string
		if runErr != nil {
			status = JobStatusError
			msg := runErr.Error()
			errText = &msg
		}
		if err := finishJob(ctx, id, status, errText); err != nil {
			validatorSchedulerLogger.WithError(err).Error("finishJob failed")
		}
		validatorSchedulerLogger.WithFields(logrus.Fields{"job": jobName, "duration": time.Since(start).String(), "status": status}).Info("job finished")
	}()

	runErr = fn()
	return runErr
}

// RunValidatorTaggerOnDemand runs one or more steps provided by name, in order, and records them as manual runs.
func RunValidatorTaggerOnDemand(ctx context.Context, stepsCSV string) error {
	names := strings.Split(stepsCSV, ",")
	for i := range names {
		names[i] = strings.TrimSpace(strings.ToLower(names[i]))
	}
	// Expand aliases and validate
	expanded, err := expandAndValidateSteps(names)
	if err != nil {
		return err
	}
	// manual runs share a group id
	runGroup := uuid.New().String()
	scheduledAt := time.Now().UTC().Truncate(time.Second)
	for _, name := range expanded {
		switch name {
		case "import":
			if err := runScheduledJob(ctx, name, scheduledAt, &runGroup, "manual", func() error { return RunStepImport(ctx) }); err != nil {
				return err
			}
		case "lido":
			if err := runScheduledJob(ctx, name, scheduledAt, &runGroup, "manual", func() error { return RunStepLido(ctx) }); err != nil {
				return err
			}
		case "lido_csm":
			if err := runScheduledJob(ctx, name, scheduledAt, &runGroup, "manual", func() error { return RunStepLidoCSM(ctx) }); err != nil {
				return err
			}
		case "rocketpool":
			if err := runScheduledJob(ctx, name, scheduledAt, &runGroup, "manual", func() error { return RunStepRocketPool(ctx) }); err != nil {
				return err
			}
		case "withdrawal_tagging":
			if err := runScheduledJob(ctx, name, scheduledAt, &runGroup, "manual", func() error { _, e := RunStepWithdrawalTagging(ctx); return e }); err != nil {
				return err
			}
		case "deposit_tagging":
			if err := runScheduledJob(ctx, name, scheduledAt, &runGroup, "manual", func() error { return RunStepDepositTagging(ctx, nil) }); err != nil {
				return err
			}
		case "populate_validator_names":
			if err := runScheduledJob(ctx, name, scheduledAt, &runGroup, "manual", func() error { return RunStepPopulateValidatorNames(ctx) }); err != nil {
				return err
			}
		case "precompute":
			if err := runScheduledJob(ctx, name, scheduledAt, &runGroup, "manual", func() error { return RunStepPrecompute(ctx) }); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown step: %s", name)
		}
	}
	return nil
}

func expandAndValidateSteps(names []string) ([]string, error) {
	set := map[string]struct{}{}
	for _, n := range names {
		if n == "all" {
			// Expand to daily steps only (no precompute) as "all" is intended for the daily chain
			for _, s := range []string{"import", "lido", "lido_csm", "rocketpool", "withdrawal_tagging", "deposit_tagging", "populate_validator_names"} {
				set[s] = struct{}{}
			}
			continue
		}
		switch n {
		case "import", "lido", "lido_csm", "rocketpool", "withdrawal_tagging", "deposit_tagging", "populate_validator_names", "precompute":
			set[n] = struct{}{}
		default:
			return nil, fmt.Errorf("invalid step name: %s", n)
		}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	// Sort to provide a deterministic order close to the daily chain when "all" or multiple provided
	order := map[string]int{
		"import":                   1,
		"lido":                     2,
		"lido_csm":                 3,
		"rocketpool":               4,
		"withdrawal_tagging":       5,
		"deposit_tagging":          6,
		"populate_validator_names": 7,
		"precompute":               100,
	}
	sort.Slice(out, func(i, j int) bool { return order[out[i]] < order[out[j]] })
	return out, nil
}
