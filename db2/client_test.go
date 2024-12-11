package db2

import (
	"context"
	"math/big"
	"net/http"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/gobitfly/eth2-beaconchain-explorer/db2/store"
	"github.com/gobitfly/eth2-beaconchain-explorer/db2/storetest"
)

const (
	chainID uint64 = 1
)

func TestBigTableClientRealCondition(t *testing.T) {
	project := os.Getenv("BIGTABLE_PROJECT")
	instance := os.Getenv("BIGTABLE_INSTANCE")
	if project == "" || instance == "" {
		t.Skip("skipping test, set BIGTABLE_PROJECT and BIGTABLE_INSTANCE")
	}

	tests := []struct {
		name  string
		block int64
	}{
		{
			name:  "test block",
			block: 6008149,
		},
		{
			name:  "test block 2",
			block: 141,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bg, err := store.NewBigTable(project, instance, nil)
			if err != nil {
				t.Fatal(err)
			}

			rawStore := NewRawStore(store.Wrap(bg, BlocksRawTable, ""))
			rpcClient, err := rpc.DialOptions(context.Background(), "http://foo.bar", rpc.WithHTTPClient(&http.Client{
				Transport: NewBigTableEthRaw(rawStore, chainID),
			}))
			if err != nil {
				t.Fatal(err)
			}
			ethClient := ethclient.NewClient(rpcClient)

			block, err := ethClient.BlockByNumber(context.Background(), big.NewInt(tt.block))
			if err != nil {
				t.Fatalf("BlockByNumber() error = %v", err)
			}
			if got, want := block.Number().Int64(), tt.block; got != want {
				t.Errorf("got %v, want %v", got, want)
			}

			receipts, err := ethClient.BlockReceipts(context.Background(), rpc.BlockNumberOrHashWithNumber(rpc.BlockNumber(tt.block)))
			if err != nil {
				t.Fatalf("BlockReceipts() error = %v", err)
			}
			if len(block.Transactions()) != 0 && len(receipts) == 0 {
				t.Errorf("receipts should not be empty")
			}

			var traces []GethTraceCallResultWrapper
			if err := rpcClient.Call(&traces, "debug_traceBlockByNumber", hexutil.EncodeBig(block.Number()), gethTracerArg); err != nil {
				t.Fatalf("debug_traceBlockByNumber() error = %v", err)
			}
			if len(block.Transactions()) != 0 && len(traces) == 0 {
				t.Errorf("traces should not be empty")
			}
		})
	}
}

func benchmarkBlockRetrieval(b *testing.B, ethClient *ethclient.Client, rpcClient *rpc.Client) {
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		blockTestNumber := int64(20978000 + b.N)
		_, err := ethClient.BlockByNumber(context.Background(), big.NewInt(blockTestNumber))
		if err != nil {
			b.Fatalf("BlockByNumber() error = %v", err)
		}

		if _, err := ethClient.BlockReceipts(context.Background(), rpc.BlockNumberOrHashWithNumber(rpc.BlockNumber(blockTestNumber))); err != nil {
			b.Fatalf("BlockReceipts() error = %v", err)
		}

		var traces []GethTraceCallResultWrapper
		if err := rpcClient.Call(&traces, "debug_traceBlockByNumber", hexutil.EncodeBig(big.NewInt(blockTestNumber)), gethTracerArg); err != nil {
			b.Fatalf("debug_traceBlockByNumber() error = %v", err)
		}
	}
}

func BenchmarkErigonNode(b *testing.B) {
	node := os.Getenv("ETH1_ERIGON_ENDPOINT")
	if node == "" {
		b.Skip("skipping test, please set ETH1_ERIGON_ENDPOINT")
	}

	rpcClient, err := rpc.DialOptions(context.Background(), node)
	if err != nil {
		b.Fatal(err)
	}

	benchmarkBlockRetrieval(b, ethclient.NewClient(rpcClient), rpcClient)
}

func BenchmarkRawBigTable(b *testing.B) {
	project := os.Getenv("BIGTABLE_PROJECT")
	instance := os.Getenv("BIGTABLE_INSTANCE")
	if project == "" || instance == "" {
		b.Skip("skipping test, set BIGTABLE_PROJECT and BIGTABLE_INSTANCE")
	}

	bt, err := store.NewBigTable(project, instance, nil)
	if err != nil {
		b.Fatal(err)
	}

	rawStore := WithCache(NewRawStore(store.Wrap(bt, BlocksRawTable, "")))
	rpcClient, err := rpc.DialOptions(context.Background(), "http://foo.bar", rpc.WithHTTPClient(&http.Client{
		Transport: NewBigTableEthRaw(rawStore, chainID),
	}))
	if err != nil {
		b.Fatal(err)
	}

	benchmarkBlockRetrieval(b, ethclient.NewClient(rpcClient), rpcClient)
}

func BenchmarkAll(b *testing.B) {
	b.Run("BenchmarkErigonNode", func(b *testing.B) {
		BenchmarkErigonNode(b)
	})
	b.Run("BenchmarkRawBigTable", func(b *testing.B) {
		BenchmarkRawBigTable(b)
	})
}

func TestBigTableClient(t *testing.T) {
	tests := []struct {
		name  string
		block FullBlockRawData
	}{
		{
			name:  "test block",
			block: testFullBlock,
		},
		{
			name:  "two uncles",
			block: testTwoUnclesFullBlock,
		},
	}

	client, admin := storetest.NewBigTable(t)
	bg, err := store.NewBigTableWithClient(context.Background(), client, admin, raw)
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rawStore := NewRawStore(store.Wrap(bg, BlocksRawTable, ""))
			if err := rawStore.AddBlocks([]FullBlockRawData{tt.block}); err != nil {
				t.Fatal(err)
			}

			rpcClient, err := rpc.DialOptions(context.Background(), "http://foo.bar", rpc.WithHTTPClient(&http.Client{
				Transport: NewBigTableEthRaw(WithCache(rawStore), tt.block.ChainID),
			}))
			if err != nil {
				t.Fatal(err)
			}
			ethClient := ethclient.NewClient(rpcClient)

			block, err := ethClient.BlockByNumber(context.Background(), big.NewInt(tt.block.BlockNumber))
			if err != nil {
				t.Fatalf("BlockByNumber() error = %v", err)
			}
			if got, want := block.Number().Int64(), tt.block.BlockNumber; got != want {
				t.Errorf("got %v, want %v", got, want)
			}

			receipts, err := ethClient.BlockReceipts(context.Background(), rpc.BlockNumberOrHashWithNumber(rpc.BlockNumber(tt.block.BlockNumber)))
			if err != nil {
				t.Fatalf("BlockReceipts() error = %v", err)
			}
			if len(block.Transactions()) != 0 && len(receipts) == 0 {
				t.Errorf("receipts should not be empty")
			}

			var traces []GethTraceCallResultWrapper
			if err := rpcClient.Call(&traces, "debug_traceBlockByNumber", hexutil.EncodeBig(block.Number()), gethTracerArg); err != nil {
				t.Fatalf("debug_traceBlockByNumber() error = %v", err)
			}
			if len(block.Transactions()) != 0 && len(traces) == 0 {
				t.Errorf("traces should not be empty")
			}
		})
	}
}

func TestBigTableClientWithFallback(t *testing.T) {
	node := os.Getenv("ETH1_ERIGON_ENDPOINT")
	if node == "" {
		t.Skip("skipping test, set ETH1_ERIGON_ENDPOINT")
	}

	tests := []struct {
		name  string
		block FullBlockRawData
	}{
		{
			name:  "test block",
			block: testFullBlock,
		},
	}

	client, admin := storetest.NewBigTable(t)
	bg, err := store.NewBigTableWithClient(context.Background(), client, admin, raw)
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rawStore := NewRawStore(store.Wrap(bg, BlocksRawTable, ""))

			rpcClient, err := rpc.DialOptions(context.Background(), node, rpc.WithHTTPClient(&http.Client{
				Transport: NewWithFallback(NewBigTableEthRaw(rawStore, tt.block.ChainID), http.DefaultTransport),
			}))
			if err != nil {
				t.Fatal(err)
			}
			ethClient := ethclient.NewClient(rpcClient)

			balance, err := ethClient.BalanceAt(context.Background(), common.Address{}, big.NewInt(tt.block.BlockNumber))
			if err != nil {
				t.Fatal(err)
			}
			if balance == nil {
				t.Errorf("empty balance")
			}

			block, err := ethClient.BlockByNumber(context.Background(), big.NewInt(tt.block.BlockNumber))
			if err != nil {
				t.Fatalf("BlockByNumber() error = %v", err)
			}
			if got, want := block.Number().Int64(), tt.block.BlockNumber; got != want {
				t.Errorf("got %v, want %v", got, want)
			}

			receipts, err := ethClient.BlockReceipts(context.Background(), rpc.BlockNumberOrHashWithNumber(rpc.BlockNumber(tt.block.BlockNumber)))
			if err != nil {
				t.Fatalf("BlockReceipts() error = %v", err)
			}
			if len(block.Transactions()) != 0 && len(receipts) == 0 {
				t.Errorf("receipts should not be empty")
			}

			var traces []GethTraceCallResultWrapper
			if err := rpcClient.Call(&traces, "debug_traceBlockByNumber", hexutil.EncodeBig(block.Number()), gethTracerArg); err != nil {
				t.Fatalf("debug_traceBlockByNumber() error = %v", err)
			}
			if len(block.Transactions()) != 0 && len(traces) == 0 {
				t.Errorf("traces should not be empty")
			}
		})
	}
}

// TODO import those 3 from somewhere
var gethTracerArg = map[string]string{
	"tracer": "callTracer",
}

type GethTraceCallResultWrapper struct {
	Result *GethTraceCallResult `json:"result,omitempty"`
}

type GethTraceCallResult struct {
	TransactionPosition int                    `json:"transaction_position,omitempty"`
	Time                string                 `json:"time,omitempty"`
	GasUsed             string                 `json:"gas_used,omitempty"`
	From                common.Address         `json:"from,omitempty"`
	To                  common.Address         `json:"to,omitempty"`
	Value               string                 `json:"value,omitempty"`
	Gas                 string                 `json:"gas,omitempty"`
	Input               string                 `json:"input,omitempty"`
	Output              string                 `json:"output,omitempty"`
	Error               string                 `json:"error,omitempty"`
	Type                string                 `json:"type,omitempty"`
	Calls               []*GethTraceCallResult `json:"calls,omitempty"`
}
