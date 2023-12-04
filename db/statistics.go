package db

import (
	"context"
	"database/sql"
	"eth2-exporter/cache"
	"eth2-exporter/metrics"
	"eth2-exporter/price"
	"eth2-exporter/rpc"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
)

func WriteValidatorStatisticsForDay(day uint64, client rpc.Client) error {
	exportStart := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_validator_stats").Observe(time.Since(exportStart).Seconds())
	}()

	firstEpoch, lastEpoch := utils.GetFirstAndLastEpochForDay(day)

	logger.Infof("exporting statistics for day %v (epoch %v to %v)", day, firstEpoch, lastEpoch)

	if err := CheckIfDayIsFinalized(day); err != nil {
		return err
	}

	logger.Infof("getting exported state for day %v", day)

	type Exported struct {
		Status bool `db:"status"`
	}
	exported := Exported{}
	err := WriterDb.Get(&exported, `
		SELECT 
			status
		FROM validator_stats_status 
		WHERE day = $1;
		`, day)

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("error retrieving exported state: %w", err)
	}

	if exported.Status {
		logger.Infof("Skipping day %v as it is already exported", day)
		return nil
	}

	previousDayExported := Exported{}
	err = WriterDb.Get(&previousDayExported, `
		SELECT 
			status
		FROM validator_stats_status 
		WHERE day = $1;
		`, int64(day)-1)

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("error retrieving previous day exported state: %w", err)
	}

	if day > 0 && !previousDayExported.Status {
		return fmt.Errorf("cannot export day %v as day %v has not yet been exported yet", day, int64(day)-1)
	}

	maxValidatorIndex, err := BigtableClient.GetMaxValidatorindexForEpoch(lastEpoch)
	if err != nil {
		return err
	}
	validators := make([]uint64, 0, maxValidatorIndex)
	validatorData := make([]*types.ValidatorStatsTableDbRow, 0, maxValidatorIndex)
	validatorDataMux := &sync.Mutex{}

	logger.Infof("processing statistics for validators 0-%d", maxValidatorIndex)
	for i := uint64(0); i <= maxValidatorIndex; i++ {
		validators = append(validators, i)
		validatorData = append(validatorData, &types.ValidatorStatsTableDbRow{
			ValidatorIndex: i,
			Day:            int64(day),
		})
	}

	g := &errgroup.Group{}

	g.Go(func() error {
		if err := gatherValidatorMissedAttestationsStatisticsForDay(validators, day, validatorData, validatorDataMux); err != nil {
			logger.Error(err)
			return fmt.Errorf("error in GatherValidatorFailedAttestationsStatisticsForDay: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		if err := GatherValidatorSyncDutiesForDay(validators, day, validatorData, validatorDataMux); err != nil {
			return fmt.Errorf("error in GatherValidatorSyncDutiesForDay: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		if err := gatherValidatorDepositWithdrawals(day, validatorData, validatorDataMux); err != nil {
			return fmt.Errorf("error in GatherValidatorDepositWithdrawals: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		if err := gatherValidatorBlockStats(day, validatorData, validatorDataMux); err != nil {
			return fmt.Errorf("error in GatherValidatorBlockStats: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		if err := gatherValidatorBalances(client, day, validatorData, validatorDataMux); err != nil {
			return fmt.Errorf("error in GatherValidatorBalances: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		if err := gatherValidatorElIcome(day, validatorData, validatorDataMux); err != nil {
			return fmt.Errorf("error in GatherValidatorElIcome: %w", err)
		}
		return nil
	})

	var statisticsData1d []*types.ValidatorStatsTableDbRow
	g.Go(func() error {
		var err error
		statisticsData1d, err = GatherStatisticsForDay(int64(day) - 1) // convert to int64 to avoid underflows
		if err != nil {
			return fmt.Errorf("error in GatherPreviousDayStatisticsData: %w", err)
		}
		return nil
	})
	var statisticsData7d []*types.ValidatorStatsTableDbRow
	g.Go(func() error {
		var err error
		statisticsData7d, err = GatherStatisticsForDay(int64(day) - 7) // convert to int64 to avoid underflows
		if err != nil {
			return fmt.Errorf("error in GatherPreviousDayStatisticsData: %w", err)
		}
		return nil
	})
	var statisticsData31d []*types.ValidatorStatsTableDbRow
	g.Go(func() error {
		var err error
		statisticsData31d, err = GatherStatisticsForDay(int64(day) - 31) // convert to int64 to avoid underflows
		if err != nil {
			return fmt.Errorf("error in GatherPreviousDayStatisticsData: %w", err)
		}
		return nil
	})
	var statisticsData365d []*types.ValidatorStatsTableDbRow
	g.Go(func() error {
		var err error
		statisticsData365d, err = GatherStatisticsForDay(int64(day) - 365) // convert to int64 to avoid underflows
		if err != nil {
			return fmt.Errorf("error in GatherPreviousDayStatisticsData: %w", err)
		}
		return nil
	})

	err = g.Wait()
	if err != nil {
		return err
	}

	logger.Infof("statistics data collection for day %v completed", day)

	// calculate cl income data & update totals
	for index, data := range validatorData {

		previousDayData := &types.ValidatorStatsTableDbRow{
			ValidatorIndex: uint64(data.ValidatorIndex),
		}

		if index < len(statisticsData1d) && day > 0 {
			previousDayData = statisticsData1d[index]
		}

		if data.ValidatorIndex != previousDayData.ValidatorIndex {
			return fmt.Errorf("logic error when retrieving previous day data for validator %v (%v wanted, %v retrieved)", index, data.ValidatorIndex, previousDayData.ValidatorIndex)
		}

		// update attestation totals
		data.MissedAttestationsTotal = previousDayData.MissedAttestationsTotal + data.MissedAttestations

		// update sync total
		data.ParticipatedSyncTotal = previousDayData.ParticipatedSyncTotal + data.ParticipatedSync
		data.MissedSyncTotal = previousDayData.MissedSyncTotal + data.MissedSync
		data.OrphanedSyncTotal = previousDayData.OrphanedSyncTotal + data.OrphanedSync

		// calculate cl reward & update totals
		data.ClRewardsGWei = data.EndBalance - previousDayData.EndBalance + data.WithdrawalsAmount - data.DepositsAmount
		data.ClRewardsGWeiTotal = previousDayData.ClRewardsGWeiTotal + data.ClRewardsGWei

		// update el reward total
		data.ElRewardsWeiTotal = previousDayData.ElRewardsWeiTotal.Add(data.ElRewardsWei)

		// update mev reward total
		data.MEVRewardsWeiTotal = previousDayData.MEVRewardsWeiTotal.Add(data.MEVRewardsWei)

		// update withdrawal total
		data.WithdrawalsTotal = previousDayData.WithdrawalsTotal + data.Withdrawals
		data.WithdrawalsAmountTotal = previousDayData.WithdrawalsAmountTotal + data.WithdrawalsAmount

		// update deposits total
		data.DepositsTotal = previousDayData.DepositsTotal + data.Deposits
		data.DepositsAmountTotal = previousDayData.DepositsAmountTotal + data.DepositsAmount

		if statisticsData1d != nil && len(statisticsData1d) > index {
			data.ClPerformance1d = data.ClRewardsGWeiTotal - statisticsData1d[index].ClRewardsGWeiTotal
			data.ElPerformance1d = data.ElRewardsWeiTotal.Sub(statisticsData1d[index].ElRewardsWeiTotal)
			data.MEVPerformance1d = data.MEVRewardsWeiTotal.Sub(statisticsData1d[index].MEVRewardsWeiTotal)
		} else {
			data.ClPerformance1d = data.ClRewardsGWeiTotal
			data.ElPerformance1d = data.ElRewardsWeiTotal
			data.MEVPerformance1d = data.MEVRewardsWeiTotal
		}
		if statisticsData7d != nil && len(statisticsData7d) > index {
			data.ClPerformance7d = data.ClRewardsGWeiTotal - statisticsData7d[index].ClRewardsGWeiTotal
			data.ElPerformance7d = data.ElRewardsWeiTotal.Sub(statisticsData7d[index].ElRewardsWeiTotal)
			data.MEVPerformance7d = data.MEVRewardsWeiTotal.Sub(statisticsData7d[index].MEVRewardsWeiTotal)
		} else {
			data.ClPerformance7d = data.ClRewardsGWeiTotal
			data.ElPerformance7d = data.ElRewardsWeiTotal
			data.MEVPerformance7d = data.MEVRewardsWeiTotal
		}
		if statisticsData31d != nil && len(statisticsData31d) > index {
			data.ClPerformance31d = data.ClRewardsGWeiTotal - statisticsData31d[index].ClRewardsGWeiTotal
			data.ElPerformance31d = data.ElRewardsWeiTotal.Sub(statisticsData31d[index].ElRewardsWeiTotal)
			data.MEVPerformance31d = data.MEVRewardsWeiTotal.Sub(statisticsData31d[index].MEVRewardsWeiTotal)
		} else {
			data.ClPerformance31d = data.ClRewardsGWeiTotal
			data.ElPerformance31d = data.ElRewardsWeiTotal
			data.MEVPerformance31d = data.MEVRewardsWeiTotal
		}
		if statisticsData365d != nil && len(statisticsData365d) > index {
			data.ClPerformance365d = data.ClRewardsGWeiTotal - statisticsData365d[index].ClRewardsGWeiTotal
			data.ElPerformance365d = data.ElRewardsWeiTotal.Sub(statisticsData365d[index].ElRewardsWeiTotal)
			data.MEVPerformance365d = data.MEVRewardsWeiTotal.Sub(statisticsData365d[index].MEVRewardsWeiTotal)
		} else {
			data.ClPerformance365d = data.ClRewardsGWeiTotal
			data.ElPerformance365d = data.ElRewardsWeiTotal
			data.MEVPerformance365d = data.MEVRewardsWeiTotal
		}
	}

	conn, err := WriterDb.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("error retrieving raw sql connection: %w", err)
	}
	defer conn.Close()

	err = conn.Raw(func(driverConn interface{}) error {
		conn := driverConn.(*stdlib.Conn).Conn()

		pgxdecimal.Register(conn.TypeMap())
		tx, err := conn.Begin(context.Background())

		if err != nil {
			return err
		}

		defer tx.Rollback(context.Background())

		logger.Infof("bulk inserting statistics data into the validator_stats table")
		_, err = tx.Exec(context.Background(), "DELETE FROM validator_stats WHERE day = $1", day)
		if err != nil {
			return err
		}

		_, err = tx.CopyFrom(context.Background(), pgx.Identifier{"validator_stats"}, []string{
			"validatorindex",
			"day",
			"start_balance",
			"end_balance",
			"min_balance",
			"max_balance",
			"start_effective_balance",
			"end_effective_balance",
			"min_effective_balance",
			"max_effective_balance",
			"missed_attestations",
			"missed_attestations_total",
			"orphaned_attestations",
			"participated_sync",
			"participated_sync_total",
			"missed_sync",
			"missed_sync_total",
			"orphaned_sync",
			"orphaned_sync_total",
			"proposed_blocks",
			"missed_blocks",
			"orphaned_blocks",
			"attester_slashings",
			"proposer_slashings",
			"deposits",
			"deposits_total",
			"deposits_amount",
			"deposits_amount_total",
			"withdrawals",
			"withdrawals_total",
			"withdrawals_amount",
			"withdrawals_amount_total",
			"cl_rewards_gwei",
			"cl_rewards_gwei_total",
			"el_rewards_wei",
			"el_rewards_wei_total",
			"mev_rewards_wei",
			"mev_rewards_wei_total",
		}, pgx.CopyFromSlice(len(validatorData), func(i int) ([]interface{}, error) {
			return []interface{}{
				validatorData[i].ValidatorIndex,
				validatorData[i].Day,
				validatorData[i].StartBalance,
				validatorData[i].EndBalance,
				validatorData[i].MinBalance,
				validatorData[i].MaxBalance,
				validatorData[i].StartEffectiveBalance,
				validatorData[i].EndEffectiveBalance,
				validatorData[i].MinEffectiveBalance,
				validatorData[i].MaxEffectiveBalance,
				validatorData[i].MissedAttestations,
				validatorData[i].MissedAttestationsTotal,
				validatorData[i].OrphanedAttestations,
				validatorData[i].ParticipatedSync,
				validatorData[i].ParticipatedSyncTotal,
				validatorData[i].MissedSync,
				validatorData[i].MissedSyncTotal,
				validatorData[i].OrphanedSync,
				validatorData[i].OrphanedSyncTotal,
				validatorData[i].ProposedBlocks,
				validatorData[i].MissedBlocks,
				validatorData[i].OrphanedBlocks,
				validatorData[i].AttesterSlashings,
				validatorData[i].ProposerSlashing,
				validatorData[i].Deposits,
				validatorData[i].DepositsTotal,
				validatorData[i].DepositsAmount,
				validatorData[i].DepositsAmountTotal,
				validatorData[i].Withdrawals,
				validatorData[i].WithdrawalsTotal,
				validatorData[i].WithdrawalsAmount,
				validatorData[i].WithdrawalsAmountTotal,
				validatorData[i].ClRewardsGWei,
				validatorData[i].ClRewardsGWeiTotal,
				validatorData[i].ElRewardsWei,
				validatorData[i].ElRewardsWeiTotal,
				validatorData[i].MEVRewardsWei,
				validatorData[i].MEVRewardsWeiTotal,
			}, nil
		}))

		if err != nil {
			return err
		}

		lastExportedStatsDay, err := GetLastExportedStatisticDay()
		if err != nil && err != ErrNoStats {
			return fmt.Errorf("error retrieving last exported statistics day: %w", err)
		}

		if day > lastExportedStatsDay {
			logger.Infof("updating validator_performance table")

			logger.Infof("deleting validator_performance table contents")

			_, err = tx.Exec(context.Background(), "TRUNCATE validator_performance")
			if err != nil {
				return err
			}
			logger.Infof("bulk loading new validator_performance table contents")

			_, err = tx.CopyFrom(context.Background(), pgx.Identifier{"validator_performance"}, []string{
				"validatorindex",
				"balance",
				"rank7d",

				"cl_performance_1d",
				"cl_performance_7d",
				"cl_performance_31d",
				"cl_performance_365d",
				"cl_performance_total",

				"el_performance_1d",
				"el_performance_7d",
				"el_performance_31d",
				"el_performance_365d",
				"el_performance_total",

				"mev_performance_1d",
				"mev_performance_7d",
				"mev_performance_31d",
				"mev_performance_365d",
				"mev_performance_total",
			}, pgx.CopyFromSlice(len(validatorData), func(i int) ([]interface{}, error) {
				return []interface{}{
					validatorData[i].ValidatorIndex,
					validatorData[i].EndBalance,
					0,

					validatorData[i].ClPerformance1d,
					validatorData[i].ClPerformance7d,
					validatorData[i].ClPerformance31d,
					validatorData[i].ClPerformance365d,
					validatorData[i].ClRewardsGWeiTotal,

					validatorData[i].ElPerformance1d,
					validatorData[i].ElPerformance7d,
					validatorData[i].ElPerformance31d,
					validatorData[i].ElPerformance365d,
					validatorData[i].ElRewardsWeiTotal,

					validatorData[i].MEVPerformance1d,
					validatorData[i].MEVPerformance7d,
					validatorData[i].MEVPerformance31d,
					validatorData[i].MEVPerformance365d,
					validatorData[i].MEVRewardsWeiTotal,
				}, nil
			}))

			if err != nil {
				return fmt.Errorf("error writing to validator_performance table: %w", err)
			}

			logger.Infof("populate validator_performance rank7d")
			_, err = tx.Exec(context.Background(), `
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
		} else {
			logger.Infof("skipping total performance export as last exported day (%v) is greater than the exported day (%v)", lastExportedStatsDay, day)
		}

		logger.Infof("marking day %v as exported", day)
		if err := WriteValidatorStatsExported(day, tx); err != nil {
			return fmt.Errorf("error in WriteValidatorStatsExported: %w", err)
		}

		err = tx.Commit(context.Background())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error during statistics data insert: %w", err)
	}

	logger.Infof("batch insert of statistics data completed")

	logger.Infof("statistics export of day %v completed, took %v", day, time.Since(exportStart))
	return nil
}

func WriteValidatorStatsExported(day uint64, tx pgx.Tx) error {

	start := time.Now()

	logger.Infof("marking day export as completed in the validator_stats_status table for day %v", day)
	_, err := tx.Exec(context.Background(), `
		INSERT INTO validator_stats_status (day, status,failed_attestations_exported,sync_duties_exported,withdrawals_deposits_exported,balance_exported,cl_rewards_exported,el_rewards_exported,total_performance_exported,block_stats_exported,total_accumulation_exported)
		VALUES ($1, true, true, true, true,true,true,true,true,true,true)
		ON CONFLICT (day) DO UPDATE
		SET status = true,
		failed_attestations_exported = true,
		sync_duties_exported = true,
		withdrawals_deposits_exported = true,
		balance_exported = true,
		cl_rewards_exported = true,
		el_rewards_exported = true,
		total_performance_exported = true,
		block_stats_exported = true,
		total_accumulation_exported = true;
		`, day)
	if err != nil {
		return fmt.Errorf("error marking day export as completed in the validator_stats_status table for day %v: %w", day, err)
	}
	logger.Infof("marking completed, took %v", time.Since(start))

	return nil
}

func WriteValidatorTotalPerformance(day uint64, tx pgx.Tx) error {
	exportStart := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_validator_total_performance_stats").Observe(time.Since(exportStart).Seconds())
	}()

	start := time.Now()

	logger.Infof("exporting total performance stats")

	_, err := tx.Exec(context.Background(), `insert into validator_performance (
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
					where vs_now.day = $1
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
			;`, day, int64(day)-1, int64(day)-7, int64(day)-31, int64(day)-365)

	if err != nil {
		return fmt.Errorf("error inserting performance into validator_performance for day [%v]: %w", day, err)
	}

	logger.Infof("export completed, took %v", time.Since(start))

	start = time.Now()
	logger.Infof("populate validator_performance rank7d")

	_, err = tx.Exec(context.Background(), `
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

	logger.Infof("total performance statistics export of day %v completed, took %v", day, time.Since(exportStart))
	return nil
}

func gatherValidatorBlockStats(day uint64, data []*types.ValidatorStatsTableDbRow, mux *sync.Mutex) error {
	exportStart := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_validator_block_stats").Observe(time.Since(exportStart).Seconds())
	}()

	firstEpoch, lastEpoch := utils.GetFirstAndLastEpochForDay(day)
	logger := logger.WithFields(logrus.Fields{
		"day":        day,
		"firstEpoch": firstEpoch,
		"lastEpoch":  lastEpoch,
	})

	type resRowBlocks struct {
		ValidatorIndex uint64 `db:"validatorindex"`
		ProposedBlocks uint64 `db:"proposed_blocks"`
		MissedBlocks   uint64 `db:"missed_blocks"`
		OrphanedBlocks uint64 `db:"orphaned_blocks"`
	}
	resBlocks := make([]*resRowBlocks, 0, 1024)

	logger.Infof("gathering proposed_blocks, missed_blocks and orphaned_blocks statistics")
	err := WriterDb.Select(&resBlocks, `select proposer AS validatorindex, sum(case when status = '1' then 1 else 0 end) AS proposed_blocks, sum(case when status = '2' then 1 else 0 end) AS missed_blocks, sum(case when status = '3' then 1 else 0 end) AS orphaned_blocks
			from blocks
			where epoch >= $1 and epoch <= $2 and proposer != $3
			group by proposer
		;`,
		firstEpoch, lastEpoch, MaxSqlInteger)
	if err != nil {
		return fmt.Errorf("error retrieving blocks for day [%v], firstEpoch [%v] and lastEpoch [%v]: %w", day, firstEpoch, lastEpoch, err)
	}

	mux.Lock()
	for _, r := range resBlocks {
		data[r.ValidatorIndex].ProposedBlocks = int64(r.ProposedBlocks)
		data[r.ValidatorIndex].MissedBlocks = int64(r.MissedBlocks)
		data[r.ValidatorIndex].OrphanedBlocks = int64(r.OrphanedBlocks)
	}
	mux.Unlock()

	type resRowSlashings struct {
		ValidatorIndex    uint64 `db:"validatorindex"`
		AttesterSlashings uint64 `db:"attester_slashings"`
		ProposerSlashing  uint64 `db:"proposer_slashings"`
	}
	resSlashings := make([]*resRowSlashings, 0, 1024)

	logger.Infof("gathering attester_slashings and proposer_slashings statistics")
	err = WriterDb.Select(&resSlashings, `
			select proposer AS validatorindex, sum(attesterslashingscount) AS attester_slashings, sum(proposerslashingscount) AS proposer_slashings
			from blocks
			where epoch >= $1 and epoch <= $2 and status = '1' and proposer != $3
			group by proposer;
		`,
		firstEpoch, lastEpoch, MaxSqlInteger)
	if err != nil {
		return fmt.Errorf("error retrieving slashings for day [%v], firstEpoch [%v] and lastEpoch [%v]: %w", day, firstEpoch, lastEpoch, err)
	}

	mux.Lock()
	for _, r := range resSlashings {
		data[r.ValidatorIndex].AttesterSlashings = int64(r.AttesterSlashings)
		data[r.ValidatorIndex].ProposerSlashing = int64(r.ProposerSlashing)
	}
	mux.Unlock()

	logger.Infof("gathering block statistics completed, took %v", time.Since(exportStart))
	return nil
}

func gatherValidatorElIcome(day uint64, data []*types.ValidatorStatsTableDbRow, mux *sync.Mutex) error {
	exportStart := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_validator_el_income_stats").Observe(time.Since(exportStart).Seconds())
	}()

	firstEpoch, lastEpoch := utils.GetFirstAndLastEpochForDay(day)
	logger := logger.WithFields(logrus.Fields{
		"day":        day,
		"firstEpoch": firstEpoch,
		"lastEpoch":  lastEpoch,
	})

	logger.Infof("gathering mev & el rewards")

	type Container struct {
		Slot            uint64 `db:"slot"`
		ExecBlockNumber uint64 `db:"exec_block_number"`
		Proposer        uint64 `db:"proposer"`
		TxFeeReward     *big.Int
		MevReward       *big.Int
	}

	blocks := make([]*Container, 0)
	blocksMap := make(map[uint64]*Container)

	err := WriterDb.Select(&blocks, "SELECT slot, exec_block_number, proposer FROM blocks WHERE epoch >= $1 AND epoch <= $2 AND exec_block_number > 0 AND status = '1'", firstEpoch, lastEpoch)
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

	mux.Lock()
	for proposer, r := range proposerRewards {
		data[proposer].ElRewardsWei = decimal.NewFromBigInt(r.TxFeeReward, 0)
		data[proposer].MEVRewardsWei = decimal.NewFromBigInt(r.MevReward, 0)
	}
	mux.Unlock()

	logger.Infof("gathering mev & el rewards statistics completed, took %v", time.Since(exportStart))
	return nil
}

func gatherValidatorBalances(client rpc.Client, day uint64, data []*types.ValidatorStatsTableDbRow, mux *sync.Mutex) error {
	exportStart := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_validator_balances_stats").Observe(time.Since(exportStart).Seconds())
	}()

	firstEpoch, lastEpoch := utils.GetFirstAndLastEpochForDay(day)
	logger := logger.WithFields(logrus.Fields{
		"day":        day,
		"firstEpoch": firstEpoch,
		"lastEpoch":  lastEpoch,
	})

	logger.Infof("gathering balance statistics")
	firstEpochBalances, err := client.GetValidatorState(firstEpoch)
	if err != nil {
		return fmt.Errorf("error in GetValidatorBalanceStatistics for firstEpoch [%v] and lastEpoch [%v]: %w", firstEpoch, lastEpoch, err)
	}
	logger.Infof("retrieved balances for first epoch of day")
	lastEpochBalances, err := client.GetValidatorState(lastEpoch)
	if err != nil {
		return fmt.Errorf("error in GetValidatorBalanceStatistics for firstEpoch [%v] and lastEpoch [%v]: %w", firstEpoch, lastEpoch, err)
	}
	logger.Infof("retrieved balances for last epoch of day")

	mux.Lock()
	for _, stat := range firstEpochBalances.Data {
		data[stat.Index].StartBalance = int64(stat.Balance)
		data[stat.Index].StartEffectiveBalance = int64(stat.Validator.EffectiveBalance)
	}
	for _, stat := range lastEpochBalances.Data {
		data[stat.Index].EndBalance = int64(stat.Balance)
		data[stat.Index].EndEffectiveBalance = int64(stat.Validator.EffectiveBalance)
	}
	mux.Unlock()

	logger.Infof("gathering balance statistics completed, took %v", time.Since(exportStart))
	return nil
}

func gatherValidatorDepositWithdrawals(day uint64, data []*types.ValidatorStatsTableDbRow, mux *sync.Mutex) error {
	exportStart := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_validator_deposit_withdrawal_stats").Observe(time.Since(exportStart).Seconds())
	}()

	// The end_balance of a day is the balance after the first slot of the last epoch of that day.
	// Therefore the last 31 slots of the day are not included in the end_balance of that day.
	// Since our income calculation is base on subtracting end_balances the deposits and withdrawals that happen during those slots must be added to the next day instead.
	firstSlot := uint64(0)
	if day > 0 {
		firstSlot = utils.GetLastBalanceInfoSlotForDay(day-1) + 1
	}
	lastSlot := utils.GetLastBalanceInfoSlotForDay(day)

	logger := logger.WithFields(logrus.Fields{
		"day":       day,
		"firstSlot": firstSlot,
		"lastSlot":  lastSlot,
	})

	logger.Infof("gathering deposits + withdrawals")

	type resRowDeposits struct {
		ValidatorIndex uint64 `db:"validatorindex"`
		Deposits       uint64 `db:"deposits"`
		DepositsAmount uint64 `db:"deposits_amount"`
	}
	resDeposits := make([]*resRowDeposits, 0, 1024)
	depositsQry := `
			select validators.validatorindex, count(*) AS deposits, sum(amount) AS deposits_amount
			from blocks_deposits
			inner join validators on blocks_deposits.publickey = validators.pubkey
			inner join blocks on blocks_deposits.block_root = blocks.blockroot
			where blocks.slot >= $1 and blocks.slot <= $2 and (blocks.status = '1' OR blocks.slot = 0) and blocks_deposits.valid_signature
			group by validators.validatorindex`

	err := WriterDb.Select(&resDeposits, depositsQry, firstSlot, lastSlot)
	if err != nil {
		return fmt.Errorf("error retrieving deposits for day [%v], firstSlot [%v] and lastSlot [%v]: %w", day, firstSlot, lastSlot, err)
	}

	mux.Lock()
	for _, r := range resDeposits {
		data[r.ValidatorIndex].Deposits = int64(r.Deposits)
		data[r.ValidatorIndex].DepositsAmount = int64(r.DepositsAmount)
	}
	mux.Unlock()

	type resRowWithdrawals struct {
		ValidatorIndex    uint64 `db:"validatorindex"`
		Withdrawals       uint64 `db:"withdrawals"`
		WithdrawalsAmount uint64 `db:"withdrawals_amount"`
	}
	resWithdrawals := make([]*resRowWithdrawals, 0, 1024)

	withdrawalsQuery := `select validatorindex, count(*) AS withdrawals, sum(amount) AS withdrawals_amount
			from blocks_withdrawals
			inner join blocks on blocks_withdrawals.block_root = blocks.blockroot
			where block_slot >= $1 and block_slot <= $2 and blocks.status = '1'
			group by validatorindex;`
	err = WriterDb.Select(&resWithdrawals, withdrawalsQuery, firstSlot, lastSlot)
	if err != nil {
		return fmt.Errorf("error retrieving withdrawals for day [%v], firstSlot [%v] and lastSlot [%v]: %w", day, firstSlot, lastSlot, err)
	}

	mux.Lock()
	for _, r := range resWithdrawals {
		data[r.ValidatorIndex].Withdrawals = int64(r.Withdrawals)
		data[r.ValidatorIndex].WithdrawalsAmount = int64(r.WithdrawalsAmount)
	}
	mux.Unlock()

	logger.Infof("gathering deposits + withdrawals completed, took %v", time.Since(exportStart))
	return nil
}

func GatherValidatorSyncDutiesForDay(validators []uint64, day uint64, data []*types.ValidatorStatsTableDbRow, mux *sync.Mutex) error {
	exportStart := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_validator_sync_stats").Observe(time.Since(exportStart).Seconds())
	}()

	firstEpoch, lastEpoch := utils.GetFirstAndLastEpochForDay(day)
	if firstEpoch < utils.Config.Chain.ClConfig.AltairForkEpoch && lastEpoch > utils.Config.Chain.ClConfig.AltairForkEpoch {
		firstEpoch = utils.Config.Chain.ClConfig.AltairForkEpoch
	} else if lastEpoch < utils.Config.Chain.ClConfig.AltairForkEpoch {
		logger.Infof("day %v is pre-altair, skipping sync committee export", day)
		return nil
	}
	logger := logger.WithFields(logrus.Fields{
		"day":         day,
		"firstEpoch":  firstEpoch,
		"lastEpoch":   lastEpoch,
		"startPeriod": utils.SyncPeriodOfEpoch(firstEpoch),
		"endPeriod":   utils.SyncPeriodOfEpoch(lastEpoch),
	})
	logger.Infof("gathering sync duties")

	//map to hold the sync committee members for a given period
	syncCommittees := make(map[types.SyncCommitteePeriod]map[types.CommitteeIndex]types.ValidatorIndex)

	// iterate over all proposed slots of the statistics day
	rows, err := ReaderDb.Query("SELECT slot, syncaggregate_bits FROM blocks WHERE epoch >= $1 AND epoch <= $2 AND status = '1'", firstEpoch, lastEpoch)

	if err != nil {
		return fmt.Errorf("error retrieving blocks for sync statistics: %w", err)
	}

	proposedSlots := make(map[types.Slot][]byte)
	for rows.Next() {
		var slot types.Slot
		var bits []byte

		err := rows.Scan(&slot, &bits)
		if err != nil {
			rows.Close()
			return fmt.Errorf("error scanning row for sync statistics: %w", err)
		}

		proposedSlots[slot] = bits
	}
	rows.Close()

	for slot := firstEpoch * utils.Config.Chain.ClConfig.SlotsPerEpoch; slot <= ((lastEpoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch)-1; slot++ {
		period := utils.SyncPeriodOfEpoch(utils.EpochOfSlot(uint64(slot)))

		committee := syncCommittees[types.SyncCommitteePeriod(period)]
		if committee == nil {
			committeeRows := []struct {
				Period         types.SyncCommitteePeriod
				ValidatorIndex types.ValidatorIndex
				CommitteeIndex types.CommitteeIndex
			}{}

			err := ReaderDb.Select(&committeeRows, "SELECT period, validatorindex, committeeindex FROM sync_committees WHERE period = $1", period)
			if err != nil {
				return fmt.Errorf("error retrieving sync period committees of period %v for sync statistics: %w", period, err)
			}

			syncCommittees[types.SyncCommitteePeriod(period)] = make(map[types.CommitteeIndex]types.ValidatorIndex)
			for _, row := range committeeRows {
				syncCommittees[types.SyncCommitteePeriod(period)][row.CommitteeIndex] = row.ValidatorIndex
			}
			committee = syncCommittees[types.SyncCommitteePeriod(period)]
			logger.Infof("retrieved committee members for period %v", period)
		}

		mux.Lock()
		bits := proposedSlots[types.Slot(slot)]
		for i := 0; i < len(committee); i++ {

			validator := committee[types.CommitteeIndex(i)]

			if len(bits) == 0 { // slot is empty
				data[validator].MissedSync++
			} else {
				participated := utils.BitAtVector(bits, i)
				if participated {
					data[validator].ParticipatedSync++
				} else {
					data[validator].MissedSync++
				}
			}
		}
		mux.Unlock()
	}

	logger.Infof("gathering sync duties completed, took %v", time.Since(exportStart))

	return nil
}

func gatherValidatorMissedAttestationsStatisticsForDay(validators []uint64, day uint64, data []*types.ValidatorStatsTableDbRow, mux *sync.Mutex) error {
	exportStart := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_validator_failed_att_stats").Observe(time.Since(exportStart).Seconds())
	}()

	firstEpoch, lastEpoch := utils.GetFirstAndLastEpochForDay(day)
	logger := logger.WithFields(logrus.Fields{
		"day":        day,
		"firstEpoch": firstEpoch,
		"lastEpoch":  lastEpoch,
	})

	start := time.Now()

	logger.Infof("gathering missed attestations statistics")

	// first retrieve activation & exit epoch for all validators
	activityData := []struct {
		ActivationEpoch types.Epoch
		ExitEpoch       types.Epoch
	}{}

	err := ReaderDb.Select(&activityData, "SELECT activationepoch, exitepoch FROM validators ORDER BY validatorindex;")
	if err != nil {
		return fmt.Errorf("error retrieving activation & exit epoch for validators: %w", err)
	}

	// next retrieve all attestation data from the db

	firstSlot := firstEpoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
	lastSlot := ((lastEpoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch - 1)
	lastQuerySlot := ((lastEpoch+2)*utils.Config.Chain.ClConfig.SlotsPerEpoch - 1)

	rows, err := ReaderDb.Query(`SELECT 
	blocks_attestations.slot, 
	validators 
	FROM blocks_attestations 
	LEFT JOIN blocks ON blocks_attestations.block_root = blocks.blockroot WHERE
	blocks_attestations.block_slot >= $1 AND blocks_attestations.block_slot <= $2 AND blocks.status = '1' ORDER BY block_slot`, firstSlot, lastQuerySlot)
	if err != nil {
		return fmt.Errorf("error retrieving attestation data from the db: %w", err)
	}
	defer rows.Close()

	epochParticipation := make(map[types.Epoch]map[types.ValidatorIndex]bool)
	for rows.Next() {
		var slot types.Slot
		var attestingValidators pq.Int64Array

		err := rows.Scan(&slot, &attestingValidators)
		if err != nil {
			logger.Error(err)
			return fmt.Errorf("error scanning attestation data: %w", err)
		}

		if slot < types.Slot(firstSlot) || slot > types.Slot(lastSlot) {
			continue
		}

		epoch := types.Epoch(utils.EpochOfSlot(uint64(slot)))

		participation := epochParticipation[epoch]

		if participation == nil {
			epochParticipation[epoch] = make(map[types.ValidatorIndex]bool)

			// logger.Infof("seeding validator duties for epoch %v", epoch)
			for _, validator := range validators {
				if activityData[validator].ActivationEpoch <= epoch && epoch < activityData[validator].ExitEpoch {
					epochParticipation[epoch][types.ValidatorIndex(validator)] = false
				}
			}

			participation = epochParticipation[epoch]
		}

		for _, validator := range attestingValidators {
			participation[types.ValidatorIndex(validator)] = true
		}

		if len(epochParticipation) == 3 { // we have data for 3 epochs now available, which means data for the earliest epoch is now complete (takes data of two epochs)
			completedEpoch := epoch - 2

			// logger.Infof("processing data for completed epoch %v", completedEpoch)
			completedEpochData := epochParticipation[completedEpoch]

			if completedEpochData == nil {
				return fmt.Errorf("logic error, did not retrieve data for epoch %v", completedEpoch)
			}

			mux.Lock()
			for validator, participated := range completedEpochData {
				if !participated {
					data[validator].MissedAttestations++
				}
			}
			mux.Unlock()

			delete(epochParticipation, completedEpoch) // delete the completed epoch to preserve memory
		}
	}

	// process the remaining epochs
	for epoch, participation := range epochParticipation {
		if epoch > types.Epoch(lastEpoch) {
			continue
		}
		mux.Lock()
		for validator, participated := range participation {
			if !participated {
				data[validator].MissedAttestations++
			}
		}
		mux.Unlock()
	}

	// mux.Lock()
	// for i := 0; i < 100; i++ {
	// 	logger.Infof("validator %v has %v missed attestations", i, data[i].MissedAttestations)
	// }
	// mux.Unlock()
	logrus.Infof("gathering missed attestations completed, took %v", time.Since(start))

	return nil
}

func GatherStatisticsForDay(day int64) ([]*types.ValidatorStatsTableDbRow, error) {

	if day < 0 {
		return nil, nil
	}

	logger := logger.WithFields(logrus.Fields{
		"day": day,
	})

	start := time.Now()

	logger.Infof("gathering existing statistics for day %v", day)

	ret := make([]*types.ValidatorStatsTableDbRow, 0)

	err := WriterDb.Select(&ret, `SELECT 
		validatorindex, 
		day, 
		COALESCE(start_balance, 0) AS start_balance,
		COALESCE(end_balance, 0) AS end_balance,
		COALESCE(min_balance, 0) AS min_balance,
		COALESCE(max_balance, 0) AS max_balance,
		COALESCE(start_effective_balance, 0) AS start_effective_balance,
		COALESCE(end_effective_balance, 0) AS end_effective_balance,
		COALESCE(min_effective_balance, 0) AS min_effective_balance,
		COALESCE(max_effective_balance, 0) AS max_effective_balance,
		COALESCE(missed_attestations, 0) AS missed_attestations,
		COALESCE(missed_attestations_total, 0) AS missed_attestations_total,
		COALESCE(orphaned_attestations, 0) AS orphaned_attestations,
		COALESCE(participated_sync, 0) AS participated_sync,
		COALESCE(participated_sync_total, 0) AS participated_sync_total,
		COALESCE(missed_sync, 0) AS missed_sync,
		COALESCE(missed_sync_total, 0) AS missed_sync_total,
		COALESCE(orphaned_sync, 0) AS orphaned_sync,
		COALESCE(orphaned_sync_total, 0) AS orphaned_sync_total,
		COALESCE(proposed_blocks, 0) AS proposed_blocks,
		COALESCE(missed_blocks, 0) AS missed_blocks,
		COALESCE(orphaned_blocks, 0) AS orphaned_blocks,
		COALESCE(attester_slashings, 0) AS attester_slashings,
		COALESCE(proposer_slashings, 0) AS proposer_slashings,
		COALESCE(deposits, 0) AS deposits,
		COALESCE(deposits_total, 0) AS deposits_total,
		COALESCE(deposits_amount, 0) AS deposits_amount,
		COALESCE(deposits_amount_total, 0) AS deposits_amount_total,
		COALESCE(withdrawals, 0) AS withdrawals,
		COALESCE(withdrawals_total, 0) AS withdrawals_total,
		COALESCE(withdrawals_amount, 0) AS withdrawals_amount,
		COALESCE(withdrawals_amount_total, 0) AS withdrawals_amount_total,
		COALESCE(cl_rewards_gwei, 0) AS cl_rewards_gwei,
		COALESCE(cl_rewards_gwei_total, 0) AS cl_rewards_gwei_total,
		COALESCE(el_rewards_wei, 0) AS el_rewards_wei,
		COALESCE(el_rewards_wei_total, 0) AS el_rewards_wei_total,
		COALESCE(mev_rewards_wei, 0) AS mev_rewards_wei,
		COALESCE(mev_rewards_wei_total, 0) AS mev_rewards_wei_total
	 from validator_stats WHERE day = $1 ORDER BY validatorindex
	`, day)

	if err != nil {
		return nil, fmt.Errorf("error statistics for day %v data: %w", day, err)
	}

	logrus.Infof("gathering existing statistics for day %v completed, took %v", day, time.Since(start))
	return ret, nil
}

func GetValidatorIncomeHistoryChart(validatorIndices []uint64, currency string, lastFinalizedEpoch uint64, lowerBoundDay uint64) ([]*types.ChartDataPoint, error) {
	incomeHistory, err := GetValidatorIncomeHistory(validatorIndices, lowerBoundDay, 0, lastFinalizedEpoch)
	if err != nil {
		return nil, err
	}
	var clRewardsSeries = make([]*types.ChartDataPoint, len(incomeHistory))

	p := price.GetPrice(utils.Config.Frontend.ClCurrency, currency)

	for i := 0; i < len(incomeHistory); i++ {
		color := "#7cb5ec"
		if incomeHistory[i].ClRewards < 0 {
			color = "#f7a35c"
		}
		balanceTs := utils.DayToTime(incomeHistory[i].Day)
		clRewardsSeries[i] = &types.ChartDataPoint{X: float64(balanceTs.Unix() * 1000), Y: p * (float64(incomeHistory[i].ClRewards)) / float64(utils.Config.Frontend.ClCurrencyDivisor), Color: color}
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

	cacheDur := time.Second * time.Duration(utils.Config.Chain.ClConfig.SecondsPerSlot*utils.Config.Chain.ClConfig.SlotsPerEpoch+10) // updates every epoch, keep 10sec longer
	cacheKey := fmt.Sprintf("%d:validatorIncomeHistory:%d:%d:%d:%s", utils.Config.Chain.ClConfig.DepositChainID, lowerBoundDay, upperBoundDay, lastFinalizedEpoch, strings.Join(validatorIndicesStr, ","))
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
		lastDay := int64(0)
		if len(result) > 0 {
			lastDay = int64(result[len(result)-1].Day)
		} else {
			lastDayDb, err := GetLastExportedStatisticDay()
			if err == nil {
				lastDay = int64(lastDayDb)
			} else if err == ErrNoStats {
				lastDay = -1
			} else {
				return nil, err
			}
		}

		currentDay := lastDay + 1
		firstSlot := uint64(0)
		if lastDay > -1 {
			firstSlot = utils.GetLastBalanceInfoSlotForDay(uint64(lastDay)) + 1
		}
		lastSlot := lastFinalizedEpoch * utils.Config.Chain.ClConfig.SlotsPerEpoch

		totalBalance := uint64(0)

		g := errgroup.Group{}
		g.Go(func() error {
			latestBalances, err := BigtableClient.GetValidatorBalanceHistory(validatorIndices, lastFinalizedEpoch, lastFinalizedEpoch)
			if err != nil {
				logger.Errorf("error in GetValidatorIncomeHistory calling BigtableClient.GetValidatorBalanceHistory: %v", err)
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

			if lastDay < 0 {
				return GetValidatorActivationBalance(validatorIndices, &lastBalance)
			} else {
				return GetValidatorBalanceForDay(validatorIndices, uint64(lastDay), &lastBalance)
			}
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
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

	epochOffset := firstSlot % utils.Config.Chain.ClConfig.SlotsPerEpoch
	firstSlot = firstSlot - epochOffset
	firstEpoch := firstSlot / utils.Config.Chain.ClConfig.SlotsPerEpoch
	// exclusive slot
	lastSlot := int64(firstSlot) + int64(epochsPerDay*utils.Config.Chain.ClConfig.SlotsPerEpoch)
	if firstSlot == 0 {
		nextDateTrunc := time.Date(startDate.Year(), startDate.Month(), startDate.Day()+1, 0, 0, 0, 0, time.UTC)
		lastSlot = int64(utils.TimeToFirstSlotOfEpoch(uint64(nextDateTrunc.Unix())))
	}
	lastEpoch := lastSlot / int64(utils.Config.Chain.ClConfig.SlotsPerEpoch)
	lastSlot = lastEpoch * int64(utils.Config.Chain.ClConfig.SlotsPerEpoch)

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
	if utils.Config.Chain.ClConfig.DepositChainID != 1 {
		// logger.Warnf("not writing chart_series for execution: chainId != 1: %v", utils.Config.Chain.ClConfig.DepositChainID)
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
	firstEpoch := firstSlot / utils.Config.Chain.ClConfig.SlotsPerEpoch
	// exclusive slot
	lastSlot := int64(firstSlot) + int64(epochsPerDay*utils.Config.Chain.ClConfig.SlotsPerEpoch)
	// The first day is not a whole day, so we take the first slot from the next day as lastSlot
	if firstSlot == 0 {
		nextDateTrunc := time.Date(startDate.Year(), startDate.Month(), startDate.Day()+1, 0, 0, 0, 0, time.UTC)
		lastSlot = int64(utils.TimeToFirstSlotOfEpoch(uint64(nextDateTrunc.Unix())))
	}
	lastEpoch := lastSlot / int64(utils.Config.Chain.ClConfig.SlotsPerEpoch)

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
		return fmt.Errorf("this function does not yet handle pre merge statistics, firstBlock is %v, firstSlot is %v", firstBlock, firstSlot)
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

	switch utils.Config.Chain.ClConfig.DepositChainID {
	case 1:
		crowdSale := 72009990.50
		logger.Infof("Exporting MARKET_CAP: %v", newEmission.Div(decimal.NewFromInt(1e18)).Add(decimal.NewFromFloat(crowdSale)).Mul(decimal.NewFromFloat(price.GetPrice(utils.Config.Frontend.MainCurrency, "USD"))).String())
		err = SaveChartSeriesPoint(dateTrunc, "MARKET_CAP", newEmission.Div(decimal.NewFromInt(1e18)).Add(decimal.NewFromFloat(crowdSale)).Mul(decimal.NewFromFloat(price.GetPrice(utils.Config.Frontend.MainCurrency, "USD"))).String())
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
	firstSlot := uint64(day) * epochsPerDay * utils.Config.Chain.ClConfig.SlotsPerEpoch
	firstSlotOfNextDay := uint64(day+1) * epochsPerDay * utils.Config.Chain.ClConfig.SlotsPerEpoch

	tx, err := WriterDb.Beginx()
	if err != nil {
		return fmt.Errorf("error starting db tx in WriteGraffitiStatisticsForDay: %w", err)
	}
	defer tx.Rollback()

	// \x are missed blocks
	// \x0000000000000000000000000000000000000000000000000000000000000000 are empty graffities
	_, err = tx.Exec(`
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

	var lastSlot uint64
	err = tx.Get(&lastSlot, `select coalesce(max(slot),0) from blocks;`)
	if err != nil {
		return fmt.Errorf("error getting lastSlot in WriteGraffitiStatisticsForDay: %w", err)
	}

	// if last exported slot is younger than the last slot of the exported day then the day is completely exported
	if day < int64(utils.DayOfSlot(lastSlot)) {
		_, err = tx.Exec(`insert into graffiti_stats_status (day, status) values ($1, true) on conflict (day) do update set status = excluded.status`, day)
		if err != nil {
			return fmt.Errorf("error updating graffiti_stats_status in WriteGraffitiStatisticsForDay: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing db tx in WriteGraffitiStatisticsForDay: %w", err)
	}
	return nil
}

func CheckIfDayIsFinalized(day uint64) error {
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
