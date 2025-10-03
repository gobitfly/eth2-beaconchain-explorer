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
// entity and sub-entity. It now uses precomputed aggregates from validator_entities_data_periods
// and sums all values in the status_counts JSONB field for period '1d'.
// When subEntity is "-" (or empty), the function counts validators for the entity aggregate row
// where sub_entity is either '-' or ”. Otherwise, it counts for the specific sub-entity row.
func CountEntityValidators(entity string, subEntity string) (int, error) {
	var count int
	// Use jsonb_each_text to iterate key/value pairs and sum numeric values safely
	// We always use the '1d' period per requirements
	query := `
		SELECT COALESCE(SUM((kv.value)::bigint), 0) AS total
		FROM validator_entities_data_periods vedp
		JOIN LATERAL jsonb_each_text(vedp.status_counts) AS kv(key, value) ON TRUE
		WHERE vedp.period = '1d'
		  AND vedp.entity = $1
		  AND ((($2 = '-') AND vedp.sub_entity IN ('-','')) OR (($2 <> '-') AND vedp.sub_entity = $2))
	`
	if subEntity == "" { // treat empty the same as '-'
		subEntity = "-"
	}
	if err := ReaderDb.Get(&count, query, entity, subEntity); err != nil {
		return 0, fmt.Errorf("count entity validators (precomputed): %w", err)
	}
	return count, nil
}

// GetEntityValidatorsByCursor returns validators for an entity/sub-entity using keyset (seek) pagination.
// If afterIndex is nil, it returns the first page ordered by validatorindex DESC.
// If afterIndex is provided, it returns the next page where validatorindex < *afterIndex.
func GetEntityValidatorsByCursor(entity string, subEntity string, limit int, afterIndex *int) ([]EntityValidatorRow, error) {
	rows := make([]EntityValidatorRow, 0, limit)
	if subEntity == "" {
		subEntity = "-"
	}
	var query string
	var args []interface{}
	if afterIndex == nil {
		// First page
		query = `
			SELECT v.validatorindex, v.status
			FROM validators v
			WHERE EXISTS (
			  SELECT 1
			  FROM validator_entities ve
			  WHERE ve.publickey = v.pubkey
			    AND ve.entity = $1
			    AND ($2 = '-' OR COALESCE(ve.sub_entity, '') = $2)
			)
			ORDER BY v.validatorindex DESC
			LIMIT $3;
		`
		args = []interface{}{entity, subEntity, limit}
	} else {
		// Next page
		query = `
			SELECT v.validatorindex, v.status
			FROM validators v
			WHERE v.validatorindex < $3
			  AND EXISTS (
			    SELECT 1 FROM validator_entities ve
			    WHERE ve.publickey = v.pubkey
			      AND ve.entity = $1
			      AND ($2 = '-' OR COALESCE(ve.sub_entity, '') = $2)
			  )
			ORDER BY v.validatorindex DESC
			LIMIT $4;
		`
		args = []interface{}{entity, subEntity, *afterIndex, limit}
	}
	if err := ReaderDb.Select(&rows, query, args...); err != nil {
		return nil, fmt.Errorf("select entity validators (keyset): %w", err)
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

	// seed the result with -1 to account for validators not yet present in the rolling table
	for _, index := range indices {
		result[index] = -1
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
