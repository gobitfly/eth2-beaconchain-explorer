package services

import (
	"eth2-exporter/db"
	"sync/atomic"
	"time"
)

var latestEth1BlockNumber uint64

func initExecutionLayerServices() {
	ready.Add(1)
	go latestBlockUpdater()
}

// latestBlockUpdater updates the most recent eth1 block number variable
func latestBlockUpdater() {
	firstRun := true

	for {
		recent, err := db.BigtableClient.GetMostRecentBlock()
		if err != nil {
			logger.WithError(err).Error("error getting most recent eth1 block")
		}
		if firstRun {
			logger.Info("initialized eth1 block updater")
			ready.Done()
			firstRun = false
		}
		atomic.StoreUint64(&latestEth1BlockNumber, recent.GetNumber())
		time.Sleep(time.Second)
	}
}

// LatestEth1BlockNumber will return the latest epoch
func LatestEth1BlockNumber() uint64 {
	return atomic.LoadUint64(&latestEth1BlockNumber)
}
