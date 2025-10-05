package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// EntityValidatorRow represents a single validator row for the entity validators table section.
// It contains the minimum fields needed for the view.
// Note: Efficiency is filled by a ClickHouse query separately and may be -1 if unavailable.
type EntityValidatorRow struct {
	Index      int    `db:"validatorindex"`
	Status     string `db:"status"`
	Efficiency float64
}

// CountEntityValidators returns the total number of validators belonging to the given
// entity and sub-entity. When subEntity is "-" (or empty), validators across all sub-entities
// of the entity are counted.
func CountEntityValidators(entity string, subEntity string) (int, error) {
	var count int
	// Sub-entity filtering: if subEntity is '-', we include all sub-entities (including NULL/empty)
	query := `
		SELECT COUNT(*)
		FROM validator_entities ve
		JOIN validators v ON v.pubkey = ve.publickey
		WHERE ve.entity = $1
		  AND ($2 = '-' OR COALESCE(ve.sub_entity, '') = $2)
	`
	if err := ReaderDb.Get(&count, query, entity, subEntity); err != nil {
		return 0, fmt.Errorf("count entity validators: %w", err)
	}
	return count, nil
}

// GetEntityValidatorsPaginated returns a page of validators for an entity and sub-entity,
// ordered by validator index descending. When subEntity is '-' (or empty), returns validators
// across all sub-entities for that entity.
func GetEntityValidatorsPaginated(entity string, subEntity string, limit int, offset int) ([]EntityValidatorRow, error) {
	rows := make([]EntityValidatorRow, 0, limit)
	query := `
		SELECT v.validatorindex, v.status
		FROM validator_entities ve
		JOIN validators v ON v.pubkey = ve.publickey
		WHERE ve.entity = $1
		  AND ($2 = '-' OR COALESCE(ve.sub_entity, '') = $2)
		ORDER BY v.validatorindex DESC
		LIMIT $3 OFFSET $4
	`
	if err := ReaderDb.Select(&rows, query, entity, subEntity, limit, offset); err != nil {
		return nil, fmt.Errorf("select entity validators: %w", err)
	}
	return rows, nil
}

// GetValidatorEfficienciesForPeriod fetches beacon score (efficiency) for the provided validator indices
// for the given period label ("1d", "7d", "30d", "all"). It returns a map from index to efficiency in [0,1]
// or -1 when not available.
// The function queries ClickHouse rolling tables consistent with precomputeEntityData PeriodSpec.
func GetValidatorEfficienciesForPeriod(period string, indices []int) (map[int]float64, error) {
	result := make(map[int]float64, len(indices))
	if len(indices) == 0 {
		return result, nil
	}
	if ClickhouseReaderDb == nil {
		// ClickHouse not enabled; leave map empty
		return result, nil
	}
	var table string
	switch period {
	case "1d":
		table = "_final_validator_dashboard_rolling_24h"
	case "7d":
		table = "_final_validator_dashboard_rolling_7d"
	case "30d":
		table = "_final_validator_dashboard_rolling_30d"
	case "all":
		table = "_final_validator_dashboard_rolling_total"
	default:
		table = "_final_validator_dashboard_rolling_24h"
	}
	// Build a temporary table parameter list (ClickHouse supports IN with array join via ANY). We pass as []int
	// and rely on the SQL driver to expand.
	type row struct {
		ValidatorIndex int     `db:"validator_index"`
		Dividend       int64   `db:"efficiency_dividend"`
		Divisor        int64   `db:"efficiency_divisor"`
		Efficiency     float64 // computed after scan
	}
	// Note: We use arrayJoin for filtering by validator_index list and aggregate in case of duplicates.
	query := fmt.Sprintf(`
		SELECT validator_index, sum(efficiency_dividend) AS efficiency_dividend, sum(efficiency_divisor) AS efficiency_divisor
		FROM %s
		WHERE validator_index IN (?)
		GROUP BY validator_index
	`, table)
	// Use sqlx.In to expand the IN clause safely
	q, args, err := sqlx.In(query, indices)
	if err != nil {
		return result, fmt.Errorf("sqlx.In build for CH: %w", err)
	}
	q = ClickhouseReaderDb.Rebind(q)
	var rows []row
	if err := ClickhouseReaderDb.Select(&rows, q, args...); err != nil {
		return result, fmt.Errorf("clickhouse select efficiencies: %w", err)
	}
	for _, r := range rows {
		if r.Divisor <= 0 {
			result[r.ValidatorIndex] = -1
			continue
		}
		eff := float64(r.Dividend) / float64(r.Divisor)
		if eff > 1 {
			eff = 1
		}
		result[r.ValidatorIndex] = eff
	}
	return result, nil
}
