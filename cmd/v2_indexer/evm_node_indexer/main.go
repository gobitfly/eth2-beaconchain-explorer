package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
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

var ErrBlockNotFound = errors.New("block not found")

type blockData struct {
	block  []byte
	txs    []string
	uncles []byte
}
type eth1RpcGetBlockNumberResponse struct {
	Result string `json:"result"`
}
type eth1RpcGetBlockInfoResponse struct {
	Id   uint64 `json:"id"`
	Hash string `json:"hash"`
}
type dbBlockHash struct {
	Hash string `json:"hash"`
}

var dbBlockCache map[uint64]string

func init() {
	dbBlockCache = make(map[uint64]string)
}

func main() {
	// read / set parameter
	elClientUrl := flag.String("elclienturl", "http://localhost:8545", "url to el client")
	btProject := flag.String("btproject", "etherchain", "bigtable project name")
	btInstance := flag.String("btinstance", "beaconchain-node-data-storage", "bigtable instance name")
	startBlockNumber := flag.Int("start-block-number", 0, "only useful in combination with end-block-number, defined block is included")
	endBlockNumber := flag.Int("end-block-number", 0, "only useful in combination with start-block-number, defined block is included")
	reorgDepth := flag.Int("reorg.depth", 20, "lookback to check and handle chain reorgs")
	concurrency := flag.Int("concurrency", 1, "maximum threads used")
	discordWebhookReportUrl := flag.String("discord-url", "", "report progress to discord url")
	discordWebhookUser := flag.String("discord-user", "", "report progress to discord user")
	flag.Parse()

	_ = startBlockNumber
	_ = endBlockNumber
	_ = reorgDepth

	// init bigtable
	btClient, err := gcp_bigtable.NewClient(context.Background(), *btProject, *btInstance, option.WithGRPCConnectionPool(1))
	if err != nil {
		utils.LogFatal(err, "creating new client for Bigtable", 0)
	}
	tableBlocksRaw := btClient.Open("blocks-raw")
	if tableBlocksRaw == nil {
		utils.LogFatal(err, "open blocks-raw table", 0)
	}

	// init el client
	client, err := ethclient.Dial(*elClientUrl)
	if err != nil {
		logrus.Fatalf("error dialing eth url: %v", err)
	}

	// get chain id
	var chainIdUint64 uint64
	{
		chainId, err := client.ChainID(context.Background())
		if err != nil {
			logrus.Fatalf("error retrieving chain id from node: %v", err)
		}
		chainIdUint64 = chainId.Uint64()
	}

	// chainIdUint64 := uint64(10)
	// latestBlockNumber := uint64(85000000)
	// retrieve the latest block number
	latestBlockNumber, err := client.BlockNumber(context.Background())
	if err != nil {
		logrus.Fatalf("error retrieving latest block number: %v", err)
	}

	// checkRead(tableBlocksRaw, chainIdUint64)

	httpClient := &http.Client{
		Timeout: time.Second * 120,
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
				key := getBlockKey(uint64(i), chainIdUint64)
				keys = append(keys, key)

				if len(keys) == 1000 {

					for ; ; time.Sleep(time.Second) {
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
						break
					}

				}
				mux.Unlock()
				blocksProcessedTotal.Add(1)

				if i%100000 == 0 {
					sendMessage(p.Sprintf("%s NODE EXPORT: currently at block %v of %v (%.1f%%)", getChainName(chainIdUint64), i, latestBlockNumber, float64(i)*100/float64(latestBlockNumber)), *discordWebhookReportUrl, *discordWebhookUser)
				}

				blocksProcessedIntv.Add(1)
				break
			}
			return nil
		})
	}

	gOuter.Wait()
}

func getChainName(chainId uint64) string {
	switch chainId {
	case 1:
		return "<:eth:1184470363967598623> ETHEREUM mainnet"
	case 10:
		return "<:op:1184470125458489354> OPTIMISM mainnet"
	case 100:
		return "<:gnosis:1184470353947398155> GNOSIS mainnet"
	case 42161:
		return "<:arbitrum:1184470344506036334> ARBITRUM mainnet"
	}
	return ""
}

func HandleChainReorgs(tableBlocksRaw *gcp_bigtable.Table, chainId uint64, elClientUrl string, httpClient *http.Client, depth uint64) error {

	// get latest node block number
	var latestNodeBlockNumber uint64
	{
		body := []byte(`{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`)

		r, err := http.NewRequest("POST", elClientUrl, bytes.NewBuffer(body))
		if err != nil {
			return fmt.Errorf("error creating post request for reorg: %w", err)
		}

		r.Header.Add("Content-Type", "application/json")
		res, err := httpClient.Do(r)
		if err != nil {
			return fmt.Errorf("error executing post request for reorg: %w", err)
		}

		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("error unexpected status code for reorg: %d", res.StatusCode)
		}
		defer res.Body.Close()

		resString, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("error reading request body for reorg: %w", err)
		}

		if strings.Contains(string(resString), `"error":{"code"`) {
			return fmt.Errorf("eth_blockNumber rpc error for reorg: %s", resString)
		}

		blockParsed := &eth1RpcGetBlockNumberResponse{}
		err = json.Unmarshal(resString, blockParsed)
		if err != nil {
			return fmt.Errorf("error decoding block response for reorg: %w", err)
		}

		latestNodeBlockNumber, err = strconv.ParseUint(blockParsed.Result, 16, 64)
		if err != nil {
			return fmt.Errorf("error parsing response for reorg: %w", err)
		}
	}

	// define start block
	if depth > latestNodeBlockNumber {
		depth = latestNodeBlockNumber
	}
	startBlock := latestNodeBlockNumber - depth

	// get all block infos from node
	var nodeBlocks []eth1RpcGetBlockInfoResponse
	{
		bodyString := fmt.Sprintf(`[{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x%x", false],"id":%d}`, startBlock, startBlock)
		for i := startBlock + 1; i <= latestNodeBlockNumber; i++ {
			bodyString += fmt.Sprintf(`,{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x%x", false],"id":%d}`, i, i)
		}
		bodyString += "]"

		r, err := http.NewRequest("POST", elClientUrl, bytes.NewBuffer([]byte(bodyString)))
		if err != nil {
			return fmt.Errorf("error creating post request for all blocks: %w", err)
		}

		r.Header.Add("Content-Type", "application/json")
		res, err := httpClient.Do(r)
		if err != nil {
			return fmt.Errorf("error executing post request for all blocks: %w", err)
		}

		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("error unexpected status code for all blocks: %d", res.StatusCode)
		}
		defer res.Body.Close()

		resString, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("error reading request body for all blocks: %w", err)
		}

		if strings.Contains(string(resString), `"error":{"code"`) {
			return fmt.Errorf("eth_blockNumber rpc error for all blocks: %s", resString)
		}

		err = json.Unmarshal(resString, &nodeBlocks)
		if err != nil {
			return fmt.Errorf("error decoding block response for all blocks: %w", err)
		}
	}

	// clean our map before adding new elements (and end up oom cause by an error causing a infinit loop)
	if len(dbBlockCache) > (int)(depth*2) {
		dbBlockCacheNew := make(map[uint64]string)
		for i, v := range dbBlockCache {
			if i >= startBlock {
				dbBlockCache[i] = v
			}
		}
		dbBlockCache = dbBlockCacheNew
	}

	// for each block check if block node hash and block db hash match
	for _, nBlock := range nodeBlocks {

		dbHash, err := getBlockHashFromBT(nBlock.Id, chainId, tableBlocksRaw)
		if err != nil {
			if err == ErrBlockNotFound { // exit if we hit a block that is not yet in the db
				return nil
			}
			return fmt.Errorf("error getting block hash from BT: %w", err)
		}

		if dbHash != nBlock.Hash {
			logrus.Warnf("found incosistency at block %d, node block hash: %s, db block hash: %s", nBlock.Id, nBlock.Hash, dbHash)

			// delete all blocks starting from the fork block up to the latest block in the db
			blocksToDelete := make([]uint64, 0, latestNodeBlockNumber-nBlock.Id+1)
			for i := nBlock.Id; i <= latestNodeBlockNumber; i++ {
				blockInBT, err := isBlockInBT(i, chainId, tableBlocksRaw)
				if err != nil {
					return fmt.Errorf("error getting block hash from BT for delete: %w", err)
				}
				if !blockInBT {
					break // stop collecting blocks if we found a block not in db
				}
				blocksToDelete = append(blocksToDelete, i)
			}

			// remove everything from cache
			dbBlockCache = make(map[uint64]string)

			// remove blocks
			err = deleteBlocksInBT(blocksToDelete, chainId, tableBlocksRaw)
			if err != nil {
				return fmt.Errorf("error deleting block from BT: %w", err)
			}

			// nothing more to do, we deleted everything
			break

		} else {
			logrus.Infof("block %d, node block hash: %s, db block hash: %s", nBlock.Id, nBlock.Hash, dbHash)
		}
	}

	return nil
}

func deleteBlocksInBT(blocks []uint64, chainId uint64, tableBlocksRaw *gcp_bigtable.Table) error {
	if len(blocks) <= 0 {
		return nil
	}

	// INCOMPLETE INCOMPLETE INCOMPLETE
	return nil
}

func isBlockInBT(number uint64, chainId uint64, tableBlocksRaw *gcp_bigtable.Table) (bool, error) {

	if _, ok := dbBlockCache[number]; ok {
		return true, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	row, err := tableBlocksRaw.ReadRow(ctx, getBlockKey(number, chainId))
	if err != nil {
		return false, fmt.Errorf("error reading row (block %d) from bigtable for db blocks: %w", number, err)
	}

	if len(row[BT_COLUMNFAMILY_BLOCK]) == 0 { // block not found
		return false, nil
	}
	return true, nil
}

func getBlockHashFromBT(number uint64, chainId uint64, tableBlocksRaw *gcp_bigtable.Table) (string, error) {

	if val, ok := dbBlockCache[number]; ok {
		return val, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	row, err := tableBlocksRaw.ReadRow(ctx, getBlockKey(number, chainId))
	if err != nil {
		return "", fmt.Errorf("error reading row (block %d) from bigtable for db blocks: %w", number, err)
	}

	if len(row[BT_COLUMNFAMILY_BLOCK]) == 0 { // block not found
		return "", ErrBlockNotFound
	}

	for _, r := range row[BT_COLUMNFAMILY_BLOCK] {
		if r.Column == BT_COLUMN_BLOCK {
			bInfo := &dbBlockHash{}
			err = json.Unmarshal(decompress(r.Value), bInfo)
			if err != nil {
				return "", fmt.Errorf("block %d has an error on unmarshal: %w", number, err)
			}
			dbBlockCache[number] = bInfo.Hash
			return bInfo.Hash, nil
		}
	}
	return "", fmt.Errorf("block %d, no entry for block in bigtable found", number)
}

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
			transactions[i] = tx.Hash.String()
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

func getBlockKey(blockNumber uint64, chainId uint64) string {
	return fmt.Sprintf("%d:%12d", chainId, MAX_EL_BLOCK_NUMBER-blockNumber)
}
