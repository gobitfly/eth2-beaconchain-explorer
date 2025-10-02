package services

import (
	"time"

	"github.com/gobitfly/eth2-beaconchain-explorer/metrics"
	"github.com/sirupsen/logrus"
)

// validatorTaggerLogger is the module-scoped logger for the validator tagger service.
var validatorTaggerLogger = logrus.StandardLogger().WithField("module", "services.validator_tagger")

// withStepMetrics returns a finisher function that records duration and errors for a step.
// Usage:
//
//	finish := withStepMetrics("validator_tagger_import")
//	defer finish(err == nil)
func withStepMetrics(step string) func(success bool) {
	start := time.Now()
	metrics.Tasks.WithLabelValues(step).Inc()
	return func(success bool) {
		metrics.TaskDuration.WithLabelValues(step).Observe(time.Since(start).Seconds())
		if !success {
			metrics.Errors.WithLabelValues(step).Inc()
		}
	}
}
