package db

import (
	"bytes"
	"database/sql"
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
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/sirupsen/logrus"

	"github.com/jackc/pgx/v4/pgxpool"
)

var DBPGX *pgxpool.Conn

// DB is a pointer to the explorer-database
var DB *sqlx.DB

var logger = logrus.New().WithField("module", "db")

func mustInitDB(username, password, host, port, name string) *sqlx.DB {
	dbConn, err := sqlx.Open("pgx", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port, name))
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
	err = dbConn.Ping()
	if err != nil {
		logger.Fatal(err)
	}
	dbConnectionTimeout.Stop()

	dbConn.SetConnMaxIdleTime(time.Second * 30)
	dbConn.SetConnMaxLifetime(time.Second * 60)

	return dbConn
}

func MustInitDB(username, password, host, port, name string) {
	DB = mustInitDB(username, password, host, port, name)
}

func GetEth1Deposits(address string, length, start uint64) ([]*types.EthOneDepositsData, error) {
	deposits := []*types.EthOneDepositsData{}

	err := DB.Select(&deposits, `
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
		}
	}
	if !hasColumn {
		orderBy = "block_ts"
	}

	var err error
	var totalCount uint64
	if query != "" {
		err = DB.Get(&totalCount, `
			SELECT COUNT(*) FROM eth1_deposits as eth1
			WHERE 
				ENCODE(eth1.publickey::bytea, 'hex') LIKE LOWER($1)
				OR ENCODE(eth1.withdrawal_credentials::bytea, 'hex') LIKE LOWER($1)
				OR ENCODE(eth1.from_address::bytea, 'hex') LIKE LOWER($1)
				OR ENCODE(tx_hash::bytea, 'hex') LIKE LOWER($1)
				OR CAST(eth1.block_number AS text) LIKE LOWER($1)`, query+"%")
	} else {
		err = DB.Get(&totalCount, "SELECT COUNT(*) FROM eth1_deposits")
	}
	if err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	}

	if query != "" {
		err = DB.Select(&deposits, fmt.Sprintf(`
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
			ENCODE(eth1.publickey::bytea, 'hex') LIKE LOWER($5)
			OR ENCODE(eth1.withdrawal_credentials::bytea, 'hex') LIKE LOWER($5)
			OR ENCODE(eth1.from_address::bytea, 'hex') LIKE LOWER($5)
			OR ENCODE(tx_hash::bytea, 'hex') LIKE LOWER($5)
			OR CAST(eth1.block_number AS text) LIKE LOWER($5)
		ORDER BY %s %s
		LIMIT $1
		OFFSET $2`, orderBy, orderDir), length, start, latestEpoch, validatorOnlineThresholdSlot, query+"%")
	} else {
		err = DB.Select(&deposits, fmt.Sprintf(`
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
	err := DB.Get(&deposits, `
	SELECT 
		Count(*)
	FROM 
		eth1_deposits
	`)
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
		err = DB.Get(&totalCount, `
		SELECT
			COUNT(from_address)
			FROM
				(
					SELECT
						from_address
					FROM
						eth1_deposits as eth1
					WHERE
						ENCODE(eth1.from_address::bytea, 'hex') LIKE LOWER($1)
						GROUP BY from_address
				) as count
		`, query+"%")
	} else {
		err = DB.Get(&totalCount, "SELECT COUNT(*) FROM (SELECT from_address FROM eth1_deposits GROUP BY from_address) as count")
	}
	if err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	}

	err = DB.Select(&deposits, fmt.Sprintf(`
		SELECT 
			from_address,
			SUM(amount) as amount,
			COUNT(CASE WHEN valid_signature = 't' THEN 1 END) as validcount,
			COUNT(CASE WHEN valid_signature = 'f' THEN 1 END) as invalidcount,
			COUNT(CASE WHEN v.slashed = 't' THEN 1 END) as slashedcount,
			COUNT(pubkey) as totalcount,
			COUNT(CASE WHEN v.slashed = 'f' and v.exitepoch > $3 and activationepoch < $3 THEN 1 END) as activecount,
			COUNT(CASE WHEN activationepoch > $3 THEN 1 END) as pendingcount,
			COUNT(CASE WHEN v.slashed = 'f' and v.exitepoch < $3 THEN 1 END) as voluntary_exit_count
		FROM
			eth1_deposits as eth1
		LEFT JOIN
			(
				SELECT 
					pubkey,
					slashed,
					exitepoch,
					activationepoch,
					COALESCE(validator_names.name, '') AS name
				FROM validators
				LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
			) as v
		ON
			v.pubkey = eth1.publickey
		WHERE
			ENCODE(eth1.from_address::bytea, 'hex') LIKE LOWER($4)
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
	// ENCODE(publickey::bytea, 'hex') LIKE $3 OR ENCODE(withdrawalcredentials::bytea, 'hex') LIKE $3 OR
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
		err := DB.Select(&deposits, fmt.Sprintf(`
			SELECT 
				blocks_deposits.block_slot,
				blocks_deposits.block_index,
				blocks_deposits.proof,
				blocks_deposits.publickey,
				blocks_deposits.withdrawalcredentials,
				blocks_deposits.amount,
				blocks_deposits.signature
			FROM blocks_deposits
			WHERE ENCODE(publickey::bytea, 'hex') LIKE $3 OR ENCODE(withdrawalcredentials::bytea, 'hex') LIKE $3 OR CAST(block_slot as varchar) LIKE $3
			ORDER BY %s %s
			LIMIT $1
			OFFSET $2`, orderBy, orderDir), length, start, query+"%")
		if err != nil {
			return nil, err
		}
	} else {
		err := DB.Select(&deposits, fmt.Sprintf(`
			SELECT 
				blocks_deposits.block_slot,
				blocks_deposits.block_index,
				blocks_deposits.proof,
				blocks_deposits.publickey,
				blocks_deposits.withdrawalcredentials,
				blocks_deposits.amount,
				blocks_deposits.signature
			FROM blocks_deposits
			ORDER BY %s %s
			LIMIT $1
			OFFSET $2`, orderBy, orderDir), length, start)
		if err != nil {
			return nil, err
		}
	}

	return deposits, nil
}

func GetEth2DepositsCount() (uint64, error) {
	deposits := uint64(0)

	err := DB.Get(&deposits, `
	SELECT 
		Count(*)
	FROM 
		blocks_deposits
	`)
	if err != nil {
		return 0, err
	}

	return deposits, nil
}
func GetSlashingCount() (uint64, error) {
	slashings := uint64(0)

	err := DB.Get(&slashings, `
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
	err := DB.Get(&epoch, "SELECT COALESCE(MAX(epoch), 0) FROM epochs")

	if err != nil {
		return 0, fmt.Errorf("error retrieving latest epoch from DB: %v", err)
	}

	return epoch, nil
}

// GetAllEpochs will return a collection of all of the epochs from the database
func GetAllEpochs() ([]uint64, error) {
	var epochs []uint64
	err := DB.Select(&epochs, "SELECT epoch FROM epochs ORDER BY epoch")

	if err != nil {
		return nil, fmt.Errorf("error retrieving all epochs from DB: %v", err)
	}

	return epochs, nil
}

// GetLastPendingAndProposedBlocks will return all proposed and pending blocks (ignores missed slots) from the database
func GetLastPendingAndProposedBlocks(startEpoch, endEpoch uint64) ([]*types.MinimalBlock, error) {
	var blocks []*types.MinimalBlock

	err := DB.Select(&blocks, "SELECT epoch, slot, blockroot FROM blocks WHERE epoch >= $1 AND epoch <= $2 AND blockroot != '\x01' ORDER BY slot DESC", startEpoch, endEpoch)

	if err != nil {
		return nil, fmt.Errorf("error retrieving last blocks (%v-%v) from DB: %v", startEpoch, endEpoch, err)
	}

	return blocks, nil
}

// GetBlocks will return all blocks for a range of epochs from the database
func GetBlocks(startEpoch, endEpoch uint64) ([]*types.MinimalBlock, error) {
	var blocks []*types.MinimalBlock

	err := DB.Select(&blocks, "SELECT epoch, slot, blockroot, parentroot FROM blocks WHERE epoch >= $1 AND epoch <= $2 AND length(blockroot) = 32 ORDER BY slot DESC", startEpoch, endEpoch)

	if err != nil {
		return nil, fmt.Errorf("error retrieving blocks for epoch %v to %v from DB: %v", startEpoch, endEpoch, err)
	}

	return blocks, nil
}

// GetValidatorPublicKey will return the public key for a specific validator from the database
func GetValidatorPublicKey(index uint64) ([]byte, error) {
	var publicKey []byte
	err := DB.Get(&publicKey, "SELECT pubkey FROM validators WHERE validatorindex = $1", index)

	return publicKey, err
}

// GetValidatorIndex will return the validator-index for a public key from the database
func GetValidatorIndex(publicKey []byte) (uint64, error) {
	var index uint64
	err := DB.Get(&index, "SELECT validatorindex FROM validators WHERE pubkey = $1", publicKey)

	return index, err
}

// GetValidatorDeposits will return eth1- and eth2-deposits for a public key from the database
func GetValidatorDeposits(publicKey []byte) (*types.ValidatorDeposits, error) {
	deposits := &types.ValidatorDeposits{}
	err := DB.Select(&deposits.Eth1Deposits, `
		SELECT tx_hash, tx_input, tx_index, block_number, EXTRACT(epoch FROM block_ts)::INT as block_ts, from_address, publickey, withdrawal_credentials, amount, signature, merkletree_index, valid_signature
		FROM eth1_deposits WHERE publickey = $1 ORDER BY block_number ASC`, publicKey)
	if err != nil {
		return nil, err
	}
	err = DB.Select(&deposits.Eth2Deposits, "SELECT * FROM blocks_deposits WHERE publickey = $1", publicKey)
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

	tx, err := DB.Begin()
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

	tx, err := DB.Begin()
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
	enteringValidatorsCount := len(validators.ActivationPublicKeys)
	exitingValidatorsCount := len(validators.ExitPublicKeys)
	_, err := DB.Exec(`
		INSERT INTO queue (ts, entering_validators_count, exiting_validators_count)
		VALUES (date_trunc('hour', now()), $1, $2)
		ON CONFLICT (ts) DO UPDATE SET
			entering_validators_count = excluded.entering_validators_count, 
			exiting_validators_count = excluded.exiting_validators_count`,
		enteringValidatorsCount, exitingValidatorsCount)
	return err
}

func SaveBlock(block *types.Block) error {

	blocksMap := make(map[uint64]map[string]*types.Block)
	if blocksMap[block.Slot] == nil {
		blocksMap[block.Slot] = make(map[string]*types.Block)
	}
	blocksMap[block.Slot][fmt.Sprintf("%x", block.BlockRoot)] = block

	tx, err := DB.Begin()
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
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("error starting db transactions: %v", err)
	}
	defer tx.Rollback()

	logger.Infof("starting export of epoch %v", data.Epoch)
	start := time.Now()

	logger.Infof("exporting block data")
	err = saveBlocks(data.Blocks, tx)
	if err != nil {
		logger.Fatalf("error saving blocks to db: %v", err)
		return fmt.Errorf("error saving blocks to db: %v", err)
	}

	logger.Infof("exporting validators data")
	err = saveValidators(data.Epoch, data.Validators, tx)
	if err != nil {
		return fmt.Errorf("error saving validators to db: %v", err)
	}

	logger.Infof("exporting proposal assignments data")
	err = saveValidatorProposalAssignments(data.Epoch, data.ValidatorAssignmentes.ProposerAssignments, tx)
	if err != nil {
		return fmt.Errorf("error saving validator assignments to db: %v", err)
	}

	logger.Infof("exporting attestation assignments data")
	err = saveValidatorAttestationAssignments(data.Epoch, data.ValidatorAssignmentes.AttestorAssignments, tx)
	if err != nil {
		return fmt.Errorf("error saving validator assignments to db: %v", err)
	}

	logger.Infof("exporting validator balance data")
	err = saveValidatorBalances(data.Epoch, data.Validators, tx)
	if err != nil {
		return fmt.Errorf("error saving validator balances to db: %v", err)
	}

	// only export recent validator balances if the epoch is within the threshold
	if uint64(utils.TimeToEpoch(time.Now())) > data.Epoch-5 {
		logger.Infof("exporting recent validator balance data")
		err = saveValidatorBalancesRecent(data.Epoch, data.Validators, tx)
		if err != nil {
			return fmt.Errorf("error saving recent validator balances to db: %v", err)
		}
	} else {
		logger.Infof("skipping export of recent validator balance data")
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
		return fmt.Errorf("error executing save epoch statement: %v", err)
	}

	err = saveGraffitiwall(data.Blocks, tx)
	if err != nil {
		return fmt.Errorf("error saving graffitiwall: %v", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing db transaction: %v", err)
	}

	logger.Infof("export of epoch %v completed, took %v", data.Epoch, time.Since(start))
	return nil
}

func saveGraffitiwall(blocks map[uint64]map[string]*types.Block, tx *sql.Tx) error {

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

	graffitiWallRegex := regexp.MustCompile("graffitiwall:([0-9]{1,3}):([0-9]{1,3}):#([0-9a-fA-F]{6})")

	for _, slot := range blocks {
		for _, block := range slot {
			matches := graffitiWallRegex.FindStringSubmatch(string(block.Graffiti))
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

func saveValidators(epoch uint64, validators []*types.Validator, tx *sql.Tx) error {
	batchSize := 4000
	var lenActivatedValidators int
	var lastActivatedValidatorIdx uint64

	for _, v := range validators {
		if !(v.ActivationEpoch <= epoch && epoch < v.ExitEpoch) {
			continue
		}
		lenActivatedValidators++
		if v.Index < lastActivatedValidatorIdx {
			continue
		}
		lastActivatedValidatorIdx = v.Index
	}

	for b := 0; b < len(validators); b += batchSize {
		start := b
		end := b + batchSize
		if len(validators) < end {
			end = len(validators)
		}

		valueStrings := make([]string, 0, batchSize)
		valueArgs := make([]interface{}, 0, batchSize*13)
		for i, v := range validators[start:end] {

			if v.WithdrawableEpoch == 18446744073709551615 {
				v.WithdrawableEpoch = 9223372036854775807
			}
			if v.ExitEpoch == 18446744073709551615 {
				v.ExitEpoch = 9223372036854775807
			}
			if v.ActivationEligibilityEpoch == 18446744073709551615 {
				v.ActivationEligibilityEpoch = 9223372036854775807
			}
			if v.ActivationEpoch == 18446744073709551615 {
				v.ActivationEpoch = 9223372036854775807
			}
			if v.ActivationEligibilityEpoch < 9223372036854775807 && v.ActivationEpoch == 9223372036854775807 {
				// see: https://github.com/ethereum/eth2.0-specs/blob/master/specs/phase0/beacon-chain.md#get_validator_churn_limit
				// validator_churn_limit = max(MIN_PER_EPOCH_CHURN_LIMIT, len(active_validator_indices) // CHURN_LIMIT_QUOTIENT)
				// validator_churn_limit = max(4, len(active_set) / 2**16)
				// validator.activationepoch = epoch + validator.positioninactivationqueue / validator_churn_limit
				// note: this is only an estimation
				positionInActivationQueue := v.Index - lastActivatedValidatorIdx
				churnLimit := float64(lenActivatedValidators) / 65536
				if churnLimit < 4 {
					churnLimit = 4
				}
				if v.ActivationEligibilityEpoch > epoch {
					v.ActivationEpoch = v.ActivationEligibilityEpoch + uint64(float64(positionInActivationQueue)/churnLimit)
				} else {
					v.ActivationEpoch = epoch + uint64(float64(positionInActivationQueue)/churnLimit)
				}
			}

			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*14+1, i*14+2, i*14+3, i*14+4, i*14+5, i*14+6, i*14+7, i*14+8, i*14+9, i*14+10, i*14+11, i*14+12, i*14+13, i*14+14))
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
			pubkeyhex
		) 
		VALUES %s
		ON CONFLICT (validatorindex) DO UPDATE SET 
			pubkey                     = EXCLUDED.pubkey,
			withdrawableepoch          = EXCLUDED.withdrawableepoch,
			withdrawalcredentials      = EXCLUDED.withdrawalcredentials,
			balance                    = EXCLUDED.balance,
			effectivebalance           = EXCLUDED.effectivebalance,
			slashed                    = EXCLUDED.slashed,
			activationeligibilityepoch = EXCLUDED.activationeligibilityepoch,
			activationepoch            = EXCLUDED.activationepoch,
			exitepoch                  = EXCLUDED.exitepoch,
			balance1d                  = EXCLUDED.balance1d,
			balance7d                  = EXCLUDED.balance7d,
			balance31d                 = EXCLUDED.balance31d,
			pubkeyhex                  = EXCLUDED.pubkeyhex`, strings.Join(valueStrings, ","))
		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			return err
		}

		logger.Infof("saving validator batch %v completed", b)
	}

	logger.Infof("saving validator status")
	var latestBlock uint64
	err := DB.Get(&latestBlock, "SELECT COALESCE(MAX(slot), 0) FROM blocks WHERE status = '1'")
	if err != nil {
		return err
	}

	thresholdSlot := latestBlock - 64
	if latestBlock < 64 {
		thresholdSlot = 0
	}

	s := time.Now()
	_, err = tx.Exec(`UPDATE validators SET status = CASE 
				WHEN exitepoch <= $1 and slashed then 'slashed'
				WHEN exitepoch <= $1 then 'exited'
				WHEN activationeligibilityepoch = 9223372036854775807 then 'deposited'
				WHEN activationepoch > $1 then 'pending'
				WHEN slashed and activationepoch < $1 and (lastattestationslot < $2 OR lastattestationslot is null) then 'slashing_offline'
				WHEN slashed then 'slashing_online'
				WHEN exitepoch < 9223372036854775807 and (lastattestationslot < $2 OR lastattestationslot is null) then 'exiting_offline'
				WHEN exitepoch < 9223372036854775807 then 'exiting_online'
				WHEN activationepoch < $1 and (lastattestationslot < $2 OR lastattestationslot is null) then 'active_offline' 
				ELSE 'active_online'
			END`, latestBlock/32, thresholdSlot)
	if err != nil {
		return err
	}
	logger.Infof("saving validator status completed, took %v", time.Since(s))

	s = time.Now()
	_, err = tx.Exec("update validators set balanceactivation = (select balance from validator_balances_p where validator_balances_p.week = validators.activationepoch / 1575 and validator_balances_p.epoch = validators.activationepoch and validator_balances_p.validatorindex = validators.validatorindex) WHERE balanceactivation IS NULL;")
	if err != nil {
		return err
	}
	logger.Infof("updating validator activation epoch balance completed, took %v", time.Since(s))

	return nil
}

func saveValidatorProposalAssignments(epoch uint64, assignments map[uint64]uint64, tx *sql.Tx) error {
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
	//args := make([][]interface{}, 0, len(assignments))
	argsWeek := make([][]interface{}, 0, len(assignments))
	for key, validator := range assignments {
		keySplit := strings.Split(key, "-")
		//args = append(args, []interface{}{epoch, validator, keySplit[0], keySplit[1], 0})
		argsWeek = append(argsWeek, []interface{}{epoch, validator, keySplit[0], keySplit[1], 0, epoch / 1575})
	}

	batchSize := 10000

	//for b := 0; b < len(args); b += batchSize {
	//	start := b
	//	end := b + batchSize
	//	if len(args) < end {
	//		end = len(args)
	//	}
	//
	//	valueStrings := make([]string, 0, batchSize)
	//	valueArgs := make([]interface{}, 0, batchSize*5)
	//	for i, v := range args[start:end] {
	//		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", i*5+1, i*5+2, i*5+3, i*5+4, i*5+5))
	//		valueArgs = append(valueArgs, v...)
	//	}
	//	stmt := fmt.Sprintf(`
	//	INSERT INTO attestation_assignments (epoch, validatorindex, attesterslot, committeeindex, status)
	//	VALUES %s
	//	ON CONFLICT (epoch, validatorindex, attesterslot, committeeindex) DO NOTHING`, strings.Join(valueStrings, ","))
	//	_, err := tx.Exec(stmt, valueArgs...)
	//	if err != nil {
	//		return fmt.Errorf("error executing save validator attestation assignment statement: %v", err)
	//	}
	//}

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
	batchSize := 10000

	//for b := 0; b < len(validators); b += batchSize {
	//	start := b
	//	end := b + batchSize
	//	if len(validators) < end {
	//		end = len(validators)
	//	}
	//
	//	valueStrings := make([]string, 0, batchSize)
	//	valueArgs := make([]interface{}, 0, batchSize*4)
	//	for i, v := range validators[start:end] {
	//		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4))
	//		valueArgs = append(valueArgs, epoch)
	//		valueArgs = append(valueArgs, v.Index)
	//		valueArgs = append(valueArgs, v.Balance)
	//		valueArgs = append(valueArgs, v.EffectiveBalance)
	//	}
	//	stmt := fmt.Sprintf(`
	//	INSERT INTO validator_balances (epoch, validatorindex, balance, effectivebalance)
	//	VALUES %s
	//	ON CONFLICT (epoch, validatorindex) DO UPDATE SET
	//		balance          = EXCLUDED.balance,
	//		effectivebalance = EXCLUDED.effectivebalance`, strings.Join(valueStrings, ","))
	//	_, err := tx.Exec(stmt, valueArgs...)
	//	if err != nil {
	//		return err
	//	}
	//}

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

	stmtBlock, err := tx.Prepare(`
		INSERT INTO blocks (epoch, slot, blockroot, parentroot, stateroot, signature, randaoreveal, graffiti, eth1data_depositroot, eth1data_depositcount, eth1data_blockhash, proposerslashingscount, attesterslashingscount, attestationscount, depositscount, voluntaryexitscount, proposer, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		ON CONFLICT (slot, blockroot) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmtBlock.Close()

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
			var dbBlockRootHash []byte
			err := DB.Get(&dbBlockRootHash, "SELECT blockroot FROM blocks WHERE slot = $1 and blockroot = $2", b.Slot, b.BlockRoot)

			if err == nil && bytes.Compare(dbBlockRootHash, b.BlockRoot) == 0 {
				logger.Printf("skipping export of block %x at slot %v as it is already present in the db", b.BlockRoot, b.Slot)
				continue
			}
			start := time.Now()
			logger.Infof("exporting block %x at slot %v", b.BlockRoot, b.Slot)

			logger.Infof("deleting placeholder block")
			_, err = tx.Exec("DELETE FROM blocks WHERE slot = $1 AND length(blockroot) = 1", b.Slot) // Delete placeholder block
			if err != nil {
				return fmt.Errorf("error deleting placeholder block: %v", err)
			}

			// Set proposer to MAX_SQL_INTEGER if it is the genesis-block (since we are using integers for validator-indices right now)
			if b.Slot == 0 {
				b.Proposer = 2147483647
			}

			n := time.Now()

			logger.Tracef("writing block data: %v", b.Eth1Data.DepositRoot)
			_, err = stmtBlock.Exec(b.Slot/utils.Config.Chain.SlotsPerEpoch,
				b.Slot,
				b.BlockRoot,
				b.ParentRoot,
				b.StateRoot,
				b.Signature,
				b.RandaoReveal,
				b.Graffiti,
				b.Eth1Data.DepositRoot,
				b.Eth1Data.DepositCount,
				b.Eth1Data.BlockHash,
				len(b.ProposerSlashings),
				len(b.AttesterSlashings),
				len(b.Attestations),
				len(b.Deposits),
				len(b.VoluntaryExits),
				b.Proposer,
				strconv.FormatUint(b.Status, 10))
			if err != nil {
				return fmt.Errorf("error executing stmtBlocks for block %v: %v", b.Slot, err)
			}

			logger.Tracef("done, took %v", time.Since(n))
			n = time.Now()
			logger.Tracef("writing proposer slashings data")
			for i, ps := range b.ProposerSlashings {
				_, err := stmtProposerSlashing.Exec(b.Slot, i, b.BlockRoot, ps.ProposerIndex, ps.Header1.Slot, ps.Header1.ParentRoot, ps.Header1.StateRoot, ps.Header1.BodyRoot, ps.Header1.Signature, ps.Header2.Slot, ps.Header2.ParentRoot, ps.Header2.StateRoot, ps.Header2.BodyRoot, ps.Header2.Signature)
				if err != nil {
					return fmt.Errorf("error executing stmtProposerSlashing for block %v: %v", b.Slot, err)
				}
			}
			logger.Tracef("done, took %v", time.Since(n))
			n = time.Now()

			logger.Tracef("writing attester slashings data")
			for i, as := range b.AttesterSlashings {
				_, err := stmtAttesterSlashing.Exec(b.Slot, i, b.BlockRoot, pq.Array(as.Attestation1.AttestingIndices), as.Attestation1.Signature, as.Attestation1.Data.Slot, as.Attestation1.Data.CommitteeIndex, as.Attestation1.Data.BeaconBlockRoot, as.Attestation1.Data.Source.Epoch, as.Attestation1.Data.Source.Root, as.Attestation1.Data.Target.Epoch, as.Attestation1.Data.Target.Root, pq.Array(as.Attestation2.AttestingIndices), as.Attestation2.Signature, as.Attestation2.Data.Slot, as.Attestation2.Data.CommitteeIndex, as.Attestation2.Data.BeaconBlockRoot, as.Attestation2.Data.Source.Epoch, as.Attestation2.Data.Source.Root, as.Attestation2.Data.Target.Epoch, as.Attestation2.Data.Target.Root)
				if err != nil {
					return fmt.Errorf("error executing stmtAttesterSlashing for block %v: %v", b.Slot, err)
				}
			}
			logger.Tracef("done, took %v", time.Since(n))
			n = time.Now()

			logger.Tracef("writing attestation data")

			for i, a := range b.Attestations {
				//attestationAssignmentsArgs := make([][]interface{}, 0, 20000)
				attestationAssignmentsArgsWeek := make([][]interface{}, 0, 20000)
				attestingValidators := make([]string, 0, 20000)

				for _, validator := range a.Attesters {
					//attestationAssignmentsArgs = append(attestationAssignmentsArgs, []interface{}{a.Data.Slot / utils.Config.Chain.SlotsPerEpoch, validator, a.Data.Slot, a.Data.CommitteeIndex, 1, b.Slot})
					attestationAssignmentsArgsWeek = append(attestationAssignmentsArgsWeek, []interface{}{a.Data.Slot / utils.Config.Chain.SlotsPerEpoch, validator, a.Data.Slot, a.Data.CommitteeIndex, 1, b.Slot, a.Data.Slot / utils.Config.Chain.SlotsPerEpoch / 1575})
					attestingValidators = append(attestingValidators, strconv.FormatUint(validator, 10))
				}

				batchSize := 10000

				//for batch := 0; batch < len(attestationAssignmentsArgs); batch += batchSize {
				//	start := batch
				//	end := batch + batchSize
				//	if len(attestationAssignmentsArgs) < end {
				//		end = len(attestationAssignmentsArgs)
				//	}
				//
				//	valueStrings := make([]string, 0, batchSize)
				//	valueArgs := make([]interface{}, 0, batchSize*6)
				//	for i, v := range attestationAssignmentsArgs[start:end] {
				//		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6))
				//		valueArgs = append(valueArgs, v...)
				//	}
				//	stmt := fmt.Sprintf(`
				//		INSERT INTO attestation_assignments (epoch, validatorindex, attesterslot, committeeindex, status, inclusionslot)
				//		VALUES %s
				//		ON CONFLICT (epoch, validatorindex, attesterslot, committeeindex) DO UPDATE SET status = excluded.status, inclusionslot = LEAST((CASE WHEN attestation_assignments.inclusionslot = 0 THEN null ELSE attestation_assignments.inclusionslot END), excluded.inclusionslot)`, strings.Join(valueStrings, ","))
				//	_, err := tx.Exec(stmt, valueArgs...)
				//	if err != nil {
				//		return fmt.Errorf("error executing stmtAttestationAssignments for block %v: %v", b.Slot, err)
				//	}
				//}

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
						return fmt.Errorf("error executing stmtAttestationAssignments_p for block %v: %v", b.Slot, err)
					}
				}

				_, err = stmtValidatorsLastAttestationSlot.Exec(a.Data.Slot, "{"+strings.Join(attestingValidators, ",")+"}")
				if err != nil {
					return fmt.Errorf("error executing stmtValidatorsLastAttestationSlot for block %v: %v", b.Slot, err)
				}

				_, err = stmtAttestations.Exec(b.Slot, i, b.BlockRoot, bitfield.Bitlist(a.AggregationBits).Bytes(), pq.Array(a.Attesters), a.Signature, a.Data.Slot, a.Data.CommitteeIndex, a.Data.BeaconBlockRoot, a.Data.Source.Epoch, a.Data.Source.Root, a.Data.Target.Epoch, a.Data.Target.Root)
				if err != nil {
					return fmt.Errorf("error executing stmtAttestations for block %v: %v", b.Slot, err)
				}
			}

			logger.Tracef("done, took %v", time.Since(n))
			n = time.Now()

			logger.Tracef("writing deposits data")
			for i, d := range b.Deposits {
				_, err := stmtDeposits.Exec(b.Slot, i, b.BlockRoot, nil, d.PublicKey, d.WithdrawalCredentials, d.Amount, d.Signature)
				if err != nil {
					return fmt.Errorf("error executing stmtDeposits for block %v: %v", b.Slot, err)
				}
			}
			logger.Tracef("done, took %v", time.Since(n))
			n = time.Now()

			logger.Tracef("writing voluntary exits data")
			for i, ve := range b.VoluntaryExits {
				_, err := stmtVoluntaryExits.Exec(b.Slot, i, b.BlockRoot, ve.Epoch, ve.ValidatorIndex, ve.Signature)
				if err != nil {
					return fmt.Errorf("error executing stmtVoluntaryExits for block %v: %v", b.Slot, err)
				}
			}
			logger.Tracef("done, took %v", time.Since(n))
			n = time.Now()

			logger.Tracef("writing proposal assignments data")
			_, err = stmtProposalAssignments.Exec(b.Slot/utils.Config.Chain.SlotsPerEpoch, b.Proposer, b.Slot, b.Status)
			if err != nil {
				return fmt.Errorf("error executing stmtProposalAssignments for block %v: %v", b.Slot, err)
			}

			logger.Infof("export of block %x at slot %v completed, took %v", b.BlockRoot, b.Slot, time.Since(start))
		}
	}

	return nil
}

// UpdateEpochStatus will update the epoch status in the database
func UpdateEpochStatus(stats *types.ValidatorParticipation) error {
	_, err := DB.Exec(`
		UPDATE epochs SET
			finalized = $1,
			eligibleether = $2,
			globalparticipationrate = $3,
			votedether = $4
		WHERE epoch = $5`,
		stats.Finalized, stats.EligibleEther, stats.GlobalParticipationRate, stats.VotedEther, stats.Epoch)

	return err
}

// GetTotalValidatorsCount will return the total-validator-count
func GetTotalValidatorsCount() (uint64, error) {
	var totalCount uint64
	err := DB.Get(&totalCount, "SELECT COUNT(*) FROM validators")
	return totalCount, err
}

// GetActiveValidatorCount will return the total-validator-count
func GetActiveValidatorCount() (uint64, error) {
	var count uint64
	err := DB.Get(&count, "select count(*) from validators where status in ('active_offline', 'active_online');")
	return count, err
}

func GetValidatorNames() (map[uint64]string, error) {
	rows, err := DB.Query(`
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
	err := DB.Get(&count, "SELECT entering_validators_count FROM queue ORDER BY ts DESC LIMIT 1")
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("error retrieving validator queue count: %v", err)
	}
	return count, nil
}

// GetValidatorChurnLimit returns the rate at which validators can enter or leave the system
func GetValidatorChurnLimit(currentEpoch uint64) (uint64, error) {
	min := utils.Config.Chain.MinPerEpochChurnLimit

	count, err := GetActiveValidatorCount()
	if err != nil {
		return 0, err
	}
	adaptable := uint64(0)
	if count > 0 {
		adaptable = utils.Config.Chain.ChurnLimitQuotient / count
	}

	if min > adaptable {
		return min, nil
	}

	return adaptable, nil
}

func GetTotalEligibleEther() (uint64, error) {
	var total uint64

	err := DB.Get(&total, `
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
	err := DB.Get(&threshold, `
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
		 `, utils.Config.Chain.MinGenesisActiveValidatorCount*32e9)
	if err != nil {
		return nil, err
	}
	return threshold, nil
}

func IsUserSubscribed(uid uint64, client string) bool {
	var dbResult []struct {
		UserID      uint64 `db:"user_id"`
		EventFilter string `db:"event_filter"`
	}

	err := DB.Select(&dbResult, `
		SELECT user_id, event_filter
		FROM users_subscriptions
		WHERE user_id=$1 AND event_filter=$2
		`,
		uid, strings.ToLower(client)) // was last notification sent 2 days ago for this client

	if err != nil {
		return false
	}

	if len(dbResult) > 0 {
		return true
	}

	return false
}
