// Copyright (C) 2025 Bitfly GmbH
//
// This file is part of Beaconchain Dashboard.
//
// Beaconchain Dashboard is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Beaconchain Dashboard is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Beaconchain Dashboard.  If not, see <https://www.gnu.org/licenses/>.

package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/gobitfly/eth2-beaconchain-explorer/cache"
	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/rpc"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func queueEstimateUpdater(wg *sync.WaitGroup) {
	firstRun := true

	for {
		data, err := getQueuesEstimate()
		if err != nil {
			logger.Warnf("error retrieving queue data: %v", err)
			time.Sleep(time.Minute)
			continue
		}

		err = cache.TieredCache.Set(getQueueCacheKey(), data, utils.Day)
		if err != nil {
			logger.Errorf("error caching queue data: %v", err)
		}

		if firstRun {
			logrus.Info("initialized queue updater")
			wg.Done()
			firstRun = false
		}

		ReportStatus("queueEstimateUpdater", "Running", nil)
		time.Sleep(1 * time.Minute)
	}
}

func LatestQueueData() *types.QueuesEstimate {
	wanted := &types.QueuesEstimate{}
	if wanted, err := cache.TieredCache.GetWithLocalTimeout(getQueueCacheKey(), time.Minute, wanted); err == nil {
		return wanted.(*types.QueuesEstimate)
	} else {
		logger.Errorf("error retrieving mempool data from cache: %v", err)
	}
	return wanted
}

func getQueueCacheKey() string {
	return fmt.Sprintf("%d:frontend:queues", utils.Config.Chain.ClConfig.DepositChainID)
}

func getQueuesEstimate() (*types.QueuesEstimate, error) {
	queue, err := rpc.CurrentClient.GetValidatorQueue()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get validator state")
	}

	type Result struct {
		EstClearEpoch         uint64 `db:"est_clear_epoch"`
		QueuedBalanceAhead    uint64 `db:"queued_balance_ahead"`
		TopUpAmount           uint64 `db:"topup_amount"`
		TopUpCount            uint64 `db:"topup_count"`
		EnteringNewValidators uint64 `db:"entering_new_validators"`
		TotalEffectiveBalance uint64 `db:"eligibleether"`
	}
	var result Result

	err = db.WriterDb.Get(&result,
		`
			WITH 
				last_deposit AS (
					SELECT * 
					FROM pending_deposits_queue 
					WHERE id = (SELECT max(id) FROM pending_deposits_queue)
				),
				deduplicated AS (
					SELECT DISTINCT pubkey, validator_index
					FROM pending_deposits_queue 
				)
			SELECT 
				COALESCE((SELECT est_clear_epoch FROM last_deposit),0) AS est_clear_epoch, 
				COALESCE((SELECT queued_balance_ahead FROM last_deposit),0) AS queued_balance_ahead, 
				(SELECT COALESCE(count(*),0) FROM deduplicated WHERE validator_index IS NULL) AS entering_new_validators, 
				(SELECT COALESCE(count(*),0) FROM deduplicated WHERE validator_index IS NOT NULL) AS topup_count, 
				(SELECT COALESCE(sum(amount),0) FROM pending_deposits_queue WHERE validator_index IS NOT NULL) AS topup_amount,
				(SELECT eligibleether FROM epochs WHERE epoch = (SELECT max(epoch) FROM epochs)) 
			
		`)
	if err != nil {
		logger.Errorf("error getting pending queue data: %v", err)
	}

	depositQueueTime := time.Until(utils.EpochToTime(result.EstClearEpoch))
	etherChurnByEpoch := utils.GetActivationExitChurnLimit(result.TotalEffectiveBalance)
	if etherChurnByEpoch == 0 {
		return nil, errors.New("etherChurnByEpoch is 0")
	}

	etherChurnByDay := etherChurnByEpoch * utils.EpochsPerDay()
	re := &types.QueuesEstimate{
		EnteringNewValidatorsCount:     result.EnteringNewValidators,
		EnteringNewValidatorsEthAmount: max(result.QueuedBalanceAhead-result.TopUpAmount, 0),
		EnteringTopUpEthAmount:         result.TopUpAmount,
		EnteringTotalEthAmount:         result.QueuedBalanceAhead,
		EnteringQueueTime:              depositQueueTime,
		EnteringTopUpCount:             result.TopUpCount,
		TotalActiveEffectiveBalance:    result.TotalEffectiveBalance,
		LeavingValidatorCount:          queue.Exiting,
		LeavingEthAmount:               queue.ExitingBalance,
		EnteringBalancePerDay:          etherChurnByDay,
		EnteringBalancePerEpoch:        etherChurnByEpoch,
	}

	return re, nil
}
