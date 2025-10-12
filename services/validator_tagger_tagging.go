package services

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// autoTagUntaggedValidatorsWithdrawal applies withdrawal address based tags for all untagged validators
// based on their withdrawal credentials. It groups untagged validators by withdrawal
// credentials (only 0x01/0x02), sums their current balances from ClickHouse
// validator_dashboard_data_rolling_total, filters groups with total balance >= 320e9 gwei,
// extracts the withdrawal address from the last 20 bytes and inserts a row into
// validator_entities for each validator pubkey in qualifying groups using the address as entity name.
func autoTagUntaggedValidatorsWithdrawal(ctx context.Context) (map[uint64]int64, error) {
	logger := validatorTaggerLogger.WithField("step", "auto_tag_by_withdrawal")
	logger.Info("starting withdrawal-based validator tagging...")
	type untagged struct {
		Index                 uint64 `db:"validatorindex"`
		Pubkey                []byte `db:"pubkey"`
		WithdrawalCredentials []byte `db:"withdrawalcredentials"`
	}
	var untaggedVals []untagged

	logger.Infof("fetching untagged validators with withdrawal credentials (0x01/0x02) from DB...")
	q := `
		SELECT v.validatorindex, v.pubkey, v.withdrawalcredentials
		FROM validators v
		LEFT JOIN validator_entities ve ON v.pubkey = ve.publickey
		WHERE ve.publickey IS NULL
		  AND substring(v.withdrawalcredentials from 1 for 1) IN (E'\\x01'::bytea, E'\\x02'::bytea)
	`
	if err := db.ReaderDb.Select(&untaggedVals, q); err != nil {
		return nil, fmt.Errorf("select untagged validators: %w", err)
	}
	if len(untaggedVals) == 0 {
		validatorTaggerLogger.Info("no untagged validators found for withdrawal tagging")
		return nil, nil
	}

	// Collect indices for CH query
	indices := make([]uint64, 0, len(untaggedVals))
	for _, u := range untaggedVals {
		indices = append(indices, u.Index)
	}

	// Fetch balances from ClickHouse in batches to avoid exceeding max_query_size
	type balRow struct {
		ValidatorIndex uint64 `db:"validator_index"`
		BalanceEnd     int64  `db:"balance_end"`
	}
	balanceByIndex := make(map[uint64]int64, len(indices))
	// Note: IN (?) with sqlx expands the slice param for clickhouse driver; keep batches small
	const chBatchSize = 10000
	chQuery := `SELECT validator_index, max(finalizeAggregation(balance_end)) AS balance_end FROM validator_dashboard_data_rolling_total WHERE validator_index IN (?) GROUP BY validator_index`
	for start := 0; start < len(indices); start += chBatchSize {
		logger.Infof("fetching balances for %d-%d from CH...", start, start+chBatchSize)
		end := start + chBatchSize
		if end > len(indices) {
			end = len(indices)
		}
		var chunkRows []balRow
		if err := db.ClickhouseReaderDb.Select(&chunkRows, chQuery, indices[start:end]); err != nil {
			return nil, fmt.Errorf("clickhouse select balances (batch %d-%d): %w", start, end, err)
		}
		for _, r := range chunkRows {
			balanceByIndex[r.ValidatorIndex] = r.BalanceEnd
		}
	}

	// Group by withdrawal credentials
	type group struct {
		sum     int64
		members []int // indices into untaggedVals slice
	}
	groups := make(map[string]*group) // key: hex withdrawal credentials (lowercase, 0x-prefixed)
	for i, u := range untaggedVals {
		wcHex := "0x" + strings.ToLower(hex.EncodeToString(u.WithdrawalCredentials))
		g := groups[wcHex]
		if g == nil {
			g = &group{}
			groups[wcHex] = g
		}
		g.sum += balanceByIndex[u.Index]
		g.members = append(g.members, i)
	}

	const threshold int64 = 320000000000 // 320 eth in gwei

	// Prepare batch inserts
	pubkeys := make([][]byte, 0, len(untaggedVals))
	entities := make([]any, 0, len(untaggedVals))

	for wcHex, g := range groups {
		if g.sum < threshold {
			continue
		}
		// Extract address from last 20 bytes
		bytesWc := untaggedVals[g.members[0]].WithdrawalCredentials
		if len(bytesWc) < 20 { // defensive
			validatorTaggerLogger.WithField("wc", wcHex).Warn("withdrawal credentials shorter than 20 bytes; skipping group")
			continue
		}
		addr := bytesWc[len(bytesWc)-20:]
		addrHex := "0x" + strings.ToLower(hex.EncodeToString(addr))

		for _, idx := range g.members {
			u := untaggedVals[idx]
			pubkeys = append(pubkeys, u.Pubkey)
			entities = append(entities, addrHex)
		}
	}

	if len(pubkeys) == 0 {
		validatorTaggerLogger.Info("no qualifying groups for withdrawal tagging (below balance threshold)")
		return balanceByIndex, nil
	}

	// Batch insert with UNNEST and ON CONFLICT DO NOTHING
	logger.Infof("inserting %d withdrawal validator entities...", len(pubkeys))
	_, err := db.WriterDb.Exec(`
		INSERT INTO validator_entities (publickey, entity)
		SELECT UNNEST($1::bytea[]), UNNEST($2::text[])
		ON CONFLICT (publickey) DO NOTHING
	`, pq.ByteaArray(pubkeys), pq.Array(entities))
	if err != nil {
		return nil, fmt.Errorf("insert validator_entities withdrawal tags: %w", err)
	}
	validatorTaggerLogger.WithFields(logrus.Fields{"inserted": len(pubkeys)}).Info("withdrawal validator tagging completed")
	return balanceByIndex, nil
}

// autoTagUntaggedValidatorsByDeposit applies deposit address based tags for remaining untagged validators
// by grouping them on their earliest eth1 deposit from_address. It sums current balances
// from ClickHouse and tags groups with total >= threshold using the from_address as entity.
func autoTagUntaggedValidatorsByDeposit(ctx context.Context, balanceByIndex map[uint64]int64) error {
	logger := validatorTaggerLogger.WithField("step", "auto_tag_by_deposit")
	logger.Info("starting deposit-based validator tagging...")
	if len(balanceByIndex) == 0 {
		validatorTaggerLogger.Warn("no balances provided; skipping deposit-based tagging")
		return nil
	}

	type untaggedDep struct {
		Index       uint64 `db:"validatorindex"`
		Pubkey      []byte `db:"pubkey"`
		FromAddress []byte `db:"from_address"`
	}
	var untaggedVals []untaggedDep

	logger.Info("fetching remaining untagged validators with earliest deposit from_address from DB...")
	q := `
		SELECT v.validatorindex, v.pubkey, d.from_address
		FROM validators v
		LEFT JOIN validator_entities ve ON v.pubkey = ve.publickey
		JOIN LATERAL (
			SELECT ed.from_address
			FROM eth1_deposits ed
			WHERE ed.publickey = v.pubkey
			ORDER BY ed.block_number ASC
			LIMIT 1
		) d ON TRUE
		WHERE ve.publickey IS NULL
	`
	if err := db.ReaderDb.Select(&untaggedVals, q); err != nil {
		return fmt.Errorf("select untagged validators by deposit: %w", err)
	}
	if len(untaggedVals) == 0 {
		validatorTaggerLogger.Info("no untagged validators found for deposit-based tagging")
		return nil
	}

	// Group by from_address
	type group2 struct {
		sum     int64
		members []int
	}
	groups := make(map[string]*group2)
	for i, u := range untaggedVals {
		addrHex := "0x" + strings.ToLower(hex.EncodeToString(u.FromAddress))
		// Validate and normalize the address length to 20 bytes if needed
		if len(u.FromAddress) == 0 {
			continue
		}
		if !utils.IsValidEth1Address(addrHex) {
			// continue on invalid address
			continue
		}
		g := groups[addrHex]
		if g == nil {
			g = &group2{}
			groups[addrHex] = g
		}
		g.sum += balanceByIndex[u.Index]
		g.members = append(g.members, i)
	}

	const threshold int64 = 320000000000 // gwei

	pubkeys := make([][]byte, 0, len(untaggedVals))
	entities := make([]any, 0, len(untaggedVals))

	for addrHex, g := range groups {
		if g.sum < threshold {
			continue
		}
		for _, idx := range g.members {
			u := untaggedVals[idx]
			pubkeys = append(pubkeys, u.Pubkey)
			entities = append(entities, addrHex)
		}
	}

	if len(pubkeys) == 0 {
		validatorTaggerLogger.Info("no qualifying deposit groups for tagging (below balance threshold)")
		return nil
	}

	logger.Infof("inserting %d deposit-based validator entities...", len(pubkeys))
	_, err := db.WriterDb.Exec(`
		INSERT INTO validator_entities (publickey, entity)
		SELECT UNNEST($1::bytea[]), UNNEST($2::text[])
		ON CONFLICT (publickey) DO NOTHING
	`, pq.ByteaArray(pubkeys), pq.Array(entities))
	if err != nil {
		return fmt.Errorf("insert validator_entities deposit tags: %w", err)
	}
	validatorTaggerLogger.WithFields(logrus.Fields{"inserted": len(pubkeys)}).Info("deposit-based validator tagging completed")
	return nil
}

// populateValidatorNamesTable populates validator_names from validator_entities in a single SQL statement.
// Existing entries are preserved (ON CONFLICT DO NOTHING). If entity starts with 0x/0X, the inserted
// name is Whale_{first_8_chars_lowercased_including_0x}. Otherwise, the entity string is used as-is.
func populateValidatorNamesTable(ctx context.Context) error {
	logger := validatorTaggerLogger.WithField("step", "populate_validator_names")
	logger.Info("populating validator_names from validator_entities")
	res, err := db.WriterDb.Exec(`
		INSERT INTO validator_names (publickey, name)
		SELECT ve.publickey,
		       CASE
		           WHEN lower(ve.entity) LIKE '0x%' THEN 'Whale_' || substr(lower(ve.entity), 1, 8)
		           ELSE ve.entity
		       END AS name
		FROM validator_entities ve
		LEFT JOIN validator_names vn ON vn.publickey = ve.publickey
		WHERE vn.publickey IS NULL
		  AND ve.entity IS NOT NULL AND ve.entity <> ''
		ON CONFLICT (publickey) DO NOTHING;
	`)
	if err != nil {
		return fmt.Errorf("populate validator_names from validator_entities: %w", err)
	}
	affected, _ := res.RowsAffected()
	logger.WithField("inserted", affected).Info("populate validator_names completed")
	return nil
}
