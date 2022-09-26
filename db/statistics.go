package db

import (
	"eth2-exporter/metrics"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

func WriteStatisticsForDay(day uint64) error {
	exportStart := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_validator_stats").Observe(time.Since(exportStart).Seconds())
	}()

	epochsPerDay := (24 * 60 * 60) / utils.Config.Chain.Config.SlotsPerEpoch / utils.Config.Chain.Config.SecondsPerSlot
	firstEpoch := day * epochsPerDay
	lastEpoch := (day+1)*epochsPerDay - 1
	// firstSlot := firstEpoch * utils.Config.Chain.Config.SlotsPerEpoch
	// lastSlot := (lastEpoch+1)*utils.Config.Chain.Config.SlotsPerEpoch - 1

	logger.Infof("exporting statistics for day %v (epoch %v to %v)", day, firstEpoch, lastEpoch)

	latestDbEpoch, err := GetLatestEpoch()
	if err != nil {
		return err
	}

	if lastEpoch > latestDbEpoch {
		return fmt.Errorf("delaying statistics export as epoch %v has not yet been indexed", lastEpoch)
	}

	start := time.Now()
	logger.Infof("exporting min_balance, max_balance, min_effective_balance, max_effective_balance, start_balance, start_effective_balance, end_balance and end_effective_balance statistics")
	balanceStatistics, err := BigtableClient.GetValidatorBalanceStatistics(firstEpoch, lastEpoch)
	if err != nil {
		return err
	}

	balanceStatsArr := make([]*types.ValidatorBalanceStatistic, 0, len(balanceStatistics))
	for _, stat := range balanceStatistics {
		balanceStatsArr = append(balanceStatsArr, stat)
	}
	tx, err := WriterDb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	batchSize := 6500 // max parameters: 65535
	for b := 0; b < len(balanceStatsArr); b += batchSize {
		start := b
		end := b + batchSize
		if len(balanceStatsArr) < end {
			end = len(balanceStatsArr)
		}

		numArgs := 10
		valueStrings := make([]string, 0, batchSize)
		valueArgs := make([]interface{}, 0, batchSize*numArgs)
		for i, stat := range balanceStatsArr[start:end] {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*numArgs+1, i*numArgs+2, i*numArgs+3, i*numArgs+4, i*numArgs+5, i*numArgs+6, i*numArgs+7, i*numArgs+8, i*numArgs+9, i*numArgs+10))
			valueArgs = append(valueArgs, stat.Index)
			valueArgs = append(valueArgs, day)
			valueArgs = append(valueArgs, stat.MinBalance)
			valueArgs = append(valueArgs, stat.MaxBalance)
			valueArgs = append(valueArgs, stat.MinEffectiveBalance)
			valueArgs = append(valueArgs, stat.MaxEffectiveBalance)
			valueArgs = append(valueArgs, stat.StartBalance)
			valueArgs = append(valueArgs, stat.StartEffectiveBalance)
			valueArgs = append(valueArgs, stat.EndBalance)
			valueArgs = append(valueArgs, stat.EndEffectiveBalance)
		}
		stmt := fmt.Sprintf(`
		insert into validator_stats (validatorindex, day, min_balance, max_balance, min_effective_balance, max_effective_balance, start_balance, start_effective_balance, end_balance, end_effective_balance) VALUES
		%s
		on conflict (validatorindex, day) do update set min_balance = excluded.min_balance, max_balance = excluded.max_balance, min_effective_balance = excluded.min_effective_balance, max_effective_balance = excluded.max_effective_balance, start_balance = excluded.start_balance, start_effective_balance = excluded.start_effective_balance, end_balance = excluded.end_balance, end_effective_balance = excluded.end_effective_balance;`,
			strings.Join(valueStrings, ","))
		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			return err
		}

		logger.Infof("saving validator balance batch %v completed", b)
	}
	logger.Infof("export completed, took %v", time.Since(start))

	start = time.Now()
	logger.Infof("exporting missed_attestations statistics")
	ma, err := BigtableClient.GetValidatorMissedAttestationsCount([]uint64{}, lastEpoch, lastEpoch-firstEpoch)
	if err != nil {
		return err
	}
	maArr := make([]*types.ValidatorMissedAttestationsStatistic, 0, len(ma))
	for _, stat := range ma {
		maArr = append(maArr, stat)
	}

	batchSize = 16000 // max parameters: 65535
	for b := 0; b < len(maArr); b += batchSize {
		start := b
		end := b + batchSize
		if len(maArr) < end {
			end = len(maArr)
		}

		numArgs := 4
		valueStrings := make([]string, 0, batchSize)
		valueArgs := make([]interface{}, 0, batchSize*numArgs)
		for i, stat := range maArr[start:end] {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", i*numArgs+1, i*numArgs+2, i*numArgs+3, i*numArgs+4))
			valueArgs = append(valueArgs, stat.Index)
			valueArgs = append(valueArgs, day)
			valueArgs = append(valueArgs, stat.MissedAttestations)
			valueArgs = append(valueArgs, 0)
		}
		stmt := fmt.Sprintf(`
		insert into validator_stats (validatorindex, day, missed_attestations, orphaned_attestations) VALUES
		%s
		on conflict (validatorindex, day) do update set missed_attestations = excluded.missed_attestations, orphaned_attestations = excluded.orphaned_attestations;`,
			strings.Join(valueStrings, ","))
		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			return err
		}

		logger.Infof("saving missed attestations batch %v completed", b)
	}

	logger.Infof("export completed, took %v", time.Since(start))

	start = time.Now()
	logger.Infof("exporting sync statistics")
	syncStats, err := BigtableClient.GetValidatorSyncDutiesStatistics([]uint64{}, lastEpoch, int64(lastEpoch-firstEpoch))
	if err != nil {
		return err
	}
	syncStatsArr := make([]*types.ValidatorSyncDutiesStatistic, 0, len(syncStats))
	for _, stat := range syncStats {
		syncStatsArr = append(syncStatsArr, stat)
	}

	batchSize = 13000 // max parameters: 65535
	for b := 0; b < len(syncStatsArr); b += batchSize {
		start := b
		end := b + batchSize
		if len(syncStatsArr) < end {
			end = len(syncStatsArr)
		}

		numArgs := 5
		valueStrings := make([]string, 0, batchSize)
		valueArgs := make([]interface{}, 0, batchSize*numArgs)
		for i, stat := range syncStatsArr[start:end] {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", i*numArgs+1, i*numArgs+2, i*numArgs+3, i*numArgs+4, i*numArgs+5))
			valueArgs = append(valueArgs, stat.Index)
			valueArgs = append(valueArgs, day)
			valueArgs = append(valueArgs, stat.ParticipatedSync)
			valueArgs = append(valueArgs, stat.MissedSync)
			valueArgs = append(valueArgs, 0)
		}
		stmt := fmt.Sprintf(`
		insert into validator_stats (validatorindex, day, participated_sync, missed_sync, orphaned_sync)  VALUES
		%s
		on conflict (validatorindex, day) do update set participated_sync = excluded.participated_sync, missed_sync = excluded.missed_sync, orphaned_sync = excluded.orphaned_sync;`,
			strings.Join(valueStrings, ","))
		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			return err
		}

		logger.Infof("saving sync statistics batch %v completed", b)
	}
	logger.Infof("export completed, took %v", time.Since(start))

	start = time.Now()
	logger.Infof("exporting proposed_blocks, missed_blocks and orphaned_blocks statistics")
	_, err = tx.Exec(`
		insert into validator_stats (validatorindex, day, proposed_blocks, missed_blocks, orphaned_blocks) 
		(
			select proposer, $3, sum(case when status = '1' then 1 else 0 end), sum(case when status = '2' then 1 else 0 end), sum(case when status = '3' then 1 else 0 end)
			from blocks
			where epoch >= $1 and epoch <= $2 and status = '1'
			group by proposer
		) 
		on conflict (validatorindex, day) do update set proposed_blocks = excluded.proposed_blocks, missed_blocks = excluded.missed_blocks, orphaned_blocks = excluded.orphaned_blocks;`,
		firstEpoch, lastEpoch, day)
	if err != nil {
		return err
	}
	logger.Infof("export completed, took %v", time.Since(start))

	start = time.Now()
	logger.Infof("exporting attester_slashings and proposer_slashings statistics")
	_, err = tx.Exec(`
		insert into validator_stats (validatorindex, day, attester_slashings, proposer_slashings) 
		(
			select proposer, $3, sum(attesterslashingscount), sum(proposerslashingscount)
			from blocks
			where epoch >= $1 and epoch <= $2 and status = '1'
			group by proposer
		) 
		on conflict (validatorindex, day) do update set attester_slashings = excluded.attester_slashings, proposer_slashings = excluded.proposer_slashings;`,
		firstEpoch, lastEpoch, day)
	if err != nil {
		return err
	}
	logger.Infof("export completed, took %v", time.Since(start))

	start = time.Now()
	logger.Infof("exporting deposits and deposits_amount statistics")
	depositsQry := `
		insert into validator_stats (validatorindex, day, deposits, deposits_amount) 
		(
			select validators.validatorindex, $3, count(*), sum(amount)
			from blocks_deposits
			inner join validators on blocks_deposits.publickey = validators.pubkey
			inner join blocks on blocks_deposits.block_root = blocks.blockroot
			where block_slot >= $1 * 32 and block_slot <= $2 * 32 and blocks.status = '1'
			group by validators.validatorindex
		) 
		on conflict (validatorindex, day) do
			update set deposits = excluded.deposits, 
			deposits_amount = excluded.deposits_amount;`
	if day == 0 {
		// genesis-deposits will be added to block 0 by the exporter which is technically not 100% correct
		// since deposits will be added to the validator-balance only after the block which includes the deposits.
		// to ease the calculation of validator-income (considering deposits) we set the day of genesis-deposits to -1.
		depositsQry = `
			insert into validator_stats (validatorindex, day, deposits, deposits_amount)
			(
				select validators.validatorindex, case when block_slot = 0 then -1 else $3 end as day, count(*), sum(amount)
				from blocks_deposits
				inner join validators on blocks_deposits.publickey = validators.pubkey
				where block_slot >= $1 * 32 and block_slot <= $2 * 32 and status = '1'
				group by validators.validatorindex, day
			) 
			on conflict (validatorindex, day) do
				update set deposits = excluded.deposits, 
				deposits_amount = excluded.deposits_amount;`
		if err != nil {
			return err
		}
	}
	_, err = tx.Exec(depositsQry, firstEpoch, lastEpoch, day)
	if err != nil {
		return err
	}
	logger.Infof("export completed, took %v", time.Since(start))

	start = time.Now()
	logger.Infof("marking day export as completed in the status table")
	_, err = tx.Exec("insert into validator_stats_status (day, status) values ($1, true)", day)
	if err != nil {
		return err
	}
	logger.Infof("export completed, took %v", time.Since(start))

	err = tx.Commit()
	if err != nil {
		return err
	}

	logger.Infof("statistics export of day %v completed, took %v", day, time.Since(exportStart))
	return nil
}
func GetValidatorIncomeHistoryChart(validator_indices []uint64, currency string) ([]*types.ChartDataPoint, error) {
	incomeHistory, err := GetValidatorIncomeHistory(validator_indices, 0, 0)
	if err != nil {
		return nil, err
	}
	var chartData = make([]*types.ChartDataPoint, len(incomeHistory))

	for i := 0; i < len(incomeHistory); i++ {
		color := "#7cb5ec"
		if incomeHistory[i].Income < 0 {
			color = "#f7a35c"
		}
		balanceTs := utils.DayToTime(incomeHistory[i].Day)
		chartData[i] = &types.ChartDataPoint{X: float64(balanceTs.Unix() * 1000), Y: utils.ExchangeRateForCurrency(currency) * (float64(incomeHistory[i].Income) / 1000000000), Color: color}
	}
	return chartData, err
}

func GetValidatorIncomeHistory(validator_indices []uint64, lowerBoundDay uint64, upperBoundDay uint64) ([]types.ValidatorIncomeHistory, error) {
	if upperBoundDay == 0 {
		upperBoundDay = 65536
	}
	queryValidatorsArr := pq.Array(validator_indices)

	var result []types.ValidatorIncomeHistory
	err := ReaderDb.Select(&result, `
	with _today as (
		select max(day) + 1 as day from validator_stats
	),
	current_balances as (
		select coalesce(sum(balance)) as end_balance, (select day from _today) as day, array_agg(pubkey) as pubkeys
		from validators
		where validatorindex = ANY($1)
	),
	current_deposits as (
		select
			coalesce(SUM(amount),0) as deposits_amount,
			(select day from _today) as day
		from blocks_deposits
		where 
			block_slot > (select (day) * 225 * 32 from _today) and
			publickey in (select pubkey
				from validators
				where validatorindex = ANY($1))
	),
	history as (
		select day, coalesce(lag(end_balance) over (order by day), start_balance) as start_balance, end_balance as end_balance, deposits_amount
		from (
			select 
				day, COALESCE(SUM(start_balance),0) AS start_balance, COALESCE(SUM(end_balance),0) AS end_balance, COALESCE(SUM(deposits_amount), 0) AS deposits_amount
			FROM validator_stats
			WHERE validatorindex = ANY($1) AND
				day BETWEEN ($2 - 1) AND $3
			GROUP BY day
			ORDER BY day
		) as foo order by day
	),
	today as (
		select 
			(select day from _today), COALESCE(SUM(end_balance),0) AS start_balance
		FROM validator_stats
		WHERE validatorindex = ANY($1) and day=(select day from _today) - 1
		GROUP BY day
	)
	select * from (
		select day, end_balance - start_balance - deposits_amount as diff, start_balance, end_balance, deposits_amount from (
			select 
				coalesce(history.day, 0) + coalesce(current_balances.day, 0) as day,
				coalesce(history.start_balance,0) + coalesce(today.start_balance,0) as start_balance,
				coalesce(history.end_balance,0) + coalesce(current_balances.end_balance,0) as end_balance,
				coalesce(history.deposits_amount,0) + coalesce(current_deposits.deposits_amount,0) as deposits_amount
			from history
			full outer join current_balances on current_balances.day = history.day
			left join current_deposits on current_balances.day = current_deposits.day
			full join today on current_balances.day = today.day
		) as foo 
	) as foo2 
	where diff <> 0  AND
		day BETWEEN $2 AND $3
	order by day;`, queryValidatorsArr, lowerBoundDay, upperBoundDay)
	return result, err
}
