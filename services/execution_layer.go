package services

import (
	"eth2-exporter/cache"
	"eth2-exporter/db"
	"eth2-exporter/utils"
	"fmt"
	"sync"
	"time"
)

// latestBlockUpdater updates the most recent eth1 block number variable
func latestBlockUpdater(wg *sync.WaitGroup) {
	firstRun := true

	for {
		recent, err := db.BigtableClient.GetMostRecentBlockFromDataTable()
		if err != nil {
			logger.WithError(err).Error("error getting most recent eth1 block")
		}
		cacheKey := fmt.Sprintf("%d:frontend:latestEth1BlockNumber", utils.Config.Chain.Config.DepositChainID)
		err = cache.TieredCache.SetUint64(cacheKey, recent.GetNumber(), time.Hour*24)
		if err != nil {
			logger.Errorf("error caching latestBlockNumber: %v", err)
		}

		if firstRun {
			logger.Info("initialized eth1 block updater")
			wg.Done()
			firstRun = false
		}
		ReportStatus("latestBlockUpdater", "Running", nil)
		time.Sleep(time.Second * 10)
	}
}

// LatestEth1BlockNumber will return the latest epoch
func LatestEth1BlockNumber() uint64 {
	cacheKey := fmt.Sprintf("%d:frontend:latestEth1BlockNumber", utils.Config.Chain.Config.DepositChainID)

	if wanted, err := cache.TieredCache.GetUint64WithLocalTimeout(cacheKey, time.Second*5); err == nil {
		return wanted
	} else {
		logger.Errorf("error retrieving latestEth1BlockNumber from cache: %v", err)
	}
	return 0
}
