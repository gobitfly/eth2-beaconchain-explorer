package services

import (
	"eth2-exporter/db"
	"eth2-exporter/types"
	"time"
)

func statsUpdater() {
	sleepDuration := time.Duration(1 * time.Minute)
	var prevEpoch uint64

	for {
		latestEpoch := LatestEpoch()
		if prevEpoch >= latestEpoch {
			time.Sleep(sleepDuration)
			continue
		}
		now := time.Now()
		statResult, err := calculateStats()
		if err != nil {
			logger.WithField("epoch", latestEpoch).Errorf("error updating stats: %v", err)
			time.Sleep(sleepDuration)
			continue
		}
		logger.WithField("epoch", latestEpoch).WithField("duration", time.Since(now)).Info("stats update completed")
		latestStats.Store(statResult)
		prevEpoch = latestEpoch
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

	return &stats, nil
}

func eth1TopDepositers() (*[]types.StatsTopDepositors, error) {
	topDepositors := []types.StatsTopDepositors{}

	err := db.DB.Select(&topDepositors, `
	SELECT 
		from_address, 
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

	err := db.DB.Get(&count, `
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

	err := db.DB.Get(&count, `
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
