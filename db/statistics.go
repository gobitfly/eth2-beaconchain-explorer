package db

import (
	"eth2-exporter/metrics"
	"eth2-exporter/price"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"math/big"
	"strings"
	"time"

	itypes "github.com/gobitfly/eth-rewards/types"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

func WriteValidatorStatisticsForDay(day uint64) error {
	exportStart := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_validator_stats").Observe(time.Since(exportStart).Seconds())
	}()

	epochsPerDay := utils.EpochsPerDay()
	firstEpoch := day * epochsPerDay
	lastEpoch := firstEpoch + epochsPerDay - 1

	logger.Infof("exporting statistics for day %v (epoch %v to %v)", day, firstEpoch, lastEpoch)

	latestDbEpoch, err := GetLatestEpoch()
	if err != nil {
		return err
	}

	if lastEpoch > latestDbEpoch {
		return fmt.Errorf("delaying statistics export as epoch %v has not yet been indexed. LatestDB: %v", lastEpoch, latestDbEpoch)
	}

	start := time.Now()

	tx, err := WriterDb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	logger.Infof("exporting deposits and deposits_amount statistics")
	depositsQry := `
		insert into validator_stats (validatorindex, day, deposits, deposits_amount) 
		(
			select validators.validatorindex, $3, count(*), sum(amount)
			from blocks_deposits
			inner join validators on blocks_deposits.publickey = validators.pubkey
			inner join blocks on blocks_deposits.block_root = blocks.blockroot
			where block_slot >= $1 and block_slot <= $2 and blocks.status = '1'
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
				where block_slot >= $1 and block_slot <= $2 and status = '1'
				group by validators.validatorindex, day
			) 
			on conflict (validatorindex, day) do
				update set deposits = excluded.deposits, 
				deposits_amount = excluded.deposits_amount;`
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(depositsQry, firstEpoch*utils.Config.Chain.Config.SlotsPerEpoch, lastEpoch*utils.Config.Chain.Config.SlotsPerEpoch, day)
	if err != nil {
		return err
	}
	logger.Infof("export completed, took %v", time.Since(start))

	start = time.Now()
	logger.Infof("exporting cl_rewards_gwei and el_rewards_wei statistics")
	incomeDetails, err := BigtableClient.GetValidatorIncomeDetailsHistory([]uint64{}, firstEpoch, lastEpoch)
	if err != nil {
		return err
	}

	incomeStats := make(map[uint64]*itypes.ValidatorEpochIncome)

	for validator, epochs := range incomeDetails {
		if incomeStats[validator] == nil {
			incomeStats[validator] = &itypes.ValidatorEpochIncome{}
		}

		for _, rewardDetails := range epochs {
			incomeStats[validator].AttestationHeadReward += rewardDetails.AttestationHeadReward
			incomeStats[validator].AttestationSourceReward += rewardDetails.AttestationSourceReward
			incomeStats[validator].AttestationSourcePenalty += rewardDetails.AttestationSourcePenalty
			incomeStats[validator].AttestationTargetReward += rewardDetails.AttestationTargetReward
			incomeStats[validator].AttestationTargetPenalty += rewardDetails.AttestationTargetPenalty
			incomeStats[validator].FinalityDelayPenalty += rewardDetails.FinalityDelayPenalty
			incomeStats[validator].ProposerSlashingInclusionReward += rewardDetails.ProposerSlashingInclusionReward
			incomeStats[validator].ProposerAttestationInclusionReward += rewardDetails.ProposerAttestationInclusionReward
			incomeStats[validator].ProposerSyncInclusionReward += rewardDetails.ProposerSyncInclusionReward
			incomeStats[validator].SyncCommitteeReward += rewardDetails.SyncCommitteeReward
			incomeStats[validator].SyncCommitteePenalty += rewardDetails.SyncCommitteePenalty
			incomeStats[validator].SlashingReward += rewardDetails.SlashingReward
			incomeStats[validator].SlashingPenalty += rewardDetails.SlashingPenalty
			incomeStats[validator].TxFeeRewardWei = utils.AddBigInts(incomeStats[validator].TxFeeRewardWei, rewardDetails.TxFeeRewardWei)
		}
	}

	batchSize := 16000 // max parameters: 65535
	for b := 0; b < len(incomeStats); b += batchSize {
		start := b
		end := b + batchSize
		if len(incomeStats) < end {
			end = len(incomeStats) - 1
		}

		logger.Info(start, end)
		numArgs := 4
		valueStrings := make([]string, 0, batchSize)
		valueArgs := make([]interface{}, 0, batchSize*numArgs)
		for i := start; i <= end; i++ {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", (i-start)*numArgs+1, (i-start)*numArgs+2, (i-start)*numArgs+3, (i-start)*numArgs+4))
			clRewards := int64(0)
			elRewards := "0"
			if incomeStats[uint64(i)] != nil {
				clRewards = incomeStats[uint64(i)].TotalClRewards()
				elRewards = new(big.Int).SetBytes(incomeStats[uint64(i)].TxFeeRewardWei).String()
			} else {
				logger.Warnf("no rewards for validator %v available", i)
			}
			valueArgs = append(valueArgs, i)
			valueArgs = append(valueArgs, day)
			valueArgs = append(valueArgs, clRewards)
			valueArgs = append(valueArgs, elRewards)
		}
		stmt := fmt.Sprintf(`
		insert into validator_stats (validatorindex, day, cl_rewards_gwei, el_rewards_wei) VALUES
		%s
		on conflict (validatorindex, day) do update set cl_rewards_gwei = excluded.cl_rewards_gwei, el_rewards_wei = excluded.el_rewards_wei;`,
			strings.Join(valueStrings, ","))
		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			return err
		}

		logger.Infof("saving validator income details batch %v completed", b)
	}
	logger.Infof("export completed, took %v", time.Since(start))
	start = time.Now()

	logrus.Infof("exporting 7d income stats")
	_, err = tx.Exec(`insert into validator_stats (validatorindex, day, cl_rewards_gwei_7d, el_rewards_wei_7d) 
		(
			select validatorindex, $1, sum(coalesce(cl_rewards_gwei, 0)), sum(coalesce(el_rewards_wei, 0)) 
			from validator_stats 
			where day <= $1 and day > $1 - 7 
			group by validatorindex
		) 
		on conflict (validatorindex, day) do update set 
		cl_rewards_gwei_7d = excluded.cl_rewards_gwei_7d, el_rewards_wei_7d=excluded.el_rewards_wei_7d;`, day)
	if err != nil {
		return err
	}

	logrus.Infof("exporting 31d income stats")
	_, err = tx.Exec(`insert into validator_stats (validatorindex, day, cl_rewards_gwei_31d, el_rewards_wei_31d) 
		(
			select validatorindex, $1, sum(coalesce(cl_rewards_gwei, 0)), sum(coalesce(el_rewards_wei, 0)) 
			from validator_stats 
			where day <= $1 and day > $1 - 31 
			group by validatorindex
		) 
		on conflict (validatorindex, day) do update set 
		cl_rewards_gwei_31d = excluded.cl_rewards_gwei_31d, el_rewards_wei_31d=excluded.el_rewards_wei_31d;`, day)
	if err != nil {
		return err
	}

	logger.Infof("exporting total income stats")
	_, err = tx.Exec(`insert into validator_stats (validatorindex, day, cl_rewards_gwei_total, el_rewards_wei_total) 
		(
			select vs1.validatorindex, $1, coalesce(vs1.cl_rewards_gwei, 0) + coalesce(vs2.cl_rewards_gwei_total, 0), coalesce(vs1.el_rewards_wei, 0) + coalesce(vs2.el_rewards_wei_total, 0) 
			from validator_stats vs1
			left join validator_stats vs2 on vs2.day = $1 - 1 and vs1.validatorindex = vs2.validatorindex
			where vs1.day = $1
		) 
		on conflict (validatorindex, day) do update set 
		cl_rewards_gwei_total = excluded.cl_rewards_gwei_total, el_rewards_wei_total=excluded.el_rewards_wei_total;`, day)
	if err != nil {
		return err
	}

	logger.Infof("populate validator_performance table")
	_, err = tx.Exec(`insert into validator_performance (validatorindex, balance, performance1d, performance7d, performance31d, performance365d, rank7d) 
		(
			select 
				validatorindex, 
				end_balance as balance, 
				cl_rewards_gwei as performance1d, 
				cl_rewards_gwei_7d as performance7d, 
				cl_rewards_gwei_31d as performance31d, 
				cl_rewards_gwei_total as performance365d, 
				row_number() over(order by cl_rewards_gwei_7d desc) as rank7d 
			from validator_stats where day = 248
		) 
		on conflict (validatorindex) do update set 
			balance = excluded.balance, 
			performance1d=excluded.performance1d,
			performance7d=excluded.performance7d,
			performance31d=excluded.performance31d,
			performance365d=excluded.performance365d,
			rank7d=excluded.rank7d
			;`) //, day)
	if err != nil {
		return err
	}

	logger.Infof("exporting min_balance, max_balance, min_effective_balance, max_effective_balance, start_balance, start_effective_balance, end_balance and end_effective_balance statistics")
	balanceStatistics, err := BigtableClient.GetValidatorBalanceStatistics(firstEpoch, lastEpoch)
	if err != nil {
		return err
	}

	balanceStatsArr := make([]*types.ValidatorBalanceStatistic, 0, len(balanceStatistics))
	for _, stat := range balanceStatistics {
		balanceStatsArr = append(balanceStatsArr, stat)
	}

	batchSize = 6500 // max parameters: 65535
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
	logger.Infof("exporting missed_attestations statistics lastEpoch: %v firstEpoch: %v", lastEpoch, firstEpoch)
	ma, err := BigtableClient.GetValidatorMissedAttestationsCount([]uint64{}, firstEpoch, lastEpoch)
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
	syncStats, err := BigtableClient.GetValidatorSyncDutiesStatistics([]uint64{}, firstEpoch, lastEpoch) //+1 is needed because the function uses limit instead of end epoch
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
	_, err = tx.Exec(withdrawalsQuery, firstEpoch*utils.Config.Chain.Config.SlotsPerEpoch, lastEpoch*utils.Config.Chain.Config.SlotsPerEpoch, day)
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

func GetValidatorIncomeHistoryChart(validator_indices []uint64, currency string) ([]*types.ChartDataPoint, int64, error) {
	incomeHistory, currentDayIncome, err := GetValidatorIncomeHistory(validator_indices, 0, 0)
	if err != nil {
		return nil, 0, err
	}
	var clRewardsSeries = make([]*types.ChartDataPoint, len(incomeHistory))

	for i := 0; i < len(incomeHistory); i++ {
		color := "#7cb5ec"
		if incomeHistory[i].ClRewards < 0 {
			color = "#f7a35c"
		}
		balanceTs := utils.DayToTime(incomeHistory[i].Day + 1)
		clRewardsSeries[i] = &types.ChartDataPoint{X: float64(balanceTs.Unix() * 1000), Y: utils.ExchangeRateForCurrency(currency) * (float64(incomeHistory[i].ClRewards) / 1e9), Color: color}
	}
	return clRewardsSeries, currentDayIncome, err
}

func GetValidatorIncomeHistory(validator_indices []uint64, lowerBoundDay uint64, upperBoundDay uint64) ([]types.ValidatorIncomeHistory, int64, error) {
	if upperBoundDay == 0 {
		upperBoundDay = 65536
	}
	queryValidatorsArr := pq.Array(validator_indices)

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
	;`, queryValidatorsArr, lowerBoundDay, upperBoundDay)

	// retrieve rewards for epochs not yet in stats
	currentDayIncome := int64(0)
	if upperBoundDay == 65536 && len(result) > 0 {
		lastDay := result[len(result)-1].Day
		currentDay := uint64(lastDay + 1)
		startEpoch := currentDay * utils.EpochsPerDay()
		endEpoch := startEpoch + utils.EpochsPerDay() - 1
		income, err := BigtableClient.GetValidatorIncomeDetailsHistory(validator_indices, startEpoch, endEpoch)

		if err != nil {
			return nil, 0, err
		}

		for _, ids := range income {
			for _, id := range ids {
				currentDayIncome += id.TotalClRewards()
			}
		}

		result = append(result, types.ValidatorIncomeHistory{
			Day:       int64(currentDay),
			ClRewards: currentDayIncome,
		})
	}

	return result, currentDayIncome, err
}

func WriteChartSeriesForDay(day int64) error {
	startTs := time.Now()

	if day < 0 {
		// before the beaconchain
		return fmt.Errorf("this function does not yet pre-beaconchain blocks")
	}

	epochsPerDay := utils.EpochsPerDay()
	beaconchainDay := day * int64(epochsPerDay)

	startDate := utils.EpochToTime(uint64(beaconchainDay))
	dateTrunc := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)

	// inclusive slot
	firstSlot := utils.TimeToSlot(uint64(dateTrunc.Unix()))

	epochOffset := firstSlot % utils.Config.Chain.Config.SlotsPerEpoch
	firstSlot = firstSlot - epochOffset
	firstEpoch := firstSlot / utils.Config.Chain.Config.SlotsPerEpoch
	// exclusive slot
	lastSlot := int64(firstSlot) + int64(epochsPerDay*utils.Config.Chain.Config.SlotsPerEpoch)
	lastEpoch := lastSlot / int64(utils.Config.Chain.Config.SlotsPerEpoch)

	latestDbEpoch, err := GetLatestEpoch()
	if err != nil {
		return err
	}

	if (uint64(lastSlot) / utils.Config.Chain.Config.SlotsPerEpoch) > latestDbEpoch {
		return fmt.Errorf("delaying statistics export as epoch %v has not yet been indexed. LatestDB: %v", (uint64(lastSlot) / utils.Config.Chain.Config.SlotsPerEpoch), latestDbEpoch)
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
			low := b - batchSize
			if int64(firstBlock) > low {
				low = int64(firstBlock - 1)
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

	avgBlockTime := decimal.NewFromInt(0)

	for blk := range blocksChan {
		// logger.Infof("analyzing block: %v with: %v transactions", blk.Number, len(blk.Transactions))
		blockCount += 1
		baseFee := decimal.NewFromBigInt(new(big.Int).SetBytes(blk.BaseFee), 0)
		totalBaseFee = totalBaseFee.Add(baseFee)
		totalGasLimit = totalGasLimit.Add(decimal.NewFromInt(int64(blk.GasLimit)))

		if prevBlock != nil {
			avgBlockTime = avgBlockTime.Add(decimal.NewFromInt(prevBlock.Time.AsTime().UnixMicro() - blk.Time.AsTime().UnixMicro())).Div(decimal.NewFromInt(2))
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

	logger.Infof("exporting consensus rewards from %v to %v", firstEpoch, lastEpoch)

	// consensus rewards are in Gwei
	totalConsensusRewards := int64(0)

	err = WriterDb.Get(&totalConsensusRewards, "SELECT SUM(COALESCE(cl_rewards_gwei, 0)) FROM validator_stats WHERE day = $1", day)
	if err != nil {
		return fmt.Errorf("error calculating totalConsensusRewards: %w", err)
	}
	logger.Infof("consensus rewards: %v", totalConsensusRewards)

	logger.Infof("Exporting BURNED_FEES %v", totalBurned.String())
	_, err = WriterDb.Exec("INSERT INTO chart_series (time, indicator, value) VALUES ($1, 'BURNED_FEES', $2) ON CONFLICT (time, indicator) DO UPDATE SET value = EXCLUDED.value", dateTrunc, totalBurned.String())
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
	if err != nil {
		return fmt.Errorf("error getting previous value for TOTAL_EMISSION chart_series: %w", err)
	}

	newEmission := decimal.NewFromFloat(lastEmission).Add(emission)
	err = SaveChartSeriesPoint(dateTrunc, "TOTAL_EMISSION", newEmission)
	if err != nil {
		return fmt.Errorf("error calculating TOTAL_EMISSION chart_series: %w", err)
	}

	if totalGasPrice.GreaterThan(decimal.NewFromInt(0)) && decimal.NewFromInt(legacyTxCount).Add(decimal.NewFromInt(accessListTxCount)).GreaterThan(decimal.NewFromInt(0)) {
		logger.Infof("Exporting AVG_GASPRICE")
		_, err = WriterDb.Exec("INSERT INTO chart_series (time, indicator, value) VALUES($1, 'AVG_GASPRICE', $2) ON CONFLICT (time, indicator) DO UPDATE SET value = EXCLUDED.value", dateTrunc, totalGasPrice.Div((decimal.NewFromInt(legacyTxCount).Add(decimal.NewFromInt(accessListTxCount)))).String())
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

	logger.Infof("Exporting MARKET_CAP: %v", newEmission.Div(decimal.NewFromInt(1e18)).Add(decimal.NewFromFloat(72009990.50)).Mul(decimal.NewFromFloat(price.GetEthPrice("USD"))).String())
	err = SaveChartSeriesPoint(dateTrunc, "MARKET_CAP", newEmission.Div(decimal.NewFromInt(1e18)).Add(decimal.NewFromFloat(72009990.50)).Mul(decimal.NewFromFloat(price.GetEthPrice("USD"))).String())
	if err != nil {
		return fmt.Errorf("error calculating MARKET_CAP chart_series: %w", err)
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

	logger.Infof("marking day export as completed in the status table")
	_, err = WriterDb.Exec("insert into chart_series_status (day, status) values ($1, true)", day)
	if err != nil {
		return err
	}

	logger.Infof("chart_series export completed: took %v", time.Since(startTs))

	return nil
}
