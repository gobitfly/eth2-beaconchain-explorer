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
	"strconv"
	"strings"

	gcp_bigtable "cloud.google.com/go/bigtable"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
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

func main() {
	btProject := flag.String("btproject", "etherchain", "bigtable project name")
	btInstance := flag.String("btinstance", "beaconchain-node-data-storage", "bigtable instance name")
	chainId := flag.Uint64("chainId", 0, "id of the chain to use")

	flag.Parse()

	btClient, err := gcp_bigtable.NewClient(context.Background(), *btProject, *btInstance, option.WithGRPCConnectionPool(1))
	if err != nil {
		utils.LogFatal(err, "creating new client for Bigtable", 0)
	}
	tableBlocksRaw := btClient.Open("blocks-raw")
	if tableBlocksRaw == nil {
		utils.LogFatal(err, "open blocks-raw table", 0)
	}

	checkBlocksFromBigtable(tableBlocksRaw, *chainId)

}
func checkBlocksFromBigtable(tbl *gcp_bigtable.Table, chainId uint64) {
	ctx := context.Background()

	start := fmt.Sprintf("%d:%d", chainId, MAX_EL_BLOCK_NUMBER-1000000)

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

			logrus.Infof("retrieved block %d", blockNumberUint64)

			if blockNumberUint64 != previousNumber-1 && previousNumber != 0 && i > 1000 {
				logrus.Fatalf("%d != %d", blockNumberUint64, previousNumber)
			}

			var blockData, receiptsData, tracesData, unclesData []byte
			blockData = decompress(r[BT_COLUMNFAMILY_BLOCK][0].Value)

			if len(r[BT_COLUMNFAMILY_RECEIPTS]) > 0 {
				receiptsData = decompress(r[BT_COLUMNFAMILY_RECEIPTS][0].Value)
			}
			if len(r[BT_COLUMNFAMILY_TRACES]) > 0 {
				tracesData = decompress(r[BT_COLUMNFAMILY_TRACES][0].Value)
			}
			if len(r[BT_COLUMNFAMILY_UNCLES]) > 0 {
				unclesData = decompress(r[BT_COLUMNFAMILY_UNCLES][0].Value)
			}

			logrus.Infof("%d blocks, %d receipts, %d traces, %d uncles", len(blockData), len(receiptsData), len(tracesData), len(unclesData))

			var blockDataParsed types.Eth1RpcGetBlockResponse
			blocksDecoder := json.NewDecoder(bytes.NewReader(blockData))
			blocksDecoder.DisallowUnknownFields()
			err = blocksDecoder.Decode(&blockDataParsed)
			if err != nil {
				fmt.Println(string(blockData))
				utils.LogFatal(err, "error decoding block", 0)
			}

			// for _, tx := range blockDataParsed.Result.Transactions {

			// 	logrus.Infof("%x", tx.ArbSubType)
			// }

			if len(receiptsData) > 0 {
				if receiptsData[0] == '[' { // receipts were retrieved via a batched eth_getTransactionReceipt rpc call
					var receiptsDataParsed []types.Eth1RpcGetBlockReceiptResponse
					receiptsDecoder := json.NewDecoder(bytes.NewReader(receiptsData))
					receiptsDecoder.DisallowUnknownFields()
					err = receiptsDecoder.Decode(&receiptsDataParsed)
					if err != nil {
						fmt.Println(string(receiptsData))
						utils.LogFatal(err, "error decoding receipts", 0)
					}
				} else if receiptsData[0] == '{' { // receipts were retrieved via the eth_getBlockReceipts rpc call
					var receiptsDataParsed types.Eth1RpcGetBlockReceiptsResponse
					receiptsDecoder := json.NewDecoder(bytes.NewReader(receiptsData))
					receiptsDecoder.DisallowUnknownFields()
					err = receiptsDecoder.Decode(&receiptsDataParsed)
					if err != nil {
						fmt.Println(string(receiptsData))
						utils.LogFatal(err, "error decoding receipts", 0)
					}
				} else {
					logrus.Fatal("invalid receipts data object")
				}
			}

			if len(tracesData) > 0 {
				if !strings.Contains(string(tracesData), `"action"`) { // heuristic to detect if traces were obtained via debug_traceBlock or (arb)trace_block style traces
					var tracesDataParsed types.Eth1RpcTraceBlockResponse
					tracessDecoder := json.NewDecoder(bytes.NewReader(tracesData))
					tracessDecoder.DisallowUnknownFields()
					err = tracessDecoder.Decode(&tracesDataParsed)
					if err != nil {
						fmt.Println(string(tracesData))
						utils.LogFatal(err, "error decoding traces (geth style)", 0)
					}

					// for _, trace := range tracesDataParsed.Result {
					// 	if trace.Result.Error != "" {
					// 		logrus.Fatal(trace.Result.Error)
					// 	}
					// }
				} else {
					var tracesDataParsed types.Eth1RpcDebugTraceBlockResponse
					tracessDecoder := json.NewDecoder(bytes.NewReader(tracesData))
					tracessDecoder.DisallowUnknownFields()
					err = tracessDecoder.Decode(&tracesDataParsed)
					if err != nil {
						fmt.Println(string(tracesData))
						utils.LogFatal(err, "error decoding traces (parity style)", 0)
					}
				}
			}

			if len(unclesData) > 0 {
				var unclesDataParsed []types.Eth1RpcGetBlockResponse
				unclesDecoder := json.NewDecoder(bytes.NewReader(unclesData))
				unclesDecoder.DisallowUnknownFields()
				err = unclesDecoder.Decode(&unclesDataParsed)
				if err != nil {
					fmt.Println(string(unclesData))
					utils.LogFatal(err, "error decoding uncles", 0)
				}
			}

			previousNumber = blockNumberUint64

			i++

			start = fmt.Sprintf("%s\x00", key)
			return true
		}, gcp_bigtable.LimitRows(1000))

		if err != nil {
			logrus.Fatal(err)
		}
	}
}

func decompress(src []byte) []byte {
	if len(src) == 0 {
		return src
	}
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
