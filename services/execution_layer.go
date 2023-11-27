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
const latestBlockHashRootCacheKey = "latestEth1BlockRootHash"

// latestBlockUpdater updates the most recent eth1 block number variable
func latestBlockUpdater(wg *sync.WaitGroup) {
	firstRun := true

	for {
		recent, err := db.BigtableClient.GetMostRecentBlockFromDataTable()
		if err != nil {
			utils.LogError(err, "error getting most recent eth1 block", 0)
		}
		cacheKey := fmt.Sprintf("%d:frontend:%s", utils.Config.Chain.ClConfig.DepositChainID, latestBlockNumberCacheKey)
		err = cache.TieredCache.SetUint64(cacheKey, recent.GetNumber(), utils.Day)
		if err != nil {
			utils.LogError(err, fmt.Sprintf("error caching latest block number with cache key %s", latestBlockNumberCacheKey), 0)
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
	cacheKey := fmt.Sprintf("%d:frontend:%s", utils.Config.Chain.ClConfig.DepositChainID, latestBlockNumberCacheKey)

	if wanted, err := cache.TieredCache.GetUint64WithLocalTimeout(cacheKey, time.Second*5); err == nil {
		return wanted
	} else {
		utils.LogError(err, fmt.Sprintf("error retrieving latest block number from cache with key %s", latestBlockNumberCacheKey), 0)
	}
	return 0
}

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
			utils.LogError(err, "error getting blockroot hash for chain head", 0)
		}
		cacheKey := fmt.Sprintf("%d:frontend:%s", utils.Config.Chain.ClConfig.DepositChainID, latestBlockHashRootCacheKey)
		err = cache.TieredCache.SetString(cacheKey, string(blockRootHash), utils.Day)
		if err != nil {
			utils.LogError(err, fmt.Sprintf("error caching latest blockroot hash with cache key %s", latestBlockHashRootCacheKey), 0)
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
	cacheKey := fmt.Sprintf("%d:frontend:%s", utils.Config.Chain.ClConfig.DepositChainID, latestBlockHashRootCacheKey)

	if wanted, err := cache.TieredCache.GetStringWithLocalTimeout(cacheKey, time.Second*5); err == nil {
		return []byte(wanted)
	} else {
		utils.LogError(err, fmt.Sprintf("error retrieving latest blockroot hash from cache with key %s", latestBlockHashRootCacheKey), 0)
	}
	return []byte{}
}
