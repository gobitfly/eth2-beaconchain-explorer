package services

import "github.com/sirupsen/logrus"

// validatorTaggerLogger is the module-scoped logger for the validator tagger service.
var validatorTaggerLogger = logrus.StandardLogger().WithField("module", "services.validator_tagger")
