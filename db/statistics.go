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

	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

func WriteValidatorStatisticsForDay(day uint64) error {
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
		return fmt.Errorf("delaying statistics export as epoch %v has not yet been indexed. LatestDB: %v", lastEpoch, latestDbEpoch)
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
	syncStats, err := BigtableClient.GetValidatorSyncDutiesStatistics([]uint64{}, lastEpoch, int64(lastEpoch-firstEpoch)+1) //+1 is needed because the function uses limit instead of end epoch
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

func GetValidatorELIncomeHistory(validator_indices []uint64, lowerBoundDay uint64, upperBoundDay uint64) ([]types.ValidatorIncomeHistory, error) {
	if upperBoundDay == 0 {
		upperBoundDay = 65536
	}
	queryValidatorsArr := pq.Array(validator_indices)

	var result []types.ValidatorIncomeHistory
	err := ReaderDb.Select(&result, `
		select 
			day, consensus_rewards_sum_wei AS income
		FROM eth_store_stats
		WHERE validator = ANY($1) AND
			day BETWEEN ($2 - 1) AND $3
		GROUP BY day
		ORDER BY day
	`, queryValidatorsArr, lowerBoundDay, upperBoundDay)

	return result, err
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

func WriteChartSeriesForDay(day int64) error {
	startTs := time.Now()

	if day < 0 {
		// before the beaconchain
		return fmt.Errorf("this function does not yet pre-beaconchain blocks")
	}

	epochsPerDay := (24 * 60 * 60) / utils.Config.Chain.Config.SlotsPerEpoch / utils.Config.Chain.Config.SecondsPerSlot
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
	historyFirst, err := BigtableClient.GetValidatorBalanceHistory(nil, firstEpoch+1, 1)
	if err != nil {
		return err
	}

	sumStartEpoch := decimal.NewFromInt(0)
	for _, balances := range historyFirst {
		for _, balance := range balances {
			sumStartEpoch = sumStartEpoch.Add(decimal.NewFromInt(int64(balance.Balance)))
		}
	}

	historyLast, err := BigtableClient.GetValidatorBalanceHistory(nil, uint64(lastEpoch+1), 1)
	if err != nil {
		return err
	}

	sumEndEpoch := decimal.NewFromInt(0)
	for _, balances := range historyLast {
		for _, balance := range balances {
			sumEndEpoch = sumEndEpoch.Add(decimal.NewFromInt(int64(balance.Balance)))
		}
	}
	// consensus rewards are in Gwei
	totalConsensusRewards := sumEndEpoch.Sub(sumStartEpoch)
	logger.Infof("consensus rewards: %v", totalConsensusRewards.String())

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
	emission := (totalBaseBlockReward.Add(totalConsensusRewards.Mul(decimal.NewFromInt(1000000000))).Add(totalTips)).Sub(totalBurned)
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
