package db

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gobitfly/eth2-beaconchain-explorer/cache"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/singleflight"
)

// EntityDetailData represents the detailed data for a specific entity and sub-entity
type EntityDetailData struct {
	Entity                            string          `db:"entity"`
	SubEntity                         string          `db:"sub_entity"`
	LastUpdatedAt                     time.Time       `db:"last_updated_at"`
	BalanceEndSumGwei                 int64           `db:"balance_end_sum_gwei"`
	Efficiency                        float64         `db:"efficiency"`
	AttestationEfficiency             float64         `db:"attestation_efficiency"`
	ProposalEfficiency                float64         `db:"proposal_efficiency"`
	SyncCommitteeEfficiency           float64         `db:"sync_committee_efficiency"`
	AttestationsScheduledSum          int64           `db:"attestations_scheduled_sum"`
	AttestationsObservedSum           int64           `db:"attestations_observed_sum"`
	AttestationsHeadExecutedSum       int64           `db:"attestations_head_executed_sum"`
	AttestationsSourceExecutedSum     int64           `db:"attestations_source_executed_sum"`
	AttestationsTargetExecutedSum     int64           `db:"attestations_target_executed_sum"`
	AttestationsMissedRewardsSum      int64           `db:"attestations_missed_rewards_sum"`
	AttestationsRewardRewardsOnlySum  int64           `db:"attestations_reward_rewards_only_sum"`
	BlocksScheduledSum                int64           `db:"blocks_scheduled_sum"`
	BlocksProposedSum                 int64           `db:"blocks_proposed_sum"`
	SyncScheduledSum                  int64           `db:"sync_scheduled_sum"`
	SyncExecutedSum                   int64           `db:"sync_executed_sum"`
	SlashedInPeriodMax                int64           `db:"slashed_in_period_max"`
	SlashedAmountSum                  int64           `db:"slashed_amount_sum"`
	BlocksClMissedMedianRewardSum     int64           `db:"blocks_cl_missed_median_reward_sum"`
	SyncLocalizedMaxRewardSum         int64           `db:"sync_localized_max_reward_sum"`
	SyncRewardRewardsOnlySum          int64           `db:"sync_reward_rewards_only_sum"`
	InclusionDelaySum                 int64           `db:"inclusion_delay_sum"`
	EfficiencyTimeBucketTimestampsSec pq.Int64Array   `db:"efficiency_time_bucket_timestamps_sec"`
	EfficiencyTimeBucketValues        pq.Float64Array `db:"efficiency_time_bucket_values"`
	NetShare                          float64         `db:"net_share"`
	StatusCountsRaw                   []byte          `db:"status_counts"`
	RoiDividend                       decimal.Decimal `db:"roi_dividend"`
	RoiDivisor                        decimal.Decimal `db:"roi_divisor"`
	ExecutionRewardsSumWei            decimal.Decimal `db:"execution_rewards_sum_wei"`
}

// EntitySummaryData represents summary data for entities list
type EntitySummaryData struct {
	Entity     string  `db:"entity"`
	Efficiency float64 `db:"efficiency"`
	NetShare   float64 `db:"net_share"`
}

// SubEntityData represents data for sub-entities
type SubEntityData struct {
	Entity     string  `db:"entity"`
	SubEntity  string  `db:"sub_entity"`
	Efficiency float64 `db:"efficiency"`
	NetShare   float64 `db:"net_share"`
}

// internal lazy Redis init for treemap caching
var treemapCacheOnce sync.Once
var treemapCache *cache.RedisCache
var treemapSF singleflight.Group

func getTreemapCache() *cache.RedisCache {
	treemapCacheOnce.Do(func() {
		if utils.Config.RedisSessionStoreEndpoint == "" {
			return
		}
		ctx := context.Background()
		c, err := cache.InitRedisCache(ctx, utils.Config.RedisSessionStoreEndpoint)
		if err != nil {
			logger.WithError(err).Warn("treemap: failed to init redis cache; falling back to DB")
			return
		}
		treemapCache = c
		logger.WithField("redis", utils.Config.RedisSessionStoreEndpoint).Info("treemap: redis cache initialized")
	})
	return treemapCache
}

// GetEntitiesTreemapData returns the pre-aggregated entity-level rows used by the treemap
// on Entities Overview and Index pages. It selects only entity rows (sub_entity is '-' or empty)
// for the requested period.
// It now reads from Redis first (key: <chainId>:entities:treemap:<period>) and falls back to DB on miss.
// On miss, it warms the cache using singleflight to avoid load spikes.
func GetEntitiesTreemapData(period string) ([]types.EntityTreemapItem, error) {
	cacheClient := getTreemapCache()
	key := fmt.Sprintf("%d:entities:treemap:%s", utils.Config.Chain.Id, period)

	// Try cache first if configured
	if cacheClient != nil {
		var cached []types.EntityTreemapItem
		if _, err := cacheClient.Get(context.Background(), key, &cached); err == nil && len(cached) > 0 {
			logger.WithFields(map[string]interface{}{"period": period, "rows": len(cached)}).Debug("treemap: cache hit")
			return cached, nil
		}
		logger.WithField("period", period).Warn("treemap: cache miss")
	}

	// Fetch from DB (singleflight) and optionally warm cache
	v, err, _ := treemapSF.Do(key, func() (interface{}, error) {
		rows := make([]types.EntityTreemapItem, 0, 4096)
		err := ReaderDb.Select(&rows, `
			SELECT entity, efficiency, net_share
			FROM validator_entities_data_periods
			WHERE period = $1 AND sub_entity IN ('-','')
		`, period)
		if err != nil {
			return nil, err
		}
		if cacheClient != nil {
			if err := cacheClient.Set(context.Background(), key, rows, 0); err != nil {
				logger.WithError(err).WithField("period", period).Warn("treemap: failed to set cache")
			} else {
				logger.WithFields(map[string]interface{}{"period": period, "rows": len(rows)}).Info("treemap: cache warmed from DB")
			}
		}
		return rows, nil
	})
	if err != nil {
		return nil, err
	}
	return v.([]types.EntityTreemapItem), nil
}

// GetEntityDetailData retrieves detailed data for a specific entity, sub-entity, and period
func GetEntityDetailData(entity, subEntity, period string) (*EntityDetailData, error) {
	var data EntityDetailData
	err := ReaderDb.Get(&data, `
		SELECT entity, sub_entity, last_updated_at,
		       balance_end_sum_gwei,
		       efficiency,
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
		       status_counts, roi_dividend, roi_divisor, execution_rewards_sum_wei
		FROM validator_entities_data_periods
		WHERE entity = $1 AND sub_entity = $2 AND period = $3
	`, entity, subEntity, period)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// HasRealSubEntities checks if there are any real sub-entities for an entity (excluding the default '-')
func HasRealSubEntities(entity, period string) (bool, error) {
	var hasReal bool
	err := ReaderDb.Get(&hasReal, `
		SELECT EXISTS(
			SELECT 1 FROM validator_entities_data_periods WHERE entity = $1 AND period = $2 AND sub_entity <> '-'
		)
	`, entity, period)
	return hasReal, err
}

// CountSubEntities returns the total count of sub-entities for a given entity and period
func CountSubEntities(entity, period string) (int, error) {
	var count int
	err := ReaderDb.Get(&count, `
		SELECT COUNT(*)
		FROM validator_entities_data_periods
		WHERE entity = $1 AND period = $2 AND sub_entity <> '-'
	`, entity, period)
	return count, err
}

// GetSubEntitiesPaginated retrieves paginated sub-entities for a given entity and period
func GetSubEntitiesPaginated(entity, period string, limit, offset int) ([]SubEntityData, error) {
	var subEntities []SubEntityData
	err := ReaderDb.Select(&subEntities, `
		SELECT sub_entity, net_share, efficiency
		FROM validator_entities_data_periods
		WHERE entity = $1 AND period = $2 AND sub_entity <> '-'
		ORDER BY net_share DESC, sub_entity ASC
		LIMIT $3 OFFSET $4
	`, entity, period, limit, offset)
	return subEntities, err
}

// CountEntitiesWithSearch counts entities that match the search criteria
func CountEntitiesWithSearch(period, searchTerm string) (int, error) {
	var count int
	err := ReaderDb.Get(&count, `
		WITH matches AS (
			SELECT DISTINCT entity
			FROM validator_entities_data_periods
			WHERE (entity ILIKE ($1 || '%')
			   OR (sub_entity NOT IN ('-','') AND sub_entity ILIKE ($1 || '%')))
			  AND period = $2
		)
		SELECT COUNT(*) FROM matches
	`, searchTerm, period)
	return count, err
}

// CountEntities counts all entities for a given period
func CountEntities(period string) (int, error) {
	var count int
	err := ReaderDb.Get(&count, `
		SELECT COUNT(*)
		FROM validator_entities_data_periods
		WHERE period = $1 AND sub_entity IN ('-','')
	`, period)
	return count, err
}

// GetEntitiesPagedWithSearch retrieves paginated entities that match search criteria
func GetEntitiesPagedWithSearch(period, searchTerm string, limit, offset int) ([]EntitySummaryData, error) {
	var entities []EntitySummaryData
	err := ReaderDb.Select(&entities, `
		WITH matches AS (
			SELECT DISTINCT entity
			FROM validator_entities_data_periods
			WHERE (entity ILIKE ($1 || '%')
			   OR (sub_entity NOT IN ('-','') AND sub_entity ILIKE ($1 || '%')))
			  AND period = $2
		)
		SELECT v.entity, v.efficiency, v.net_share
		FROM validator_entities_data_periods v
		JOIN matches m ON m.entity = v.entity
		WHERE v.sub_entity IN ('-','') AND v.period = $2
		ORDER BY v.net_share DESC, v.entity ASC
		LIMIT $3 OFFSET $4
	`, searchTerm, period, limit, offset)
	return entities, err
}

// GetEntitiesPaged retrieves paginated entities without search
func GetEntitiesPaged(period string, limit, offset int) ([]EntitySummaryData, error) {
	var entities []EntitySummaryData
	err := ReaderDb.Select(&entities, `
		SELECT entity, efficiency, net_share
		FROM validator_entities_data_periods
		WHERE period = $1 AND sub_entity IN ('-','')
		ORDER BY net_share DESC, entity ASC
		LIMIT $2 OFFSET $3
	`, period, limit, offset)
	return entities, err
}

// GetSubEntitiesForEntitiesWithSearch retrieves sub-entities for given entities that match search criteria
func GetSubEntitiesForEntitiesWithSearch(entityNames []string, searchTerm, period string) ([]SubEntityData, error) {
	var subEntities []SubEntityData
	err := ReaderDb.Select(&subEntities, `
		WITH ranked AS (
			SELECT entity, sub_entity, efficiency, net_share,
			       ROW_NUMBER() OVER (PARTITION BY entity ORDER BY net_share DESC, sub_entity ASC) AS rn
			FROM validator_entities_data_periods
			WHERE period = $3
			  AND entity = ANY($1)
			  AND sub_entity NOT IN ('-','')
			  AND sub_entity ILIKE ($2 || '%')
		)
		SELECT entity, sub_entity, efficiency, net_share
		FROM ranked
		WHERE rn <= 101
		ORDER BY entity ASC, net_share DESC, sub_entity ASC
	`, pq.Array(entityNames), searchTerm, period)
	return subEntities, err
}

// GetSubEntitiesForEntities retrieves sub-entities for given entities without search
func GetSubEntitiesForEntities(entityNames []string, period string) ([]SubEntityData, error) {
	var subEntities []SubEntityData
	err := ReaderDb.Select(&subEntities, `
		WITH ranked AS (
			SELECT entity, sub_entity, efficiency, net_share,
			       ROW_NUMBER() OVER (PARTITION BY entity ORDER BY net_share DESC, sub_entity ASC) AS rn
			FROM validator_entities_data_periods
			WHERE period = $2
			  AND entity = ANY($1)
			  AND sub_entity NOT IN ('-','')
		)
		SELECT entity, sub_entity, efficiency, net_share
		FROM ranked
		WHERE rn <= 101
		ORDER BY entity ASC, net_share DESC, sub_entity ASC
	`, pq.Array(entityNames), period)
	return subEntities, err
}

// GetSubEntityCountsForEntities returns the number of real sub-entities (sub_entity <> '-') per entity for the given period.
func GetSubEntityCountsForEntities(entityNames []string, period string) (map[string]int, error) {
	type countRow struct {
		Entity string `db:"entity"`
		Count  int    `db:"cnt"`
	}
	var rows []countRow
	err := ReaderDb.Select(&rows, `
		SELECT entity, COUNT(*) AS cnt
		FROM validator_entities_data_periods
		WHERE period = $2 AND entity = ANY($1) AND sub_entity <> '-'
		GROUP BY entity
	`, pq.Array(entityNames), period)
	if err != nil {
		return nil, err
	}
	result := make(map[string]int, len(rows))
	for _, r := range rows {
		result[r.Entity] = r.Count
	}
	return result, nil
}
