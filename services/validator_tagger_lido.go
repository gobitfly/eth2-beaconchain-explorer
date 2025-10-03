package services

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/lido"
	"github.com/gobitfly/eth2-beaconchain-explorer/metrics"
	"github.com/gobitfly/eth2-beaconchain-explorer/rpc"
	"github.com/lib/pq"
)

var (
	operatorRegistryAddress = common.HexToAddress("0x55032650b14df07b85bF18A3a3eC8E0Af2e028d5")
	csmModuleAddress        = common.HexToAddress("0xdA7dE2ECdDfccC6c3AF10108Db212ACBBf9EA83F")
	simpleDVTModuleAddress  = common.HexToAddress("0xaE7B191A31f627b4eB1d4DaC64eaB9976995b433")
	lidoLogger              = logger.WithField("module", "lido")
)

// indexLidoValidators indexes Lido node operators and their signing keys from the on-chain
// OperatorRegistry contract and persists them to Postgres.
//
// Summary of behavior:
//   - Connects to the Lido OperatorRegistry (operatorRegistryAddress) via the current
//     Erigon RPC client and reads the total number of node operators.
//   - Iterates all operators and loads their metadata: active flag, name, reward address
//     and validator totals, plus the total signing key count. These are stored in-memory
//     as operator rows.
//   - Opens a single write-transaction and reads existing rows from lido_node_operators.
//     It compares on-chain data with DB state and only upserts operators whose properties
//     changed. Upserts are done in batches using parameterized arrays with UNNEST and
//     ON CONFLICT (operator_id) DO UPDATE, minimizing unnecessary writes.
//   - After operators are persisted (satisfying the FK), it fetches signing keys for each
//     operator from the chain in fixed-size batches (48-byte pubkeys). It then bulks inserts
//     (operator_id, pubkey) pairs into lido_signing_keys using UNNEST with
//     ON CONFLICT DO NOTHING to make the operation idempotent.
//   - All SQL statements are parameterized; no per-row statements are executed.
//   - The whole run is executed within a single transaction for consistency; the
//     transaction is committed at the end or rolled back on error.
//
// Tables used/updated:
//   - lido_node_operators(operator_id PK, active, name, reward_address, total_* fields, signing_key_count)
//   - lido_signing_keys(operator_id FK -> lido_node_operators, pubkey, PK(operator_id, pubkey))
//
// Batching and performance:
//   - Operator upserts are performed in batches of up to 1000 rows per statement.
//   - Signing keys are fetched per-operator in batches of up to 1000 keys from chain,
//     sliced into 48-byte pubkeys for DB insertion.
//
// Preconditions:
//   - db.MustInitDB has been called and db.WriterDb is available.
//   - rpc.CurrentErigonClient is initialized and connected to an Ethereum node.
//
// Returns nil on success or a descriptive error. It does not log-fatal; callers (binaries)
// are expected to handle errors appropriately.
func indexLidoValidators() error {
	operatorRegistry, err := lido.NewOperatorRegistry(operatorRegistryAddress, rpc.CurrentErigonClient.GetNativeClient())
	if err != nil {
		return fmt.Errorf("failed to create operator registry: %w", err)
	}

	nodeOperatorsCounter, err := operatorRegistry.GetNodeOperatorsCount(nil)
	if err != nil {
		return fmt.Errorf("failed to retrieve node operators count: %w", err)
	}
	nodeOperatorsCounterInt64 := nodeOperatorsCounter.Int64()
	lidoLogger.Infof("node operators count: %d", nodeOperatorsCounterInt64)

	// collect operator rows to batch-insert after we process all operators
	type operatorRow struct {
		OperatorID      int64  `db:"operator_id"`
		Active          bool   `db:"active"`
		Name            string `db:"name"`
		RewardAddress   []byte `db:"reward_address"`
		TotalVetted     int64  `db:"total_vetted_validators"`
		TotalExited     int64  `db:"total_exited_validators"`
		TotalAdded      int64  `db:"total_added_validators"`
		TotalDeposited  int64  `db:"total_deposited_validators"`
		SigningKeyCount int64  `db:"signing_key_count"`
	}
	operatorRows := make([]operatorRow, 0, nodeOperatorsCounterInt64)

	// use a single transaction for the whole run to keep things consistent
	tx, err := db.WriterDb.Beginx()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// now process each node operator one by one
	for nodeOperatorIndex := int64(0); nodeOperatorIndex < nodeOperatorsCounterInt64; nodeOperatorIndex++ {
		nodeOperatorIndexBig := big.NewInt(nodeOperatorIndex)
		// retrieve details on the node operator
		nodeOperatorDetails, err := operatorRegistry.GetNodeOperator(nil, nodeOperatorIndexBig, true)
		if err != nil {
			return fmt.Errorf("failed to retrieve node operator details for operator %d: %w", nodeOperatorIndex, err)
		}

		operatorSigningKeyCount, err := operatorRegistry.GetTotalSigningKeyCount(nil, nodeOperatorIndexBig)
		if err != nil {
			return fmt.Errorf("failed to retrieve signing key count for operator %d: %w", nodeOperatorIndex, err)
		}
		operatorSigningKeyCountInt64 := operatorSigningKeyCount.Int64()

		// add the operator row now that we also have the signing key count
		operatorRows = append(operatorRows, operatorRow{
			OperatorID:      nodeOperatorIndex,
			Active:          nodeOperatorDetails.Active,
			Name:            nodeOperatorDetails.Name,
			RewardAddress:   nodeOperatorDetails.RewardAddress.Bytes(),
			TotalVetted:     int64(nodeOperatorDetails.TotalVettedValidators),
			TotalExited:     int64(nodeOperatorDetails.TotalExitedValidators),
			TotalAdded:      int64(nodeOperatorDetails.TotalAddedValidators),
			TotalDeposited:  int64(nodeOperatorDetails.TotalDepositedValidators),
			SigningKeyCount: operatorSigningKeyCountInt64,
		})

		lidoLogger.Infof("processed node operator %s (%d)", nodeOperatorDetails.Name, nodeOperatorIndex)

	}

	// Compare with DB and only upsert changed operators; also track signing_key_count changes
	var existingRows []operatorRow
	if err := tx.Select(&existingRows, `
		SELECT operator_id, active, name, reward_address, total_vetted_validators, total_exited_validators, total_added_validators, total_deposited_validators, signing_key_count
		FROM lido_node_operators
		WHERE operator_id >= $1 AND operator_id < $2
	`, 0, nodeOperatorsCounterInt64); err != nil {
		return fmt.Errorf("select existing lido_node_operators: %w", err)
	}
	existingByID := make(map[int64]operatorRow, len(existingRows))
	for _, e := range existingRows {
		existingByID[e.OperatorID] = e
	}

	rowsToUpsert := make([]operatorRow, 0, len(operatorRows))
	countsChanged := make(map[int64]bool, len(operatorRows))
	for _, r := range operatorRows {
		if e, ok := existingByID[r.OperatorID]; !ok {
			rowsToUpsert = append(rowsToUpsert, r)
			countsChanged[r.OperatorID] = true
			continue
		} else {
			same := e.Active == r.Active &&
				e.Name == r.Name &&
				bytes.Equal(e.RewardAddress, r.RewardAddress) &&
				e.TotalVetted == r.TotalVetted &&
				e.TotalExited == r.TotalExited &&
				e.TotalAdded == r.TotalAdded &&
				e.TotalDeposited == r.TotalDeposited &&
				e.SigningKeyCount == r.SigningKeyCount
			if !same {
				rowsToUpsert = append(rowsToUpsert, r)
			}
			if e.SigningKeyCount != r.SigningKeyCount {
				countsChanged[r.OperatorID] = true
			}
		}
	}

	lidoLogger.Infof("lido operators total: %d, to upsert: %d, signing-key-count changed: %d", len(operatorRows), len(rowsToUpsert), len(countsChanged))

	// upsert only changed operator rows in batches using UNNEST arrays
	if len(rowsToUpsert) > 0 {
		batchSize := 1000
		for b := 0; b < len(rowsToUpsert); b += batchSize {
			start := b
			end := b + batchSize
			if len(rowsToUpsert) < end {
				end = len(rowsToUpsert)
			}

			ids := make([]int64, 0, end-start)
			actives := make([]bool, 0, end-start)
			names := make([]string, 0, end-start)
			rewards := make([][]byte, 0, end-start)
			vetted := make([]int64, 0, end-start)
			exited := make([]int64, 0, end-start)
			added := make([]int64, 0, end-start)
			deposited := make([]int64, 0, end-start)
			keyCounts := make([]int64, 0, end-start)

			for _, r := range rowsToUpsert[start:end] {
				ids = append(ids, r.OperatorID)
				actives = append(actives, r.Active)
				names = append(names, r.Name)
				rewards = append(rewards, r.RewardAddress)
				vetted = append(vetted, r.TotalVetted)
				exited = append(exited, r.TotalExited)
				added = append(added, r.TotalAdded)
				deposited = append(deposited, r.TotalDeposited)
				keyCounts = append(keyCounts, r.SigningKeyCount)
			}

			if res, err := tx.Exec(`
				INSERT INTO lido_node_operators (
					operator_id, active, name, reward_address, total_vetted_validators, total_exited_validators, total_added_validators, total_deposited_validators, signing_key_count
				)
				SELECT
					UNNEST($1::bigint[]),
					UNNEST($2::boolean[]),
					UNNEST($3::text[]),
					UNNEST($4::bytea[]),
					UNNEST($5::bigint[]),
					UNNEST($6::bigint[]),
					UNNEST($7::bigint[]),
					UNNEST($8::bigint[]),
					UNNEST($9::bigint[])
				ON CONFLICT (operator_id) DO UPDATE SET
					active = EXCLUDED.active,
					name = EXCLUDED.name,
					reward_address = EXCLUDED.reward_address,
					total_vetted_validators = EXCLUDED.total_vetted_validators,
					total_exited_validators = EXCLUDED.total_exited_validators,
					total_added_validators = EXCLUDED.total_added_validators,
					total_deposited_validators = EXCLUDED.total_deposited_validators,
					signing_key_count = EXCLUDED.signing_key_count
			`, pq.Array(ids), pq.Array(actives), pq.Array(names), pq.ByteaArray(rewards), pq.Array(vetted), pq.Array(exited), pq.Array(added), pq.Array(deposited), pq.Array(keyCounts)); err != nil {
				return fmt.Errorf("upsert lido_node_operators: %w", err)
			} else {
				if n, err2 := res.RowsAffected(); err2 == nil {
					metrics.Counter.WithLabelValues("validator_tagger_lido_operators_upserted").Add(float64(n))
				}
			}
		}
	} else {
		lidoLogger.Info("no operator property changes detected; skipping lido_node_operators upsert")
	}

	// second pass: after operators are persisted, fetch and insert signing keys per operator
	for _, r := range operatorRows {
		nodeOperatorIndex := r.OperatorID
		// Only fetch signing keys if the signing_key_count changed for this operator
		if !countsChanged[nodeOperatorIndex] {
			continue
		}
		nodeOperatorIndexBig := big.NewInt(nodeOperatorIndex)
		operatorSigningKeyCountInt64 := r.SigningKeyCount

		keysBatchSize := int64(1000)
		for i := int64(0); i < operatorSigningKeyCountInt64; i += keysBatchSize {
			offset := i
			limit := keysBatchSize
			if offset+limit > operatorSigningKeyCountInt64 {
				limit = operatorSigningKeyCountInt64 - offset
			}

			lidoLogger.Infof("retrieving signing keys for operator %d from %d to %d", nodeOperatorIndex, offset, limit)
			signingsKeys, err := operatorRegistry.GetSigningKeys(nil, nodeOperatorIndexBig, big.NewInt(offset), big.NewInt(limit))
			if err != nil {
				return fmt.Errorf("failed to retrieve signing keys for operator %d: %w", nodeOperatorIndex, err)
			}
			count := len(signingsKeys.Pubkeys) / 48
			lidoLogger.Infof("retrieved %d signing keys for operator %d", count, nodeOperatorIndex)
			if count == 0 {
				continue
			}

			// build arrays for UNNEST upsert of signing keys (operator_id, pubkey)
			operatorIDs := make([]int64, 0, count)
			pubkeys := make([][]byte, 0, count)
			for k := 0; k < count; k++ {
				start := k * 48
				end := start + 48
				pubkey := signingsKeys.Pubkeys[start:end]
				operatorIDs = append(operatorIDs, nodeOperatorIndex)
				pubkeys = append(pubkeys, pubkey)
			}

			if res, err := tx.Exec(`
				INSERT INTO lido_signing_keys (operator_id, pubkey)
				SELECT UNNEST($1::bigint[]), UNNEST($2::bytea[])
				ON CONFLICT (operator_id, pubkey) DO NOTHING
			`, pq.Array(operatorIDs), pq.ByteaArray(pubkeys)); err != nil {
				return fmt.Errorf("upsert lido_signing_keys (operator %d): %w", nodeOperatorIndex, err)
			} else {
				if n, err2 := res.RowsAffected(); err2 == nil {
					metrics.Counter.WithLabelValues("validator_tagger_lido_signing_keys_inserted").Add(float64(n))
				}
			}
		}
	}

	// Upsert validator_entities for 'Lido' using a single INSERT ... SELECT with joins.
	// This creates missing rows and updates existing ones to keep sub_entity in sync with operator name.
	lidoLogger.Info("upserting Lido validator_entities from lido_signing_keys and lido_node_operators")
	if res, err := tx.Exec(`
		INSERT INTO validator_entities (publickey, entity, sub_entity)
		SELECT lsk.pubkey, 'Lido', lno.name
		FROM lido_signing_keys lsk
		JOIN lido_node_operators lno ON lno.operator_id = lsk.operator_id
		ON CONFLICT (publickey) DO UPDATE
		SET entity = 'Lido',
		    sub_entity = EXCLUDED.sub_entity
	`); err != nil {
		return fmt.Errorf("upsert validator_entities (Lido): %w", err)
	} else {
		if n, err2 := res.RowsAffected(); err2 == nil {
			lidoLogger.WithField("affected", n).Info("upserted Lido validator_entities rows")
			metrics.Counter.WithLabelValues("validator_tagger_lido_validator_entities_upserted").Add(float64(n))
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func indexLidoCSMValidators() error {
	csmModule, err := lido.NewCSMModule(csmModuleAddress, rpc.CurrentErigonClient.GetNativeClient())
	if err != nil {
		return fmt.Errorf("failed to create csm module: %w", err)
	}

	nodeOperatorsCounter, err := csmModule.GetNodeOperatorsCount(nil)
	if err != nil {
		return fmt.Errorf("failed to retrieve CSM node operators count: %w", err)
	}
	nodeOperatorsCounterInt64 := nodeOperatorsCounter.Int64()
	lidoLogger.Infof("csm node operators count: %d", nodeOperatorsCounterInt64)

	// Collect operator metadata rows first
	type csmOperatorRow struct {
		OperatorID      int64 `db:"operator_id"`
		SigningKeyCount int64 `db:"signing_key_count"`
	}

	operatorRows := make([]csmOperatorRow, 0, nodeOperatorsCounterInt64)
	operatorIDs := make([]int64, 0, nodeOperatorsCounterInt64)

	// List operator IDs using GetNodeOperatorIds (IDs are not necessarily contiguous)
	idsBatch := int64(1000)
	for offset := int64(0); offset < nodeOperatorsCounterInt64; offset += idsBatch {
		limit := idsBatch
		if offset+limit > nodeOperatorsCounterInt64 {
			limit = nodeOperatorsCounterInt64 - offset
		}
		ids, err := csmModule.GetNodeOperatorIds(nil, big.NewInt(offset), big.NewInt(limit))
		if err != nil {
			return fmt.Errorf("GetNodeOperatorIds(offset=%d, limit=%d): %w", offset, limit, err)
		}
		for _, id := range ids {
			if id == nil {
				continue
			}
			operatorID := id.Int64()
			no, err := csmModule.GetNodeOperator(nil, id)
			if err != nil {
				return fmt.Errorf("GetNodeOperator(id=%d): %w", operatorID, err)
			}
			row := csmOperatorRow{
				OperatorID:      operatorID,
				SigningKeyCount: int64(no.TotalAddedKeys), // use TotalAddedKeys as key count
			}
			operatorRows = append(operatorRows, row)
			operatorIDs = append(operatorIDs, operatorID)
		}
	}

	// Open a single transaction for consistency
	tx, err := db.WriterDb.Beginx()
	if err != nil {
		return fmt.Errorf("begin tx (csm): %w", err)
	}
	defer tx.Rollback()

	// Load existing operator rows for diffing
	existingByID := make(map[int64]csmOperatorRow, len(operatorRows))
	if len(operatorIDs) > 0 {
		var existingRows []csmOperatorRow
		if err := tx.Select(&existingRows, `
			SELECT operator_id, signing_key_count
			FROM lido_csm_node_operators
			WHERE operator_id = ANY($1)
		`, pq.Array(operatorIDs)); err != nil {
			return fmt.Errorf("select existing lido_csm_node_operators: %w", err)
		}
		for _, e := range existingRows {
			existingByID[e.OperatorID] = e
		}
	}

	// Diff and collect rows to upsert and which key counts changed
	rowsToUpsert := make([]csmOperatorRow, 0, len(operatorRows))
	countsChanged := make(map[int64]bool, len(operatorRows))
	for _, r := range operatorRows {
		if e, ok := existingByID[r.OperatorID]; !ok {
			rowsToUpsert = append(rowsToUpsert, r)
			countsChanged[r.OperatorID] = true
			continue
		} else {
			same :=
				e.SigningKeyCount == r.SigningKeyCount
			if !same {
				rowsToUpsert = append(rowsToUpsert, r)
			}
			if e.SigningKeyCount != r.SigningKeyCount {
				countsChanged[r.OperatorID] = true
			}
		}
	}

	lidoLogger.Infof("lido CSM operators total: %d, to upsert: %d, signing-key-count changed: %d", len(operatorRows), len(rowsToUpsert), len(countsChanged))

	// Upsert only changed operator rows
	if len(rowsToUpsert) > 0 {
		batchSize := 1000
		for b := 0; b < len(rowsToUpsert); b += batchSize {
			start := b
			end := b + batchSize
			if end > len(rowsToUpsert) {
				end = len(rowsToUpsert)
			}

			ids := make([]int64, 0, end-start)
			keyCounts := make([]int64, 0, end-start)

			for _, r := range rowsToUpsert[start:end] {
				ids = append(ids, r.OperatorID)
				keyCounts = append(keyCounts, r.SigningKeyCount)
			}

			if res, err := tx.Exec(`
				INSERT INTO lido_csm_node_operators (
					operator_id, signing_key_count
				)
				SELECT
					UNNEST($1::bigint[]),
					UNNEST($2::bigint[])
				ON CONFLICT (operator_id) DO UPDATE SET
					signing_key_count = EXCLUDED.signing_key_count
			`, pq.Array(ids), pq.Array(keyCounts)); err != nil {
				return fmt.Errorf("upsert lido_csm_node_operators: %w", err)
			} else {
				if n, err2 := res.RowsAffected(); err2 == nil {
					metrics.Counter.WithLabelValues("validator_tagger_lido_csm_operators_upserted").Add(float64(n))
				}
			}
		}
	} else {
		lidoLogger.Info("no CSM operator property changes detected; skipping lido_csm_node_operators upsert")
	}

	// Fetch and upsert keys only for operators whose count changed
	for _, r := range operatorRows {
		operatorID := r.OperatorID
		if !countsChanged[operatorID] {
			continue
		}
		idBig := big.NewInt(operatorID)
		totalAdded := r.SigningKeyCount
		keysBatch := int64(1000)
		for start := int64(0); start < totalAdded; start += keysBatch {
			cnt := keysBatch
			if start+cnt > totalAdded {
				cnt = totalAdded - start
			}
			lidoLogger.WithFields(map[string]interface{}{`operator_id`: operatorID, `offset`: start, `limit`: cnt}).Info("retrieving CSM signing keys")
			keysBytes, err := csmModule.GetSigningKeys(nil, idBig, big.NewInt(start), big.NewInt(cnt))
			if err != nil {
				return fmt.Errorf("GetSigningKeys(id=%d, start=%d, count=%d): %w", operatorID, start, cnt, err)
			}
			if len(keysBytes) == 0 {
				continue
			}
			if len(keysBytes)%48 != 0 {
				return fmt.Errorf("unexpected CSM keysBytes length %d (not multiple of 48) for operator %d", len(keysBytes), operatorID)
			}
			count := len(keysBytes) / 48
			operatorIDs := make([]int64, 0, count)
			pubkeys := make([][]byte, 0, count)
			for k := 0; k < count; k++ {
				startPos := k * 48
				endPos := startPos + 48
				pubkeys = append(pubkeys, keysBytes[startPos:endPos])
				operatorIDs = append(operatorIDs, operatorID)
			}

			if _, err := tx.Exec(`
				INSERT INTO lido_csm_signing_keys (operator_id, pubkey)
				SELECT UNNEST($1::bigint[]), UNNEST($2::bytea[])
				ON CONFLICT (operator_id, pubkey) DO NOTHING
			`, pq.Array(operatorIDs), pq.ByteaArray(pubkeys)); err != nil {
				return fmt.Errorf("upsert lido_csm_signing_keys (operator %d): %w", operatorID, err)
			}
		}
	}

	// Upsert validator_entities for 'Lido CSM' from lido_csm_signing_keys and operators (join)
	lidoLogger.Info("upserting Lido CSM validator_entities from lido_csm_signing_keys and lido_csm_node_operators")
	if res, err := tx.Exec(`
		INSERT INTO validator_entities (publickey, entity, sub_entity)
		SELECT lsk.pubkey, 'Lido CSM', 'CSM Operator ' || lco.operator_id::text
		FROM lido_csm_signing_keys lsk
		JOIN lido_csm_node_operators lco ON lco.operator_id = lsk.operator_id
		ON CONFLICT (publickey) DO UPDATE
		SET entity = 'Lido CSM',
		    sub_entity = EXCLUDED.sub_entity
	`); err != nil {
		return fmt.Errorf("upsert validator_entities (Lido CSM): %w", err)
	} else {
		if n, err2 := res.RowsAffected(); err2 == nil {
			lidoLogger.WithField("affected", n).Info("upserted Lido CSM validator_entities rows")
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx (csm): %w", err)
	}
	return nil
}

func indexLidoSimpleDVTValidators() error {
	simpleDVTModule, err := lido.NewSimpleDVTModule(simpleDVTModuleAddress, rpc.CurrentErigonClient.GetNativeClient())
	if err != nil {
		return fmt.Errorf("failed to create simple_dvt module: %w", err)
	}

	nodeOperatorsCounter, err := simpleDVTModule.GetNodeOperatorsCount(nil)
	if err != nil {
		return fmt.Errorf("failed to retrieve Simple DVT node operators count: %w", err)
	}
	nodeOperatorsCounterInt64 := nodeOperatorsCounter.Int64()
	lidoLogger.Infof("simple dvt node operators count: %d", nodeOperatorsCounterInt64)

	// Collect operator metadata rows first
	type simpleDVTOperatorRow struct {
		OperatorID      int64  `db:"operator_id"`
		SigningKeyCount int64  `db:"signing_key_count"`
		Name            string `db:"name"`
	}

	operatorRows := make([]simpleDVTOperatorRow, 0, nodeOperatorsCounterInt64)
	operatorIDs := make([]int64, 0, nodeOperatorsCounterInt64)

	// List operator IDs using GetNodeOperatorIds (IDs are not necessarily contiguous)
	idsBatch := int64(1000)
	for offset := int64(0); offset < nodeOperatorsCounterInt64; offset += idsBatch {
		limit := idsBatch
		if offset+limit > nodeOperatorsCounterInt64 {
			limit = nodeOperatorsCounterInt64 - offset
		}
		ids, err := simpleDVTModule.GetNodeOperatorIds(nil, big.NewInt(offset), big.NewInt(limit))
		if err != nil {
			return fmt.Errorf("GetNodeOperatorIds(offset=%d, limit=%d): %w", offset, limit, err)
		}
		for _, id := range ids {
			if id == nil {
				continue
			}
			operatorID := id.Int64()
			no, err := simpleDVTModule.GetNodeOperator(nil, id, true)
			if err != nil {
				return fmt.Errorf("GetNodeOperator(id=%d): %w", operatorID, err)
			}
			row := simpleDVTOperatorRow{
				OperatorID:      operatorID,
				Name:            no.Name,
				SigningKeyCount: int64(no.TotalAddedValidators),
			}

			operatorRows = append(operatorRows, row)
			operatorIDs = append(operatorIDs, operatorID)
		}
	}

	// Open a single transaction for consistency
	tx, err := db.WriterDb.Beginx()
	if err != nil {
		return fmt.Errorf("begin tx (simple_dvt): %w", err)
	}
	defer tx.Rollback()

	// Load existing operator rows for diffing
	existingByID := make(map[int64]simpleDVTOperatorRow, len(operatorRows))
	if len(operatorIDs) > 0 {
		var existingRows []simpleDVTOperatorRow
		if err := tx.Select(&existingRows, `
			SELECT operator_id, signing_key_count, name
			FROM lido_simple_dvt_node_operators
			WHERE operator_id = ANY($1)
		`, pq.Array(operatorIDs)); err != nil {
			return fmt.Errorf("select existing lido_simple_dvt_node_operators: %w", err)
		}
		for _, e := range existingRows {
			existingByID[e.OperatorID] = e
		}
	}

	// Diff and collect rows to upsert and which key counts changed
	rowsToUpsert := make([]simpleDVTOperatorRow, 0, len(operatorRows))
	countsChanged := make(map[int64]bool, len(operatorRows))
	for _, r := range operatorRows {
		if e, ok := existingByID[r.OperatorID]; !ok {
			rowsToUpsert = append(rowsToUpsert, r)
			countsChanged[r.OperatorID] = true
			continue
		} else {
			same :=
				e.SigningKeyCount == r.SigningKeyCount
			if !same {
				rowsToUpsert = append(rowsToUpsert, r)
			}
			if e.SigningKeyCount != r.SigningKeyCount {
				countsChanged[r.OperatorID] = true
			}
		}
	}

	lidoLogger.Infof("lido Simple DVT operators total: %d, to upsert: %d, signing-key-count changed: %d", len(operatorRows), len(rowsToUpsert), len(countsChanged))

	// Upsert only changed operator rows
	if len(rowsToUpsert) > 0 {
		batchSize := 1000
		for b := 0; b < len(rowsToUpsert); b += batchSize {
			start := b
			end := b + batchSize
			if end > len(rowsToUpsert) {
				end = len(rowsToUpsert)
			}

			ids := make([]int64, 0, end-start)
			keyCounts := make([]int64, 0, end-start)
			names := make([]string, 0, end-start)

			for _, r := range rowsToUpsert[start:end] {
				ids = append(ids, r.OperatorID)
				keyCounts = append(keyCounts, r.SigningKeyCount)
				names = append(names, r.Name)
			}

			if res, err := tx.Exec(`
				INSERT INTO lido_simple_dvt_node_operators (
					operator_id, signing_key_count, name
				)
				SELECT
					UNNEST($1::bigint[]),
					UNNEST($2::bigint[]),
					UNNEST($3::text[])
				ON CONFLICT (operator_id) DO UPDATE SET
					signing_key_count = EXCLUDED.signing_key_count,
					name = EXCLUDED.name
			`, pq.Array(ids), pq.Array(keyCounts), pq.Array(names)); err != nil {
				return fmt.Errorf("upsert lido_simple_dvt_node_operators: %w", err)
			} else {
				if n, err2 := res.RowsAffected(); err2 == nil {
					metrics.Counter.WithLabelValues("validator_tagger_lido_simple_dvt_operators_upserted").Add(float64(n))
				}
			}
		}
	} else {
		lidoLogger.Info("no Simple DVT operator property changes detected; skipping lido_simple_dvt_node_operators upsert")
	}

	// Fetch and upsert keys only for operators whose count changed
	for _, r := range operatorRows {
		operatorID := r.OperatorID
		if !countsChanged[operatorID] {
			continue
		}
		idBig := big.NewInt(operatorID)
		totalAdded := r.SigningKeyCount
		keysBatch := int64(1000)
		for start := int64(0); start < totalAdded; start += keysBatch {
			cnt := keysBatch
			if start+cnt > totalAdded {
				cnt = totalAdded - start
			}
			lidoLogger.WithFields(map[string]interface{}{`operator_id`: operatorID, `offset`: start, `limit`: cnt}).Info("retrieving Simple DVT signing keys")
			keysBytes, err := simpleDVTModule.GetSigningKeys(nil, idBig, big.NewInt(start), big.NewInt(cnt))
			if err != nil {
				return fmt.Errorf("GetSigningKeys(id=%d, start=%d, count=%d): %w", operatorID, start, cnt, err)
			}
			if len(keysBytes.Pubkeys) == 0 {
				continue
			}
			if len(keysBytes.Pubkeys)%48 != 0 {
				return fmt.Errorf("unexpected Simple DVT keysBytes length %d (not multiple of 48) for operator %d", len(keysBytes.Pubkeys), operatorID)
			}
			count := len(keysBytes.Pubkeys) / 48
			operatorIDs := make([]int64, 0, count)
			pubkeys := make([][]byte, 0, count)
			for k := 0; k < count; k++ {
				startPos := k * 48
				endPos := startPos + 48
				pubkeys = append(pubkeys, keysBytes.Pubkeys[startPos:endPos])
				operatorIDs = append(operatorIDs, operatorID)
			}

			if _, err := tx.Exec(`
				INSERT INTO lido_simple_dvt_signing_keys (operator_id, pubkey)
				SELECT UNNEST($1::bigint[]), UNNEST($2::bytea[])
				ON CONFLICT (operator_id, pubkey) DO NOTHING
			`, pq.Array(operatorIDs), pq.ByteaArray(pubkeys)); err != nil {
				return fmt.Errorf("upsert lido_simple_dvt_signing_keys (operator %d): %w", operatorID, err)
			}
		}
	}

	// Upsert validator_entities for 'Lido Simple DVT' from lido_simple_dvt_signing_keys and operators (join)
	lidoLogger.Info("upserting Lido simple DVT validator_entities from lido_simple_dvt_signing_keys and lido_simple_dvt_node_operators")
	if res, err := tx.Exec(`
		INSERT INTO validator_entities (publickey, entity, sub_entity)
		SELECT lsk.pubkey, 'Lido Simple DVT', lco.name
		FROM lido_simple_dvt_signing_keys lsk
		JOIN lido_simple_dvt_node_operators lco ON lco.operator_id = lsk.operator_id
		ON CONFLICT (publickey) DO UPDATE
		SET entity = 'Lido Simple DVT',
		    sub_entity = EXCLUDED.sub_entity
	`); err != nil {
		return fmt.Errorf("upsert validator_entities (Lido Simple DVT): %w", err)
	} else {
		if n, err2 := res.RowsAffected(); err2 == nil {
			lidoLogger.WithField("affected", n).Info("upserted Lido Simple DVT validator_entities rows")
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx (simple_dvt): %w", err)
	}
	return nil
}
