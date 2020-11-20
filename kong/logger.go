package kong

import "github.com/sirupsen/logrus"

var logger = logrus.New().WithField("module", "kong")
