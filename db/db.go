package db

import (
	"bytes"
	"database/sql"
	"eth2-exporter/metrics"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"math/big"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"github.com/jackc/pgx/v4/pgxpool"
)

var DBPGX *pgxpool.Conn

// DB is a pointer to the explorer-database
var WriterDb *sqlx.DB
var ReaderDb *sqlx.DB

var logger = logrus.StandardLogger().WithField("module", "db")

func mustInitDB(writer *types.DatabaseConfig, reader *types.DatabaseConfig) (*sqlx.DB, *sqlx.DB) {
	dbConnWriter, err := sqlx.Open("pgx", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", writer.Username, writer.Password, writer.Host, writer.Port, writer.Name))
	if err != nil {
		logger.Fatal(err)
	}

	// The golang sql driver does not properly implement PingContext
	// therefore we use a timer to catch db connection timeouts
	dbConnectionTimeout := time.NewTimer(15 * time.Second)
	go func() {
		<-dbConnectionTimeout.C
		logger.Fatalf("timeout while connecting to the database")
	}()
	err = dbConnWriter.Ping()
	if err != nil {
		logger.Fatal(err)
	}
	dbConnectionTimeout.Stop()

	dbConnWriter.SetConnMaxIdleTime(time.Second * 30)
	dbConnWriter.SetConnMaxLifetime(time.Second * 60)
	dbConnWriter.SetMaxOpenConns(200)
	dbConnWriter.SetMaxIdleConns(200)

	if reader == nil {

		return dbConnWriter, dbConnWriter
	}

	dbConnReader, err := sqlx.Open("pgx", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", reader.Username, reader.Password, reader.Host, reader.Port, reader.Name))
	if err != nil {
		logger.Fatal(err)
	}

	// The golang sql driver does not properly implement PingContext
	// therefore we use a timer to catch db connection timeouts
	dbConnectionTimeout = time.NewTimer(15 * time.Second)
	go func() {
		<-dbConnectionTimeout.C
		logger.Fatalf("timeout while connecting to the read replica database")
	}()
	err = dbConnReader.Ping()
	if err != nil {
		logger.Fatal(err)
	}
	dbConnectionTimeout.Stop()

	dbConnReader.SetConnMaxIdleTime(time.Second * 30)
	dbConnReader.SetConnMaxLifetime(time.Second * 60)
	dbConnReader.SetMaxOpenConns(200)
	dbConnReader.SetMaxIdleConns(200)
	return dbConnWriter, dbConnReader
}

func MustInitDB(writer *types.DatabaseConfig, reader *types.DatabaseConfig) {
	WriterDb, ReaderDb = mustInitDB(writer, reader)
}

func GetEth1Deposits(address string, length, start uint64) ([]*types.EthOneDepositsData, error) {
	deposits := []*types.EthOneDepositsData{}

	err := ReaderDb.Select(&deposits, `
	SELECT 
		tx_hash,
		tx_input,
		tx_index,
		block_number,
		block_ts as block_ts,
		from_address,
		publickey,
		withdrawal_credentials,
		amount,
		signature,
		merkletree_index
	FROM 
		eth1_deposits
	ORDER BY block_ts DESC
	LIMIT $1
	OFFSET $2`, length, start)
	if err != nil {
		return nil, err
	}

	return deposits, nil
}

var searchLikeHash = regexp.MustCompile(`^0?x?[0-9a-fA-F]{2,96}`) // only search for pubkeys if string consists of 96 hex-chars

func GetEth1DepositsJoinEth2Deposits(query string, length, start uint64, orderBy, orderDir string, latestEpoch, validatorOnlineThresholdSlot uint64) ([]*types.EthOneDepositsData, uint64, error) {
	deposits := []*types.EthOneDepositsData{}

	if orderDir != "desc" && orderDir != "asc" {
		orderDir = "desc"
	}
	columns := []string{"tx_hash", "tx_input", "tx_index", "block_number", "block_ts", "from_address", "publickey", "withdrawal_credentials", "amount", "signature", "merkletree_index", "state", "valid_signature"}
	hasColumn := false
	for _, column := range columns {
		if orderBy == column {
			hasColumn = true
			break
		}
	}
	if !hasColumn {
		orderBy = "block_ts"
	}

	var totalCount uint64
	var err error

	query = strings.Replace(query, "0x", "", -1)

	if searchLikeHash.MatchString(query) {
		if query != "" {
			err = ReaderDb.Get(&totalCount, `
				SELECT COUNT(*) FROM eth1_deposits as eth1
				WHERE 
					ENCODE(eth1.publickey, 'hex') LIKE LOWER($1)
					OR ENCODE(eth1.withdrawal_credentials, 'hex') LIKE LOWER($1)
					OR ENCODE(eth1.from_address, 'hex') LIKE LOWER($1)
					OR ENCODE(tx_hash, 'hex') LIKE LOWER($1)
					OR CAST(eth1.block_number AS text) LIKE LOWER($1)`, query+"%")
		}
	} else {
		if query != "" {
			err = ReaderDb.Get(&totalCount, `
				SELECT COUNT(*) FROM eth1_deposits as eth1
				WHERE 
				CAST(eth1.block_number AS text) LIKE LOWER($1)`, query+"%")
		}
	}

	if query == "" {
		err = ReaderDb.Get(&totalCount, "SELECT COUNT(*) FROM eth1_deposits")
	}

	if err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	}

	if query != "" {
		wholeQuery := fmt.Sprintf(`
		SELECT 
			eth1.tx_hash as tx_hash,
			eth1.tx_input as tx_input,
			eth1.tx_index as tx_index,
			eth1.block_number as block_number,
			eth1.block_ts as block_ts,
			eth1.from_address as from_address,
			eth1.publickey as publickey,
			eth1.withdrawal_credentials as withdrawal_credentials,
			eth1.amount as amount,
			eth1.signature as signature,
			eth1.merkletree_index as merkletree_index,
			eth1.valid_signature as valid_signature,
			COALESCE(v.state, 'deposited') as state
		FROM
			eth1_deposits as eth1
		LEFT JOIN
			(
				SELECT pubkey,
				CASE 
					WHEN exitepoch <= $3 then 'exited'
					WHEN activationepoch > $3 then 'pending'
					WHEN slashed and activationepoch < $3 and (lastattestationslot < $4 OR lastattestationslot is null) then 'slashing_offline'
					WHEN slashed then 'slashing_online'
					WHEN activationepoch < $3 and (lastattestationslot < $4 OR lastattestationslot is null) then 'active_offline'
					ELSE 'active_online'
				END AS state
				FROM validators
			) as v
		ON
			v.pubkey = eth1.publickey
		WHERE
			ENCODE(eth1.publickey, 'hex') LIKE LOWER($5)
			OR ENCODE(eth1.withdrawal_credentials, 'hex') LIKE LOWER($5)
			OR ENCODE(eth1.from_address, 'hex') LIKE LOWER($5)
			OR ENCODE(tx_hash, 'hex') LIKE LOWER($5)
			OR CAST(eth1.block_number AS text) LIKE LOWER($5)
		ORDER BY %s %s
		LIMIT $1
		OFFSET $2`, orderBy, orderDir)
		err = ReaderDb.Select(&deposits, wholeQuery, length, start, latestEpoch, validatorOnlineThresholdSlot, query+"%")
	} else {
		err = ReaderDb.Select(&deposits, fmt.Sprintf(`
		SELECT 
			eth1.tx_hash as tx_hash,
			eth1.tx_input as tx_input,
			eth1.tx_index as tx_index,
			eth1.block_number as block_number,
			eth1.block_ts as block_ts,
			eth1.from_address as from_address,
			eth1.publickey as publickey,
			eth1.withdrawal_credentials as withdrawal_credentials,
			eth1.amount as amount,
			eth1.signature as signature,
			eth1.merkletree_index as merkletree_index,
			eth1.valid_signature as valid_signature,
			COALESCE(v.state, 'deposited') as state
		FROM
			eth1_deposits as eth1
			LEFT JOIN
			(
				SELECT pubkey,
				CASE 
					WHEN exitepoch <= $3 then 'exited'
					WHEN activationepoch > $3 then 'pending'
					WHEN slashed and activationepoch < $3 and (lastattestationslot < $4 OR lastattestationslot is null) then 'slashing_offline'
					WHEN slashed then 'slashing_online'
					WHEN activationepoch < $3 and (lastattestationslot < $4 OR lastattestationslot is null) then 'active_offline'
					ELSE 'active_online'
				END AS state
				FROM validators
			) as v
		ON
			v.pubkey = eth1.publickey
		ORDER BY %s %s
		LIMIT $1
		OFFSET $2`, orderBy, orderDir), length, start, latestEpoch, validatorOnlineThresholdSlot)
	}
	if err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	}

	return deposits, totalCount, nil
}

func GetEth1DepositsCount() (uint64, error) {
	deposits := uint64(0)
	err := ReaderDb.Get(&deposits, `SELECT COUNT(*) FROM eth1_deposits`)
	if err != nil {
		return 0, err
	}
	return deposits, nil
}

func GetEth1DepositsLeaderboard(query string, length, start uint64, orderBy, orderDir string, latestEpoch uint64) ([]*types.EthOneDepositLeaderboardData, uint64, error) {
	deposits := []*types.EthOneDepositLeaderboardData{}

	if orderDir != "desc" && orderDir != "asc" {
		orderDir = "desc"
	}
	columns := []string{
		"from_address",
		"amount",
		"validcount",
		"invalidcount",
		"slashedcount",
		"totalcount",
		"activecount",
		"pendingcount",
		"voluntary_exit_count",
	}
	hasColumn := false
	for _, column := range columns {
		if orderBy == column {
			hasColumn = true
		}
	}
	if !hasColumn {
		orderBy = "amount"
	}

	var err error
	var totalCount uint64
	if query != "" {
		err = ReaderDb.Get(&totalCount, `
		SELECT
			COUNT(from_address)
			FROM
				(
					SELECT
						from_address
					FROM
						eth1_deposits as eth1
					WHERE
					ENCODE(eth1.from_address, 'hex') LOWER($1)
						GROUP BY from_address
				) as count
		`, query+"%")
	} else {
		err = ReaderDb.Get(&totalCount, "SELECT COUNT(*) FROM (SELECT from_address FROM eth1_deposits GROUP BY from_address) as count")
	}
	if err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	}

	err = ReaderDb.Select(&deposits, fmt.Sprintf(`
		SELECT
			eth1.from_address,
			SUM(eth1.amount) as amount,
			SUM(eth1.validcount) AS validcount,
			SUM(eth1.invalidcount) AS invalidcount,
			COUNT(CASE WHEN v.slashed = 't' THEN 1 END) AS slashedcount,
			COUNT(v.pubkey) AS totalcount,
			COUNT(CASE WHEN v.slashed = 'f' AND v.exitepoch > $3 AND v.activationepoch < $3 THEN 1 END) as activecount,
			COUNT(CASE WHEN v.activationepoch > $3 THEN 1 END) AS pendingcount,
			COUNT(CASE WHEN v.slashed = 'f' AND v.exitepoch < $3 THEN 1 END) AS voluntary_exit_count
		FROM (
			SELECT 
				from_address,
				publickey,
				SUM(amount) AS amount,
				COUNT(CASE WHEN valid_signature = 't' THEN 1 END) AS validcount,
				COUNT(CASE WHEN valid_signature = 'f' THEN 1 END) AS invalidcount
			FROM eth1_deposits
			GROUP BY from_address, publickey
		) eth1
		LEFT JOIN (
			SELECT 
				pubkey,
				slashed,
				exitepoch,
				activationepoch,
				COALESCE(validator_names.name, '') AS name
			FROM validators
			LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
		) v ON v.pubkey = eth1.publickey
		WHERE ENCODE(eth1.from_address, 'hex') LIKE LOWER($4)
		GROUP BY eth1.from_address
		ORDER BY %s %s
		LIMIT $1
		OFFSET $2`, orderBy, orderDir), length, start, latestEpoch, query+"%")
	if err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	}
	return deposits, totalCount, nil
}

func GetEth2Deposits(query string, length, start uint64, orderBy, orderDir string) ([]*types.EthTwoDepositData, error) {
	deposits := []*types.EthTwoDepositData{}
	// ENCODE(publickey, 'hex') LIKE $3 OR ENCODE(withdrawalcredentials, 'hex') LIKE $3 OR
	if orderDir != "desc" && orderDir != "asc" {
		orderDir = "desc"
	}
	columns := []string{"block_slot", "publickey", "amount", "withdrawalcredentials", "signature"}
	hasColumn := false
	for _, column := range columns {
		if orderBy == column {
			hasColumn = true
		}
	}
	if !hasColumn {
		orderBy = "block_slot"
	}

	if query != "" {
		err := ReaderDb.Select(&deposits, fmt.Sprintf(`
			SELECT 
				blocks_deposits.block_slot,
				blocks_deposits.block_index,
				blocks_deposits.proof,
				blocks_deposits.publickey,
				blocks_deposits.withdrawalcredentials,
				blocks_deposits.amount,
				blocks_deposits.signature
			FROM blocks_deposits
			INNER JOIN blocks ON blocks_deposits.block_root = blocks.blockroot AND blocks.status = '1'
			WHERE ENCODE(publickey, 'hex') LIKE LOWER($3)
				OR ENCODE(withdrawalcredentials, 'hex') LIKE LOWER($3)
				OR CAST(block_slot as varchar) LIKE LOWER($3)
			ORDER BY %s %s
			LIMIT $1
			OFFSET $2`, orderBy, orderDir), length, start, query+"%")
		if err != nil {
			return nil, err
		}
	} else {
		err := ReaderDb.Select(&deposits, fmt.Sprintf(`
			SELECT 
				blocks_deposits.block_slot,
				blocks_deposits.block_index,
				blocks_deposits.proof,
				blocks_deposits.publickey,
				blocks_deposits.withdrawalcredentials,
				blocks_deposits.amount,
				blocks_deposits.signature
			FROM blocks_deposits
			INNER JOIN blocks ON blocks_deposits.block_root = blocks.blockroot AND blocks.status = '1'
			ORDER BY %s %s
			LIMIT $1
			OFFSET $2`, orderBy, orderDir), length, start)
		if err != nil {
			return nil, err
		}
	}

	return deposits, nil
}

func GetEth2DepositsCount(search string) (uint64, error) {
	deposits := uint64(0)
	var err error
	if search == "" {
		err = ReaderDb.Get(&deposits, `
		SELECT COUNT(*)
		FROM blocks_deposits
		INNER JOIN blocks ON blocks_deposits.block_root = blocks.blockroot AND blocks.status = '1'`)
	} else {
		err = ReaderDb.Get(&deposits, `
		SELECT COUNT(*)
		FROM blocks_deposits
		INNER JOIN blocks ON blocks_deposits.block_root = blocks.blockroot AND blocks.status = '1'
		WHERE 
			ENCODE(publickey, 'hex') LIKE LOWER($1)
			OR ENCODE(withdrawalcredentials, 'hex') LIKE LOWER($1)
			OR CAST(block_slot as varchar) LIKE LOWER($1)
		`, search+"%")
	}
	if err != nil {
		return 0, err
	}

	return deposits, nil
}
func GetSlashingCount() (uint64, error) {
	slashings := uint64(0)

	err := ReaderDb.Get(&slashings, `
		SELECT SUM(count)
		FROM 
		(
			SELECT COUNT(*) 
			FROM 
				blocks_attesterslashings 
				INNER JOIN blocks on blocks.slot = blocks_attesterslashings.block_slot and blocks.status = '1'
			UNION 
			SELECT COUNT(*) 
			FROM 
				blocks_proposerslashings
				INNER JOIN blocks on blocks.slot = blocks_proposerslashings.block_slot and blocks.status = '1'
		) as tbl`)
	if err != nil {
		return 0, err
	}

	return slashings, nil
}

// GetLatestEpoch will return the latest epoch from the database
func GetLatestEpoch() (uint64, error) {
	var epoch uint64
	err := WriterDb.Get(&epoch, "SELECT COALESCE(MAX(epoch), 0) FROM epochs")

	if err != nil {
		return 0, fmt.Errorf("error retrieving latest epoch from DB: %v", err)
	}

	return epoch, nil
}

// GetAllEpochs will return a collection of all of the epochs from the database
func GetAllEpochs() ([]uint64, error) {
	var epochs []uint64
	err := WriterDb.Select(&epochs, "SELECT epoch FROM epochs ORDER BY epoch")

	if err != nil {
		return nil, fmt.Errorf("error retrieving all epochs from DB: %v", err)
	}

	return epochs, nil
}

// GetLastPendingAndProposedBlocks will return all proposed and pending blocks (ignores missed slots) from the database
func GetLastPendingAndProposedBlocks(startEpoch, endEpoch uint64) ([]*types.MinimalBlock, error) {
	var blocks []*types.MinimalBlock

	err := WriterDb.Select(&blocks, "SELECT epoch, slot, blockroot FROM blocks WHERE epoch >= $1 AND epoch <= $2 AND blockroot != '\x01' ORDER BY slot DESC", startEpoch, endEpoch)

	if err != nil {
		return nil, fmt.Errorf("error retrieving last blocks (%v-%v) from DB: %v", startEpoch, endEpoch, err)
	}

	return blocks, nil
}

// GetBlocks will return all blocks for a range of epochs from the database
func GetBlocks(startEpoch, endEpoch uint64) ([]*types.MinimalBlock, error) {
	var blocks []*types.MinimalBlock

	err := ReaderDb.Select(&blocks, "SELECT epoch, slot, blockroot, parentroot FROM blocks WHERE epoch >= $1 AND epoch <= $2 AND length(blockroot) = 32 ORDER BY slot DESC", startEpoch, endEpoch)

	if err != nil {
		return nil, fmt.Errorf("error retrieving blocks for epoch %v to %v from DB: %v", startEpoch, endEpoch, err)
	}

	return blocks, nil
}

// GetValidatorPublicKey will return the public key for a specific validator from the database
func GetValidatorPublicKey(index uint64) ([]byte, error) {
	var publicKey []byte
	err := ReaderDb.Get(&publicKey, "SELECT pubkey FROM validators WHERE validatorindex = $1", index)

	return publicKey, err
}

// GetValidatorIndex will return the validator-index for a public key from the database
func GetValidatorIndex(publicKey []byte) (uint64, error) {
	var index uint64
	err := ReaderDb.Get(&index, "SELECT validatorindex FROM validators WHERE pubkey = $1", publicKey)

	return index, err
}

// GetValidatorDeposits will return eth1- and eth2-deposits for a public key from the database
func GetValidatorDeposits(publicKey []byte) (*types.ValidatorDeposits, error) {
	deposits := &types.ValidatorDeposits{}
	err := ReaderDb.Select(&deposits.Eth1Deposits, `
		SELECT tx_hash, tx_input, tx_index, block_number, EXTRACT(epoch FROM block_ts)::INT as block_ts, from_address, publickey, withdrawal_credentials, amount, signature, merkletree_index, valid_signature
		FROM eth1_deposits WHERE publickey = $1 ORDER BY block_number ASC`, publicKey)
	if err != nil {
		return nil, err
	}
	if len(deposits.Eth1Deposits) > 0 {
		deposits.LastEth1DepositTs = deposits.Eth1Deposits[len(deposits.Eth1Deposits)-1].BlockTs
	}

	err = ReaderDb.Select(&deposits.Eth2Deposits, `
		SELECT 
			blocks_deposits.block_slot,
			blocks_deposits.block_index,
			blocks_deposits.block_root,
			blocks_deposits.proof,
			blocks_deposits.publickey,
			blocks_deposits.withdrawalcredentials,
			blocks_deposits.amount,
			blocks_deposits.signature
		FROM blocks_deposits
		INNER JOIN blocks ON (blocks_deposits.block_root = blocks.blockroot AND blocks.status = '1') OR (blocks_deposits.block_slot = 0 AND blocks_deposits.block_slot = blocks.slot AND blocks_deposits.publickey = $1)
		WHERE blocks_deposits.publickey = $1`, publicKey)
	if err != nil {
		return nil, err
	}
	return deposits, nil
}

// UpdateCanonicalBlocks will update the blocks for an epoch range in the database
func UpdateCanonicalBlocks(startEpoch, endEpoch uint64, blocks []*types.MinimalBlock) error {
	if len(blocks) == 0 {
		return nil
	}
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_canonical_blocks").Observe(time.Since(start).Seconds())
	}()

	tx, err := WriterDb.Begin()
	if err != nil {
		return fmt.Errorf("error starting db transactions: %v", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE blocks SET status = 1 WHERE epoch >= $1 AND epoch <= $2 AND (status = '1' OR status = '3')", startEpoch, endEpoch)
	if err != nil {
		return err
	}

	for _, block := range blocks {
		if block.Canonical {
			continue
		}
		logger.Printf("marking block %x at slot %v as orphaned", block.BlockRoot, block.Slot)
		_, err = tx.Exec("UPDATE blocks SET status = '3' WHERE blockroot = $1", block.BlockRoot)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func SetBlockStatus(blocks []*types.CanonBlock) error {
	if len(blocks) == 0 {
		return nil
	}

	tx, err := WriterDb.Begin()
	if err != nil {
		return fmt.Errorf("error starting db transactions: %v", err)
	}
	defer tx.Rollback()

	canonBlocks := make(pq.ByteaArray, 0)
	orphanedBlocks := make(pq.ByteaArray, 0)
	for _, block := range blocks {
		if !block.Canonical {
			logger.Printf("marking block %x at slot %v as orphaned", block.BlockRoot, block.Slot)
			orphanedBlocks = append(orphanedBlocks, block.BlockRoot)
		} else {
			logger.Printf("marking block %x at slot %v as canonical", block.BlockRoot, block.Slot)
			canonBlocks = append(canonBlocks, block.BlockRoot)
		}

	}

	_, err = tx.Exec("UPDATE blocks SET status = '1' WHERE blockroot = ANY($1)", canonBlocks)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE blocks SET status = '3' WHERE blockroot = ANY($1)", orphanedBlocks)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// SaveValidatorQueue will save the validator queue into the database
func SaveValidatorQueue(validators *types.ValidatorQueue) error {
	_, err := WriterDb.Exec(`
		INSERT INTO queue (ts, entering_validators_count, exiting_validators_count)
		VALUES (date_trunc('hour', now()), $1, $2)
		ON CONFLICT (ts) DO UPDATE SET
			entering_validators_count = excluded.entering_validators_count, 
			exiting_validators_count = excluded.exiting_validators_count`,
		validators.Activating, validators.Exititing)
	return err
}

func SaveBlock(block *types.Block) error {

	blocksMap := make(map[uint64]map[string]*types.Block)
	if blocksMap[block.Slot] == nil {
		blocksMap[block.Slot] = make(map[string]*types.Block)
	}
	blocksMap[block.Slot][fmt.Sprintf("%x", block.BlockRoot)] = block

	tx, err := WriterDb.Begin()
	if err != nil {
		return fmt.Errorf("error starting db transactions: %v", err)
	}
	defer tx.Rollback()

	logger.Infof("exporting block data")
	err = saveBlocks(blocksMap, tx)
	if err != nil {
		logger.Fatalf("error saving blocks to db: %v", err)
		return fmt.Errorf("error saving blocks to db: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing db transaction: %v", err)
	}

	return nil
}

// SaveEpoch will stave the epoch data into the database
func SaveEpoch(data *types.EpochData) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_save_epoch").Observe(time.Since(start).Seconds())
		logger.WithFields(logrus.Fields{"epoch": data.Epoch, "duration": time.Since(start)}).Info("completed saving epoch")
	}()

	tx, err := WriterDb.Begin()
	if err != nil {
		return fmt.Errorf("error starting db transactions: %w", err)
	}
	defer tx.Rollback()

	logger.WithFields(logrus.Fields{"chainEpoch": utils.TimeToEpoch(time.Now()), "exportEpoch": data.Epoch}).Infof("starting export of epoch %v", data.Epoch)

	logger.Infof("exporting block data")
	err = saveBlocks(data.Blocks, tx)
	if err != nil {
		logger.Fatalf("error saving blocks to db: %v", err)
		return fmt.Errorf("error saving blocks to db: %w", err)
	}

	if uint64(utils.TimeToEpoch(time.Now())) > data.Epoch+10 {
		logger.WithFields(logrus.Fields{"exportEpoch": data.Epoch, "chainEpoch": utils.TimeToEpoch(time.Now())}).Infof("skipping exporting validators because epoch is far behind head")
	} else {
		logger.Infof("exporting validators")
		err = saveValidators(data, tx)
		if err != nil {
			return fmt.Errorf("error saving validators to db: %w", err)
		}
	}

	logger.Infof("exporting proposal assignments data")
	err = saveValidatorProposalAssignments(data.Epoch, data.ValidatorAssignmentes.ProposerAssignments, tx)
	if err != nil {
		return fmt.Errorf("error saving validator proposal assignments to db: %w", err)
	}

	logger.Infof("exporting attestation assignments data")
	err = saveValidatorAttestationAssignments(data.Epoch, data.ValidatorAssignmentes.AttestorAssignments, tx)
	if err != nil {
		return fmt.Errorf("error saving validator attestation assignments to db: %w", err)
	}

	logger.Infof("exporting validator balance data")
	err = saveValidatorBalances(data.Epoch, data.Validators, tx)
	if err != nil {
		return fmt.Errorf("error saving validator balances to db: %w", err)
	}

	// only export recent validator balances if the epoch is within the threshold
	if uint64(utils.TimeToEpoch(time.Now())) > data.Epoch+10 {
		logger.WithFields(logrus.Fields{"exportEpoch": data.Epoch, "chainEpoch": utils.TimeToEpoch(time.Now())}).Infof("skipping exporting recent validator balance because epoch is far behind head")
	} else {
		logger.Infof("exporting recent validator balance")
		err = saveValidatorBalancesRecent(data.Epoch, data.Validators, tx)
		if err != nil {
			return fmt.Errorf("error saving recent validator balances to db: %w", err)
		}
	}

	logger.Infof("exporting epoch statistics data")
	proposerSlashingsCount := 0
	attesterSlashingsCount := 0
	attestationsCount := 0
	depositCount := 0
	voluntaryExitCount := 0

	for _, slot := range data.Blocks {
		for _, b := range slot {
			proposerSlashingsCount += len(b.ProposerSlashings)
			attesterSlashingsCount += len(b.AttesterSlashings)
			attestationsCount += len(b.Attestations)
			depositCount += len(b.Deposits)
			voluntaryExitCount += len(b.VoluntaryExits)
		}
	}

	validatorBalanceSum := new(big.Int)
	validatorsCount := 0
	for _, v := range data.Validators {
		if v.ExitEpoch > data.Epoch && v.ActivationEpoch <= data.Epoch {
			validatorsCount++
			validatorBalanceSum = new(big.Int).Add(validatorBalanceSum, new(big.Int).SetUint64(v.Balance))
		}
	}

	validatorBalanceAverage := new(big.Int).Div(validatorBalanceSum, new(big.Int).SetInt64(int64(validatorsCount)))

	_, err = tx.Exec(`
		INSERT INTO epochs (
			epoch, 
			blockscount, 
			proposerslashingscount, 
			attesterslashingscount, 
			attestationscount, 
			depositscount, 
			voluntaryexitscount, 
			validatorscount, 
			averagevalidatorbalance, 
			totalvalidatorbalance,
			finalized, 
			eligibleether, 
			globalparticipationrate, 
			votedether
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) 
		ON CONFLICT (epoch) DO UPDATE SET 
			blockscount             = excluded.blockscount, 
			proposerslashingscount  = excluded.proposerslashingscount,
			attesterslashingscount  = excluded.attesterslashingscount,
			attestationscount       = excluded.attestationscount,
			depositscount           = excluded.depositscount,
			voluntaryexitscount     = excluded.voluntaryexitscount,
			validatorscount         = excluded.validatorscount,
			averagevalidatorbalance = excluded.averagevalidatorbalance,
			totalvalidatorbalance   = excluded.totalvalidatorbalance,
			finalized               = excluded.finalized,
			eligibleether           = excluded.eligibleether,
			globalparticipationrate = excluded.globalparticipationrate,
			votedether              = excluded.votedether`,
		data.Epoch,
		len(data.Blocks),
		proposerSlashingsCount,
		attesterSlashingsCount,
		attestationsCount,
		depositCount,
		voluntaryExitCount,
		validatorsCount,
		validatorBalanceAverage.Uint64(),
		validatorBalanceSum.Uint64(),
		data.EpochParticipationStats.Finalized,
		data.EpochParticipationStats.EligibleEther,
		data.EpochParticipationStats.GlobalParticipationRate,
		data.EpochParticipationStats.VotedEther)

	if err != nil {
		return fmt.Errorf("error executing save epoch statement: %w", err)
	}

	err = saveGraffitiwall(data.Blocks, tx)
	if err != nil {
		return fmt.Errorf("error saving graffitiwall: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing db transaction: %w", err)
	}

	logger.Infof("export of epoch %v completed, took %v", data.Epoch, time.Since(start))
	return nil
}

func saveGraffitiwall(blocks map[uint64]map[string]*types.Block, tx *sql.Tx) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_save_graffitiwall").Observe(time.Since(start).Seconds())
	}()

	stmtGraffitiwall, err := tx.Prepare(`
		INSERT INTO graffitiwall (
			x,
			y,
			color,
			slot,
			validator
		)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (x, y) DO UPDATE SET
										 color         = EXCLUDED.color,
										 slot          = EXCLUDED.slot,
										 validator     = EXCLUDED.validator
		WHERE excluded.slot > graffitiwall.slot;
		`)
	if err != nil {
		return err
	}
	defer stmtGraffitiwall.Close()

	regexes := [...]*regexp.Regexp{
		regexp.MustCompile("graffitiwall:([0-9]{1,3}):([0-9]{1,3}):#([0-9a-fA-F]{6})"),
		regexp.MustCompile("gw:([0-9]{3})([0-9]{3})([0-9a-fA-F]{6})"),
	}

	for _, slot := range blocks {
		for _, block := range slot {
			var matches []string
			for _, regex := range regexes {
				matches = regex.FindStringSubmatch(string(block.Graffiti))
				if len(matches) > 0 {
					break
				}
			}
			if len(matches) == 4 {
				x, err := strconv.Atoi(matches[1])
				if err != nil || x >= 1000 {
					return fmt.Errorf("error parsing x coordinate for graffiti %v of block %v", string(block.Graffiti), block.Slot)
				}

				y, err := strconv.Atoi(matches[2])
				if err != nil || y >= 1000 {
					return fmt.Errorf("error parsing y coordinate for graffiti %v of block %v", string(block.Graffiti), block.Slot)
				}
				color := matches[3]

				logger.Infof("set graffiti at %v - %v with color %v for slot %v by validator %v", x, y, color, block.Slot, block.Proposer)
				_, err = stmtGraffitiwall.Exec(x, y, color, block.Slot, block.Proposer)

				if err != nil {
					return fmt.Errorf("error executing graffitiwall statement: %v", err)
				}
			}
		}
	}
	return nil
}

func saveValidators(data *types.EpochData, tx *sql.Tx) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_save_validators").Observe(time.Since(start).Seconds())
	}()

	validators := data.Validators

	validatorsByIndex := make(map[uint64]*types.Validator, len(data.Validators))
	for _, v := range data.Validators {
		validatorsByIndex[v.Index] = v
	}
	slots := make([]uint64, 0, len(data.Blocks))
	for slot := range data.Blocks {
		slots = append(slots, slot)
	}
	sort.Slice(slots, func(i, j int) bool {
		return slots[i] < slots[j]
	})
	for _, slot := range slots {
		for _, b := range data.Blocks[slot] {
			if !b.Canonical {
				continue
			}
			propVal := validatorsByIndex[b.Proposer]
			if propVal != nil {
				propVal.LastProposalSlot = b.Slot
			}
			for _, a := range b.Attestations {
				for _, v := range a.Attesters {
					attVal := validatorsByIndex[v]
					if attVal != nil {
						attVal.LastAttestationSlot = a.Data.Slot
					}
				}
			}
		}
	}

	var latestBlock uint64
	err := WriterDb.Get(&latestBlock, "SELECT COALESCE(MAX(slot), 0) FROM blocks WHERE status = '1'")
	if err != nil {
		return err
	}

	thresholdSlot := latestBlock - 64
	if latestBlock < 64 {
		thresholdSlot = 0
	}

	farFutureEpoch := uint64(18446744073709551615)
	maxSqlNumber := uint64(9223372036854775807)

	for _, v := range validators {
		if v.WithdrawableEpoch == farFutureEpoch {
			v.WithdrawableEpoch = maxSqlNumber
		}
		if v.ExitEpoch == farFutureEpoch {
			v.ExitEpoch = maxSqlNumber
		}
		if v.ActivationEligibilityEpoch == farFutureEpoch {
			v.ActivationEligibilityEpoch = maxSqlNumber
		}
		if v.ActivationEpoch == farFutureEpoch {
			v.ActivationEpoch = maxSqlNumber
		}
	}

	batchSize := 4000 // max parameters: 65535
	for b := 0; b < len(validators); b += batchSize {
		start := b
		end := b + batchSize
		if len(validators) < end {
			end = len(validators)
		}

		numArgs := 16
		valueStrings := make([]string, 0, batchSize)
		valueArgs := make([]interface{}, 0, batchSize*numArgs)
		for i, v := range validators[start:end] {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*numArgs+1, i*numArgs+2, i*numArgs+3, i*numArgs+4, i*numArgs+5, i*numArgs+6, i*numArgs+7, i*numArgs+8, i*numArgs+9, i*numArgs+10, i*numArgs+11, i*numArgs+12, i*numArgs+13, i*numArgs+14, i*numArgs+15, i*numArgs+16))
			valueArgs = append(valueArgs, v.Index)
			valueArgs = append(valueArgs, v.PublicKey)
			valueArgs = append(valueArgs, v.WithdrawableEpoch)
			valueArgs = append(valueArgs, v.WithdrawalCredentials)
			valueArgs = append(valueArgs, v.Balance)
			valueArgs = append(valueArgs, v.EffectiveBalance)
			valueArgs = append(valueArgs, v.Slashed)
			valueArgs = append(valueArgs, v.ActivationEligibilityEpoch)
			valueArgs = append(valueArgs, v.ActivationEpoch)
			valueArgs = append(valueArgs, v.ExitEpoch)
			valueArgs = append(valueArgs, v.Balance1d)
			valueArgs = append(valueArgs, v.Balance7d)
			valueArgs = append(valueArgs, v.Balance31d)
			valueArgs = append(valueArgs, fmt.Sprintf("%x", v.PublicKey))
			valueArgs = append(valueArgs, v.Status)
			valueArgs = append(valueArgs, v.LastAttestationSlot)
		}
		stmt := fmt.Sprintf(`
			INSERT INTO validators (
				validatorindex,
				pubkey,
				withdrawableepoch,
				withdrawalcredentials,
				balance,
				effectivebalance,
				slashed,
				activationeligibilityepoch,
				activationepoch,
				exitepoch,
				balance1d,
				balance7d,
				balance31d,
				pubkeyhex,
				status,
				lastattestationslot
			) 
			VALUES %[3]s
			ON CONFLICT (validatorindex) DO UPDATE SET 
				withdrawableepoch          = EXCLUDED.withdrawableepoch,
				balance                    = EXCLUDED.balance,
				effectivebalance           = EXCLUDED.effectivebalance,
				slashed                    = EXCLUDED.slashed,
				activationeligibilityepoch = EXCLUDED.activationeligibilityepoch,
				activationepoch            = EXCLUDED.activationepoch,
				exitepoch                  = EXCLUDED.exitepoch,
				balance1d                  = EXCLUDED.balance1d,
				balance7d                  = EXCLUDED.balance7d,
				balance31d                 = EXCLUDED.balance31d,
				lastattestationslot        = 
					CASE 
					WHEN EXCLUDED.lastattestationslot > COALESCE(validators.lastattestationslot, 0) THEN EXCLUDED.lastattestationslot 
					ELSE validators.lastattestationslot 
					END,
				status                     = 
					CASE 
					WHEN EXCLUDED.exitepoch <= %[1]d AND EXCLUDED.slashed THEN 'slashed'
					WHEN EXCLUDED.exitepoch <= %[1]d THEN 'exited'
					WHEN EXCLUDED.activationeligibilityepoch = 9223372036854775807 THEN 'deposited'
					WHEN EXCLUDED.activationepoch > %[1]d THEN 'pending'
					WHEN EXCLUDED.slashed AND EXCLUDED.activationepoch < %[1]d AND GREATEST(EXCLUDED.lastattestationslot, validators.lastattestationslot) < %[2]d THEN 'slashing_offline'
					WHEN EXCLUDED.slashed THEN 'slashing_online'
					WHEN EXCLUDED.exitepoch < 9223372036854775807 AND GREATEST(EXCLUDED.lastattestationslot, validators.lastattestationslot) < %[2]d THEN 'exiting_offline'
					WHEN EXCLUDED.exitepoch < 9223372036854775807 THEN 'exiting_online'
					WHEN EXCLUDED.activationepoch < %[1]d AND GREATEST(EXCLUDED.lastattestationslot, validators.lastattestationslot) < %[2]d THEN 'active_offline' 
					ELSE 'active_online'
					END`,
			latestBlock, thresholdSlot, strings.Join(valueStrings, ","))
		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			return err
		}

		logger.Infof("saving validator batch %v completed", b)
	}

	s := time.Now()
	_, err = tx.Exec("update validators set balanceactivation = (select balance from validator_balances_p where validator_balances_p.week = validators.activationepoch / 1575 and validator_balances_p.epoch = validators.activationepoch and validator_balances_p.validatorindex = validators.validatorindex) WHERE balanceactivation IS NULL;")
	if err != nil {
		return err
	}
	logger.Infof("updating validator activation epoch balance completed, took %v", time.Since(s))

	return nil
}

func saveValidatorProposalAssignments(epoch uint64, assignments map[uint64]uint64, tx *sql.Tx) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_save_proposal_assignments").Observe(time.Since(start).Seconds())
	}()

	stmt, err := tx.Prepare(`
		INSERT INTO proposal_assignments (epoch, validatorindex, proposerslot, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (epoch, validatorindex, proposerslot) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for slot, validator := range assignments {
		_, err := stmt.Exec(epoch, validator, slot, 0)
		if err != nil {
			return fmt.Errorf("error executing save validator proposal assignment statement: %v", err)
		}
	}

	return nil
}

func saveValidatorAttestationAssignments(epoch uint64, assignments map[string]uint64, tx *sql.Tx) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_save_attestation_assignments").Observe(time.Since(start).Seconds())
	}()

	//args := make([][]interface{}, 0, len(assignments))
	argsWeek := make([][]interface{}, 0, len(assignments))
	for key, validator := range assignments {
		keySplit := strings.Split(key, "-")
		//args = append(args, []interface{}{epoch, validator, keySplit[0], keySplit[1], 0})
		argsWeek = append(argsWeek, []interface{}{epoch, validator, keySplit[0], keySplit[1], 0, epoch / 1575})
	}

	batchSize := 10000

	for b := 0; b < len(argsWeek); b += batchSize {
		start := b
		end := b + batchSize
		if len(argsWeek) < end {
			end = len(argsWeek)
		}

		valueStrings := make([]string, 0, batchSize)
		valueArgs := make([]interface{}, 0, batchSize*6)
		for i, v := range argsWeek[start:end] {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6))
			valueArgs = append(valueArgs, v...)
		}
		stmt := fmt.Sprintf(`
		INSERT INTO attestation_assignments_p (epoch, validatorindex, attesterslot, committeeindex, status, week)
		VALUES %s
		ON CONFLICT (validatorindex, week, epoch) DO UPDATE SET attesterslot = EXCLUDED.attesterslot, committeeindex = EXCLUDED.committeeindex`, strings.Join(valueStrings, ","))
		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			return fmt.Errorf("error executing save validator attestation assignment statement: %v", err)
		}
	}

	return nil
}

func saveValidatorBalances(epoch uint64, validators []*types.Validator, tx *sql.Tx) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_save_validator_balances").Observe(time.Since(start).Seconds())
	}()

	batchSize := 10000

	for b := 0; b < len(validators); b += batchSize {
		start := b
		end := b + batchSize
		if len(validators) < end {
			end = len(validators)
		}

		valueStrings := make([]string, 0, batchSize)
		valueArgs := make([]interface{}, 0, batchSize*5)
		for i, v := range validators[start:end] {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", i*5+1, i*5+2, i*5+3, i*5+4, i*5+5))
			valueArgs = append(valueArgs, epoch)
			valueArgs = append(valueArgs, v.Index)
			valueArgs = append(valueArgs, v.Balance)
			valueArgs = append(valueArgs, v.EffectiveBalance)
			valueArgs = append(valueArgs, epoch/1575)
		}
		stmt := fmt.Sprintf(`
		INSERT INTO validator_balances_p (epoch, validatorindex, balance, effectivebalance, week)
		VALUES %s
		ON CONFLICT (epoch, validatorindex, week) DO UPDATE SET
			balance          = EXCLUDED.balance,
			effectivebalance = EXCLUDED.effectivebalance`, strings.Join(valueStrings, ","))
		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			return err
		}
	}

	return nil
}

func saveValidatorBalancesRecent(epoch uint64, validators []*types.Validator, tx *sql.Tx) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_save_validator_balances_recent").Observe(time.Since(start).Seconds())
	}()

	batchSize := 10000

	for b := 0; b < len(validators); b += batchSize {
		start := b
		end := b + batchSize
		if len(validators) < end {
			end = len(validators)
		}

		valueStrings := make([]string, 0, batchSize)
		valueArgs := make([]interface{}, 0, batchSize*3)
		for i, v := range validators[start:end] {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
			valueArgs = append(valueArgs, epoch)
			valueArgs = append(valueArgs, v.Index)
			valueArgs = append(valueArgs, v.Balance)
		}
		stmt := fmt.Sprintf(`
			INSERT INTO validator_balances_recent (epoch, validatorindex, balance)
			VALUES %s
			ON CONFLICT (epoch, validatorindex) DO UPDATE SET
				balance = EXCLUDED.balance`, strings.Join(valueStrings, ","))

		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			return err
		}
	}

	if epoch > 10 {
		_, err := tx.Exec("DELETE FROM validator_balances_recent WHERE epoch < $1", epoch-10)
		if err != nil {
			return err
		}
	}

	return nil
}

func saveBlocks(blocks map[uint64]map[string]*types.Block, tx *sql.Tx) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_save_blocks").Observe(time.Since(start).Seconds())
	}()

	stmtBlock, err := tx.Prepare(`
		INSERT INTO blocks (epoch, slot, blockroot, parentroot, stateroot, signature, randaoreveal, graffiti, graffiti_text, eth1data_depositroot, eth1data_depositcount, eth1data_blockhash, syncaggregate_bits, syncaggregate_signature, proposerslashingscount, attesterslashingscount, attestationscount, depositscount, voluntaryexitscount, syncaggregate_participation, proposer, status, exec_parent_hash, exec_fee_recipient, exec_state_root, exec_receipts_root, exec_logs_bloom, exec_random, exec_block_number, exec_gas_limit, exec_gas_used, exec_timestamp, exec_extra_data, exec_base_fee_per_gas, exec_block_hash, exec_transactions_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36)
		ON CONFLICT (slot, blockroot) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtBlock.Close()

	stmtTransaction, err := tx.Prepare(`
		INSERT INTO blocks_transactions (block_slot, block_index, block_root, raw, txhash, nonce, gas_price, gas_limit, sender, recipient, amount, payload, max_priority_fee_per_gas, max_fee_per_gas)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (block_slot, block_index) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtTransaction.Close()

	stmtProposerSlashing, err := tx.Prepare(`
		INSERT INTO blocks_proposerslashings (block_slot, block_index, block_root, proposerindex, header1_slot, header1_parentroot, header1_stateroot, header1_bodyroot, header1_signature, header2_slot, header2_parentroot, header2_stateroot, header2_bodyroot, header2_signature)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (block_slot, block_index) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtProposerSlashing.Close()

	stmtAttesterSlashing, err := tx.Prepare(`
		INSERT INTO blocks_attesterslashings (block_slot, block_index, block_root, attestation1_indices, attestation1_signature, attestation1_slot, attestation1_index, attestation1_beaconblockroot, attestation1_source_epoch, attestation1_source_root, attestation1_target_epoch, attestation1_target_root, attestation2_indices, attestation2_signature, attestation2_slot, attestation2_index, attestation2_beaconblockroot, attestation2_source_epoch, attestation2_source_root, attestation2_target_epoch, attestation2_target_root)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		ON CONFLICT (block_slot, block_index) DO UPDATE SET attestation1_indices = excluded.attestation1_indices, attestation2_indices = excluded.attestation2_indices`)
	if err != nil {
		return err
	}
	defer stmtAttesterSlashing.Close()

	stmtAttestations, err := tx.Prepare(`
		INSERT INTO blocks_attestations (block_slot, block_index, block_root, aggregationbits, validators, signature, slot, committeeindex, beaconblockroot, source_epoch, source_root, target_epoch, target_root)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (block_slot, block_index) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtAttestations.Close()

	stmtDeposits, err := tx.Prepare(`
		INSERT INTO blocks_deposits (block_slot, block_index, block_root, proof, publickey, withdrawalcredentials, amount, signature)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		ON CONFLICT (block_slot, block_index) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtDeposits.Close()

	stmtVoluntaryExits, err := tx.Prepare(`
		INSERT INTO blocks_voluntaryexits (block_slot, block_index, block_root, epoch, validatorindex, signature)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (block_slot, block_index) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtVoluntaryExits.Close()

	stmtProposalAssignments, err := tx.Prepare(`
		INSERT INTO proposal_assignments (epoch, validatorindex, proposerslot, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (epoch, validatorindex, proposerslot) DO UPDATE SET status = excluded.status`)
	if err != nil {
		return err
	}
	defer stmtProposalAssignments.Close()

	stmtValidatorsLastAttestationSlot, err := tx.Prepare(`UPDATE validators SET lastattestationslot = $1 WHERE validatorindex = ANY($2::int[])`)
	if err != nil {
		return err
	}
	defer stmtValidatorsLastAttestationSlot.Close()

	slots := make([]uint64, 0, len(blocks))
	for slot := range blocks {
		slots = append(slots, slot)
	}
	sort.Slice(slots, func(i, j int) bool {
		return slots[i] < slots[j]
	})

	for _, slot := range slots {
		for _, b := range blocks[slot] {
			start := time.Now()
			blockLog := logger.WithFields(logrus.Fields{"slot": b.Slot, "blockRoot": fmt.Sprintf("%x", b.BlockRoot)})

			var dbBlockRootHash []byte
			err := WriterDb.Get(&dbBlockRootHash, "SELECT blockroot FROM blocks WHERE slot = $1 and blockroot = $2", b.Slot, b.BlockRoot)
			if err == nil && bytes.Compare(dbBlockRootHash, b.BlockRoot) == 0 {
				blockLog.Infof("skipping export of block as it is already present in the db")
				continue
			}
			blockLog.WithField("duration", time.Since(start)).Tracef("check if exists")
			t := time.Now()

			res, err := tx.Exec("DELETE FROM blocks WHERE slot = $1 AND length(blockroot) = 1", b.Slot) // Delete placeholder block
			if err != nil {
				return fmt.Errorf("error deleting placeholder block: %w", err)
			}
			ra, err := res.RowsAffected()
			if err != nil && ra > 0 {
				blockLog.Infof("deleted placeholder block")
			}
			blockLog.WithField("duration", time.Since(t)).Tracef("delete placeholder")
			t = time.Now()

			// Set proposer to MAX_SQL_INTEGER if it is the genesis-block (since we are using integers for validator-indices right now)
			if b.Slot == 0 {
				b.Proposer = 2147483647
			}
			syncAggBits := []byte{}
			syncAggSig := []byte{}
			syncAggParticipation := 0.0
			if b.SyncAggregate != nil {
				syncAggBits = b.SyncAggregate.SyncCommitteeBits
				syncAggSig = b.SyncAggregate.SyncCommitteeSignature
				syncAggParticipation = b.SyncAggregate.SyncAggregateParticipation
				// blockLog = blockLog.WithField("syncParticipation", b.SyncAggregate.SyncAggregateParticipation)
			}

			parentHash := []byte{}
			feeRecipient := []byte{}
			stateRoot := []byte{}
			receiptRoot := []byte{}
			logsBloom := []byte{}
			random := []byte{}
			blockNumber := uint64(0)
			gasLimit := uint64(0)
			gasUsed := uint64(0)
			timestamp := uint64(0)
			extraData := []byte{}
			baseFeePerGas := uint64(0)
			blockHash := []byte{}
			txCount := 0
			if b.ExecutionPayload != nil {
				parentHash = b.ExecutionPayload.ParentHash
				feeRecipient = b.ExecutionPayload.FeeRecipient
				stateRoot = b.ExecutionPayload.StateRoot
				receiptRoot = b.ExecutionPayload.ReceiptsRoot
				logsBloom = b.ExecutionPayload.LogsBloom
				random = b.ExecutionPayload.Random
				blockNumber = b.ExecutionPayload.BlockNumber
				gasLimit = b.ExecutionPayload.GasLimit
				gasUsed = b.ExecutionPayload.GasUsed
				timestamp = b.ExecutionPayload.Timestamp
				extraData = b.ExecutionPayload.ExtraData
				baseFeePerGas = b.ExecutionPayload.BaseFeePerGas
				blockHash = b.ExecutionPayload.BlockHash
				txCount = len(b.ExecutionPayload.Transactions)
			}
			_, err = stmtBlock.Exec(
				b.Slot/utils.Config.Chain.Config.SlotsPerEpoch,
				b.Slot,
				b.BlockRoot,
				b.ParentRoot,
				b.StateRoot,
				b.Signature,
				b.RandaoReveal,
				b.Graffiti,
				utils.GraffitiToSring(b.Graffiti),
				b.Eth1Data.DepositRoot,
				b.Eth1Data.DepositCount,
				b.Eth1Data.BlockHash,
				syncAggBits,
				syncAggSig,
				len(b.ProposerSlashings),
				len(b.AttesterSlashings),
				len(b.Attestations),
				len(b.Deposits),
				len(b.VoluntaryExits),
				syncAggParticipation,
				b.Proposer,
				strconv.FormatUint(b.Status, 10),
				parentHash,
				feeRecipient,
				stateRoot,
				receiptRoot,
				logsBloom,
				random,
				blockNumber,
				gasLimit,
				gasUsed,
				timestamp,
				extraData,
				baseFeePerGas,
				blockHash,
				txCount,
			)
			if err != nil {
				return fmt.Errorf("error executing stmtBlocks for block %v: %w", b.Slot, err)
			}
			blockLog.WithField("duration", time.Since(t)).Tracef("stmtBlock")
			t = time.Now()

			n := time.Now()
			logger.Tracef("done, took %v", time.Since(n))
			logger.Tracef("writing transactions data")
			if payload := b.ExecutionPayload; payload != nil {
				for i, tx := range payload.Transactions {
					_, err := stmtTransaction.Exec(b.Slot, i, b.BlockRoot,
						tx.Raw, tx.TxHash, tx.AccountNonce, tx.Price, tx.GasLimit, tx.Sender, tx.Recipient, tx.Amount, tx.Payload, tx.MaxPriorityFeePerGas, tx.MaxFeePerGas)
					if err != nil {
						return fmt.Errorf("error executing stmtTransaction for block %v: %v", b.Slot, err)
					}
				}
			}
			logger.Tracef("done, took %v", time.Since(n))
			n = time.Now()
			logger.Tracef("writing proposer slashings data")
			for i, ps := range b.ProposerSlashings {
				_, err := stmtProposerSlashing.Exec(b.Slot, i, b.BlockRoot, ps.ProposerIndex, ps.Header1.Slot, ps.Header1.ParentRoot, ps.Header1.StateRoot, ps.Header1.BodyRoot, ps.Header1.Signature, ps.Header2.Slot, ps.Header2.ParentRoot, ps.Header2.StateRoot, ps.Header2.BodyRoot, ps.Header2.Signature)
				if err != nil {
					return fmt.Errorf("error executing stmtProposerSlashing for block %v: %w", b.Slot, err)
				}
			}
			blockLog.WithField("duration", time.Since(t)).Tracef("stmtProposerSlashing")
			t = time.Now()

			for i, as := range b.AttesterSlashings {
				_, err := stmtAttesterSlashing.Exec(b.Slot, i, b.BlockRoot, pq.Array(as.Attestation1.AttestingIndices), as.Attestation1.Signature, as.Attestation1.Data.Slot, as.Attestation1.Data.CommitteeIndex, as.Attestation1.Data.BeaconBlockRoot, as.Attestation1.Data.Source.Epoch, as.Attestation1.Data.Source.Root, as.Attestation1.Data.Target.Epoch, as.Attestation1.Data.Target.Root, pq.Array(as.Attestation2.AttestingIndices), as.Attestation2.Signature, as.Attestation2.Data.Slot, as.Attestation2.Data.CommitteeIndex, as.Attestation2.Data.BeaconBlockRoot, as.Attestation2.Data.Source.Epoch, as.Attestation2.Data.Source.Root, as.Attestation2.Data.Target.Epoch, as.Attestation2.Data.Target.Root)
				if err != nil {
					return fmt.Errorf("error executing stmtAttesterSlashing for block %v: %w", b.Slot, err)
				}
			}
			blockLog.WithField("duration", time.Since(t)).Tracef("stmtAttesterSlashing")
			t = time.Now()

			if b.Status == 2 {
				// set status of sync_assignments to 3 if block is missed
				_, err := tx.Exec(`UPDATE sync_assignments_p SET status = 3 WHERE week = $1 AND slot = $2`, utils.WeekOfSlot(b.Slot), b.Slot)
				if err != nil {
					return fmt.Errorf("error updating status of sync_assignments to orhphan for block %v: %w", b.Slot, err)
				}
			} else if b.SyncAggregate != nil && len(b.SyncAggregate.SyncCommitteeValidators) > 0 {
				// update sync_assignments table if block is a post-altair-activation-block with sync-aggregate
				bitLen := len(b.SyncAggregate.SyncCommitteeBits) * 8
				valLen := len(b.SyncAggregate.SyncCommitteeValidators)
				if bitLen < valLen {
					return fmt.Errorf("error getting sync_committee participants: bitLen != valLen: %v != %v", bitLen, valLen)
				}
				nArgs := 4
				valueStrings := make([]string, valLen)
				valueArgs := make([]interface{}, valLen*nArgs)
				for i, valIndex := range b.SyncAggregate.SyncCommitteeValidators {
					valueStrings[i] = fmt.Sprintf("($%d, $%d, $%d, $%d)", i*nArgs+1, i*nArgs+2, i*nArgs+3, i*nArgs+4)
					valueArgs[i*nArgs] = b.Slot
					valueArgs[i*nArgs+1] = valIndex
					if utils.BitAtVector(b.SyncAggregate.SyncCommitteeBits, i) {
						valueArgs[i*nArgs+2] = 1
					} else {
						valueArgs[i*nArgs+2] = 2
					}
					valueArgs[i*nArgs+3] = utils.WeekOfSlot(b.Slot)
				}
				stmt := fmt.Sprintf(`
					INSERT INTO sync_assignments_p (slot, validatorindex, status, week)
					VALUES %s
					ON CONFLICT (slot, validatorindex, week) DO UPDATE SET status = excluded.status`, strings.Join(valueStrings, ","))
				_, err := tx.Exec(stmt, valueArgs...)
				if err != nil {
					return fmt.Errorf("error executing sync_assignments insert for block %v: %w", b.Slot, err)
				}
			}
			blockLog.WithField("duration", time.Since(t)).Tracef("sync_assignments_p")
			t = time.Now()

			for i, a := range b.Attestations {
				attestationAssignmentsArgsWeek := make([][]interface{}, 0, 20000)
				attestingValidators := make([]string, 0, 20000)

				for _, validator := range a.Attesters {
					attestationAssignmentsArgsWeek = append(attestationAssignmentsArgsWeek, []interface{}{a.Data.Slot / utils.Config.Chain.Config.SlotsPerEpoch, validator, a.Data.Slot, a.Data.CommitteeIndex, 1, b.Slot, a.Data.Slot / utils.Config.Chain.Config.SlotsPerEpoch / 1575})
					attestingValidators = append(attestingValidators, strconv.FormatUint(validator, 10))
				}

				batchSize := 20000

				for batch := 0; batch < len(attestationAssignmentsArgsWeek); batch += batchSize {
					start := batch
					end := batch + batchSize
					if len(attestationAssignmentsArgsWeek) < end {
						end = len(attestationAssignmentsArgsWeek)
					}

					valueStrings := make([]string, 0, batchSize)
					valueArgs := make([]interface{}, 0, batchSize*7)
					for i, v := range attestationAssignmentsArgsWeek[start:end] {
						valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*7+1, i*7+2, i*7+3, i*7+4, i*7+5, i*7+6, i*7+7))
						valueArgs = append(valueArgs, v...)
					}
					stmt := fmt.Sprintf(`
						INSERT INTO attestation_assignments_p (epoch, validatorindex, attesterslot, committeeindex, status, inclusionslot, week)
						VALUES %s
						ON CONFLICT (validatorindex, week, epoch) DO UPDATE SET status = excluded.status, inclusionslot = LEAST((CASE WHEN attestation_assignments_p.inclusionslot = 0 THEN null ELSE attestation_assignments_p.inclusionslot END), excluded.inclusionslot)`, strings.Join(valueStrings, ","))
					_, err := tx.Exec(stmt, valueArgs...)
					if err != nil {
						return fmt.Errorf("error executing stmtAttestationAssignments_p for block %v: %w", b.Slot, err)
					}
				}

				// _, err = stmtValidatorsLastAttestationSlot.Exec(a.Data.Slot, "{"+strings.Join(attestingValidators, ",")+"}")
				// if err != nil {
				// 	return fmt.Errorf("error executing stmtValidatorsLastAttestationSlot for block %v: %w", b.Slot, err)
				// }

				_, err = stmtAttestations.Exec(b.Slot, i, b.BlockRoot, a.AggregationBits, pq.Array(a.Attesters), a.Signature, a.Data.Slot, a.Data.CommitteeIndex, a.Data.BeaconBlockRoot, a.Data.Source.Epoch, a.Data.Source.Root, a.Data.Target.Epoch, a.Data.Target.Root)
				if err != nil {
					return fmt.Errorf("error executing stmtAttestations for block %v: %w", b.Slot, err)
				}
			}
			blockLog.WithField("duration", time.Since(t)).Tracef("attestations")
			t = time.Now()

			for i, d := range b.Deposits {
				_, err := stmtDeposits.Exec(b.Slot, i, b.BlockRoot, nil, d.PublicKey, d.WithdrawalCredentials, d.Amount, d.Signature)
				if err != nil {
					return fmt.Errorf("error executing stmtDeposits for block %v: %w", b.Slot, err)
				}
			}
			blockLog.WithField("duration", time.Since(t)).Tracef("deposits")
			t = time.Now()

			for i, ve := range b.VoluntaryExits {
				_, err := stmtVoluntaryExits.Exec(b.Slot, i, b.BlockRoot, ve.Epoch, ve.ValidatorIndex, ve.Signature)
				if err != nil {
					return fmt.Errorf("error executing stmtVoluntaryExits for block %v: %w", b.Slot, err)
				}
			}
			blockLog.WithField("duration", time.Since(t)).Tracef("exits")
			t = time.Now()

			_, err = stmtProposalAssignments.Exec(b.Slot/utils.Config.Chain.Config.SlotsPerEpoch, b.Proposer, b.Slot, b.Status)
			if err != nil {
				return fmt.Errorf("error executing stmtProposalAssignments for block %v: %w", b.Slot, err)
			}
			blockLog.WithField("duration", time.Since(t)).Tracef("stmtProposalAssignments")
			t = time.Now()

			blockLog.Infof("export of block completed, took %v", time.Since(start))
		}
	}

	return nil
}

// UpdateEpochStatus will update the epoch status in the database
func UpdateEpochStatus(stats *types.ValidatorParticipation) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_epochs_status").Observe(time.Since(start).Seconds())
	}()

	_, err := WriterDb.Exec(`
		UPDATE epochs SET
			finalized = $1,
			eligibleether = $2,
			globalparticipationrate = $3,
			votedether = $4
		WHERE epoch = $5`,
		stats.Finalized, stats.EligibleEther, stats.GlobalParticipationRate, stats.VotedEther, stats.Epoch)

	return err
}

// UpdateEpochFinalization will update finalized-flag of all epochs before the last finalized epoch
func UpdateEpochFinalization() error {
	_, err := WriterDb.Exec(`UPDATE epochs SET finalized = true WHERE epoch < (SELECT MAX(epoch) FROM epochs WHERE finalized = true)`)
	return err
}

// GetTotalValidatorsCount will return the total-validator-count
func GetTotalValidatorsCount() (uint64, error) {
	var totalCount uint64
	err := ReaderDb.Get(&totalCount, "SELECT COUNT(*) FROM validators")
	return totalCount, err
}

// GetActiveValidatorCount will return the total-validator-count
func GetActiveValidatorCount() (uint64, error) {
	var count uint64
	err := ReaderDb.Get(&count, "select count(*) from validators where status in ('active_offline', 'active_online');")
	return count, err
}

func GetValidatorNames() (map[uint64]string, error) {
	rows, err := ReaderDb.Query(`
		SELECT validatorindex, validator_names.name 
		FROM validators 
		LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
		WHERE validator_names.name IS NOT NULL`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	validatorIndexToNameMap := make(map[uint64]string, 30000)

	for rows.Next() {
		var index uint64
		var name string

		err := rows.Scan(&index, &name)

		if err != nil {
			return nil, err
		}
		validatorIndexToNameMap[index] = name
	}

	return validatorIndexToNameMap, nil
}

// GetPendingValidatorCount queries the pending validators currently in the queue
func GetPendingValidatorCount() (uint64, error) {
	count := uint64(0)
	err := ReaderDb.Get(&count, "SELECT entering_validators_count FROM queue ORDER BY ts DESC LIMIT 1")
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("error retrieving validator queue count: %v", err)
	}
	return count, nil
}

func GetTotalEligibleEther() (uint64, error) {
	var total uint64

	err := ReaderDb.Get(&total, `
		SELECT eligibleether FROM epochs ORDER BY epoch desc LIMIT 1
	`)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return total / 1e9, nil
}

func GetDepositThresholdTime() (*time.Time, error) {
	var threshold *time.Time
	err := ReaderDb.Get(&threshold, `
	select min(block_ts) from (
		select block_ts, block_number, sum(amount) over (order by block_ts) as totalsum
			from (
				SELECT
					publickey,
					32e9 AS amount,
					MAX(block_ts) as block_ts,
					MAX(block_number) as block_number
				FROM eth1_deposits
				WHERE valid_signature = true
				GROUP BY publickey
				HAVING SUM(amount) >= 32e9
			) a
		) b
		where totalsum > $1;
		 `, utils.Config.Chain.Config.MinGenesisActiveValidatorCount*32e9)
	if err != nil {
		return nil, err
	}
	return threshold, nil
}

// GetValidatorsBalanceDecrease returns all validators whose balance decreased for 3 consecutive epochs. It looks 10 epochs back for when the balance increased the last time
func GetValidatorsBalanceDecrease(epoch uint64) ([]struct {
	Pubkey         string `db:"pubkey"`
	ValidatorIndex uint64 `db:"validatorindex"`
	StartBalance   uint64 `db:"startbalance"`
	EndBalance     uint64 `db:"endbalance"`
}, error) {

	var dbResult []struct {
		Pubkey         string `db:"pubkey"`
		ValidatorIndex uint64 `db:"validatorindex"`
		StartBalance   uint64 `db:"startbalance"`
		EndBalance     uint64 `db:"endbalance"`
	}

	err := ReaderDb.Select(&dbResult, `
	SELECT validatorindex, startbalance, endbalance, a.pubkey AS pubkey FROM (
		SELECT 
			v.validatorindex,
			v.pubkeyhex AS pubkey, 
			vb0.balance AS endbalance, 
			vb3.balance AS startbalance, 
			(SELECT MAX(epoch) FROM (
				SELECT epoch, balance-LAG(balance) OVER (ORDER BY epoch) AS diff
				FROM validator_balances_recent 
				WHERE validatorindex = v.validatorindex AND epoch > $1 - 10
			) b WHERE diff > 0) AS lastbalanceincreaseepoch
		from validators v
		INNER JOIN validator_balances_recent vb0 ON v.validatorindex = vb0.validatorindex AND vb0.epoch = $1
		INNER JOIN validator_balances_recent vb1 ON v.validatorindex = vb1.validatorindex AND vb1.epoch = $1 - 1 AND vb1.balance > vb0.balance
		INNER JOIN validator_balances_recent vb2 ON v.validatorindex = vb2.validatorindex AND vb2.epoch = $1 - 2 AND vb2.balance > vb1.balance
		INNER JOIN validator_balances_recent vb3 ON v.validatorindex = vb3.validatorindex AND vb3.epoch = $1 - 3 AND vb3.balance > vb2.balance
	) a WHERE lastbalanceincreaseepoch IS NOT NULL
	`, epoch)

	if err != nil {
		return nil, err
	}

	return dbResult, nil
}

// GetValidatorsGotSlashed returns the validators that got slashed after `epoch` either by an attestation violation or a proposer violation
func GetValidatorsGotSlashed(epoch uint64) ([]struct {
	Epoch                  uint64 `db:"epoch"`
	SlasherIndex           uint64 `db:"slasher"`
	SlasherPubkey          string `db:"slasher_pubkey"`
	SlashedValidatorIndex  uint64 `db:"slashedvalidator"`
	SlashedValidatorPubkey []byte `db:"slashedvalidator_pubkey"`
	Reason                 string `db:"reason"`
}, error) {

	var dbResult []struct {
		Epoch                  uint64 `db:"epoch"`
		SlasherIndex           uint64 `db:"slasher"`
		SlasherPubkey          string `db:"slasher_pubkey"`
		SlashedValidatorIndex  uint64 `db:"slashedvalidator"`
		SlashedValidatorPubkey []byte `db:"slashedvalidator_pubkey"`
		Reason                 string `db:"reason"`
	}
	err := ReaderDb.Select(&dbResult, `
		WITH
			slashings AS (
				SELECT DISTINCT ON (slashedvalidator) 
					slot,
					epoch,
					slasher,
					slashedvalidator,
					reason
				FROM (
					SELECT
						blocks.slot, 
						blocks.epoch, 
						blocks.proposer AS slasher, 
						UNNEST(ARRAY(
							SELECT UNNEST(attestation1_indices)
								INTERSECT
							SELECT UNNEST(attestation2_indices)
						)) AS slashedvalidator, 
						'Attestation Violation' AS reason
					FROM blocks_attesterslashings 
					LEFT JOIN blocks ON blocks_attesterslashings.block_slot = blocks.slot
					WHERE blocks.status = '1' AND blocks.epoch > $1
					UNION ALL
						SELECT
							blocks.slot, 
							blocks.epoch, 
							blocks.proposer AS slasher, 
							blocks_proposerslashings.proposerindex AS slashedvalidator,
							'Proposer Violation' AS reason 
						FROM blocks_proposerslashings
						LEFT JOIN blocks ON blocks_proposerslashings.block_slot = blocks.slot
						WHERE blocks.status = '1' AND blocks.epoch > $1
				) a
				ORDER BY slashedvalidator, slot
			)
		SELECT slasher, vk.pubkey as slasher_pubkey, slashedvalidator, vv.pubkey as slashedvalidator_pubkey, epoch, reason
		FROM slashings s
	    INNER JOIN validators vk ON s.slasher = vk.validatorindex
		INNER JOIN validators vv ON s.slashedvalidator = vv.validatorindex`, epoch)
	if err != nil {
		return nil, err
	}
	return dbResult, nil
}
