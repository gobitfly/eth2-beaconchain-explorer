package rpc

import (
	"context"
	"eth2-exporter/types"
	"fmt"
	"math/big"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	geth_rpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"

	geth_types "github.com/ethereum/go-ethereum/core/types"
)

type ErigonClient struct {
	endpoint  string
	rpcClient *geth_rpc.Client
	ethClient *ethclient.Client
}

func NewErigonClient(endpoint string) (*ErigonClient, error) {
	client := &ErigonClient{
		endpoint: endpoint,
	}

	rpcClient, err := geth_rpc.Dial(client.endpoint)
	if err != nil {
		return nil, fmt.Errorf("error dialing rpc node: %v", err)
	}
	client.rpcClient = rpcClient

	ethClient, err := ethclient.Dial(client.endpoint)
	if err != nil {
		return nil, fmt.Errorf("error dialing rpc node: %v", err)
	}
	client.ethClient = ethClient

	return client, nil
}

func (client *ErigonClient) Close() {
	client.rpcClient.Close()
	client.ethClient.Close()
}

func (client *ErigonClient) GetBlock(number int64) (*types.Eth1Block, *types.GetBlockTimings, error) {
	start := time.Now()
	timings := &types.GetBlockTimings{}

	block, err := client.ethClient.BlockByNumber(context.Background(), big.NewInt(int64(number)))
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
	reqs := make([]rpc.BatchElem, len(block.Transactions()))

	txs := block.Transactions()

	for _, tx := range txs {

		msg, err := tx.AsMessage(geth_types.NewLondonSigner(tx.ChainId()), big.NewInt(1))
		if err != nil {
			return nil, nil, fmt.Errorf("error converting tx %v to msg: %v", tx.Hash(), err)
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
			From:                 msg.From().Bytes(),
			ChainId:              tx.ChainId().Bytes(),
			AccessList:           []*types.AccessList{},
			Hash:                 tx.Hash().Bytes(),
			Itx:                  []*types.Eth1InternalTransaction{},
		}

		if tx.To() != nil {
			pbTx.To = tx.To().Bytes()
		}
		c.Transactions = append(c.Transactions, pbTx)

	}

	g := new(errgroup.Group)

	g.Go(func() error {
		traces, err := client.TraceParity(block.NumberU64())

		if err != nil {
			return fmt.Errorf("error tracing block (%v), %v: %v", block.Number(), block.Hash(), err)
		}

		timings.Traces = time.Since(start)

		// logrus.Infof("retrieved %v traces for %v txs", len(traces), len(c.Transactions))
		for _, trace := range traces {
			if trace.Type == "reward" {
				continue
			}

			if trace.TransactionHash == "" {
				continue
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

			if trace.Type == "create" {
				tracePb.From = common.FromHex(trace.Action.From)
				tracePb.To = common.FromHex(trace.Result.Address)
				tracePb.Value = common.FromHex(trace.Action.Value)
			} else if trace.Type == "suicide" {
				tracePb.From = common.FromHex(trace.Action.Address)
				tracePb.To = common.FromHex(trace.Action.RefundAddress)
				tracePb.Value = common.FromHex(trace.Action.Balance)
			} else if trace.Type == "call" || trace.Type == "delegatecall" {
				tracePb.From = common.FromHex(trace.Action.From)
				tracePb.To = common.FromHex(trace.Action.To)
				tracePb.Value = common.FromHex(trace.Action.Value)
			} else {
				spew.Dump(trace)
				logrus.Fatalf("unknown trace type %v in tx %v", trace.Type, trace.TransactionHash)
			}

			c.Transactions[trace.TransactionPosition].Itx = append(c.Transactions[trace.TransactionPosition].Itx, tracePb)
		}
		return nil
	})

	for i := range reqs {
		reqs[i] = rpc.BatchElem{
			Method: "eth_getTransactionReceipt",
			Args:   []interface{}{txs[i].Hash().String()},
			Result: &receipts[i],
		}
	}

	if len(reqs) > 0 {
		if err := client.rpcClient.BatchCallContext(context.Background(), reqs); err != nil {
			return nil, nil, fmt.Errorf("error retrieving receipts for block %v: %v", block.Number(), err)
		}
	}
	timings.Receipts = time.Since(start)
	start = time.Now()

	for i := range reqs {
		if reqs[i].Error != nil {
			return nil, nil, fmt.Errorf("error retrieving receipt %v for block %v: %v", i, block.Number(), reqs[i].Error)
		}
		if receipts[i] == nil {
			return nil, nil, fmt.Errorf("got null value for receipt %d of block %v", i, block.Number())
		}

		r := receipts[i]
		c.Transactions[i].ContractAddress = r.ContractAddress[:]
		c.Transactions[i].CommulativeGasUsed = r.CumulativeGasUsed
		c.Transactions[i].GasUsed = r.GasUsed
		c.Transactions[i].LogsBloom = r.Bloom[:]
		c.Transactions[i].Logs = make([]*types.Eth1Log, 0, len(r.Logs))

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
		return nil, nil, fmt.Errorf("error retrieving traces for block %v: %v", block.Number(), err)
	}

	return c, timings, nil
}

func (client *ErigonClient) GetLatestEth1BlockNumber() (uint64, error) {
	latestBlock, err := client.ethClient.BlockByNumber(context.Background(), nil)
	if err != nil {
		return 0, fmt.Errorf("error getting latest block: %v", err)
	}

	return latestBlock.NumberU64(), nil
}

type GethTraceCallResult struct {
	Time    string
	GasUsed string
	From    common.Address
	To      common.Address
	Value   string
	Gas     string
	Input   string
	Output  string
	Error   string
	Type    string
	Calls   []*GethTraceCallResult
}

type GethTraceCallData struct {
	From     common.Address
	To       common.Address
	Gas      hexutil.Uint64
	GasPrice hexutil.Big
	Value    hexutil.Big
	Data     hexutil.Bytes
}

var gethTracerArg = map[string]string{
	"tracer": "callTracer",
}

func (client *ErigonClient) TraceGeth(blockHash common.Hash) ([]*GethTraceCallResult, error) {
	var res []*GethTraceCallResult

	err := client.rpcClient.Call(&res, "debug_traceBlockByHash", blockHash, gethTracerArg)
	if err != nil {
		return nil, err
	}

	return res, nil
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
