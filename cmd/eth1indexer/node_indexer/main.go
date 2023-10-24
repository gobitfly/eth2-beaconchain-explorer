package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"eth2-exporter/types"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func main() {

	url := flag.String("url", "http://localhost:8545", "")
	blockNumber := flag.Int("block-number", -1, "")
	startBlockNumber := flag.Int("start-block-number", -1, "")
	step := flag.Int("step", 1, "")

	flag.Parse()

	client, err := ethclient.Dial(*url)

	if err != nil {
		logrus.Fatalf("error dialing eth url: %v", err)
	}

	// retrieve the latest block number
	if *startBlockNumber == -1 {
		latestBlockNumber, err := client.BlockNumber(context.Background())
		if err != nil {
			logrus.Fatalf("error retrieving latest block number: %v", err)
		}
		*startBlockNumber = int(latestBlockNumber)
	}

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

	blocksProcessed := int64(0)

	for i := *startBlockNumber; i >= 0; i = i - *step {

		i := i

		gOuter.Go(
			func() error {
				g := &errgroup.Group{}

				var block, receipts, traces []byte

				g.Go(func() error {
					block = getBlock(*url, httpClient, i)
					return nil
				})
				g.Go(func() error {
					receipts = getReceipts(*url, httpClient, i)
					return nil
				})
				g.Go(func() error {
					traces = getTraces(*url, httpClient, i)
					return nil
				})

				g.Wait()

				logrus.Infof("block: %v bytes, receipts: %v bytes, traces: %v bytes, total: %v bytes", len(block), len(receipts), len(traces), len(block)+len(receipts)+len(traces))
				new := atomic.AddInt64(&blocksProcessed, 1)

				if new%1000 == 0 {
					logrus.Infof("scanning blocks for unknown rpc fields, currently at: %v", i)
				}

				return nil
			})

		// if i%100 == 0 || *step > 1 {
		// }

		// logrus.Infof("retrieved block with number %v and hash %v", blockData.Result.Number, blockData.Result.Hash)

	}

}

func getBlock(url string, httpClient *http.Client, number int) []byte {
	body := []byte(fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x%x", true],"id":1}`, number))

	r, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		logrus.Fatalf("error creating post request: %v", err)
	}

	r.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(r)
	if err != nil {
		logrus.Fatalf("error executing post request: %v", err)
	}

	defer res.Body.Close()

	resString, err := io.ReadAll(res.Body)
	if err != nil {
		logrus.Fatalf("error reading request body: %v", err)
	}

	return compress(resString)
}

func getReceipts(url string, httpClient *http.Client, number int) []byte {
	body := []byte(fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockReceipts","params":["0x%x"],"id":1}`, number))

	r, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		logrus.Fatalf("error creating post request: %v", err)
	}

	r.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(r)
	if err != nil {
		logrus.Fatalf("error executing post request: %v", err)
	}

	defer res.Body.Close()

	resString, err := io.ReadAll(res.Body)
	if err != nil {
		logrus.Fatalf("error reading request body: %v", err)
	}

	return compress(resString)
}

func getTraces(url string, httpClient *http.Client, number int) []byte {
	body := []byte(fmt.Sprintf(`{"jsonrpc":"2.0","method":"debug_traceBlockByNumber","params":["0x%x", {"tracer": "callTracer"}],"id":1}`, number))

	r, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		logrus.Fatalf("error creating post request: %v", err)
	}

	r.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(r)
	if err != nil {
		logrus.Fatalf("error executing post request: %v", err)
	}

	defer res.Body.Close()

	resString, err := io.ReadAll(res.Body)
	if err != nil {
		logrus.Fatalf("error reading request body: %v", err)
	}

	return compress(resString)
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
