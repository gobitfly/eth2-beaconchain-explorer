package rpc

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gobitfly/eth2-beaconchain-explorer/contracts/oneinchoracle"
	"github.com/gobitfly/eth2-beaconchain-explorer/db2"
	"github.com/gobitfly/eth2-beaconchain-explorer/db2/store"
	"github.com/gobitfly/eth2-beaconchain-explorer/erc20"
	"github.com/gobitfly/eth2-beaconchain-explorer/metrics"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	geth_rpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"

	geth_types "github.com/ethereum/go-ethereum/core/types"
)

type ErigonClient struct {
	endpoint     string
	rpcClient    *geth_rpc.Client
	ethClient    *ethclient.Client
	chainID      *big.Int
	multiChecker *Balance

	rawStore *db2.CachedRawStore
}

var CurrentErigonClient *ErigonClient

func NewErigonClient(endpoint string) (*ErigonClient, error) {
	logger.Infof("initializing erigon client at %v", endpoint)
	client := &ErigonClient{
		endpoint: endpoint,
	}

	var opts []geth_rpc.ClientOption
	if utils.Config != nil {
		if utils.Config.RawBigtable.Project != "" && utils.Config.RawBigtable.Instance != "" {
			if utils.Config.RawBigtable.Emulator {
				_ = os.Setenv("BIGTABLE_EMULATOR_HOST", fmt.Sprintf("%s:%d", utils.Config.RawBigtable.EmulatorHost, utils.Config.RawBigtable.EmulatorPort))
			}

			project, instance := utils.Config.RawBigtable.Project, utils.Config.RawBigtable.Instance
			bg, err := store.NewBigTable(project, instance, nil)
			if err != nil {
				return nil, err
			}
			rawStore := db2.WithCache(db2.NewRawStore(store.Wrap(bg, db2.BlocRawTable, "")))
			roundTripper := db2.NewBigTableEthRaw(rawStore, utils.Config.Chain.Id)
			opts = append(opts, geth_rpc.WithHTTPClient(&http.Client{
				Transport: db2.NewWithFallback(roundTripper, http.DefaultTransport),
			}))
			client.rawStore = rawStore
		}
	}

	rpcClient, err := geth_rpc.DialOptions(context.Background(), client.endpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("error dialing rpc node: %w", err)
	}
	client.rpcClient = rpcClient
	client.ethClient = ethclient.NewClient(rpcClient)

	client.multiChecker, err = NewBalance(common.HexToAddress("0xb1F8e55c7f64D203C1400B9D8555d050F94aDF39"), client.ethClient)
	if err != nil {
		return nil, fmt.Errorf("error initiation balance checker contract: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	chainID, err := client.ethClient.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting chainid of rpcclient: %w", err)
	}
	client.chainID = chainID

	return client, nil
}

func (client *ErigonClient) Close() {
	client.rpcClient.Close()
	client.ethClient.Close()
}

func (client *ErigonClient) GetChainID() *big.Int {
	return client.chainID
}

func (client *ErigonClient) GetNativeClient() *ethclient.Client {
	return client.ethClient
}

func (client *ErigonClient) GetRPCClient() *geth_rpc.Client {
	return client.rpcClient
}

func (client *ErigonClient) GetBlock(number int64, traceMode string) (*types.Eth1Block, *types.GetBlockTimings, error) {
	start := time.Now()
	timings := &types.GetBlockTimings{}
	mu := sync.Mutex{}

	defer func() {
		metrics.TaskDuration.WithLabelValues("rpc_el_get_block").Observe(time.Since(start).Seconds())
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var traces []*Eth1InternalTransactionWithPosition
	var block *geth_types.Block
	var receipts []*geth_types.Receipt
	g := new(errgroup.Group)
	g.Go(func() error {
		b, err := client.ethClient.BlockByNumber(ctx, big.NewInt(number))
		if err != nil {
			return err
		}
		mu.Lock()
		timings.Headers = time.Since(start)
		mu.Unlock()
		block = b
		return nil
	})
	g.Go(func() error {
		if err := client.rpcClient.CallContext(ctx, &receipts, "eth_getBlockReceipts", fmt.Sprintf("0x%x", number)); err != nil {
			return fmt.Errorf("error retrieving receipts for block %v: %w", number, err)
		}
		mu.Lock()
		timings.Receipts = time.Since(start)
		mu.Unlock()
		return nil
	})
	g.Go(func() error {
		t, err := client.getTrace(traceMode, big.NewInt(number))
		if err != nil {
			return fmt.Errorf("error retrieving traces for block %v: %w", number, err)
		}
		traces = t
		mu.Lock()
		timings.Traces = time.Since(start)
		mu.Unlock()
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, nil, err
	}

	var withdrawals []*types.Eth1Withdrawal
	for _, withdrawal := range block.Withdrawals() {
		withdrawals = append(withdrawals, &types.Eth1Withdrawal{
			Index:          withdrawal.Index,
			ValidatorIndex: withdrawal.Validator,
			Address:        withdrawal.Address.Bytes(),
			Amount:         new(big.Int).SetUint64(withdrawal.Amount).Bytes(),
		})
	}

	var transactions []*types.Eth1Transaction
	traceIndex := 0
	for txPosition, receipt := range receipts {
		logs := make([]*types.Eth1Log, 0, len(receipt.Logs))
		for _, log := range receipt.Logs {
			topics := make([][]byte, 0, len(log.Topics))
			for _, topic := range log.Topics {
				topics = append(topics, topic.Bytes())
			}
			logs = append(logs, &types.Eth1Log{
				Address: log.Address.Bytes(),
				Data:    log.Data,
				Removed: log.Removed,
				Topics:  topics,
			})
		}

		var internals []*types.Eth1InternalTransaction
		for ; traceIndex < len(traces) && traces[traceIndex].txPosition == txPosition; traceIndex++ {
			internals = append(internals, &traces[traceIndex].Eth1InternalTransaction)
		}

		tx := block.Transactions()[txPosition]
		transactions = append(transactions, &types.Eth1Transaction{
			Type:                 uint32(tx.Type()),
			Nonce:                tx.Nonce(),
			GasPrice:             tx.GasPrice().Bytes(),
			MaxPriorityFeePerGas: tx.GasTipCap().Bytes(),
			MaxFeePerGas:         tx.GasFeeCap().Bytes(),
			Gas:                  tx.Gas(),
			Value:                tx.Value().Bytes(),
			Data:                 tx.Data(),
			To: func() []byte {
				if tx.To() != nil {
					return tx.To().Bytes()
				}
				return nil
			}(),
			From: func() []byte {
				sender, err := geth_types.Sender(geth_types.NewCancunSigner(tx.ChainId()), tx)
				if err != nil {
					from, _ := hex.DecodeString("abababababababababababababababababababab")
					logrus.Errorf("error converting tx %v to msg: %v", tx.Hash(), err)
					return from
				}
				return sender.Bytes()
			}(),
			ChainId:            tx.ChainId().Bytes(),
			AccessList:         []*types.AccessList{},
			Hash:               tx.Hash().Bytes(),
			ContractAddress:    receipt.ContractAddress[:],
			CommulativeGasUsed: receipt.CumulativeGasUsed,
			GasUsed:            receipt.GasUsed,
			LogsBloom:          receipt.Bloom[:],
			Status:             receipt.Status,
			Logs:               logs,
			Itx:                internals,
			MaxFeePerBlobGas: func() []byte {
				if tx.BlobGasFeeCap() != nil {
					return tx.BlobGasFeeCap().Bytes()
				}
				return nil
			}(),
			BlobVersionedHashes: func() (b [][]byte) {
				for _, h := range tx.BlobHashes() {
					b = append(b, h.Bytes())
				}
				return b
			}(),
			BlobGasPrice: func() []byte {
				if receipt.BlobGasPrice != nil {
					return receipt.BlobGasPrice.Bytes()
				}
				return nil
			}(),
			BlobGasUsed: receipt.BlobGasUsed,
		})
	}

	var uncles []*types.Eth1Block
	for _, uncle := range block.Uncles() {
		uncles = append(uncles, &types.Eth1Block{
			Hash:        uncle.Hash().Bytes(),
			ParentHash:  uncle.ParentHash.Bytes(),
			UncleHash:   uncle.UncleHash.Bytes(),
			Coinbase:    uncle.Coinbase.Bytes(),
			Root:        uncle.Root.Bytes(),
			TxHash:      uncle.TxHash.Bytes(),
			ReceiptHash: uncle.ReceiptHash.Bytes(),
			Difficulty:  uncle.Difficulty.Bytes(),
			Number:      uncle.Number.Uint64(),
			GasLimit:    uncle.GasLimit,
			GasUsed:     uncle.GasUsed,
			Time:        timestamppb.New(time.Unix(int64(uncle.Time), 0)),
			Extra:       uncle.Extra,
			MixDigest:   uncle.MixDigest.Bytes(),
			Bloom:       uncle.Bloom.Bytes(),
		})
	}

	return &types.Eth1Block{
		Hash:        block.Hash().Bytes(),
		ParentHash:  block.ParentHash().Bytes(),
		UncleHash:   block.UncleHash().Bytes(),
		Coinbase:    block.Coinbase().Bytes(),
		Root:        block.Root().Bytes(),
		TxHash:      block.TxHash().Bytes(),
		ReceiptHash: block.ReceiptHash().Bytes(),
		Difficulty:  block.Difficulty().Bytes(),
		Number:      block.NumberU64(),
		GasLimit:    block.GasLimit(),
		GasUsed:     block.GasUsed(),
		Time:        timestamppb.New(time.Unix(int64(block.Time()), 0)),
		Extra:       block.Extra(),
		MixDigest:   block.MixDigest().Bytes(),
		Bloom:       block.Bloom().Bytes(),
		BaseFee: func() []byte {
			if block.BaseFee() != nil {
				return block.BaseFee().Bytes()
			}
			return nil
		}(),
		Uncles:       uncles,
		Transactions: transactions,
		Withdrawals:  withdrawals,
		BlobGasUsed: func() uint64 {
			blobGasUsed := block.BlobGasUsed()
			if blobGasUsed != nil {
				return *blobGasUsed
			}
			return 0
		}(),
		ExcessBlobGas: func() uint64 {
			excessBlobGas := block.ExcessBlobGas()
			if excessBlobGas != nil {
				return *excessBlobGas
			}
			return 0
		}(),
	}, timings, nil
}

func (client *ErigonClient) GetBlocks(start, end int64, traceMode string) ([]*types.Eth1Block, error) {
	_, err := client.rawStore.ReadBlocksByNumber(client.chainID.Uint64(), start, end)
	if err != nil {
		return nil, err
	}
	var blocks []*types.Eth1Block
	for i := start; i <= end; i++ {
		block, _, err := client.GetBlock(i, traceMode)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func (client *ErigonClient) GetBlockNumberByHash(hash string) (uint64, error) {
	startTime := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("rpc_el_get_block_number_by_hash").Observe(time.Since(startTime).Seconds())
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	block, err := client.ethClient.BlockByHash(ctx, common.HexToHash(hash))
	if err != nil {
		return 0, err
	}
	return block.NumberU64(), nil
}

func (client *ErigonClient) GetLatestEth1BlockNumber() (uint64, error) {
	startTime := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("rpc_el_get_latest_eth1_block_number").Observe(time.Since(startTime).Seconds())
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	latestBlock, err := client.ethClient.BlockByNumber(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("error getting latest block: %w", err)
	}

	return latestBlock.NumberU64(), nil
}

type GethTraceCallResultWrapper struct {
	Result *GethTraceCallResult
}

type GethTraceCallResult struct {
	TransactionPosition int
	Time                string
	GasUsed             string
	From                common.Address
	To                  common.Address
	Value               string
	Gas                 string
	Input               string
	Output              string
	Error               string
	Type                string
	Calls               []*GethTraceCallResult
}

var gethTracerArg = map[string]string{
	"tracer": "callTracer",
}

func extractCalls(r *GethTraceCallResult, d *[]*GethTraceCallResult) {
	if r == nil {
		return
	}
	*d = append(*d, r)

	if r.Calls == nil {
		return
	}
	for _, c := range r.Calls {
		c.TransactionPosition = r.TransactionPosition
		extractCalls(c, d)
	}
}

func (client *ErigonClient) TraceGeth(blockNumber *big.Int) ([]*GethTraceCallResult, error) {
	var res []*GethTraceCallResultWrapper

	err := client.rpcClient.Call(&res, "debug_traceBlockByNumber", hexutil.EncodeBig(blockNumber), gethTracerArg)
	if err != nil {
		return nil, err
	}

	data := make([]*GethTraceCallResult, 0, 20)
	for i, r := range res {
		r.Result.TransactionPosition = i
		extractCalls(r.Result, &data)
	}

	return data, nil
}

type ParityTraceResult struct {
	Action struct {
		CallType      string `json:"callType"`
		From          string `json:"from"`
		Gas           string `json:"gas"`
		Input         string `json:"input"`
		To            string `json:"to"`
		Value         string `json:"value"`
		Init          string `json:"init"`
		Address       string `json:"address"`
		Balance       string `json:"balance"`
		RefundAddress string `json:"refundAddress"`
		Author        string `json:"author"`
		RewardType    string `json:"rewardType"`
	} `json:"action"`
	BlockHash   string `json:"blockHash"`
	BlockNumber int    `json:"blockNumber"`
	Error       string `json:"error"`
	Result      struct {
		GasUsed string `json:"gasUsed"`
		Code    string `json:"code"`
		Output  string `json:"output"`
		Address string `json:"address"`
	} `json:"result"`

	Subtraces           int     `json:"subtraces"`
	TraceAddress        []int64 `json:"traceAddress"`
	TransactionHash     string  `json:"transactionHash"`
	TransactionPosition int     `json:"transactionPosition"`
	Type                string  `json:"type"`
}

func (trace *ParityTraceResult) ConvertFields() ([]byte, []byte, []byte, string) {
	var from, to, value []byte
	tx_type := trace.Type

	switch trace.Type {
	case "create":
		from = common.FromHex(trace.Action.From)
		to = common.FromHex(trace.Result.Address)
		value = common.FromHex(trace.Action.Value)
	case "suicide":
		from = common.FromHex(trace.Action.Address)
		to = common.FromHex(trace.Action.RefundAddress)
		value = common.FromHex(trace.Action.Balance)
	case "call":
		from = common.FromHex(trace.Action.From)
		to = common.FromHex(trace.Action.To)
		value = common.FromHex(trace.Action.Value)
		tx_type = trace.Action.CallType
	default:
		spew.Dump(trace)
		utils.LogFatal(nil, "unknown trace type", 0, map[string]interface{}{"trace type": trace.Type, "tx hash": trace.TransactionHash})
	}
	return from, to, value, tx_type
}

func (client *ErigonClient) TraceParity(blockNumber uint64) ([]*ParityTraceResult, error) {
	var res []*ParityTraceResult

	err := client.rpcClient.Call(&res, "trace_block", blockNumber)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (client *ErigonClient) TraceParityTx(txHash string) ([]*ParityTraceResult, error) {
	var res []*ParityTraceResult

	err := client.rpcClient.Call(&res, "trace_transaction", txHash)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (client *ErigonClient) GetBalances(pairs []*types.Eth1AddressBalance, addressIndex, tokenIndex int) ([]*types.Eth1AddressBalance, error) {
	startTime := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("rpc_el_get_balances").Observe(time.Since(startTime).Seconds())
	}()

	batchElements := make([]geth_rpc.BatchElem, 0, len(pairs))

	ret := make([]*types.Eth1AddressBalance, len(pairs))

	for i, pair := range pairs {

		// if s[1] != "B" {
		// 	logrus.Fatalf("%v has invalid balance update prefix", pair)
		// }

		result := ""

		ret[i] = &types.Eth1AddressBalance{
			Address: pair.Address,
			Token:   pair.Token,
		}

		// logger.Infof("retrieving balance for %x / %x", ret[i].Address, ret[i].Token)

		if len(pair.Token) < 20 {
			batchElements = append(batchElements, geth_rpc.BatchElem{
				Method: "eth_getBalance",
				Args:   []interface{}{common.BytesToAddress(pair.Address), "latest"},
				Result: &result,
			})
		} else {
			to := common.BytesToAddress(pair.Token)
			msg := ethereum.CallMsg{
				To:   &to,
				Gas:  1000000,
				Data: common.Hex2Bytes(fmt.Sprintf("70a08231000000000000000000000000%x", pair.Address)),
			}

			batchElements = append(batchElements, geth_rpc.BatchElem{
				Method: "eth_call",
				Args:   []interface{}{toCallArg(msg), "latest"},
				Result: &result,
			})
		}
	}

	err := client.rpcClient.BatchCall(batchElements)
	if err != nil {
		return nil, fmt.Errorf("error during batch request: %w", err)
	}

	for i, el := range batchElements {
		if el.Error != nil {
			logrus.Warnf("error in batch call: %v", el.Error) // PPR: are smart contracts that pretend to implement the erc20 standard but are somehow buggy
		}

		res := strings.TrimPrefix(*el.Result.(*string), "0x")
		ret[i].Balance = new(big.Int).SetBytes(common.FromHex(res)).Bytes()

		// logger.Infof("retrieved balance %x / %x: %x (%v)", ret[i].Address, ret[i].Token, ret[i].Balance, *el.Result.(*string))
	}

	return ret, nil
}

func (client *ErigonClient) GetBalancesForAddresse(address string, tokenStr []string) ([]*types.Eth1AddressBalance, error) {
	opts := &bind.CallOpts{
		BlockNumber: nil,
	}

	tokens := make([]common.Address, 0, len(tokenStr))

	for _, token := range tokenStr {
		tokens = append(tokens, common.HexToAddress(token))
	}
	balancesInt, err := client.multiChecker.Balances(opts, []common.Address{common.HexToAddress(address)}, tokens)
	if err != nil {
		return nil, err
	}

	res := make([]*types.Eth1AddressBalance, len(tokenStr))
	for tokenIdx := range tokens {

		res[tokenIdx] = &types.Eth1AddressBalance{
			Address: common.FromHex(address),
			Token:   common.FromHex(string(tokens[tokenIdx].Bytes())),
			Balance: balancesInt[tokenIdx].Bytes(),
		}
	}

	return res, nil
}

func (client *ErigonClient) GetNativeBalance(address string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	balance, err := client.ethClient.BalanceAt(ctx, common.HexToAddress(address), nil)

	if err != nil {
		return nil, err
	}
	return balance.Bytes(), nil
}

func (client *ErigonClient) GetERC20TokenBalance(address string, token string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	to := common.HexToAddress(token)
	balance, err := client.ethClient.CallContract(ctx, ethereum.CallMsg{
		To:   &to,
		Gas:  1000000,
		Data: common.Hex2Bytes("70a08231000000000000000000000000" + address),
	}, nil)

	if err != nil && !strings.HasPrefix(err.Error(), "execution reverted") {
		return nil, err
	}
	return balance, nil
}

func (client *ErigonClient) GetERC20TokenMetadata(token []byte) (*types.ERC20Metadata, error) {
	logger.Infof("retrieving metadata for token %x", token)

	oracle, err := oneinchoracle.NewOneInchOracleByChainID(client.GetChainID(), client.ethClient)
	if err != nil {
		return nil, err
	}

	contract, err := erc20.NewErc20(common.BytesToAddress(token), client.ethClient)
	if err != nil {
		return nil, err
	}

	g := new(errgroup.Group)

	ret := &types.ERC20Metadata{}

	g.Go(func() error {
		symbol, err := contract.Symbol(nil)
		if err != nil {
			if strings.Contains(err.Error(), "abi") {
				ret.Symbol = "UNKNOWN"
				return nil
			}

			return fmt.Errorf("error retrieving symbol: %w", err)
		}

		ret.Symbol = symbol
		return nil
	})

	g.Go(func() error {
		totalSupply, err := contract.TotalSupply(nil)
		if err != nil {
			return fmt.Errorf("error retrieving total supply: %w", err)
		}
		ret.TotalSupply = totalSupply.Bytes()
		return nil
	})

	g.Go(func() error {
		decimals, err := contract.Decimals(nil)
		if err != nil {
			return fmt.Errorf("error retrieving decimals: %w", err)
		}
		ret.Decimals = big.NewInt(int64(decimals)).Bytes()
		return nil
	})

	g.Go(func() error {
		rate, err := oracle.GetRateToEth(nil, common.BytesToAddress(token), false)
		if err != nil {
			return fmt.Errorf("error calling oneinchoracle.GetRateToEth: %w", err)
		}
		ret.Price = rate.Bytes()
		return nil
	})

	err = g.Wait()
	if err != nil {
		return ret, err
	}

	if err == nil && len(ret.Decimals) == 0 && ret.Symbol == "" && len(ret.TotalSupply) == 0 {
		// it's possible that a token contract implements the ERC20 interfaces but does not return any values; we use a backup in this case
		ret = &types.ERC20Metadata{
			Decimals:    []byte{0x0},
			Symbol:      "UNKNOWN",
			TotalSupply: []byte{0x0}}
	}

	return ret, err
}

func toCallArg(msg ethereum.CallMsg) interface{} {
	arg := map[string]interface{}{
		"from": msg.From,
		"to":   msg.To,
	}
	if len(msg.Data) > 0 {
		arg["data"] = hexutil.Bytes(msg.Data)
	}
	if msg.Value != nil {
		arg["value"] = (*hexutil.Big)(msg.Value)
	}
	if msg.Gas != 0 {
		arg["gas"] = hexutil.Uint64(msg.Gas)
	}
	if msg.GasPrice != nil {
		arg["gasPrice"] = (*hexutil.Big)(msg.GasPrice)
	}
	return arg
}

func (client *ErigonClient) getTrace(traceMode string, blockNumber *big.Int) ([]*Eth1InternalTransactionWithPosition, error) {
	if blockNumber.Uint64() == 0 { // genesis block is not traceable
		return nil, nil
	}
	switch traceMode {
	case "parity":
		return client.getTraceParity(blockNumber)
	case "parity/geth":
		traces, err := client.getTraceParity(blockNumber)
		if err == nil {
			return traces, nil
		}
		logger.Errorf("error tracing block via parity style traces (%v): %v", blockNumber, err)
		// fallback to geth traces
		fallthrough
	case "geth":
		return client.getTraceGeth(blockNumber)
	}
	return nil, fmt.Errorf("unknown trace mode '%s'", traceMode)
}

func (client *ErigonClient) getTraceParity(blockNumber *big.Int) ([]*Eth1InternalTransactionWithPosition, error) {
	traces, err := client.TraceParity(blockNumber.Uint64())
	if err != nil {
		return nil, fmt.Errorf("error tracing block via parity style traces (%v): %w", blockNumber, err)
	}

	var indexedTraces []*Eth1InternalTransactionWithPosition
	for _, trace := range traces {
		if trace.Type == "reward" {
			continue
		}
		if trace.TransactionHash == "" {
			continue
		}
		// if trace.TransactionPosition >= txsLen {
		// 	return nil, fmt.Errorf("error transaction position %v out of range", trace.TransactionPosition)
		// }

		from, to, value, traceType := trace.ConvertFields()
		indexedTraces = append(indexedTraces, &Eth1InternalTransactionWithPosition{
			Eth1InternalTransaction: types.Eth1InternalTransaction{
				Type:     traceType,
				From:     from,
				To:       to,
				Value:    value,
				ErrorMsg: trace.Error,
				Path:     fmt.Sprint(trace.TraceAddress),
			},
			txPosition: trace.TransactionPosition,
		})
	}
	return indexedTraces, nil
}

func (client *ErigonClient) getTraceGeth(blockNumber *big.Int) ([]*Eth1InternalTransactionWithPosition, error) {
	traces, err := client.TraceGeth(blockNumber)
	if err != nil {
		return nil, fmt.Errorf("error tracing block via geth style traces (%v): %w", blockNumber, err)
	}

	var indexedTraces []*Eth1InternalTransactionWithPosition
	var txPosition int //, tracePath int
	paths := make(map[*GethTraceCallResult]string)
	for i, trace := range traces {
		switch trace.Type {
		case "CREATE2":
			trace.Type = "CREATE"
		case "CREATE", "SELFDESTRUCT", "SUICIDE", "CALL", "DELEGATECALL", "STATICCALL":
		case "":
			logrus.WithFields(logrus.Fields{"type": trace.Type, "block.Number": blockNumber}).Errorf("geth style trace without type")
			spew.Dump(trace)
			continue
		default:
			spew.Dump(trace)
			logrus.Fatalf("unknown trace type %v in tx %v", trace.Type, trace.TransactionPosition)
		}
		if txPosition != trace.TransactionPosition {
			txPosition = trace.TransactionPosition
			paths = make(map[*GethTraceCallResult]string)
		}
		for index, call := range trace.Calls {
			paths[call] = fmt.Sprintf("%s %d", paths[trace], index)
		}

		logger.Tracef("appending trace %v to tx %d:%x from %v to %v value %v", i, blockNumber, trace.TransactionPosition, trace.From, trace.To, trace.Value)
		indexedTraces = append(indexedTraces, &Eth1InternalTransactionWithPosition{
			Eth1InternalTransaction: types.Eth1InternalTransaction{
				Type:     strings.ToLower(trace.Type),
				From:     trace.From.Bytes(),
				To:       trace.To.Bytes(),
				Value:    common.FromHex(trace.Value),
				ErrorMsg: trace.Error,
				Path:     fmt.Sprintf("[%s]", strings.TrimPrefix(paths[trace], " ")),
			},
			txPosition: trace.TransactionPosition,
		})
	}
	return indexedTraces, nil
}

type Eth1InternalTransactionWithPosition struct {
	types.Eth1InternalTransaction
	txPosition int
}

type BlockResponse struct {
	Hash          string                    `json:"hash"`
	ParentHash    string                    `json:"parentHash"`
	UncleHash     string                    `json:"uncleHash"`
	Coinbase      string                    `json:"coinbase"`
	Root          string                    `json:"stateRoot"`
	TxHash        string                    `json:"transactionsHash"`
	ReceiptHash   string                    `json:"receiptsHash"`
	Difficulty    string                    `json:"difficulty"`
	Number        string                    `json:"number"`
	GasLimit      string                    `json:"gasLimit"`
	GasUsed       string                    `json:"gasUsed"`
	Time          string                    `json:"timestamp"`
	Extra         string                    `json:"extraData"`
	MixDigest     string                    `json:"mixHash"`
	Bloom         string                    `json:"logsBloom"`
	Transactions  []*geth_types.Transaction `json:"transactions"`
	Withdrawals   []*geth_types.Withdrawal  `json:"withdrawals"`
	BlobGasUsed   *string                   `json:"blobGasUsed"`
	ExcessBlobGas *string                   `json:"excessBlobGas"`
	BaseFee       string                    `json:"baseFee"`
}

type BlockResponseWithUncles struct {
	BlockResponse
	Uncles []*geth_types.Block
}

type RPCBlock struct {
	Hash        common.Hash   `json:"hash"`
	UncleHashes []common.Hash `json:"uncles"`
}

func (client *ErigonClient) GetBlocksByBatch(blockNumbers []int64) ([]*types.Eth1Block, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var ethBlock []*types.Eth1Block
	var batchCall []geth_rpc.BatchElem
	batchCallNums := 3

	if len(blockNumbers) == 0 {
		return nil, fmt.Errorf("block numbers slice is empty")
	}

	for _, blockNumber := range blockNumbers {
		batchCall = append(batchCall, geth_rpc.BatchElem{
			Method: "eth_getBlockByNumber",
			Args:   []interface{}{blockNumber, true},
			Result: new(json.RawMessage),
		})

		batchCall = append(batchCall, geth_rpc.BatchElem{
			Method: "eth_getBlockReceipts",
			Args:   []interface{}{blockNumber},
			Result: new([]geth_types.Receipt),
		})

		batchCall = append(batchCall, geth_rpc.BatchElem{
			Method: "trace_block",
			Args:   []interface{}{blockNumber},
			Result: new([]ParityTraceResult),
		})
	}

	if len(batchCall) == 0 {
		return ethBlock, nil
	}

	err := client.rpcClient.BatchCallContext(ctx, batchCall)
	if err != nil {
		logger.Errorf("error while batch calling rpc for block details, error: %s", err)
		return nil, err
	}

	for i := 0; i < len(batchCall)/batchCallNums; i++ {
		blockResult := batchCall[i*batchCallNums].Result.(*json.RawMessage)
		receiptsResult := batchCall[i*batchCallNums+1].Result.(*[]geth_types.Receipt)
		tracesResults := batchCall[i*batchCallNums+2].Result.(*[]ParityTraceResult)

		var head *geth_types.Header
		if err := json.Unmarshal(*blockResult, &head); err != nil {
			return nil, fmt.Errorf("error while unmarshaling block results to Header type, error: %v", err)
		}
		var body RPCBlock
		if err := json.Unmarshal(*blockResult, &body); err != nil {
			return nil, fmt.Errorf("error while unmarshaling block results to RPCBlock type, error: %v", err)
		}

		if head.UncleHash == geth_types.EmptyUncleHash && len(body.UncleHashes) > 0 {
			return nil, fmt.Errorf("server returned non-empty uncle list but block header indicates no uncles")
		}
		if head.UncleHash != geth_types.EmptyUncleHash && len(body.UncleHashes) == 0 {
			return nil, fmt.Errorf("server returned empty uncle list but block header indicates uncles")
		}

		var uncles []*geth_types.Block
		if len(body.UncleHashes) > 0 {
			uncles = make([]*geth_types.Block, len(body.UncleHashes))
			uncleHashes := make([]geth_rpc.BatchElem, len(body.UncleHashes))
			for i := range uncleHashes {
				uncleHashes[i] = geth_rpc.BatchElem{
					Method: "eth_getUncleByBlockHashAndIndex",
					Args:   []interface{}{body.Hash, hexutil.EncodeUint64(uint64(i))},
					Result: &uncles[i],
				}
			}
			if err := client.rpcClient.BatchCallContext(ctx, uncleHashes); err != nil {
				return nil, fmt.Errorf("error while batch calling uncle hashes, error: %v", err)
			}

			for i := range uncleHashes {
				if uncleHashes[i].Error != nil {
					return nil, fmt.Errorf("error in uncle hash, error: %v", uncleHashes[i].Error)
				}
				if uncles[i] == nil {
					return nil, fmt.Errorf("got null header for uncle %d of block %x", i, body.Hash[:])
				}
			}
		}

		var blockResponse BlockResponse
		err := json.Unmarshal(*blockResult, &blockResponse)
		if err != nil {
			logger.Errorf("error while unmarshalling block results to BlockResponse type: %s", err)
			continue
		}

		blockResp := BlockResponseWithUncles{
			BlockResponse: blockResponse,
			Uncles:        uncles,
		}

		blockDetails := client.processBlockResult(blockResp)
		client.processReceiptsAndTraces(blockDetails, *receiptsResult, *tracesResults)
		ethBlock = append(ethBlock, blockDetails)
	}

	return ethBlock, nil
}

func (client *ErigonClient) processBlockResult(block BlockResponseWithUncles) *types.Eth1Block {
	blockNumber, err := strconv.ParseUint(block.Number, 0, 64)
	if err != nil {
		logger.Errorf("error while parsing block number to uint64, error: %s", err)
	}
	gasLimit, err := strconv.ParseUint(block.GasLimit, 0, 64)
	if err != nil {
		logger.Errorf("error while parsing gas limit, block: %d, error: %s", blockNumber, err)
	}
	gasUsed, err := strconv.ParseUint(block.GasUsed, 0, 64)
	if err != nil {
		logger.Errorf("error while parsing gas used, block: %d, error: %s", blockNumber, err)
	}
	blockTime, err := strconv.ParseInt(block.Time, 0, 64)
	if err != nil {
		logger.Errorf("error while parsing block time, block: %d, error: %s", blockNumber, err)
	}

	var blobGasUsed, excessBlobGas uint64
	if block.BlobGasUsed != nil {
		blobGasUsedStr := *block.BlobGasUsed
		blobGasUsed, err = strconv.ParseUint(blobGasUsedStr[2:], 16, 64) // remove "0x" and parse as hex
		if err != nil {
			logger.Errorf("error while parsing blob gas used, block: %d, error: %s", blockNumber, err)
		}
	}
	if block.ExcessBlobGas != nil {
		excessBlobGasStr := *block.ExcessBlobGas
		excessBlobGas, err = strconv.ParseUint(excessBlobGasStr[2:], 16, 64)
		if err != nil {
			logger.Errorf("error while parsing excess blob gas, block: %d, error: %s", blockNumber, err)
		}
	}

	ethBlock := &types.Eth1Block{
		Hash:          []byte(block.Hash),
		ParentHash:    []byte(block.ParentHash),
		UncleHash:     []byte(block.UncleHash),
		Coinbase:      []byte(block.Coinbase),
		Root:          []byte(block.Root),
		TxHash:        []byte(block.TxHash),
		ReceiptHash:   []byte(block.ReceiptHash),
		Difficulty:    []byte(block.Difficulty),
		Number:        blockNumber,
		GasLimit:      gasLimit,
		GasUsed:       gasUsed,
		Time:          timestamppb.New(time.Unix(blockTime, 0)),
		Extra:         []byte(block.Extra),
		MixDigest:     []byte(block.MixDigest),
		Bloom:         []byte(block.Bloom),
		Uncles:        []*types.Eth1Block{},
		Transactions:  []*types.Eth1Transaction{},
		Withdrawals:   []*types.Eth1Withdrawal{},
		BlobGasUsed:   blobGasUsed,
		ExcessBlobGas: excessBlobGas,
		BaseFee:       []byte(block.BaseFee),
	}

	if len(block.Withdrawals) > 0 {
		withdrawalsIndexed := make([]*types.Eth1Withdrawal, 0, len(block.Withdrawals))
		for _, w := range block.Withdrawals {
			withdrawalsIndexed = append(withdrawalsIndexed, &types.Eth1Withdrawal{
				Index:          w.Index,
				ValidatorIndex: w.Validator,
				Address:        w.Address.Bytes(),
				Amount:         new(big.Int).SetUint64(w.Amount).Bytes(),
			})
		}
		ethBlock.Withdrawals = withdrawalsIndexed
	}

	txs := block.Transactions

	for _, tx := range txs {

		var from []byte
		sender, err := geth_types.Sender(geth_types.NewCancunSigner(tx.ChainId()), tx)
		if err != nil {
			from, _ = hex.DecodeString("abababababababababababababababababababab")
			logrus.Errorf("error converting tx %v to msg: %v", tx.Hash(), err)
		} else {
			from = sender.Bytes()
		}

		pbTx := &types.Eth1Transaction{
			Type:                 uint32(tx.Type()),
			Nonce:                tx.Nonce(),
			GasPrice:             tx.GasPrice().Bytes(),
			MaxPriorityFeePerGas: tx.GasTipCap().Bytes(),
			MaxFeePerGas:         tx.GasFeeCap().Bytes(),
			Gas:                  tx.Gas(),
			Value:                tx.Value().Bytes(),
			Data:                 tx.Data(),
			From:                 from,
			ChainId:              tx.ChainId().Bytes(),
			AccessList:           []*types.AccessList{},
			Hash:                 tx.Hash().Bytes(),
			Itx:                  []*types.Eth1InternalTransaction{},
			BlobVersionedHashes:  [][]byte{},
		}

		if tx.BlobGasFeeCap() != nil {
			pbTx.MaxFeePerBlobGas = tx.BlobGasFeeCap().Bytes()
		}
		for _, h := range tx.BlobHashes() {
			pbTx.BlobVersionedHashes = append(pbTx.BlobVersionedHashes, h.Bytes())
		}

		if tx.To() != nil {
			pbTx.To = tx.To().Bytes()
		}

		ethBlock.Transactions = append(ethBlock.Transactions, pbTx)

	}

	return ethBlock
}

func (client *ErigonClient) processReceiptsAndTraces(ethBlock *types.Eth1Block, receipts []geth_types.Receipt, traces []ParityTraceResult) {
	traceIndex := 0
	var indexedTraces []*Eth1InternalTransactionWithPosition

	for _, trace := range traces {
		if trace.Type == "reward" {
			continue
		}
		if trace.TransactionHash == "" {
			continue
		}
		if trace.TransactionPosition >= len(ethBlock.Transactions) {
			logrus.Errorf("error transaction position %v out of range", trace.TransactionPosition)
			return
		}

		from, to, value, traceType := trace.ConvertFields()
		indexedTraces = append(indexedTraces, &Eth1InternalTransactionWithPosition{
			Eth1InternalTransaction: types.Eth1InternalTransaction{
				Type:     traceType,
				From:     from,
				To:       to,
				Value:    value,
				ErrorMsg: trace.Error,
				Path:     fmt.Sprint(trace.TraceAddress),
			},
			txPosition: trace.TransactionPosition,
		})
	}

	for txPosition, receipt := range receipts {
		ethBlock.Transactions[txPosition].ContractAddress = receipt.ContractAddress[:]
		ethBlock.Transactions[txPosition].CommulativeGasUsed = receipt.CumulativeGasUsed
		ethBlock.Transactions[txPosition].GasUsed = receipt.GasUsed
		ethBlock.Transactions[txPosition].LogsBloom = receipt.Bloom[:]
		ethBlock.Transactions[txPosition].Logs = make([]*types.Eth1Log, 0, len(receipt.Logs))
		ethBlock.Transactions[txPosition].Status = receipt.Status

		if receipt.BlobGasPrice != nil {
			ethBlock.Transactions[txPosition].BlobGasPrice = receipt.BlobGasPrice.Bytes()
		}
		ethBlock.Transactions[txPosition].BlobGasUsed = receipt.BlobGasUsed

		for _, l := range receipt.Logs {
			topics := make([][]byte, 0, len(l.Topics))
			for _, t := range l.Topics {
				topics = append(topics, t.Bytes())
			}
			ethBlock.Transactions[txPosition].Logs = append(ethBlock.Transactions[txPosition].Logs, &types.Eth1Log{
				Address: l.Address.Bytes(),
				Data:    l.Data,
				Removed: l.Removed,
				Topics:  topics,
			})
		}
		if len(indexedTraces) == 0 {
			continue
		}
		for ; traceIndex < len(indexedTraces) && indexedTraces[traceIndex].txPosition == txPosition; traceIndex++ {
			ethBlock.Transactions[txPosition].Itx = append(ethBlock.Transactions[txPosition].Itx, &indexedTraces[traceIndex].Eth1InternalTransaction)
		}
	}

}
