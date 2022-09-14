package services

import (
	"eth2-exporter/db"
	"sync"
	"sync/atomic"
	"time"
)

var latestEth1BlockNumber uint64

// latestBlockUpdater updates the most recent eth1 block number variable
func latestBlockUpdater(wg *sync.WaitGroup) {
	firstRun := true

	for {
		recent, err := db.BigtableClient.GetMostRecentBlockFromDataTable()
		if err != nil {
			logger.WithError(err).Error("error getting most recent eth1 block")
		}
		if firstRun {
			logger.Info("initialized eth1 block updater")
			wg.Done()
			firstRun = false
		}
		atomic.StoreUint64(&latestEth1BlockNumber, recent.GetNumber())
		time.Sleep(time.Second * 10)
	}
}

// LatestEth1BlockNumber will return the latest epoch
func LatestEth1BlockNumber() uint64 {
	return atomic.LoadUint64(&latestEth1BlockNumber)
}
