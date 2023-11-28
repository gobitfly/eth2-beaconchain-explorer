package userService

import (
	"github.com/sirupsen/logrus"
)

var logger = logrus.New().WithField("module", "userService")

func Init() {
	logger.Info("starting user service")
	go stripeEmailUpdater()
}
