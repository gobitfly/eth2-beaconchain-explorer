package db

import (
	"context"
	"database/sql"
	"eth2-exporter/cache"
	"eth2-exporter/metrics"
	"eth2-exporter/price"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func WriteValidatorStatisticsForDay(day uint64, concurrencyTotal uint64, concurrencyCl uint64, concurrencyFailedAttestations uint64) error {
	exportStart := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_validator_stats").Observe(time.Since(exportStart).Seconds())
	}()

	firstEpoch, lastEpoch := utils.GetFirstAndLastEpochForDay(day)

	logger.Infof("exporting statistics for day %v (epoch %v to %v)", day, firstEpoch, lastEpoch)

	if err := checkIfDayIsFinalized(day); err != nil {
		return err
	}

	logger.Infof("getting exported state for day %v", day)
	start := time.Now()

	type Exported struct {
		Status              bool `db:"status"`
		FailedAttestations  bool `db:"failed_attestations_exported"`
		SyncDuties          bool `db:"sync_duties_exported"`
		WithdrawalsDeposits bool `db:"withdrawals_deposits_exported"`
		Balance             bool `db:"balance_exported"`
		ClRewards           bool `db:"cl_rewards_exported"`
		ElRewards           bool `db:"el_rewards_exported"`
		TotalPerformance    bool `db:"total_performance_exported"`
		BlockStats          bool `db:"block_stats_exported"`
		TotalAccumulation   bool `db:"total_accumulation_exported"`
	}
	exported := Exported{}

	err := ReaderDb.Get(&exported, `
		SELECT 
			status,
			failed_attestations_exported,
			sync_duties_exported,
			withdrawals_deposits_exported,
			balance_exported,
			cl_rewards_exported,
			el_rewards_exported,
			total_performance_exported,
			block_stats_exported,
			total_accumulation_exported
		FROM validator_stats_status 
		WHERE day = $1;
		`, day)

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("error retrieving exported state: %w", err)
	}
	logger.Infof("getting exported state took %v", time.Since(start))

	if exported.FailedAttestations && exported.SyncDuties && exported.WithdrawalsDeposits && exported.Balance && exported.ClRewards && exported.ElRewards && exported.TotalAccumulation && exported.TotalPerformance && exported.BlockStats && exported.Status {
		logger.Infof("Skipping day %v as it is already exported", day)
		return nil
	}

	if exported.FailedAttestations {
		logger.Infof("Skipping failed attestations")
	} else if err := WriteValidatorFailedAttestationsStatisticsForDay(day, concurrencyFailedAttestations); err != nil {
		return fmt.Errorf("error in WriteValidatorFailedAttestationsStatisticsForDay: %w", err)
	}

	if exported.SyncDuties {
		logger.Infof("Skipping sync duties")
	} else if err := WriteValidatorSyncDutiesForDay(day); err != nil {
		return fmt.Errorf("error in WriteValidatorSyncDutiesForDay: %w", err)
	}

	if exported.WithdrawalsDeposits {
		logger.Infof("Skipping withdrawals / deposits")
	} else if err := WriteValidatorDepositWithdrawals(day); err != nil {
		return fmt.Errorf("error in WriteValidatorDepositWithdrawals: %w", err)
	}

	if exported.BlockStats {
		logger.Infof("Skipping block stats")
	} else if err := WriteValidatorBlockStats(day); err != nil {
		return fmt.Errorf("error in WriteValidatorBlockStats: %w", err)
	}

	if exported.Balance {
		logger.Infof("Skipping balances")
	} else if err := WriteValidatorBalances(day); err != nil {
		return fmt.Errorf("error in WriteValidatorBalances: %w", err)
	}

	if exported.ClRewards {
		logger.Infof("Skipping cl rewards")
	} else if err := WriteValidatorClIcome(day, concurrencyCl); err != nil {
		return fmt.Errorf("error in WriteValidatorClIcome: %w", err)
	}

	if exported.ElRewards {
		logger.Infof("Skipping el rewards")
	} else if err := WriteValidatorElIcome(day); err != nil {
		return fmt.Errorf("error in WriteValidatorElIcome: %w", err)
	}

	if exported.TotalAccumulation {
		logger.Infof("Skipping total accumulation")
	} else if err := WriteValidatorTotalAccumulation(day, concurrencyTotal); err != nil {
		return fmt.Errorf("error in WriteValidatorTotalAccumulation: %w", err)
	}

	if exported.TotalPerformance {
		logger.Infof("Skipping total performance")
	} else if err := WriteValidatorTotalPerformance(day, concurrencyTotal); err != nil {
		return fmt.Errorf("error in WriteValidatorTotalPerformance: %w", err)
	}

	if err := WriteValidatorStatsExported(day); err != nil {
		return fmt.Errorf("error in WriteValidatorStatsExported: %w", err)
	}

	logger.Infof("statistics export of day %v completed, took %v", day, time.Since(exportStart))
	return nil
}

func WriteValidatorStatsExported(day uint64) error {
	tx, err := WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	start := time.Now()

	logger.Infof("marking day export as completed in the validator_stats_status table for day %v", day)
	_, err = tx.Exec(`
		UPDATE validator_stats_status
		SET status = true
		WHERE day=$1
		AND failed_attestations_exported = true
		AND sync_duties_exported = true
		AND withdrawals_deposits_exported = true
		AND balance_exported = true
		AND cl_rewards_exported = true
		AND el_rewards_exported = true
		AND total_performance_exported = true
		AND block_stats_exported = true
		AND total_accumulation_exported = true;
		`, day)
	if err != nil {
		return fmt.Errorf("error marking day export as completed in the validator_stats_status table for day %v: %w", day, err)
	}
	logger.Infof("marking completed, took %v", time.Since(start))

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func WriteValidatorTotalAccumulation(day uint64, concurrency uint64) error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*10))
	defer cancel()
	exportStart := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_validator_total_accumulation_stats").Observe(time.Since(exportStart).Seconds())
	}()

	if err := checkIfDayIsFinalized(day); err != nil {
		return err
	}

	start := time.Now()
	logger.Infof("validating if required data has been exported for total accumulation")
	type Exported struct {
		LastTotalAccumulation     bool `db:"last_total_accumulation_exported"`
		CurrentCLRewards          bool `db:"cur_cl_rewards_exported"`
		CurrentElRewards          bool `db:"cur_el_rewards_exported"`
		CurrentSyncDuties         bool `db:"cur_sync_duties_exported"`
		CurrentFailedAttestations bool `db:"cur_failed_attestations_exported"`
	}
	exported := Exported{}
	err := ReaderDb.Get(&exported, `
		SELECT 
			last.total_accumulation_exported as last_total_accumulation_exported, 
			cur.cl_rewards_exported as cur_cl_rewards_exported, 
			cur.el_rewards_exported as cur_el_rewards_exported,
			cur.sync_duties_exported as cur_sync_duties_exported,
			cur.failed_attestations_exported as cur_failed_attestations_exported
		FROM validator_stats_status cur
		INNER JOIN validator_stats_status last 
				ON last.day = GREATEST(cur.day - 1, 0)
		WHERE cur.day = $1;
	`, day)

	if err != nil {
		return fmt.Errorf("error retrieving required data: %w", err)
	} else if !(exported.LastTotalAccumulation || day == 0) || !exported.CurrentCLRewards || !exported.CurrentElRewards || !exported.CurrentSyncDuties || !exported.CurrentFailedAttestations {
		return fmt.Errorf("missing required export: last total accumulation: %v, cur cl rewards: %v, cur el rewards: %v, cur sync duties: %v, cur failed attestations: %v",
			!exported.LastTotalAccumulation, !exported.CurrentCLRewards, !exported.CurrentElRewards, !exported.CurrentSyncDuties, !exported.CurrentFailedAttestations)
	}
	logger.Infof("validating completed, took %v", time.Since(start))

	start = time.Now()

	logger.Infof("exporting total accumulation stats")
	maxValidatorIndex, err := GetTotalValidatorsCount()
	if err != nil {
		return err
	}
	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(int(concurrency))
	batchSize := 1000
	for b := 0; b <= int(maxValidatorIndex); b += batchSize {
		start := b
		end := b + batchSize
		if int(maxValidatorIndex) < end {
			end = int(maxValidatorIndex)
		}
		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
			}
			_, err = WriterDb.Exec(`INSERT INTO validator_stats (
				validatorindex,
				day,

				cl_rewards_gwei_total,
				el_rewards_wei_total,
				mev_rewards_wei_total,

				missed_attestations_total,

				participated_sync_total,
				missed_sync_total,
				orphaned_sync_total
				) (
					SELECT 
						vs1.validatorindex, 
						vs1.day, 
						COALESCE(vs1.cl_rewards_gwei, 0) + COALESCE(vs2.cl_rewards_gwei_total, 0),
						COALESCE(vs1.el_rewards_wei, 0) + COALESCE(vs2.el_rewards_wei_total, 0),
						COALESCE(vs1.mev_rewards_wei, 0) + COALESCE(vs2.mev_rewards_wei_total, 0),
						COALESCE(vs1.missed_attestations, 0) + COALESCE(vs2.missed_attestations_total, 0),
						COALESCE(vs1.participated_sync, 0) + COALESCE(vs2.participated_sync_total, 0),
						COALESCE(vs1.missed_sync, 0) + COALESCE(vs2.missed_sync_total, 0),
						COALESCE(vs1.orphaned_sync, 0) + COALESCE(vs2.orphaned_sync_total, 0)
					FROM validator_stats vs1 LEFT JOIN validator_stats vs2 ON vs2.day = vs1.day - 1 AND vs2.validatorindex = vs1.validatorindex WHERE vs1.day = $1 AND vs1.validatorindex >= $2 AND vs1.validatorindex < $3
				) 
				ON CONFLICT (validatorindex, day) DO UPDATE SET 
					cl_rewards_gwei_total = excluded.cl_rewards_gwei_total,
					el_rewards_wei_total = excluded.el_rewards_wei_total,
					mev_rewards_wei_total = excluded.mev_rewards_wei_total,
					missed_attestations_total = excluded.missed_attestations_total,
					participated_sync_total = excluded.participated_sync_total,
					missed_sync_total = excluded.missed_sync_total,
					orphaned_sync_total = excluded.orphaned_sync_total;
				`, day, start, end)
			if err != nil {
				return fmt.Errorf("error inserting accumulated data into validator_stats for day [%v], start [%v] and end [%v]: %w", day, start, end, err)
			}

			logger.Infof("populate total accumulation for validator stats table done for batch %v", start)
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		logrus.Error(err)
		return err
	}
	logger.Infof("export completed, took %v", time.Since(start))

	if err = markColumnExported(day, "total_accumulation_exported"); err != nil {
		return err
	}

	logger.Infof("total accumulation for statistics export of day %v completed, took %v", day, time.Since(exportStart))
	return nil
}

func WriteValidatorTotalPerformance(day uint64, concurrency uint64) error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*10))
	defer cancel()
	exportStart := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_validator_total_performance_stats").Observe(time.Since(exportStart).Seconds())
	}()

	if err := checkIfDayIsFinalized(day); err != nil {
		return err
	}

	start := time.Now()
	logger.Infof("validating if required data has been exported for total performance")
	type Exported struct {
		LastTotalPerformance     bool `db:"last_total_performance_exported"`
		CurrentTotalAccumulation bool `db:"cur_total_accumulation_exported"`
	}
	exported := Exported{}
	err := ReaderDb.Get(&exported, `
		SELECT 
			last.total_performance_exported as last_total_performance_exported, 
			cur.total_accumulation_exported as cur_total_accumulation_exported
		FROM validator_stats_status cur
		INNER JOIN validator_stats_status last 
				ON last.day = GREATEST(cur.day - 1, 0)
		WHERE cur.day = $1;
	`, day)

	if err != nil {
		return fmt.Errorf("error retrieving required data: %w", err)
	} else if !(exported.LastTotalPerformance || day == 0) || !exported.CurrentTotalAccumulation {
		return fmt.Errorf("missing required export: last total performance: %v, cur total exported: %v",
			!exported.LastTotalPerformance, !exported.CurrentTotalAccumulation)
	}
	logger.Infof("validating completed, took %v", time.Since(start))

	start = time.Now()

	logger.Infof("exporting total performance stats")
	maxValidatorIndex, err := GetTotalValidatorsCount()
	if err != nil {
		return err
	}
	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(int(concurrency))
	batchSize := 1000
	for b := 0; b <= int(maxValidatorIndex); b += batchSize {
		start := b
		end := b + batchSize
		if int(maxValidatorIndex) < end {
			end = int(maxValidatorIndex)
		}
		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
			}

			_, err = WriterDb.Exec(`insert into validator_performance (
				validatorindex,
				balance,
				rank7d,

				cl_performance_1d,
				cl_performance_7d,
				cl_performance_31d,
				cl_performance_365d,
				cl_performance_total,

				el_performance_1d,
				el_performance_7d,
				el_performance_31d,
				el_performance_365d,
				el_performance_total,

				mev_performance_1d,
				mev_performance_7d,
				mev_performance_31d,
				mev_performance_365d,
				mev_performance_total
				) (
					select 
					vs_now.validatorindex, 
						COALESCE(vs_now.end_balance, 0) as balance,
						0 as rank7d,

						coalesce(vs_now.cl_rewards_gwei_total, 0) - coalesce(vs_1d.cl_rewards_gwei_total, 0) as cl_performance_1d, 
						coalesce(vs_now.cl_rewards_gwei_total, 0) - coalesce(vs_7d.cl_rewards_gwei_total, 0) as cl_performance_7d, 
						coalesce(vs_now.cl_rewards_gwei_total, 0) - coalesce(vs_31d.cl_rewards_gwei_total, 0) as cl_performance_31d, 
						coalesce(vs_now.cl_rewards_gwei_total, 0) - coalesce(vs_365d.cl_rewards_gwei_total, 0) as cl_performance_365d,
						coalesce(vs_now.cl_rewards_gwei_total, 0) as cl_performance_total, 
						
						coalesce(vs_now.el_rewards_wei_total, 0) - coalesce(vs_1d.el_rewards_wei_total, 0) as el_performance_1d, 
						coalesce(vs_now.el_rewards_wei_total, 0) - coalesce(vs_7d.el_rewards_wei_total, 0) as el_performance_7d, 
						coalesce(vs_now.el_rewards_wei_total, 0) - coalesce(vs_31d.el_rewards_wei_total, 0) as el_performance_31d, 
						coalesce(vs_now.el_rewards_wei_total, 0) - coalesce(vs_365d.el_rewards_wei_total, 0) as el_performance_365d,
						coalesce(vs_now.el_rewards_wei_total, 0) as el_performance_total, 
						
						coalesce(vs_now.mev_rewards_wei_total, 0) - coalesce(vs_1d.mev_rewards_wei_total, 0) as mev_performance_1d, 
						coalesce(vs_now.mev_rewards_wei_total, 0) - coalesce(vs_7d.mev_rewards_wei_total, 0) as mev_performance_7d, 
						coalesce(vs_now.mev_rewards_wei_total, 0) - coalesce(vs_31d.mev_rewards_wei_total, 0) as mev_performance_31d, 
						coalesce(vs_now.mev_rewards_wei_total, 0) - coalesce(vs_365d.mev_rewards_wei_total, 0) as mev_performance_365d,
						coalesce(vs_now.mev_rewards_wei_total, 0) as mev_performance_total
					from validator_stats vs_now
					left join validator_stats vs_1d on vs_1d.validatorindex = vs_now.validatorindex and vs_1d.day = $2
					left join validator_stats vs_7d on vs_7d.validatorindex = vs_now.validatorindex and vs_7d.day = $3
					left join validator_stats vs_31d on vs_31d.validatorindex = vs_now.validatorindex and vs_31d.day = $4
					left join validator_stats vs_365d on vs_365d.validatorindex = vs_now.validatorindex and vs_365d.day = $5
					where vs_now.day = $1 AND vs_now.validatorindex >= $6 AND vs_now.validatorindex < $7
				) 
				on conflict (validatorindex) do update set 
					balance = excluded.balance,
					rank7d=excluded.rank7d,

					cl_performance_1d=excluded.cl_performance_1d,
					cl_performance_7d=excluded.cl_performance_7d,
					cl_performance_31d=excluded.cl_performance_31d,
					cl_performance_365d=excluded.cl_performance_365d,
					cl_performance_total=excluded.cl_performance_total,

					el_performance_1d=excluded.el_performance_1d,
					el_performance_7d=excluded.el_performance_7d,
					el_performance_31d=excluded.el_performance_31d,
					el_performance_365d=excluded.el_performance_365d,
					el_performance_total=excluded.el_performance_total,

					mev_performance_1d=excluded.mev_performance_1d,
					mev_performance_7d=excluded.mev_performance_7d,
					mev_performance_31d=excluded.mev_performance_31d,
					mev_performance_365d=excluded.mev_performance_365d,
					mev_performance_total=excluded.mev_performance_total
			;`, day, int64(day)-1, int64(day)-7, int64(day)-31, int64(day)-365, start, end)

			if err != nil {
				return fmt.Errorf("error inserting performance into validator_performance for day [%v], start [%v] and end [%v]: %w", day, start, end, err)
			}

			logger.Infof("populate validator_performance table done for batch %v", start)
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		logrus.Error(err)
		return err
	}
	logger.Infof("export completed, took %v", time.Since(start))

	start = time.Now()
	logger.Infof("populate validator_performance rank7d")

	_, err = WriterDb.Exec(`
		WITH ranked_performance AS (
			SELECT
				validatorindex, 
				row_number() OVER (ORDER BY cl_performance_7d DESC) AS rank7d
			FROM validator_performance
		)
		UPDATE validator_performance vp
		SET rank7d = rp.rank7d
		FROM ranked_performance rp
		WHERE vp.validatorindex = rp.validatorindex
		`)
	if err != nil {
		return fmt.Errorf("error updating rank7d while exporting day [%v]: %w", day, err)
	}

	logger.Infof("export completed, took %v", time.Since(start))

	if err = markColumnExported(day, "total_performance_exported"); err != nil {
		return err
	}

	logger.Infof("total performance statistics export of day %v completed, took %v", day, time.Since(exportStart))
	return nil
}

func WriteValidatorBlockStats(day uint64) error {
	exportStart := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_validator_block_stats").Observe(time.Since(exportStart).Seconds())
	}()

	if err := checkIfDayIsFinalized(day); err != nil {
		return err
	}

	firstEpoch, lastEpoch := utils.GetFirstAndLastEpochForDay(day)

	tx, err := WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	start := time.Now()

	logger.Infof("exporting proposed_blocks, missed_blocks and orphaned_blocks statistics")
	_, err = tx.Exec(`
		insert into validator_stats (validatorindex, day, proposed_blocks, missed_blocks, orphaned_blocks) 
		(
			select proposer, $3, sum(case when status = '1' then 1 else 0 end), sum(case when status = '2' then 1 else 0 end), sum(case when status = '3' then 1 else 0 end)
			from blocks
			where epoch >= $1 and epoch <= $2 and proposer != $4
			group by proposer
		) 
		on conflict (validatorindex, day) do update set proposed_blocks = excluded.proposed_blocks, missed_blocks = excluded.missed_blocks, orphaned_blocks = excluded.orphaned_blocks;`,
		firstEpoch, lastEpoch, day, MaxSqlInteger)
	if err != nil {
		return fmt.Errorf("error inserting blocks into validator_stats for day [%v], firstEpoch [%v] and lastEpoch [%v]: %w", day, firstEpoch, lastEpoch, err)
	}
	logger.Infof("export completed, took %v", time.Since(start))

	start = time.Now()
	logger.Infof("exporting attester_slashings and proposer_slashings statistics")
	_, err = tx.Exec(`
		insert into validator_stats (validatorindex, day, attester_slashings, proposer_slashings) 
		(
			select proposer, $3, sum(attesterslashingscount), sum(proposerslashingscount)
			from blocks
			where epoch >= $1 and epoch <= $2 and status = '1' and proposer != $4
			group by proposer
		) 
		on conflict (validatorindex, day) do update set attester_slashings = excluded.attester_slashings, proposer_slashings = excluded.proposer_slashings;`,
		firstEpoch, lastEpoch, day, MaxSqlInteger)
	if err != nil {
		return fmt.Errorf("error inserting slashings into validator_stats for day [%v], firstEpoch [%v] and lastEpoch [%v]: %w", day, firstEpoch, lastEpoch, err)
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	logger.Infof("export completed, took %v", time.Since(start))

	if err = markColumnExported(day, "block_stats_exported"); err != nil {
		return err
	}

	logger.Infof("block statistics export of day %v completed, took %v", day, time.Since(exportStart))
	return nil
}

func WriteValidatorElIcome(day uint64) error {
	exportStart := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_validator_el_income_stats").Observe(time.Since(exportStart).Seconds())
	}()

	if err := checkIfDayIsFinalized(day); err != nil {
		return err
	}

	firstEpoch, lastEpoch := utils.GetFirstAndLastEpochForDay(day)

	tx, err := WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	start := time.Now()

	logger.Infof("exporting mev & el rewards")

	type Container struct {
		Slot            uint64 `db:"slot"`
		ExecBlockNumber uint64 `db:"exec_block_number"`
		Proposer        uint64 `db:"proposer"`
		TxFeeReward     *big.Int
		MevReward       *big.Int
	}

	blocks := make([]*Container, 0)
	blocksMap := make(map[uint64]*Container)

	err = tx.Select(&blocks, "SELECT slot, exec_block_number, proposer FROM blocks WHERE epoch >= $1 AND epoch <= $2 AND exec_block_number > 0 AND status = '1'", firstEpoch, lastEpoch)
	if err != nil {
		return fmt.Errorf("error retrieving blocks data for firstEpoch [%v] and lastEpoch [%v]: %w", firstEpoch, lastEpoch, err)
	}

	numbers := make([]uint64, 0, len(blocks))

	for _, b := range blocks {
		numbers = append(numbers, b.ExecBlockNumber)
		blocksMap[b.ExecBlockNumber] = b
	}

	blocksData, err := BigtableClient.GetBlocksIndexedMultiple(numbers, uint64(len(numbers)))
	if err != nil {
		return fmt.Errorf("error in GetBlocksIndexedMultiple: %w", err)
	}

	relaysData, err := GetRelayDataForIndexedBlocks(blocksData)
	if err != nil {
		return fmt.Errorf("error in GetRelayDataForIndexedBlocks: %w", err)
	}

	proposerRewards := make(map[uint64]*Container)
	for _, b := range blocksData {
		proposer := blocksMap[b.Number].Proposer

		if proposerRewards[proposer] == nil {
			proposerRewards[proposer] = &Container{
				MevReward:   big.NewInt(0),
				TxFeeReward: big.NewInt(0),
			}
		}

		txFeeReward := new(big.Int).SetBytes(b.TxReward)
		proposerRewards[proposer].TxFeeReward = new(big.Int).Add(txFeeReward, proposerRewards[proposer].TxFeeReward)

		mevReward, ok := relaysData[common.BytesToHash(b.Hash)]

		if ok {
			proposerRewards[proposer].MevReward = new(big.Int).Add(mevReward.MevBribe.BigInt(), proposerRewards[proposer].MevReward)
		} else {
			proposerRewards[proposer].MevReward = new(big.Int).Add(txFeeReward, proposerRewards[proposer].MevReward)
		}
	}
	logrus.Infof("retrieved mev / el rewards data for %v proposer", len(proposerRewards))

	if len(proposerRewards) > 0 {
		numArgs := 4
		valueStrings := make([]string, 0, len(proposerRewards))
		valueArgs := make([]interface{}, 0, len(proposerRewards)*numArgs)
		i := 0
		for proposer, rewards := range proposerRewards {

			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", i*numArgs+1, i*numArgs+2, i*numArgs+3, i*numArgs+4))
			valueArgs = append(valueArgs, proposer)
			valueArgs = append(valueArgs, day)
			valueArgs = append(valueArgs, rewards.TxFeeReward.String())
			valueArgs = append(valueArgs, rewards.MevReward.String())

			i++
		}
		stmt := fmt.Sprintf(`
				INSERT INTO validator_stats (validatorindex, day, el_rewards_wei, mev_rewards_wei) VALUES
				%s
				ON CONFLICT(validatorindex, day) DO UPDATE SET el_rewards_wei = excluded.el_rewards_wei, mev_rewards_wei = excluded.mev_rewards_wei;`,
			strings.Join(valueStrings, ","))
		_, err = tx.Exec(stmt, valueArgs...)
		if err != nil {
			return fmt.Errorf("error inserting el_rewards into validator_stats for day [%v]: %w", day, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	logger.Infof("export completed, took %v", time.Since(start))

	if err = markColumnExported(day, "el_rewards_exported"); err != nil {
		return err
	}

	logger.Infof("el rewards statistics export of day %v completed, took %v", day, time.Since(exportStart))
	return nil
}

func WriteValidatorClIcome(day uint64, concurrency uint64) error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*10))
	defer cancel()
	exportStart := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_validator_cl_income_stats").Observe(time.Since(exportStart).Seconds())
	}()

	if err := checkIfDayIsFinalized(day); err != nil {
		return err
	}

	start := time.Now()
	logger.Infof("validating if required data has been exported for cl rewards")
	type Exported struct {
		LastBalanceExported                bool `db:"last_balance_exported"`
		CurrentBalanceExported             bool `db:"cur_balance_exported"`
		CurrentWithdrawalsDepositsExported bool `db:"cur_withdrawals_deposits_exported"`
	}
	exported := Exported{}
	err := ReaderDb.Get(&exported, `
		SELECT last.balance_exported as last_balance_exported, cur.balance_exported as cur_balance_exported, cur.withdrawals_deposits_exported as cur_withdrawals_deposits_exported
		FROM validator_stats_status cur
		INNER JOIN validator_stats_status last 
				ON last.day = GREATEST(cur.day - 1, 0)
		WHERE cur.day = $1;
	`, day)

	if err != nil {
		return fmt.Errorf("error retrieving required data: %w", err)
	} else if !exported.CurrentBalanceExported || !exported.CurrentWithdrawalsDepositsExported || !exported.LastBalanceExported {
		return fmt.Errorf("missing required export: cur balance: %v, cur withdrwals/deposits: %v, last balance: %v", !exported.CurrentBalanceExported, !exported.CurrentWithdrawalsDepositsExported, !exported.LastBalanceExported)
	}
	logger.Infof("validating took %v", time.Since(start))

	start = time.Now()
	_, lastEpoch := utils.GetFirstAndLastEpochForDay(day)

	logger.Infof("exporting cl_rewards_wei statistics")

	maxValidatorIndex, err := BigtableClient.GetMaxValidatorindexForEpoch(lastEpoch)
	if err != nil {
		return fmt.Errorf("error in GetAggregatedValidatorIncomeDetailsHistory: could not get max validator index from validator income history for last epoch [%v] of day [%v]: %v", lastEpoch, day, err)
	} else if maxValidatorIndex == uint64(0) {
		return fmt.Errorf("error in GetAggregatedValidatorIncomeDetailsHistory: no validator found for last epoch [%v] of day [%v]: %v", lastEpoch, day, err)
	}

	maxValidatorIndex++

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(int(concurrency))

	batchSize := 100 // max parameters: 65535 / 3, but it's faster in smaller batches
	for b := 0; b < int(maxValidatorIndex); b += batchSize {
		start := b
		end := b + batchSize
		if int(maxValidatorIndex) < end {
			end = int(maxValidatorIndex)
		}

		logrus.Info(start, end)

		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
			}
			stmt := `
				INSERT INTO validator_stats (validatorindex, day, cl_rewards_gwei) 
				(
					SELECT cur.validatorindex, cur.day, COALESCE(cur.end_balance, 0) - COALESCE(last.end_balance, 0) + COALESCE(cur.withdrawals_amount, 0) - COALESCE(cur.deposits_amount, 0) AS cl_rewards_gwei
					FROM validator_stats cur
					LEFT JOIN validator_stats last 
						ON cur.validatorindex = last.validatorindex AND last.day = GREATEST(cur.day - 1, 0)
					WHERE cur.day = $1 AND cur.validatorindex >= $2 AND cur.validatorindex < $3
				)
				ON CONFLICT (validatorindex, day) DO
					UPDATE SET cl_rewards_gwei = excluded.cl_rewards_gwei;`
			if day == 0 {
				stmt = `
					INSERT INTO validator_stats (validatorindex, day, cl_rewards_gwei) 
					(
						SELECT cur.validatorindex, cur.day, COALESCE(cur.end_balance, 0) - COALESCE(cur.start_balance,0) + COALESCE(cur.withdrawals_amount, 0) - COALESCE(cur.deposits_amount, 0) AS cl_rewards_gwei
						FROM validator_stats cur
						WHERE cur.day = $1 AND cur.validatorindex >= $2 AND cur.validatorindex < $3
					)
					ON CONFLICT (validatorindex, day) DO
						UPDATE SET cl_rewards_gwei = excluded.cl_rewards_gwei;`
			}
			_, err = WriterDb.Exec(stmt, day, start, end)
			if err != nil {
				return fmt.Errorf("error inserting cl_rewards_gwei into validator_stats for day [%v], start [%v] and end [%v]: %w", day, start, end, err)
			}
			logrus.Infof("saving validator cl rewards gwei batch %v completed", start)
			return nil
		})
	}

	if err = g.Wait(); err != nil {
		logrus.Error(err)
		return err
	}

	logger.Infof("export completed, took %v", time.Since(start))

	if err = markColumnExported(day, "cl_rewards_exported"); err != nil {
		return err
	}

	logger.Infof("cl rewards statistics export of day %v completed, took %v", day, time.Since(exportStart))
	return nil
}

func WriteValidatorBalances(day uint64) error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*10))
	defer cancel()

	exportStart := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_validator_balances_stats").Observe(time.Since(exportStart).Seconds())
	}()

	if err := checkIfDayIsFinalized(day); err != nil {
		return err
	}

	firstEpoch, lastEpoch := utils.GetFirstAndLastEpochForDay(day)

	start := time.Now()

	logger.Infof("exporting min_balance, max_balance, min_effective_balance, max_effective_balance, start_balance, start_effective_balance, end_balance and end_effective_balance statistics")
	balanceStatistics, err := BigtableClient.GetValidatorBalanceStatistics(firstEpoch, lastEpoch)
	if err != nil {
		return fmt.Errorf("error in GetValidatorBalanceStatistics for firstEpoch [%v] and lastEpoch [%v]: %w", firstEpoch, lastEpoch, err)
	}

	balanceStatsArr := make([]*types.ValidatorBalanceStatistic, 0, len(balanceStatistics))
	for _, stat := range balanceStatistics {
		balanceStatsArr = append(balanceStatsArr, stat)
	}
	logger.Infof("fetching balance completed, took %v, now we save it", time.Since(start))
	start = time.Now()

	g, gCtx := errgroup.WithContext(ctx)

	batchSize := 100 // max parameters: 65535 / 10, but we are faster with smaller batch sizes
	for b := 0; b < len(balanceStatsArr); b += batchSize {
		start := b
		end := b + batchSize
		if len(balanceStatsArr) < end {
			end = len(balanceStatsArr)
		}

		numArgs := 10
		valueStrings := make([]string, 0, batchSize)
		valueArgs := make([]interface{}, 0, batchSize*numArgs)

		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
			}
			defer logger.Infof("saving validator balance batch %v completed", start)
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
			_, err := WriterDb.Exec(stmt, valueArgs...)

			if err != nil {
				return fmt.Errorf("error inserting balances into validator_stats for day [%v], start [%v] and end [%v]: %w", day, start, end, err)
			}

			return nil
		})
	}

	if err = g.Wait(); err != nil {
		logrus.Error(err)
		return err
	}

	logger.Infof("export completed, took %v", time.Since(start))

	if err = markColumnExported(day, "balance_exported"); err != nil {
		return err
	}

	logger.Infof("balance statistics export of day %v completed, took %v", day, time.Since(exportStart))
	return nil
}

func WriteValidatorDepositWithdrawals(day uint64) error {
	exportStart := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_validator_deposit_withdrawal_stats").Observe(time.Since(exportStart).Seconds())
	}()

	if err := checkIfDayIsFinalized(day); err != nil {
		return err
	}

	// The end_balance of a day is the balance after the first slot of the last epoch of that day.
	// Therefore the last 31 slots of the day are not included in the end_balance of that day.
	// Since our income calculation is base on subtracting end_balances the deposits and withdrawals that happen during those slots must be added to the next day instead.
	firstSlot := uint64(0)
	if day > 0 {
		firstSlot = utils.GetLastBalanceInfoSlotForDay(day-1) + 1
	}
	lastSlot := utils.GetLastBalanceInfoSlotForDay(day)

	tx, err := WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	start := time.Now()
	logrus.Infof("Resetting Withdrawals + Deposits for day [%v]", day)

	firstDayExtraCondition := ""
	if day == 0 {
		// genesis-deposits will be added to day -1.
		firstDayExtraCondition = " OR day = -1"
	}

	resetQry := fmt.Sprintf(`
		UPDATE validator_stats SET 
			deposits = NULL, 
			deposits_amount = NULL,
			withdrawals = NULL, 
			withdrawals_amount = NULL
		WHERE day = $1%s;`, firstDayExtraCondition)

	_, err = tx.Exec(resetQry, day)
	if err != nil {
		return fmt.Errorf("error resetting validator_stats for day [%v]: %w", day, err)
	}
	logger.Infof("reset completed, took %v", time.Since(start))

	start = time.Now()
	logrus.Infof("Update Withdrawals + Deposits for day [%v] slot %v -> %v", day, firstSlot, lastSlot)

	logger.Infof("exporting deposits and deposits_amount statistics")
	depositsQry := `
		insert into validator_stats (validatorindex, day, deposits, deposits_amount) 
		(
			select validators.validatorindex, $3, count(*), sum(amount)
			from blocks_deposits
			inner join validators on blocks_deposits.publickey = validators.pubkey
			inner join blocks on blocks_deposits.block_root = blocks.blockroot
			where blocks.slot >= $1 and blocks.slot <= $2 and blocks.status = '1' and blocks_deposits.valid_signature
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
				inner join blocks on blocks_deposits.block_root = blocks.blockroot
				where blocks.slot >= $1 and blocks.slot <= $2 and blocks.status = '1'
				group by validators.validatorindex, day
			) 
			on conflict (validatorindex, day) do
				update set deposits = excluded.deposits, 
				deposits_amount = excluded.deposits_amount;`
	}

	_, err = tx.Exec(depositsQry, firstSlot, lastSlot, day)
	if err != nil {
		return fmt.Errorf("error inserting deposits into validator_stats for day [%v], firstSlot [%v] and lastSlot [%v]: %w", day, firstSlot, lastSlot, err)
	}
	logger.Infof("export completed, took %v", time.Since(start))

	start = time.Now()
	logger.Infof("exporting withdrawals and withdrawals_amount statistics")
	withdrawalsQuery := `
		insert into validator_stats (validatorindex, day, withdrawals, withdrawals_amount) 
		(
			select validatorindex, $3, count(*), sum(amount)
			from blocks_withdrawals
			inner join blocks on blocks_withdrawals.block_root = blocks.blockroot
			where block_slot >= $1 and block_slot <= $2 and blocks.status = '1'
			group by validatorindex
		) 
		on conflict (validatorindex, day) do
			update set withdrawals = excluded.withdrawals, 
			withdrawals_amount = excluded.withdrawals_amount;`
	_, err = tx.Exec(withdrawalsQuery, firstSlot, lastSlot, day)
	if err != nil {
		return fmt.Errorf("error inserting withdrawals into validator_stats for day [%v], firstSlot [%v] and lastSlot [%v]: %w", day, firstSlot, lastSlot, err)
	}
	if err = tx.Commit(); err != nil {
		return err
	}

	logger.Infof("export completed, took %v", time.Since(start))

	if err = markColumnExported(day, "withdrawals_deposits_exported"); err != nil {
		return err
	}

	logger.Infof("deposits and withdrawals statistics export of day %v completed, took %v", day, time.Since(exportStart))
	return nil
}

func WriteValidatorSyncDutiesForDay(day uint64) error {
	exportStart := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_validator_sync_stats").Observe(time.Since(exportStart).Seconds())
	}()

	if err := checkIfDayIsFinalized(day); err != nil {
		return err
	}

	startEpoch, endEpoch := utils.GetFirstAndLastEpochForDay(day)

	start := time.Now()
	logrus.Infof("Update Sync duties for day [%v] epoch %v -> %v", day, startEpoch, endEpoch)

	syncStats, err := BigtableClient.GetValidatorSyncDutiesStatistics([]uint64{}, startEpoch, endEpoch)
	if err != nil {
		return fmt.Errorf("error in GetValidatorSyncDutiesStatistics for startEpoch [%v] and endEpoch [%v]: %w", startEpoch, endEpoch, err)
	}
	logrus.Infof("getting sync duties done in %v, now we export them to the db", time.Since(start))
	start = time.Now()

	syncStatsArr := make([]*types.ValidatorSyncDutiesStatistic, 0, len(syncStats))
	for _, stat := range syncStats {
		syncStatsArr = append(syncStatsArr, stat)
	}

	tx, err := WriterDb.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	batchSize := 13000 // max parameters: 65535
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
			valueArgs = append(valueArgs, stat.OrphanedSync)
		}
		stmt := fmt.Sprintf(`
			insert into validator_stats (validatorindex, day, participated_sync, missed_sync, orphaned_sync)  VALUES
			%s
			on conflict (validatorindex, day) do update set participated_sync = excluded.participated_sync, missed_sync = excluded.missed_sync, orphaned_sync = excluded.orphaned_sync;`,
			strings.Join(valueStrings, ","))
		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			return fmt.Errorf("error inserting sync information into validator_stats for day [%v], start [%v] and end [%v]: %w", day, start, end, err)
		}

		logrus.Infof("saving sync statistics batch %v completed", b)
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	logger.Infof("export completed, took %v", time.Since(start))

	if err = markColumnExported(day, "sync_duties_exported"); err != nil {
		return err
	}

	logger.Infof("sync duties and statistics export of day %v completed, took %v", day, time.Since(exportStart))
	return nil
}

func WriteValidatorFailedAttestationsStatisticsForDay(day uint64, concurrency uint64) error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*10))
	defer cancel()
	exportStart := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_validator_failed_att_stats").Observe(time.Since(exportStart).Seconds())
	}()

	if err := checkIfDayIsFinalized(day); err != nil {
		return err
	}

	firstEpoch, lastEpoch := utils.GetFirstAndLastEpochForDay(day)

	start := time.Now()

	logrus.Infof("exporting 'failed attestations' statistics firstEpoch: %v lastEpoch: %v", firstEpoch, lastEpoch)

	// first key is the batch start index and the second is the validator id
	failed := map[uint64]map[uint64]*types.ValidatorMissedAttestationsStatistic{}
	mux := sync.Mutex{}
	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(int(concurrency))
	epochBatchSize := uint64(2) // Fetching 2 Epochs per batch seems to be the fastest way to go
	for i := firstEpoch; i < lastEpoch; i += epochBatchSize {
		fromEpoch := i
		toEpoch := fromEpoch + epochBatchSize
		if toEpoch >= lastEpoch {
			toEpoch = lastEpoch
		} else {
			toEpoch--
		}
		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
			}
			ma, err := BigtableClient.GetValidatorMissedAttestationsCount([]uint64{}, fromEpoch, toEpoch)
			if err != nil {
				return fmt.Errorf("error in GetValidatorMissedAttestationsCount for fromEpoch [%v] and toEpoch [%v]: %w", fromEpoch, toEpoch, err)
			}
			mux.Lock()
			failed[fromEpoch] = ma
			mux.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	validatorMap := map[uint64]*types.ValidatorMissedAttestationsStatistic{}
	for _, f := range failed {

		for key, val := range f {
			if validatorMap[key] == nil {
				validatorMap[key] = val
			} else {
				validatorMap[key].MissedAttestations += val.MissedAttestations
			}
		}
	}

	logrus.Infof("fetching 'failed attestations' done in %v, now we export them to the db", time.Since(start))
	start = time.Now()
	maArr := make([]*types.ValidatorMissedAttestationsStatistic, 0, len(validatorMap))

	for _, stat := range validatorMap {
		maArr = append(maArr, stat)
	}

	g, gCtx = errgroup.WithContext(ctx)

	batchSize := 100 // max: 65535 / 4, but we are faster with smaller batches
	for b := 0; b < len(maArr); b += batchSize {

		start := b
		end := b + batchSize
		if len(maArr) < end {
			end = len(maArr)
		}

		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
			}

			err := saveFailedAttestationBatch(maArr[start:end], day)
			if err != nil {
				return fmt.Errorf("error in saveFailedAttestationBatch for day [%v], start [%v] and end [%v]: %w", day, start, end, err)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}
	logger.Infof("export completed, took %v", time.Since(start))

	if err := markColumnExported(day, "failed_attestations_exported"); err != nil {
		return err
	}

	logger.Infof("'failed attestation' statistics export of day %v completed, took %v", day, time.Since(exportStart))
	return nil
}

func saveFailedAttestationBatch(batch []*types.ValidatorMissedAttestationsStatistic, day uint64) error {
	var failedAttestationBatchNumArgs int = 4
	batchSize := len(batch)
	valueStrings := make([]string, 0, failedAttestationBatchNumArgs)
	valueArgs := make([]interface{}, 0, batchSize*failedAttestationBatchNumArgs)

	for i, stat := range batch {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", i*failedAttestationBatchNumArgs+1, i*failedAttestationBatchNumArgs+2, i*failedAttestationBatchNumArgs+3, i*failedAttestationBatchNumArgs+4))
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
	_, err := WriterDb.Exec(stmt, valueArgs...)
	if err != nil {
		return fmt.Errorf("error inserting failed attestations into validator_stats for day [%v]: %w", day, err)
	}

	return nil
}

func markColumnExported(day uint64, column string) error {
	start := time.Now()
	logger.Infof("marking [%v] exported for day [%v] as completed in the status table", column, day)

	_, err := WriterDb.Exec(fmt.Sprintf(`	
		INSERT INTO validator_stats_status (day, status, %[1]v) 
		VALUES ($1, false, true) 
		ON CONFLICT (day) 
			DO UPDATE SET %[1]v=EXCLUDED.%[1]v;
			`, column), day)
	if err != nil {
		return fmt.Errorf("error marking [%v] exported for day [%v] as completed in the status table: %w", column, day, err)
	}
	logrus.Infof("Marking complete in %v", time.Since(start))
	return nil
}

func GetValidatorIncomeHistoryChart(validatorIndices []uint64, currency string, lastFinalizedEpoch uint64, lowerBoundDay uint64) ([]*types.ChartDataPoint, error) {
	incomeHistory, err := GetValidatorIncomeHistory(validatorIndices, lowerBoundDay, 0, lastFinalizedEpoch)
	if err != nil {
		return nil, err
	}
	var clRewardsSeries = make([]*types.ChartDataPoint, len(incomeHistory))

	for i := 0; i < len(incomeHistory); i++ {
		color := "#7cb5ec"
		if incomeHistory[i].ClRewards < 0 {
			color = "#f7a35c"
		}
		balanceTs := utils.DayToTime(incomeHistory[i].Day)
		clRewardsSeries[i] = &types.ChartDataPoint{X: float64(balanceTs.Unix() * 1000), Y: utils.ExchangeRateForCurrency(currency) * (float64(incomeHistory[i].ClRewards) / 1e9), Color: color}
	}
	return clRewardsSeries, err
}

func GetValidatorIncomeHistory(validatorIndices []uint64, lowerBoundDay uint64, upperBoundDay uint64, lastFinalizedEpoch uint64) ([]types.ValidatorIncomeHistory, error) {
	if len(validatorIndices) == 0 {
		return []types.ValidatorIncomeHistory{}, nil
	}

	if upperBoundDay == 0 {
		upperBoundDay = 65536
	}

	validatorIndices = utils.SortedUniqueUint64(validatorIndices)
	validatorIndicesStr := make([]string, len(validatorIndices))
	for i, v := range validatorIndices {
		validatorIndicesStr[i] = fmt.Sprintf("%d", v)
	}

	validatorIndicesPqArr := pq.Array(validatorIndices)

	cacheDur := time.Second * time.Duration(utils.Config.Chain.Config.SecondsPerSlot*utils.Config.Chain.Config.SlotsPerEpoch+10) // updates every epoch, keep 10sec longer
	cacheKey := fmt.Sprintf("%d:validatorIncomeHistory:%d:%d:%d:%s", utils.Config.Chain.Config.DepositChainID, lowerBoundDay, upperBoundDay, lastFinalizedEpoch, strings.Join(validatorIndicesStr, ","))
	cached := []types.ValidatorIncomeHistory{}
	if _, err := cache.TieredCache.GetWithLocalTimeout(cacheKey, cacheDur, &cached); err == nil {
		return cached, nil
	}

	var result []types.ValidatorIncomeHistory
	err := ReaderDb.Select(&result, `
		SELECT 
			day, 
			SUM(COALESCE(cl_rewards_gwei, 0)) AS cl_rewards_gwei,
			SUM(COALESCE(end_balance, 0)) AS end_balance
		FROM validator_stats 
		WHERE validatorindex = ANY($1) AND day BETWEEN $2 AND $3 
		GROUP BY day 
		ORDER BY day
	;`, validatorIndicesPqArr, lowerBoundDay, upperBoundDay)
	if err != nil {
		return nil, err
	}

	// retrieve rewards for epochs not yet in stats
	if upperBoundDay == 65536 {
		lastDay := uint64(0)
		if len(result) > 0 {
			lastDay = uint64(result[len(result)-1].Day)
		} else {
			lastDay, err = GetLastExportedStatisticDay()
			if err != nil {
				return nil, err
			}
		}

		currentDay := lastDay + 1
		firstSlot := utils.GetLastBalanceInfoSlotForDay(lastDay) + 1
		lastSlot := lastFinalizedEpoch * utils.Config.Chain.Config.SlotsPerEpoch

		totalBalance := uint64(0)

		g := errgroup.Group{}
		g.Go(func() error {
			latestBalances, err := BigtableClient.GetValidatorBalanceHistory(validatorIndices, lastFinalizedEpoch, lastFinalizedEpoch)
			if err != nil {
				logger.Errorf("error getting validator balance data in GetValidatorEarnings: %v", err)
				return err
			}

			for _, balance := range latestBalances {
				if len(balance) == 0 {
					continue
				}

				totalBalance += balance[0].Balance
			}
			return nil
		})

		var lastBalance uint64
		g.Go(func() error {
			return GetValidatorBalanceForDay(validatorIndices, lastDay, &lastBalance)
		})

		var lastDeposits uint64
		g.Go(func() error {
			return GetValidatorDepositsForSlots(validatorIndices, firstSlot, lastSlot, &lastDeposits)
		})

		var lastWithdrawals uint64
		g.Go(func() error {
			return GetValidatorWithdrawalsForSlots(validatorIndices, firstSlot, lastSlot, &lastWithdrawals)
		})

		err = g.Wait()
		if err != nil {
			return nil, err
		}

		result = append(result, types.ValidatorIncomeHistory{
			Day:       int64(currentDay),
			ClRewards: int64(totalBalance - lastBalance - lastDeposits + lastWithdrawals),
		})
	}

	go func() {
		err := cache.TieredCache.Set(cacheKey, &result, cacheDur)
		if err != nil {
			utils.LogError(err, fmt.Errorf("error setting tieredCache for GetValidatorIncomeHistory with key %v", cacheKey), 0)
		}
	}()

	return result, nil
}

func WriteChartSeriesForDay(day int64) error {
	startTs := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		select {
		case <-gCtx.Done():
			return gCtx.Err()
		default:
		}
		err := WriteExecutionChartSeriesForDay(day)
		if err != nil {
			return err
		}
		return nil
	})

	g.Go(func() error {
		select {
		case <-gCtx.Done():
			return gCtx.Err()
		default:
		}
		err := WriteConsensusChartSeriesForDay(day)
		if err != nil {
			return err
		}
		return nil
	})

	err := g.Wait()
	if err != nil {
		return err
	}

	logger.Infof("marking day export as completed in the chart_series_status table for day %v", day)
	_, err = WriterDb.Exec("insert into chart_series_status (day, status) values ($1, true)", day)
	if err != nil {
		return err
	}

	logger.Infof("chart_series export completed: took %v", time.Since(startTs))
	return nil
}

func WriteConsensusChartSeriesForDay(day int64) error {
	if day < 0 {
		logger.Warnf("no consensus-charts for day < 0: %v", day)
		return nil
	}

	epochsPerDay := utils.EpochsPerDay()
	beaconchainDay := day * int64(epochsPerDay)

	startDate := utils.EpochToTime(uint64(beaconchainDay))
	dateTrunc := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)

	// inclusive slot
	firstSlot := utils.TimeToFirstSlotOfEpoch(uint64(dateTrunc.Unix()))

	epochOffset := firstSlot % utils.Config.Chain.Config.SlotsPerEpoch
	firstSlot = firstSlot - epochOffset
	firstEpoch := firstSlot / utils.Config.Chain.Config.SlotsPerEpoch
	// exclusive slot
	lastSlot := int64(firstSlot) + int64(epochsPerDay*utils.Config.Chain.Config.SlotsPerEpoch)
	if firstSlot == 0 {
		nextDateTrunc := time.Date(startDate.Year(), startDate.Month(), startDate.Day()+1, 0, 0, 0, 0, time.UTC)
		lastSlot = int64(utils.TimeToFirstSlotOfEpoch(uint64(nextDateTrunc.Unix())))
	}
	lastEpoch := lastSlot / int64(utils.Config.Chain.Config.SlotsPerEpoch)
	lastSlot = lastEpoch * int64(utils.Config.Chain.Config.SlotsPerEpoch)

	logrus.WithFields(logrus.Fields{"day": day, "firstSlot": firstSlot, "lastSlot": lastSlot, "firstEpoch": firstEpoch, "lastEpoch": lastEpoch, "startDate": startDate, "dateTrunc": dateTrunc}).Infof("exporting consensus chart_series")

	var err error

	_, err = WriterDb.Exec(`insert into chart_series select $1 as time, 'STAKED_ETH' as indicator, eligibleether/1e9 as value from epochs where epoch = $2 limit 1 on conflict (time, indicator) do update set time = excluded.time, indicator = excluded.indicator, value = excluded.value`, dateTrunc, lastEpoch-1)
	if err != nil {
		return fmt.Errorf("error inserting STAKED_ETH into chart_series: %w", err)
	}

	_, err = WriterDb.Exec(`insert into chart_series select $1 as time, 'AVG_VALIDATOR_BALANCE_ETH' as indicator, avg(averagevalidatorbalance)/1e9 as value from epochs where epoch >= $2 and epoch < $3 on conflict (time, indicator) do update set time = excluded.time, indicator = excluded.indicator, value = excluded.value`, dateTrunc, firstEpoch, lastEpoch)
	if err != nil {
		return fmt.Errorf("error inserting AVG_VALIDATOR_BALANCE_ETH into chart_series: %w", err)
	}

	_, err = WriterDb.Exec(`insert into chart_series select $1 as time, 'AVG_PARTICIPATION_RATE' as indicator, avg(globalparticipationrate) as value from epochs where epoch >= $2 and epoch < $3 on conflict (time, indicator) do update set time = excluded.time, indicator = excluded.indicator, value = excluded.value`, dateTrunc, firstEpoch, lastEpoch)
	if err != nil {
		return fmt.Errorf("error inserting AVG_PARTICIPATION_RATE into chart_series: %w", err)
	}

	_, err = WriterDb.Exec(`insert into chart_series select $1 as time, 'AVG_STAKE_EFFECTIVENESS' as indicator, coalesce(avg(eligibleether) / avg(totalvalidatorbalance), 0) as value from epochs where totalvalidatorbalance != 0 AND eligibleether != 0 and epoch >= $2 and epoch < $3 on conflict (time, indicator) do update set time = excluded.time, indicator = excluded.indicator, value = excluded.value`, dateTrunc, firstEpoch, lastEpoch)
	if err != nil {
		return fmt.Errorf("error inserting AVG_STAKE_EFFECTIVENESS into chart_series: %w", err)
	}

	_, err = WriterDb.Exec(`insert into chart_series select $1 as time, 'EL_VALID_DEPOSITS_ETH' as indicator, coalesce(sum(amount)/1e9,0) as value from eth1_deposits where valid_signature = true and block_ts >= $1 and block_ts < ($1 + interval '24 hour') on conflict (time, indicator) do update set time = excluded.time, indicator = excluded.indicator, value = excluded.value`, dateTrunc)
	if err != nil {
		return fmt.Errorf("error inserting EL_VALID_DEPOSITS_ETH into chart_series: %w", err)
	}

	_, err = WriterDb.Exec(`insert into chart_series select $1 as time, 'EL_INVALID_DEPOSITS_ETH' as indicator, coalesce(sum(amount)/1e9,0) as value from eth1_deposits where valid_signature = false and block_ts >= $1 and block_ts < ($1 + interval '24 hour') on conflict (time, indicator) do update set time = excluded.time, indicator = excluded.indicator, value = excluded.value`, dateTrunc)
	if err != nil {
		return fmt.Errorf("error inserting EL_INVALID_DEPOSITS_ETH into chart_series: %w", err)
	}

	_, err = WriterDb.Exec(`insert into chart_series select $1 as time, 'CL_DEPOSITS_ETH' as indicator, coalesce(sum(amount)/1e9,0) as value from blocks_deposits where block_slot >= $2 and block_slot < $3 on conflict (time, indicator) do update set time = excluded.time, indicator = excluded.indicator, value = excluded.value`, dateTrunc, firstSlot, lastSlot)
	if err != nil {
		return fmt.Errorf("error inserting CL_DEPOSITS_ETH into chart_series: %w", err)
	}

	_, err = WriterDb.Exec(`insert into chart_series select $1 as time, 'WITHDRAWALS_ETH' as indicator, coalesce(sum(w.amount)/1e9,0) as value from blocks_withdrawals w inner join blocks b ON w.block_root = b.blockroot AND b.status = '1' where w.block_slot >= $2 and w.block_slot < $3 on conflict (time, indicator) do update set time = excluded.time, indicator = excluded.indicator, value = excluded.value`, dateTrunc, firstSlot, lastSlot)
	if err != nil {
		return fmt.Errorf("error inserting WITHDRAWALS_ETH into chart_series: %w", err)
	}

	_, err = WriterDb.Exec(`insert into chart_series select $1 as time, 'PROPOSED_BLOCKS' as indicator, count(*) as value from blocks where status = '1' and slot >= $2 and slot < $3 on conflict (time, indicator) do update set time = excluded.time, indicator = excluded.indicator, value = excluded.value`, dateTrunc, firstSlot, lastSlot)
	if err != nil {
		return fmt.Errorf("error inserting PROPOSED_BLOCKS into chart_series: %w", err)
	}

	_, err = WriterDb.Exec(`insert into chart_series select $1 as time, 'MISSED_BLOCKS' as indicator, count(*) as value from blocks where status = '2' and slot >= $2 and slot < $3 on conflict (time, indicator) do update set time = excluded.time, indicator = excluded.indicator, value = excluded.value`, dateTrunc, firstSlot, lastSlot)
	if err != nil {
		return fmt.Errorf("error inserting MISSED_BLOCKS into chart_series: %w", err)
	}

	_, err = WriterDb.Exec(`insert into chart_series select $1 as time, 'ORPHANED_BLOCKS' as indicator, count(*) as value from blocks where status = '3' and slot >= $2 and slot < $3 on conflict (time, indicator) do update set time = excluded.time, indicator = excluded.indicator, value = excluded.value`, dateTrunc, firstSlot, lastSlot)
	if err != nil {
		return fmt.Errorf("error inserting ORPHANED_BLOCKS into chart_series: %w", err)
	}

	return nil
}

func WriteExecutionChartSeriesForDay(day int64) error {
	if utils.Config.Chain.Config.DepositChainID != 1 {
		// logger.Warnf("not writing chart_series for execution: chainId != 1: %v", utils.Config.Chain.Config.DepositChainID)
		return nil
	}

	if day < 0 {
		// before the beaconchain
		logger.Warnf("no execution charts for days before beaconchain")
		return nil
	}

	epochsPerDay := utils.EpochsPerDay()
	beaconchainDay := day * int64(epochsPerDay)

	startDate := utils.EpochToTime(uint64(beaconchainDay))
	dateTrunc := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)

	// inclusive slot
	firstSlot := utils.TimeToFirstSlotOfEpoch(uint64(dateTrunc.Unix()))
	firstEpoch := firstSlot / utils.Config.Chain.Config.SlotsPerEpoch
	// exclusive slot
	lastSlot := int64(firstSlot) + int64(epochsPerDay*utils.Config.Chain.Config.SlotsPerEpoch)
	// The first day is not a whole day, so we take the first slot from the next day as lastSlot
	if firstSlot == 0 {
		nextDateTrunc := time.Date(startDate.Year(), startDate.Month(), startDate.Day()+1, 0, 0, 0, 0, time.UTC)
		lastSlot = int64(utils.TimeToFirstSlotOfEpoch(uint64(nextDateTrunc.Unix())))
	}
	lastEpoch := lastSlot / int64(utils.Config.Chain.Config.SlotsPerEpoch)

	latestFinalizedEpoch, err := GetLatestFinalizedEpoch()
	if err != nil {
		return fmt.Errorf("error getting latest finalized epoch from db %w", err)
	}

	if lastEpoch > int64(latestFinalizedEpoch) {
		return fmt.Errorf("delaying chart series export as not all epochs for day %v finalized. last epoch of the day [%v] last finalized epoch [%v]", day, lastEpoch, latestFinalizedEpoch)
	}

	firstBlock, err := GetBlockNumber(uint64(firstSlot))
	if err != nil {
		return fmt.Errorf("error getting block number for slot: %v err: %w", firstSlot, err)
	}

	if firstBlock <= 15537394 {
		return fmt.Errorf("this function does not yet handle pre merge statistics")
	}

	lastBlock, err := GetBlockNumber(uint64(lastSlot))
	if err != nil {
		return fmt.Errorf("error getting block number for slot: %v err: %w", lastSlot, err)
	}
	logger.Infof("exporting chart_series for day %v ts: %v (slot %v to %v, block %v to %v)", day, dateTrunc, firstSlot, lastSlot, firstBlock, lastBlock)

	blocksChan := make(chan *types.Eth1Block, 360)
	batchSize := int64(360)
	go func(stream chan *types.Eth1Block) {
		logger.Infof("querying blocks from %v to %v", firstBlock, lastBlock)
		for b := int64(lastBlock) - 1; b > int64(firstBlock); b -= batchSize {
			high := b
			low := b - batchSize + 1
			if int64(firstBlock) > low {
				low = int64(firstBlock)
			}

			err := BigtableClient.GetFullBlocksDescending(stream, uint64(high), uint64(low))
			if err != nil {
				logger.Errorf("error getting blocks descending high: %v low: %v err: %v", high, low, err)
			}

		}
		close(stream)
	}(blocksChan)

	// logger.Infof("got %v blocks", len(blocks))

	blockCount := int64(0)
	txCount := int64(0)

	totalBaseFee := decimal.NewFromInt(0)
	totalGasPrice := decimal.NewFromInt(0)
	totalTxSavings := decimal.NewFromInt(0)
	totalTxFees := decimal.NewFromInt(0)
	totalBurned := decimal.NewFromInt(0)
	totalGasUsed := decimal.NewFromInt(0)

	legacyTxCount := int64(0)
	accessListTxCount := int64(0)
	eip1559TxCount := int64(0)
	failedTxCount := int64(0)
	successTxCount := int64(0)

	totalFailedGasUsed := decimal.NewFromInt(0)
	totalFailedTxFee := decimal.NewFromInt(0)

	totalBaseBlockReward := decimal.NewFromInt(0)

	totalGasLimit := decimal.NewFromInt(0)
	totalTips := decimal.NewFromInt(0)

	// totalSize := decimal.NewFromInt(0)

	// blockCount := len(blocks)

	// missedBlockCount := (firstSlot - uint64(lastSlot)) - uint64(blockCount)

	var prevBlock *types.Eth1Block

	accumulatedBlockTime := decimal.NewFromInt(0)

	for blk := range blocksChan {
		// logger.Infof("analyzing block: %v with: %v transactions", blk.Number, len(blk.Transactions))
		blockCount += 1
		baseFee := decimal.NewFromBigInt(new(big.Int).SetBytes(blk.BaseFee), 0)
		totalBaseFee = totalBaseFee.Add(baseFee)
		totalGasLimit = totalGasLimit.Add(decimal.NewFromInt(int64(blk.GasLimit)))

		if prevBlock != nil {
			accumulatedBlockTime = accumulatedBlockTime.Add(decimal.NewFromInt(prevBlock.Time.AsTime().UnixMicro() - blk.Time.AsTime().UnixMicro()))
		}

		totalBaseBlockReward = totalBaseBlockReward.Add(decimal.NewFromBigInt(utils.Eth1BlockReward(blk.Number, blk.Difficulty), 0))

		for _, tx := range blk.Transactions {
			// for _, itx := range tx.Itx {
			// }
			// blk.Time
			txCount += 1
			maxFee := decimal.NewFromBigInt(new(big.Int).SetBytes(tx.MaxFeePerGas), 0)
			prioFee := decimal.NewFromBigInt(new(big.Int).SetBytes(tx.MaxPriorityFeePerGas), 0)
			gasUsed := decimal.NewFromBigInt(new(big.Int).SetUint64(tx.GasUsed), 0)
			gasPrice := decimal.NewFromBigInt(new(big.Int).SetBytes(tx.GasPrice), 0)

			var tipFee decimal.Decimal
			var txFees decimal.Decimal
			switch tx.Type {
			case 0:
				legacyTxCount += 1
				totalGasPrice = totalGasPrice.Add(gasPrice)
				txFees = gasUsed.Mul(gasPrice)
				tipFee = gasPrice.Sub(baseFee)

			case 1:
				accessListTxCount += 1
				totalGasPrice = totalGasPrice.Add(gasPrice)
				txFees = gasUsed.Mul(gasPrice)
				tipFee = gasPrice.Sub(baseFee)

			case 2:
				// priority fee is capped because the base fee is filled first
				tipFee = decimal.Min(prioFee, maxFee.Sub(baseFee))
				eip1559TxCount += 1
				// totalMinerTips = totalMinerTips.Add(tipFee.Mul(gasUsed))
				txFees = baseFee.Mul(gasUsed).Add(tipFee.Mul(gasUsed))
				totalTxSavings = totalTxSavings.Add(maxFee.Mul(gasUsed).Sub(baseFee.Mul(gasUsed).Add(tipFee.Mul(gasUsed))))

			default:
				logger.Fatalf("error unknown tx type %v hash: %x", tx.Status, tx.Hash)
			}
			totalTxFees = totalTxFees.Add(txFees)

			switch tx.Status {
			case 0:
				failedTxCount += 1
				totalFailedGasUsed = totalFailedGasUsed.Add(gasUsed)
				totalFailedTxFee = totalFailedTxFee.Add(txFees)
			case 1:
				successTxCount += 1
			default:
				logger.Fatalf("error unknown status code %v hash: %x", tx.Status, tx.Hash)
			}
			totalGasUsed = totalGasUsed.Add(gasUsed)
			totalBurned = totalBurned.Add(baseFee.Mul(gasUsed))
			if blk.Number < 12244000 {
				totalTips = totalTips.Add(gasUsed.Mul(gasPrice))
			} else {
				totalTips = totalTips.Add(gasUsed.Mul(tipFee))
			}
		}
		prevBlock = blk
	}

	avgBlockTime := accumulatedBlockTime.Div(decimal.NewFromInt(blockCount - 1))

	logger.Infof("exporting consensus rewards from %v to %v", firstEpoch, lastEpoch)

	// consensus rewards are in Gwei
	totalConsensusRewards := int64(0)

	err = WriterDb.Get(&totalConsensusRewards, "SELECT SUM(COALESCE(cl_rewards_gwei, 0)) FROM validator_stats WHERE day = $1", day)
	if err != nil {
		return fmt.Errorf("error calculating totalConsensusRewards: %w", err)
	}
	logger.Infof("consensus rewards: %v", totalConsensusRewards)

	logger.Infof("Exporting BURNED_FEES %v", totalBurned.String())
	_, err = WriterDb.Exec("INSERT INTO chart_series (time, indicator, value) VALUES ($1, 'BURNED_FEES', $2) on conflict (time, indicator) do update set time = excluded.time, indicator = excluded.indicator, value = excluded.value", dateTrunc, totalBurned.String())
	if err != nil {
		return fmt.Errorf("error calculating BURNED_FEES chart_series: %w", err)
	}

	logger.Infof("Exporting NON_FAILED_TX_GAS_USAGE %v", totalGasUsed.Sub(totalFailedGasUsed).String())
	err = SaveChartSeriesPoint(dateTrunc, "NON_FAILED_TX_GAS_USAGE", totalGasUsed.Sub(totalFailedGasUsed).String())
	if err != nil {
		return fmt.Errorf("error calculating NON_FAILED_TX_GAS_USAGE chart_series: %w", err)
	}
	logger.Infof("Exporting BLOCK_COUNT %v", blockCount)
	err = SaveChartSeriesPoint(dateTrunc, "BLOCK_COUNT", blockCount)
	if err != nil {
		return fmt.Errorf("error calculating BLOCK_COUNT chart_series: %w", err)
	}

	// convert microseconds to seconds
	logger.Infof("Exporting BLOCK_TIME_AVG %v", avgBlockTime.Div(decimal.NewFromInt(1e6)).Abs().String())
	err = SaveChartSeriesPoint(dateTrunc, "BLOCK_TIME_AVG", avgBlockTime.Div(decimal.NewFromInt(1e6)).String())
	if err != nil {
		return fmt.Errorf("error calculating BLOCK_TIME_AVG chart_series: %w", err)
	}
	// convert consensus rewards to gwei
	emission := (totalBaseBlockReward.Add(decimal.NewFromInt(totalConsensusRewards).Mul(decimal.NewFromInt(1000000000))).Add(totalTips)).Sub(totalBurned)
	logger.Infof("Exporting TOTAL_EMISSION %v day emission", emission)

	var lastEmission float64
	err = ReaderDb.Get(&lastEmission, "SELECT value FROM chart_series WHERE indicator = 'TOTAL_EMISSION' AND time < $1 ORDER BY time DESC LIMIT 1", dateTrunc)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("error getting previous value for TOTAL_EMISSION chart_series: %w", err)
	}

	newEmission := decimal.NewFromFloat(lastEmission).Add(emission)
	err = SaveChartSeriesPoint(dateTrunc, "TOTAL_EMISSION", newEmission)
	if err != nil {
		return fmt.Errorf("error calculating TOTAL_EMISSION chart_series: %w", err)
	}

	if totalGasPrice.GreaterThan(decimal.NewFromInt(0)) && decimal.NewFromInt(legacyTxCount).Add(decimal.NewFromInt(accessListTxCount)).GreaterThan(decimal.NewFromInt(0)) {
		logger.Infof("Exporting AVG_GASPRICE")
		_, err = WriterDb.Exec("INSERT INTO chart_series (time, indicator, value) VALUES($1, 'AVG_GASPRICE', $2) on conflict (time, indicator) do update set time = excluded.time, indicator = excluded.indicator, value = excluded.value", dateTrunc, totalGasPrice.Div((decimal.NewFromInt(legacyTxCount).Add(decimal.NewFromInt(accessListTxCount)))).String())
		if err != nil {
			return fmt.Errorf("error calculating AVG_GASPRICE chart_series err: %w", err)
		}
	}

	if txCount > 0 {
		logger.Infof("Exporting AVG_GASUSED %v", totalGasUsed.Div(decimal.NewFromInt(blockCount)).String())
		err = SaveChartSeriesPoint(dateTrunc, "AVG_GASUSED", totalGasUsed.Div(decimal.NewFromInt(blockCount)).String())
		if err != nil {
			return fmt.Errorf("error calculating AVG_GASUSED chart_series: %w", err)
		}
	}

	logger.Infof("Exporting TOTAL_GASUSED %v", totalGasUsed.String())
	err = SaveChartSeriesPoint(dateTrunc, "TOTAL_GASUSED", totalGasUsed.String())
	if err != nil {
		return fmt.Errorf("error calculating TOTAL_GASUSED chart_series: %w", err)
	}

	if blockCount > 0 {
		logger.Infof("Exporting AVG_GASLIMIT %v", totalGasLimit.Div(decimal.NewFromInt(blockCount)))
		err = SaveChartSeriesPoint(dateTrunc, "AVG_GASLIMIT", totalGasLimit.Div(decimal.NewFromInt(blockCount)))
		if err != nil {
			return fmt.Errorf("error calculating AVG_GASLIMIT chart_series: %w", err)
		}
	}

	if !totalGasLimit.IsZero() {
		logger.Infof("Exporting AVG_BLOCK_UTIL %v", totalGasUsed.Div(totalGasLimit).Mul(decimal.NewFromInt(100)))
		err = SaveChartSeriesPoint(dateTrunc, "AVG_BLOCK_UTIL", totalGasUsed.Div(totalGasLimit).Mul(decimal.NewFromInt(100)))
		if err != nil {
			return fmt.Errorf("error calculating AVG_BLOCK_UTIL chart_series: %w", err)
		}
	}

	switch utils.Config.Chain.Config.DepositChainID {
	case 1:
		crowdSale := 72009990.50
		logger.Infof("Exporting MARKET_CAP: %v", newEmission.Div(decimal.NewFromInt(1e18)).Add(decimal.NewFromFloat(crowdSale)).Mul(decimal.NewFromFloat(price.GetEthPrice("USD"))).String())
		err = SaveChartSeriesPoint(dateTrunc, "MARKET_CAP", newEmission.Div(decimal.NewFromInt(1e18)).Add(decimal.NewFromFloat(crowdSale)).Mul(decimal.NewFromFloat(price.GetEthPrice("USD"))).String())
		if err != nil {
			return fmt.Errorf("error calculating MARKET_CAP chart_series: %w", err)
		}
	}

	logger.Infof("Exporting TX_COUNT %v", txCount)
	err = SaveChartSeriesPoint(dateTrunc, "TX_COUNT", txCount)
	if err != nil {
		return fmt.Errorf("error calculating TX_COUNT chart_series: %w", err)
	}

	// Not sure how this is currently possible (where do we store the size, i think this is missing)
	// logger.Infof("Exporting AVG_SIZE %v", totalSize.div)
	// err = SaveChartSeriesPoint(dateTrunc, "AVG_SIZE", totalSize.div)
	// if err != nil {
	// 	return fmt.Errorf("error calculating AVG_SIZE chart_series: %w", err)
	// }

	// logger.Infof("Exporting POWER_CONSUMPTION %v", avgBlockTime.String())
	// err = SaveChartSeriesPoint(dateTrunc, "POWER_CONSUMPTION", avgBlockTime.String())
	// if err != nil {
	// 	return fmt.Errorf("error calculating POWER_CONSUMPTION chart_series: %w", err)
	// }

	// logger.Infof("Exporting NEW_ACCOUNTS %v", avgBlockTime.String())
	// err = SaveChartSeriesPoint(dateTrunc, "NEW_ACCOUNTS", avgBlockTime.String())
	// if err != nil {
	// 	return fmt.Errorf("error calculating NEW_ACCOUNTS chart_series: %w", err)
	// }

	return nil
}

func WriteGraffitiStatisticsForDay(day int64) error {
	if day < 0 {
		logger.Warnf("no graffiti-stats for days before beaconchain")
		return nil
	}

	epochsPerDay := utils.EpochsPerDay()
	firstSlot := uint64(day) * epochsPerDay * utils.Config.Chain.Config.SlotsPerEpoch
	firstSlotOfNextDay := uint64(day+1) * epochsPerDay * utils.Config.Chain.Config.SlotsPerEpoch

	// \x are missed blocks
	// \x0000000000000000000000000000000000000000000000000000000000000000 are empty graffities
	_, err := WriterDb.Exec(`
		insert into graffiti_stats
		select $1::int as day, graffiti, graffiti_text, count(*), count(distinct proposer) as proposer_count
		from blocks 
		where slot >= $2 and slot < $3 and status = '1' and graffiti <> '\x' and graffiti <> '\x0000000000000000000000000000000000000000000000000000000000000000'
		group by day, graffiti, graffiti_text
		on conflict (graffiti, day) do update set
			graffiti       = excluded.graffiti,
			day            = excluded.day,
			graffiti_text  = excluded.graffiti_text,
			count          = excluded.count,
			proposer_count = excluded.proposer_count`, day, firstSlot, firstSlotOfNextDay)
	if err != nil {
		return err
	}

	return nil
}

func checkIfDayIsFinalized(day uint64) error {
	_, lastEpoch := utils.GetFirstAndLastEpochForDay(day)

	latestFinalizedEpoch, err := GetLatestFinalizedEpoch()
	if err != nil {
		return fmt.Errorf("error getting latest finalized epoch from db %w", err)
	}

	if lastEpoch > latestFinalizedEpoch {
		return fmt.Errorf("delaying statistics export as not all epochs for day %v are finalized. Last epoch of the day [%v] last finalized epoch [%v]", day, lastEpoch, latestFinalizedEpoch)
	}

	return nil
}
