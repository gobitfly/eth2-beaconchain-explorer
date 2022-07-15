package rpc

import (
	"context"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	geth_rpc "github.com/ethereum/go-ethereum/rpc"
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

func (client *ErigonClient) GetBlock(number uint64) (*types.Eth1BlockContainer, *types.GetBlockTimings, error) {
	start := time.Now()
	timings := &types.GetBlockTimings{}

	block, err := client.ethClient.BlockByNumber(context.Background(), big.NewInt(int64(number)))

	if err != nil {
		return nil, nil, err
	}

	timings.Headers = time.Since(start)
	start = time.Now()

	// initialize fields for block reward calculation
	reward := new(big.Int).Set(utils.Eth1BlockReward(block.NumberU64()))
	tempReward := new(big.Int)
	uncleInclusionReward := new(big.Int)

	c := &types.Eth1BlockContainer{
		Header: &types.Eth1Header{
			Hash:              block.Hash().Bytes(),
			ParentHash:        block.ParentHash().Bytes(),
			UncleHash:         block.UncleHash().Bytes(),
			Coinbase:          block.Coinbase().Bytes(),
			Root:              block.Root().Bytes(),
			TxHash:            block.TxHash().Bytes(),
			ReceiptHash:       block.ReceiptHash().Bytes(),
			Difficulty:        block.Difficulty().Bytes(),
			Number:            block.NumberU64(),
			GasLimit:          block.GasLimit(),
			GasUsed:           block.GasUsed(),
			Time:              timestamppb.New(time.Unix(int64(block.Time()), 0)),
			Extra:             block.Extra(),
			MixDigest:         block.MixDigest().Bytes(),
			TransactionHashes: make([][]byte, 0, len(block.Transactions())), // TODO: Populate tx hashes
			UncleHashes:       make([][]byte, 0, len(block.Uncles())),
			Bloom:             block.Bloom().Bytes(),
			TxCount:           uint64(len(block.Transactions())),
			UnclesCount:       uint64(len(block.Uncles())),
		},
		Transactions:         map[string]*types.Eth1Transaction{},
		Receipts:             map[string]*types.Eth1TransactionReceipt{},
		Logs:                 map[string]*types.Eth1Log{},
		Uncles:               make([]*types.Eth1Header, 0, len(block.Uncles())),
		Erc20Transfers:       make([]*types.ERC20Transfer, 0, 1000),
		Erc721Transfers:      make([]*types.ERC721Transfer, 0, 1000),
		InternalTransactions: map[string]*types.Eth1InternalTransactionList{},
	}

	for _, uncle := range block.Uncles() {
		pbUncle := &types.Eth1Header{
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

		// Add the uncle inclusion reward to the block reward
		tempReward.Add(big.NewInt(uncle.Number.Int64()), big.NewInt(8))
		tempReward.Sub(tempReward, big.NewInt(block.Number().Int64()))
		tempReward.Mul(tempReward, utils.Eth1BlockReward(block.Number().Uint64()))
		tempReward.Div(tempReward, big.NewInt(8))

		pbUncle.MiningReward = tempReward.Bytes()
		pbUncle.BaseBlockReward = pbUncle.MiningReward
		pbUncle.UncleInclusionReward = big.NewInt(0).Bytes()
		pbUncle.TotalFee = big.NewInt(0).Bytes()

		c.Uncles = append(c.Uncles, pbUncle)

		tempReward.Div(utils.Eth1BlockReward(block.Number().Uint64()), big.NewInt(32))
		uncleInclusionReward.Add(uncleInclusionReward, tempReward)
		reward.Add(reward, tempReward)

		pbUncle.MinerMinGasPrice = big.NewInt(0).Bytes()
	}

	totalFee := new(big.Int)

	receipts := make([]*geth_types.Receipt, len(block.Transactions()))
	reqs := make([]rpc.BatchElem, len(block.Transactions()))

	txs := block.Transactions()

	for i, tx := range txs {
		c.Header.TransactionHashes = append(c.Header.TransactionHashes, tx.Hash().Bytes())

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
			ChainId:              tx.ChainId().Bytes(),
			AccessList:           []*types.Eth1Transaction_AccessList{},
			From:                 msg.From().Bytes(),
			Time:                 timestamppb.New(time.Unix(int64(block.Time()), 0)),
			Index:                uint64(i),
		}

		if tx.To() != nil {
			pbTx.To = tx.To().Bytes()
		}
		c.Transactions[utils.StripPrefix(tx.Hash().String())] = pbTx

	}

	g := new(errgroup.Group)

	g.Go(func() error {
		traces, err := client.TraceParity(block.NumberU64())

		if err != nil {
			return fmt.Errorf("error tracing block (%v), %v: %v", block.Number(), block.Hash(), err)
		}

		timings.Traces = time.Since(start)

		for _, trace := range traces {
			txHash := trace.TransactionHash
			tracePb := &types.Eth1InternalTransaction{
				Hash:  common.FromHex(txHash),
				Block: block.NumberU64(),
				Index: trace.TraceAddress,
				Time:  timestamppb.New(time.Unix(int64(block.Time()), 0)),
				Type:  trace.Type,
			}

			if trace.Type == "create" {
				tracePb.From = common.FromHex(trace.Action.From)
				tracePb.To = common.FromHex(trace.Result.Address)
				tracePb.Value = common.FromHex(trace.Action.Value)
				tracePb.Input = common.FromHex(trace.Action.Init)
				tracePb.Output = common.FromHex(trace.Result.Code)
			} else if trace.Type == "suicide" {
				tracePb.From = common.FromHex(trace.Action.Address)
				tracePb.To = common.FromHex(trace.Action.RefundAddress)
				tracePb.Value = common.FromHex(trace.Action.Balance)
				tracePb.Input = []byte{}
				tracePb.Output = []byte{}
			} else {
				tracePb.From = common.FromHex(trace.Action.From)
				tracePb.To = common.FromHex(trace.Action.To)
				tracePb.Value = common.FromHex(trace.Action.Value)
				tracePb.Input = common.FromHex(trace.Action.Input)
				tracePb.Output = common.FromHex(trace.Result.Output)
			}

			if c.InternalTransactions[txHash] == nil {
				c.InternalTransactions[txHash] = &types.Eth1InternalTransactionList{
					List: make([]*types.Eth1InternalTransaction, 0, 10),
				}
			}

			c.InternalTransactions[txHash].List = append(c.InternalTransactions[txHash].List, tracePb)
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
		c.Receipts[utils.StripPrefix(r.TxHash.String())] = &types.Eth1TransactionReceipt{
			BlockHash:          r.BlockHash.Bytes(),
			BlockNumber:        r.BlockNumber.Uint64(),
			ContractAddress:    r.ContractAddress[:],
			CommulativeGasUsed: r.CumulativeGasUsed,
			GasUsed:            r.GasUsed,
			LogsBloom:          r.Bloom.Bytes(),
			Status:             r.Status,
			TransactionIndex:   uint64(r.TransactionIndex),
		}

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
			c.Logs[fmt.Sprintf("%v#%010d", utils.StripPrefix(r.TxHash.String()), l.Index)] = pbLog

			if len(l.Topics) > 2 && l.Topics[0] == utils.Erc20TransferEventHash && len(l.Topics) == 3 { // ERC20 token transfer event
				pbErc20Transfer := &types.ERC20Transfer{
					TokenAddress: l.Address[:],
					From:         l.Topics[1][:],
					To:           l.Topics[2][:],
					Value:        l.Data,
					TxHash:       r.TxHash[:],
				}
				c.Erc20Transfers = append(c.Erc20Transfers, pbErc20Transfer)
			}

			if len(l.Topics) > 2 && l.Topics[0] == utils.Erc20TransferEventHash && len(l.Topics) == 4 { // ERC721 token transfer event
				pbErc721Transfer := &types.ERC721Transfer{
					TokenAddress: l.Address[:],
					From:         l.Topics[1][:],
					To:           l.Topics[2][:],
					TokenId:      l.Topics[3][:],
					TxHash:       r.TxHash[:],
				}
				c.Erc721Transfers = append(c.Erc721Transfers, pbErc721Transfer)
			}
		}

		fees := big.NewInt(int64(r.GasUsed))
		if block.NumberU64() >= 12965000 {
			minerTip := new(big.Int).Sub(new(big.Int).SetBytes(c.Transactions[utils.StripPrefix(r.TxHash.String())].GasPrice), block.BaseFee())
			fees.Mul(fees, minerTip)
		} else {
			fees.Mul(fees, new(big.Int).SetBytes(c.Transactions[utils.StripPrefix(r.TxHash.String())].GasPrice))
		}
		reward.Add(reward, fees)
		totalFee.Add(totalFee, fees)
	}

	if len(txs) > 0 {
		c.Header.MinerMinGasPrice = txs[len(txs)-1].GasPrice().Bytes()
	} else {
		c.Header.MinerMinGasPrice = big.NewInt(0).Bytes()
	}

	c.Header.MiningReward = reward.Bytes()
	c.Header.BaseBlockReward = new(big.Int).Set(utils.Eth1BlockReward(block.NumberU64())).Bytes()

	if block.BaseFee() != nil {
		c.Header.BaseFee = block.BaseFee().Bytes()
	}

	if err := g.Wait(); err != nil {
		return nil, nil, fmt.Errorf("error retrieving traces for block %v: %v", block.Number(), err)
	}

	return c, timings, nil
}

func (client *ErigonClient) LatestEth1BlockNumber() (uint64, error) {
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
