package services

import (
	"eth2-exporter/cache"
	"eth2-exporter/db"
	"eth2-exporter/utils"
	"fmt"
	"sync"
	"time"
)

const latestBlockNumberCacheKey = "latestEth1BlockNumber"

// latestBlockUpdater updates the most recent eth1 block number variable
func latestBlockUpdater(wg *sync.WaitGroup) {
	firstRun := true

	for {
		recent, err := db.BigtableClient.GetMostRecentBlockFromDataTable()
		if err != nil {
			logger.WithError(err).Error("error getting most recent eth1 block")
		}
		cacheKey := fmt.Sprintf("%d:frontend:%s", utils.Config.Chain.Config.DepositChainID, latestBlockNumberCacheKey)
		err = cache.TieredCache.SetUint64(cacheKey, recent.GetNumber(), time.Hour*24)
		if err != nil {
			logger.Errorf("error caching %s: %v", latestBlockNumberCacheKey, err)
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

// LatestEth1BlockNumber will return most recent eth1 block number
func LatestEth1BlockNumber() uint64 {
	cacheKey := fmt.Sprintf("%d:frontend:%s", utils.Config.Chain.Config.DepositChainID, latestBlockNumberCacheKey)

	if wanted, err := cache.TieredCache.GetUint64WithLocalTimeout(cacheKey, time.Second*5); err == nil {
		return wanted
	} else {
		logger.Errorf("error retrieving %s from cache: %v", latestBlockNumberCacheKey, err)
	}
	return 0
}

const latestBlockHashRootCacheKey = "latestEth1BlockRootHash"

// headBlockRootHashUpdater updates the hash of the current chain head block
func headBlockRootHashUpdater(wg *sync.WaitGroup) {
	firstRun := true

	for {
		blockRootHash := []byte{}
		err := db.ReaderDb.Get(&blockRootHash, `
		SELECT blockroot
		FROM blocks
		WHERE status = '1'
		ORDER BY slot DESC
		LIMIT 1`)

		if err != nil {
			logger.WithError(err).Error("error getting blockrroot hash for chain head")
		}
		cacheKey := fmt.Sprintf("%d:frontend:%s", utils.Config.Chain.Config.DepositChainID, latestBlockHashRootCacheKey)
		err = cache.TieredCache.SetString(cacheKey, string(blockRootHash), time.Hour*24)
		if err != nil {
			logger.Errorf("error caching %s: %v", latestBlockHashRootCacheKey, err)
		}

		if firstRun {
			logger.Info("initialized eth1 head block root hash updater")
			wg.Done()
			firstRun = false
		}
		ReportStatus("headBlockRootHashUpdater", "Running", nil)
		time.Sleep(time.Second * 10)
	}
}

// Eth1HeadBlockRootHash will return the hash of the current chain head block
func Eth1HeadBlockRootHash() []byte {
	cacheKey := fmt.Sprintf("%d:frontend:%s", utils.Config.Chain.Config.DepositChainID, latestBlockHashRootCacheKey)

	if wanted, err := cache.TieredCache.GetStringWithLocalTimeout(cacheKey, time.Second*5); err == nil {
		return []byte(wanted)
	} else {
		logger.Errorf("error retrieving %s from cache: %v", latestBlockHashRootCacheKey, err)
	}
	return []byte{}
}
