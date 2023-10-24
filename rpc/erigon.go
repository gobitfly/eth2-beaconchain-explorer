package rpc

import (
	"context"
	"encoding/hex"
	"eth2-exporter/contracts/oneinchoracle"
	"eth2-exporter/erc20"
	"eth2-exporter/types"
	"fmt"
	"math/big"
	"strings"
	"time"

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
}

var CurrentErigonClient *ErigonClient

func NewErigonClient(endpoint string) (*ErigonClient, error) {
	logger.Infof("initializing erigon client at %v", endpoint)
	client := &ErigonClient{
		endpoint: endpoint,
	}

	rpcClient, err := geth_rpc.Dial(client.endpoint)
	if err != nil {
		return nil, fmt.Errorf("error dialing rpc node: %w", err)
	}
	client.rpcClient = rpcClient

	ethClient, err := ethclient.Dial(client.endpoint)
	if err != nil {
		return nil, fmt.Errorf("error dialing rpc node: %w", err)
	}
	client.ethClient = ethClient

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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	start := time.Now()
	timings := &types.GetBlockTimings{}

	block, err := client.ethClient.BlockByNumber(ctx, big.NewInt(int64(number)))
	if err != nil {
		return nil, nil, err
	}

	timings.Headers = time.Since(start)
	start = time.Now()

	c := &types.Eth1Block{
		Hash:         block.Hash().Bytes(),
		ParentHash:   block.ParentHash().Bytes(),
		UncleHash:    block.UncleHash().Bytes(),
		Coinbase:     block.Coinbase().Bytes(),
		Root:         block.Root().Bytes(),
		TxHash:       block.TxHash().Bytes(),
		ReceiptHash:  block.ReceiptHash().Bytes(),
		Difficulty:   block.Difficulty().Bytes(),
		Number:       block.NumberU64(),
		GasLimit:     block.GasLimit(),
		GasUsed:      block.GasUsed(),
		Time:         timestamppb.New(time.Unix(int64(block.Time()), 0)),
		Extra:        block.Extra(),
		MixDigest:    block.MixDigest().Bytes(),
		Bloom:        block.Bloom().Bytes(),
		Uncles:       []*types.Eth1Block{},
		Transactions: []*types.Eth1Transaction{},
		Withdrawals:  []*types.Eth1Withdrawal{},
	}
	blobGasUsed := block.BlobGasUsed()
	if blobGasUsed != nil {
		c.BlobGasUsed = *blobGasUsed
	}
	excessBlobGas := block.ExcessBlobGas()
	if excessBlobGas != nil {
		c.ExcessBlobGas = *excessBlobGas
	}

	if block.BaseFee() != nil {
		c.BaseFee = block.BaseFee().Bytes()
	}

	for _, uncle := range block.Uncles() {
		pbUncle := &types.Eth1Block{
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
		}

		c.Uncles = append(c.Uncles, pbUncle)
	}

	receipts := make([]*geth_types.Receipt, len(block.Transactions()))

	if len(block.Withdrawals()) > 0 {
		withdrawalsIndexed := make([]*types.Eth1Withdrawal, 0, len(block.Withdrawals()))
		for _, w := range block.Withdrawals() {
			withdrawalsIndexed = append(withdrawalsIndexed, &types.Eth1Withdrawal{
				Index:          w.Index,
				ValidatorIndex: w.Validator,
				Address:        w.Address.Bytes(),
				Amount:         new(big.Int).SetUint64(w.Amount).Bytes(),
			})
		}
		c.Withdrawals = withdrawalsIndexed
	}

	txs := block.Transactions()

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

		c.Transactions = append(c.Transactions, pbTx)

	}

	g := new(errgroup.Group)

	g.Go(func() error {
		if block.NumberU64() == 0 { // genesis block is not traceable
			return nil
		}

		var traceError error
		if traceMode == "parity" || traceMode == "parity/geth" {
			traces, err := client.TraceParity(block.NumberU64())

			if err != nil {
				if traceMode == "parity" {
					return fmt.Errorf("error tracing block via parity style traces (%v), %v: %w", block.Number(), block.Hash(), err)
				} else {
					logger.Errorf("error tracing block via parity style traces (%v), %v: %v", block.Number(), block.Hash(), err)

				}
				traceError = err
			} else {
				for _, trace := range traces {
					if trace.Type == "reward" {
						continue
					}

					if trace.TransactionHash == "" {
						continue
					}

					if trace.TransactionPosition >= len(c.Transactions) {
						return fmt.Errorf("error transaction position %v out of range", trace.TransactionPosition)
					}

					if trace.Error == "" {
						c.Transactions[trace.TransactionPosition].Status = 1
					} else {
						c.Transactions[trace.TransactionPosition].Status = 0
						c.Transactions[trace.TransactionPosition].ErrorMsg = trace.Error
					}

					tracePb := &types.Eth1InternalTransaction{
						Type: trace.Type,
						Path: fmt.Sprint(trace.TraceAddress),
					}

					if tracePb.Type == "call" {
						tracePb.Type = trace.Action.CallType
					}

					if trace.Type == "create" {
						tracePb.From = common.FromHex(trace.Action.From)
						tracePb.To = common.FromHex(trace.Result.Address)
						tracePb.Value = common.FromHex(trace.Action.Value)
					} else if trace.Type == "suicide" {
						tracePb.From = common.FromHex(trace.Action.Address)
						tracePb.To = common.FromHex(trace.Action.RefundAddress)
						tracePb.Value = common.FromHex(trace.Action.Balance)
					} else if trace.Type == "call" {
						tracePb.From = common.FromHex(trace.Action.From)
						tracePb.To = common.FromHex(trace.Action.To)
						tracePb.Value = common.FromHex(trace.Action.Value)
					} else {
						spew.Dump(trace)
						logrus.Fatalf("unknown trace type %v in tx %v", trace.Type, trace.TransactionHash)
					}

					c.Transactions[trace.TransactionPosition].Itx = append(c.Transactions[trace.TransactionPosition].Itx, tracePb)
				}
			}
		}

		if traceMode == "geth" || (traceError != nil && traceMode == "parity/geth") {

			gethTraceData, err := client.TraceGeth(block.Hash())

			if err != nil {
				return fmt.Errorf("error tracing block via geth style traces (%v), %v: %w", block.Number(), block.Hash(), err)
			}

			// logger.Infof("retrieved %v calls via geth", len(gethTraceData))

			for _, trace := range gethTraceData {
				if trace.Error == "" {
					c.Transactions[trace.TransactionPosition].Status = 1
				} else {
					c.Transactions[trace.TransactionPosition].Status = 0
					c.Transactions[trace.TransactionPosition].ErrorMsg = trace.Error
				}

				if trace.Type == "CREATE2" {
					trace.Type = "CREATE"
				}

				tracePb := &types.Eth1InternalTransaction{
					Type: strings.ToLower(trace.Type),
					Path: "0",
				}

				tracePb.From = trace.From.Bytes()
				tracePb.To = trace.To.Bytes()
				tracePb.Value = common.FromHex(trace.Value)
				if trace.Type == "CREATE" {
				} else if trace.Type == "SELFDESTRUCT" {
				} else if trace.Type == "SUICIDE" {
				} else if trace.Type == "CALL" || trace.Type == "DELEGATECALL" || trace.Type == "STATICCALL" {
				} else if trace.Type == "" {
					logrus.WithFields(logrus.Fields{"type": trace.Type, "block.Number": block.Number(), "block.Hash": block.Hash()}).Errorf("geth style trace without type")
					spew.Dump(trace)
					continue
				} else {
					spew.Dump(trace)
					logrus.Fatalf("unknown trace type %v in tx %v", trace.Type, trace.TransactionPosition)
				}

				logger.Tracef("appending trace %v to tx %x from %v to %v value %v", trace.TransactionPosition, c.Transactions[trace.TransactionPosition].Hash, trace.From, trace.To, trace.Value)

				c.Transactions[trace.TransactionPosition].Itx = append(c.Transactions[trace.TransactionPosition].Itx, tracePb)
			}
		}

		timings.Traces = time.Since(start)

		// logrus.Infof("retrieved %v traces for %v txs", len(traces), len(c.Transactions))

		return nil
	})

	if err = client.rpcClient.CallContext(ctx, &receipts, "eth_getBlockReceipts", fmt.Sprintf("0x%x", block.NumberU64())); err != nil {
		return nil, nil, fmt.Errorf("error retrieving receipts for block %v: %w", block.Number(), err)
	}

	timings.Receipts = time.Since(start)
	start = time.Now()

	for i, r := range receipts {
		c.Transactions[i].ContractAddress = r.ContractAddress[:]
		c.Transactions[i].CommulativeGasUsed = r.CumulativeGasUsed
		c.Transactions[i].GasUsed = r.GasUsed
		c.Transactions[i].LogsBloom = r.Bloom[:]
		c.Transactions[i].Logs = make([]*types.Eth1Log, 0, len(r.Logs))

		if r.BlobGasPrice != nil {
			c.Transactions[i].BlobGasPrice = r.BlobGasPrice.Bytes()
		}
		c.Transactions[i].BlobGasUsed = r.BlobGasUsed

		for _, l := range r.Logs {
			pbLog := &types.Eth1Log{
				Address: l.Address.Bytes(),
				Data:    l.Data,
				Removed: l.Removed,
				Topics:  make([][]byte, 0, len(l.Topics)),
			}

			for _, t := range l.Topics {
				pbLog.Topics = append(pbLog.Topics, t.Bytes())
			}
			c.Transactions[i].Logs = append(c.Transactions[i].Logs, pbLog)
		}
	}

	if err := g.Wait(); err != nil {
		return nil, nil, fmt.Errorf("error retrieving traces for block %v: %w", block.Number(), err)
	}

	return c, timings, nil
}

func (client *ErigonClient) GetBlockNumberByHash(hash string) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	block, err := client.ethClient.BlockByHash(ctx, common.HexToHash(hash))
	if err != nil {
		return 0, err
	}
	return block.NumberU64(), nil
}

func (client *ErigonClient) GetLatestEth1BlockNumber() (uint64, error) {
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

func (client *ErigonClient) TraceGeth(blockHash common.Hash) ([]*GethTraceCallResult, error) {
	var res []*GethTraceCallResultWrapper

	err := client.rpcClient.Call(&res, "debug_traceBlockByHash", blockHash, gethTracerArg)
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
