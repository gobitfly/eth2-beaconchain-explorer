package main

// imports
import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"eth2-exporter/db"
	"eth2-exporter/hexutil"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/option"

	gcp_bigtable "cloud.google.com/go/bigtable"
)

// defines
const MAX_EL_BLOCK_NUMBER = int64(1_000_000_000_000 - 1)

const BT_COLUMNFAMILY_BLOCK = "b"
const BT_COLUMN_BLOCK = "b"
const BT_COLUMNFAMILY_RECEIPTS = "r"
const BT_COLUMN_RECEIPTS = "r"
const BT_COLUMNFAMILY_TRACES = "t"
const BT_COLUMN_TRACES = "t"
const BT_COLUMNFAMILY_UNCLES = "u"
const BT_COLUMN_UNCLES = "u"

// const MAINNET_CHAINID = 1
// const GOERLI_CHAINID = 5
// const OPTIMISM_CHAINID = 10
// const GNOSIS_CHAINID = 100
// const HOLESKY_CHAINID = 17000
// const SEPOLIA_CHAINID = 11155111
const ARBITRUM_CHAINID = 42161
const ARBITRUM_NITRO_BLOCKNUMBER = 22207815

const HTTP_TIMEOUT_IN_SECONDS = 2 * 120
const MAX_REORG_DEPTH = 100           // that number of block we are looking 'back', includes latest block
const MAX_NODE_REQUESTS_AT_ONCE = 100 // Maximum node requests allowed
const OUTPUT_CYCLE_IN_SECONDS = 10    // duration between 2 outputs / updates, just a visual thing

// errors
var errContextDeadlineExceeded = errors.New("context deadline exceeded (Client.Timeout or context cancellation while reading body)")

// structs
type jsonRpcReturnId struct {
	Id int64 `json:"id"`
}
type fullBlockRawData struct {
	blockNumber      int64
	blockHash        hexutil.Bytes
	blockUnclesCount int
	blockTxs         []string

	blockCompressed    hexutil.Bytes
	receiptsCompressed hexutil.Bytes
	tracesCompressed   hexutil.Bytes
	unclesCompressed   hexutil.Bytes
}
type intRange struct {
	start int64
	end   int64
}

// local globals
var currentNodeBlockNumber atomic.Int64
var elClient *ethclient.Client
var reorgDepth *int64
var httpClient *http.Client

// init
func init() {
	httpClient = &http.Client{Timeout: time.Second * HTTP_TIMEOUT_IN_SECONDS}
}

// main
func main() {
	// read / set parameter
	configPath := flag.String("config", "config/default.config.yml", "Path to the config file")
	startBlockNumber := flag.Int64("start-block-number", -1, "only working in combination with end-block-number, defined block is included")
	endBlockNumber := flag.Int64("end-block-number", -1, "only working in combination with start-block-number, defined block is included")
	reorgDepth = flag.Int64("reorg.depth", 20, fmt.Sprintf("lookback to check and handle chain reorgs (MAX %d)", MAX_REORG_DEPTH))
	concurrency := flag.Int("concurrency", 10, "maximum threads used")
	nodeRequestsAtOnce := flag.Int("node-requests-at-once", 50, fmt.Sprintf("bulk size per node request (MAX %d)", MAX_NODE_REQUESTS_AT_ONCE))
	skipHoleCheck := flag.Bool("skip-hole-check", false, "skips the initial check for holes")
	flag.Parse()

	// tell the user about all parameter
	{
		logrus.Infof("config set to '%s'", *configPath)
		if *startBlockNumber >= 0 {
			logrus.Infof("start-block-number set to '%d'", *startBlockNumber)
		}
		if *endBlockNumber >= 0 {
			logrus.Infof("end-block-number set to '%d'", *endBlockNumber)
		}
		logrus.Infof("reorg.depth set to '%d'", *reorgDepth)
		logrus.Infof("concurrency set to '%d'", *concurrency)
		logrus.Infof("node-requests-at-once set to '%d'", *nodeRequestsAtOnce)
		if *skipHoleCheck {
			logrus.Infof("skip-hole-check set true")
		}
	}

	// check config
	{
		logrus.WithField("config", *configPath).WithField("version", version.Version).Printf("starting")
		cfg := &types.Config{}
		err := utils.ReadConfig(cfg, *configPath)
		if err != nil {
			utils.LogFatal(err, "error reading config file", 0) // fatal, as there is no point without a config
		} else {
			logrus.Info("reading config completed")
		}
		utils.Config = cfg
	}

	// check parameters
	if *nodeRequestsAtOnce < 1 {
		logrus.Warnf("node-requests-at-once set to %d, corrected to 1", *nodeRequestsAtOnce)
		*nodeRequestsAtOnce = 1
	}
	if *nodeRequestsAtOnce > MAX_NODE_REQUESTS_AT_ONCE {
		logrus.Warnf("node-requests-at-once set to %d, corrected to %d", *nodeRequestsAtOnce, MAX_NODE_REQUESTS_AT_ONCE)
		*nodeRequestsAtOnce = MAX_NODE_REQUESTS_AT_ONCE
	}
	if *reorgDepth < 0 || *reorgDepth > MAX_REORG_DEPTH {
		logrus.Warnf("reorg.depth parameter set to %d, corrected to %d", *reorgDepth, MAX_REORG_DEPTH)
		*reorgDepth = MAX_REORG_DEPTH
	}
	if *concurrency < 1 {
		logrus.Warnf("concurrency parameter set to %d, corrected to 1", *concurrency)
		*concurrency = 1
	}

	// init postgres
	{
		db.MustInitDB(&types.DatabaseConfig{
			Username:     utils.Config.WriterDatabase.Username,
			Password:     utils.Config.WriterDatabase.Password,
			Name:         utils.Config.WriterDatabase.Name,
			Host:         utils.Config.WriterDatabase.Host,
			Port:         utils.Config.WriterDatabase.Port,
			MaxOpenConns: utils.Config.WriterDatabase.MaxOpenConns,
			MaxIdleConns: utils.Config.WriterDatabase.MaxIdleConns,
		}, &types.DatabaseConfig{
			Username:     utils.Config.ReaderDatabase.Username,
			Password:     utils.Config.ReaderDatabase.Password,
			Name:         utils.Config.ReaderDatabase.Name,
			Host:         utils.Config.ReaderDatabase.Host,
			Port:         utils.Config.ReaderDatabase.Port,
			MaxOpenConns: utils.Config.ReaderDatabase.MaxOpenConns,
			MaxIdleConns: utils.Config.ReaderDatabase.MaxIdleConns,
		})
		defer db.ReaderDb.Close()
		defer db.WriterDb.Close()
		logrus.Info("starting postgres completed")
	}

	// init bigtable
	logrus.Info("init BT...")
	btClient, err := gcp_bigtable.NewClient(context.Background(), utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, option.WithGRPCConnectionPool(1))
	if err != nil {
		utils.LogFatal(err, "creating new client for Bigtable", 0) // fatal, no point to continue without BT
	}
	tableBlocksRaw := btClient.Open("blocks-raw")
	if tableBlocksRaw == nil {
		utils.LogFatal(err, "open blocks-raw table", 0) // fatal, no point to continue without BT
	}
	defer btClient.Close()
	logrus.Info("...init BT done.")

	// init el client
	logrus.Info("init el client endpoint...")
	elClient, err = ethclient.Dial(utils.Config.Eth1RpcEndpoint)
	if err != nil {
		utils.LogFatal(err, "error dialing eth url", 0) // fatal, no point to continue without node connection
	}
	logrus.Info("...init el client endpoint done.")

	// check chain id
	{
		// utils.Config.Chain.Id = ARBITRUM_CHAINID // #RECY REMOVE
		logrus.Info("check chain id...")
		chainID, err := rpciGetChainId()
		if err != nil {
			utils.LogFatal(err, "error get chain id", 0) // fatal, no point to continue without chain id
		}
		if chainID != utils.Config.Chain.Id { // if the chain id is removed from the config, just remove this if, there is no point, except checking consistency
			utils.LogFatal(err, "node chain different from config chain", 0) // fatal, config doesn't match node
		}
		logrus.Info("...check chain id done.")
	}

	// get latest block (as it's global, so we have a initial value)
	logrus.Info("get latest block from node...")
	updateBlockNumber(false)
	logrus.Info("...get latest block from node done.")

	// //////////////////////////////////////////
	// Config done, now actually "doing" stuff //
	// //////////////////////////////////////////

	/* #RECY REMOVE
	prefix := "11155111:999999826267"
	readBT(tableBlocksRaw, prefix, BT_COLUMNFAMILY_BLOCK, BT_COLUMN_BLOCK)
	readBT(tableBlocksRaw, prefix, BT_COLUMNFAMILY_RECEIPTS, BT_COLUMN_RECEIPTS)
	readBT(tableBlocksRaw, prefix, BT_COLUMNFAMILY_TRACES, BT_COLUMN_TRACES)
	readBT(tableBlocksRaw, prefix, BT_COLUMNFAMILY_UNCLES, BT_COLUMN_UNCLES)
	return
	*/

	// check if reexport requested
	if *startBlockNumber >= 0 && *endBlockNumber >= 0 && *startBlockNumber <= *endBlockNumber {
		logrus.Infof("Found REEXPORT for block %v to %v...", *startBlockNumber, *endBlockNumber)
		err := bulkExportBlocksRange(tableBlocksRaw, []intRange{intRange{start: *startBlockNumber, end: *endBlockNumber}}, *concurrency, *nodeRequestsAtOnce)
		if err != nil {
			utils.LogFatal(err, "error while reexport blocks for bigtable (reexport range)", 0) // fatal, as there is nothing more todo anyway
		}
		logrus.Info("Job done, have a nice day :)")
		return
	}

	// find holes in our previous runs / sanity check
	if *skipHoleCheck {
		logrus.Warn("Skipping hole check!")
	} else {
		logrus.Info("Checking for holes...")
		startTime := time.Now()
		missingBlocks, err := psqlFindGaps() // find the holes
		findHolesTook := time.Since(startTime)
		if err != nil {
			utils.LogFatal(err, "error checking for holes", 0) // fatal, as we highly depend on postgres, if this is not working, we can quit
		}
		l := len(missingBlocks)
		if l > 0 { // some holes found
			logrus.Warnf("Found %d missing block ranges in %v, fixing them now...", l, findHolesTook)
			if l <= 10 {
				logrus.Warnf("%v", missingBlocks)
			} else {
				logrus.Warnf("%v<...>", missingBlocks[:10])
			}
			startTime = time.Now()
			err := bulkExportBlocksRange(tableBlocksRaw, missingBlocks, *concurrency, *nodeRequestsAtOnce) // reexport the holes
			if err != nil {
				utils.LogFatal(err, "error while reexport blocks for bigtable (fixing holes)", 0) // fatal, as if we wanna start with holes, we should set the skip-hole-check parameter
			}
			logrus.Warnf("...fixed them in %v", time.Since(startTime))
		} else {
			logrus.Infof("...no missing block found in %v", findHolesTook)
		}
	}

	// waiting for new blocks and export them, while checking reorg before every new block
	latestPGBlock, err := psqlGetLatestBlock(false)
	if err != nil {
		utils.LogFatal(err, "error while using psqlGetLatestBlock (start / read)", 0) // fatal, as if there is no inital value, we have nothing to start from
	}
	var consecutiveErrorCount int
	consecutiveErrorCountThreshold := 5 // after threshold + 1 errors it will be fatal instead
	for {
		currentNodeBN := currentNodeBlockNumber.Load()
		if currentNodeBN < latestPGBlock {
			// fatal, as this is an impossible error
			utils.LogFatal(err, "impossible error currentNodeBN < lastestPGBlock", 0, map[string]interface{}{"currentNodeBN": currentNodeBN, "latestPGBlock": latestPGBlock})
		} else if currentNodeBN == latestPGBlock {
			time.Sleep(time.Second)
			continue // still the same block
		} else {
			consecutiveErrorCountOld := consecutiveErrorCount

			// checking for reorg
			if *reorgDepth > 0 && latestPGBlock >= 0 {
				// define length to check
				l := *reorgDepth
				if l > latestPGBlock+1 {
					l = latestPGBlock + 1
				}

				// fill array with block numbers to check
				blockRawData := make([]fullBlockRawData, l)
				for i := int64(0); i < l; i++ {
					blockRawData[i].blockNumber = latestPGBlock + i - l + 1
				}

				// get all hashes from node
				err = rpciGetBulkBlockRawHash(blockRawData, *nodeRequestsAtOnce)
				if err != nil {
					consecutiveErrorCount++
					if consecutiveErrorCount <= consecutiveErrorCountThreshold {
						utils.LogError(err, "error when bulk getting raw block hashes", 0, map[string]interface{}{"reorgErrorCount": consecutiveErrorCount, "latestPGBlock": latestPGBlock, "reorgDepth": *reorgDepth})
					} else {
						utils.LogFatal(err, "error when bulk getting raw block hashes", 0, map[string]interface{}{"reorgErrorCount": consecutiveErrorCount, "latestPGBlock": latestPGBlock, "reorgDepth": *reorgDepth})
					}
				}

				// get a list of all block_ids where the hashes are fine
				var matchingHashesBlockIdList []int64
				matchingHashesBlockIdList, err = psqlGetHashHitsIdList(blockRawData)
				if err != nil {
					consecutiveErrorCount++
					if consecutiveErrorCount <= consecutiveErrorCountThreshold {
						utils.LogError(err, "error when getting hash hits id list", 0, map[string]interface{}{"reorgErrorCount": consecutiveErrorCount, "latestPGBlock": latestPGBlock, "reorgDepth": *reorgDepth})
					} else {
						utils.LogFatal(err, "error when getting hash hits id list", 0, map[string]interface{}{"reorgErrorCount": consecutiveErrorCount, "latestPGBlock": latestPGBlock, "reorgDepth": *reorgDepth})
					}
				}

				matchingLength := len(matchingHashesBlockIdList)
				if len(blockRawData) != matchingLength { // nothing todo if all elements are fine, but if not...
					if len(blockRawData) < matchingLength {
						// fatal, as this is an impossible error
						utils.LogFatal(err, "impossible error len(blockRawData) < matchingLength", 0, map[string]interface{}{"latestPGBlock": latestPGBlock, "matchingLength": matchingLength})
					}

					// reverse the "fine" list, so we have a "not fine" list
					wrongHashRanges := []intRange{intRange{start: -1}}
					wrongHashRangesIndex := 0
					var i int
					var failCounter int
					for _, v := range blockRawData {
						for i < matchingLength && v.blockNumber > matchingHashesBlockIdList[i] {
							i++
						}
						if i > matchingLength || v.blockNumber != matchingHashesBlockIdList[i] {
							failCounter++
							if wrongHashRanges[wrongHashRangesIndex].start < 0 {
								wrongHashRanges[wrongHashRangesIndex].start = v.blockNumber
								wrongHashRanges[wrongHashRangesIndex].end = v.blockNumber
							} else if wrongHashRanges[wrongHashRangesIndex].end+1 == v.blockNumber {
								wrongHashRanges[wrongHashRangesIndex].end = v.blockNumber
							} else {
								wrongHashRangesIndex++
								wrongHashRanges[wrongHashRangesIndex].start = v.blockNumber
								wrongHashRanges[wrongHashRangesIndex].end = v.blockNumber
							}
						}
					}
					if failCounter != len(blockRawData)-matchingLength {
						// fatal, as this is an impossible error
						utils.LogFatal(err, "impossible error failureLength != len(blockRawData)-matchingLength", 0, map[string]interface{}{"failCounter": failCounter, "len(blockRawData)-matchingLength": len(blockRawData) - matchingLength})
					}
					logrus.Infof("found %d wrong hashes when checking for reorgs, reexporting them now...", failCounter)
					logrus.Infof("%v", wrongHashRanges)

					// export the hits again
					err = bulkExportBlocksRange(tableBlocksRaw, wrongHashRanges, *concurrency, *nodeRequestsAtOnce)
					if err != nil {
						consecutiveErrorCount++
						if consecutiveErrorCount <= consecutiveErrorCountThreshold {
							utils.LogError(err, "error exporting hits on reorg", 0, map[string]interface{}{"reorgErrorCount": consecutiveErrorCount, "len(blockRawData)": len(blockRawData), "reorgDepth": *reorgDepth, "matchingHashesBlockIdList": matchingHashesBlockIdList, "wrongHashRanges": wrongHashRanges})
						} else {
							utils.LogFatal(err, "error exporting hits on reorg", 0, map[string]interface{}{"reorgErrorCount": consecutiveErrorCount, "len(blockRawData)": len(blockRawData), "reorgDepth": *reorgDepth, "matchingHashesBlockIdList": matchingHashesBlockIdList, "wrongHashRanges": wrongHashRanges})
						}
					} else {
						logrus.Info("...done. Everything fine with reorgs again.")
					}
				}
			}

			// export all new blocks
			{
				newerNodeBN := currentNodeBlockNumber.Load() // just in case it took a while doing the reorg stuff, no problem if range > reorg limit, as the exported blocks will be newest also
				if newerNodeBN < currentNodeBN {
					// fatal, as this is an impossible error
					utils.LogFatal(err, "impossible error newerNodeBN < currentNodeBN", 0, map[string]interface{}{"newerNodeBN": newerNodeBN, "currentNodeBN": currentNodeBN})
				}
				err = bulkExportBlocksRange(tableBlocksRaw, []intRange{intRange{start: latestPGBlock + 1, end: newerNodeBN}}, *concurrency, *nodeRequestsAtOnce)
				if err != nil {
					consecutiveErrorCount++
					if consecutiveErrorCount <= consecutiveErrorCountThreshold {
						utils.LogError(err, "error while reexport blocks for bigtable (newest blocks)", 0, map[string]interface{}{"reorgErrorCount": consecutiveErrorCount, "latestPGBlock+1": latestPGBlock + 1, "newerNodeBN": newerNodeBN})
					} else {
						utils.LogFatal(err, "error while reexport blocks for bigtable (newest blocks)", 0, map[string]interface{}{"reorgErrorCount": consecutiveErrorCount, "latestPGBlock+1": latestPGBlock + 1, "newerNodeBN": newerNodeBN})
					}
				} else {
					latestPGBlock, err = psqlGetLatestBlock(true)
					if err != nil {
						consecutiveErrorCount++
						if consecutiveErrorCount <= consecutiveErrorCountThreshold {
							utils.LogError(err, "error while using psqlGetLatestBlock (ongoing / write)", 0, map[string]interface{}{"reorgErrorCount": consecutiveErrorCount})
						} else {
							utils.LogFatal(err, "error while using psqlGetLatestBlock (ongoing / write)", 0, map[string]interface{}{"reorgErrorCount": consecutiveErrorCount})
						}
					} else if latestPGBlock != newerNodeBN {
						// fatal, as this is a nearly impossible error
						utils.LogFatal(err, "impossible error latestPGBlock != newerNodeBN", 0, map[string]interface{}{"latestPGBlock": latestPGBlock, "newerNodeBN": newerNodeBN})
					}
				}
			}

			// reset consecutive error count if no change during this run
			if consecutiveErrorCount > 0 && consecutiveErrorCountOld == consecutiveErrorCount {
				consecutiveErrorCount = 0
				logrus.Infof("reset consecutive error count to 0, as no error in this run (was %d)", consecutiveErrorCountOld)
			}
		}
	}
}

// #RECY REMOVE after testing
func readBT(tableBlocksRaw *gcp_bigtable.Table, prefix string, family string, columnFilter string) error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancel()

	rowRange := gcp_bigtable.PrefixRange(prefix)
	rowHandler := func(row gcp_bigtable.Row) bool {
		logrus.Warnf("%s", decompress(row[family][0].Value))
		return true
	}

	err := tableBlocksRaw.ReadRows(ctx, rowRange, rowHandler, gcp_bigtable.LimitRows(1), gcp_bigtable.RowFilter(gcp_bigtable.ColumnFilter(columnFilter)))
	if err != nil {
		logrus.Errorf("%v", err)
	}

	return nil
}

// export all blocks, heavy use of bulk & concurrency, providing a block raw data array (used by the other bulkExportBlocks+ functions)
func _bulkExportBlocks(tableBlocksRaw *gcp_bigtable.Table, blockRawData []fullBlockRawData, nodeRequestsAtOnce int) error {
	// check values
	{
		if tableBlocksRaw == nil {
			return fmt.Errorf("tableBlocksRaw == nil")
		}

		l := len(blockRawData)
		if l < 1 || l > nodeRequestsAtOnce {
			return fmt.Errorf("blockRawData length (%d) is 0 or greater 'node requests at once' (%d)", l, nodeRequestsAtOnce)
		}
	}

	// get block_hash, block_unclesCount, block_compressed & block_txs
	err := rpciGetBulkBlockRawData(blockRawData, nodeRequestsAtOnce)
	if err != nil {
		return err
	}
	err = rpciGetBulkRawUncles(blockRawData, nodeRequestsAtOnce)
	if err != nil {
		return err
	}
	err = rpciGetBulkRawReceipts(blockRawData, nodeRequestsAtOnce)
	if err != nil {
		return err
	}
	err = rpciGetBulkRawTraces(blockRawData, nodeRequestsAtOnce)
	if err != nil {
		return err
	}

	// write to bigtable
	{
		// prepare array
		muts := []*gcp_bigtable.Mutation{}
		keys := []string{}
		for _, v := range blockRawData {
			mut := gcp_bigtable.NewMutation()
			mut.Set(BT_COLUMNFAMILY_BLOCK, BT_COLUMN_BLOCK, gcp_bigtable.Timestamp(0), v.blockCompressed)
			mut.Set(BT_COLUMNFAMILY_RECEIPTS, BT_COLUMN_RECEIPTS, gcp_bigtable.Timestamp(0), v.receiptsCompressed)
			mut.Set(BT_COLUMNFAMILY_TRACES, BT_COLUMN_TRACES, gcp_bigtable.Timestamp(0), v.tracesCompressed)
			if v.blockUnclesCount > 0 {
				mut.Set(BT_COLUMNFAMILY_UNCLES, BT_COLUMN_UNCLES, gcp_bigtable.Timestamp(0), v.unclesCompressed)
			}
			muts = append(muts, mut)
			keys = append(keys, fmt.Sprintf("%d:%12d", utils.Config.Chain.Id, MAX_EL_BLOCK_NUMBER-int64(v.blockNumber)))
		}

		// write
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		var errs []error
		errs, err = tableBlocksRaw.ApplyBulk(ctx, keys, muts)
		if err != nil {
			return err
		}
		for _, e := range errs {
			return e
		}
	}

	// write to SQL
	err = psqlAddElements(blockRawData)
	if err != nil {
		return err
	}

	return nil
}

// export all blocks, heavy use of bulk & concurrency, providing a range array
func bulkExportBlocksRange(tableBlocksRaw *gcp_bigtable.Table, blockRanges []intRange, concurrency int, nodeRequestsAtOnce int) error {
	var blocksTotalCount int64
	{
		l := len(blockRanges)
		if l <= 0 {
			return fmt.Errorf("got empty blockRanges array")
		}
		for i, v := range blockRanges {
			if v.start <= v.end {
				blocksTotalCount += v.end - v.start + 1
			} else {
				return fmt.Errorf("blockRanges at index %d has wrong start (%d) > end (%d) combination", i, v.start, v.end)
			}
		}

		if l == 1 {
			logrus.Infof("Only 1 range found, started export of blocks %d to %d, total block amount %d, using an updater every %d seconds for more details.", blockRanges[0].start, blockRanges[0].end, blocksTotalCount, OUTPUT_CYCLE_IN_SECONDS)
		} else {
			logrus.Infof("%d ranges found, total block amount %d, using an updater every %d seconds for more details.", l, blocksTotalCount, OUTPUT_CYCLE_IN_SECONDS)
		}
	}

	gOuterMustStop := atomic.Bool{}
	gOuter := &errgroup.Group{}
	gOuter.SetLimit(concurrency)

	t := time.NewTicker(time.Second * OUTPUT_CYCLE_IN_SECONDS)
	totalStart := time.Now()
	exportStart := totalStart
	blocksProcessedTotal := atomic.Int64{}
	blocksProcessedIntv := atomic.Int64{}

	go func() {
		for {
			<-t.C
			if gOuterMustStop.Load() {
				break
			}

			bpi := blocksProcessedIntv.Swap(0)
			newStart := time.Now()
			blocksProcessedTotal.Add(bpi)
			bpt := blocksProcessedTotal.Load()

			remainingBlocks := blocksTotalCount - bpt
			blocksPerSecond := float64(bpi) / time.Since(exportStart).Seconds()
			blocksPerSecondTotal := float64(bpt) / time.Since(totalStart).Seconds()
			secondsRemaining := float64(remainingBlocks) / float64(blocksPerSecond)
			durationRemaining := time.Second * time.Duration(secondsRemaining)
			durationRemainingTotal := time.Second * time.Duration(float64(remainingBlocks)/float64(blocksPerSecondTotal))

			logrus.Infof("current speed: %0.1f blocks/sec, %0.1f total/sec, %d blocks remaining (%v / %v to go)", blocksPerSecond, blocksPerSecondTotal, remainingBlocks, durationRemaining, durationRemainingTotal)
			exportStart = newStart
		}
	}()
	defer gOuterMustStop.Store(true) // kill the updater

	blockRawData := make([]fullBlockRawData, 0, nodeRequestsAtOnce)
	blockRawDataLen := 0
Loop:
	for _, blockRange := range blockRanges {
		current := blockRange.start
		for blockRange.end-current+1 > 0 {
			if gOuterMustStop.Load() {
				break Loop
			}
			for blockRawDataLen < nodeRequestsAtOnce && current <= blockRange.end {
				blockRawData = append(blockRawData, fullBlockRawData{blockNumber: current})
				blockRawDataLen++
				current++
			}
			if blockRawDataLen == nodeRequestsAtOnce {
				brd := blockRawData
				gOuter.Go(func() error {
					err := _bulkExportBlocks(tableBlocksRaw, brd, nodeRequestsAtOnce)
					if err != nil {
						gOuterMustStop.Store(true)
						return err
					}
					blocksProcessedIntv.Add(int64(len(brd)))
					return nil
				})
				blockRawData = make([]fullBlockRawData, 0, nodeRequestsAtOnce)
				blockRawDataLen = 0
			}
		}
	}

	// write the rest
	if !gOuterMustStop.Load() && blockRawDataLen > 0 {
		brd := blockRawData
		gOuter.Go(func() error {
			err := _bulkExportBlocks(tableBlocksRaw, brd, nodeRequestsAtOnce)
			if err != nil {
				gOuterMustStop.Store(true)
				return err
			}
			blocksProcessedIntv.Add(int64(len(brd)))
			return nil
		})
	}

	return gOuter.Wait()
}

// //////////
// HELPERs //
// //////////
// compress given byte slice
func compress(src []byte) []byte {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(src); err != nil {
		utils.LogFatal(err, "error writing to gzip writer", 0) // fatal, as if this is not working in the first place, it will never work
	}
	if err := zw.Close(); err != nil {
		utils.LogFatal(err, "error closing gzip writer", 0) // fatal, as if this is not working in the first place, it will never work
	}
	return buf.Bytes()
}

// decompress given byte slice
func decompress(src []byte) []byte {
	zr, err := gzip.NewReader(bytes.NewReader(src))
	if err != nil {
		utils.LogFatal(err, "error creating gzip reader", 0) // fatal, as if this is not working in the first place, it will never work
	}
	data, err := io.ReadAll(zr)
	if err != nil {
		utils.LogFatal(err, "error reading from gzip reader", 0) // fatal, as if this is not working in the first place, it will never work
	}
	return data
}

// used by splitAndVerifyJsonArray to add an element to the list depending on its Id
func _splitAndVerifyJsonArrayAddElement(r *[][]byte, element []byte, lastId int64) (int64, error) {
	// adding empty elements will cause issues, so we don't allow it
	if len(element) <= 0 {
		return -1, fmt.Errorf("error, tried to add empty element, lastId (%d)", lastId)
	}

	// unmarshal
	data := &jsonRpcReturnId{}
	err := json.Unmarshal(element, data)
	if err != nil {
		return -1, fmt.Errorf("error decoding '%s': %w", element, err)
	}

	// negativ ids signals an issue
	if data.Id < 0 {
		return -1, fmt.Errorf("error, provided Id (%d) < 0", data.Id)
	}
	// id must ascending or equal
	if data.Id < lastId {
		return -1, fmt.Errorf("error, provided Id (%d) < lastId (%d)", data.Id, lastId)
	}

	// new element
	if data.Id != lastId {
		*r = append(*r, element)
	} else { // append element (same id)
		i := len(*r) - 1
		if (*r)[i][0] == byte('[') {
			(*r)[i] = (*r)[i][1 : len((*r)[i])-1]
		}
		(*r)[i] = append(append(append(append([]byte("["), (*r)[i]...), byte(',')), element...), byte(']'))
	}

	return data.Id, nil
}

// split a bulk json request in single requests
func _splitAndVerifyJsonArray(jArray []byte, providedElementCount int) ([][]byte, error) {
	endDigit := byte('}')
	searchValue := []byte(`{"jsonrpc":"`)
	searchLen := len(searchValue)
	foundElementCount := 0

	// remove everything before the first hit
	i := bytes.Index(jArray, searchValue)
	if i < 0 {
		return nil, fmt.Errorf("no element found")
	}
	jArray = jArray[i:]

	// find all elements
	var err error
	lastId := int64(-1)
	r := make([][]byte, 0)
	for {
		if len(jArray) < searchLen { // weird corner case, shouldn't happen at all
			i = -1
		} else { // get next hit / ignore current (at index 0)
			i = bytes.Index(jArray[searchLen:], searchValue)
		}
		// handle last element
		if i < 0 {
			for l := len(jArray) - 1; l >= 0 && jArray[l] != endDigit; l-- {
				jArray = jArray[:l]
			}
			foundElementCount++
			_, err = _splitAndVerifyJsonArrayAddElement(&r, jArray, lastId)
			if err != nil {
				return nil, fmt.Errorf("error calling split and verify json array add element - last element: %w", err)
			}
			break
		}
		// handle normal element
		foundElementCount++
		lastId, err = _splitAndVerifyJsonArrayAddElement(&r, jArray[:i+searchLen-1], lastId)
		if err != nil {
			return nil, fmt.Errorf("error calling split and verify json array add element: %w", err)
		}
		// set cursor to new start
		jArray = jArray[i+searchLen:]
	}
	if foundElementCount != providedElementCount {
		return r, fmt.Errorf("provided element count %d doesn't match found %d", providedElementCount, foundElementCount)
	}
	return r, nil
}

// #RECY REMOVE after testing
// join int ranges for a better "look"
/*
func joinIntRanges(iRange []intRange) []intRange {
	if len(iRange) < 1 {
		return iRange
	}
	for cleanRun := false; !cleanRun; {
		cleanRun = true
		l := len(iRange)
		for i := 0; cleanRun && i < l; i++ {
			for k, v := range iRange {
				if i != k && iRange[i].end+1 == v.start {
					iRange[i].end = v.end
					iRange[k] = iRange[0]
					iRange = iRange[1:]
					cleanRun = false
					break
				}
			}
		}
	}
	return iRange
}
*/

// get newest block number from node, should be called always with FALSE
func updateBlockNumber(loop bool) {
	/*
		#RECY_ QUESTION wanna use eth_subscribe, but seems not possible?
		Command: curl -X POST -H "Content-Type: application/json" --data '{"id":1,"jsonrpc":"2.0","method":"eth_subscribe","params":["newHeads"]}' localhost:18545
		Result: {"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"notifications not supported"}}
		Environment: our current Sepolia archive node
		=> Look like we have to use ws:// endpoint to make this working
	*/
	sleepTime := time.Second * time.Duration(utils.Config.Chain.ClConfig.SecondsPerSlot/2) // checking twice in the avg duration
	consecutivelyErrorCount := -1
	noNewBlockCount := 0
	for {
		blockNumber, err := rpciGetLatestBlock()
		if err == nil {
			consecutivelyErrorCount = 0
			if currentNodeBlockNumber.Load() != blockNumber {
				currentNodeBlockNumber.Store(blockNumber)
				noNewBlockCount = 0
			} else {
				noNewBlockCount++
				if noNewBlockCount >= 5*60/int(sleepTime.Seconds()) { // 5 minutes
					// fatal, as if it's not working after so many tries, it's very likly it will never work
					utils.LogFatal(nil, "fatal, no new block from node for a while", 0, map[string]interface{}{"noNewBlockCount": noNewBlockCount, "blockNumber": blockNumber})
				} else if noNewBlockCount >= 3*60/int(sleepTime.Seconds()) { // 3 minutes
					utils.LogError(nil, "error, no new block from node for a while", 0, map[string]interface{}{"noNewBlockCount": noNewBlockCount, "blockNumber": blockNumber})
				} else if noNewBlockCount >= 1*60/int(sleepTime.Seconds()) { // 1 minute
					logrus.Warnf("warning, no new block from node for a while (%d tries), block number : %d", noNewBlockCount, blockNumber)
				}
			}
		} else if consecutivelyErrorCount < 0 {
			// fatal, as we need an initial value, otherwise the result will be random
			utils.LogFatal(err, "error get latest block number from node, first try", 0)
		} else {
			consecutivelyErrorCount++
			if consecutivelyErrorCount <= 3 {
				logrus.Warnf("error get latest block number from node. consecutivelyErrorCount: %d", consecutivelyErrorCount)
			} else if consecutivelyErrorCount <= 10 {
				utils.LogError(err, "error get latest block number from node", 0, map[string]interface{}{"consecutivelyErrorCount": consecutivelyErrorCount})
			} else {
				// fatal, as if it's not working after so many tries, it's very likly it will never work
				utils.LogFatal(err, "error get latest block number from node", 0, map[string]interface{}{"consecutivelyErrorCount": consecutivelyErrorCount})
			}
		}
		if !loop {
			go updateBlockNumber(true)
			break
		}
		time.Sleep(sleepTime)
	}
}

// /////////////////////
// Postgres interface //
// /////////////////////

// #RECY REMOVE after testing
// used by findHoles function, doing recursion stuff
/*
func _psqlCheckHoles(start int64, end int64) ([]intRange, error) {
	targetAmount := end - start + 1
	if targetAmount < 1 {
		return nil, fmt.Errorf("error end (%d) > start (%d) in _psqlCheckHoles", end, start)
	}

	var blockAmount int64
	err := db.ReaderDb.Get(&blockAmount, `SELECT COUNT(*) FROM raw_block_status WHERE block_id >= $1 AND block_id <= $2;`, start, end)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			blockAmount = 0
		} else {
			return nil, fmt.Errorf("error at 'SELECT COUNT(*) FROM raw_block_status WHERE block_id >= %d AND block_id <= %d;': %w", start, end, err)
		}
	}

	if targetAmount == blockAmount {
		return nil, nil // best case, every as expected
	} else if targetAmount > blockAmount {
		// complete range not found
		if blockAmount == 0 {
			return []intRange{intRange{start: start, end: start + targetAmount - 1}}, nil
		}

		// split range in low / high
		middle := start + targetAmount/2
		lowBound, err := _psqlCheckHoles(start, middle-1)
		if err != nil {
			return nil, err
		}
		if blockAmount+int64(len(lowBound)) == targetAmount { // no need to check high bound, if already enough missing elements found
			return lowBound, nil
		}
		highBound, err := _psqlCheckHoles(middle, end)
		if err != nil {
			return nil, err
		}
		return append(lowBound, highBound...), nil
	}

	return nil, fmt.Errorf("impossible error, targetAmount (%d) < blockAmount (%d)", targetAmount, blockAmount)
}
*/

// #RECY REMOVE after testing
// find holes (missing ids) in raw_block_status. Starting at 0 and ending at current highest index.
// using _checkHoles function for the recursion stuff
/*
func psqlFindHoles() ([]intRange, error) {
	latestBlock, err := psqlGetLatestBlock(false)
	if err != nil {
		return nil, err
	} else if latestBlock < 0 { // no holes if no entries
		return nil, nil
	}
	iRange, err := _psqlCheckHoles(0, latestBlock)
	if err != nil {
		return nil, err
	}
	return joinIntRanges(iRange), nil
}
*/

func psqlFindGaps() ([]intRange, error) {
	gaps := []intRange{}

	// check for a gap at the beginning
	{
		var firstBlock int64
		err := db.ReaderDb.Get(&firstBlock, `SELECT block_id FROM raw_block_status ORDER BY block_id LIMIT 1;`)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) { // no entries = no gaps
				return []intRange{}, nil
			}
			return []intRange{}, fmt.Errorf("error reading first block from postgres: %w", err)
		}
		if firstBlock != 0 {
			gaps = append(gaps, intRange{start: 0, end: firstBlock - 1})
		}
	}

	// check for gaps everywhere else
	rows, err := db.ReaderDb.Query(`
		SELECT 
			block_id + 1 as gapStart, 
			nextNumber - 1 as gapEnd
		FROM 
			(
			SELECT 
				block_id, LEAD(block_id) OVER (ORDER BY block_id) as nextNumber
			FROM
				raw_block_status
			) number
		WHERE 
			block_id + 1 <> nextNumber
		ORDER BY
			gapStart;`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return gaps, nil
		}
		return []intRange{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var gap intRange
		err := rows.Scan(&gap.start, &gap.end)
		if err != nil {
			return []intRange{}, err
		}
		gaps = append(gaps, gap)
	}

	return gaps, nil
}

// get latest block in postgres db
func psqlGetLatestBlock(useWriterDb bool) (int64, error) {
	var err error
	var latestBlock int64
	query := `SELECT block_id FROM raw_block_status ORDER BY block_id DESC LIMIT 1;`
	if useWriterDb {
		err = db.WriterDb.Get(&latestBlock, query)
	} else {
		err = db.ReaderDb.Get(&latestBlock, query)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return -1, nil
		}
		return -1, fmt.Errorf("error reading latest block in postgres: %w", err)
	}
	return latestBlock, nil
}

// will add elements to sql, based on blockRawData
// on conflict, it will only overwrite / change current entry if hash is different
func psqlAddElements(blockRawData []fullBlockRawData) error {
	l := len(blockRawData)
	if l <= 0 {
		return fmt.Errorf("error, got empty blockRawData array (%d)", l)
	}

	block_number := make([]int64, l)
	block_hash := make(pq.ByteaArray, l)
	for i, v := range blockRawData {
		block_number[i] = v.blockNumber
		block_hash[i] = v.blockHash
	}

	_, err := db.WriterDb.Exec(`
		INSERT INTO raw_block_status
			(block_id, block_hash)
		SELECT
			UNNEST($1::int[]),
			UNNEST($2::bytea[][])
		ON CONFLICT (block_id) DO
			UPDATE SET
				block_hash = excluded.block_hash,
				indexed_bt = FALSE
			WHERE
				raw_block_status.block_hash != excluded.block_hash;`,
		pq.Array(block_number), block_hash)
	return err
}

// will return a list of all provided block_ids where the hash in the database matches the provided list
func psqlGetHashHitsIdList(blockRawData []fullBlockRawData) ([]int64, error) {
	l := len(blockRawData)
	if l <= 0 {
		return nil, fmt.Errorf("error, got empty blockRawData array (%d)", l)
	}

	block_number := make([]int64, l)
	block_hash := make(pq.ByteaArray, l)
	for i, v := range blockRawData {
		block_number[i] = v.blockNumber
		block_hash[i] = v.blockHash
	}

	// as there are corner cases, to be on the safe side, we will use WriterDb here
	rows, err := db.WriterDb.Query(`
		SELECT 
			raw_block_status.block_id 
		FROM 
			raw_block_status, 
			(SELECT UNNEST($1::int[]) as block_id, UNNEST($2::bytea[][]) as block_hash) as node_block_status 
		WHERE 
			raw_block_status.block_id = node_block_status.block_id 
			AND 
			raw_block_status.block_hash = node_block_status.block_hash 
		ORDER 
			by raw_block_status.block_id;`,
		pq.Array(block_number), block_hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []int64{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	result := []int64{}
	for rows.Next() {
		var block_id int64
		err := rows.Scan(&block_id)
		if err != nil {
			return nil, err
		}
		result = append(result, block_id)
	}

	return result, nil
}

// ////////////////
// RPC interface //
// ////////////////
// get chain id from node
func rpciGetChainId() (uint64, error) {
	chainId, err := elClient.ChainID(context.Background())
	if err != nil {
		return 0, fmt.Errorf("error retrieving chain id from node: %w", err)
	}
	return chainId.Uint64(), nil
}

// get latest block number from node
func rpciGetLatestBlock() (int64, error) {
	latestBlockNumber, err := elClient.BlockNumber(context.Background())
	if err != nil {
		return 0, fmt.Errorf("error retrieving latest block number: %w", err)
	}
	return int64(latestBlockNumber), nil
}

// do all the http stuff
func _rpciGetHttpResult(body []byte, nodeRequestsAtOnce int, count int) ([][]byte, error) {
	startTime := time.Now()
	r, err := http.NewRequest("POST", utils.Config.Eth1RpcEndpoint, bytes.NewBuffer(body))
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
	resByte, err := io.ReadAll(res.Body)
	if err != nil {
		if strings.Compare(err.Error(), errContextDeadlineExceeded.Error()) == 0 {
			err = errContextDeadlineExceeded
			logrus.Warnf("Exceeded after %v, if this happens a lot, you might have increase the http timeout (currently at %d seconds), or reduce the 'node requests at once' count (currently at %d)", time.Second*time.Duration(time.Since(startTime).Seconds()), HTTP_TIMEOUT_IN_SECONDS, nodeRequestsAtOnce)
		}
		return nil, fmt.Errorf("error reading request body: %w", err)
	}

	errorToCheck := []byte(`"error":{"code"`)
	if bytes.Contains(resByte, errorToCheck) {
		const keepDigitsTotal = 1000
		const keepDigitsFront = 100
		if len(resByte) > keepDigitsTotal {
			i := bytes.Index(resByte, errorToCheck)
			if i >= keepDigitsFront {
				resByte = append([]byte(`<...>`), resByte[i-keepDigitsFront:]...)
			}
			if len(resByte) > keepDigitsTotal {
				resByte = append(resByte[:keepDigitsTotal-5], []byte(`<...>`)...)
			}
		}
		return nil, fmt.Errorf("rpc error: %s", resByte)
	}

	return _splitAndVerifyJsonArray(resByte, count)
}

// will fill only receipts_compressed based on block, used by rpciGetBulkRawReceipts function
func _rpciGetBulkRawBlockReceipts(blockRawData []fullBlockRawData, nodeRequestsAtOnce int) error {
	// check
	{
		l := len(blockRawData)
		if l < 1 {
			return fmt.Errorf("empty blockRawData array received")
		}
		if l > nodeRequestsAtOnce {
			return fmt.Errorf("blockRawData array received with more elements (%d) than allowed (%d)", l, nodeRequestsAtOnce)
		}
	}

	// get array
	var rawData [][]byte
	{
		bodyStr := "["
		for i, v := range blockRawData {
			if i != 0 {
				bodyStr += ","
			}
			bodyStr += fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockReceipts","params":["0x%x"],"id":%d}`, v.blockNumber, i)
		}
		bodyStr += "]"
		var err error
		rawData, err = _rpciGetHttpResult([]byte(bodyStr), nodeRequestsAtOnce, len(blockRawData))
		if err != nil {
			return fmt.Errorf("error (_rpciGetBulkRawBlockReceipts) split and verify json array: %w", err)
		}
	}

	// get data
	for i, v := range rawData {
		blockRawData[i].receiptsCompressed = compress(v)
	}

	return nil
}

// will fill only receipts_compressed based on transaction, used by rpciGetBulkRawReceipts function
func _rpciGetBulkRawTransactionReceipts(blockRawData []fullBlockRawData, nodeRequestsAtOnce int) error {
	// check
	{
		l := len(blockRawData)
		if l < 1 {
			return fmt.Errorf("empty blockRawData array received")
		}
		if l > nodeRequestsAtOnce {
			return fmt.Errorf("blockRawData array received with more elements (%d) than allowed (%d)", l, nodeRequestsAtOnce)
		}
	}

	// iterate through array and get data when threshold reached
	var currentElementCount int
	bodyStr := "["
	for i, v := range blockRawData {
		l := len(v.blockTxs)

		// threshold reached, getting data...
		if i != 0 {
			if currentElementCount+l > nodeRequestsAtOnce {
				bodyStr += "]"
				rawData, err := _rpciGetHttpResult([]byte(bodyStr), nodeRequestsAtOnce, currentElementCount)
				if err != nil {
					return fmt.Errorf("error (_rpciGetBulkRawTransactionReceipts) split and verify json array: %w", err)
				}

				for ii, vv := range rawData {
					blockRawData[ii].receiptsCompressed = compress(vv)
				}

				currentElementCount = 0
				bodyStr = "["
			} else {
				bodyStr += ","
			}
		}

		// adding txs of current block
		currentElementCount += l
		for txIndex, txValue := range v.blockTxs {
			if txIndex != 0 {
				bodyStr += ","
			}
			bodyStr += fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getTransactionReceipt","params":["%s"],"id":%d}`, txValue, i)
		}
	}

	// getting data for the rest...
	{
		bodyStr += "]"
		rawData, err := _rpciGetHttpResult([]byte(bodyStr), nodeRequestsAtOnce, currentElementCount)
		if err != nil {
			return fmt.Errorf("error (_rpciGetBulkRawTransactionReceipts) split and verify json array: %w", err)
		}

		for ii, vv := range rawData {
			blockRawData[ii].receiptsCompressed = compress(vv)
		}
	}

	return nil
}

// will fill only block_hash, block_unclesCount, block_compressed & block_txs
func rpciGetBulkBlockRawData(blockRawData []fullBlockRawData, nodeRequestsAtOnce int) error {
	// check
	{
		l := len(blockRawData)
		if l < 1 {
			return fmt.Errorf("empty blockRawData array received")
		}
		if l > nodeRequestsAtOnce {
			return fmt.Errorf("blockRawData array received with more elements (%d) than allowed (%d)", l, nodeRequestsAtOnce)
		}
	}

	// get array
	var rawData [][]byte
	{
		bodyStr := "["
		for i, v := range blockRawData {
			if i != 0 {
				bodyStr += ","
			}
			bodyStr += fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x%x", true],"id":%d}`, v.blockNumber, i)
		}
		bodyStr += "]"
		var err error
		rawData, err = _rpciGetHttpResult([]byte(bodyStr), nodeRequestsAtOnce, len(blockRawData))
		if err != nil {
			return fmt.Errorf("error (rpciGetBulkBlockRawData) split and verify json array: %w", err)
		}
	}

	// get data
	blockParsed := &types.Eth1RpcGetBlockResponse{}
	for i, v := range rawData {
		// block
		{
			blockRawData[i].blockCompressed = compress(v)
			err := json.Unmarshal(v, blockParsed)
			if err != nil {
				return fmt.Errorf("error decoding block '%d' response: %w", blockRawData[i].blockNumber, err)
			}
		}

		// id
		if i != blockParsed.Id {
			return fmt.Errorf("impossible error, i '%d' doesn't match blockParsed.Id '%d'", i, blockParsed.Id)
		}

		// number
		{
			blockParsedResultNumber := int64(binary.BigEndian.Uint64(append(make([]byte, 8-len(blockParsed.Result.Number)), blockParsed.Result.Number...)))
			if blockRawData[i].blockNumber != blockParsedResultNumber {
				logrus.Errorf("blockRawData[i].block_number '%d' doesn't match blockParsed.Result.Number '%d'", blockRawData[i].blockNumber, blockParsedResultNumber)
			}
		}

		// hash
		if blockParsed.Result.Hash == nil {
			return fmt.Errorf("blockParsed.Result.Hash is nil at block '%d'", blockRawData[i].blockNumber)
		}
		blockRawData[i].blockHash = blockParsed.Result.Hash

		// transaction
		if blockParsed.Result.Transactions == nil {
			return fmt.Errorf("blockParsed.Result.Transactions is nil at block '%d'", blockRawData[i].blockNumber)
		}
		blockRawData[i].blockTxs = make([]string, len(blockParsed.Result.Transactions))
		for ii, tx := range blockParsed.Result.Transactions {
			blockRawData[i].blockTxs[ii] = tx.Hash.String()
		}

		// uncle count
		if blockParsed.Result.Uncles != nil {
			blockRawData[i].blockUnclesCount = len(blockParsed.Result.Uncles)
		}
	}

	return nil
}

// will fill only block_hash
func rpciGetBulkBlockRawHash(blockRawData []fullBlockRawData, nodeRequestsAtOnce int) error {
	// check
	{
		l := len(blockRawData)
		if l < 1 {
			return fmt.Errorf("empty blockRawData array received")
		}
		if l > nodeRequestsAtOnce {
			return fmt.Errorf("blockRawData array received with more elements (%d) than allowed (%d)", l, nodeRequestsAtOnce)
		}
	}

	// get array
	var rawData [][]byte
	{
		bodyStr := "["
		for i, v := range blockRawData {
			if i != 0 {
				bodyStr += ","
			}
			bodyStr += fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x%x", true],"id":%d}`, v.blockNumber, i)
		}
		bodyStr += "]"
		var err error
		rawData, err = _rpciGetHttpResult([]byte(bodyStr), nodeRequestsAtOnce, len(blockRawData))
		if err != nil {
			return fmt.Errorf("error (rpciGetBulkBlockRawHash) split and verify json array: %w", err)
		}
	}

	// get data
	blockParsed := &types.Eth1RpcGetBlockResponse{}
	for i, v := range rawData {
		err := json.Unmarshal(v, blockParsed)
		if err != nil {
			return fmt.Errorf("error decoding block '%d' response: %w", blockRawData[i].blockNumber, err)
		}
		if i != blockParsed.Id {
			return fmt.Errorf("impossible error, i '%d' doesn't match blockParsed.Id '%d'", i, blockParsed.Id)
		}
		{
			blockParsedResultNumber := int64(binary.BigEndian.Uint64(append(make([]byte, 8-len(blockParsed.Result.Number)), blockParsed.Result.Number...)))
			if blockRawData[i].blockNumber != blockParsedResultNumber {
				logrus.Errorf("blockRawData[i].block_number '%d' doesn't match blockParsed.Result.Number '%d'", blockRawData[i].blockNumber, blockParsedResultNumber)
			}
		}
		if blockParsed.Result.Hash == nil {
			return fmt.Errorf("blockParsed.Result.Hash is nil at block '%d'", blockRawData[i].blockNumber)
		}
		blockRawData[i].blockHash = blockParsed.Result.Hash
	}

	return nil
}

// will fill only uncles (if available)
func rpciGetBulkRawUncles(blockRawData []fullBlockRawData, nodeRequestsAtOnce int) error {
	// check
	{
		l := len(blockRawData)
		if l < 1 {
			return fmt.Errorf("empty blockRawData array received")
		}
		if l > nodeRequestsAtOnce {
			// I know, in the case of uncles, it's very unlikly that we need all slots, but handling this separate, would be way to much, so whatever
			return fmt.Errorf("blockRawData array received with more elements (%d) than allowed (%d)", l, nodeRequestsAtOnce)
		}
	}

	// get array
	var rawData [][]byte
	{
		requestedCount := 0
		firstElement := true
		bodyStr := "["
		for _, v := range blockRawData {
			if v.blockUnclesCount > 2 || v.blockUnclesCount < 0 {
				// fatal, as this is an impossible error
				utils.LogFatal(nil, "impossible error, found impossible uncle count, expected 0, 1 or 2", 0, map[string]interface{}{"block_unclesCount": v.blockUnclesCount, "block_number": v.blockNumber})
			} else if v.blockUnclesCount == 0 {
				continue
			} else {
				if firstElement {
					firstElement = false
				} else {
					bodyStr += ","
				}
				if v.blockUnclesCount == 1 {
					requestedCount++
					bodyStr += fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getUncleByBlockNumberAndIndex","params":["0x%x", "0x0"],"id":%d}`, v.blockNumber, v.blockNumber)
				} else /* if v.block_unclesCount == 2 */ {
					requestedCount++
					bodyStr += fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getUncleByBlockNumberAndIndex","params":["0x%x", "0x0"],"id":%d},`, v.blockNumber, v.blockNumber)
					requestedCount++
					bodyStr += fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getUncleByBlockNumberAndIndex","params":["0x%x", "0x1"],"id":%d}`, v.blockNumber, v.blockNumber)
				}
			}
		}
		bodyStr += "]"
		if requestedCount == 0 { // nothing todo, no uncles in set
			return nil
		}

		var err error
		rawData, err = _rpciGetHttpResult([]byte(bodyStr), nodeRequestsAtOnce, requestedCount)
		if err != nil {
			return fmt.Errorf("error (rpciGetBulkRawUncles) split and verify json array: %w", err)
		}
	}

	// get data
	rdIndex := 0
	for i, v := range blockRawData {
		if v.blockUnclesCount > 0 { // Not the prettiest way, but the unmarshal would take much longer with the same result
			blockRawData[i].unclesCompressed = compress(rawData[rdIndex])
			rdIndex++
		}
	}

	return nil
}

// will fill only receipts_compressed
func rpciGetBulkRawReceipts(blockRawData []fullBlockRawData, nodeRequestsAtOnce int) error {
	if utils.Config.Chain.Id == ARBITRUM_CHAINID {
		return _rpciGetBulkRawTransactionReceipts(blockRawData, nodeRequestsAtOnce)
	}
	return _rpciGetBulkRawBlockReceipts(blockRawData, nodeRequestsAtOnce)
}

// will fill only traces_compressed
func rpciGetBulkRawTraces(blockRawData []fullBlockRawData, nodeRequestsAtOnce int) error {
	// check
	{
		l := len(blockRawData)
		if l < 1 {
			return fmt.Errorf("empty blockRawData array received")
		}
		if l > nodeRequestsAtOnce {
			return fmt.Errorf("blockRawData array received with more elements (%d) than allowed (%d)", l, nodeRequestsAtOnce)
		}
	}

	// get array
	var rawData [][]byte
	{
		bodyStr := "["
		for i, v := range blockRawData {
			if i != 0 {
				bodyStr += ","
			}
			if utils.Config.Chain.Id == ARBITRUM_CHAINID && v.blockNumber <= ARBITRUM_NITRO_BLOCKNUMBER {
				bodyStr += fmt.Sprintf(`{"jsonrpc":"2.0","method":"arbtrace_block","params":["0x%x"],"id":%d}`, v.blockNumber, i)
			} else {
				bodyStr += fmt.Sprintf(`{"jsonrpc":"2.0","method":"debug_traceBlockByNumber","params":["0x%x", {"tracer": "callTracer"}],"id":%d}`, v.blockNumber, i)
			}
		}
		bodyStr += "]"
		var err error
		rawData, err = _rpciGetHttpResult([]byte(bodyStr), nodeRequestsAtOnce, len(blockRawData))
		if err != nil {
			return fmt.Errorf("error (rpciGetBulkRawTraces) split and verify json array: %w", err)
		}
	}

	// get data
	for i, v := range rawData {
		blockRawData[i].tracesCompressed = compress(v)
	}

	return nil
}
