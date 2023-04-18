package db

import (
	"bytes"
	"crypto/sha1"
	"database/sql"
	"embed"
	"encoding/hex"
	"eth2-exporter/metrics"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"math/big"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/patrickmn/go-cache"
	"github.com/pressly/goose/v3"
	prysm_deposit "github.com/prysmaticlabs/prysm/v3/contracts/deposit"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"github.com/sirupsen/logrus"

	"eth2-exporter/rpc"

	"github.com/jackc/pgx/v4/pgxpool"
)

//go:embed migrations/*.sql
var EmbedMigrations embed.FS

var DBPGX *pgxpool.Conn

// DB is a pointer to the explorer-database
var WriterDb *sqlx.DB
var ReaderDb *sqlx.DB

var logger = logrus.StandardLogger().WithField("module", "db")

var epochsCache = cache.New(time.Hour, time.Minute)
var saveValidatorsMux = &sync.Mutex{}

func dbTestConnection(dbConn *sqlx.DB, dataBaseName string) {
	// The golang sql driver does not properly implement PingContext
	// therefore we use a timer to catch db connection timeouts
	dbConnectionTimeout := time.NewTimer(15 * time.Second)

	go func() {
		<-dbConnectionTimeout.C
		logger.Fatalf("timeout while connecting to %s", dataBaseName)
	}()

	err := dbConn.Ping()
	if err != nil {
		logger.Fatalf("unable to Ping %s: %s", dataBaseName, err)
	}

	dbConnectionTimeout.Stop()
}

func mustInitDB(writer *types.DatabaseConfig, reader *types.DatabaseConfig) (*sqlx.DB, *sqlx.DB) {
	dbConnWriter, err := sqlx.Open("pgx", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", writer.Username, writer.Password, writer.Host, writer.Port, writer.Name))
	if err != nil {
		utils.LogFatal(err, "error getting Connection Writer database", 0)
	}

	dbTestConnection(dbConnWriter, "database")
	dbConnWriter.SetConnMaxIdleTime(time.Second * 30)
	dbConnWriter.SetConnMaxLifetime(time.Second * 60)
	dbConnWriter.SetMaxOpenConns(200)
	dbConnWriter.SetMaxIdleConns(200)

	if reader == nil {
		return dbConnWriter, dbConnWriter
	}

	dbConnReader, err := sqlx.Open("pgx", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", reader.Username, reader.Password, reader.Host, reader.Port, reader.Name))
	if err != nil {
		utils.LogFatal(err, "error getting Connection Reader database", 0)
	}

	dbTestConnection(dbConnReader, "read replica database")
	dbConnReader.SetConnMaxIdleTime(time.Second * 30)
	dbConnReader.SetConnMaxLifetime(time.Second * 60)
	dbConnReader.SetMaxOpenConns(200)
	dbConnReader.SetMaxIdleConns(200)
	return dbConnWriter, dbConnReader
}

func MustInitDB(writer *types.DatabaseConfig, reader *types.DatabaseConfig) {
	WriterDb, ReaderDb = mustInitDB(writer, reader)
}

func ApplyEmbeddedDbSchema(version int64) error {
	goose.SetBaseFS(EmbedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if version == -2 {
		if err := goose.Up(WriterDb.DB, "migrations"); err != nil {
			return err
		}
	} else if version == -1 {
		if err := goose.UpByOne(WriterDb.DB, "migrations"); err != nil {
			return err
		}
	} else {
		if err := goose.UpTo(WriterDb.DB, "migrations", version); err != nil {
			return err
		}
	}

	return nil
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

var searchLikeHash = regexp.MustCompile(`^(0x)?[0-9a-fA-F]{2,96}`) // only search for pubkeys if string consists of 96 hex-chars

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
				SELECT pubkey, status AS state
				FROM validators
			) as v
		ON
			v.pubkey = eth1.publickey
		WHERE
			ENCODE(eth1.publickey, 'hex') LIKE LOWER($3)
			OR ENCODE(eth1.withdrawal_credentials, 'hex') LIKE LOWER($3)
			OR ENCODE(eth1.from_address, 'hex') LIKE LOWER($3)
			OR ENCODE(tx_hash, 'hex') LIKE LOWER($3)
			OR CAST(eth1.block_number AS text) LIKE LOWER($3)
		ORDER BY %s %s
		LIMIT $1
		OFFSET $2`, orderBy, orderDir)
		err = ReaderDb.Select(&deposits, wholeQuery, length, start, query+"%")
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
				SELECT pubkey, status AS state
				FROM validators
			) as v
		ON
			v.pubkey = eth1.publickey
		ORDER BY %s %s
		LIMIT $1
		OFFSET $2`, orderBy, orderDir), length, start)
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

func GetEth1DepositsLeaderboard(query string, length, start uint64, orderBy, orderDir string) ([]*types.EthOneDepositLeaderboardData, uint64, error) {
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
			break
		}
	}
	if !hasColumn {
		orderBy = "amount"
	}

	var err error
	var totalCount uint64
	if query != "" {
		err = ReaderDb.Get(&totalCount, `
		SELECT COUNT(*) FROM eth1_deposits_aggregated WHERE ENCODE(from_address, 'hex') LIKE LOWER($1)`, query+"%")
	} else {
		err = ReaderDb.Get(&totalCount, "SELECT COUNT(*) FROM eth1_deposits_aggregated AS count")
	}
	if err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	}

	if query != "" {
		err = ReaderDb.Select(&deposits, fmt.Sprintf(`
			SELECT from_address, amount, validcount, invalidcount, slashedcount, totalcount, activecount, pendingcount, voluntary_exit_count
			FROM eth1_deposits_aggregated
			WHERE ENCODE(from_address, 'hex') LIKE LOWER($3)
			ORDER BY %s %s
			LIMIT $1
			OFFSET $2`, orderBy, orderDir), length, start, query+"%")
	} else {
		err = ReaderDb.Select(&deposits, fmt.Sprintf(`
			SELECT from_address, amount, validcount, invalidcount, slashedcount, totalcount, activecount, pendingcount, voluntary_exit_count
			FROM eth1_deposits_aggregated
			ORDER BY %s %s
			LIMIT $1
			OFFSET $2`, orderBy, orderDir), length, start)
	}
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
			break
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
		return 0, fmt.Errorf("error retrieving latest epoch from DB: %w", err)
	}

	return epoch, nil
}

// GetAllEpochs will return a collection of all of the epochs from the database
func GetAllEpochs() ([]uint64, error) {
	var epochs []uint64
	err := WriterDb.Select(&epochs, "SELECT epoch FROM epochs ORDER BY epoch")

	if err != nil {
		return nil, fmt.Errorf("error retrieving all epochs from DB: %w", err)
	}

	return epochs, nil
}

// GetLastPendingAndProposedBlocks will return all proposed and pending blocks (ignores missed slots) from the database
func GetLastPendingAndProposedBlocks(startEpoch, endEpoch uint64) ([]*types.MinimalBlock, error) {
	var blocks []*types.MinimalBlock

	err := WriterDb.Select(&blocks, "SELECT epoch, slot, blockroot FROM blocks WHERE epoch >= $1 AND epoch <= $2 AND blockroot != '\x01' ORDER BY slot DESC", startEpoch, endEpoch)

	if err != nil {
		return nil, fmt.Errorf("error retrieving last blocks (%v-%v) from DB: %w", startEpoch, endEpoch, err)
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

// UpdateMissedBlocks will update the missed blocks for an epoch range in the database
func UpdateMissedBlocks(startEpoch, endEpoch uint64) error {
	_, err := WriterDb.Exec(`UPDATE blocks SET status = '2', blockroot = '\x01' WHERE status = '0' AND epoch >= $1 AND epoch <= $2`, startEpoch, endEpoch)
	return err
}

func UpdateMissedBlocksInEpochWithSlotCutoff(slot uint64) error {
	_, err := WriterDb.Exec(`UPDATE blocks SET status = '2', blockroot = '\x01' WHERE status = '0' AND epoch = $1 AND slot < $2`, slot/utils.Config.Chain.Config.SlotsPerEpoch, slot)
	return err
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
		return fmt.Errorf("error starting db transactions: %w", err)
	}
	defer tx.Rollback()

	lastSlotNumber := uint64(0)
	for _, block := range blocks {
		if block.Slot > lastSlotNumber {
			lastSlotNumber = block.Slot
		}
	}

	_, err = tx.Exec("UPDATE blocks SET status = 3 WHERE epoch >= $1 AND epoch <= $2 AND (status = '1' OR status = '3') AND slot <= $3", startEpoch, endEpoch, lastSlotNumber)
	if err != nil {
		return err
	}

	for _, block := range blocks {
		if block.Canonical {
			logger.Printf("marking block %x at slot %v as canonical", block.BlockRoot, block.Slot)
			_, err = tx.Exec("UPDATE blocks SET status = '1' WHERE blockroot = $1", block.BlockRoot)
			if err != nil {
				return err
			}
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
		return fmt.Errorf("error starting db transactions: %w", err)
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
		validators.Activating, validators.Exiting)
	return err
}

func SaveBlock(block *types.Block) error {

	blocksMap := make(map[uint64]map[string]*types.Block)
	if blocksMap[block.Slot] == nil {
		blocksMap[block.Slot] = make(map[string]*types.Block)
	}
	blocksMap[block.Slot][fmt.Sprintf("%x", block.BlockRoot)] = block

	tx, err := WriterDb.Beginx()
	if err != nil {
		return fmt.Errorf("error starting db transactions: %v", err)
	}
	defer tx.Rollback()

	logger.Infof("exporting block data")
	err = saveBlocks(blocksMap, tx)
	if err != nil {
		logger.Fatalf("error saving blocks to db: %v", err)
		return fmt.Errorf("error saving blocks to db: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing db transaction: %w", err)
	}

	return nil
}

// SaveEpoch will stave the epoch data into the database
func SaveEpoch(data *types.EpochData, client rpc.Client) error {
	// Check if we need to export the epoch
	hasher := sha1.New()
	slots := make([]uint64, 0, len(data.Blocks))
	for slot := range data.Blocks {
		slots = append(slots, slot)
	}
	sort.Slice(slots, func(i, j int) bool {
		return slots[i] < slots[j]
	})

	for _, slot := range slots {
		for _, b := range data.Blocks[slot] {
			hasher.Write(b.BlockRoot)
		}
	}

	epochCacheKey := fmt.Sprintf("%x", hasher.Sum(nil))
	logger.Infof("cache key for epoch %v is %v", data.Epoch, epochCacheKey)

	cachedEpochKey, found := epochsCache.Get(fmt.Sprintf("%v", data.Epoch))
	if found && epochCacheKey == cachedEpochKey.(string) {
		logger.Infof("skipping export of epoch %v as it did not change compared to the previous export run", data.Epoch)
		return nil
	}

	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_save_epoch").Observe(time.Since(start).Seconds())
		logger.WithFields(logrus.Fields{"epoch": data.Epoch, "duration": time.Since(start)}).Info("completed saving epoch")
	}()

	tx, err := WriterDb.Beginx()
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
		go func() {
			logger.Infof("exporting validators for epoch %v", data.Epoch)
			saveValidatorsMux.Lock()
			defer saveValidatorsMux.Unlock()
			logger.Infof("acquired saveValidatorsMux lock for epoch %v", data.Epoch)

			validatorsTx, err := WriterDb.Beginx()
			if err != nil {
				logger.Errorf("error starting validators tx: %v", err)
				return
			}
			defer validatorsTx.Rollback()

			err = saveValidators(data, validatorsTx, client)
			if err != nil {
				logger.Errorf("error saving validators to db: %v", err)
			}
			err = updateQueueDeposits()
			if err != nil {
				logger.Errorf("error updating queue deposits cache: %v", err)
			}

			err = validatorsTx.Commit()
			if err != nil {
				logger.Errorf("error committing validators tx: %v", err)
			}
		}()
	}
	logger.Infof("exporting proposal assignments data")
	err = saveValidatorProposalAssignments(data.Epoch, data.ValidatorAssignmentes.ProposerAssignments, tx)
	if err != nil {
		return fmt.Errorf("error saving validator proposal assignments to db: %w", err)
	}

	logger.Infof("exporting attestation assignments data")
	// only export validator balances for epoch zero (validator_balances_recent is only needed for genesis deposits)
	if data.Epoch == 0 {
		logger.Infof("exporting validator balances for epoch 0")
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
	withdrawalCount := 0

	for _, slot := range data.Blocks {
		for _, b := range slot {
			proposerSlashingsCount += len(b.ProposerSlashings)
			attesterSlashingsCount += len(b.AttesterSlashings)
			attestationsCount += len(b.Attestations)
			depositCount += len(b.Deposits)
			voluntaryExitCount += len(b.VoluntaryExits)
			if b.ExecutionPayload != nil {
				withdrawalCount += len(b.ExecutionPayload.Withdrawals)
			}

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

	finalized := false
	if data.Epoch == 0 {
		finalized = true
	}

	_, err = tx.Exec(`
		INSERT INTO epochs (
			epoch, 
			blockscount, 
			proposerslashingscount, 
			attesterslashingscount, 
			attestationscount, 
			depositscount,
			withdrawalcount,
			voluntaryexitscount, 
			validatorscount, 
			averagevalidatorbalance, 
			totalvalidatorbalance,
			finalized, 
			eligibleether, 
			globalparticipationrate, 
			votedether
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) 
		ON CONFLICT (epoch) DO UPDATE SET 
			blockscount             = excluded.blockscount, 
			proposerslashingscount  = excluded.proposerslashingscount,
			attesterslashingscount  = excluded.attesterslashingscount,
			attestationscount       = excluded.attestationscount,
			depositscount           = excluded.depositscount,
			withdrawalcount         = excluded.withdrawalcount,
			voluntaryexitscount     = excluded.voluntaryexitscount,
			validatorscount         = excluded.validatorscount,
			averagevalidatorbalance = excluded.averagevalidatorbalance,
			totalvalidatorbalance   = excluded.totalvalidatorbalance,
			eligibleether           = excluded.eligibleether,
			globalparticipationrate = excluded.globalparticipationrate,
			votedether              = excluded.votedether`,
		data.Epoch,
		len(data.Blocks),
		proposerSlashingsCount,
		attesterSlashingsCount,
		attestationsCount,
		depositCount,
		withdrawalCount,
		voluntaryExitCount,
		validatorsCount,
		validatorBalanceAverage.Uint64(),
		validatorBalanceSum.Uint64(),
		finalized,
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

	lookback := uint64(0)
	if data.Epoch > 3 {
		lookback = data.Epoch - 3
	}
	// delete duplicate scheduled slots
	_, err = WriterDb.Exec("delete from blocks where slot in (select slot from blocks where epoch >= $1 group by slot having count(*) > 1) and blockroot = $2;", lookback, []byte{0x0})
	if err != nil {
		return fmt.Errorf("error cleaning up blocks table: %w", err)
	}

	// delete duplicate missed blocks
	_, err = WriterDb.Exec("delete from blocks where slot in (select slot from blocks where epoch >= $1 group by slot having count(*) > 1) and blockroot = $2;", lookback, []byte{0x1})
	if err != nil {
		return fmt.Errorf("error cleaning up blocks table: %w", err)
	}

	epochsCache.Set(fmt.Sprintf("%v", data.Epoch), epochCacheKey, cache.DefaultExpiration)
	return nil
}

func saveGraffitiwall(blocks map[uint64]map[string]*types.Block, tx *sqlx.Tx) error {
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
					return fmt.Errorf("error executing graffitiwall statement: %w", err)
				}
			}
		}
	}
	return nil
}

func saveValidators(data *types.EpochData, tx *sqlx.Tx, client rpc.Client) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_save_validators").Observe(time.Since(start).Seconds())
	}()

	var genesisBalances map[uint64][]*types.ValidatorBalance

	if data.Epoch == 0 {
		var err error
		genesisBalances, err = BigtableClient.GetValidatorBalanceHistory([]uint64{}, 0, 0)
		if err != nil {
			return err
		}
	}

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
				propVal.LastProposalSlot = sql.NullInt64{Int64: int64(b.Slot), Valid: true}
			}
			for _, a := range b.Attestations {
				for _, v := range a.Attesters {
					attVal := validatorsByIndex[v]
					if attVal != nil {
						attVal.LastAttestationSlot = sql.NullInt64{
							Int64: int64(a.Data.Slot),
							Valid: true,
						}
					}
				}
			}
		}
	}

	var currentState []*types.Validator
	err := tx.Select(&currentState, "SELECT validatorindex, withdrawableepoch, withdrawalcredentials, slashed, activationeligibilityepoch, activationepoch, exitepoch, lastattestationslot, status FROM validators;")

	if err != nil {
		return err
	}

	currentStateMap := make(map[uint64]*types.Validator, len(currentState))
	latestBlock := uint64(0)

	for _, v := range currentState {
		if uint64(v.LastAttestationSlot.Int64) > latestBlock {
			latestBlock = uint64(v.LastAttestationSlot.Int64)
		}
		currentStateMap[v.Index] = v
	}

	thresholdSlot := latestBlock - 64
	if latestBlock < 64 {
		thresholdSlot = 0
	}

	latestEpoch := latestBlock / utils.Config.Chain.Config.SlotsPerEpoch
	farFutureEpoch := uint64(18446744073709551615)
	maxSqlNumber := uint64(9223372036854775807)

	var queries strings.Builder
	updates := 0
	for _, v := range data.Validators {

		// exchange farFutureEpoch with the corresponding max sql value
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

		if currentStateMap[v.Index] != nil && currentStateMap[v.Index].LastAttestationSlot.Int64 > v.LastAttestationSlot.Int64 {
			v.LastAttestationSlot.Int64 = currentStateMap[v.Index].LastAttestationSlot.Int64
		}

		c := currentStateMap[v.Index]

		if c == nil {
			logger.Infof("validator %v is new", v.Index)

			_, err = tx.Exec(`INSERT INTO validators (
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
				pubkeyhex,
				status,
				lastattestationslot
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);`,
				v.Index,
				v.PublicKey,
				v.WithdrawableEpoch,
				v.WithdrawalCredentials,
				0,
				0,
				v.Slashed,
				v.ActivationEligibilityEpoch,
				v.ActivationEpoch,
				v.ExitEpoch,
				fmt.Sprintf("%x", v.PublicKey),
				v.Status,
				v.LastAttestationSlot,
			)

			if err != nil {
				logger.Errorf("error saving new validator %v: %v", v.Index, err)
			}
		} else {
			// status                     =
			// CASE
			// WHEN EXCLUDED.exitepoch <= %[1]d AND EXCLUDED.slashed THEN 'slashed'
			// WHEN EXCLUDED.exitepoch <= %[1]d THEN 'exited'
			// WHEN EXCLUDED.activationeligibilityepoch = 9223372036854775807 THEN 'deposited'
			// WHEN EXCLUDED.activationepoch > %[1]d THEN 'pending'
			// WHEN EXCLUDED.slashed AND EXCLUDED.activationepoch < %[1]d AND GREATEST(EXCLUDED.lastattestationslot, validators.lastattestationslot) < %[2]d THEN 'slashing_offline'
			// WHEN EXCLUDED.slashed THEN 'slashing_online'
			// WHEN EXCLUDED.exitepoch < 9223372036854775807 AND GREATEST(EXCLUDED.lastattestationslot, validators.lastattestationslot) < %[2]d THEN 'exiting_offline'
			// WHEN EXCLUDED.exitepoch < 9223372036854775807 THEN 'exiting_online'
			// WHEN EXCLUDED.activationepoch < %[1]d AND GREATEST(EXCLUDED.lastattestationslot, validators.lastattestationslot) < %[2]d THEN 'active_offline'
			// ELSE 'active_online'
			// END

			offline := false

			lastSeen := c.LastAttestationSlot.Int64
			if v.LastAttestationSlot.Int64 > lastSeen {
				lastSeen = v.LastAttestationSlot.Int64
			}

			if lastSeen < int64(thresholdSlot) {
				offline = true
			}

			if v.ExitEpoch <= latestEpoch && v.Slashed {
				v.Status = "slashed"
			} else if v.ExitEpoch <= latestEpoch {
				v.Status = "exited"
			} else if v.ActivationEligibilityEpoch == 9223372036854775807 {
				v.Status = "deposited"
			} else if v.ActivationEpoch > latestEpoch {
				v.Status = "pending"
			} else if v.Slashed && v.ActivationEpoch < latestEpoch && offline {
				v.Status = "slashing_offline"
			} else if v.Slashed {
				v.Status = "slashing_online"
			} else if v.ExitEpoch < 9223372036854775807 && offline {
				v.Status = "exiting_offline"
			} else if v.ExitEpoch < 9223372036854775807 {
				v.Status = "exiting_online"
			} else if v.ActivationEpoch < latestEpoch && offline {
				v.Status = "active_offline"
			} else {
				v.Status = "active_online"
			}

			// if c.LastAttestationSlot != v.LastAttestationSlot && v.LastAttestationSlot.Valid {
			// 	// logger.Infof("LastAttestationSlot changed for validator %v from %v to %v", v.Index, c.LastAttestationSlot.Int64, v.LastAttestationSlot.Int64)

			// 	queries.WriteString(fmt.Sprintf("UPDATE validators SET lastattestationslot = %d WHERE validatorindex = %d;\n", v.LastAttestationSlot.Int64, c.Index))
			// 	updates++
			// }

			if c.Status != v.Status {
				logger.Tracef("Status changed for validator %v from %v to %v", v.Index, c.Status, v.Status)
				queries.WriteString(fmt.Sprintf("UPDATE validators SET status = '%s' WHERE validatorindex = %d;\n", v.Status, c.Index))
				updates++
			}
			// if c.Balance != v.Balance {
			// 	// logger.Infof("Balance changed for validator %v from %v to %v", v.Index, c.Balance, v.Balance)
			// 	queries.WriteString(fmt.Sprintf("UPDATE validators SET balance = %d WHERE validatorindex = %d;\n", v.Balance, c.Index))
			// 	updates++
			// }
			// if c.EffectiveBalance != v.EffectiveBalance {
			// 	// logger.Infof("EffectiveBalance changed for validator %v from %v to %v", v.Index, c.EffectiveBalance, v.EffectiveBalance)
			// 	queries.WriteString(fmt.Sprintf("UPDATE validators SET effectivebalance = %d WHERE validatorindex = %d;\n", v.EffectiveBalance, c.Index))
			// 	updates++
			// }
			if c.Slashed != v.Slashed {
				logger.Infof("Slashed changed for validator %v from %v to %v", v.Index, c.Slashed, v.Slashed)
				queries.WriteString(fmt.Sprintf("UPDATE validators SET slashed = %v WHERE validatorindex = %d;\n", v.Slashed, c.Index))
				updates++
			}
			if c.ActivationEligibilityEpoch != v.ActivationEligibilityEpoch {
				logger.Infof("ActivationEligibilityEpoch changed for validator %v from %v to %v", v.Index, c.ActivationEligibilityEpoch, v.ActivationEligibilityEpoch)
				queries.WriteString(fmt.Sprintf("UPDATE validators SET activationeligibilityepoch = %d WHERE validatorindex = %d;\n", v.ActivationEligibilityEpoch, c.Index))
				updates++
			}
			if c.ActivationEpoch != v.ActivationEpoch {
				logger.Infof("ActivationEpoch changed for validator %v from %v to %v", v.Index, c.ActivationEpoch, v.ActivationEpoch)
				queries.WriteString(fmt.Sprintf("UPDATE validators SET activationepoch = %d WHERE validatorindex = %d;\n", v.ActivationEpoch, c.Index))
				updates++
			}
			if c.ExitEpoch != v.ExitEpoch {
				logger.Infof("ExitEpoch changed for validator %v from %v to %v", v.Index, c.ExitEpoch, v.ExitEpoch)
				queries.WriteString(fmt.Sprintf("UPDATE validators SET exitepoch = %d WHERE validatorindex = %d;\n", v.ExitEpoch, c.Index))
				updates++
			}
			if c.WithdrawableEpoch != v.WithdrawableEpoch {
				logger.Infof("WithdrawableEpoch changed for validator %v from %v to %v", v.Index, c.WithdrawableEpoch, v.WithdrawableEpoch)
				queries.WriteString(fmt.Sprintf("UPDATE validators SET withdrawableepoch = %d WHERE validatorindex = %d;\n", v.WithdrawableEpoch, c.Index))
				updates++
			}
			if !bytes.Equal(c.WithdrawalCredentials, v.WithdrawalCredentials) {
				logger.Infof("WithdrawalCredentials changed for validator %v from %x to %x", v.Index, c.WithdrawalCredentials, v.WithdrawalCredentials)
				queries.WriteString(fmt.Sprintf("UPDATE validators SET withdrawalcredentials = '\\x%x' WHERE validatorindex = %d;\n", v.WithdrawalCredentials, c.Index))
				updates++
			}
		}
	}

	if updates > 0 {
		updateStart := time.Now()
		logger.Infof("applying %v update queries", updates)
		_, err = tx.Exec(queries.String())
		if err != nil {
			logger.Errorf("error executing validator update query: %v", err)
			return err
		}
		logger.Infof("update completed, took %v", time.Since(updateStart))
	}

	batchSize := 30000 // max parameters: 65535
	for b := 0; b < len(validators); b += batchSize {
		start := b
		end := b + batchSize
		if len(validators) < end {
			end = len(validators)
		}

		numArgs := 2
		valueStrings := make([]string, 0, batchSize)
		valueArgs := make([]interface{}, 0, batchSize*numArgs)
		for i, v := range validators[start:end] {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d::int, $%d::int)", i*numArgs+1, i*numArgs+2))
			valueArgs = append(valueArgs, v.Index)
			valueArgs = append(valueArgs, v.LastAttestationSlot.Int64)
		}

		stmt := fmt.Sprintf(`
			UPDATE validators AS v SET
			lastattestationslot = GREATEST(v.lastattestationslot, v2.lastattestationslot)
			FROM (VALUES
				%[1]s
			) AS v2(validatorindex, lastattestationslot)
			WHERE v2.validatorindex = v.validatorindex;
	`, strings.Join(valueStrings, ","))

		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			utils.LogError(err, "error executing transaction", 0)
			return err
		}

		logger.Infof("saving validator batch %v completed", b)
	}

	s := time.Now()
	newValidators := []struct {
		Validatorindex  uint64
		ActivationEpoch uint64
	}{}

	err = tx.Select(&newValidators, "SELECT validatorindex, activationepoch FROM validators WHERE balanceactivation IS NULL ORDER BY activationepoch LIMIT 10000")
	if err != nil {
		return err
	}

	balanceCache := make(map[uint64]map[uint64]uint64)
	currentActivationEpoch := uint64(0)

	// get genesis balances of all validators for performance

	for _, newValidator := range newValidators {
		if newValidator.ActivationEpoch > data.Epoch {
			continue
		}

		if newValidator.ActivationEpoch != currentActivationEpoch {
			logger.Infof("removing epoch %v from the activation epoch balance cache", currentActivationEpoch)
			delete(balanceCache, currentActivationEpoch) // remove old items from the map
			currentActivationEpoch = newValidator.ActivationEpoch
		}

		var balance map[uint64][]*types.ValidatorBalance
		if newValidator.ActivationEpoch == 0 {
			balance = genesisBalances
		} else {
			balance, err = BigtableClient.GetValidatorBalanceHistory([]uint64{newValidator.Validatorindex}, newValidator.ActivationEpoch, newValidator.ActivationEpoch)
			if err != nil {
				return err
			}
		}

		foundBalance := uint64(0)
		if balance[newValidator.Validatorindex] == nil || len(balance[newValidator.Validatorindex]) == 0 {
			logger.Errorf("no activation epoch balance found for validator %v for epoch %v in bigtable, trying node", newValidator.Validatorindex, newValidator.ActivationEpoch)

			if balanceCache[newValidator.ActivationEpoch] == nil {
				balances, err := client.GetBalancesForEpoch(int64(newValidator.ActivationEpoch))
				if err != nil {
					return fmt.Errorf("error retrieving balances for epoch %d: %v", newValidator.ActivationEpoch, err)
				}
				balanceCache[newValidator.ActivationEpoch] = balances
			}
			foundBalance = balanceCache[newValidator.ActivationEpoch][newValidator.Validatorindex]
		} else {
			foundBalance = balance[newValidator.Validatorindex][0].Balance
		}

		logger.Infof("retrieved activation epoch balance of %v for validator %v", foundBalance, newValidator.Validatorindex)

		_, err = tx.Exec("update validators set balanceactivation = $1 WHERE validatorindex = $2 AND balanceactivation IS NULL;", foundBalance, newValidator.Validatorindex)
		if err != nil {
			return err
		}
	}

	logger.Infof("updating validator activation epoch balance completed, took %v", time.Since(s))

	s = time.Now()
	_, err = tx.Exec("ANALYZE (SKIP_LOCKED) validators;")
	if err != nil {
		return err
	}
	logger.Infof("analyze of validators table completed, took %v", time.Since(s))

	return nil
}

func saveValidatorProposalAssignments(epoch uint64, assignments map[uint64]uint64, tx *sqlx.Tx) error {
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
			return fmt.Errorf("error executing save validator proposal assignment statement: %w", err)
		}
	}

	return nil
}

func saveValidatorBalancesRecent(epoch uint64, validators []*types.Validator, tx *sqlx.Tx) error {
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
		_, err := tx.Exec("DELETE FROM validator_balances_recent WHERE epoch < $1 AND epoch <> 0", epoch-10)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetRelayDataForIndexedBlocks(blocks []*types.Eth1BlockIndexed) (map[common.Hash]types.RelaysData, error) {
	var execBlockHashes [][]byte
	var relaysData []types.RelaysData

	for _, block := range blocks {
		execBlockHashes = append(execBlockHashes, block.Hash)
	}
	// try to get mev rewards from relys_blocks table
	err := ReaderDb.Select(&relaysData,
		`SELECT proposer_fee_recipient, value, exec_block_hash, tag_id, builder_pubkey FROM relays_blocks WHERE relays_blocks.exec_block_hash = ANY($1)`,
		pq.ByteaArray(execBlockHashes),
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	var relaysDataMap = make(map[common.Hash]types.RelaysData)
	for _, relayData := range relaysData {
		relaysDataMap[common.BytesToHash(relayData.ExecBlockHash)] = relayData
	}

	return relaysDataMap, nil
}

func saveBlocks(blocks map[uint64]map[string]*types.Block, tx *sqlx.Tx) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_save_blocks").Observe(time.Since(start).Seconds())
	}()

	domain, err := utils.GetSigningDomain()
	if err != nil {
		return err
	}

	stmtBlock, err := tx.Prepare(`
		INSERT INTO blocks (epoch, slot, blockroot, parentroot, stateroot, signature, randaoreveal, graffiti, graffiti_text, eth1data_depositroot, eth1data_depositcount, eth1data_blockhash, syncaggregate_bits, syncaggregate_signature, proposerslashingscount, attesterslashingscount, attestationscount, depositscount, withdrawalcount, voluntaryexitscount, syncaggregate_participation, proposer, status, exec_parent_hash, exec_fee_recipient, exec_state_root, exec_receipts_root, exec_logs_bloom, exec_random, exec_block_number, exec_gas_limit, exec_gas_used, exec_timestamp, exec_extra_data, exec_base_fee_per_gas, exec_block_hash, exec_transactions_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37)
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

	stmtWithdrawals, err := tx.Prepare(`
	INSERT INTO blocks_withdrawals (block_slot, block_root, withdrawalindex, validatorindex, address, amount)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (block_slot, block_root, withdrawalindex) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtWithdrawals.Close()

	stmtBLSChange, err := tx.Prepare(`
	INSERT INTO blocks_bls_change (block_slot, block_root, validatorindex, signature, pubkey, address)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (block_slot, block_root, validatorindex) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtBLSChange.Close()

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
		INSERT INTO blocks_deposits (block_slot, block_index, block_root, proof, publickey, withdrawalcredentials, amount, signature, valid_signature)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
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
			if err == nil && bytes.Equal(dbBlockRootHash, b.BlockRoot) {
				blockLog.Infof("skipping export of block as it is already present in the db")
				continue
			} else if err != nil && err != sql.ErrNoRows {
				return fmt.Errorf("error checking for block in db: %w", err)
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
			withdrawalCount := 0
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
				withdrawalCount = len(b.ExecutionPayload.Withdrawals)
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
				withdrawalCount,
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

			n := time.Now()
			logger.Tracef("done, took %v", time.Since(n))
			logger.Tracef("writing transactions and withdrawal data")
			if payload := b.ExecutionPayload; payload != nil {
				for i, tx := range payload.Transactions {
					_, err := stmtTransaction.Exec(b.Slot, i, b.BlockRoot,
						tx.Raw, tx.TxHash, tx.AccountNonce, tx.Price, tx.GasLimit, tx.Sender, tx.Recipient, tx.Amount, tx.Payload, tx.MaxPriorityFeePerGas, tx.MaxFeePerGas)
					if err != nil {
						return fmt.Errorf("error executing stmtTransaction for block %v: %v", b.Slot, err)
					}
				}
				for _, w := range payload.Withdrawals {
					_, err := stmtWithdrawals.Exec(b.Slot, b.BlockRoot, w.Index, w.ValidatorIndex, w.Address, w.Amount)
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
			blockLog.WithField("duration", time.Since(n)).Tracef("stmtProposerSlashing")

			n = time.Now()
			logger.Tracef("writing bls change data")
			for _, bls := range b.SignedBLSToExecutionChange {
				_, err := stmtBLSChange.Exec(b.Slot, b.BlockRoot, bls.Message.Validatorindex, bls.Signature, bls.Message.BlsPubkey, bls.Message.Address)
				if err != nil {
					return fmt.Errorf("error executing stmtBLSChange for block %v: %w", b.Slot, err)
				}
			}
			blockLog.WithField("duration", time.Since(n)).Tracef("stmtBLSChange")
			t = time.Now()

			for i, as := range b.AttesterSlashings {
				_, err := stmtAttesterSlashing.Exec(b.Slot, i, b.BlockRoot, pq.Array(as.Attestation1.AttestingIndices), as.Attestation1.Signature, as.Attestation1.Data.Slot, as.Attestation1.Data.CommitteeIndex, as.Attestation1.Data.BeaconBlockRoot, as.Attestation1.Data.Source.Epoch, as.Attestation1.Data.Source.Root, as.Attestation1.Data.Target.Epoch, as.Attestation1.Data.Target.Root, pq.Array(as.Attestation2.AttestingIndices), as.Attestation2.Signature, as.Attestation2.Data.Slot, as.Attestation2.Data.CommitteeIndex, as.Attestation2.Data.BeaconBlockRoot, as.Attestation2.Data.Source.Epoch, as.Attestation2.Data.Source.Root, as.Attestation2.Data.Target.Epoch, as.Attestation2.Data.Target.Root)
				if err != nil {
					return fmt.Errorf("error executing stmtAttesterSlashing for block %v: %w", b.Slot, err)
				}
			}
			blockLog.WithField("duration", time.Since(t)).Tracef("stmtAttesterSlashing")
			t = time.Now()
			for i, a := range b.Attestations {
				_, err = stmtAttestations.Exec(b.Slot, i, b.BlockRoot, a.AggregationBits, pq.Array(a.Attesters), a.Signature, a.Data.Slot, a.Data.CommitteeIndex, a.Data.BeaconBlockRoot, a.Data.Source.Epoch, a.Data.Source.Root, a.Data.Target.Epoch, a.Data.Target.Root)
				if err != nil {
					return fmt.Errorf("error executing stmtAttestations for block %v: %w", b.Slot, err)
				}
			}
			blockLog.WithField("duration", time.Since(t)).Tracef("attestations")
			t = time.Now()

			for i, d := range b.Deposits {

				err := prysm_deposit.VerifyDepositSignature(&ethpb.Deposit_Data{
					PublicKey:             d.PublicKey,
					WithdrawalCredentials: d.WithdrawalCredentials,
					Amount:                d.Amount,
					Signature:             d.Signature,
				}, domain)

				signatureValid := err == nil

				_, err = stmtDeposits.Exec(b.Slot, i, b.BlockRoot, nil, d.PublicKey, d.WithdrawalCredentials, d.Amount, d.Signature, signatureValid)
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

			blockLog.Infof("! export of block completed, took %v", time.Since(start))
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
			eligibleether = $1,
			globalparticipationrate = $2,
			votedether = $3
		WHERE epoch = $4`,
		stats.EligibleEther, stats.GlobalParticipationRate, stats.VotedEther, stats.Epoch)

	return err
}

// UpdateEpochFinalization will update finalized-flag of unfinalized epochs
func UpdateEpochFinalization(finality_epoch uint64) error {
	// to prevent a full table scan, the query is constrained to update only between the last epoch that was tagged finalized and the passed finality_epoch
	// will not fill gaps in the db in finalization this way, but makes the query much faster.
	_, err := WriterDb.Exec(`
	UPDATE epochs
	SET finalized = true
	WHERE epoch BETWEEN COALESCE((
			SELECT epoch
			FROM   epochs
			WHERE  finalized = true
			ORDER  BY epoch DESC
			LIMIT  1
		),0) AND $1`, finality_epoch)
	return err
}

// GetTotalValidatorsCount will return the total-validator-count
func GetTotalValidatorsCount() (uint64, error) {
	var totalCount uint64
	err := ReaderDb.Get(&totalCount, "select coalesce(max(validatorindex) + 1, 0) from validators;")
	return totalCount, err
}

// GetActiveValidatorCount will return the total-validator-count
func GetActiveValidatorCount() (uint64, error) {
	var count uint64
	err := ReaderDb.Get(&count, "select count(*) from validators where status in ('active_offline', 'active_online');")
	return count, err
}

func updateQueueDeposits() error {
	start := time.Now()
	defer func() {
		logger.Infof("took %v seconds to update queue deposits", time.Since(start).Seconds())
		metrics.TaskDuration.WithLabelValues("update_queue_deposits").Observe(time.Since(start).Seconds())
	}()

	// first we remove any validator that isn't queued anymore
	_, err := WriterDb.Exec(`
		DELETE FROM validator_queue_deposits
		WHERE validator_queue_deposits.validatorindex NOT IN (
			SELECT validatorindex 
			FROM validators 
			WHERE activationepoch=9223372036854775807 and status='pending')`)
	if err != nil {
		logger.Errorf("error removing queued publickeys from validator_queue_deposits: %v", err)
		return err
	}

	// then we add any new ones that are queued
	_, err = WriterDb.Exec(`
		INSERT INTO validator_queue_deposits
		SELECT validatorindex FROM validators WHERE activationepoch=9223372036854775807 and status='pending' ON CONFLICT DO NOTHING`)
	if err != nil {
		logger.Errorf("error adding queued publickeys to validator_queue_deposits: %v", err)
		return err
	}

	// efficiently collect the tnx that pushed each validator over 32 ETH.
	_, err = WriterDb.Exec(`
		UPDATE validator_queue_deposits 
		SET 
			block_slot=data.block_slot,
			block_index=data.block_index
		FROM (
			WITH CumSum AS
			(
				SELECT publickey, block_slot, block_index,
					/* generate partion per publickey ordered by newest to oldest. store cum sum of deposits */
					SUM(amount) OVER (partition BY publickey ORDER BY (block_slot, block_index) ASC) AS cumTotal
				FROM blocks_deposits
				WHERE publickey IN (
					/* get the pubkeys of the indexes */
					select pubkey from validators where validators.validatorindex in (
						/* get the indexes we need to update */
						select validatorindex from validator_queue_deposits where block_slot is null and block_index is null
					)
				)
				ORDER BY block_slot, block_index ASC
			)
			/* we only care about one deposit per vali */
			SELECT DISTINCT ON(publickey) validators.validatorindex, block_slot, block_index
			FROM CumSum
			/* join so we can retrieve the validator index again */
			left join validators on validators.pubkey = CumSum.publickey
			/* we want the deposit that pushed the cum sum over 32 ETH */
			WHERE cumTotal>=32000000000
			ORDER BY publickey, cumTotal asc 
		) AS data
		WHERE validator_queue_deposits.validatorindex=data.validatorindex`)
	if err != nil {
		logger.Errorf("error updating validator_queue_deposits: %v", err)
		return err
	}
	return nil
}

func GetQueueAheadOfValidator(validatorIndex uint64) (uint64, error) {
	var res uint64
	err := ReaderDb.Get(&res, `
	with SelectedValidator as (
		select block_slot, block_index, validators.activationeligibilityepoch from validator_queue_deposits vqd
		inner join validators on validators.validatorindex = $1
		where vqd.validatorindex = validators.validatorindex 
	)
	select count(*)
	from (
		select validatorindex, block_slot, block_index 
		from validator_queue_deposits
	) as vqd 
	inner join validators on
		validators.validatorindex = vqd.validatorindex and
		validators.activationeligibilityepoch <= (select activationeligibilityepoch from SelectedValidator)
	where 
		validators.activationeligibilityepoch < (select activationeligibilityepoch from SelectedValidator) or 
		vqd.block_slot < (select block_slot from SelectedValidator) or
		vqd.block_slot = (select block_slot from SelectedValidator) and vqd.block_index < (select block_index from SelectedValidator)`, validatorIndex)
	return res, err
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
		return 0, fmt.Errorf("error retrieving validator queue count: %w", err)
	}
	return count, nil
}

func GetTotalEligibleEther() (uint64, error) {
	var total uint64

	err := ReaderDb.Get(&total, `
		SELECT eligibleether FROM epochs ORDER BY epoch DESC LIMIT 1
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

func GetSlotVizData(latestEpoch uint64) ([]*types.SlotVizEpochs, error) {
	type sqlBlocks struct {
		Slot                    uint64
		BlockRoot               []byte
		Epoch                   uint64
		Status                  string
		Globalparticipationrate float64
		Finalized               bool
		Justified               bool
		Previousjustified       bool
	}

	var blks []sqlBlocks = []sqlBlocks{}
	if latestEpoch > 4 {
		latestEpoch = latestEpoch - 4
	} else {
		latestEpoch = 0
	}

	err := ReaderDb.Select(&blks, `
	SELECT
		b.slot,
		b.blockroot,
		CASE
			WHEN b.status = '0' THEN 'scheduled'
			WHEN b.status = '1' THEN 'proposed'
			WHEN b.status = '2' THEN 'missed'
			WHEN b.status = '3' THEN 'orphaned'
			ELSE 'unknown'
		END AS status,
		b.epoch,
		COALESCE(e.globalparticipationrate, 0) AS globalparticipationrate,
		COALESCE(e.finalized, false) AS finalized
	FROM blocks b
		LEFT JOIN epochs e ON e.epoch = b.epoch
	WHERE b.epoch >= $1
	ORDER BY slot DESC;
`, latestEpoch)
	if err != nil {
		return nil, err
	}

	currentSlot := utils.TimeToSlot(uint64(time.Now().Unix()))

	epochMap := map[uint64]*types.SlotVizEpochs{}

	res := []*types.SlotVizEpochs{}

	for _, b := range blks {
		if b.Globalparticipationrate == 1 && !b.Finalized {
			b.Globalparticipationrate = 0
		}
		_, exists := epochMap[b.Epoch]
		if !exists {
			r := types.SlotVizEpochs{
				Epoch:          b.Epoch,
				Finalized:      b.Finalized,
				Particicpation: b.Globalparticipationrate,
				Slots:          []*types.SlotVizSlots{},
			}
			r.Slots = make([]*types.SlotVizSlots, utils.Config.Chain.Config.SlotsPerEpoch)
			epochMap[b.Epoch] = &r
		}

		slotIndex := b.Slot - (b.Epoch * utils.Config.Chain.Config.SlotsPerEpoch)

		// if epochMap[b.Epoch].Slots[slotIndex] != nil && len(b.BlockRoot) > len(epochMap[b.Epoch].Slots[slotIndex].BlockRoot) {
		// 	logger.Infof("CONFLICTING block found for slotindex %v", slotIndex)
		// }

		if epochMap[b.Epoch].Slots[slotIndex] == nil || len(b.BlockRoot) > len(epochMap[b.Epoch].Slots[slotIndex].BlockRoot) {
			epochMap[b.Epoch].Slots[slotIndex] = &types.SlotVizSlots{
				Epoch:     b.Epoch,
				Slot:      b.Slot,
				Status:    b.Status,
				Active:    b.Slot == currentSlot,
				BlockRoot: b.BlockRoot,
			}
		}

	}

	for _, epoch := range epochMap {
		for i := uint64(0); i < utils.Config.Chain.Config.SlotsPerEpoch; i++ {
			if epoch.Slots[i] == nil {
				status := "scheduled"
				slot := (epoch.Epoch * utils.Config.Chain.Config.SlotsPerEpoch) + i
				if slot < currentSlot {
					status = "scheduled-missed"
				}
				// logger.Infof("FILLING MISSING SLOT: %v", slot)
				epoch.Slots[i] = &types.SlotVizSlots{
					Epoch:  epoch.Epoch,
					Slot:   slot,
					Status: status,
					Active: slot == currentSlot,
				}
			}
		}
	}

	for _, epoch := range epochMap {
		for _, slot := range epoch.Slots {
			slot.Active = slot.Slot == currentSlot
			if slot.Status != "proposed" && slot.Status != "missed" {
				if slot.Status == "scheduled" && slot.Slot < currentSlot {
					slot.Status = "scheduled-missed"
				}

				if slot.Slot >= currentSlot {
					slot.Status = "scheduled"
				}
			}
		}
		if epoch.Finalized {
			epoch.Justified = true
		}
		res = append(res, epoch)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Epoch > res[j].Epoch
	})

	for i := 0; i < len(res); i++ {
		if !res[i].Finalized && i != 0 {
			res[i-1].Justifying = true
		}
		if res[i].Finalized && i != 0 {
			res[i-1].Justified = true
			break
		}
	}

	return res, nil
}

func GetBlockNumber(slot uint64) (block uint64, err error) {
	err = ReaderDb.Get(&block, `SELECT exec_block_number FROM blocks where slot = $1`, slot)
	return
}

func SaveChartSeriesPoint(date time.Time, indicator string, value any) error {
	_, err := WriterDb.Exec(`INSERT INTO chart_series (time, indicator, value) VALUES($1, $2, $3) ON CONFLICT (time, indicator) DO UPDATE SET value = EXCLUDED.value`, date, indicator, value)
	if err != nil {
		return fmt.Errorf("error calculating NON_FAILED_TX_GAS_USAGE chart_series: %w", err)
	}
	return err
}

func GetSlotWithdrawals(slot uint64) ([]*types.Withdrawals, error) {
	var withdrawals []*types.Withdrawals

	err := ReaderDb.Select(&withdrawals, `
		SELECT
			w.withdrawalindex as index,
			w.validatorindex,
			w.address,
			w.amount
		FROM
			blocks_withdrawals w
		LEFT JOIN blocks b ON b.blockroot = w.block_root
		WHERE w.block_slot = $1 AND b.status = '1'
		ORDER BY w.withdrawalindex
	`, slot)
	if err != nil {
		if err == sql.ErrNoRows {
			return withdrawals, nil
		}
		return nil, fmt.Errorf("error getting blocks_withdrawals for slot: %d: %w", slot, err)
	}

	return withdrawals, nil
}

func GetTotalWithdrawals() (total uint64, err error) {
	err = ReaderDb.Get(&total, `
	SELECT
		COALESCE(MAX(withdrawalindex), 0)
	FROM 
		blocks_withdrawals`)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return
}

func GetWithdrawals(query string, length, start uint64, orderBy, orderDir string) ([]*types.Withdrawals, error) {
	withdrawals := []*types.Withdrawals{}

	if orderDir != "desc" && orderDir != "asc" {
		orderDir = "desc"
	}
	columns := []string{"block_slot", "withdrawalindex", "validatorindex", "address", "amount"}
	hasColumn := false
	for _, column := range columns {
		if orderBy == column {
			hasColumn = true
			break
		}
	}
	if !hasColumn {
		orderBy = "block_slot"
	}

	if query != "" {
		bquery, _ := hex.DecodeString(strings.TrimPrefix(query, "0x"))

		err := ReaderDb.Select(&withdrawals, fmt.Sprintf(`
			SELECT 
				w.block_slot as slot,
				w.withdrawalindex as index,
				w.validatorindex,
				w.address,
				w.amount
			FROM blocks_withdrawals w
			INNER JOIN blocks b ON w.block_root = b.blockroot AND b.status = '1'
			WHERE CAST(w.validatorindex as varchar) LIKE $3 || '%%'
				OR address LIKE $4 || '%%'::bytea
				OR CAST(block_slot as varchar) LIKE $3 || '%%'
				OR CAST(block_slot / $5 as varchar) LIKE $3 || '%%'
			ORDER BY %s %s
			LIMIT $1
			OFFSET $2`, orderBy, orderDir), length, start, strings.ToLower(query), bquery, utils.Config.Chain.Config.SlotsPerEpoch)
		if err != nil {
			return nil, err
		}
	} else {
		err := ReaderDb.Select(&withdrawals, fmt.Sprintf(`
			SELECT 
				w.block_slot as slot,
				w.withdrawalindex as index,
				w.validatorindex,
				w.address,
				w.amount
			FROM blocks_withdrawals w
			INNER JOIN blocks b ON w.block_root = b.blockroot AND b.status = '1'
			ORDER BY %s %s
			LIMIT $1
			OFFSET $2`, orderBy, orderDir), length, start)
		if err != nil {
			return nil, err
		}
	}

	return withdrawals, nil
}

func GetTotalAmountWithdrawn() (sum uint64, count uint64, err error) {
	var res = struct {
		Sum   uint64 `db:"sum"`
		Count uint64 `db:"count"`
	}{}
	err = ReaderDb.Get(&res, `
	SELECT 
		COALESCE(sum(w.amount), 0) as sum,
		COALESCE(count(*), 0) as count
	FROM blocks_withdrawals w
	INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'`)
	return res.Sum, res.Count, err
}

func GetTotalAmountDeposited() (uint64, error) {
	var total uint64
	err := ReaderDb.Get(&total, `
	SELECT 
		COALESCE(sum(d.amount), 0) as sum 
	FROM blocks_deposits d
	INNER JOIN blocks b ON b.blockroot = d.block_root AND b.status = '1'`)
	return total, err
}

func GetBLSChangeCount() (uint64, error) {
	var total uint64
	err := ReaderDb.Get(&total, `
	SELECT 
		COALESCE(count(*), 0) as count
	FROM blocks_bls_change bls
	INNER JOIN blocks b ON b.blockroot = bls.block_root AND b.status = '1'`)
	return total, err
}

func GetEpochWithdrawalsTotal(epoch uint64) (total uint64, err error) {
	err = ReaderDb.Get(&total, `
	SELECT 
		COALESCE(sum(w.amount), 0) as sum 
	FROM blocks_withdrawals w
	INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
	WHERE w.block_slot >= $1 AND w.block_slot < $2`, epoch*utils.Config.Chain.Config.SlotsPerEpoch, (epoch+1)*utils.Config.Chain.Config.SlotsPerEpoch)
	return
}

// GetAddressWithdrawals returns the withdrawals for an address
func GetAddressWithdrawals(address []byte, limit uint64, offset uint64) ([]*types.Withdrawals, error) {
	var withdrawals []*types.Withdrawals
	if limit == 0 {
		limit = 100
	}

	err := ReaderDb.Select(&withdrawals, `
	SELECT 
		w.block_slot as slot, 
		w.withdrawalindex as index, 
		w.validatorindex, 
		w.address, 
		w.amount 
	FROM blocks_withdrawals w
	INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
	WHERE w.address = $1 ORDER BY w.withdrawalindex DESC LIMIT $2 OFFSET $3`, address, limit, offset)
	if err != nil {
		if err == sql.ErrNoRows {
			return withdrawals, nil
		}
		return nil, fmt.Errorf("error getting blocks_withdrawals for address: %x: %w", address, err)
	}

	return withdrawals, nil
}

func GetEpochWithdrawals(epoch uint64) ([]*types.WithdrawalsNotification, error) {
	var withdrawals []*types.WithdrawalsNotification

	err := ReaderDb.Select(&withdrawals, `
	SELECT 
		w.block_slot as slot, 
		w.withdrawalindex as index, 
		w.validatorindex, 
		w.address, 
		w.amount,
		v.pubkey as pubkey
	FROM blocks_withdrawals w
	INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
	LEFT JOIN validators v on v.validatorindex = w.validatorindex
	WHERE w.block_slot >= $1 AND w.block_slot < $2 ORDER BY w.withdrawalindex`, epoch*utils.Config.Chain.Config.SlotsPerEpoch, (epoch+1)*utils.Config.Chain.Config.SlotsPerEpoch)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting blocks_withdrawals for epoch: %d: %w", epoch, err)
	}

	return withdrawals, nil
}

func GetValidatorWithdrawals(validator uint64, limit uint64, offset uint64, orderBy string, orderDir string) ([]*types.Withdrawals, error) {
	var withdrawals []*types.Withdrawals
	if limit == 0 {
		limit = 100
	}

	err := ReaderDb.Select(&withdrawals, fmt.Sprintf(`
	SELECT 
		w.block_slot as slot, 
		w.withdrawalindex as index, 
		w.validatorindex, 
		w.address, 
		w.amount 
	FROM blocks_withdrawals w
	INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
	WHERE validatorindex = $1 
	ORDER BY  w.%s %s 
	LIMIT $2 OFFSET $3`, orderBy, orderDir), validator, limit, offset)
	if err != nil {
		if err == sql.ErrNoRows {
			return withdrawals, nil
		}
		return nil, fmt.Errorf("error getting blocks_withdrawals for validator: %d: %w", validator, err)
	}

	return withdrawals, nil
}

func GetValidatorsWithdrawals(validators []uint64, fromEpoch uint64, toEpoch uint64) ([]*types.Withdrawals, error) {
	var withdrawals []*types.Withdrawals

	err := ReaderDb.Select(&withdrawals, `
	SELECT 
		w.block_slot as slot, 
		w.withdrawalindex as index, 
		w.block_root as blockroot,
		w.validatorindex, 
		w.address, 
		w.amount 
	FROM blocks_withdrawals w
	INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
	WHERE validatorindex = ANY($1)
	AND (w.block_slot / $4) >= $2 AND (w.block_slot / $4) <= $3 
	ORDER BY w.withdrawalindex`, pq.Array(validators), fromEpoch, toEpoch, utils.Config.Chain.Config.SlotsPerEpoch)
	if err != nil {
		if err == sql.ErrNoRows {
			return withdrawals, nil
		}
		return nil, fmt.Errorf("error getting blocks_withdrawals for validators: %+v: %w", validators, err)
	}

	return withdrawals, nil
}

func GetValidatorsWithdrawalsByEpoch(validator []uint64, startEpoch uint64, endEpoch uint64) ([]*types.WithdrawalsByEpoch, error) {
	if startEpoch > endEpoch {
		startEpoch = 0
	}

	var withdrawals []*types.WithdrawalsByEpoch

	err := ReaderDb.Select(&withdrawals, `
	SELECT 
		w.validatorindex,
		w.block_slot / $4 as epoch, 
		sum(w.amount) as amount
	FROM blocks_withdrawals w
	INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1' AND b.slot >= $2 AND b.slot <= $3
	WHERE validatorindex = ANY($1) 
	GROUP BY w.validatorindex, w.block_slot / $4
	ORDER BY w.block_slot / $4 DESC LIMIT 100`, pq.Array(validator), startEpoch*utils.Config.Chain.Config.SlotsPerEpoch, endEpoch*utils.Config.Chain.Config.SlotsPerEpoch+utils.Config.Chain.Config.SlotsPerEpoch-1, utils.Config.Chain.Config.SlotsPerEpoch)
	if err != nil {
		if err == sql.ErrNoRows {
			return withdrawals, nil
		}
		return nil, fmt.Errorf("error getting blocks_withdrawals for validator: %d: %w", validator, err)
	}
	return withdrawals, nil
}

// GetAddressWithdrawalsTotal returns the total withdrawals for an address
func GetAddressWithdrawalsTotal(address []byte) (uint64, error) {
	var total uint64

	err := ReaderDb.Get(&total, `
	SELECT 
		COALESCE(sum(w.amount), 0) as total
	FROM blocks_withdrawals w
	INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
	WHERE w.address = $1`, address)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("error getting blocks_withdrawals for address: %x: %w", address, err)
	}

	return total, nil
}

func GetDashboardWithdrawalsCount(validators []uint64) (uint64, error) {
	var count uint64
	validatorFilter := pq.Array(validators)
	err := ReaderDb.Get(&count, `
	SELECT count(*) 
	FROM blocks_withdrawals w
	INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
	WHERE w.validatorindex = Any($1)`, validatorFilter)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("error getting dashboard validator blocks_withdrawals count for validators: %d: %w", validators, err)
	}

	return count, nil
}

func GetDashboardWithdrawals(validators []uint64, limit uint64, offset uint64, orderBy string, orderDir string) ([]*types.Withdrawals, error) {
	var withdrawals []*types.Withdrawals
	if limit == 0 {
		limit = 100
	}
	validatorFilter := pq.Array(validators)
	err := ReaderDb.Select(&withdrawals, fmt.Sprintf(`
		SELECT 
			w.block_slot as slot, 
			w.withdrawalindex as index, 
			w.validatorindex, 
			w.address, 
			w.amount 
		FROM blocks_withdrawals w
		INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
		WHERE validatorindex = ANY($1)
		ORDER BY  w.%s %s 
		LIMIT $2 OFFSET $3`, orderBy, orderDir), validatorFilter, limit, offset)
	if err != nil {
		if err == sql.ErrNoRows {
			return withdrawals, nil
		}
		return nil, fmt.Errorf("error getting dashboard blocks_withdrawals for validators: %d: %w", validators, err)
	}

	return withdrawals, nil
}

func GetValidatorWithdrawalsCount(validator uint64) (count, lastWithdrawalEpoch uint64, err error) {

	type dbResponse struct {
		Count              uint64 `db:"withdrawals_count"`
		LastWithdrawalSlot uint64 `db:"last_withdawal_slot"`
	}

	r := &dbResponse{}
	err = ReaderDb.Get(r, `
	SELECT count(*) as withdrawals_count, COALESCE(max(block_slot), 0) as last_withdawal_slot
	FROM blocks_withdrawals w
	INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
	WHERE w.validatorindex = $1`, validator)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, nil
		}
		return 0, 0, fmt.Errorf("error getting validator blocks_withdrawals count for validator: %d: %w", validator, err)
	}

	return r.Count, r.LastWithdrawalSlot / utils.Config.Chain.Config.SlotsPerEpoch, nil
}

func GetLastWithdrawalEpoch(validators []uint64) (map[uint64]uint64, error) {

	type dbResponse struct {
		ValidatorIndex     uint64 `db:"validatorindex"`
		LastWithdrawalSlot uint64 `db:"last_withdawal_slot"`
	}

	res := make(map[uint64]uint64)

	r := make([]*dbResponse, 0)
	err := ReaderDb.Get(r, `
	SELECT w.validatorindex, COALESCE(max(block_slot), 0) as last_withdawal_slot
	FROM blocks_withdrawals w
	INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
	WHERE w.validatorindex = ANY($1)
	GROUP BY w.validatorindex`, validators)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, nil
		}
		return nil, fmt.Errorf("error getting validator blocks_withdrawals count for validators: %d: %w", validators, err)
	}

	for _, row := range r {
		res[row.ValidatorIndex] = row.LastWithdrawalSlot / utils.Config.Chain.Config.SlotsPerEpoch
	}

	return res, nil
}

func GetMostRecentWithdrawalValidator() (uint64, error) {
	var validatorindex uint64

	err := ReaderDb.Get(&validatorindex, `
	SELECT 
		w.validatorindex 
	FROM 
		blocks_withdrawals w
	INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
	ORDER BY 
		withdrawalindex DESC LIMIT 1;`)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("error getting most recent blocks_withdrawals validatorindex: %w", err)
	}

	return validatorindex, nil
}

// get all ad configurations
func GetAdConfigurations() ([]*types.AdConfig, error) {
	var adConfigs []*types.AdConfig

	err := ReaderDb.Select(&adConfigs, `
	SELECT 
		id, 
		template_id, 
		jquery_selector, 
		insert_mode, 
		refresh_interval, 
		enabled, 
		for_all_users,
		banner_id, 
		html_content
	FROM 
		ad_configurations`)
	if err != nil {
		if err == sql.ErrNoRows {
			return []*types.AdConfig{}, nil
		}
		return nil, fmt.Errorf("error getting ad configurations: %w", err)
	}

	return adConfigs, nil
}

// get the ad configuration for a specific template that are active
func GetAdConfigurationsForTemplate(ids []string, noAds bool) ([]*types.AdConfig, error) {
	var adConfigs []*types.AdConfig
	forAllUsers := ""
	if noAds {
		forAllUsers = " AND for_all_users = true"
	}
	err := ReaderDb.Select(&adConfigs, fmt.Sprintf(`
	SELECT 
		id, 
		template_id, 
		jquery_selector, 
		insert_mode, 
		refresh_interval, 
		enabled, 
		for_all_users,
		banner_id, 
		html_content
	FROM 
		ad_configurations
	WHERE 
		template_id = ANY($1) AND
		enabled = true %v`, forAllUsers), pq.Array(ids))
	if err != nil {
		if err == sql.ErrNoRows {
			return []*types.AdConfig{}, nil
		}
		return nil, fmt.Errorf("error getting ad configurations for template: %v %s", err, ids)
	}

	return adConfigs, nil
}

// insert new ad configuration
func InsertAdConfigurations(adConfig types.AdConfig) error {
	_, err := WriterDb.Exec(`
		INSERT INTO ad_configurations (
			id, 
			template_id, 
			jquery_selector,
			insert_mode,
			refresh_interval, 
			enabled,
			for_all_users,
			banner_id,
			html_content) 
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9) 
		ON CONFLICT DO NOTHING`,
		adConfig.Id,
		adConfig.TemplateId,
		adConfig.JQuerySelector,
		adConfig.InsertMode,
		adConfig.RefreshInterval,
		adConfig.Enabled,
		adConfig.ForAllUsers,
		adConfig.BannerId,
		adConfig.HtmlContent)
	if err != nil {
		return fmt.Errorf("error inserting ad configuration: %w", err)
	}
	return nil
}

// update exisiting ad configuration
func UpdateAdConfiguration(adConfig types.AdConfig) error {
	tx, err := WriterDb.Begin()
	if err != nil {
		return fmt.Errorf("error starting db transactions: %w", err)
	}
	defer tx.Rollback()
	_, err = tx.Exec(`
		UPDATE ad_configurations SET 
			template_id = $2,
			jquery_selector = $3,
			insert_mode = $4,
			refresh_interval = $5,
			enabled = $6,
			for_all_users = $7,
			banner_id = $8,
			html_content = $9
		WHERE id = $1;`,
		adConfig.Id,
		adConfig.TemplateId,
		adConfig.JQuerySelector,
		adConfig.InsertMode,
		adConfig.RefreshInterval,
		adConfig.Enabled,
		adConfig.ForAllUsers,
		adConfig.BannerId,
		adConfig.HtmlContent)
	if err != nil {
		return fmt.Errorf("error updating ad configuration: %w", err)
	}
	return tx.Commit()
}

// delete ad configuration
func DeleteAdConfiguration(id string) error {

	tx, err := WriterDb.Beginx()
	if err != nil {
		return fmt.Errorf("error starting db transactions: %w", err)
	}
	defer tx.Rollback()

	// delete ad configuration
	_, err = WriterDb.Exec(`
		DELETE FROM ad_configurations 
		WHERE 
			id = $1;`,
		id)
	return err
}

// get all explorer configurations
func GetExplorerConfigurations() ([]*types.ExplorerConfig, error) {
	var configs []*types.ExplorerConfig

	err := ReaderDb.Select(&configs, `
	SELECT 
		category, 
		key, 
		value, 
		data_type 
	FROM 
		explorer_configurations`)
	if err != nil {
		if err == sql.ErrNoRows {
			return []*types.ExplorerConfig{}, nil
		}
		return nil, fmt.Errorf("error getting explorer configurations: %w", err)
	}

	return configs, nil
}

// save current configurations
func SaveExplorerConfiguration(configs []types.ExplorerConfig) error {
	valueStrings := []string{}
	valueArgs := []interface{}{}
	for i, config := range configs {
		valueStrings = append(valueStrings, fmt.Sprintf("($%v, $%v, $%v, $%v)", i*4+1, i*4+2, i*4+3, i*4+4))

		valueArgs = append(valueArgs, config.Category)
		valueArgs = append(valueArgs, config.Key)
		valueArgs = append(valueArgs, config.Value)
		valueArgs = append(valueArgs, config.DataType)
	}
	query := fmt.Sprintf(`
		INSERT INTO explorer_configurations (
			category, 
			key, 
			value, 
			data_type)
    	VALUES %s 
		ON CONFLICT 
			(category, key) 
		DO UPDATE SET 
			value = excluded.value,
			data_type = excluded.data_type
			`, strings.Join(valueStrings, ","))

	_, err := WriterDb.Exec(query, valueArgs...)
	if err != nil {
		return fmt.Errorf("error inserting/updating explorer configurations: %w", err)
	}
	return nil
}

func GetTotalBLSChanges() (uint64, error) {
	var count uint64
	err := ReaderDb.Get(&count, `
		SELECT count(*) FROM blocks_bls_change`)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("error getting total blocks_bls_change: %w", err)
	}

	return count, nil
}

func GetBLSChanges(query string, length, start uint64, orderBy, orderDir string) ([]*types.BLSChange, error) {
	blsChange := []*types.BLSChange{}

	if orderDir != "desc" && orderDir != "asc" {
		orderDir = "desc"
	}
	columns := []string{"block_slot", "validatorindex"}
	hasColumn := false
	for _, column := range columns {
		if orderBy == column {
			hasColumn = true
			break
		}
	}
	if !hasColumn {
		orderBy = "block_slot"
	}

	if query != "" {

		bquery, _ := hex.DecodeString(strings.TrimPrefix(query, "0x"))
		err := ReaderDb.Select(&blsChange, fmt.Sprintf(`
			SELECT 
				bls.block_slot as slot,
				bls.validatorindex,
				bls.signature,
				bls.pubkey,
				bls.address
			FROM blocks_bls_change bls
			INNER JOIN blocks b ON bls.block_root = b.blockroot AND b.status = '1'
			WHERE CAST(bls.validatorindex as varchar) LIKE $3 || '%%'
				OR pubkey LIKE $4::bytea || '%%'::bytea
				OR CAST(block_slot as varchar) LIKE $3 || '%%'
				OR CAST((block_slot / $5) as varchar) LIKE $3 || '%%'
			ORDER BY bls.%s %s
			LIMIT $1
			OFFSET $2`, orderBy, orderDir), length, start, strings.ToLower(query), bquery, utils.Config.Chain.Config.SlotsPerEpoch)
		if err != nil {
			return nil, err
		}
	} else {
		err := ReaderDb.Select(&blsChange, fmt.Sprintf(`
			SELECT 
				bls.block_slot as slot,
				bls.validatorindex,
				bls.signature,
				bls.pubkey,
				bls.address
			FROM blocks_bls_change bls
			INNER JOIN blocks b ON bls.block_root = b.blockroot AND b.status = '1'
			ORDER BY %s %s
			LIMIT $1
			OFFSET $2`, orderBy, orderDir), length, start)
		if err != nil {
			return nil, err
		}
	}

	return blsChange, nil
}

func GetSlotBLSChange(slot uint64) ([]*types.BLSChange, error) {
	var change []*types.BLSChange

	err := ReaderDb.Select(&change, `
	SELECT 
		bls.validatorindex, 
		bls.signature, 
		bls.pubkey, 
		bls.address 
	FROM blocks_bls_change bls 
	INNER JOIN blocks b ON b.blockroot = bls.block_root AND b.status = '1'
	WHERE block_slot = $1
	ORDER BY bls.validatorindex`, slot)
	if err != nil {
		if err == sql.ErrNoRows {
			return change, nil
		}
		return nil, fmt.Errorf("error getting slot blocks_bls_change: %w", err)
	}

	return change, nil
}

func GetValidatorBLSChange(validatorindex uint64) (*types.BLSChange, error) {
	change := &types.BLSChange{}

	err := ReaderDb.Get(change, `
	SELECT 
		bls.block_slot as slot, 
		bls.signature, 
		bls.pubkey, 
		bls.address 
	FROM blocks_bls_change bls
	INNER JOIN blocks b ON b.blockroot = bls.block_root AND b.status = '1'
	WHERE validatorindex = $1 
	ORDER BY bls.block_slot`, validatorindex)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting validator blocks_bls_change: %w", err)
	}

	return change, nil
}

// GetValidatorsBLSChange returns the BLS change for a list of validators
func GetValidatorsBLSChange(validators []uint64) ([]*types.ValidatorsBLSChange, error) {
	change := make([]*types.ValidatorsBLSChange, 0, len(validators))

	err := ReaderDb.Select(&change, `	
	SELECT
		bls.block_slot AS slot,
		bls.block_root,
		bls.signature,
		bls.pubkey,
		bls.validatorindex,
		bls.address,
		d.withdrawalcredentials
	FROM blocks_bls_change bls
	INNER JOIN blocks b ON b.blockroot = bls.block_root AND b.status = '1'
	LEFT JOIN validators v ON v.validatorindex = bls.validatorindex
	LEFT JOIN (
		SELECT ROW_NUMBER() OVER (PARTITION BY publickey ORDER BY block_slot) AS rn, withdrawalcredentials, publickey, block_root FROM blocks_deposits d
		INNER JOIN blocks b ON b.blockroot = d.block_root AND b.status = '1'
	) AS d ON d.publickey = v.pubkey AND rn = 1
	WHERE bls.validatorindex = ANY($1)
	ORDER BY bls.block_slot DESC
	`, pq.Array(validators))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting validators blocks_bls_change: %w", err)
	}

	return change, nil
}

func GetValidatorsInitialWithdrawalCredentials(validators []uint64) ([][]byte, error) {
	var withdrawalCredentials [][]byte

	err := ReaderDb.Select(&withdrawalCredentials, `
	SELECT 
		withdrawalcredentials 
	FROM 
		blocks_deposits d
	LEFT JOIN blocks b ON b.blockroot = d.block_root AND b.status = '1'
	WHERE
		validatorindex = ANY($1)`, pq.Array(validators))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting validator initial withdrawal credentials: %w", err)
	}

	return withdrawalCredentials, nil
}

func GetWithdrawableValidatorCount(epoch uint64) (uint64, error) {
	var count uint64
	err := ReaderDb.Get(&count, `
	SELECT 
		count(*) 
	FROM 
		validators 
	INNER JOIN (
		SELECT validatorindex, 
                end_effective_balance, 
                end_balance,
                DAY
        FROM
                validator_stats
        WHERE DAY = (SELECT COALESCE(MAX(day), 0) FROM validator_stats_status)) as stats 
	ON stats.validatorindex = validators.validatorindex
	WHERE 
		validators.withdrawalcredentials LIKE '\x01' || '%'::bytea AND ((stats.end_effective_balance = $1 AND stats.end_balance > $1) OR (validators.withdrawableepoch <= $2 AND stats.end_balance > 0));`, utils.Config.Chain.Config.MaxEffectiveBalance, epoch)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("error getting withdrawable validator count: %w", err)
	}

	return count, nil
}

func GetPendingBLSChangeValidatorCount() (uint64, error) {
	var count uint64

	err := ReaderDb.Get(&count, `
	SELECT 
		count(*) 
	FROM 
		validators 
	WHERE 
		withdrawalcredentials LIKE '\x00' || '%'::bytea`)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("error getting withdrawable validator count: %w", err)
	}

	return count, nil
}
