package db

import (
	"eth2-exporter/metrics"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"strings"
	"time"
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
	ma, err := BigtableClient.GetValidatorMissedAttestationsCount([]uint64{}, lastEpoch, int64(lastEpoch-firstEpoch))
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
	syncStatsArr := make([]*types.ValidatorSyncDutiesStatistic, 0, len(ma))
	for _, stat := range syncStats {
		syncStatsArr = append(syncStatsArr, stat)
	}

	batchSize = 13000 // max parameters: 65535
	for b := 0; b < len(maArr); b += batchSize {
		start := b
		end := b + batchSize
		if len(maArr) < end {
			end = len(maArr)
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
