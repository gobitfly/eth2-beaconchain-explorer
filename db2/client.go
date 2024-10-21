package db2

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/eth2-beaconchain-explorer/db2/store"
)

var ErrNotFoundInCache = fmt.Errorf("cannot find hash in cache")
var ErrMethodNotSupported = fmt.Errorf("methode not supported")

type RawStoreReader interface {
	ReadBlockByNumber(chainID uint64, number int64) (*FullBlockRawData, error)
	ReadBlockByHash(chainID uint64, hash string) (*FullBlockRawData, error)
	ReadBlocksByNumbers(chainID uint64, numbers []int64) (map[int64]*FullBlockRawData, error)
}

type WithFallback struct {
	roundTripper http.RoundTripper
	fallback     http.RoundTripper
}

func NewWithFallback(roundTripper, fallback http.RoundTripper) *WithFallback {
	return &WithFallback{
		roundTripper: roundTripper,
		fallback:     fallback,
	}
}

func (r *WithFallback) RoundTrip(request *http.Request) (*http.Response, error) {
	resp, err := r.roundTripper.RoundTrip(request)
	if err == nil {
		// no fallback needed
		return resp, nil
	}

	var e1 *json.SyntaxError

	if !errors.As(err, &e1) ||
		!errors.Is(err, ErrNotFoundInCache) ||
		!errors.Is(err, ErrMethodNotSupported) ||
		!errors.Is(err, store.ErrNotFound) {
		return nil, err
	}

	return r.fallback.RoundTrip(request)
}

type BigTableEthRaw struct {
	db      RawStoreReader
	chainID uint64
}

func NewBigTableEthRaw(db RawStoreReader, chainID uint64) *BigTableEthRaw {
	return &BigTableEthRaw{
		db:      db,
		chainID: chainID,
	}
}

func (r *BigTableEthRaw) RoundTrip(request *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}
	defer func() {
		request.Body = io.NopCloser(bytes.NewBuffer(body))
	}()

	var messages []*jsonrpcMessage
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&messages); err != nil {
		message := new(jsonrpcMessage)
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(message); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	var resps []*jsonrpcMessage
	for _, message := range messages {
		resp, err := r.handle(request.Context(), message)
		if err != nil {
			return nil, err
		}
		resps = append(resps, resp)
	}

	respBody, err := makeBody(len(resps) == 1, resps)
	if err != nil {
		return nil, err
	}

	return &http.Response{
		Body:       respBody,
		StatusCode: http.StatusOK,
	}, nil
}

func (r *BigTableEthRaw) handle(ctx context.Context, message *jsonrpcMessage) (*jsonrpcMessage, error) {
	var params []interface{}
	err := json.Unmarshal(message.Params, &params)
	if err != nil {
		return nil, err
	}

	var blockNums []string
	if len(params) > 0 {
		if n, ok := params[0].(string); ok {
			blockNums = append(blockNums, n)
		} else {
			return nil, fmt.Errorf("expected string for a block number, got: %v", params[0])
		}
	}

	var numbers []int64
	for _, num := range blockNums {
		if strings.HasPrefix(num, "0x") {
			if n, err := strconv.ParseInt(num[2:], 16, 64); err == nil {
				numbers = append(numbers, n)
			} else {
				return nil, fmt.Errorf("invalid block number: %s", num)
			}
		} else {
			if n, err := strconv.ParseInt(num, 10, 64); err == nil {
				numbers = append(numbers, n)
			} else {
				return nil, fmt.Errorf("invalid block number: %s", num)
			}
		}
	}

	var respBody []byte
	switch message.Method {
	case "eth_getBlockByNumber":
		respBody, err = r.BlocksByNumbers(ctx, numbers)
		if err != nil {
			return nil, err
		}
	case "debug_traceBlockByNumber":
		respBody, err = r.TraceBlocksByNumbers(ctx, numbers)
		if err != nil {
			return nil, err
		}
	case "eth_getBlockReceipts":
		respBody, err = r.BlocksByReceipts(ctx, numbers)
		if err != nil {
			return nil, err
		}
	// case "eth_getUncleByBlockHashAndIndex":
	// 	index, err := hexutil.DecodeBig(args[1].(string))
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	respBody, err = r.UncleByBlockHashAndIndex(ctx, args[0].(string), index.Int64())
	// 	if err != nil {
	// 		return nil, err
	// 	}
	default:
		return nil, ErrMethodNotSupported
	}

	resp := jsonrpcMessage{
		Version: message.Version,
		ID:      message.ID,
	}

	if len(respBody) == 0 {
		resp.Result = []byte("[]")
	} else {
		resp.Result = json.RawMessage(respBody)
	}

	return &resp, nil
}

func makeBody(isSingle bool, messages []*jsonrpcMessage) (io.ReadCloser, error) {
	var b []byte
	var err error
	if isSingle {
		b, err = json.Marshal(messages[0])
	} else {
		b, err = json.Marshal(messages)
	}
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}

func (r *BigTableEthRaw) BlocksByNumbers(ctx context.Context, numbers []int64) ([]byte, error) {
	blocks, err := r.db.ReadBlocksByNumbers(r.chainID, numbers)
	if err != nil {
		return nil, err
	}

	if len(blocks) == 0 {
		return json.Marshal([]interface{}{})
	}

	var results []hexutil.Bytes
	for _, block := range blocks {
		results = append(results, block.Block)
	}

	return json.Marshal(results)
}

func (r *BigTableEthRaw) TraceBlocksByNumbers(ctx context.Context, numbers []int64) ([]byte, error) {
	blocks, err := r.db.ReadBlocksByNumbers(r.chainID, numbers)
	if err != nil {
		return nil, err
	}

	var results []hexutil.Bytes
	for _, block := range blocks {
		results = append(results, block.Traces)
	}

	return json.Marshal(results)
}

func (r *BigTableEthRaw) BlocksByReceipts(ctx context.Context, numbers []int64) ([]byte, error) {
	blocks, err := r.db.ReadBlocksByNumbers(r.chainID, numbers)
	if err != nil {
		return nil, err
	}

	var results []hexutil.Bytes
	for _, block := range blocks {
		results = append(results, block.Receipts)
	}

	return json.Marshal(results)
}

func (r *BigTableEthRaw) UncleByBlockNumberAndIndex(ctx context.Context, number *big.Int, index int64) ([]byte, error) {
	block, err := r.db.ReadBlockByNumber(r.chainID, number.Int64())
	if err != nil {
		return nil, err
	}
	var uncles []*jsonrpcMessage
	err = json.Unmarshal(block.Uncles, &uncles)
	if err != nil {
		return nil, err
	}
	return json.Marshal(uncles[index])
}

func (r *BigTableEthRaw) UncleByBlockHashAndIndex(ctx context.Context, hash string, index int64) ([]byte, error) {
	block, err := r.db.ReadBlockByHash(r.chainID, hash)
	if err != nil {
		return nil, err
	}

	var uncles []*jsonrpcMessage
	err = json.Unmarshal(block.Uncles, &uncles)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal uncles: %v", err)
	}

	if index < 0 || index >= int64(len(uncles)) {
		return nil, fmt.Errorf("index %d out of bounds for uncles array of length %d", index, len(uncles))
	}

	return json.Marshal(uncles[index])
}

// A value of this type can a JSON-RPC request, notification, successful response or
// error response. Which one it is depends on the fields.
type jsonrpcMessage struct {
	Version string          `json:"jsonrpc,omitempty"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Error   *jsonError      `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

type jsonError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
