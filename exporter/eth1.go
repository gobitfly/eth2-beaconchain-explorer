package exporter

import (
	"context"
	"database/sql"
	"eth2-exporter/db"
	"eth2-exporter/metrics"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"math/big"
	"regexp"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	gethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	gethRPC "github.com/ethereum/go-ethereum/rpc"
	"github.com/prysmaticlabs/prysm/v3/contracts/deposit"
	"github.com/prysmaticlabs/prysm/v3/crypto/hash"
	"github.com/prysmaticlabs/prysm/v3/encoding/bytesutil"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"github.com/sirupsen/logrus"
)

var eth1LookBack = uint64(100)
var eth1MaxFetch = uint64(1000)
var eth1DepositEventSignature = hash.HashKeccak256([]byte("DepositEvent(bytes,bytes,bytes,bytes,bytes)"))
var eth1DepositContractFirstBlock uint64
var eth1DepositContractAddress common.Address
var eth1Client *ethclient.Client
var eth1RPCClient *gethRPC.Client
var infuraToMuchResultsErrorRE = regexp.MustCompile("query returned more than [0-9]+ results")
var gethRequestEntityTooLargeRE = regexp.MustCompile("413 Request Entity Too Large")

// eth1DepositsExporter regularly fetches the depositcontract-logs of the
// last 100 blocks and exports the deposits into the database.
// If a reorg of the eth1-chain happened within these 100 blocks it will delete
// removed deposits.
func eth1DepositsExporter() {
	eth1DepositContractAddress = common.HexToAddress(utils.Config.Chain.ClConfig.DepositContractAddress)
	eth1DepositContractFirstBlock = utils.Config.Indexer.Eth1DepositContractFirstBlock

	rpcClient, err := gethRPC.Dial(utils.Config.Eth1GethEndpoint)
	if err != nil {
		utils.LogFatal(err, "new exporter geth client error", 0)
	}
	eth1RPCClient = rpcClient
	client := ethclient.NewClient(rpcClient)
	eth1Client = client

	lastFetchedBlock := uint64(0)

	for {
		t0 := time.Now()

		var lastDepositBlock uint64
		err = db.WriterDb.Get(&lastDepositBlock, "select coalesce(max(block_number),0) from eth1_deposits")
		if err != nil {
			logger.WithError(err).Errorf("error retrieving highest block_number of eth1-deposits from db")
			time.Sleep(time.Second * 5)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		header, err := eth1Client.HeaderByNumber(ctx, nil)
		if err != nil {
			logger.WithError(err).Errorf("error getting header from eth1-client")
			cancel()
			time.Sleep(time.Second * 5)
			continue
		}
		cancel()

		blockHeight := header.Number.Uint64()

		fromBlock := lastDepositBlock + 1
		toBlock := blockHeight

		// start from the first block
		if fromBlock < eth1DepositContractFirstBlock {
			fromBlock = eth1DepositContractFirstBlock
		}
		// make sure we are progressing even if there are no deposits in the last batch
		if fromBlock < lastFetchedBlock+1 {
			fromBlock = lastFetchedBlock + 1
		}
		// if we are not synced to the head yet fetch missing blocks in batches of size 1000
		if toBlock > fromBlock+eth1MaxFetch {
			toBlock = fromBlock + eth1MaxFetch
		}
		if toBlock > blockHeight {
			toBlock = blockHeight
		}
		// if we are synced to the head look at the last 100 blocks
		if toBlock < fromBlock+eth1LookBack {
			if toBlock > eth1LookBack {
				fromBlock = toBlock - eth1LookBack
			} else {
				fromBlock = 0
			}
		}

		depositsToSave, err := fetchEth1Deposits(fromBlock, toBlock)
		if err != nil {
			if infuraToMuchResultsErrorRE.MatchString(err.Error()) || gethRequestEntityTooLargeRE.MatchString(err.Error()) {
				toBlock = fromBlock + 100
				if toBlock > blockHeight {
					toBlock = blockHeight
				}
				logger.Infof("limiting block-range to %v-%v when fetching eth1-deposits due to too much results", fromBlock, toBlock)
				depositsToSave, err = fetchEth1Deposits(fromBlock, toBlock)
			}
			if err != nil {
				logger.WithError(err).WithField("fromBlock", fromBlock).WithField("toBlock", toBlock).Errorf("error fetching eth1-deposits")
				time.Sleep(time.Second * 5)
				continue
			}
		}

		err = saveEth1Deposits(depositsToSave)
		if err != nil {
			logger.WithError(err).Errorf("error saving eth1-deposits")
			time.Sleep(time.Second * 5)
			continue
		}

		if len(depositsToSave) > 0 {
			err = aggregateDeposits()
			if err != nil {
				logger.WithError(err).Errorf("error saving eth1-deposits-leaderboard")
				time.Sleep(time.Second * 5)
				continue
			}
		}

		// make sure we are progressing even if there are no deposits in the last batch
		lastFetchedBlock = toBlock

		if len(depositsToSave) > 0 {
			logger.WithFields(logrus.Fields{
				"duration":      time.Since(t0),
				"blockHeight":   blockHeight,
				"fromBlock":     fromBlock,
				"toBlock":       toBlock,
				"depositsSaved": len(depositsToSave),
			}).Info("exported eth1-deposits")
		}

		// progress faster if we are not synced to head yet
		if blockHeight != toBlock {
			time.Sleep(time.Second * 5)
			continue
		}

		time.Sleep(time.Minute)
	}
}

func fetchEth1Deposits(fromBlock, toBlock uint64) (depositsToSave []*types.Eth1Deposit, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	topic := common.BytesToHash(eth1DepositEventSignature[:])
	qry := ethereum.FilterQuery{
		Addresses: []common.Address{
			eth1DepositContractAddress,
		},
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(toBlock),
		Topics:    [][]common.Hash{{topic}},
	}

	depositLogs, err := eth1Client.FilterLogs(ctx, qry)
	if err != nil {
		return depositsToSave, fmt.Errorf("error getting logs from eth1-client: %w", err)
	}

	blocksToFetch := []uint64{}
	txsToFetch := []string{}

	domain, err := utils.GetSigningDomain()
	if err != nil {
		return nil, err
	}

	for _, depositLog := range depositLogs {
		if depositLog.Topics[0] != eth1DepositEventSignature {
			continue
		}
		pubkey, withdrawalCredentials, amount, signature, merkletreeIndex, err := deposit.UnpackDepositLogData(depositLog.Data)
		if err != nil {
			return depositsToSave, fmt.Errorf("error unpacking eth1-deposit-log: %x: %w", depositLog.Data, err)
		}
		err = deposit.VerifyDepositSignature(&ethpb.Deposit_Data{
			PublicKey:             pubkey,
			WithdrawalCredentials: withdrawalCredentials,
			Amount:                bytesutil.FromBytes8(amount),
			Signature:             signature,
		}, domain)
		validSignature := err == nil
		blocksToFetch = append(blocksToFetch, depositLog.BlockNumber)
		txsToFetch = append(txsToFetch, depositLog.TxHash.Hex())
		depositsToSave = append(depositsToSave, &types.Eth1Deposit{
			TxHash:                depositLog.TxHash.Bytes(),
			TxIndex:               uint64(depositLog.TxIndex),
			BlockNumber:           depositLog.BlockNumber,
			PublicKey:             pubkey,
			WithdrawalCredentials: withdrawalCredentials,
			Amount:                bytesutil.FromBytes8(amount),
			Signature:             signature,
			MerkletreeIndex:       merkletreeIndex,
			Removed:               depositLog.Removed,
			ValidSignature:        validSignature,
		})
	}

	headers, txs, err := eth1BatchRequestHeadersAndTxs(blocksToFetch, txsToFetch)
	if err != nil {
		return depositsToSave, fmt.Errorf("error getting eth1-blocks: %w\nblocks to fetch: %v\n tx to fetch: %v", err, blocksToFetch, txsToFetch)
	}

	for _, d := range depositsToSave {
		// get corresponding block (for the tx-time)
		b, exists := headers[d.BlockNumber]
		if !exists {
			return depositsToSave, fmt.Errorf("error getting block for eth1-deposit: block does not exist in fetched map")
		}
		d.BlockTs = int64(b.Time)

		// get corresponding tx (for input and from-address)
		tx, exists := txs[fmt.Sprintf("0x%x", d.TxHash)]
		if !exists {
			return depositsToSave, fmt.Errorf("error getting tx for eth1-deposit: tx does not exist in fetched map")
		}
		d.TxInput = tx.Data()
		chainID := tx.ChainId()
		if chainID == nil {
			return depositsToSave, fmt.Errorf("error getting tx-chainId for eth1-deposit")
		}
		signer := gethTypes.NewCancunSigner(chainID)
		sender, err := signer.Sender(tx)
		if err != nil {
			return depositsToSave, fmt.Errorf("error getting sender for eth1-deposit (txHash: %x, chainID: %v): %w", d.TxHash, chainID, err)
		}
		d.FromAddress = sender.Bytes()
	}

	return depositsToSave, nil
}

func saveEth1Deposits(depositsToSave []*types.Eth1Deposit) error {
	tx, err := db.WriterDb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	insertDepositStmt, err := tx.Prepare(`
		INSERT INTO eth1_deposits (
			tx_hash,
			tx_input,
			tx_index,
			block_number,
			block_ts,
			from_address,
			from_address_text,
			publickey,
			withdrawal_credentials,
			amount,
			signature,
			merkletree_index,
			removed,
			valid_signature
		)
		VALUES ($1, $2, $3, $4, TO_TIMESTAMP($5), $6, ENCODE($7, 'hex'), $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (tx_hash, merkletree_index) DO UPDATE SET
			tx_input               = EXCLUDED.tx_input,
			tx_index               = EXCLUDED.tx_index,
			block_number           = EXCLUDED.block_number,
			block_ts               = EXCLUDED.block_ts,
			from_address           = EXCLUDED.from_address,
			from_address_text      = EXCLUDED.from_address_text,
			publickey              = EXCLUDED.publickey,
			withdrawal_credentials = EXCLUDED.withdrawal_credentials,
			amount                 = EXCLUDED.amount,
			signature              = EXCLUDED.signature,
			merkletree_index       = EXCLUDED.merkletree_index,
			removed                = EXCLUDED.removed,
			valid_signature        = EXCLUDED.valid_signature`)
	if err != nil {
		return err
	}
	defer insertDepositStmt.Close()

	for _, d := range depositsToSave {
		_, err := insertDepositStmt.Exec(d.TxHash, d.TxInput, d.TxIndex, d.BlockNumber, d.BlockTs, d.FromAddress, d.FromAddress, d.PublicKey, d.WithdrawalCredentials, d.Amount, d.Signature, d.MerkletreeIndex, d.Removed, d.ValidSignature)
		if err != nil {
			return fmt.Errorf("error saving eth1-deposit to db: %v: %w", fmt.Sprintf("%x", d.TxHash), err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing db-tx for eth1-deposits: %w", err)
	}

	return nil
}

// eth1BatchRequestHeadersAndTxs requests the block range specified in the arguments.
// Instead of requesting each block in one call, it batches all requests into a single rpc call.
// This code is shamelessly stolen and adapted from https://github.com/prysmaticlabs/prysm/blob/2eac24c/beacon-chain/powchain/service.go#L473
func eth1BatchRequestHeadersAndTxs(blocksToFetch []uint64, txsToFetch []string) (map[uint64]*gethTypes.Header, map[string]*gethTypes.Transaction, error) {
	elems := make([]gethRPC.BatchElem, 0, len(blocksToFetch)+len(txsToFetch))
	headers := make(map[uint64]*gethTypes.Header, len(blocksToFetch))
	txs := make(map[string]*gethTypes.Transaction, len(txsToFetch))
	errors := make([]error, 0, len(blocksToFetch)+len(txsToFetch))

	for _, b := range blocksToFetch {
		header := &gethTypes.Header{}
		err := error(nil)
		elems = append(elems, gethRPC.BatchElem{
			Method: "eth_getBlockByNumber",
			Args:   []interface{}{hexutil.EncodeBig(big.NewInt(int64(b))), false},
			Result: header,
			Error:  err,
		})
		headers[b] = header
		errors = append(errors, err)
	}

	for _, txHashHex := range txsToFetch {
		tx := &gethTypes.Transaction{}
		err := error(nil)
		elems = append(elems, gethRPC.BatchElem{
			Method: "eth_getTransactionByHash",
			Args:   []interface{}{txHashHex},
			Result: tx,
			Error:  err,
		})
		txs[txHashHex] = tx
		errors = append(errors, err)
	}

	lenElems := len(elems)

	if lenElems == 0 {
		return headers, txs, nil
	}

	for i := 0; (i * 100) < lenElems; i++ {
		start := (i * 100)
		end := start + 100

		if end > lenElems {
			end = lenElems
		}

		ioErr := eth1RPCClient.BatchCall(elems[start:end])
		if ioErr != nil {
			return nil, nil, ioErr
		}
	}

	for _, e := range errors {
		if e != nil {
			return nil, nil, e
		}
	}

	return headers, txs, nil
}

func aggregateDeposits() error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("exporter_aggregate_eth1_deposits").Observe(time.Since(start).Seconds())
	}()
	_, err := db.WriterDb.Exec(`
		INSERT INTO eth1_deposits_aggregated (from_address, amount, validcount, invalidcount, slashedcount, totalcount, activecount, pendingcount, voluntary_exit_count)
		SELECT
			eth1.from_address,
			SUM(eth1.amount) as amount,
			SUM(eth1.validcount) AS validcount,
			SUM(eth1.invalidcount) AS invalidcount,
			COUNT(CASE WHEN v.status = 'slashed' THEN 1 END) AS slashedcount,
			COUNT(v.pubkey) AS totalcount,
			COUNT(CASE WHEN v.status = 'active_online' OR v.status = 'active_offline' THEN 1 END) as activecount,
			COUNT(CASE WHEN v.status = 'deposited' THEN 1 END) AS pendingcount,
			COUNT(CASE WHEN v.status = 'exited' THEN 1 END) AS voluntary_exit_count
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
		LEFT JOIN (SELECT pubkey, status FROM validators) v ON v.pubkey = eth1.publickey
		GROUP BY eth1.from_address
		ON CONFLICT (from_address) DO UPDATE SET
			amount               = excluded.amount,
			validcount           = excluded.validcount,
			invalidcount         = excluded.invalidcount,
			slashedcount         = excluded.slashedcount,
			totalcount           = excluded.totalcount,
			activecount          = excluded.activecount,
			pendingcount         = excluded.pendingcount,
			voluntary_exit_count = excluded.voluntary_exit_count`)
	if err != nil && err != sql.ErrNoRows {
		return nil
	}
	return err
}
