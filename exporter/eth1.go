package exporter

import (
	"context"
	"eth2-exporter/db"
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
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/go-ssz"
	contracts "github.com/prysmaticlabs/prysm/contracts/deposit-contract"
	pb "github.com/prysmaticlabs/prysm/proto/beacon/p2p/v1"
	"github.com/prysmaticlabs/prysm/shared/bls"
	"github.com/prysmaticlabs/prysm/shared/bytesutil"
	"github.com/prysmaticlabs/prysm/shared/hashutil"
	"github.com/prysmaticlabs/prysm/shared/params"
	"github.com/sirupsen/logrus"
)

var eth1LookBack = uint64(100)
var eth1MaxFetch = uint64(1000)
var eth1DepositEventSignature = hashutil.HashKeccak256([]byte("DepositEvent(bytes,bytes,bytes,bytes,bytes)"))
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
	eth1DepositContractAddress = common.HexToAddress(utils.Config.Indexer.Eth1DepositContractAddress)
	eth1DepositContractFirstBlock = utils.Config.Indexer.Eth1DepositContractFirstBlock

	rpcClient, err := gethRPC.Dial(utils.Config.Indexer.Eth1Endpoint)
	if err != nil {
		logger.Fatal(err)
	}
	eth1RPCClient = rpcClient
	client := ethclient.NewClient(rpcClient)
	eth1Client = client

	lastFetchedBlock := uint64(0)

	for {
		t0 := time.Now()

		var lastDepositBlock uint64
		err = db.DB.Get(&lastDepositBlock, "select coalesce(max(block_number),0) from eth1_deposits")
		if err != nil {
			logger.WithError(err).Errorf("error retrieving highest block_number of eth1-deposits from db")
			time.Sleep(time.Second * 5)
			continue
		}
		header, err := eth1Client.HeaderByNumber(context.Background(), nil)
		if err != nil {
			logger.WithError(err).Errorf("error getting header from eth1-client")
			time.Sleep(time.Second * 5)
			continue
		}
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
		if toBlock-fromBlock > eth1MaxFetch {
			toBlock = fromBlock + 1000
		}
		if toBlock > blockHeight {
			toBlock = blockHeight
		}
		// if we are synced to the head look at the last 100 blocks
		if toBlock-fromBlock < eth1LookBack {
			fromBlock = toBlock - 100
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

		// make sure we are progressing even if there are no deposits in the last batch
		lastFetchedBlock = toBlock

		logger.WithFields(logrus.Fields{
			"duration":      time.Since(t0),
			"blockHeight":   blockHeight,
			"fromBlock":     fromBlock,
			"toBlock":       toBlock,
			"depositsSaved": len(depositsToSave),
		}).Info("exported eth1-deposits")

		// progress faster if we are not synced to head yet
		if blockHeight != toBlock {
			time.Sleep(time.Second * 5)
			continue
		}

		time.Sleep(time.Second * 60)
	}
}

func fetchEth1Deposits(fromBlock, toBlock uint64) (depositsToSave []*types.Eth1Deposit, err error) {
	qry := ethereum.FilterQuery{
		Addresses: []common.Address{
			eth1DepositContractAddress,
		},
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(toBlock),
	}

	depositLogs, err := eth1Client.FilterLogs(context.Background(), qry)
	if err != nil {
		return depositsToSave, fmt.Errorf("error getting logs from eth1-client: %w", err)
	}

	blocksToFetch := []uint64{}
	txsToFetch := []string{}

	for _, depositLog := range depositLogs {
		if depositLog.Topics[0] != eth1DepositEventSignature {
			continue
		}
		pubkey, withdrawalCredentials, amount, signature, merkletreeIndex, err := contracts.UnpackDepositLogData(depositLog.Data)
		if err != nil {
			return depositsToSave, fmt.Errorf("error unpacking eth1-deposit-log: %x: %w", depositLog.Data, err)
		}
		err = VerifyEth1DepositSignature(&ethpb.Deposit_Data{
			PublicKey:             pubkey,
			WithdrawalCredentials: withdrawalCredentials,
			Amount:                bytesutil.FromBytes8(amount),
			Signature:             signature,
		})
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
		return depositsToSave, fmt.Errorf("error getting eth1-blocks: %w", err)
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
		signer := gethTypes.NewEIP155Signer(chainID)
		sender, err := signer.Sender(tx)
		if err != nil {
			return depositsToSave, fmt.Errorf("error getting sender for eth1-deposit")
		}
		d.FromAddress = sender.Bytes()
	}

	return depositsToSave, nil
}

func saveEth1Deposits(depositsToSave []*types.Eth1Deposit) error {
	tx, err := db.DB.Begin()
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
			publickey,
			withdrawal_credentials,
			amount,
			signature,
			merkletree_index,
			removed,
			valid_signature
		)
		VALUES ($1, $2, $3, $4, TO_TIMESTAMP($5), $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (tx_hash, merkletree_index) DO UPDATE SET
			tx_input               = EXCLUDED.tx_input,
			tx_index               = EXCLUDED.tx_index,
			block_number           = EXCLUDED.block_number,
			block_ts               = EXCLUDED.block_ts,
			from_address           = EXCLUDED.from_address,
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
		_, err := insertDepositStmt.Exec(d.TxHash, d.TxInput, d.TxIndex, d.BlockNumber, d.BlockTs, d.FromAddress, d.PublicKey, d.WithdrawalCredentials, d.Amount, d.Signature, d.MerkletreeIndex, d.Removed, d.ValidSignature)
		if err != nil {
			return fmt.Errorf("error saving eth1-deposit to db: %v: %w", fmt.Sprintf("%x", d.TxHash), err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error commiting db-tx for eth1-deposits: %w", err)
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

	if len(elems) == 0 {
		return headers, txs, nil
	}

	ioErr := eth1RPCClient.BatchCall(elems)
	if ioErr != nil {
		return nil, nil, ioErr
	}

	for _, e := range errors {
		if e != nil {
			return nil, nil, e
		}
	}

	return headers, txs, nil
}

// From: "github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
// Avoid including dependency directly as it triggers a
// Cloudflare roughtime call that blocks startup for
// several seconds
// ForkVersionByteLength length of fork version byte array.
const ForkVersionByteLength = 4

// DomainByteLength length of domain byte array.
const DomainByteLength = 4

func ComputeDomain(domainType [DomainByteLength]byte, forkVersion []byte, genesisValidatorsRoot []byte) ([]byte, error) {
	if forkVersion == nil {
		forkVersion = params.BeaconConfig().GenesisForkVersion
	}
	if genesisValidatorsRoot == nil {
		genesisValidatorsRoot = params.BeaconConfig().ZeroHash[:]
	}
	forkBytes := [ForkVersionByteLength]byte{}
	copy(forkBytes[:], forkVersion)

	forkDataRoot, err := computeForkDataRoot(forkBytes[:], genesisValidatorsRoot)
	if err != nil {
		return nil, err
	}

	return domain(domainType, forkDataRoot[:]), nil
}

func domain(domainType [DomainByteLength]byte, forkDataRoot []byte) []byte {
	b := []byte{}
	b = append(b, domainType[:4]...)
	b = append(b, forkDataRoot[:28]...)
	return b
}

func computeForkDataRoot(version []byte, root []byte) ([32]byte, error) {
	r, err := ssz.HashTreeRoot(&pb.ForkData{
		CurrentVersion:        version,
		GenesisValidatorsRoot: root,
	})
	if err != nil {
		return [32]byte{}, err
	}
	return r, nil
}

func VerifyEth1DepositSignature(obj *ethpb.Deposit_Data) error {
	cfg := params.BeaconConfig()
	if utils.Config.Chain.Network == "altona" {
		cfg = params.AltonaConfig()
	} else if utils.Config.Chain.Network == "medalla" {
		cfg = params.MedallaConfig()
	} else if utils.Config.Chain.Network == "spadina" {
		cfg = params.SpadinaConfig()
	}
	domain, err := ComputeDomain(
		cfg.DomainDeposit,
		cfg.GenesisForkVersion,
		cfg.ZeroHash[:],
	)
	if utils.Config.Chain.Network == "zinken" {
		domain, err = ComputeDomain(
			cfg.DomainDeposit,
			[]byte{0x00, 0x00, 0x00, 0x03},
			cfg.ZeroHash[:],
		)
	}
	if utils.Config.Chain.Network == "toledo" {
		domain, err = ComputeDomain(
			cfg.DomainDeposit,
			[]byte{0x00, 0x70, 0x1E, 0xD0},
			cfg.ZeroHash[:],
		)
	}
	if utils.Config.Chain.Network == "pyrmont" {
		domain, err = ComputeDomain(
			cfg.DomainDeposit,
			[]byte{0x00, 0x00, 0x20, 0x09},
			cfg.ZeroHash[:],
		)
	}
	if utils.Config.Chain.Network == "prater" {
		domain, err = ComputeDomain(
			cfg.DomainDeposit,
			[]byte{0x00, 0x00, 0x10, 0x20},
			cfg.ZeroHash[:],
		)
	}
	if err != nil {
		return fmt.Errorf("could not get domain: %w", err)
	}
	blsPubkey, err := bls.PublicKeyFromBytes(obj.PublicKey)
	if err != nil {
		return fmt.Errorf("could not get pubkey: %w", err)
	}
	blsSig, err := bls.SignatureFromBytes(obj.Signature)
	if err != nil {
		return fmt.Errorf("could not get sig %w", err)
	}
	root, err := ssz.SigningRoot(obj)
	if err != nil {
		return fmt.Errorf("could not get root: %w", err)
	}
	signingData := &pb.SigningData{
		ObjectRoot: root[:],
		Domain:     domain,
	}
	ctrRoot, err := ssz.HashTreeRoot(signingData)
	if err != nil {
		return fmt.Errorf("could not get ctr root: %w", err)
	}
	if !blsSig.Verify(blsPubkey, ctrRoot[:]) {
		return fmt.Errorf("invalid signature")
	}
	return nil
}
