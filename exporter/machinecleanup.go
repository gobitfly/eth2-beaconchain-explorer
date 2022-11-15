package exporter

import (
	"eth2-exporter/db"
	"eth2-exporter/utils"
	"time"
)

// todo: remove once migrated
func cleanupOldMachineStats() {
	if !utils.Config.Frontend.CleanupOldMachineStats {
		return
	}
	for {
		start := time.Now()

		err := db.CleanupOldMachineStats()

		if err != nil {
			logger.Errorf("error machineclean data db: %v", err)
			return
		}

		logger.WithField("duration", time.Since(start)).Info("machineclean completed")
		time.Sleep(time.Second * 60 * 60 * 1)
	}
}
