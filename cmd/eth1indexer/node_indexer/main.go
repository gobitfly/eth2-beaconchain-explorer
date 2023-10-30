package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"eth2-exporter/types"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/davecgh/go-spew/spew"
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

func main() {

	url := flag.String("url", "http://localhost:8545", "")

	discordWebhookReportUrl := flag.String("discord-url", "", "")
	discordWebhookUser := flag.String("discord-user", "", "")
	blockNumber := flag.Int("block-number", -1, "")
	startBlockNumber := flag.Int("start-block-number", 0, "")

	flag.Parse()

	btClient, err := gcp_bigtable.NewClient(context.Background(), "etherchain", "beaconchain-node-data-storage", option.WithGRPCConnectionPool(50))
	if err != nil {
		logrus.Fatal(err)
	}

	tableBlocksRaw := btClient.Open("blocks-raw")

	client, err := ethclient.Dial(*url)

	if err != nil {
		logrus.Fatalf("error dialing eth url: %v", err)
	}

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

	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}

	if *blockNumber != -1 {
		logrus.Infof("checking block %v", *blockNumber)
		getBlock(*url, httpClient, *blockNumber)
		logrus.Info("OK")
		return
	}

	gOuter := &errgroup.Group{}
	gOuter.SetLimit(10)

	muts := []*gcp_bigtable.Mutation{}
	keys := []string{}
	mux := &sync.Mutex{}

	blocksProcessed := int64(0)

	p := message.NewPrinter(language.English)

	for i := *startBlockNumber; i <= int(latestBlockNumber); i++ {

		i := i

		gOuter.Go(func() error {
			for ; ; time.Sleep(time.Second) {

				g := &errgroup.Group{}

				var block, receipts, traces []byte

				g.Go(func() error {
					var err error
					block, err = getBlock(*url, httpClient, i)
					return err
				})
				g.Go(func() error {
					var err error
					receipts, err = getReceipts(*url, httpClient, i)
					return err
				})
				g.Go(func() error {
					var err error
					traces, err = getTraces(*url, httpClient, i)
					return err
				})

				err := g.Wait()

				if err != nil {
					logrus.Errorf("error processing block %v: %v", i, err)
					continue
				}

				mux.Lock()
				mut := gcp_bigtable.NewMutation()
				mut.Set("b", "b", gcp_bigtable.Timestamp(0), block)
				mut.Set("r", "r", gcp_bigtable.Timestamp(0), receipts)
				mut.Set("t", "t", gcp_bigtable.Timestamp(0), traces)

				muts = append(muts, mut)
				key := getBlockKey(i, chainIdUint64)
				keys = append(keys, key)

				if len(keys) == 1000 {
					ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
					errs, err := tableBlocksRaw.ApplyBulk(ctx, keys, muts)

					if err != nil {
						logrus.Fatalf("error writing data to bigtable: %v", err)
					}

					for _, err := range errs {
						logrus.Fatalf("error writing data to bigtable: %v", err)
					}
					cancel()
					logrus.Infof("completed processing block %v (block: %v bytes, receipts: %v bytes, traces: %v bytes, total: %v bytes)", i, len(block), len(receipts), len(traces), len(block)+len(receipts)+len(traces))

					muts = []*gcp_bigtable.Mutation{}
					keys = []string{}

				}
				mux.Unlock()

				if atomic.AddInt64(&blocksProcessed, 1)%100000 == 0 {
					sendMessage(p.Sprintf("OP MAINNET NODE EXPORT: currently at block %v of %v (%.1f%%)", i, latestBlockNumber, float64(i)*100/float64(latestBlockNumber)), *discordWebhookReportUrl, *discordWebhookUser)
				}
				break

			}
			return nil
		})
	}

	gOuter.Wait()

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
func getBlock(url string, httpClient *http.Client, number int) ([]byte, error) {
	body := []byte(fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x%x", true],"id":1}`, number))

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

	resString, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading request body: %v", err)
	}

	if strings.Contains(string(resString), `"error":{"code"`) {
		return nil, fmt.Errorf("rpc error: %s", resString)
	}

	return compress(resString), nil
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

	resString, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading request body: %v", err)
	}

	if strings.Contains(string(resString), `"error":{"code"`) {
		return nil, fmt.Errorf("rpc error: %s", resString)
	}

	return compress(resString), nil
}

func getTraces(url string, httpClient *http.Client, number int) ([]byte, error) {

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

	defer res.Body.Close()

	resString, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading request body: %v", err)
	}

	if strings.Contains(string(resString), `"error":{"code"`) {
		return nil, fmt.Errorf("rpc error: %s", resString)
	}

	return compress(resString), nil
}

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

func getBlockKey(blockNumber int, chainId uint64) string {
	return fmt.Sprintf("%d:%12d", chainId, MAX_EL_BLOCK_NUMBER-blockNumber)
}
