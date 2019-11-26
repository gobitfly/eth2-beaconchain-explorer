package services

import (
	"eth2-exporter/db"
	"eth2-exporter/utils"
	"github.com/sirupsen/logrus"
	"sync/atomic"
	"time"
)

var latestEpoch uint64
var logger = logrus.New().WithField("module", "services")

func Init() {
	go epochUpdater()
}

func epochUpdater() {
	for true {
		var epoch uint64
		err := db.DB.Get(&epoch, "SELECT MAX(epoch)FROM epochs")

		if err != nil {
			logger.Printf("Error retrieving latest epoch from the database: %v", err)
		} else {
			atomic.StoreUint64(&latestEpoch, epoch)
		}
		time.Sleep(time.Second)
	}
}

func LatestEpoch() uint64 {
	return atomic.LoadUint64(&latestEpoch)
}

func IsSyncing() bool {
	return time.Now().Add(time.Minute * -5).After(utils.EpochToTime(LatestEpoch()))
}
