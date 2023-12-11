package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gtuk/discordwebhook"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"google.golang.org/api/option"

	gcp_bigtable "cloud.google.com/go/bigtable"
)

const MAX_EL_BLOCK_NUMBER = 1_000_000_000_000 - 1

const BT_COLUMNFAMILY_BLOCK = "b"
const BT_COLUMN_BLOCK = "b"
const BT_COLUMNFAMILY_RECEIPTS = "r"
const BT_COLUMN_RECEIPTS = "r"
const BT_COLUMNFAMILY_TRACES = "t"
const BT_COLUMN_TRACES = "t"
const BT_COLUMNFAMILY_UNCLES = "u"
const BT_COLUMN_UNCLES = "u"

type blockData struct {
	block  []byte
	txs    []string
	uncles []byte
}

func main() {
	elClientUrl := flag.String("elclienturl", "http://localhost:8545", "")
	discordWebhookReportUrl := flag.String("discord-url", "", "")
	discordWebhookUser := flag.String("discord-user", "", "")
	blockNumber := flag.Int("block-number", -1, "")
	startBlockNumber := flag.Int("start-block-number", 0, "")
	concurrency := flag.Int("concurrency", 1, "")
	btProject := flag.String("btproject", "etherchain", "")
	btInstance := flag.String("btinstance", "beaconchain-node-data-storage", "")
	flag.Parse()

	btClient, err := gcp_bigtable.NewClient(context.Background(), *btProject, *btInstance, option.WithGRPCConnectionPool(1))
	if err != nil {
		utils.LogFatal(err, "creating new client for Bigtable", 0)
	}

	tableBlocksRaw := btClient.Open("blocks-raw")

	client, err := ethclient.Dial(*elClientUrl)

	if err != nil {
		logrus.Fatalf("error dialing eth url: %v", err)
	}

	// chainIdUint64 := uint64(10)
	// latestBlockNumber := uint64(85000000)
	// retrieve the latest block number
	latestBlockNumber, err := client.BlockNumber(context.Background())
	if err != nil {
		logrus.Fatalf("error retrieving latest block number: %v", err)
	}

	chainId, err := client.ChainID(context.Background())
	if err != nil {
		logrus.Fatalf("error retrieving chain id from node: %v", err)
	}
	chainIdUint64 := chainId.Uint64()

	networkName := ""
	if chainIdUint64 == 10 {
		networkName = "OPTIMISM MAINNET"
	} else if chainIdUint64 == 42161 {
		networkName = "ARBITRUM MAINNET"
	}
	// checkRead(tableBlocksRaw, chainIdUint64)

	httpClient := &http.Client{
		Timeout: time.Second * 120,
	}

	if *blockNumber != -1 {
		logrus.Infof("checking block %v", *blockNumber)
		getBlock(*elClientUrl, httpClient, *blockNumber)
		logrus.Info("OK")
		return
	}

	gOuter := &errgroup.Group{}
	gOuter.SetLimit(*concurrency)

	muts := []*gcp_bigtable.Mutation{}
	keys := []string{}
	mux := &sync.Mutex{}

	blocksProcessedTotal := atomic.Int64{}
	blocksProcessedIntv := atomic.Int64{}
	exportStart := time.Now()

	t := time.NewTicker(time.Second * 10)

	go func() {
		for {
			<-t.C

			remainingBlocks := int64(latestBlockNumber) - int64(*startBlockNumber) - blocksProcessedTotal.Load()
			blocksPerSecond := float64(blocksProcessedIntv.Load()) / time.Since(exportStart).Seconds()
			secondsRemaining := float64(remainingBlocks) / float64(blocksPerSecond)

			durationRemaining := time.Second * time.Duration(secondsRemaining)
			logrus.Infof("current speed: %0.1f blocks/sec, %d blocks processed, %d blocks remaining (%0.1f days to go)", blocksPerSecond, blocksProcessedIntv.Load(), remainingBlocks, durationRemaining.Hours()/24)
			blocksProcessedIntv.Store(0)
			exportStart = time.Now()
		}
	}()

	p := message.NewPrinter(language.English)

	for i := *startBlockNumber; i <= int(latestBlockNumber); i++ {

		i := i

		gOuter.Go(func() error {
			for ; ; time.Sleep(time.Second) {

				start := time.Now()

				var bData *blockData
				var receipts, traces []byte
				var blockDuration, receiptsDuration, tracesDuration time.Duration
				var err error
				bData, err = getBlock(*elClientUrl, httpClient, i)

				if err != nil {
					utils.LogError(err, "error processing block", 0, map[string]interface{}{"block": i})
					continue
				}
				blockDuration = time.Since(start)

				if len(bData.txs) > 0 { // only request receipts & traces for blocks with tx
					if chainIdUint64 == 42161 {
						receipts, err = getBatchedReceipts(*elClientUrl, httpClient, bData.txs)
					} else {
						receipts, err = getReceipts(*elClientUrl, httpClient, i)
					}
					receiptsDuration = time.Since(start)
					if err != nil {
						utils.LogError(err, "error processing block", 0, map[string]interface{}{"block": i})
						continue
					}

					if chainIdUint64 == 42161 && i <= 22207815 {
						traces, err = getArbitrumTraces(*elClientUrl, httpClient, i)
					} else {
						traces, err = getGethTraces(*elClientUrl, httpClient, i)
					}
					tracesDuration = time.Since(start)
					if err != nil {
						utils.LogError(err, "error processing block", 0, map[string]interface{}{"block": i})
						continue
					}
				}

				mux.Lock()
				mut := gcp_bigtable.NewMutation()
				mut.Set(BT_COLUMNFAMILY_BLOCK, BT_COLUMN_BLOCK, gcp_bigtable.Timestamp(0), bData.block)
				mut.Set(BT_COLUMNFAMILY_RECEIPTS, BT_COLUMN_RECEIPTS, gcp_bigtable.Timestamp(0), receipts)
				mut.Set(BT_COLUMNFAMILY_TRACES, BT_COLUMN_TRACES, gcp_bigtable.Timestamp(0), traces)
				if len(bData.uncles) > 0 {
					mut.Set(BT_COLUMNFAMILY_UNCLES, BT_COLUMN_UNCLES, gcp_bigtable.Timestamp(0), bData.uncles)
				}

				muts = append(muts, mut)
				key := getBlockKey(i, chainIdUint64)
				keys = append(keys, key)

				if len(keys) == 1000 {
					ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
					errs, err := tableBlocksRaw.ApplyBulk(ctx, keys, muts)

					if err != nil {
						logrus.Errorf("error writing data to bigtable: %v", err)
						cancel()
						continue
					}

					for _, err := range errs {
						logrus.Errorf("error writing data to bigtable: %v", err)
						cancel()
						continue
					}
					cancel()
					logrus.Infof("completed processing block %v (block: %v bytes (%v), receipts: %v bytes (%v), traces: %v bytes (%v), total: %v bytes)", i, len(bData.block), blockDuration, len(receipts), receiptsDuration, len(traces), tracesDuration, len(bData.block)+len(receipts)+len(traces))

					muts = []*gcp_bigtable.Mutation{}
					keys = []string{}

				}
				mux.Unlock()

				if blocksProcessedTotal.Add(1)%100000 == 0 {
					sendMessage(p.Sprintf("%s NODE EXPORT: currently at block %v of %v (%.1f%%)", networkName, i, latestBlockNumber, float64(i)*100/float64(latestBlockNumber)), *discordWebhookReportUrl, *discordWebhookUser)
				}

				blocksProcessedIntv.Add(1)
				break
			}
			return nil
		})
	}

	gOuter.Wait()
}

/*
func HandleChainReorgs(tableBlocksRaw *gcp_bigtable.Table, elClientUrl string, httpClient *http.Client, depth int) error {

	ctx := context.Background()
	// get latest block from the node
	latestNodeBlock, err := client.GetNativeClient().BlockByNumber(ctx, nil)
	if err != nil {
		return err
	}
	latestNodeBlockNumber := latestNodeBlock.NumberU64()

	// for each block check if block node hash and block db hash match
	for i := latestNodeBlockNumber - uint64(depth); i <= latestNodeBlockNumber; i++ {
		nodeBlock, err := client.GetNativeClient().HeaderByNumber(ctx, big.NewInt(int64(i)))
		if err != nil {
			return err
		}

		dbBlock, err := bt.GetBlockFromBlocksTable(i)
		if err != nil {
			if err == db.ErrBlockNotFound { // exit if we hit a block that is not yet in the db
				return nil
			}
			return err
		}

		if !bytes.Equal(nodeBlock.Hash().Bytes(), dbBlock.Hash) {
			logrus.Warnf("found incosistency at height %v, node block hash: %x, db block hash: %x", i, nodeBlock.Hash().Bytes(), dbBlock.Hash)

			// delete all blocks starting from the fork block up to the latest block in the db
			for j := i; j <= latestNodeBlockNumber; j++ {
				dbBlock, err := bt.GetBlockFromBlocksTable(j)
				if err != nil {
					if err == db.ErrBlockNotFound { // exit if we hit a block that is not yet in the db
						return nil
					}
					return err
				}
				logrus.Infof("deleting block at height %v with hash %x", dbBlock.Number, dbBlock.Hash)
				err = bt.DeleteBlock(dbBlock.Number, dbBlock.Hash)
				if err != nil {
					return err
				}
			}
		} else {
			logrus.Infof("height %v, node block hash: %x, db block hash: %x", i, nodeBlock.Hash().Bytes(), dbBlock.Hash)
		}
	}

	return nil
}
*/
/*
func checkRead(tbl *gcp_bigtable.Table, chainId uint64) {
	ctx := context.Background()

	start := fmt.Sprintf("%d:", chainId)

	previousNumber := uint64(0)
	i := 0
	for {
		filter := gcp_bigtable.NewRange(start, "")
		err := tbl.ReadRows(ctx, filter, func(r gcp_bigtable.Row) bool {
			key := r.Key()
			blockNumberString := strings.Replace(key, fmt.Sprintf("%d:", chainId), "", 1)
			blockNumberUint64, err := strconv.ParseUint(blockNumberString, 10, 64)
			if err != nil {
				logrus.Fatal(err)
			}
			blockNumberUint64 = MAX_EL_BLOCK_NUMBER - blockNumberUint64

			if blockNumberUint64%10000 == 0 {
				logrus.Infof("retrieved block %d", blockNumberUint64)
			}
			if blockNumberUint64 != previousNumber-1 && previousNumber != 0 && i > 1000 {
				logrus.Fatalf("%d != %d", blockNumberUint64, previousNumber)
			}

			previousNumber = blockNumberUint64

			i++

			start = fmt.Sprintf("%s\x00", key)
			return true
		}, gcp_bigtable.RowFilter(gcp_bigtable.ChainFilters(gcp_bigtable.CellsPerRowLimitFilter(1), gcp_bigtable.StripValueFilter())), gcp_bigtable.LimitRows(1000))

		if err != nil {
			logrus.Fatal(err)
		}
	}
}
*/

func sendMessage(content, webhookUrl, username string) {

	message := discordwebhook.Message{
		Username: &username,
		Content:  &content,
	}

	err := discordwebhook.SendMessage(webhookUrl, message)
	if err != nil {
		log.Fatal(err)
	}
}

func getBlock(elClientUrl string, httpClient *http.Client, number int) (*blockData, error) {
	// block
	var resString []byte
	blockParsed := &types.Eth1RpcGetBlockResponse{}
	{
		body := []byte(fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x%x", true],"id":1}`, number))

		r, err := http.NewRequest("POST", elClientUrl, bytes.NewBuffer(body))
		if err != nil {
			return nil, fmt.Errorf("error creating post request: %w", err)
		}

		r.Header.Add("Content-Type", "application/json")

		res, err := httpClient.Do(r)
		if err != nil {
			return nil, fmt.Errorf("error executing post request: %w", err)
		}

		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("error unexpected status code: %d", res.StatusCode)
		}

		defer res.Body.Close()

		resString, err = io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading request body: %w", err)
		}

		err = json.Unmarshal(resString, blockParsed)
		if err != nil {
			return nil, fmt.Errorf("error decoding block response: %w", err)
		}

		if strings.Contains(string(resString), `"error":{"code"`) {
			return nil, fmt.Errorf("eth_getBlockByNumber rpc error: %s", resString)
		}
	}

	// transactions
	var transactions []string
	if blockParsed.Result.Transactions != nil {
		transactions = make([]string, len(blockParsed.Result.Transactions))
		for i, tx := range blockParsed.Result.Transactions {
			transactions[i] = tx.Hash
		}
	} else {
		return nil, fmt.Errorf("blockParsed.Result.Transactions is nil")
	}

	// uncles
	if blockParsed.Result.Uncles != nil {
		uncleCount := len(blockParsed.Result.Uncles)
		if uncleCount > 0 {
			if uncleCount > 2 {
				return nil, fmt.Errorf("found more than 2 uncles: %d", len(blockParsed.Result.Uncles))
			}

			var body []byte
			if uncleCount == 1 {
				body = []byte(fmt.Sprintf(`[{"jsonrpc":"2.0","method":"eth_getUncleByBlockNumberAndIndex","params":["0x%x", "0x0"],"id":1}]`, number))
			} else {
				body = []byte(fmt.Sprintf(`[{"jsonrpc":"2.0","method":"eth_getUncleByBlockNumberAndIndex","params":["0x%x", "0x0"],"id":1}, {"jsonrpc":"2.0","method":"eth_getUncleByBlockNumberAndIndex","params":["0x%x", "0x1"],"id":2}]`, number, number))
			}

			r, err := http.NewRequest("POST", elClientUrl, bytes.NewBuffer(body))
			if err != nil {
				return nil, fmt.Errorf("error creating post request for uncle: %w", err)
			}

			r.Header.Add("Content-Type", "application/json")

			res, err := httpClient.Do(r)
			if err != nil {
				return nil, fmt.Errorf("error executing post request for uncle: %w", err)
			}

			if res.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("error unexpected status code: %d", res.StatusCode)
			}

			defer res.Body.Close()

			resStringUncle, err := io.ReadAll(res.Body)
			if err != nil {
				return nil, fmt.Errorf("error reading request body for uncle: %w", err)
			}

			if strings.Contains(string(resStringUncle), `"error":{"code"`) {
				return nil, fmt.Errorf("eth_getUncleByBlockNumberAndIndex rpc error: %s", resStringUncle)
			}

			return &blockData{block: compress(resString), txs: transactions, uncles: compress(resStringUncle)}, nil
		}
	}

	return &blockData{block: compress(resString), txs: transactions, uncles: nil}, nil
}

func getReceipts(url string, httpClient *http.Client, number int) ([]byte, error) {
	body := []byte(fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockReceipts","params":["0x%x"],"id":1}`, number))

	r, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("error creating post request: %v", err)
	}

	r.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("error executing post request: %v", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error unexpected status code: %v", res.StatusCode)
	}

	resString, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading request body: %v", err)
	}

	if strings.Contains(string(resString), `"error":{"code"`) {
		return nil, fmt.Errorf("eth_getBlockReceipts rpc error: %s", resString)
	}

	// fmt.Println(string(resString))

	return compress(resString), nil
}

func getBatchedReceipts(url string, httpClient *http.Client, txs []string) ([]byte, error) {

	body := strings.Builder{}
	body.WriteString("[")
	for i, tx := range txs {
		body.WriteString(fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getTransactionReceipt","params":["%s"],"id":%d}`, tx, i))

		if i != len(txs)-1 {
			body.WriteString(",")
		}
	}
	body.WriteString("]")

	r, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(body.String())))
	if err != nil {
		return nil, fmt.Errorf("error creating post request: %v", err)
	}

	r.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("error executing post request: %v", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error unexpected status code: %v", res.StatusCode)
	}

	resString, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading request body: %v", err)
	}

	if strings.Contains(string(resString), `"error":{"code"`) {
		return nil, fmt.Errorf("eth_getTransactionReceipt rpc error: %s", resString)
	}

	// fmt.Println(string(resString))

	return compress(resString), nil
}

func getGethTraces(url string, httpClient *http.Client, number int) ([]byte, error) {

	if number == 0 { // genesis block can't be traced
		return []byte{}, nil
	}

	body := []byte(fmt.Sprintf(`{"jsonrpc":"2.0","method":"debug_traceBlockByNumber","params":["0x%x", {"tracer": "callTracer"}],"id":1}`, number))

	r, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("error creating post request: %v", err)
	}

	r.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("error executing post request: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error unexpected status code: %v", res.StatusCode)
	}

	defer res.Body.Close()

	resString, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading request body: %v", err)
	}

	if strings.Contains(string(resString), `"error":{"code"`) {
		return nil, fmt.Errorf("debug_traceBlockByNumber rpc error: %s", resString)
	}

	return compress(resString), nil
}

func getArbitrumTraces(url string, httpClient *http.Client, number int) ([]byte, error) {

	if number == 0 { // genesis block can't be traced
		return []byte{}, nil
	}

	body := []byte(fmt.Sprintf(`{"jsonrpc":"2.0","method":"arbtrace_block","params":["0x%x"],"id":1}`, number))

	r, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("error creating post request: %v", err)
	}

	r.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("error executing post request: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error unexpected status code: %v", res.StatusCode)
	}

	defer res.Body.Close()

	resString, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading request body: %v", err)
	}

	if strings.Contains(string(resString), `"error":{"code"`) {
		return nil, fmt.Errorf("arbtrace_block rpc error: %s", resString)
	}

	return compress(resString), nil
}

/*
func printCall(calls []types.Eth1RpcTraceCall, txHash string) {
	for _, call := range calls {
		if call.Type != "STATICCALL" && call.Type != "DELEGATECALL" && call.Type != "CALL" && call.Type != "CREATE" && call.Type != "CREATE2" && call.Type != "SELFDESTRUCT" {
			logrus.Infof("%v in %v", call.Type, txHash)
			spew.Dump(call)
		}
		if len(call.Calls) > 0 {
			printCall(call.Calls, txHash)
		}
	}
}
*/

func compress(src []byte) []byte {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write(src)
	if err != nil {
		logrus.Fatalf("error writing to gzip writer: %v", err)
	}
	if err := zw.Close(); err != nil {
		logrus.Fatalf("error closing gzip writer: %v", err)
	}
	return buf.Bytes()
}

/*
func decompress(src []byte) []byte {
	zr, err := gzip.NewReader(bytes.NewReader(src))
	if err != nil {
		logrus.Fatalf("error creating gzip reader: %v", err)
	}

	data, err := io.ReadAll(zr)
	if err != nil {
		logrus.Fatalf("error reading from gzip reader: %v", err)
	}
	return data
}
*/

func getBlockKey(blockNumber int, chainId uint64) string {
	return fmt.Sprintf("%d:%12d", chainId, MAX_EL_BLOCK_NUMBER-blockNumber)
}
