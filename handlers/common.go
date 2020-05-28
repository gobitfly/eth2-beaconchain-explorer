package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"sync"
	"time"

	"github.com/lib/pq"
)

func GetValidatorOnlineThresholdSlot() uint64 {
	latestProposedSlot := services.LatestProposedSlot()
	var validatorOnlineThresholdSlot uint64
	if latestProposedSlot < 1 {
		validatorOnlineThresholdSlot = 0
	} else {
		validatorOnlineThresholdSlot = latestProposedSlot - utils.Config.Chain.SlotsPerEpoch*2
	}

	return validatorOnlineThresholdSlot
}

// GetValidatorEarnings will return the earnings (last day, week, month and total) of selected validators
func GetValidatorEarnings(validators []uint64) (*types.ValidatorEarnings, error) {
	validatorsPQArray := pq.Array(validators)
	latestEpoch := services.LatestEpoch()
	now := utils.EpochToTime(latestEpoch)
	lastDayEpoch := utils.TimeToEpoch(now.Add(time.Hour * 24 * 1 * -1))
	lastWeekEpoch := utils.TimeToEpoch(now.Add(time.Hour * 24 * 7 * -1))
	lastMonthEpoch := utils.TimeToEpoch(now.Add(time.Hour * 24 * 31 * -1))

	query := `
		WITH 
			minmaxepoch AS (
				SELECT
					validatorindex,
					MIN(epoch) AS firstepoch,
					MAX(epoch) AS lastepoch
				FROM validator_balances
				WHERE validatorindex = ANY($1) AND epoch > $2
				GROUP by validatorindex
			),
			deposits AS (
				SELECT vv.validatorindex, COALESCE(SUM(bd.amount),0) AS amount
				FROM minmaxepoch
				INNER JOIN validators vv
					ON vv.validatorindex = minmaxepoch.validatorindex
				LEFT JOIN blocks_deposits bd 
					ON bd.publickey = vv.pubkey
					AND (bd.block_slot/32)-1 > minmaxepoch.firstepoch
				GROUP BY vv.validatorindex
			)
		SELECT
			SUM(last.balance - first.balance - d.amount) AS earnings
		FROM minmaxepoch
		INNER JOIN validator_balances first
			ON first.validatorindex = minmaxepoch.validatorindex
			AND first.epoch = minmaxepoch.firstepoch
		INNER JOIN validator_balances last
			ON last.validatorindex = minmaxepoch.validatorindex
			AND last.epoch = minmaxepoch.lastepoch
		LEFT JOIN deposits d ON d.validatorindex = minmaxepoch.validatorindex`

	var earningsTotal int64
	var earningsLastDay int64
	var earningsLastWeek int64
	var earningsLastMonth int64

	wg := sync.WaitGroup{}
	wg.Add(4)
	errs := make(chan error, 4)

	go func() {
		defer wg.Done()
		err := db.DB.Get(&earningsTotal, query, validatorsPQArray, 0)
		if err != nil {
			err = fmt.Errorf("error retrieving total earnings: %w", err)
		}
		errs <- err
	}()

	go func() {
		defer wg.Done()
		err := db.DB.Get(&earningsLastDay, query, validatorsPQArray, lastDayEpoch)
		if err != nil {
			err = fmt.Errorf("error retrieving earnings of last day: %w", err)
		}
		errs <- err
	}()

	go func() {
		defer wg.Done()
		err := db.DB.Get(&earningsLastWeek, query, validatorsPQArray, lastWeekEpoch)
		if err != nil {
			err = fmt.Errorf("error retrieving earnings of last week: %w", err)
		}
		errs <- err
	}()

	go func() {
		defer wg.Done()
		err := db.DB.Get(&earningsLastMonth, query, validatorsPQArray, lastMonthEpoch)
		if err != nil {
			err = fmt.Errorf("error retrieving earnings of last month: %w", err)
		}
		errs <- err
	}()

	go func() {
		wg.Wait()
		close(errs)
	}()

	for err := range errs {
		if err != nil {
			return nil, err
		}
	}

	earnings := &types.ValidatorEarnings{
		Total:     earningsTotal,
		LastDay:   earningsLastDay,
		LastWeek:  earningsLastWeek,
		LastMonth: earningsLastMonth,
	}

	return earnings, nil
}
