package services

import (
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"time"
)

func statsUpdater() {
	sleepDuration := time.Duration(time.Minute)

	for {
		latestEpoch := LatestEpoch()

		now := time.Now()
		statResult, err := calculateStats()
		if err != nil {
			logger.WithField("epoch", latestEpoch).Errorf("error updating stats: %v", err)
			time.Sleep(sleepDuration)
			continue
		}
		logger.WithField("epoch", latestEpoch).WithField("duration", time.Since(now)).Info("stats update completed")
		latestStats.Store(statResult)
		time.Sleep(sleepDuration)
	}
}

func calculateStats() (*types.Stats, error) {
	stats := types.Stats{}

	topDeposits, err := eth1TopDepositers()
	if err != nil {
		return nil, err
	}
	stats.TopDepositors = topDeposits
	invalidCount, err := eth1InvalidDeposits()
	if err != nil {
		return nil, err
	}
	stats.InvalidDepositCount = invalidCount

	uniqueValidatorCount, err := eth1UniqueValidatorsCount()
	if err != nil {
		return nil, err
	}
	stats.UniqueValidatorCount = uniqueValidatorCount

	totalValidatorCount, err := db.GetTotalValidatorsCount()
	if err != nil {
		logger.WithError(err).Error("error getting total validator count")
	}
	stats.TotalValidatorCount = &totalValidatorCount

	activeValidatorCount, err := db.GetActiveValidatorCount()
	if err != nil {
		logger.WithError(err).Error("error getting active validator count")
	}

	stats.ActiveValidatorCount = &activeValidatorCount

	pendingValidatorCount, err := db.GetPendingValidatorCount()
	if err != nil {
		logger.WithError(err).Error("error getting pending validator count")
	}

	stats.PendingValidatorCount = &pendingValidatorCount

	validatorChurnLimit, err := GetValidatorChurnLimit()
	if err != nil {
		logger.WithError(err).Error("error getting total validator churn limit")
	}

	stats.ValidatorChurnLimit = &validatorChurnLimit

	return &stats, nil
}

func eth1TopDepositers() (*[]types.StatsTopDepositors, error) {
	topDepositors := []types.StatsTopDepositors{}

	err := db.WriterDb.Select(&topDepositors, `
	SELECT 
		ENCODE(from_address::bytea, 'hex') as from_address, 
		count(from_address) as count 
	FROM eth1_deposits
	where valid_signature = true 
	GROUP BY 
		from_address
	ORDER BY count DESC LIMIT 5;
	`)
	if err != nil {
		return nil, err
	}

	return &topDepositors, nil
}

func eth1InvalidDeposits() (*uint64, error) {
	count := uint64(0)

	err := db.WriterDb.Get(&count, `
	SELECT 
		count(*) as count
	FROM eth1_deposits
	WHERE
	  valid_signature = false
	`)
	if err != nil {
		return nil, err
	}

	return &count, nil
}

func eth1UniqueValidatorsCount() (*uint64, error) {
	count := uint64(0)

	err := db.WriterDb.Get(&count, `
	SELECT 
		count(*) as count
	FROM 
	(
		SELECT 
			publickey, 
			sum(amount) 
		FROM 
			eth1_deposits 
		WHERE 
			valid_signature = true 
		GROUP BY 
			publickey 
		HAVING sum(amount) >= 32e9
	) as q;
	`)
	if err != nil {
		return nil, err
	}

	return &count, nil
}

// GetValidatorChurnLimit returns the rate at which validators can enter or leave the system
func GetValidatorChurnLimit() (uint64, error) {
	min := utils.Config.Chain.Config.MinPerEpochChurnLimit

	stats := GetLatestStats()
	count := stats.ActiveValidatorCount

	if count == nil {
		count = new(uint64)
	}

	adaptable := uint64(0)
	if *count > 0 {
		adaptable = utils.Config.Chain.Config.ChurnLimitQuotient / *count
	}

	if min > adaptable {
		return min, nil
	}

	return adaptable, nil
}
