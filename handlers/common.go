package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"net/http"
	"regexp"
	"time"

	"github.com/lib/pq"
)

var pkeyRegex = regexp.MustCompile("[^0-9A-Fa-f]+")

func GetValidatorOnlineThresholdSlot() uint64 {
	latestProposedSlot := services.LatestProposedSlot()
	threshold := utils.Config.Chain.SlotsPerEpoch * 2

	var validatorOnlineThresholdSlot uint64
	if latestProposedSlot < 1 || latestProposedSlot < threshold {
		validatorOnlineThresholdSlot = 0
	} else {
		validatorOnlineThresholdSlot = latestProposedSlot - threshold
	}

	return validatorOnlineThresholdSlot
}

// GetValidatorEarnings will return the earnings (last day, week, month and total) of selected validators
func GetValidatorEarnings(validators []uint64) (*types.ValidatorEarnings, error) {
	validatorsPQArray := pq.Array(validators)
	latestEpoch := services.LatestEpoch()
	now := utils.EpochToTime(latestEpoch)
	lastDayEpoch := uint64(utils.TimeToEpoch(now.Add(time.Hour * 24 * 1 * -1)))
	lastWeekEpoch := uint64(utils.TimeToEpoch(now.Add(time.Hour * 24 * 7 * -1)))
	lastMonthEpoch := uint64(utils.TimeToEpoch(now.Add(time.Hour * 24 * 31 * -1)))

	var activationEpoch uint64
	err := db.DB.Get(&activationEpoch, "SELECT CAST(MIN(activationepoch) AS BIGINT) FROM validators WHERE validatorindex = ANY($1)", validatorsPQArray)
	if err != nil {
		return nil, err
	}

	if activationEpoch == 9223372036854775807 {
		activationEpoch = 0
	}

	balances := []struct {
		Epoch   uint64
		Balance int64
	}{}

	err = db.DB.Select(&balances, "SELECT epoch, COALESCE(SUM(balance), 0) AS balance FROM validator_balances WHERE epoch = ANY($1) AND validatorindex = ANY($2) GROUP BY epoch", pq.Array([]uint64{latestEpoch, lastDayEpoch, lastWeekEpoch, lastMonthEpoch, activationEpoch}), validatorsPQArray)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	balancesEpochMap := make(map[uint64]int64)
	for _, b := range balances {
		balancesEpochMap[b.Epoch] = b.Balance
	}

	deposits := []struct {
		Epoch   uint64
		Deposit int64
	}{}

	err = db.DB.Select(&deposits, "SELECT block_slot / 32 AS epoch, SUM(amount) AS deposit FROM blocks_deposits WHERE publickey IN (SELECT pubkey FROM validators WHERE validatorindex = ANY($1)) GROUP BY epoch", validatorsPQArray)
	if err != nil {
		return nil, err
	}

	var earningsTotal int64
	var earningsLastDay int64
	var earningsLastWeek int64
	var earningsLastMonth int64

	// Calculate earnings
	start := activationEpoch
	end := latestEpoch
	initialBalance := balancesEpochMap[start]
	endBalance := balancesEpochMap[end]
	depositSum := int64(0)
	for _, d := range deposits {
		if d.Epoch > start && d.Epoch < end {
			depositSum += d.Deposit
		}
	}
	earningsTotal = endBalance - initialBalance - depositSum

	start = lastMonthEpoch
	initialBalance = balancesEpochMap[start]
	endBalance = balancesEpochMap[end]
	depositSum = int64(0)
	for _, d := range deposits {
		if d.Epoch > start && d.Epoch < end {
			depositSum += d.Deposit
		}
	}
	earningsLastMonth = endBalance - initialBalance - depositSum

	start = lastWeekEpoch
	initialBalance = balancesEpochMap[start]
	endBalance = balancesEpochMap[end]
	depositSum = int64(0)
	for _, d := range deposits {
		if d.Epoch > start && d.Epoch < end {
			depositSum += d.Deposit
		}
	}
	earningsLastWeek = endBalance - initialBalance - depositSum

	start = lastDayEpoch
	initialBalance = balancesEpochMap[start]
	endBalance = balancesEpochMap[end]
	depositSum = int64(0)
	for _, d := range deposits {
		if d.Epoch > start && d.Epoch < end {
			depositSum += d.Deposit
		}
	}
	earningsLastDay = endBalance - initialBalance - depositSum

	return &types.ValidatorEarnings{
		Total:     earningsTotal,
		LastDay:   earningsLastDay,
		LastWeek:  earningsLastWeek,
		LastMonth: earningsLastMonth,
	}, nil
}

// LatestState will return common information that about the current state of the eth2 chain
func LatestState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(services.LatestState())

	if err != nil {
		logger.Errorf("error sending latest index page data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

func GetCurrency(r *http.Request) string {
	if langCookie, err := r.Cookie("currency"); err == nil {
		return langCookie.Value
	}

	return "ETH"
}
