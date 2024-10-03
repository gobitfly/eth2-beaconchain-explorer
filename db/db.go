package db

import (
	"bytes"
	"database/sql"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gobitfly/eth2-beaconchain-explorer/metrics"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pressly/goose/v3"
	prysm_deposit "github.com/prysmaticlabs/prysm/v3/contracts/deposit"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"github.com/sirupsen/logrus"

	"github.com/gobitfly/eth2-beaconchain-explorer/rpc"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var EmbedMigrations embed.FS

var DBPGX *pgxpool.Conn

// DB is a pointer to the explorer-database
var WriterDb *sqlx.DB
var ReaderDb *sqlx.DB

var logger = logrus.StandardLogger().WithField("module", "db")

var farFutureEpoch = uint64(18446744073709551615)
var maxSqlNumber = uint64(9223372036854775807)

const WithdrawalsQueryLimit = 10000
const BlsChangeQueryLimit = 10000
const MaxSqlInteger = 2147483647

const DefaultInfScrollRows = 25

var ErrNoStats = errors.New("no stats available")

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

func mustInitDB(writer *types.DatabaseConfig, reader *types.DatabaseConfig, driverName string, databaseBrand string) (*sqlx.DB, *sqlx.DB) {
	if writer.MaxOpenConns == 0 {
		writer.MaxOpenConns = 50
	}
	if writer.MaxIdleConns == 0 {
		writer.MaxIdleConns = 10
	}
	if writer.MaxOpenConns < writer.MaxIdleConns {
		writer.MaxIdleConns = writer.MaxOpenConns
	}

	if reader.MaxOpenConns == 0 {
		reader.MaxOpenConns = 50
	}
	if reader.MaxIdleConns == 0 {
		reader.MaxIdleConns = 10
	}
	if reader.MaxOpenConns < reader.MaxIdleConns {
		reader.MaxIdleConns = reader.MaxOpenConns
	}

	var sslParam string
	if driverName == "clickhouse" {
		sslParam = "secure=false"
		if writer.SSL {
			sslParam = "secure=true"
		}
		// debug
		// sslParam += "&debug=true"
	} else {
		sslParam = "sslmode=disable"
		if writer.SSL {
			sslParam = "sslmode=require"
		}
	}

	logger.Infof("connecting to %s database %s:%s/%s as writer with %d/%d max open/idle connections", databaseBrand, writer.Host, writer.Port, writer.Name, writer.MaxOpenConns, writer.MaxIdleConns)
	dbConnWriter, err := sqlx.Open(driverName, fmt.Sprintf("%s://%s:%s@%s/%s?%s", databaseBrand, writer.Username, writer.Password, net.JoinHostPort(writer.Host, writer.Port), writer.Name, sslParam))
	if err != nil {
		logger.Fatal(err, "error getting Connection Writer database", 0)
	}

	dbTestConnection(dbConnWriter, fmt.Sprintf("database %v:%v/%v", writer.Host, writer.Port, writer.Name))
	dbConnWriter.SetConnMaxIdleTime(time.Second * 30)
	dbConnWriter.SetConnMaxLifetime(time.Minute)
	dbConnWriter.SetMaxOpenConns(writer.MaxOpenConns)
	dbConnWriter.SetMaxIdleConns(writer.MaxIdleConns)

	if reader == nil {
		return dbConnWriter, dbConnWriter
	}

	if driverName == "clickhouse" {
		sslParam = "secure=false"
		if writer.SSL {
			sslParam = "secure=true"
		}
		// debug
		// sslParam += "&debug=true"
	} else {
		sslParam = "sslmode=disable"
		if writer.SSL {
			sslParam = "sslmode=require"
		}
	}

	logger.Infof("connecting to %s database %s:%s/%s as reader with %d/%d max open/idle connections", databaseBrand, reader.Host, reader.Port, reader.Name, reader.MaxOpenConns, reader.MaxIdleConns)
	dbConnReader, err := sqlx.Open(driverName, fmt.Sprintf("%s://%s:%s@%s/%s?%s", databaseBrand, reader.Username, reader.Password, net.JoinHostPort(reader.Host, reader.Port), reader.Name, sslParam))
	if err != nil {
		logger.Fatal(err, "error getting Connection Reader database", 0)
	}

	dbTestConnection(dbConnReader, fmt.Sprintf("database %v:%v/%v", writer.Host, writer.Port, writer.Name))
	dbConnReader.SetConnMaxIdleTime(time.Second * 30)
	dbConnReader.SetConnMaxLifetime(time.Minute)
	dbConnReader.SetMaxOpenConns(reader.MaxOpenConns)
	dbConnReader.SetMaxIdleConns(reader.MaxIdleConns)
	return dbConnWriter, dbConnReader
}

func MustInitDB(writer *types.DatabaseConfig, reader *types.DatabaseConfig, driverName string, databaseBrand string) {
	WriterDb, ReaderDb = mustInitDB(writer, reader, driverName, databaseBrand)
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

func GetEth1DepositsJoinEth2Deposits(query string, length, start uint64, orderBy, orderDir string, latestEpoch, validatorOnlineThresholdSlot uint64) ([]*types.EthOneDepositsData, uint64, error) {
	// Initialize the return values
	deposits := []*types.EthOneDepositsData{}
	totalCount := uint64(0)

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

	var param interface{}
	var searchQuery string
	var err error

	// Define the base queries
	deposistsCountQuery := `
		SELECT COUNT(*) FROM eth1_deposits as eth1
		%s`

	deposistsQuery := `
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
		%s
		ORDER BY %s %s
		LIMIT $1
		OFFSET $2`

	// Get the search query and parameter for it
	trimmedQuery := strings.ToLower(strings.TrimPrefix(query, "0x"))
	var hash []byte
	if len(trimmedQuery)%2 == 0 && utils.HashLikeRegex.MatchString(trimmedQuery) {
		hash, err = hex.DecodeString(trimmedQuery)
		if err != nil {
			return nil, 0, err
		}
	}
	if trimmedQuery == "" {
		err = ReaderDb.Get(&totalCount, fmt.Sprintf(deposistsCountQuery, ""))
		if err != nil {
			return nil, 0, err
		}

		err = ReaderDb.Select(&deposits, fmt.Sprintf(deposistsQuery, "", orderBy, orderDir), length, start)
		if err != nil && err != sql.ErrNoRows {
			return nil, 0, err
		}

		return deposits, totalCount, nil
	}

	param = hash
	if utils.IsHash(trimmedQuery) {
		searchQuery = `WHERE eth1.publickey = $3`
	} else if utils.IsEth1Tx(trimmedQuery) {
		// Withdrawal credentials have the same length as a tx hash
		if utils.IsValidWithdrawalCredentials(trimmedQuery) {
			searchQuery = `
				WHERE 
					eth1.tx_hash = $3
					OR eth1.withdrawal_credentials = $3`
		} else {
			searchQuery = `WHERE eth1.tx_hash = $3`
		}
	} else if utils.IsEth1Address(trimmedQuery) {
		searchQuery = `WHERE eth1.from_address = $3`
	} else if uiQuery, parseErr := strconv.ParseUint(query, 10, 31); parseErr == nil { // Limit to 31 bits to stay within math.MaxInt32
		param = uiQuery
		searchQuery = `WHERE eth1.block_number = $3`
	} else {
		// The query does not fulfill any of the requirements for a search
		return deposits, totalCount, nil
	}

	// The deposits count query only has one parameter for the search
	countSearchQuery := strings.ReplaceAll(searchQuery, "$3", "$1")

	err = ReaderDb.Get(&totalCount, fmt.Sprintf(deposistsCountQuery, countSearchQuery), param)
	if err != nil {
		return nil, 0, err
	}

	err = ReaderDb.Select(&deposits, fmt.Sprintf(deposistsQuery, searchQuery, orderBy, orderDir), length, start, param)
	if err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	}

	return deposits, totalCount, nil
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

func GetEth2Deposits(query string, length, start uint64, orderBy, orderDir string) ([]*types.EthTwoDepositData, uint64, error) {
	// Initialize the return values
	deposits := []*types.EthTwoDepositData{}
	totalCount := uint64(0)

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

	var param interface{}
	var searchQuery string
	var err error

	// Define the base queries
	deposistsCountQuery := `
		SELECT COUNT(*)
		FROM blocks_deposits
		INNER JOIN blocks ON blocks_deposits.block_root = blocks.blockroot AND blocks.status = '1'
		%s`

	deposistsQuery := `
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
			%s
			ORDER BY %s %s
			LIMIT $1
			OFFSET $2`

	// Get the search query and parameter for it
	trimmedQuery := strings.ToLower(strings.TrimPrefix(query, "0x"))
	var hash []byte
	if len(trimmedQuery)%2 == 0 && utils.HashLikeRegex.MatchString(trimmedQuery) {
		hash, err = hex.DecodeString(trimmedQuery)
		if err != nil {
			return nil, 0, err
		}
	}
	if trimmedQuery == "" {
		err = ReaderDb.Get(&totalCount, fmt.Sprintf(deposistsCountQuery, ""))
		if err != nil {
			return nil, 0, err
		}

		err = ReaderDb.Select(&deposits, fmt.Sprintf(deposistsQuery, "", orderBy, orderDir), length, start)
		if err != nil && err != sql.ErrNoRows {
			return nil, 0, err
		}

		return deposits, totalCount, nil
	}

	if utils.IsHash(trimmedQuery) {
		param = hash
		searchQuery = `WHERE blocks_deposits.publickey = $3`
	} else if utils.IsValidWithdrawalCredentials(trimmedQuery) {
		param = hash
		searchQuery = `WHERE blocks_deposits.withdrawalcredentials = $3`
	} else if utils.IsEth1Address(trimmedQuery) {
		param = hash
		searchQuery = `
				LEFT JOIN eth1_deposits ON blocks_deposits.publickey = eth1_deposits.publickey
				WHERE eth1_deposits.from_address = $3`
	} else if uiQuery, parseErr := strconv.ParseUint(query, 10, 31); parseErr == nil { // Limit to 31 bits to stay within math.MaxInt32
		param = uiQuery
		searchQuery = `WHERE blocks_deposits.block_slot = $3`
	} else {
		// The query does not fulfill any of the requirements for a search
		return deposits, totalCount, nil
	}

	// The deposits count query only has one parameter for the search
	countSearchQuery := strings.ReplaceAll(searchQuery, "$3", "$1")

	err = ReaderDb.Get(&totalCount, fmt.Sprintf(deposistsCountQuery, countSearchQuery), param)
	if err != nil {
		return nil, 0, err
	}

	err = ReaderDb.Select(&deposits, fmt.Sprintf(deposistsQuery, searchQuery, orderBy, orderDir), length, start, param)
	if err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	}

	return deposits, totalCount, nil
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

func GetAllSlots(tx *sqlx.Tx) ([]uint64, error) {
	var slots []uint64
	err := tx.Select(&slots, "SELECT slot FROM blocks ORDER BY slot")

	if err != nil {
		return nil, fmt.Errorf("error retrieving all slots from the DB: %w", err)
	}

	return slots, nil
}

func SetSlotFinalizationAndStatus(slot uint64, finalized bool, status string, tx *sqlx.Tx) error {
	_, err := tx.Exec("UPDATE blocks SET finalized = $1, status = $2 WHERE slot = $3", finalized, status, slot)

	if err != nil {
		return fmt.Errorf("error setting slot finalization and status: %w", err)
	}

	return nil
}

type GetAllNonFinalizedSlotsRow struct {
	Slot      uint64 `db:"slot"`
	BlockRoot []byte `db:"blockroot"`
	Finalized bool   `db:"finalized"`
	Status    string `db:"status"`
}

func GetAllNonFinalizedSlots() ([]*GetAllNonFinalizedSlotsRow, error) {
	var slots []*GetAllNonFinalizedSlotsRow
	err := WriterDb.Select(&slots, "SELECT slot, blockroot, finalized, status FROM blocks WHERE NOT finalized ORDER BY slot")

	if err != nil {
		return nil, fmt.Errorf("error retrieving all non finalized slots from the DB: %w", err)
	}

	return slots, nil
}

// Get latest finalized epoch
func GetLatestFinalizedEpoch() (uint64, error) {
	var latestFinalized uint64
	err := WriterDb.Get(&latestFinalized, "SELECT epoch FROM epochs WHERE finalized ORDER BY epoch DESC LIMIT 1")
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		utils.LogError(err, "error retrieving latest exported finalized epoch from the database", 0)
		return 0, err
	}

	return latestFinalized, nil
}

// GetValidatorPublicKeys will return the public key for a list of validator indices and or public keys
func GetValidatorPublicKeys(indices []uint64, keys [][]byte) ([][]byte, error) {
	var publicKeys [][]byte
	err := ReaderDb.Select(&publicKeys, "SELECT pubkey FROM validators WHERE validatorindex = ANY($1) OR pubkey = ANY($2)", indices, keys)

	return publicKeys, err
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

		// retrieve address names from bigtable
		names := make(map[string]string)
		for _, v := range deposits.Eth1Deposits {
			names[string(v.FromAddress)] = ""
		}
		names, _, err = BigtableClient.GetAddressesNamesArMetadata(&names, nil)
		if err != nil {
			return nil, err
		}

		for k, v := range deposits.Eth1Deposits {
			deposits.Eth1Deposits[k].FromName = names[string(v.FromAddress)]
		}
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
func SaveValidatorQueue(validators *types.ValidatorQueue, tx *sqlx.Tx) error {
	_, err := tx.Exec(`
		INSERT INTO queue (ts, entering_validators_count, exiting_validators_count)
		VALUES (date_trunc('hour', now()), $1, $2)
		ON CONFLICT (ts) DO UPDATE SET
			entering_validators_count = excluded.entering_validators_count, 
			exiting_validators_count = excluded.exiting_validators_count`,
		validators.Activating, validators.Exiting)
	return err
}

func SaveBlock(block *types.Block, forceSlotUpdate bool, tx *sqlx.Tx) error {

	blocksMap := make(map[uint64]map[string]*types.Block)
	if blocksMap[block.Slot] == nil {
		blocksMap[block.Slot] = make(map[string]*types.Block)
	}
	blocksMap[block.Slot][fmt.Sprintf("%x", block.BlockRoot)] = block

	err := saveBlocks(blocksMap, tx, forceSlotUpdate)
	if err != nil {
		logger.Fatalf("error saving blocks to db: %v", err)
		return fmt.Errorf("error saving blocks to db: %w", err)
	}

	return nil
}

// SaveEpoch will save the epoch data into the database
func SaveEpoch(epoch uint64, validators []*types.Validator, client rpc.Client, tx *sqlx.Tx) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_save_epoch").Observe(time.Since(start).Seconds())
		logger.WithFields(logrus.Fields{"epoch": epoch, "duration": time.Since(start)}).Info("completed saving epoch")
	}()

	logger.WithFields(logrus.Fields{"chainEpoch": utils.TimeToEpoch(time.Now()), "exportEpoch": epoch}).Infof("starting export of epoch %v", epoch)

	logger.Infof("exporting epoch statistics data")
	proposerSlashingsCount := 0
	attesterSlashingsCount := 0
	attestationsCount := 0
	depositCount := 0
	voluntaryExitCount := 0
	withdrawalCount := 0

	// for _, slot := range data.Blocks {
	// 	for _, b := range slot {
	// 		proposerSlashingsCount += len(b.ProposerSlashings)
	// 		attesterSlashingsCount += len(b.AttesterSlashings)
	// 		attestationsCount += len(b.Attestations)
	// 		depositCount += len(b.Deposits)
	// 		voluntaryExitCount += len(b.VoluntaryExits)
	// 		if b.ExecutionPayload != nil {
	// 			withdrawalCount += len(b.ExecutionPayload.Withdrawals)
	// 		}
	// 	}
	// }

	validatorBalanceSum := new(big.Int)
	validatorEffectiveBalanceSum := new(big.Int)
	validatorsCount := 0
	for _, v := range validators {
		if v.ExitEpoch > epoch && v.ActivationEpoch <= epoch {
			validatorsCount++
			validatorBalanceSum = new(big.Int).Add(validatorBalanceSum, new(big.Int).SetUint64(v.Balance))
			validatorEffectiveBalanceSum = new(big.Int).Add(validatorEffectiveBalanceSum, new(big.Int).SetUint64(v.EffectiveBalance))

		}
	}

	validatorBalanceAverage := new(big.Int).Div(validatorBalanceSum, new(big.Int).SetInt64(int64(validatorsCount)))

	_, err := tx.Exec(`
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
			eligibleether, 
			globalparticipationrate, 
			votedether,
			finalized
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
			votedether              = excluded.votedether,
			finalized               = excluded.finalized`,
		epoch,
		0,
		proposerSlashingsCount,
		attesterSlashingsCount,
		attestationsCount,
		depositCount,
		withdrawalCount,
		voluntaryExitCount,
		validatorsCount,
		validatorBalanceAverage.Uint64(),
		validatorBalanceSum.Uint64(),
		validatorEffectiveBalanceSum.Uint64(),
		0,
		0,
		false)

	if err != nil {
		return fmt.Errorf("error executing save epoch statement: %w", err)
	}

	lookback := uint64(0)
	if epoch > 3 {
		lookback = epoch - 3
	}
	// delete duplicate scheduled slots
	_, err = tx.Exec("delete from blocks where slot in (select slot from blocks where epoch >= $1 group by slot having count(*) > 1) and blockroot = $2;", lookback, []byte{0x0})
	if err != nil {
		return fmt.Errorf("error cleaning up blocks table: %w", err)
	}

	// delete duplicate missed blocks
	_, err = tx.Exec("delete from blocks where slot in (select slot from blocks where epoch >= $1 group by slot having count(*) > 1) and blockroot = $2;", lookback, []byte{0x1})
	if err != nil {
		return fmt.Errorf("error cleaning up blocks table: %w", err)
	}
	return nil
}

func saveGraffitiwall(block *types.Block, tx *sqlx.Tx) error {
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
        ON CONFLICT (slot) DO UPDATE SET
            x = EXCLUDED.x,
            y = EXCLUDED.y,
            color = EXCLUDED.color,
            validator = EXCLUDED.validator;
		`)
	if err != nil {
		return err
	}
	defer stmtGraffitiwall.Close()

	regexes := [...]*regexp.Regexp{
		regexp.MustCompile("graffitiwall:([0-9]{1,3}):([0-9]{1,3}):#([0-9a-fA-F]{6})"),
		regexp.MustCompile("gw:([0-9]{3})([0-9]{3})([0-9a-fA-F]{6})"),
	}

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
	return nil
}

func SaveValidators(epoch uint64, validators []*types.Validator, client rpc.Client, activationBalanceBatchSize int, tx *sqlx.Tx) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_save_validators").Observe(time.Since(start).Seconds())
	}()

	if activationBalanceBatchSize <= 0 {
		activationBalanceBatchSize = 10000
	}

	var genesisBalances map[uint64][]*types.ValidatorBalance

	if epoch == 0 {
		var err error

		indices := make([]uint64, 0, len(validators))

		for _, validator := range validators {
			indices = append(indices, validator.Index)
		}
		genesisBalances, err = BigtableClient.GetValidatorBalanceHistory(indices, 0, 0)
		if err != nil {
			return fmt.Errorf("error retrieving genesis validator balances: %w", err)
		}
	}

	validatorsByIndex := make(map[uint64]*types.Validator, len(validators))
	for _, v := range validators {
		validatorsByIndex[v.Index] = v
	}

	var currentState []*types.Validator
	err := tx.Select(&currentState, "SELECT validatorindex, withdrawableepoch, withdrawalcredentials, slashed, activationeligibilityepoch, activationepoch, exitepoch, status FROM validators;")

	if err != nil {
		return fmt.Errorf("error retrieving current validator state set: %v", err)
	}

	for ; ; time.Sleep(time.Second) { // wait till the last attestation in memory cache has been populated by the exporter
		BigtableClient.LastAttestationCacheMux.Lock()
		if BigtableClient.LastAttestationCache != nil {
			BigtableClient.LastAttestationCacheMux.Unlock()
			break
		}
		BigtableClient.LastAttestationCacheMux.Unlock()
		logger.Infof("waiting until LastAttestation in memory cache is available")
	}

	currentStateMap := make(map[uint64]*types.Validator, len(currentState))
	latestBlock := uint64(0)
	BigtableClient.LastAttestationCacheMux.Lock()
	for _, v := range currentState {
		if BigtableClient.LastAttestationCache[v.Index] > latestBlock {
			latestBlock = BigtableClient.LastAttestationCache[v.Index]
		}
		currentStateMap[v.Index] = v
	}
	BigtableClient.LastAttestationCacheMux.Unlock()

	thresholdSlot := uint64(0)
	if latestBlock >= 64 {
		thresholdSlot = latestBlock - 64
	}

	latestEpoch := latestBlock / utils.Config.Chain.ClConfig.SlotsPerEpoch

	var queries strings.Builder

	insertStmt, err := tx.Prepare(`INSERT INTO validators (
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
		status
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);`)
	if err != nil {
		return fmt.Errorf("error preparing insert validator statement: %w", err)
	}

	validatorStatusCounts := make(map[string]int)

	updates := 0
	for _, v := range validators {

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

		c := currentStateMap[v.Index]

		if c == nil {
			if v.Index%1000 == 0 {
				logger.Infof("validator %v is new", v.Index)
			}

			_, err = insertStmt.Exec(
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
			)

			if err != nil {
				logger.Errorf("error saving new validator %v: %v", v.Index, err)
			}
			validatorStatusCounts[v.Status]++
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
			BigtableClient.LastAttestationCacheMux.Lock()
			offline := BigtableClient.LastAttestationCache[v.Index] < thresholdSlot
			BigtableClient.LastAttestationCacheMux.Unlock()

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

			validatorStatusCounts[v.Status]++
			if c.Status != v.Status {
				logger.Tracef("Status changed for validator %v from %v to %v", v.Index, c.Status, v.Status)
				// logger.Tracef("v.ActivationEpoch %v, latestEpoch %v, lastAttestationSlots[v.Index] %v, thresholdSlot %v", v.ActivationEpoch, latestEpoch, lastAttestationSlots[v.Index], thresholdSlot)
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

	err = insertStmt.Close()
	if err != nil {
		return fmt.Errorf("error closing insert validator statement: %w", err)
	}

	if updates > 0 {
		updateStart := time.Now()
		logger.Infof("applying %v validator table update queries", updates)
		_, err = tx.Exec(queries.String())
		if err != nil {
			logger.Errorf("error executing validator update query: %v", err)
			return err
		}
		logger.Infof("validator table update completed, took %v", time.Since(updateStart))
	}

	s := time.Now()
	newValidators := []struct {
		Validatorindex  uint64
		ActivationEpoch uint64
	}{}

	err = tx.Select(&newValidators, "SELECT validatorindex, activationepoch FROM validators WHERE balanceactivation IS NULL ORDER BY activationepoch LIMIT $1", activationBalanceBatchSize)
	if err != nil {
		return fmt.Errorf("error retreiving activation epoch balances from db: %w", err)
	}

	balanceCache := make(map[uint64]map[uint64]uint64)
	currentActivationEpoch := uint64(0)

	// get genesis balances of all validators for performance

	for _, newValidator := range newValidators {
		if newValidator.ActivationEpoch > epoch {
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
				return fmt.Errorf("error retreiving validator balance history: %w", err)
			}
		}

		foundBalance := uint64(0)
		if balance[newValidator.Validatorindex] == nil || len(balance[newValidator.Validatorindex]) == 0 {
			logger.Warnf("no activation epoch balance found for validator %v for epoch %v in bigtable, trying node", newValidator.Validatorindex, newValidator.ActivationEpoch)

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
			return fmt.Errorf("error updating activation epoch balance for validator %v: %w", newValidator.Validatorindex, err)
		}
	}

	logger.Infof("updating validator activation epoch balance completed, took %v", time.Since(s))

	logger.Infof("updating validator status counts")
	s = time.Now()
	_, err = tx.Exec("TRUNCATE TABLE validators_status_counts;")
	if err != nil {
		return fmt.Errorf("error truncating validators_status_counts table: %w", err)
	}
	for status, count := range validatorStatusCounts {
		_, err = tx.Exec("INSERT INTO validators_status_counts (status, validator_count) VALUES ($1, $2);", status, count)
		if err != nil {
			return fmt.Errorf("error updating validator status counts: %w", err)
		}
	}
	logger.Infof("updating validator status counts completed, took %v", time.Since(s))

	s = time.Now()
	_, err = tx.Exec("ANALYZE (SKIP_LOCKED) validators;")
	if err != nil {
		return fmt.Errorf("analyzing validators table: %w", err)
	}
	logger.Infof("analyze of validators table completed, took %v", time.Since(s))

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

func saveBlocks(blocks map[uint64]map[string]*types.Block, tx *sqlx.Tx, forceSlotUpdate bool) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_save_blocks").Observe(time.Since(start).Seconds())
	}()

	domain, err := utils.GetSigningDomain()
	if err != nil {
		return err
	}

	stmtExecutionPayload, err := tx.Prepare(`
		INSERT INTO execution_payloads (block_hash)
		VALUES ($1)
		ON CONFLICT (block_hash) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtExecutionPayload.Close()

	stmtBlock, err := tx.Prepare(`
		INSERT INTO blocks (epoch, slot, blockroot, parentroot, stateroot, signature, randaoreveal, graffiti, graffiti_text, eth1data_depositroot, eth1data_depositcount, eth1data_blockhash, syncaggregate_bits, syncaggregate_signature, proposerslashingscount, attesterslashingscount, attestationscount, depositscount, withdrawalcount, voluntaryexitscount, syncaggregate_participation, proposer, status, exec_parent_hash, exec_fee_recipient, exec_state_root, exec_receipts_root, exec_logs_bloom, exec_random, exec_block_number, exec_gas_limit, exec_gas_used, exec_timestamp, exec_extra_data, exec_base_fee_per_gas, exec_block_hash, exec_transactions_count, exec_blob_gas_used, exec_excess_blob_gas, exec_blob_transactions_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40)
		ON CONFLICT (slot, blockroot) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtBlock.Close()

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

	stmtBlobs, err := tx.Prepare(`
		INSERT INTO blocks_blob_sidecars (block_slot, block_root, index, kzg_commitment, kzg_proof, blob_versioned_hash)
		VALUES ($1, $2, $3, $4, $5, $6) 
		ON CONFLICT (block_root, index) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtBlobs.Close()

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

			if !forceSlotUpdate {
				var dbBlockRootHash []byte
				err := tx.Get(&dbBlockRootHash, "SELECT blockroot FROM blocks WHERE slot = $1 and blockroot = $2", b.Slot, b.BlockRoot)
				if err == nil && bytes.Equal(dbBlockRootHash, b.BlockRoot) {
					blockLog.Infof("skipping export of block as it is already present in the db")
					continue
				} else if err != nil && err != sql.ErrNoRows {
					return fmt.Errorf("error checking for block in db: %w", err)
				}
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
				b.Proposer = MaxSqlInteger
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

			type exectionPayloadData struct {
				ParentHash      []byte
				FeeRecipient    []byte
				StateRoot       []byte
				ReceiptRoot     []byte
				LogsBloom       []byte
				Random          []byte
				BlockNumber     *uint64
				GasLimit        *uint64
				GasUsed         *uint64
				Timestamp       *uint64
				ExtraData       []byte
				BaseFeePerGas   *uint64
				BlockHash       []byte
				TxCount         *int64
				WithdrawalCount *int64
				BlobGasUsed     *uint64
				ExcessBlobGas   *uint64
				BlobTxCount     *int64
			}

			execData := new(exectionPayloadData)

			if b.ExecutionPayload != nil {
				txCount := int64(len(b.ExecutionPayload.Transactions))
				withdrawalCount := int64(len(b.ExecutionPayload.Withdrawals))
				blobTxCount := int64(len(b.BlobKZGCommitments))
				execData = &exectionPayloadData{
					ParentHash:      b.ExecutionPayload.ParentHash,
					FeeRecipient:    b.ExecutionPayload.FeeRecipient,
					StateRoot:       b.ExecutionPayload.StateRoot,
					ReceiptRoot:     b.ExecutionPayload.ReceiptsRoot,
					LogsBloom:       b.ExecutionPayload.LogsBloom,
					Random:          b.ExecutionPayload.Random,
					BlockNumber:     &b.ExecutionPayload.BlockNumber,
					GasLimit:        &b.ExecutionPayload.GasLimit,
					GasUsed:         &b.ExecutionPayload.GasUsed,
					Timestamp:       &b.ExecutionPayload.Timestamp,
					ExtraData:       b.ExecutionPayload.ExtraData,
					BaseFeePerGas:   &b.ExecutionPayload.BaseFeePerGas,
					BlockHash:       b.ExecutionPayload.BlockHash,
					TxCount:         &txCount,
					WithdrawalCount: &withdrawalCount,
					BlobGasUsed:     &b.ExecutionPayload.BlobGasUsed,
					ExcessBlobGas:   &b.ExecutionPayload.ExcessBlobGas,
					BlobTxCount:     &blobTxCount,
				}
				_, err = stmtExecutionPayload.Exec(execData.BlockHash)
				if err != nil {
					return fmt.Errorf("error executing stmtExecutionPayload for block %v: %w", b.Slot, err)
				}
			}
			_, err = stmtBlock.Exec(
				b.Slot/utils.Config.Chain.ClConfig.SlotsPerEpoch,
				b.Slot,
				b.BlockRoot,
				b.ParentRoot,
				b.StateRoot,
				b.Signature,
				b.RandaoReveal,
				b.Graffiti,
				utils.GraffitiToString(b.Graffiti),
				b.Eth1Data.DepositRoot,
				b.Eth1Data.DepositCount,
				b.Eth1Data.BlockHash,
				syncAggBits,
				syncAggSig,
				len(b.ProposerSlashings),
				len(b.AttesterSlashings),
				len(b.Attestations),
				len(b.Deposits),
				execData.WithdrawalCount,
				len(b.VoluntaryExits),
				syncAggParticipation,
				b.Proposer,
				strconv.FormatUint(b.Status, 10),
				execData.ParentHash,
				execData.FeeRecipient,
				execData.StateRoot,
				execData.ReceiptRoot,
				execData.LogsBloom,
				execData.Random,
				execData.BlockNumber,
				execData.GasLimit,
				execData.GasUsed,
				execData.Timestamp,
				execData.ExtraData,
				execData.BaseFeePerGas,
				execData.BlockHash,
				execData.TxCount,
				execData.BlobGasUsed,
				execData.ExcessBlobGas,
				execData.BlobTxCount,
			)
			if err != nil {
				return fmt.Errorf("error executing stmtBlocks for block %v: %w", b.Slot, err)
			}
			blockLog.WithField("duration", time.Since(t)).Tracef("stmtBlock")
			logger.Tracef("done, took %v", time.Since(t))

			t = time.Now()
			logger.Tracef("writing BlobKZGCommitments data")
			for i, c := range b.BlobKZGCommitments {
				_, err := stmtBlobs.Exec(b.Slot, b.BlockRoot, i, c, b.BlobKZGProofs[i], utils.VersionedBlobHash(c).Bytes())
				if err != nil {
					return fmt.Errorf("error executing stmtBlobs for block at slot %v index %v: %w", b.Slot, i, err)
				}
			}
			logger.Tracef("done, took %v", time.Since(t))
			t = time.Now()
			logger.Tracef("writing transactions and withdrawal data")
			if payload := b.ExecutionPayload; payload != nil {
				for i, w := range payload.Withdrawals {
					_, err := stmtWithdrawals.Exec(b.Slot, b.BlockRoot, w.Index, w.ValidatorIndex, w.Address, w.Amount)
					if err != nil {
						return fmt.Errorf("error executing stmtWithdrawals for block at slot %v index %v: %w", b.Slot, i, err)
					}
				}
			}
			logger.Tracef("done, took %v", time.Since(t))
			t = time.Now()
			logger.Tracef("writing proposer slashings data")
			for i, ps := range b.ProposerSlashings {
				_, err := stmtProposerSlashing.Exec(b.Slot, i, b.BlockRoot, ps.ProposerIndex, ps.Header1.Slot, ps.Header1.ParentRoot, ps.Header1.StateRoot, ps.Header1.BodyRoot, ps.Header1.Signature, ps.Header2.Slot, ps.Header2.ParentRoot, ps.Header2.StateRoot, ps.Header2.BodyRoot, ps.Header2.Signature)
				if err != nil {
					return fmt.Errorf("error executing stmtProposerSlashing for block at slot %v index %v: %w", b.Slot, i, err)
				}
			}
			blockLog.WithField("duration", time.Since(t)).Tracef("stmtProposerSlashing")
			t = time.Now()
			logger.Tracef("writing bls change data")
			for i, bls := range b.SignedBLSToExecutionChange {
				_, err := stmtBLSChange.Exec(b.Slot, b.BlockRoot, bls.Message.Validatorindex, bls.Signature, bls.Message.BlsPubkey, bls.Message.Address)
				if err != nil {
					return fmt.Errorf("error executing stmtBLSChange for block %v index %v: %w", b.Slot, i, err)
				}
			}
			blockLog.WithField("duration", time.Since(t)).Tracef("stmtBLSChange")
			t = time.Now()

			for i, as := range b.AttesterSlashings {
				_, err := stmtAttesterSlashing.Exec(b.Slot, i, b.BlockRoot, pq.Array(as.Attestation1.AttestingIndices), as.Attestation1.Signature, as.Attestation1.Data.Slot, as.Attestation1.Data.CommitteeIndex, as.Attestation1.Data.BeaconBlockRoot, as.Attestation1.Data.Source.Epoch, as.Attestation1.Data.Source.Root, as.Attestation1.Data.Target.Epoch, as.Attestation1.Data.Target.Root, pq.Array(as.Attestation2.AttestingIndices), as.Attestation2.Signature, as.Attestation2.Data.Slot, as.Attestation2.Data.CommitteeIndex, as.Attestation2.Data.BeaconBlockRoot, as.Attestation2.Data.Source.Epoch, as.Attestation2.Data.Source.Root, as.Attestation2.Data.Target.Epoch, as.Attestation2.Data.Target.Root)
				if err != nil {
					return fmt.Errorf("error executing stmtAttesterSlashing for block %v index %v: %w", b.Slot, i, err)
				}
			}
			blockLog.WithField("duration", time.Since(t)).Tracef("stmtAttesterSlashing")
			t = time.Now()
			for i, a := range b.Attestations {
				_, err = stmtAttestations.Exec(b.Slot, i, b.BlockRoot, a.AggregationBits, pq.Array(a.Attesters), a.Signature, a.Data.Slot, a.Data.CommitteeIndex, a.Data.BeaconBlockRoot, a.Data.Source.Epoch, a.Data.Source.Root, a.Data.Target.Epoch, a.Data.Target.Root)
				if err != nil {
					return fmt.Errorf("error executing stmtAttestations for block %v index %v: %w", b.Slot, i, err)
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
					return fmt.Errorf("error executing stmtDeposits for block %v index %v: %w", b.Slot, i, err)
				}
			}
			blockLog.WithField("duration", time.Since(t)).Tracef("deposits")
			t = time.Now()

			for i, ve := range b.VoluntaryExits {
				_, err := stmtVoluntaryExits.Exec(b.Slot, i, b.BlockRoot, ve.Epoch, ve.ValidatorIndex, ve.Signature)
				if err != nil {
					return fmt.Errorf("error executing stmtVoluntaryExits for block %v index %v: %w", b.Slot, i, err)
				}
			}
			blockLog.WithField("duration", time.Since(t)).Tracef("exits")
			t = time.Now()

			_, err = stmtProposalAssignments.Exec(b.Slot/utils.Config.Chain.ClConfig.SlotsPerEpoch, b.Proposer, b.Slot, b.Status)
			if err != nil {
				return fmt.Errorf("error executing stmtProposalAssignments for block %v: %w", b.Slot, err)
			}
			blockLog.WithField("duration", time.Since(t)).Tracef("stmtProposalAssignments")

			// save the graffitiwall data of the block the the db
			t = time.Now()
			err = saveGraffitiwall(b, tx)
			if err != nil {
				return fmt.Errorf("error saving graffitiwall data to the db: %v", err)
			}
			blockLog.WithField("duration", time.Since(t)).Tracef("saveGraffitiwall")
		}
	}

	return nil
}

// UpdateEpochStatus will update the epoch status in the database
func UpdateEpochStatus(stats *types.ValidatorParticipation, tx *sqlx.Tx) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("db_update_epochs_status").Observe(time.Since(start).Seconds())
	}()

	_, err := tx.Exec(`
		UPDATE epochs SET
			eligibleether = $1,
			globalparticipationrate = $2,
			votedether = $3,
			finalized = $4,
			blockscount = (SELECT COUNT(*) FROM blocks WHERE epoch = $5 AND status = '1'),
			proposerslashingscount = (SELECT COALESCE(SUM(proposerslashingscount),0) FROM blocks WHERE epoch = $5 AND status = '1'),
			attesterslashingscount = (SELECT COALESCE(SUM(attesterslashingscount),0) FROM blocks WHERE epoch = $5 AND status = '1'),
			attestationscount = (SELECT COALESCE(SUM(attestationscount),0) FROM blocks WHERE epoch = $5 AND status = '1'),
			depositscount = (SELECT COALESCE(SUM(depositscount),0) FROM blocks WHERE epoch = $5 AND status = '1'),
			withdrawalcount = (SELECT COALESCE(SUM(withdrawalcount),0) FROM blocks WHERE epoch = $5 AND status = '1'),
			voluntaryexitscount = (SELECT COALESCE(SUM(voluntaryexitscount),0) FROM blocks WHERE epoch = $5 AND status = '1')
		WHERE epoch = $5`,
		stats.EligibleEther, stats.GlobalParticipationRate, stats.VotedEther, stats.Finalized, stats.Epoch)

	return err
}

// GetValidatorIndices will return the total-validator-indices
func GetValidatorIndices() ([]uint64, error) {
	indices := []uint64{}
	err := ReaderDb.Select(&indices, "select validatorindex from validators order by validatorindex;")
	return indices, err
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

func UpdateQueueDeposits(tx *sqlx.Tx) error {
	start := time.Now()
	defer func() {
		logger.Infof("took %v seconds to update queue deposits", time.Since(start).Seconds())
		metrics.TaskDuration.WithLabelValues("update_queue_deposits").Observe(time.Since(start).Seconds())
	}()

	// first we remove any validator that isn't queued anymore
	_, err := tx.Exec(`
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
	_, err = tx.Exec(`
		INSERT INTO validator_queue_deposits
		SELECT validatorindex FROM validators WHERE activationepoch=$1 and status='pending' ON CONFLICT DO NOTHING
	`, maxSqlNumber)
	if err != nil {
		logger.Errorf("error adding queued publickeys to validator_queue_deposits: %v", err)
		return err
	}

	// now we add the activationeligibilityepoch where it is missing
	_, err = tx.Exec(`
		UPDATE validator_queue_deposits 
		SET 
			activationeligibilityepoch=validators.activationeligibilityepoch
		FROM validators
		WHERE 
			validator_queue_deposits.activationeligibilityepoch IS NULL AND
			validator_queue_deposits.validatorindex = validators.validatorindex
	`)
	if err != nil {
		logger.Errorf("error updating activationeligibilityepoch on validator_queue_deposits: %v", err)
		return err
	}

	// efficiently collect the tnx that pushed each validator over 32 ETH.
	_, err = tx.Exec(`
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
						select validatorindex from validator_queue_deposits where block_slot is null or block_index is null
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
	var selected struct {
		BlockSlot                  uint64 `db:"block_slot"`
		BlockIndex                 uint64 `db:"block_index"`
		ActivationEligibilityEpoch uint64 `db:"activationeligibilityepoch"`
	}
	err := ReaderDb.Get(&selected, `
		SELECT 
			COALESCE(block_index, 0) as block_index, 
			COALESCE(block_slot, 0) as block_slot, 
			COALESCE(activationeligibilityepoch, $2) as activationeligibilityepoch
		FROM validator_queue_deposits
		WHERE 
			validatorindex = $1
		`, validatorIndex, maxSqlNumber)
	if err == sql.ErrNoRows {
		// If we did not find our validator in the queue it is most likly that he has not yet been added so we put him as last
		err = ReaderDb.Get(&res, `
			SELECT count(*)
			FROM validator_queue_deposits
		`)
		if err == nil {
			return res, nil
		}
	}
	if err != nil {
		return res, err
	}
	err = ReaderDb.Get(&res, `
	SELECT count(*)
	FROM validator_queue_deposits
	WHERE 
		COALESCE(activationeligibilityepoch, 0) < $1 OR 
		block_slot < $2 OR
		block_slot = $2 AND block_index < $3`, selected.ActivationEligibilityEpoch, selected.BlockSlot, selected.BlockIndex)
	return res, err
}

func GetValidatorNames(validators []uint64) (map[uint64]string, error) {
	logger.Infof("getting validator names for %d validators", len(validators))
	rows, err := ReaderDb.Query(`
		SELECT validatorindex, validator_names.name 
		FROM validators 
		LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
		WHERE validators.validatorindex = ANY($1) AND validator_names.name IS NOT NULL`, pq.Array(validators))

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

	latestFinalizedEpoch, err := GetLatestFinalizedEpoch()
	if err != nil {
		return nil, err
	}
	err = ReaderDb.Select(&blks, `
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
		(b.epoch <= $2) AS finalized
	FROM blocks b
		LEFT JOIN epochs e ON e.epoch = b.epoch
	WHERE b.epoch >= $1
	ORDER BY slot DESC;
`, latestEpoch, latestFinalizedEpoch)
	if err != nil {
		return nil, err
	}

	currentSlot := utils.TimeToSlot(uint64(time.Now().Unix()))

	epochMap := map[uint64]*types.SlotVizEpochs{}

	res := []*types.SlotVizEpochs{}

	for _, b := range blks {
		_, exists := epochMap[b.Epoch]
		if !exists {
			r := types.SlotVizEpochs{
				Epoch:          b.Epoch,
				Finalized:      b.Finalized,
				Particicpation: b.Globalparticipationrate,
				Slots:          []*types.SlotVizSlots{},
			}
			r.Slots = make([]*types.SlotVizSlots, utils.Config.Chain.ClConfig.SlotsPerEpoch)
			epochMap[b.Epoch] = &r
		}

		slotIndex := b.Slot - (b.Epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch)

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
		for i := uint64(0); i < utils.Config.Chain.ClConfig.SlotsPerEpoch; i++ {
			if epoch.Slots[i] == nil {
				status := "scheduled"
				slot := (epoch.Epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch) + i
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
	err = ReaderDb.Get(&block, `SELECT exec_block_number FROM blocks where slot >= $1 AND exec_block_number > 0 ORDER BY slot LIMIT 1`, slot)
	return
}

func SaveChartSeriesPoint(date time.Time, indicator string, value any) error {
	_, err := WriterDb.Exec(`INSERT INTO chart_series (time, indicator, value) VALUES($1, $2, $3) ON CONFLICT (time, indicator) DO UPDATE SET value = EXCLUDED.value`, date, indicator, value)
	if err != nil {
		return fmt.Errorf("error saving chart_series: %v: %w", indicator, err)
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

func GetWithdrawalsCountForQuery(query string) (uint64, error) {
	t0 := time.Now()
	defer func() {
		logger.WithFields(logrus.Fields{"duration": time.Since(t0)}).Infof("finished GetWithdrawalsCountForQuery")
	}()
	count := uint64(0)

	withdrawalsQuery := `
		SELECT COUNT(*) FROM (
			SELECT b.slot
			FROM blocks_withdrawals w
			INNER JOIN blocks b ON w.block_root = b.blockroot AND b.status = '1'
			%s
			LIMIT %d
		) a`

	var err error = nil

	trimmedQuery := strings.ToLower(strings.TrimPrefix(query, "0x"))
	if utils.IsEth1Address(query) {
		searchQuery := `WHERE w.address = $1`
		addr, decErr := hex.DecodeString(trimmedQuery)
		if err != nil {
			return 0, decErr
		}
		err = ReaderDb.Get(&count, fmt.Sprintf(withdrawalsQuery, searchQuery, WithdrawalsQueryLimit),
			addr)
	} else if uiQuery, parseErr := strconv.ParseUint(query, 10, 64); parseErr == nil {
		// Check whether the query can be used for a validator, slot or epoch search
		searchQuery := `
			WHERE w.validatorindex = $1
				OR w.block_slot = $1
				OR w.block_slot BETWEEN $1*$2 AND ($1+1)*$2-1`
		err = ReaderDb.Get(&count, fmt.Sprintf(withdrawalsQuery, searchQuery, WithdrawalsQueryLimit),
			uiQuery, utils.Config.Chain.ClConfig.SlotsPerEpoch)
	}

	if err != nil {
		return 0, err
	}

	return count, nil
}

func GetWithdrawals(query string, length, start uint64, orderBy, orderDir string) ([]*types.Withdrawals, error) {
	t0 := time.Now()
	defer func() {
		logger.WithFields(logrus.Fields{"duration": time.Since(t0)}).Infof("finished GetWithdrawals")
	}()
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

	withdrawalsQuery := `
		SELECT 
			w.block_slot as slot,
			w.withdrawalindex as index,
			w.validatorindex,
			w.address,
			w.amount
		FROM blocks_withdrawals w
		INNER JOIN blocks b ON w.block_root = b.blockroot AND b.status = '1'
		%s 
		ORDER BY %s %s
		LIMIT $1
		OFFSET $2`

	var err error = nil

	trimmedQuery := strings.ToLower(strings.TrimPrefix(query, "0x"))
	if trimmedQuery != "" {
		if utils.IsEth1Address(query) {
			searchQuery := `WHERE w.address = $3`
			addr, decErr := hex.DecodeString(trimmedQuery)
			if decErr != nil {
				return nil, decErr
			}
			err = ReaderDb.Select(&withdrawals, fmt.Sprintf(withdrawalsQuery, searchQuery, orderBy, orderDir),
				length, start, addr)
		} else if uiQuery, parseErr := strconv.ParseUint(query, 10, 64); parseErr == nil {
			// Check whether the query can be used for a validator, slot or epoch search
			searchQuery := `
				WHERE w.validatorindex = $3
					OR w.block_slot = $3
					OR w.block_slot BETWEEN $3*$4 AND ($3+1)*$4-1`
			err = ReaderDb.Select(&withdrawals, fmt.Sprintf(withdrawalsQuery, searchQuery, orderBy, orderDir),
				length, start, uiQuery, utils.Config.Chain.ClConfig.SlotsPerEpoch)
		}
	} else {
		err = ReaderDb.Select(&withdrawals, fmt.Sprintf(withdrawalsQuery, "", orderBy, orderDir), length, start)
	}

	if err != nil {
		return nil, err
	}

	return withdrawals, nil
}

func GetTotalAmountWithdrawn() (sum uint64, count uint64, err error) {
	var res = struct {
		Sum   uint64 `db:"sum"`
		Count uint64 `db:"count"`
	}{}
	lastExportedDay, err := GetLastExportedStatisticDay()
	if err != nil {
		return 0, 0, fmt.Errorf("error getting latest exported statistic day for withdrawals count: %w", err)
	}
	_, lastEpochOfDay := utils.GetFirstAndLastEpochForDay(lastExportedDay)
	cutoffSlot := (lastEpochOfDay * utils.Config.Chain.ClConfig.SlotsPerEpoch) + 1

	err = ReaderDb.Get(&res, `
		WITH today AS (
			SELECT
				COALESCE(SUM(w.amount), 0) as sum,
				COUNT(*) as count
			FROM blocks_withdrawals w
			INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
			WHERE w.block_slot >= $1
		),
		stats AS (
			SELECT
				COALESCE(SUM(withdrawals_amount_total), 0) as sum,
				COALESCE(SUM(withdrawals_total), 0) as count
			FROM validator_stats
			WHERE day = $2
		)
		SELECT
			today.sum + stats.sum as sum,
			today.count + stats.count as count
		FROM today, stats`, cutoffSlot, lastExportedDay)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, nil
		}
		return 0, 0, fmt.Errorf("error fetching total withdrawal count and amount: %w", err)
	}

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
	WHERE w.block_slot >= $1 AND w.block_slot < $2`, epoch*utils.Config.Chain.ClConfig.SlotsPerEpoch, (epoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch)
	return
}

// GetAddressWithdrawalTableData returns the withdrawal data for an address
func GetAddressWithdrawalTableData(address []byte, pageToken string, currency string) (*types.DataTableResponse, error) {
	const endOfWithdrawalsData = "End of withdrawals data"
	const limit = DefaultInfScrollRows

	var withdrawals []*types.Withdrawals
	var withdrawalIndex uint64
	var err error
	var nextPageToken string
	var emptyData = &types.DataTableResponse{
		Data:        make([][]interface{}, 0),
		PagingToken: "",
	}

	tmr := time.AfterFunc(REPORT_TIMEOUT, func() {
		logger.WithFields(logrus.Fields{
			"address":   address,
			"pageToken": pageToken,
		}).Warnf("%s call took longer than %v", utils.GetCurrentFuncName(), REPORT_TIMEOUT)
	})
	defer tmr.Stop()

	if pageToken == "" {
		// Start from the beginning
		withdrawalIndex, err = GetTotalWithdrawals()
		if err != nil {
			return emptyData, fmt.Errorf("error getting total withdrawals for address: %x, %w", address, err)
		}
	} else if pageToken == endOfWithdrawalsData {
		// Last page already shown, end the infinite scroll
		return emptyData, nil
	} else {
		withdrawalIndex, err = strconv.ParseUint(pageToken, 10, 64)
		if err != nil {
			return emptyData, fmt.Errorf("error parsing page token: %w", err)
		}
	}

	err = ReaderDb.Select(&withdrawals, `
	SELECT 
		w.block_slot as slot, 
		w.withdrawalindex as index, 
		w.validatorindex, 
		w.address, 
		w.amount 
	FROM blocks_withdrawals w
	INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
	WHERE w.address = $1 AND w.withdrawalindex <= $2
	ORDER BY w.withdrawalindex DESC LIMIT $3`, address, withdrawalIndex, limit+1)
	if err != nil {
		if err == sql.ErrNoRows {
			return emptyData, nil
		}
		return emptyData, fmt.Errorf("error getting blocks_withdrawals for address: %x: %w", address, err)
	}
	// Get the next page token and remove that withdrawal from the results
	nextPageToken = endOfWithdrawalsData
	if len(withdrawals) == int(limit+1) {
		nextPageToken = fmt.Sprintf("%d", withdrawals[limit].Index)
		withdrawals = withdrawals[:limit]
	}

	withdrawalsData := make([][]interface{}, len(withdrawals))
	for i, w := range withdrawals {
		withdrawalsData[i] = []interface{}{
			utils.FormatEpoch(utils.EpochOfSlot(w.Slot)),
			utils.FormatBlockSlot(w.Slot),
			utils.FormatTimestamp(utils.SlotToTime(w.Slot).Unix()),
			utils.FormatValidator(w.ValidatorIndex),
			utils.FormatClCurrency(w.Amount, currency, 6, true, false, false, true),
		}
	}

	data := &types.DataTableResponse{
		Data:        withdrawalsData,
		PagingToken: nextPageToken,
	}

	return data, nil
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
	WHERE w.block_slot >= $1 AND w.block_slot < $2 ORDER BY w.withdrawalindex`, epoch*utils.Config.Chain.ClConfig.SlotsPerEpoch, (epoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch)
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
	ORDER BY w.withdrawalindex`, pq.Array(validators), fromEpoch, toEpoch, utils.Config.Chain.ClConfig.SlotsPerEpoch)
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
	ORDER BY w.block_slot / $4 DESC LIMIT 100`, pq.Array(validator), startEpoch*utils.Config.Chain.ClConfig.SlotsPerEpoch, endEpoch*utils.Config.Chain.ClConfig.SlotsPerEpoch+utils.Config.Chain.ClConfig.SlotsPerEpoch-1, utils.Config.Chain.ClConfig.SlotsPerEpoch)
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
	// #TODO: BIDS-2879
	if true {
		return 0, nil
	}
	var total uint64

	err := ReaderDb.Get(&total, `
	/*+
	BitmapScan(w)
	NestLoop(b w)
	*/
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

func GetDashboardWithdrawals(validators []uint64, limit uint64, offset uint64, orderBy string, orderDir string) ([]*types.Withdrawals, error) {
	var withdrawals []*types.Withdrawals
	if limit == 0 {
		limit = 100
	}
	validatorFilter := pq.Array(validators)
	err := ReaderDb.Select(&withdrawals, fmt.Sprintf(`
		/*+
		BitmapScan(w)
		NestLoop(b w)
		*/
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

func GetTotalWithdrawalsCount(validators []uint64) (uint64, error) {
	var count uint64
	validatorFilter := pq.Array(validators)
	lastExportedDay, err := GetLastExportedStatisticDay()
	if err != nil && err != ErrNoStats {
		return 0, fmt.Errorf("error getting latest exported statistic day for withdrawals count: %w", err)
	}

	cutoffSlot := uint64(0)
	if err == nil {
		_, lastEpochOfDay := utils.GetFirstAndLastEpochForDay(lastExportedDay)
		cutoffSlot = (lastEpochOfDay * utils.Config.Chain.ClConfig.SlotsPerEpoch) + 1
	}

	err = ReaderDb.Get(&count, `
		WITH today AS (
			SELECT COUNT(*) as count
			FROM blocks_withdrawals w
			INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
			WHERE w.validatorindex = ANY($1) AND w.block_slot >= $2
		),
		stats AS (
			SELECT COALESCE(SUM(withdrawals_total), 0) as count
			FROM validator_stats
			WHERE validatorindex = ANY($1) AND day = $3
		)
		SELECT today.count + stats.count
		FROM today, stats`, validatorFilter, cutoffSlot, lastExportedDay)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("error getting dashboard validator blocks_withdrawals count for validators: %d: %w", validators, err)
	}

	return count, nil
}

func GetLastWithdrawalEpoch(validators []uint64) (map[uint64]uint64, error) {
	var dbResponse []struct {
		ValidatorIndex     uint64 `db:"validatorindex"`
		LastWithdrawalSlot uint64 `db:"last_withdawal_slot"`
	}

	res := make(map[uint64]uint64)
	err := ReaderDb.Select(&dbResponse, `
		SELECT w.validatorindex as validatorindex, COALESCE(max(block_slot), 0) as last_withdawal_slot
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

	for _, row := range dbResponse {
		res[row.ValidatorIndex] = row.LastWithdrawalSlot / utils.Config.Chain.ClConfig.SlotsPerEpoch
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

func GetBLSChangesCountForQuery(query string) (uint64, error) {
	count := uint64(0)

	blsQuery := `
		SELECT COUNT(*) FROM (
			SELECT b.slot
			FROM blocks_bls_change bls
			INNER JOIN blocks b ON bls.block_root = b.blockroot AND b.status = '1'
			%s
			LIMIT %d
		) a
		`

	trimmedQuery := strings.ToLower(strings.TrimPrefix(query, "0x"))
	var err error = nil

	if utils.IsHash(query) {
		searchQuery := `WHERE bls.pubkey = $1`
		pubkey, decErr := hex.DecodeString(trimmedQuery)
		if decErr != nil {
			return 0, decErr
		}
		err = ReaderDb.Get(&count, fmt.Sprintf(blsQuery, searchQuery, BlsChangeQueryLimit),
			pubkey)
	} else if uiQuery, parseErr := strconv.ParseUint(query, 10, 64); parseErr == nil {
		// Check whether the query can be used for a validator, slot or epoch search
		searchQuery := `
			WHERE bls.validatorindex = $1			
				OR bls.block_slot = $1
				OR bls.block_slot BETWEEN $1*$2 AND ($1+1)*$2-1`
		err = ReaderDb.Get(&count, fmt.Sprintf(blsQuery, searchQuery, BlsChangeQueryLimit),
			uiQuery, utils.Config.Chain.ClConfig.SlotsPerEpoch)
	}
	if err != nil {
		return 0, err
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

	blsQuery := `
		SELECT 
			bls.block_slot as slot,
			bls.validatorindex,
			bls.signature,
			bls.pubkey,
			bls.address
		FROM blocks_bls_change bls
		INNER JOIN blocks b ON bls.block_root = b.blockroot AND b.status = '1'
		%s
		ORDER BY bls.%s %s
		LIMIT $1
		OFFSET $2`

	trimmedQuery := strings.ToLower(strings.TrimPrefix(query, "0x"))
	var err error = nil

	if trimmedQuery != "" {
		if utils.IsHash(query) {
			searchQuery := `WHERE bls.pubkey = $3`
			pubkey, decErr := hex.DecodeString(trimmedQuery)
			if decErr != nil {
				return nil, decErr
			}
			err = ReaderDb.Select(&blsChange, fmt.Sprintf(blsQuery, searchQuery, orderBy, orderDir),
				length, start, pubkey)
		} else if uiQuery, parseErr := strconv.ParseUint(query, 10, 64); parseErr == nil {
			// Check whether the query can be used for a validator, slot or epoch search
			searchQuery := `
				WHERE bls.validatorindex = $3			
					OR bls.block_slot = $3
					OR bls.block_slot BETWEEN $3*$4 AND ($3+1)*$4-1`
			err = ReaderDb.Select(&blsChange, fmt.Sprintf(blsQuery, searchQuery, orderBy, orderDir),
				length, start, uiQuery, utils.Config.Chain.ClConfig.SlotsPerEpoch)
		}
		if err != nil {
			return nil, err
		}
	} else {
		err := ReaderDb.Select(&blsChange, fmt.Sprintf(blsQuery, "", orderBy, orderDir), length, start)
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
		validators.withdrawalcredentials LIKE '\x01' || '%'::bytea AND ((stats.end_effective_balance = $1 AND stats.end_balance > $1) OR (validators.withdrawableepoch <= $2 AND stats.end_balance > 0));`, utils.Config.Chain.ClConfig.MaxEffectiveBalance, epoch)
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

func GetLastExportedStatisticDay() (uint64, error) {
	var lastStatsDay sql.NullInt64
	err := ReaderDb.Get(&lastStatsDay, "SELECT MAX(day) FROM validator_stats_status WHERE status")

	if err != nil {
		return 0, fmt.Errorf("error getting lastStatsDay %v", err)
	}

	if !lastStatsDay.Valid {
		return 0, ErrNoStats
	}
	return uint64(lastStatsDay.Int64), nil
}

// GetValidatorIncomePerformance gets all rewards of a validator in WEI for 1d, 7d, 365d and total
func GetValidatorIncomePerformance(validators []uint64, incomePerformance *types.ValidatorIncomePerformance) error {
	validatorsPQArray := pq.Array(validators)
	return ReaderDb.Get(incomePerformance, `
		SELECT 
			COALESCE(SUM(cl_performance_1d    ), 0)*1e9 AS cl_performance_wei_1d,
			COALESCE(SUM(cl_performance_7d    ), 0)*1e9 AS cl_performance_wei_7d,
			COALESCE(SUM(cl_performance_31d   ), 0)*1e9 AS cl_performance_wei_31d,
			COALESCE(SUM(cl_performance_365d  ), 0)*1e9 AS cl_performance_wei_365d,
			COALESCE(SUM(cl_performance_total ), 0)*1e9 AS cl_performance_wei_total,
			COALESCE(SUM(mev_performance_1d   ), 0)     AS el_performance_wei_1d,
			COALESCE(SUM(mev_performance_7d   ), 0)     AS el_performance_wei_7d,
			COALESCE(SUM(mev_performance_31d  ), 0)     AS el_performance_wei_31d,
			COALESCE(SUM(mev_performance_365d ), 0)     AS el_performance_wei_365d,
			COALESCE(SUM(mev_performance_total), 0)     AS el_performance_wei_total
		FROM validator_performance 
		WHERE validatorindex = ANY($1)`, validatorsPQArray)
}

func GetTotalValidatorDeposits(validators []uint64, totalDeposits *uint64) error {
	validatorsPQArray := pq.Array(validators)
	return ReaderDb.Get(totalDeposits, `
		SELECT 
			COALESCE(SUM(amount), 0) 
		FROM blocks_deposits d
		INNER JOIN blocks b ON b.blockroot = d.block_root AND b.status = '1' 
		WHERE publickey IN (SELECT pubkey FROM validators WHERE validatorindex = ANY($1))
	`, validatorsPQArray)
}

func GetFirstActivationEpoch(validators []uint64, firstActivationEpoch *uint64) error {
	validatorsPQArray := pq.Array(validators)
	return ReaderDb.Get(firstActivationEpoch, `
		SELECT 
			activationepoch
		FROM validators
		WHERE validatorindex = ANY($1) 
		ORDER BY activationepoch LIMIT 1
	`, validatorsPQArray)
}

func GetValidatorDepositsForSlots(validators []uint64, fromSlot uint64, toSlot uint64, deposits *uint64) error {
	validatorsPQArray := pq.Array(validators)
	return ReaderDb.Get(deposits, `
		SELECT 
			COALESCE(SUM(amount), 0) 
		FROM blocks_deposits d
		INNER JOIN blocks b ON b.blockroot = d.block_root AND b.status = '1' and b.slot >= $2 and b.slot <= $3
		WHERE publickey IN (SELECT pubkey FROM validators WHERE validatorindex = ANY($1))
	`, validatorsPQArray, fromSlot, toSlot)
}

func GetValidatorWithdrawalsForSlots(validators []uint64, fromSlot uint64, toSlot uint64, withdrawals *uint64) error {
	validatorsPQArray := pq.Array(validators)
	return ReaderDb.Get(withdrawals, `
		SELECT 
			COALESCE(SUM(amount), 0) 
		FROM blocks_withdrawals d
		INNER JOIN blocks b ON b.blockroot = d.block_root AND b.status = '1' and b.slot >= $2 and b.slot <= $3        
		WHERE validatorindex = ANY($1)
	`, validatorsPQArray, fromSlot, toSlot)
}

func GetValidatorBalanceForDay(validators []uint64, day uint64, balance *uint64) error {
	validatorsPQArray := pq.Array(validators)
	return ReaderDb.Get(balance, `
		SELECT 
			COALESCE(SUM(end_balance), 0) 
		FROM validator_stats     
		WHERE validatorindex = ANY($1) AND day = $2
	`, validatorsPQArray, day)
}

func GetValidatorActivationBalance(validators []uint64, balance *uint64) error {
	if len(validators) == 0 {
		return fmt.Errorf("passing empty validator array is unsupported")
	}

	validatorsPQArray := pq.Array(validators)
	return ReaderDb.Get(balance, `
		SELECT 
			SUM(balanceactivation)
		FROM validators     
		WHERE validatorindex = ANY($1)
	`, validatorsPQArray)
}

func GetValidatorPropsosals(validators []uint64, proposals *[]types.ValidatorProposalInfo) error {
	validatorsPQArray := pq.Array(validators)

	return ReaderDb.Select(proposals, `
		SELECT
			slot,
			status,
			COALESCE(exec_block_number, 0) as exec_block_number
		FROM blocks
		WHERE proposer = ANY($1)
		ORDER BY slot ASC
		`, validatorsPQArray)
}

func GetMissedSlots(slots []uint64) ([]uint64, error) {
	slotsPQArray := pq.Array(slots)
	missed := []uint64{}

	err := ReaderDb.Select(&missed, `
		SELECT
			slot
		FROM blocks
		WHERE slot = ANY($1) AND status = '2'
		`, slotsPQArray)

	return missed, err
}

func GetMissedSlotsMap(slots []uint64) (map[uint64]bool, error) {
	missedSlots, err := GetMissedSlots(slots)
	if err != nil {
		return nil, err
	}
	missedSlotsMap := make(map[uint64]bool, len(missedSlots))
	for _, slot := range missedSlots {
		missedSlotsMap[slot] = true
	}
	return missedSlotsMap, nil
}

func GetOrphanedSlots(slots []uint64) ([]uint64, error) {
	slotsPQArray := pq.Array(slots)
	orphaned := []uint64{}

	err := ReaderDb.Select(&orphaned, `
		SELECT
			slot
		FROM blocks
		WHERE slot = ANY($1) AND status = '3'
		`, slotsPQArray)

	return orphaned, err
}

func GetOrphanedSlotsMap(slots []uint64) (map[uint64]bool, error) {
	orphanedSlots, err := GetOrphanedSlots(slots)
	if err != nil {
		return nil, err
	}
	orphanedSlotsMap := make(map[uint64]bool, len(orphanedSlots))
	for _, slot := range orphanedSlots {
		orphanedSlotsMap[slot] = true
	}
	return orphanedSlotsMap, nil
}

func GetBlockStatus(block int64, latestFinalizedEpoch uint64, epochInfo *types.EpochInfo) error {
	return ReaderDb.Get(epochInfo, `
				SELECT (epochs.epoch <= $2) AS finalized, epochs.globalparticipationrate 
				FROM blocks 
				LEFT JOIN epochs ON blocks.epoch = epochs.epoch 
				WHERE blocks.exec_block_number = $1 
				AND blocks.status='1'`,
		block, latestFinalizedEpoch)
}

// Returns the participation rate for every slot between startSlot and endSlot (both inclusive) as a map with the slot as key
//
// If a slot is missed, the map will not contain an entry for it
func GetSyncParticipationBySlotRange(startSlot, endSlot uint64) (map[uint64]uint64, error) {

	rows := []struct {
		Slot         uint64
		Participated uint64
	}{}

	err := ReaderDb.Select(&rows, `SELECT slot, syncaggregate_participation * $1 AS participated FROM blocks WHERE slot >= $2 AND slot <= $3 AND status = '1'`,
		utils.Config.Chain.ClConfig.SyncCommitteeSize,
		startSlot,
		endSlot)

	if err != nil {
		return nil, err
	}

	ret := make(map[uint64]uint64)

	for _, row := range rows {
		ret[row.Slot] = row.Participated
	}

	return ret, nil
}

// Should be used when retrieving data for a very large amount of validators (for the notifications process)
func GetValidatorAttestationHistoryForNotifications(startEpoch uint64, endEpoch uint64) (map[types.Epoch]map[types.ValidatorIndex]bool, error) {
	// first retrieve activation & exit epoch for all validators
	activityData := []struct {
		ValidatorIndex  types.ValidatorIndex
		ActivationEpoch types.Epoch
		ExitEpoch       types.Epoch
	}{}

	err := ReaderDb.Select(&activityData, "SELECT validatorindex, activationepoch, exitepoch FROM validators ORDER BY validatorindex;")
	if err != nil {
		return nil, fmt.Errorf("error retrieving activation & exit epoch for validators: %w", err)
	}

	logger.Info("retrieved activation & exit epochs")

	// next retrieve all attestation data from the db (need to retrieve data for the endEpoch+1 epoch as that could still contain attestations for the endEpoch)
	firstSlot := startEpoch * utils.Config.Chain.ClConfig.SlotsPerEpoch
	lastSlot := ((endEpoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch - 1)
	lastQuerySlot := ((endEpoch+2)*utils.Config.Chain.ClConfig.SlotsPerEpoch - 1)

	rows, err := ReaderDb.Query(`SELECT 
	blocks_attestations.slot, 
	validators 
	FROM blocks_attestations 
	LEFT JOIN blocks ON blocks_attestations.block_root = blocks.blockroot WHERE
	blocks_attestations.block_slot >= $1 AND blocks_attestations.block_slot <= $2 AND blocks.status = '1' ORDER BY block_slot`, firstSlot, lastQuerySlot)
	if err != nil {
		return nil, fmt.Errorf("error retrieving attestation data from the db: %w", err)
	}
	defer rows.Close()

	logger.Info("retrieved attestation raw data")

	// next process the data and fill up the epoch participation
	// validators that participated in an epoch will have the flag set to true
	// validators that missed their participation will have it set to false
	epochParticipation := make(map[types.Epoch]map[types.ValidatorIndex]bool)
	for rows.Next() {
		var slot types.Slot
		var attestingValidators pq.Int64Array

		err := rows.Scan(&slot, &attestingValidators)
		if err != nil {
			return nil, fmt.Errorf("error scanning attestation data: %w", err)
		}

		if slot < types.Slot(firstSlot) || slot > types.Slot(lastSlot) {
			continue
		}

		epoch := types.Epoch(utils.EpochOfSlot(uint64(slot)))

		participation := epochParticipation[epoch]

		if participation == nil {
			epochParticipation[epoch] = make(map[types.ValidatorIndex]bool)

			// logger.Infof("seeding validator duties for epoch %v", epoch)
			for _, data := range activityData {
				if data.ActivationEpoch <= epoch && epoch < data.ExitEpoch {
					epochParticipation[epoch][types.ValidatorIndex(data.ValidatorIndex)] = false
				}
			}

			participation = epochParticipation[epoch]
		}

		for _, validator := range attestingValidators {
			participation[types.ValidatorIndex(validator)] = true
		}
	}

	return epochParticipation, nil
}
