package exporter

import (
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"sort"
	"time"
)

func performanceDataUpdater() {
	for true {
		time.Sleep(time.Hour)
		logger.Info("updating validator performance data")
		start := time.Now()
		err := updateValidatorPerformance()
		if err != nil {
			logger.WithError(err).Errorf("error updating validator performance data")
		} else {
			logger.WithField("duration", time.Since(start)).Info("validator performance data update completed")
		}
	}
}

func updateValidatorPerformance() error {
	tx, err := db.DB.Beginx()
	if err != nil {
		return fmt.Errorf("error starting db transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec("TRUNCATE validator_performance")
	if err != nil {
		return fmt.Errorf("error truncating validator performance table: %w", err)
	}

	var currentEpoch uint64

	err = tx.Get(&currentEpoch, "SELECT MAX(epoch) FROM validator_balances")
	if err != nil {
		return fmt.Errorf("error retrieving latest epoch from validator_balances table: %w", err)
	}

	now := utils.EpochToTime(currentEpoch)
	epoch1d := utils.TimeToEpoch(now.Add(time.Hour * 24 * -1))
	epoch7d := utils.TimeToEpoch(now.Add(time.Hour * 24 * 7 * -1))
	epoch31d := utils.TimeToEpoch(now.Add(time.Hour * 24 * 31 * -1))
	epoch365d := utils.TimeToEpoch(now.Add(time.Hour * 24 * 356 * -1))

	if epoch1d < 0 {
		epoch1d = 0
	}
	if epoch7d < 0 {
		epoch7d = 0
	}
	if epoch31d < 0 {
		epoch31d = 0
	}
	if epoch365d < 0 {
		epoch365d = 0
	}

	var startBalances []struct {
		Index           uint64
		Balance         uint64
		Activationepoch int64
	}
	err = tx.Select(&startBalances, `
		SELECT 
			validator_balances_historical.validatorindex as index,
			validator_balances_historical.balance,
			validators.activationepoch
		FROM validators
			LEFT JOIN validator_balances
				ON validators.activationepoch = validator_balances_historical.epoch
				AND validators.validatorindex = validator_balances_historical.validatorindex
		WHERE validator_balances_historical.validatorindex IS NOT NULL`)
	if err != nil {
		return fmt.Errorf("error retrieving initial validator balances data: %w", err)
	}

	startEpochMap := make(map[uint64]int64)
	startBalanceMap := make(map[uint64]uint64)
	for _, balance := range startBalances {
		startEpochMap[balance.Index] = balance.Activationepoch
		startBalanceMap[balance.Index] = balance.Balance
	}

	var balances []*types.ValidatorBalance
	err = tx.Select(&balances, `
		SELECT
			validator_balances_historical.epoch,
			validator_balances_historical.validatorindex,
			validator_balances_historical.balance
		FROM validator_balances_historical
		WHERE validator_balances_historical.epoch IN ($1, $2, $3, $4, $5)`,
		currentEpoch, epoch1d, epoch7d, epoch31d, epoch365d)
	if err != nil {
		return fmt.Errorf("error retrieving validator performance data: %w", err)
	}

	performance := make(map[uint64]map[int64]int64)
	for _, balance := range balances {
		if performance[balance.Index] == nil {
			performance[balance.Index] = make(map[int64]int64)
		}
		performance[balance.Index][int64(balance.Epoch)] = int64(balance.Balance)
	}

	deposits := []struct {
		Validatorindex uint64
		Epoch          int64
		Amount         int64
	}{}

	err = tx.Select(&deposits, `
		SELECT
			v.validatorindex,
			(d.block_slot/32) AS epoch,
			SUM(d.amount) AS amount
		FROM validators v
			INNER JOIN blocks_deposits d
				ON d.publickey = v.pubkey
				AND (d.block_slot/32) > v.activationepoch
		GROUP BY (d.block_slot/32), v.validatorindex
		ORDER BY epoch`)
	if err != nil {
		return fmt.Errorf("error retrieving validator deposits data: %w", err)
	}

	depositsMap := make(map[uint64]map[int64]int64)
	for _, d := range deposits {
		if _, exists := depositsMap[d.Validatorindex]; !exists {
			depositsMap[d.Validatorindex] = make(map[int64]int64)
		}
		depositsMap[d.Validatorindex][d.Epoch] = d.Amount
	}

	data := make([]*types.ValidatorPerformance, 0, len(performance))

	for validator, balances := range performance {

		currentBalance := balances[int64(currentEpoch)]
		startBalance := int64(startBalanceMap[validator])

		if currentBalance == 0 || startBalance == 0 {
			continue
		}

		balance1d := balances[epoch1d]
		if balance1d == 0 || startEpochMap[validator] > epoch1d {
			balance1d = startBalance
		}
		balance7d := balances[epoch7d]
		if balance7d == 0 || startEpochMap[validator] > epoch7d {
			balance7d = startBalance
		}
		balance31d := balances[epoch31d]
		if balance31d == 0 || startEpochMap[validator] > epoch31d {
			balance31d = startBalance
		}
		balance365d := balances[epoch365d]
		if balance365d == 0 || startEpochMap[validator] > epoch365d {
			balance365d = startBalance
		}

		performance1d := currentBalance - balance1d
		performance7d := currentBalance - balance7d
		performance31d := currentBalance - balance31d
		performance365d := currentBalance - balance365d

		if depositsMap[validator] != nil {
			for depositEpoch, depositAmount := range depositsMap[validator] {
				if depositEpoch > epoch1d {
					performance1d -= depositAmount
				}
				if depositEpoch > epoch7d {
					performance7d -= depositAmount
				}
				if depositEpoch > epoch31d {
					performance31d -= depositAmount
				}
				if depositEpoch > epoch365d {
					performance365d -= depositAmount
				}
			}
		}

		data = append(data, &types.ValidatorPerformance{
			Rank:            0,
			Index:           validator,
			PublicKey:       nil,
			Name:            "",
			Balance:         uint64(currentBalance),
			Performance1d:   performance1d,
			Performance7d:   performance7d,
			Performance31d:  performance31d,
			Performance365d: performance365d,
		})
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].Performance7d > data[j].Performance7d
	})

	for i, d := range data {
		_, err := tx.Exec(`
			INSERT INTO validator_performance (validatorindex, balance, performance1d, performance7d, performance31d, performance365d, rank7d)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			d.Index, d.Balance, d.Performance1d, d.Performance7d, d.Performance31d, d.Performance365d, i+1)

		if err != nil {
			return fmt.Errorf("error saving validator performance data: %w", err)
		}
	}

	return tx.Commit()
}
