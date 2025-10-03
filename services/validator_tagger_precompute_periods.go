package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gobitfly/eth2-beaconchain-explorer/cache"
	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

// PeriodSpec defines ClickHouse sources and labels for a precomputation period.
type PeriodSpec struct {
	Label              string // one of: "1d", "7d", "30d", "90d"
	RollingTable       string // e.g., validator_dashboard_data_rolling_24h / _7d / _30d / _90d
	HistoryTable       string // e.g., validator_dashboard_data_hourly / _daily / _monthly
	HistoryWhereClause string // optional WHERE clause for limiting the history time window (without the WHERE keyword)
}

// precomputeEntityData runs entity precomputation for all supported periods in sequence.
func precomputeEntityData(ctx context.Context) error {
	specs := []PeriodSpec{
		{Label: "1d", RollingTable: "_final_validator_dashboard_rolling_24h", HistoryTable: "validator_dashboard_data_hourly", HistoryWhereClause: "t >= now() - INTERVAL 24 HOUR"},
		{Label: "7d", RollingTable: "_final_validator_dashboard_rolling_7d", HistoryTable: "validator_dashboard_data_daily", HistoryWhereClause: "t >= today() - 7"},
		{Label: "30d", RollingTable: "_final_validator_dashboard_rolling_30d", HistoryTable: "validator_dashboard_data_daily", HistoryWhereClause: "t >= today() - 30"},
		{Label: "90d", RollingTable: "_final_validator_dashboard_rolling_90d", HistoryTable: "validator_dashboard_data_daily", HistoryWhereClause: "t >= today() - 90"},
	}

	// 1) retrieve the current balances for all validators
	validatorTaggerLogger.Info("retrieving validator balance data")
	validatorBalances, err := fetchValidatorBalancesFromMapping()
	if err != nil {
		return fmt.Errorf("failed to fetch validator balances from redis: %w", err)
	}
	validatorTaggerLogger.Infof("retrieved %d validator balances", len(validatorBalances))

	// 2) Fetch validator index, status, and entity data
	// we retrieve this only once and reuse it for all time periods
	// first start a db transaction
	tx, err := db.WriterDb.Beginx()
	if err != nil {
		return fmt.Errorf("begin db transaction: %w", err)
	}
	defer tx.Rollback()

	validatorTaggerLogger.Info("retrieving validator entity and status data")
	var validatorEntityRows []ValidatorEntityJoinRow
	const joinSQL = `
        SELECT v.validatorindex,
               v.status,
               COALESCE(ve.entity, 'Unknown') AS entity,
               ve.sub_entity
        FROM validators v
        LEFT JOIN validator_entities ve ON v.pubkey = ve.publickey
    `
	if err := tx.Select(&validatorEntityRows, joinSQL); err != nil {
		return fmt.Errorf("join validators and validator_entities: %w", err)
	}
	if len(validatorEntityRows) == 0 {
		logger.Info("no validators found; skipping")
		return nil
	}

	for _, spec := range specs {
		err = precomputeEntityDataForPeriod(ctx, tx, spec, validatorEntityRows, validatorBalances)
		if err != nil {
			return fmt.Errorf("error precomputing entity data for period %s: %w", spec.Label, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit db transaction: %w", err)
	}
	return nil
}

type ValidatorEntityJoinRow struct {
	ValidatorIndex uint64  `db:"validatorindex"`
	Status         string  `db:"status"`
	Entity         string  `db:"entity"`
	SubEntity      *string `db:"sub_entity"`
}

func precomputeEntityDataForPeriod(ctx context.Context, tx *sqlx.Tx, spec PeriodSpec, validatorEntityStatusMapping []ValidatorEntityJoinRow, balanceMap map[int]uint64) error {
	logger := validatorTaggerLogger.WithField("period", spec.Label)
	logger.Info("precomputeEntityDataForPeriod: start")

	// Clear old rows for this period within the transaction to avoid TRUNCATE locks
	res, err := tx.Exec(`DELETE FROM validator_entities_data_periods WHERE period = $1`, spec.Label)
	if err != nil {
		return fmt.Errorf("delete old rows for period %s: %w", spec.Label, err)
	}
	if rows, errRA := res.RowsAffected(); errRA == nil {
		logger.WithField("deleted_rows", rows).Info("cleared old rows for period")
	}

	// Aggregation scaffolding
	type AggregationKey struct{ Entity, SubEntity string }
	type AggregationValue struct {
		BalanceEndSumGwei int64

		EfficiencyDividend decimal.Decimal
		EfficiencyDivisor  decimal.Decimal

		RoiDividend decimal.Decimal
		RoiDivisor  decimal.Decimal

		EfficiencyAttestationsDividend decimal.Decimal
		EfficiencyAttestationsDivisor  decimal.Decimal
		EfficiencyProposalsDividend    decimal.Decimal
		EfficiencyProposalsDivisor     decimal.Decimal
		EfficiencySyncDividend         decimal.Decimal
		EfficiencySyncDivisor          decimal.Decimal

		ExecutionRewardsSumWei decimal.Decimal

		AttestationsScheduledSum      int64
		AttestationsObservedSum       int64
		AttestationsHeadExecutedSum   int64
		AttestationsSourceExecutedSum int64
		AttestationsTargetExecutedSum int64
		AttestationsIdealRewardSum    int64
		AttestationsRewardsOnlySum    int64
		BlocksScheduledSum            int64
		BlocksProposedSum             int64
		SyncScheduledSum              int64
		SyncExecutedSum               int64
		SlashedInPeriodMax            int64
		SlashedAmountSum              int64
		BlocksClMissedMedianRewardSum int64
		SyncLocalizedMaxRewardSum     int64
		SyncRewardRewardsOnlySum      int64
		InclusionDelaySum             int64

		StatusCountsByStatus map[string]int
	}
	aggregations := make(map[AggregationKey]*AggregationValue, 4096)
	entityByValidatorIndex := make(map[uint64]struct{ entity, subEntity string }, len(validatorEntityStatusMapping))
	ensureAggregation := func(key AggregationKey) *AggregationValue {
		value := aggregations[key]
		if value == nil {
			value = &AggregationValue{StatusCountsByStatus: make(map[string]int, 8)}
			aggregations[key] = value
		}
		return value
	}
	for _, row := range validatorEntityStatusMapping {
		subEntity := ""
		if row.SubEntity != nil {
			subEntity = strings.TrimSpace(*row.SubEntity)
		}
		entityByValidatorIndex[row.ValidatorIndex] = struct{ entity, subEntity string }{entity: row.Entity, subEntity: subEntity}
		keyEntityOnly := AggregationKey{Entity: row.Entity, SubEntity: ""}
		ensureAggregation(keyEntityOnly).StatusCountsByStatus[row.Status]++
		if subEntity != "" {
			keyEntityWithSub := AggregationKey{Entity: row.Entity, SubEntity: subEntity}
			ensureAggregation(keyEntityWithSub).StatusCountsByStatus[row.Status]++
		}
	}

	// Parquet row from ClickHouse rolling table
	type ClickHouseDataRow struct {
		ValidatorIndex                 uint64 `parquet:"validator_index"`
		EpochStart                     int64  `parquet:"epoch_start"`
		EpochEnd                       int64  `parquet:"epoch_end"`
		EfficiencyDividend             int64  `parquet:"efficiency_dividend"`
		EfficiencyDivisor              int64  `parquet:"efficiency_divisor"`
		RoiDividendLE                  []byte `parquet:"roi_dividend"`
		RoiDivisorLE                   []byte `parquet:"roi_divisor"`
		EfficiencyAttestationsDividend int64  `parquet:"efficiency_attestations_dividend"`
		EfficiencyAttestationsDivisor  int64  `parquet:"efficiency_attestations_divisor"`
		EfficiencyProposalsDividend    int64  `parquet:"efficiency_proposals_dividend"`
		EfficiencyProposalsDivisor     int64  `parquet:"efficiency_proposals_divisor"`
		EfficiencySyncDividend         int64  `parquet:"efficiency_sync_dividend"`
		EfficiencySyncDivisor          int64  `parquet:"efficiency_sync_divisor"`
		AttestationsScheduledSum       int64  `parquet:"attestations_scheduled"`
		AttestationsObservedSum        int64  `parquet:"attestations_observed"`
		AttestationsHeadExecutedSum    int64  `parquet:"attestations_head_executed"`
		AttestationsSourceExecutedSum  int64  `parquet:"attestations_source_executed"`
		AttestationsTargetExecutedSum  int64  `parquet:"attestations_target_executed"`
		AttestationsIdealRewardSum     int64  `parquet:"attestations_ideal_reward"`
		AttestationsRewardsOnlySum     int64  `parquet:"attestations_reward_rewards_only"`
		BlocksScheduledSum             int64  `parquet:"blocks_scheduled"`
		BlocksProposedSum              int64  `parquet:"blocks_proposed"`
		SyncScheduledSum               int64  `parquet:"sync_scheduled"`
		SyncExecutedSum                int64  `parquet:"sync_executed"`
		Slashed                        int64  `parquet:"slashed"`
		BlocksSlashingCount            int64  `parquet:"blocks_slashing_count"`
		BlocksClMissedMedianRewardSum  int64  `parquet:"blocks_cl_missed_median_reward"`
		SyncLocalizedMaxRewardSum      int64  `parquet:"sync_localized_max_reward"`
		SyncRewardRewardsOnlySum       int64  `parquet:"sync_reward_rewards_only"`
		InclusionDelaySum              int64  `parquet:"inclusion_delay_sum"`
	}

	// Int128 LE to decimal
	int128LEToDecimal := func(b []byte) decimal.Decimal {
		if len(b) == 0 {
			return decimal.Zero
		}
		bCopy := make([]byte, len(b))
		for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
			bCopy[i], bCopy[j] = b[j], b[i]
		}
		z := new(big.Int).SetBytes(bCopy)
		return decimal.NewFromBigInt(z, 0)
	}

	// aggregation adder
	addRowToAggregation := func(agg *AggregationValue, row ClickHouseDataRow, validatorSeen bool) {
		if !validatorSeen {
			agg.BalanceEndSumGwei += int64(balanceMap[int(row.ValidatorIndex)])
		}
		agg.EfficiencyDividend = agg.EfficiencyDividend.Add(decimal.NewFromInt(row.EfficiencyDividend))
		agg.EfficiencyDivisor = agg.EfficiencyDivisor.Add(decimal.NewFromInt(row.EfficiencyDivisor))
		agg.RoiDividend = agg.RoiDividend.Add(int128LEToDecimal(row.RoiDividendLE))
		agg.RoiDivisor = agg.RoiDivisor.Add(int128LEToDecimal(row.RoiDivisorLE))
		agg.EfficiencyAttestationsDividend = agg.EfficiencyAttestationsDividend.Add(decimal.NewFromInt(row.EfficiencyAttestationsDividend))
		agg.EfficiencyAttestationsDivisor = agg.EfficiencyAttestationsDivisor.Add(decimal.NewFromInt(row.EfficiencyAttestationsDivisor))
		agg.EfficiencyProposalsDividend = agg.EfficiencyProposalsDividend.Add(decimal.NewFromInt(row.EfficiencyProposalsDividend))
		agg.EfficiencyProposalsDivisor = agg.EfficiencyProposalsDivisor.Add(decimal.NewFromInt(row.EfficiencyProposalsDivisor))
		agg.EfficiencySyncDividend = agg.EfficiencySyncDividend.Add(decimal.NewFromInt(row.EfficiencySyncDividend))
		agg.EfficiencySyncDivisor = agg.EfficiencySyncDivisor.Add(decimal.NewFromInt(row.EfficiencySyncDivisor))
		agg.AttestationsScheduledSum += row.AttestationsScheduledSum
		agg.AttestationsObservedSum += row.AttestationsObservedSum
		agg.AttestationsHeadExecutedSum += row.AttestationsHeadExecutedSum
		agg.AttestationsSourceExecutedSum += row.AttestationsSourceExecutedSum
		agg.AttestationsTargetExecutedSum += row.AttestationsTargetExecutedSum
		agg.AttestationsIdealRewardSum += row.AttestationsIdealRewardSum
		agg.AttestationsRewardsOnlySum += row.AttestationsRewardsOnlySum
		agg.BlocksScheduledSum += row.BlocksScheduledSum
		agg.BlocksProposedSum += row.BlocksProposedSum
		agg.SyncScheduledSum += row.SyncScheduledSum
		agg.SyncExecutedSum += row.SyncExecutedSum
		if row.Slashed > agg.SlashedInPeriodMax {
			agg.SlashedInPeriodMax = row.Slashed
		}
		agg.SlashedAmountSum += row.BlocksSlashingCount
		agg.BlocksClMissedMedianRewardSum += row.BlocksClMissedMedianRewardSum
		agg.SyncLocalizedMaxRewardSum += row.SyncLocalizedMaxRewardSum
		agg.SyncRewardRewardsOnlySum += row.SyncRewardRewardsOnlySum
		agg.InclusionDelaySum += row.InclusionDelaySum
	}

	// 2) Stream aggregated rows from ClickHouse (rolling table)
	clickhouseSQL := fmt.Sprintf(`
		SELECT
		  validator_index,
		  efficiency_dividend,
		  efficiency_divisor,
		  roi_dividend,
		  roi_divisor,
		  efficiency_attestations_dividend,
		  efficiency_attestations_divisor,
		  efficiency_proposals_dividend,
		  efficiency_proposals_divisor,
		  efficiency_sync_dividend,
		  efficiency_sync_divisor,
		  attestations_scheduled,
		  attestations_observed,
		  attestations_head_executed,
		  attestations_source_executed,
		  attestations_target_executed,
		  attestations_ideal_reward,
		  attestations_reward_rewards_only,
		  blocks_scheduled,
		  blocks_proposed,
		  sync_scheduled,
		  sync_executed,
		  slashed, -- max
		  blocks_slashing_count,
		  blocks_cl_missed_median_reward,
		  sync_localized_max_reward,
		  sync_reward_rewards_only,
		  inclusion_delay_sum,
		  epoch_start, -- min
		  epoch_end -- max
		FROM %s
		FORMAT Parquet
		SETTINGS output_format_parquet_compression_method='zstd'
	`, spec.RollingTable)

	var totalBalanceEndGwei int64
	epochStart := int64(math.MaxInt64)
	epochEnd := int64(math.MinInt64)
	validatorsSeen := make(map[uint64]bool) // map to keep track when we see a validator for the first time
	if err := db.FetchClickhouseParquet[ClickHouseDataRow](ctx, clickhouseSQL, func(row ClickHouseDataRow) bool {
		mapping, ok := entityByValidatorIndex[row.ValidatorIndex]
		if !ok {
			return true
		}
		if row.EpochStart < epochStart {
			epochStart = row.EpochStart
		}
		if row.EpochEnd > epochEnd {
			epochEnd = row.EpochEnd
		}
		keyEntityOnly := AggregationKey{Entity: mapping.entity, SubEntity: ""}
		addRowToAggregation(ensureAggregation(keyEntityOnly), row, validatorsSeen[row.ValidatorIndex])
		if mapping.subEntity != "" {
			keyEntityWithSub := AggregationKey{Entity: mapping.entity, SubEntity: mapping.subEntity}
			addRowToAggregation(ensureAggregation(keyEntityWithSub), row, validatorsSeen[row.ValidatorIndex])
		}
		if !validatorsSeen[row.ValidatorIndex] {
			totalBalanceEndGwei += int64(balanceMap[int(row.ValidatorIndex)])
		}
		validatorsSeen[row.ValidatorIndex] = true
		return true
	}); err != nil {
		return fmt.Errorf("fetch clickhouse parquet balances (%s): %w", spec.Label, err)
	}

	// 2b) History buckets
	type ClickHouseTimeBucketRow struct {
		ValidatorIndex     uint64 `parquet:"validator_index"`
		TUnix              int64  `parquet:"t_unix"`
		EfficiencyDividend int64  `parquet:"efficiency_dividend"`
		EfficiencyDivisor  int64  `parquet:"efficiency_divisor"`
	}
	type TimeBucket struct {
		Dividend decimal.Decimal
		Divisor  decimal.Decimal
	}
	timeBucketAgg := make(map[AggregationKey]map[int64]*TimeBucket, 4096)
	whereClause := ""
	if strings.TrimSpace(spec.HistoryWhereClause) != "" {
		whereClause = "WHERE " + spec.HistoryWhereClause
	}
	timeBucketSQL := fmt.Sprintf(`
		SELECT
		  validator_index,
		  toInt64(toUnixTimestamp(t)) AS t_unix,
		  efficiency_dividend,
		  efficiency_divisor
		FROM %s
		%s
		FORMAT Parquet
		SETTINGS output_format_parquet_compression_method='zstd'
	`, spec.HistoryTable, whereClause)
	if err := db.FetchClickhouseParquet[ClickHouseTimeBucketRow](ctx, timeBucketSQL, func(row ClickHouseTimeBucketRow) bool {
		mapping, ok := entityByValidatorIndex[row.ValidatorIndex]
		if !ok {
			return true
		}
		update := func(key AggregationKey) {
			bucketsForKey := timeBucketAgg[key]
			if bucketsForKey == nil {
				bucketsForKey = make(map[int64]*TimeBucket, 64)
				timeBucketAgg[key] = bucketsForKey
			}
			bucket := bucketsForKey[row.TUnix]
			if bucket == nil {
				bucket = &TimeBucket{}
				bucketsForKey[row.TUnix] = bucket
			}
			bucket.Dividend = bucket.Dividend.Add(decimal.NewFromInt(row.EfficiencyDividend))
			bucket.Divisor = bucket.Divisor.Add(decimal.NewFromInt(row.EfficiencyDivisor))
		}
		update(AggregationKey{Entity: mapping.entity, SubEntity: ""})
		if mapping.subEntity != "" {
			update(AggregationKey{Entity: mapping.entity, SubEntity: mapping.subEntity})
		}
		return true
	}); err != nil {
		return fmt.Errorf("fetch clickhouse parquet history (%s): %w", spec.Label, err)
	}

	// 3) Execution layer rewards from Postgres
	if epochStart <= epochEnd && epochStart != int64(math.MaxInt64) && epochEnd != int64(math.MinInt64) {
		logger.WithField("epoch_start", epochStart).WithField("epoch_end", epochEnd).Info("fetching EL rewards")
		type ELRewardRow struct {
			Proposer uint64          `db:"proposer"`
			Value    decimal.Decimal `db:"value"`
		}
		var elRows []ELRewardRow
		if err := tx.Select(&elRows, `
			SELECT proposer, value
			FROM execution_rewards_finalized
			WHERE epoch >= $1 AND epoch <= $2
		`, epochStart, epochEnd); err != nil {
			return fmt.Errorf("fetch execution_rewards_finalized: %w", err)
		}
		for _, r := range elRows {
			mapping, ok := entityByValidatorIndex[r.Proposer]
			if !ok {
				continue
			}
			keyEntityOnly := AggregationKey{Entity: mapping.entity, SubEntity: ""}
			ensureAggregation(keyEntityOnly).ExecutionRewardsSumWei = ensureAggregation(keyEntityOnly).ExecutionRewardsSumWei.Add(r.Value)
			if mapping.subEntity != "" {
				keyEntityWithSub := AggregationKey{Entity: mapping.entity, SubEntity: mapping.subEntity}
				ensureAggregation(keyEntityWithSub).ExecutionRewardsSumWei = ensureAggregation(keyEntityWithSub).ExecutionRewardsSumWei.Add(r.Value)
			}
		}
	} else {
		logger.WithField("epoch_start", epochStart).WithField("epoch_end", epochEnd).Warn("invalid epoch range; skipping EL rewards fetch")
	}

	// Build per-group time bucket arrays
	buildTimeBucketArrays := func(key AggregationKey) (timestamps []int64, values []float64) {
		bucketsByTimestamp := timeBucketAgg[key]
		if len(bucketsByTimestamp) == 0 {
			return nil, nil
		}
		timestamps = make([]int64, 0, len(bucketsByTimestamp))
		for ts := range bucketsByTimestamp {
			timestamps = append(timestamps, ts)
		}
		sort.Slice(timestamps, func(i, j int) bool { return timestamps[i] < timestamps[j] })
		values = make([]float64, 0, len(timestamps))
		for _, ts := range timestamps {
			bucket := bucketsByTimestamp[ts]
			values = append(values, utils.CalcEfficiency(bucket.Dividend, bucket.Divisor))
		}
		return timestamps, values
	}

	// Final output rows
	type OutputEntry struct {
		Entity                            string          `json:"entity"`
		SubEntity                         string          `json:"sub_entity"`
		BalanceEndSumGwei                 int64           `json:"balance_end_sum_gwei"`
		EfficiencyDividend                decimal.Decimal `json:"efficiency_dividend"`
		EfficiencyDivisor                 decimal.Decimal `json:"efficiency_divisor"`
		Efficiency                        float64         `json:"efficiency"`
		RoiDividend                       decimal.Decimal `json:"roi_dividend"`
		RoiDivisor                        decimal.Decimal `json:"roi_divisor"`
		AttestationEfficiency             float64         `json:"attestation_efficiency"`
		ProposalEfficiency                float64         `json:"proposal_efficiency"`
		SyncCommitteeEfficiency           float64         `json:"sync_committee_efficiency"`
		EfficiencyTimeBucketTimestampsSec []int64         `json:"efficiency_time_bucket_ts"`
		EfficiencyTimeBucketValues        []float64       `json:"efficiency_time_bucket_values"`
		AttestationsScheduledSum          int64           `json:"attestations_scheduled_sum"`
		AttestationsObservedSum           int64           `json:"attestations_observed_sum"`
		AttestationsHeadExecutedSum       int64           `json:"attestations_head_executed_sum"`
		AttestationsSourceExecutedSum     int64           `json:"attestations_source_executed_sum"`
		AttestationsTargetExecutedSum     int64           `json:"attestations_target_executed_sum"`
		AttestationsMissedRewardsSum      int64           `json:"attestations_missed_rewards_sum"`
		AttestationsRewardsOnlySum        int64           `json:"attestations_reward_rewards_only_sum"`
		BlocksScheduledSum                int64           `json:"blocks_scheduled_sum"`
		BlocksProposedSum                 int64           `json:"blocks_proposed_sum"`
		SyncScheduledSum                  int64           `json:"sync_scheduled_sum"`
		SyncExecutedSum                   int64           `json:"sync_executed_sum"`
		SlashedInPeriodMax                int64           `json:"slashed_in_period_max"`
		SlashedAmountSum                  int64           `json:"slashed_amount_sum"`
		BlocksClMissedMedianRewardSum     int64           `json:"blocks_cl_missed_median_reward_sum"`
		SyncLocalizedMaxRewardSum         int64           `json:"sync_localized_max_reward_sum"`
		SyncRewardRewardsOnlySum          int64           `json:"sync_reward_rewards_only_sum"`
		InclusionDelaySum                 int64           `json:"inclusion_delay_sum"`
		ExecutionRewardsSumWei            decimal.Decimal `json:"execution_rewards_sum_wei"`
		NetShare                          float64         `json:"net_share"`
		StatusCounts                      map[string]int  `json:"status_counts"`
	}

	outputEntries := make([]OutputEntry, 0, len(aggregations))
	for key, value := range aggregations {
		subEntityOutput := key.SubEntity
		if strings.TrimSpace(subEntityOutput) == "" {
			subEntityOutput = "-"
		}
		var netShare float64
		if totalBalanceEndGwei > 0 {
			netShare = float64(value.BalanceEndSumGwei) / float64(totalBalanceEndGwei)
		}
		bucketTimestamps, bucketValues := buildTimeBucketArrays(key)
		outputEntries = append(outputEntries, OutputEntry{
			Entity:                            key.Entity,
			SubEntity:                         subEntityOutput,
			BalanceEndSumGwei:                 value.BalanceEndSumGwei,
			EfficiencyDividend:                value.EfficiencyDividend,
			EfficiencyDivisor:                 value.EfficiencyDivisor,
			Efficiency:                        utils.CalcEfficiency(value.EfficiencyDividend, value.EfficiencyDivisor),
			RoiDividend:                       value.RoiDividend,
			RoiDivisor:                        value.RoiDivisor,
			AttestationEfficiency:             utils.CalcEfficiency(value.EfficiencyAttestationsDividend, value.EfficiencyAttestationsDivisor),
			ProposalEfficiency:                utils.CalcEfficiency(value.EfficiencyProposalsDividend, value.EfficiencyProposalsDivisor),
			SyncCommitteeEfficiency:           utils.CalcEfficiency(value.EfficiencySyncDividend, value.EfficiencySyncDivisor),
			EfficiencyTimeBucketTimestampsSec: bucketTimestamps,
			EfficiencyTimeBucketValues:        bucketValues,
			AttestationsScheduledSum:          value.AttestationsScheduledSum,
			AttestationsObservedSum:           value.AttestationsObservedSum,
			AttestationsHeadExecutedSum:       value.AttestationsHeadExecutedSum,
			AttestationsSourceExecutedSum:     value.AttestationsSourceExecutedSum,
			AttestationsTargetExecutedSum:     value.AttestationsTargetExecutedSum,
			AttestationsMissedRewardsSum:      value.AttestationsIdealRewardSum - value.AttestationsRewardsOnlySum,
			AttestationsRewardsOnlySum:        value.AttestationsRewardsOnlySum,
			BlocksScheduledSum:                value.BlocksScheduledSum,
			BlocksProposedSum:                 value.BlocksProposedSum,
			SyncScheduledSum:                  value.SyncScheduledSum,
			SyncExecutedSum:                   value.SyncExecutedSum,
			SlashedInPeriodMax:                value.SlashedInPeriodMax,
			SlashedAmountSum:                  value.SlashedAmountSum,
			BlocksClMissedMedianRewardSum:     value.BlocksClMissedMedianRewardSum,
			SyncLocalizedMaxRewardSum:         value.SyncLocalizedMaxRewardSum,
			SyncRewardRewardsOnlySum:          value.SyncRewardRewardsOnlySum,
			InclusionDelaySum:                 value.InclusionDelaySum,
			ExecutionRewardsSumWei:            value.ExecutionRewardsSumWei,
			NetShare:                          netShare,
			StatusCounts:                      value.StatusCountsByStatus,
		})
	}
	sort.SliceStable(outputEntries, func(i, j int) bool { return outputEntries[i].NetShare < outputEntries[j].NetShare })
	logger.Info("precomputed set calculation completed for period")

	if len(outputEntries) == 0 {
		logger.Info("no output entries to persist")
		return nil
	}

	// Prepare batched arrays for upsert into validator_entities_data_periods
	n := len(outputEntries)
	periods := make([]string, 0, n)
	entities := make([]string, 0, n)
	subEntities := make([]string, 0, n)
	updatedAts := make([]time.Time, 0, n)
	balanceEndSumGwei := make([]int64, 0, n)
	effDividends := make([]int64, 0, n)
	effDivisors := make([]int64, 0, n)
	efficiencies := make([]float64, 0, n)
	roiDividends := make([]string, 0, n)
	roiDivisors := make([]string, 0, n)
	attEffs := make([]float64, 0, n)
	propEffs := make([]float64, 0, n)
	syncEffs := make([]float64, 0, n)
	attSched := make([]int64, 0, n)
	attObs := make([]int64, 0, n)
	attHead := make([]int64, 0, n)
	attSource := make([]int64, 0, n)
	attTarget := make([]int64, 0, n)
	attMissed := make([]int64, 0, n)
	attRewardsOnly := make([]int64, 0, n)
	blkSched := make([]int64, 0, n)
	blkProp := make([]int64, 0, n)
	syncSched := make([]int64, 0, n)
	syncExec := make([]int64, 0, n)
	slashedMax := make([]int64, 0, n)
	slashedSum := make([]int64, 0, n)
	blkClMissed := make([]int64, 0, n)
	syncLocMax := make([]int64, 0, n)
	syncRewardOnly := make([]int64, 0, n)
	inclusionDelay := make([]int64, 0, n)
	netShares := make([]float64, 0, n)
	timeBucketTimestampsLiterals := make([]string, 0, n)
	timeBucketValuesLiterals := make([]string, 0, n)
	statusCountsJSON := make([]string, 0, n)
	execRewardsWei := make([]string, 0, n)

	int64SliceToPgArray := func(a []int64) string {
		if len(a) == 0 {
			return "{}"
		}
		var b strings.Builder
		b.WriteByte('{')
		for i, v := range a {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(fmt.Sprintf("%d", v))
		}
		b.WriteByte('}')
		return b.String()
	}
	float64SliceToPgArray := func(a []float64) string {
		if len(a) == 0 {
			return "{}"
		}
		var b strings.Builder
		b.WriteByte('{')
		for i, v := range a {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
		}
		b.WriteByte('}')
		return b.String()
	}

	now := time.Now().UTC()
	for _, e := range outputEntries {
		periods = append(periods, spec.Label)
		entities = append(entities, e.Entity)
		subEntities = append(subEntities, e.SubEntity)
		updatedAts = append(updatedAts, now)
		balanceEndSumGwei = append(balanceEndSumGwei, e.BalanceEndSumGwei)
		effDividends = append(effDividends, e.EfficiencyDividend.IntPart())
		effDivisors = append(effDivisors, e.EfficiencyDivisor.IntPart())
		efficiencies = append(efficiencies, e.Efficiency)
		roiDividends = append(roiDividends, e.RoiDividend.String())
		roiDivisors = append(roiDivisors, e.RoiDivisor.String())
		attEffs = append(attEffs, e.AttestationEfficiency)
		propEffs = append(propEffs, e.ProposalEfficiency)
		syncEffs = append(syncEffs, e.SyncCommitteeEfficiency)
		bucketTimestampsRow := e.EfficiencyTimeBucketTimestampsSec
		if bucketTimestampsRow == nil {
			bucketTimestampsRow = []int64{}
		}
		bucketValuesRow := e.EfficiencyTimeBucketValues
		if bucketValuesRow == nil {
			bucketValuesRow = []float64{}
		}
		timeBucketTimestampsLiterals = append(timeBucketTimestampsLiterals, int64SliceToPgArray(bucketTimestampsRow))
		timeBucketValuesLiterals = append(timeBucketValuesLiterals, float64SliceToPgArray(bucketValuesRow))
		attSched = append(attSched, e.AttestationsScheduledSum)
		attObs = append(attObs, e.AttestationsObservedSum)
		attHead = append(attHead, e.AttestationsHeadExecutedSum)
		attSource = append(attSource, e.AttestationsSourceExecutedSum)
		attTarget = append(attTarget, e.AttestationsTargetExecutedSum)
		attMissed = append(attMissed, e.AttestationsMissedRewardsSum)
		attRewardsOnly = append(attRewardsOnly, e.AttestationsRewardsOnlySum)
		blkSched = append(blkSched, e.BlocksScheduledSum)
		blkProp = append(blkProp, e.BlocksProposedSum)
		syncSched = append(syncSched, e.SyncScheduledSum)
		syncExec = append(syncExec, e.SyncExecutedSum)
		slashedMax = append(slashedMax, e.SlashedInPeriodMax)
		slashedSum = append(slashedSum, e.SlashedAmountSum)
		blkClMissed = append(blkClMissed, e.BlocksClMissedMedianRewardSum)
		syncLocMax = append(syncLocMax, e.SyncLocalizedMaxRewardSum)
		syncRewardOnly = append(syncRewardOnly, e.SyncRewardRewardsOnlySum)
		inclusionDelay = append(inclusionDelay, e.InclusionDelaySum)
		execRewardsWei = append(execRewardsWei, e.ExecutionRewardsSumWei.String())
		netShares = append(netShares, e.NetShare)
		m := e.StatusCounts
		if m == nil {
			m = map[string]int{}
		}
		b, err := json.Marshal(m)
		if err != nil {
			return fmt.Errorf("marshal status_counts: %w", err)
		}
		statusCountsJSON = append(statusCountsJSON, string(b))
	}

	logger.WithField("rows", n).Info("persisting validator_entities_data_periods rows")
	_, err = tx.Exec(`
		INSERT INTO validator_entities_data_periods (
			entity, sub_entity, period, last_updated_at,
			balance_end_sum_gwei,
			efficiency_dividend, efficiency_divisor, efficiency,
			roi_dividend, roi_divisor,
			attestation_efficiency, proposal_efficiency, sync_committee_efficiency,
			attestations_scheduled_sum, attestations_observed_sum, attestations_head_executed_sum,
			attestations_source_executed_sum, attestations_target_executed_sum,
			attestations_missed_rewards_sum, attestations_reward_rewards_only_sum,
			blocks_scheduled_sum, blocks_proposed_sum,
			sync_scheduled_sum, sync_executed_sum,
			slashed_in_period_max, slashed_amount_sum,
			blocks_cl_missed_median_reward_sum, sync_localized_max_reward_sum, sync_reward_rewards_only_sum,
			inclusion_delay_sum,
			efficiency_time_bucket_timestamps_sec, efficiency_time_bucket_values,
			net_share,
			status_counts,
			execution_rewards_sum_wei
		)
		SELECT
			UNNEST($1::text[]),
			UNNEST($2::text[]),
			UNNEST($3::text[]),
			UNNEST($4::timestamptz[]),
			UNNEST($5::bigint[]),
			UNNEST($6::bigint[]),
			UNNEST($7::bigint[]),
			UNNEST($8::double precision[]),
			UNNEST($9::text[])::numeric,
			UNNEST($10::text[])::numeric,
			UNNEST($11::double precision[]),
			UNNEST($12::double precision[]),
			UNNEST($13::double precision[]),
			UNNEST($14::bigint[]),
			UNNEST($15::bigint[]),
			UNNEST($16::bigint[]),
			UNNEST($17::bigint[]),
			UNNEST($18::bigint[]),
			UNNEST($19::bigint[]),
			UNNEST($20::bigint[]),
			UNNEST($21::bigint[]),
			UNNEST($22::bigint[]),
			UNNEST($23::bigint[]),
			UNNEST($24::bigint[]),
			UNNEST($25::bigint[]),
			UNNEST($26::bigint[]),
			UNNEST($27::bigint[]),
			UNNEST($28::bigint[]),
			UNNEST($29::bigint[]),
			UNNEST($30::bigint[]),
			UNNEST($31::text[])::bigint[],
			UNNEST($32::text[])::double precision[],
			UNNEST($33::double precision[]),
			UNNEST($34::jsonb[]),
			UNNEST($35::text[])::numeric
		ON CONFLICT (entity, sub_entity, period) DO UPDATE SET
			last_updated_at = EXCLUDED.last_updated_at,
			balance_end_sum_gwei = EXCLUDED.balance_end_sum_gwei,
			efficiency_dividend = EXCLUDED.efficiency_dividend,
			efficiency_divisor = EXCLUDED.efficiency_divisor,
			efficiency = EXCLUDED.efficiency,
			roi_dividend = EXCLUDED.roi_dividend,
			roi_divisor = EXCLUDED.roi_divisor,
			attestation_efficiency = EXCLUDED.attestation_efficiency,
			proposal_efficiency = EXCLUDED.proposal_efficiency,
			sync_committee_efficiency = EXCLUDED.sync_committee_efficiency,
			attestations_scheduled_sum = EXCLUDED.attestations_scheduled_sum,
			attestations_observed_sum = EXCLUDED.attestations_observed_sum,
			attestations_head_executed_sum = EXCLUDED.attestations_head_executed_sum,
			attestations_source_executed_sum = EXCLUDED.attestations_source_executed_sum,
			attestations_target_executed_sum = EXCLUDED.attestations_target_executed_sum,
			attestations_missed_rewards_sum = EXCLUDED.attestations_missed_rewards_sum,
			attestations_reward_rewards_only_sum = EXCLUDED.attestations_reward_rewards_only_sum,
			blocks_scheduled_sum = EXCLUDED.blocks_scheduled_sum,
			blocks_proposed_sum = EXCLUDED.blocks_proposed_sum,
			sync_scheduled_sum = EXCLUDED.sync_scheduled_sum,
			sync_executed_sum = EXCLUDED.sync_executed_sum,
			slashed_in_period_max = EXCLUDED.slashed_in_period_max,
			slashed_amount_sum = EXCLUDED.slashed_amount_sum,
			blocks_cl_missed_median_reward_sum = EXCLUDED.blocks_cl_missed_median_reward_sum,
			sync_localized_max_reward_sum = EXCLUDED.sync_localized_max_reward_sum,
			sync_reward_rewards_only_sum = EXCLUDED.sync_reward_rewards_only_sum,
			inclusion_delay_sum = EXCLUDED.inclusion_delay_sum,
			efficiency_time_bucket_timestamps_sec = EXCLUDED.efficiency_time_bucket_timestamps_sec,
			efficiency_time_bucket_values = EXCLUDED.efficiency_time_bucket_values,
			net_share = EXCLUDED.net_share,
			status_counts = EXCLUDED.status_counts,
			execution_rewards_sum_wei = EXCLUDED.execution_rewards_sum_wei
		`,
		pq.Array(entities), pq.Array(subEntities), pq.Array(periods), pq.Array(updatedAts),
		pq.Array(balanceEndSumGwei),
		pq.Array(effDividends), pq.Array(effDivisors), pq.Array(efficiencies),
		pq.Array(roiDividends), pq.Array(roiDivisors),
		pq.Array(attEffs), pq.Array(propEffs), pq.Array(syncEffs),
		pq.Array(attSched), pq.Array(attObs), pq.Array(attHead), pq.Array(attSource), pq.Array(attTarget),
		pq.Array(attMissed), pq.Array(attRewardsOnly),
		pq.Array(blkSched), pq.Array(blkProp),
		pq.Array(syncSched), pq.Array(syncExec),
		pq.Array(slashedMax), pq.Array(slashedSum),
		pq.Array(blkClMissed), pq.Array(syncLocMax), pq.Array(syncRewardOnly),
		pq.Array(inclusionDelay),
		pq.Array(timeBucketTimestampsLiterals), pq.Array(timeBucketValuesLiterals),
		pq.Array(netShares),
		pq.Array(statusCountsJSON),
		pq.Array(execRewardsWei),
	)
	if err != nil {
		return fmt.Errorf("upsert validator_entities_data_periods: %w", err)
	}
	logger.WithField("rows", n).Info("persisted validator_entities_data_periods rows for period")

	// After persisting DB, update Redis cache for index treemap for 30d period
	if utils.Config.RedisSessionStoreEndpoint != "" {
		ctx := context.Background()
		rdc, errInit := cache.InitRedisCache(ctx, utils.Config.RedisSessionStoreEndpoint)
		if errInit != nil {
			logger.WithError(errInit).Warn("treemap cache: failed to init redis; skipping cache write")
		} else {
			var treemapRows []types.EntityTreemapItem
			if errSel := tx.Select(&treemapRows, `
				SELECT entity, efficiency, net_share
				FROM validator_entities_data_periods
				WHERE period = $1 AND sub_entity IN ('-','')
			`, spec.Label); errSel != nil {
				logger.WithError(errSel).Warn("treemap cache: failed to select rows for cache")
			} else {
				key := fmt.Sprintf("%d:entities:treemap:%s", utils.Config.Chain.Id, spec.Label)
				if errSet := rdc.Set(ctx, key, treemapRows, 0); errSet != nil {
					logger.WithError(errSet).WithField("rows", len(treemapRows)).Warn("treemap cache: failed to set redis key")
				} else {
					logger.WithField("rows", len(treemapRows)).Info("treemap cache: updated redis key for 1d treemap")
				}
			}
		}
	}
	return nil
}
