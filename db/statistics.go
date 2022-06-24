package db

import (
	"eth2-exporter/metrics"
	"eth2-exporter/utils"
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
	firstSlot := firstEpoch * utils.Config.Chain.Config.SlotsPerEpoch
	lastSlot := (lastEpoch+1)*utils.Config.Chain.Config.SlotsPerEpoch - 1

	logger.Infof("exporting statistics for day %v (epoch %v to %v)", day, firstEpoch, lastEpoch)

	tx, err := WriterDb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	start := time.Now()
	logger.Infof("exporting min_balance, max_balance, min_effective_balance, max_effective_balance, start_balance, start_effective_balance, end_balance and end_effective_balance statistics")
	_, err = tx.Exec(`
		insert into validator_stats (validatorindex, day, min_balance, max_balance, min_effective_balance, max_effective_balance, start_balance, start_effective_balance, end_balance, end_effective_balance)
		(
			select validatorindex, $3, min(balance), max(balance), min(effectivebalance), max(effectivebalance), max(case when epoch = $1 then balance else 0 end), max(case when epoch = $1 then effectivebalance else 0 end), max(case when epoch = $2 then balance else 0 end), max(case when epoch = $2 then effectivebalance else 0 end) 
			from validator_balances_p 
			where week >= $1 / 1575 AND week <= $2 / 1575 and epoch >= $1 and epoch <= $2
			group by validatorindex
		) 
		on conflict (validatorindex, day) do update set min_balance = excluded.min_balance, max_balance = excluded.max_balance, min_effective_balance = excluded.min_effective_balance, max_effective_balance = excluded.max_effective_balance, start_balance = excluded.start_balance, start_effective_balance = excluded.start_effective_balance, end_balance = excluded.end_balance, end_effective_balance = excluded.end_effective_balance;`,
		firstEpoch, lastEpoch, day)
	if err != nil {
		return err
	}
	logger.Infof("export completed, took %v", time.Since(start))

	start = time.Now()
	logger.Infof("exporting missed_attestations and orphaned_attestations statistics")
	_, err = tx.Exec(`
		insert into validator_stats (validatorindex, day, missed_attestations, orphaned_attestations) 
		(
			select validatorindex, $3, sum(case when status = 0 then 1 else 0 end), sum(case when status = 3 then 1 else 0 end)
			from attestation_assignments_p
			where week >= $1 / 1575 AND week <= $2 / 1575 and epoch >= $1 and epoch <= $2
			group by validatorindex
		) 
		on conflict (validatorindex, day) do update set missed_attestations = excluded.missed_attestations, orphaned_attestations = excluded.orphaned_attestations;`,
		firstEpoch, lastEpoch, day)
	if err != nil {
		return err
	}
	logger.Infof("export completed, took %v", time.Since(start))

	start = time.Now()
	logger.Infof("exporting sync statistics")
	_, err = tx.Exec(`
		insert into validator_stats (validatorindex, day, participated_sync, missed_sync, orphaned_sync) 
		(
			select validatorindex, $3, sum(case when status = 1 then 1 else 0 end), sum(case when status = 2 then 1 else 0 end), sum(case when status = 3 then 1 else 0 end)
			from sync_assignments_p
			where week >= $1 / 1575 AND week <= $2 / 1575 and slot >= $1 and slot <= $2
			group by validatorindex
		) 
		on conflict (validatorindex, day) do update set participated_sync = excluded.participated_sync, missed_sync = excluded.missed_sync, orphaned_sync = excluded.orphaned_sync;`,
		firstSlot, lastSlot, day)
	if err != nil {
		return err
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
			where block_slot >= $1 * 32 and block_slot <= $2 * 32 and status = '1'
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
