package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/erc20"
	"eth2-exporter/rpc"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"flag"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/coocood/freecache"
	"github.com/ethereum/go-ethereum/common"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	_ "net/http/pprof"
)

func main() {
	erigonEndpoint := flag.String("erigon", "", "Erigon archive node enpoint")
	block := flag.Int64("block", 0, "Index a specific block")

	reorgDepth := flag.Int("reorg.depth", 20, "Lookback to check and handle chain reorgs")

	concurrencyBlocks := flag.Int64("blocks.concurrency", 30, "Concurrency to use when indexing blocks from erigon")
	startBlocks := flag.Int64("blocks.start", 0, "Block to start indexing")
	endBlocks := flag.Int64("blocks.end", 0, "Block to finish indexing")
	bulkBlocks := flag.Int64("blocks.bulk", 8000, "Maximum number of blocks to be processed before saving")
	offsetBlocks := flag.Int64("blocks.offset", 100, "Blocks offset")
	checkBlocksGaps := flag.Bool("blocks.gaps", false, "Check for gaps in the blocks table")
	checkBlocksGapsLookback := flag.Int("blocks.gaps.lookback", 1000000, "Lookback for gaps check of the blocks table")
	traceMode := flag.String("blocks.tracemode", "parity/geth", "Trace mode to use, can bei either 'parity', 'geth' or 'parity/geth' for both")

	concurrencyData := flag.Int64("data.concurrency", 30, "Concurrency to use when indexing data from bigtable")
	startData := flag.Int64("data.start", 0, "Block to start indexing")
	endData := flag.Int64("data.end", 0, "Block to finish indexing")
	bulkData := flag.Int64("data.bulk", 8000, "Maximum number of blocks to be processed before saving")
	offsetData := flag.Int64("data.offset", 1000, "Data offset")
	checkDataGaps := flag.Bool("data.gaps", false, "Check for gaps in the data table")
	checkDataGapsLookback := flag.Int("data.gaps.lookback", 1000000, "Lookback for gaps check of the blocks table")

	enableBalanceUpdater := flag.Bool("balances.enabled", false, "Enable balance update process")
	enableFullBalanceUpdater := flag.Bool("balances.full.enabled", false, "Enable full balance update process")
	balanceUpdaterBatchSize := flag.Int("balances.batch", 1000, "Batch size for balance updates")

	tokenPriceExport := flag.Bool("token.price.enabled", false, "Enable token export process")
	tokenPriceExportList := flag.String("token.price.list", "", "Tokenlist path to use for the token price export")
	tokenPriceExportFrequency := flag.Duration("token.price.frequency", time.Hour, "Token price export interval")

	versionFlag := flag.Bool("version", false, "Print version and exit")

	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")

	enableEnsUpdater := flag.Bool("ens.enabled", false, "Enable ens update process")

	flag.Parse()

	if *versionFlag {
		fmt.Println(version.Version)
		fmt.Println(version.GoVersion)
		return
	}

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg
	logrus.WithField("config", *configPath).WithField("version", version.Version).WithField("chainName", utils.Config.Chain.ClConfig.ConfigName).Printf("starting")

	// enable pprof endpoint if requested
	if utils.Config.Pprof.Enabled {
		go func() {
			logrus.Infof("starting pprof http server on port %s", utils.Config.Pprof.Port)
			logrus.Info(http.ListenAndServe(fmt.Sprintf("localhost:%s", utils.Config.Pprof.Port), nil))
		}()
	}

	db.MustInitDB(&types.DatabaseConfig{
		Username:     cfg.WriterDatabase.Username,
		Password:     cfg.WriterDatabase.Password,
		Name:         cfg.WriterDatabase.Name,
		Host:         cfg.WriterDatabase.Host,
		Port:         cfg.WriterDatabase.Port,
		MaxOpenConns: cfg.WriterDatabase.MaxOpenConns,
		MaxIdleConns: cfg.WriterDatabase.MaxIdleConns,
	}, &types.DatabaseConfig{
		Username:     cfg.ReaderDatabase.Username,
		Password:     cfg.ReaderDatabase.Password,
		Name:         cfg.ReaderDatabase.Name,
		Host:         cfg.ReaderDatabase.Host,
		Port:         cfg.ReaderDatabase.Port,
		MaxOpenConns: cfg.ReaderDatabase.MaxOpenConns,
		MaxIdleConns: cfg.ReaderDatabase.MaxIdleConns,
	})
	defer db.ReaderDb.Close()
	defer db.WriterDb.Close()

	if erigonEndpoint == nil || *erigonEndpoint == "" {

		if utils.Config.Eth1ErigonEndpoint == "" {

			utils.LogFatal(nil, "no erigon node url provided", 0)
		} else {
			logrus.Info("applying erigon endpoint from config")
			*erigonEndpoint = utils.Config.Eth1ErigonEndpoint
		}

	}

	logrus.Infof("using erigon node at %v", *erigonEndpoint)
	client, err := rpc.NewErigonClient(*erigonEndpoint)
	if err != nil {
		utils.LogFatal(err, "erigon client creation error", 0)
	}

	chainId := strconv.FormatUint(utils.Config.Chain.ClConfig.DepositChainID, 10)

	balanceUpdaterPrefix := chainId + ":B:"

	nodeChainId, err := client.GetNativeClient().ChainID(context.Background())
	if err != nil {
		utils.LogFatal(err, "node chain id error", 0)
	}

	if nodeChainId.String() != chainId {
		logrus.Fatalf("node chain id mismatch, wanted %v got %v", chainId, nodeChainId.String())
	}

	bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, chainId, utils.Config.RedisCacheEndpoint)
	if err != nil {
		logrus.Fatalf("error connecting to bigtable: %v", err)
	}
	defer bt.Close()

	if *tokenPriceExport {
		go func() {
			for {
				err = UpdateTokenPrices(bt, client, *tokenPriceExportList)
				if err != nil {
					utils.LogError(err, "error while updating token prices", 0)
					time.Sleep(*tokenPriceExportFrequency)
				}
				time.Sleep(*tokenPriceExportFrequency)
			}
		}()
	}
	// err = UpdateTokenPrices(bt, client, "tokenlists/tokens.uniswap.org.json")
	// if err != nil {
	// 	logrus.Fatal(err)
	// }
	// return
	if *enableFullBalanceUpdater {
		ProcessMetadataUpdates(bt, client, balanceUpdaterPrefix, *balanceUpdaterBatchSize, -1)
		return
		// currentKey := balanceUpdaterPrefix // "1:00028ebf7d36c5779c1deddf3ba72761fd46c8aa"
		// for {
		// 	keys, pairs, err := bt.GetMetadata(currentKey, *balanceUpdaterBatchSize)
		// 	if err != nil {
		// 		logrus.Fatal(err)
		// 	}

		// 	if len(keys) == 0 {
		// 		logrus.Infof("done")
		// 		return
		// 	}
		// 	// for _, pair := range pairs {
		// 	// 	logrus.Info(pair)
		// 	// }

		// 	logrus.Infof("currently at %v, processing balances for %v pairs", currentKey, len(pairs))
		// 	balances, err := client.GetBalances(pairs, 1, 4)
		// 	if err != nil {
		// 		logrus.Fatal(err)
		// 	}
		// 	// for _, balance := range balances {
		// 	// 	logrus.Infof("%x %x %s", balance.Address, balance.Token, new(big.Int).SetBytes(balance.Balance))
		// 	// }

		// 	err = bt.SaveBalances(balances, []string{})
		// 	if err != nil {
		// 		logrus.Fatal(err)
		// 	}
		// 	currentKey = keys[len(keys)-1]
		// }
	}

	transforms := make([]func(blk *types.Eth1Block, cache *freecache.Cache) (*types.BulkMutations, *types.BulkMutations, error), 0)
	transforms = append(transforms,
		bt.TransformBlock,
		bt.TransformTx,
		bt.TransformItx,
		bt.TransformBlobTx,
		bt.TransformERC20,
		bt.TransformERC721,
		bt.TransformERC1155,
		bt.TransformUncle,
		bt.TransformWithdrawals,
		bt.TransformEnsNameRegistered)

	cache := freecache.NewCache(100 * 1024 * 1024) // 100 MB limit

	if *block != 0 {
		err = IndexFromNode(bt, client, *block, *block, *concurrencyBlocks, *traceMode)
		if err != nil {
			logrus.WithError(err).Fatalf("error indexing from node, start: %v end: %v concurrency: %v", *block, *block, *concurrencyBlocks)
		}
		err = bt.IndexEventsWithTransformers(*block, *block, transforms, *concurrencyData, cache)
		if err != nil {
			logrus.WithError(err).Fatalf("error indexing from bigtable")
		}
		cache.Clear()

		logrus.Infof("indexing of block %v completed", *block)
		return
	}

	if *checkBlocksGaps {
		bt.CheckForGapsInBlocksTable(*checkBlocksGapsLookback)
		return
	}

	if *checkDataGaps {
		bt.CheckForGapsInDataTable(*checkDataGapsLookback)
		return
	}

	if *endBlocks != 0 && *startBlocks < *endBlocks {
		err = IndexFromNode(bt, client, *startBlocks, *endBlocks, *concurrencyBlocks, *traceMode)
		if err != nil {
			logrus.WithError(err).Fatalf("error indexing from node, start: %v end: %v concurrency: %v", *startBlocks, *endBlocks, *concurrencyBlocks)
		}
		return
	}

	if *endData != 0 && *startData < *endData {
		err = bt.IndexEventsWithTransformers(int64(*startData), int64(*endData), transforms, *concurrencyData, cache)
		if err != nil {
			logrus.WithError(err).Fatalf("error indexing from bigtable")
		}
		cache.Clear()
		return
	}

	lastSuccessulBlockIndexingTs := time.Now()
	for ; ; time.Sleep(time.Second * 14) {
		err := HandleChainReorgs(bt, client, *reorgDepth)
		if err != nil {
			logrus.Errorf("error handling chain reorgs: %v", err)
			continue
		}

		lastBlockFromNode, err := client.GetLatestEth1BlockNumber()
		if err != nil {
			logrus.Errorf("error retrieving latest eth block number: %v", err)
			continue
		}

		lastBlockFromBlocksTable, err := bt.GetLastBlockInBlocksTable()
		if err != nil {
			logrus.Errorf("error retrieving last blocks from blocks table: %v", err)
			continue
		}

		lastBlockFromDataTable, err := bt.GetLastBlockInDataTable()
		if err != nil {
			logrus.Errorf("error retrieving last blocks from data table: %v", err)
			continue
		}

		logrus.WithFields(
			logrus.Fields{
				"node":   lastBlockFromNode,
				"blocks": lastBlockFromBlocksTable,
				"data":   lastBlockFromDataTable,
			},
		).Infof("last blocks")

		continueAfterError := false
		if lastBlockFromNode > 0 {
			if lastBlockFromBlocksTable < int(lastBlockFromNode) {
				logrus.Infof("missing blocks %v to %v in blocks table, indexing ...", lastBlockFromBlocksTable, lastBlockFromNode)

				startBlock := int64(lastBlockFromBlocksTable) - *offsetBlocks
				if startBlock < 0 {
					startBlock = 0
				}

				if *bulkBlocks <= 0 || *bulkBlocks > int64(lastBlockFromNode)-startBlock+1 {
					*bulkBlocks = int64(lastBlockFromNode) - startBlock + 1
				}

				for startBlock <= int64(lastBlockFromNode) && !continueAfterError {
					endBlock := startBlock + *bulkBlocks - 1
					if endBlock > int64(lastBlockFromNode) {
						endBlock = int64(lastBlockFromNode)
					}

					err = IndexFromNode(bt, client, startBlock, endBlock, *concurrencyBlocks, *traceMode)
					if err != nil {
						errMsg := "error indexing from node"
						errFields := map[string]interface{}{
							"start":       startBlock,
							"end":         endBlock,
							"concurrency": *concurrencyBlocks}
						if time.Since(lastSuccessulBlockIndexingTs) > time.Minute*30 {
							utils.LogFatal(err, errMsg, 0, errFields)
						} else {
							utils.LogError(err, errMsg, 0, errFields)
						}
						continueAfterError = true
						continue
					} else {
						lastSuccessulBlockIndexingTs = time.Now()
					}

					startBlock = endBlock + 1
				}
				if continueAfterError {
					continue
				}
			}

			if lastBlockFromDataTable < int(lastBlockFromNode) {
				logrus.Infof("missing blocks %v to %v in data table, indexing ...", lastBlockFromDataTable, lastBlockFromNode)

				startBlock := int64(lastBlockFromDataTable) - *offsetData
				if startBlock < 0 {
					startBlock = 0
				}

				if *bulkData <= 0 || *bulkData > int64(lastBlockFromNode)-startBlock+1 {
					*bulkData = int64(lastBlockFromNode) - startBlock + 1
				}

				for startBlock <= int64(lastBlockFromNode) && !continueAfterError {
					endBlock := startBlock + *bulkData - 1
					if endBlock > int64(lastBlockFromNode) {
						endBlock = int64(lastBlockFromNode)
					}

					err = bt.IndexEventsWithTransformers(startBlock, endBlock, transforms, *concurrencyData, cache)
					if err != nil {
						utils.LogError(err, "error indexing from bigtable", 0, map[string]interface{}{"start": startBlock, "end": endBlock, "concurrency": *concurrencyData})
						cache.Clear()
						continueAfterError = true
						continue
					}
					cache.Clear()

					startBlock = endBlock + 1
				}
				if continueAfterError {
					continue
				}
			}
		}

		if *enableBalanceUpdater {
			ProcessMetadataUpdates(bt, client, balanceUpdaterPrefix, *balanceUpdaterBatchSize, 10)
		}

		if *enableEnsUpdater {
			err := bt.ImportEnsUpdates(client.GetNativeClient(), 1000)
			if err != nil {
				utils.LogError(err, "error importing ens updates", 0, nil)
				continue
			}
		}

		logrus.Infof("index run completed")
		services.ReportStatus("eth1indexer", "Running", nil)
	}

	// utils.WaitForCtrlC()
}

func UpdateTokenPrices(bt *db.Bigtable, client *rpc.ErigonClient, tokenListPath string) error {

	tokenListContent, err := os.ReadFile(tokenListPath)
	if err != nil {
		return err
	}

	tokenList := &erc20.ERC20TokenList{}

	err = json.Unmarshal(tokenListContent, tokenList)
	if err != nil {
		return err
	}

	type defillamaPriceRequest struct {
		Coins []string `json:"coins"`
	}
	coinsList := make([]string, 0, len(tokenList.Tokens))
	for _, token := range tokenList.Tokens {
		coinsList = append(coinsList, "ethereum:"+token.Address)
	}

	req := &defillamaPriceRequest{
		Coins: coinsList,
	}

	reqEncoded, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpClient := &http.Client{Timeout: time.Second * 10}

	resp, err := httpClient.Post("https://coins.llama.fi/prices", "application/json", bytes.NewReader(reqEncoded))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error querying defillama api: %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	type defillamaCoin struct {
		Decimals  int64            `json:"decimals"`
		Price     *decimal.Decimal `json:"price"`
		Symbol    string           `json:"symbol"`
		Timestamp int64            `json:"timestamp"`
	}

	type defillamaResponse struct {
		Coins map[string]defillamaCoin `json:"coins"`
	}

	respParsed := &defillamaResponse{}
	err = json.Unmarshal(body, respParsed)
	if err != nil {
		return err
	}

	tokenPrices := make([]*types.ERC20TokenPrice, 0, len(respParsed.Coins))
	for address, data := range respParsed.Coins {
		tokenPrices = append(tokenPrices, &types.ERC20TokenPrice{
			Token: common.FromHex(strings.TrimPrefix(address, "ethereum:0x")),
			Price: []byte(data.Price.String()),
		})
	}

	g := new(errgroup.Group)
	g.SetLimit(20)
	for i := range tokenPrices {
		i := i
		g.Go(func() error {

			metadata, err := client.GetERC20TokenMetadata(tokenPrices[i].Token)
			if err != nil {
				return err
			}
			tokenPrices[i].TotalSupply = metadata.TotalSupply
			// logrus.Infof("price for token %x is %s @ %v", tokenPrices[i].Token, tokenPrices[i].Price, new(big.Int).SetBytes(tokenPrices[i].TotalSupply))
			return nil
		})
	}
	err = g.Wait()
	if err != nil {
		return err
	}

	return bt.SaveERC20TokenPrices(tokenPrices)
}

func HandleChainReorgs(bt *db.Bigtable, client *rpc.ErigonClient, depth int) error {
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

			// first we set the cached marker of the last block in the blocks/data table to the block prior to the forked one
			if i > 0 {
				previousBlock := i - 1
				err := bt.SetLastBlockInBlocksTable(int64(previousBlock))
				if err != nil {
					return fmt.Errorf("error setting last block [%v] in blocks table: %w", previousBlock, err)
				}
				err = bt.SetLastBlockInDataTable(int64(previousBlock))
				if err != nil {
					return fmt.Errorf("error setting last block [%v] in data table: %w", previousBlock, err)
				}
				// now we can proceed to delete all blocks including and after the forked block
			}
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

func ProcessMetadataUpdates(bt *db.Bigtable, client *rpc.ErigonClient, prefix string, batchSize int, iterations int) {
	lastKey := prefix
	// for {
	// 	updates, err := bt.GetMetadataUpdates(lastKey, batchSize)
	// 	if err != nil {
	// 		logrus.Fatal(err)
	// 	}

	// 	currentAddress := ""
	// 	tokens := make([]string, 0, 100)
	// 	pairs := make([]string, 0, batchSize)
	// 	for _, update := range updates {
	// 		s := strings.Split(update, ":")

	// 		if len(s) != 3 {
	// 			logrus.Fatalf("%v has an invalid format", update)
	// 		}

	// 		if s[0] != "B" {
	// 			logrus.Fatalf("%v has invalid balance update prefix", update)
	// 		}

	// 		address := s[1]
	// 		token := s[2]
	// 		pairs = append(pairs, update)

	// 		if currentAddress == "" {
	// 			currentAddress = address
	// 		} else if address != currentAddress {
	// 			logrus.Infof("retrieving %v token balances for address %v", len(tokens), currentAddress)
	// 			start := time.Now()
	// 			balances, err := client.GetBalancesForAddresse(currentAddress, tokens)

	// 			if err != nil {
	// 				logrus.Errorf("error during balance checker contract call: %v", err)
	// 				logrus.Infof("retrieving balances via batch rpc calls")
	// 				balances, err = client.GetBalances(pairs)
	// 				if err != nil {
	// 					logrus.Fatal(err)
	// 				}
	// 			}

	// 			logrus.Infof("retrieved %v balances in %v", len(balances), time.Since(start))
	// 			// for i, t := range tokens {
	// 			// 	if len(balances[i]) > 0 {
	// 			// 		logrus.Infof("balance of address %v of token %v is %x", currentAddress, t, balances[i])
	// 			// 	}
	// 			// }
	// 			currentAddress = address
	// 			tokens = make([]string, 0, 100)
	// 			pairs = make([]string, 0, 1000)
	// 		}

	// 		tokens = append(tokens, token)
	// 	}
	// 	logrus.Infof("retrieving %v token balances for address %v", len(tokens), currentAddress)
	// 	start := time.Now()
	// 	balances, err := client.GetBalancesForAddresse(currentAddress, tokens)

	// 	if err != nil {
	// 		logrus.Errorf("error during balance checker contract call: %v", err)
	// 		logrus.Infof("retrieving balances via batch rpc calls")
	// 		balances, err = client.GetBalances(pairs)
	// 		if err != nil {
	// 			logrus.Fatal(err)
	// 		}
	// 	}

	// 	logrus.Infof("retrieved %v balances in %v", len(balances), time.Since(start))
	// 	// for i, t := range tokens {
	// 	// 	if len(balances[i]) > 0 {
	// 	// 		logrus.Infof("balance of address %v of token %v is %x", currentAddress, t, balances[i])
	// 	// 	}
	// 	// }
	// 	lastKey = updates[len(updates)-1]
	// }

	its := 0
	for {
		start := time.Now()
		keys, pairs, err := bt.GetMetadataUpdates(prefix, lastKey, batchSize)
		if err != nil {
			logrus.Errorf("error retrieving metadata updates from bigtable: %v", err)
			return
		}

		if len(keys) == 0 {
			return
		}

		// for _, b := range balances {
		// 	logrus.Infof("retrieved balance %x for token %x of address %x", b.Balance, b.Token, b.Address)
		// }

		balances := make([]*types.Eth1AddressBalance, 0, len(pairs))
		for b := 0; b < len(pairs); b += batchSize {
			start := b
			end := b + batchSize
			if len(pairs) < end {
				end = len(pairs)
			}

			logrus.Infof("processing batch %v with start %v and end %v", b, start, end)

			b, err := client.GetBalances(pairs[start:end], 2, 4)

			if err != nil {
				logrus.Errorf("error retrieving balances from node: %v", err)
				return
			}
			balances = append(balances, b...)
		}

		err = bt.SaveBalances(balances, keys)
		if err != nil {
			logrus.Errorf("error saving balances to bigtable: %v", err)
			return
		}
		// for i, b := range balances {

		// 	if len(b) > 0 {
		// 		logrus.Infof("balance for key %v is %x", updates[i], b)
		// 	}
		// }

		lastKey = keys[len(keys)-1]
		logrus.Infof("retrieved %v balances in %v, currently at %v", len(balances), time.Since(start), lastKey)

		its++

		if iterations != -1 && its > iterations {
			return
		}
	}
	// g := new(errgroup.Group)
	// g.SetLimit(batchSize)

	// for _, update := range updates {
	// 	update := update

	// 	g.Go(func() error {
	// 		// logrus.Infof("updating balance of key %v", update)
	// 		s := strings.Split(update, ":")

	// 		if len(s) != 3 {
	// 			logrus.Fatalf("%v has an invalid format", update)
	// 		}

	// 		if s[0] != "B" {
	// 			logrus.Fatalf("%v has invalid balance update prefix", update)
	// 		}

	// 		address := s[1]
	// 		token := s[2]

	// 		if token == "00" {
	// 			balance, err := client.GetNativeBalance(address)
	// 			if err != nil {
	// 				logrus.Fatal(err)
	// 			}

	// 			balanceInt := new(big.Int).SetBytes(balance)

	// 			if balanceInt.Cmp(big.NewInt(0)) != 0 {
	// 				logrus.Infof("native balance of %v is %x", address, balanceInt.String())
	// 			}
	// 		} else {
	// 			balance, err := client.GetERC20TokenBalance(address, token)
	// 			if err != nil {
	// 				logrus.Fatal(err)
	// 			}

	// 			balanceInt := new(big.Int).SetBytes(balance)
	// 			if balanceInt.Cmp(big.NewInt(0)) != 0 {
	// 				logrus.Infof("token %v balance of %v is %v", token, address, balanceInt.String())
	// 			}
	// 		}
	// 		return nil
	// 	})
	// }

	// err = g.Wait()

	// if err != nil {
	// 	logrus.Fatal(err)
	// }
}

func IndexFromNode(bt *db.Bigtable, client *rpc.ErigonClient, start, end, concurrency int64, traceMode string) error {
	ctx := context.Background()
	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(int(concurrency))

	startTs := time.Now()
	lastTickTs := time.Now()

	processedBlocks := int64(0)

	for i := start; i <= end; i++ {

		i := i
		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
			}

			blockStartTs := time.Now()
			bc, timings, err := client.GetBlock(i, traceMode)
			if err != nil {
				return fmt.Errorf("error getting block: %v from ethereum node err: %w", i, err)
			}

			dbStart := time.Now()
			err = bt.SaveBlock(bc)
			if err != nil {
				return fmt.Errorf("error saving block: %v to bigtable: %w", i, err)

			}
			current := atomic.AddInt64(&processedBlocks, 1)
			if current%100 == 0 {
				r := end - start
				if r == 0 {
					r = 1
				}
				perc := float64(i-start) * 100 / float64(r)

				logrus.Infof("retrieved & saved block %v (0x%x) in %v (header: %v, receipts: %v, traces: %v, db: %v)", bc.Number, bc.Hash, time.Since(blockStartTs), timings.Headers, timings.Receipts, timings.Traces, time.Since(dbStart))
				logrus.Infof("processed %v blocks in %v (%.1f blocks / sec); sync is %.1f%% complete", current, time.Since(startTs), float64((current))/time.Since(lastTickTs).Seconds(), perc)

				lastTickTs = time.Now()
				atomic.StoreInt64(&processedBlocks, 0)
			}
			return nil
		})

	}

	err := g.Wait()

	if err != nil {
		return err
	}

	lastBlockInCache, err := bt.GetLastBlockInBlocksTable()
	if err != nil {
		return err
	}

	if end > int64(lastBlockInCache) {
		err := bt.SetLastBlockInBlocksTable(end)

		if err != nil {
			return err
		}
	}
	return nil
}

func ImportMainnetERC20TokenMetadataFromTokenDirectory(bt *db.Bigtable) {

	client := &http.Client{Timeout: time.Second * 10}

	resp, err := client.Get("<INSERT_TOKENLIST_URL>")

	if err != nil {
		utils.LogFatal(err, "getting client error", 0)
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		utils.LogFatal(err, "reading body for ERC20 tokens error", 0)
	}

	type TokenDirectory struct {
		ChainID       int64    `json:"chainId"`
		Keywords      []string `json:"keywords"`
		LogoURI       string   `json:"logoURI"`
		Name          string   `json:"name"`
		Timestamp     string   `json:"timestamp"`
		TokenStandard string   `json:"tokenStandard"`
		Tokens        []struct {
			Address    string `json:"address"`
			ChainID    int64  `json:"chainId"`
			Decimals   int64  `json:"decimals"`
			Extensions struct {
				Description   string      `json:"description"`
				Link          string      `json:"link"`
				OgImage       interface{} `json:"ogImage"`
				OriginAddress string      `json:"originAddress"`
				OriginChainID int64       `json:"originChainId"`
			} `json:"extensions"`
			LogoURI string `json:"logoURI"`
			Name    string `json:"name"`
			Symbol  string `json:"symbol"`
		} `json:"tokens"`
	}

	td := &TokenDirectory{}

	err = json.Unmarshal(body, td)

	if err != nil {
		utils.LogFatal(err, "unmarshal json body error", 0)
	}

	for _, token := range td.Tokens {

		address, err := hex.DecodeString(strings.TrimPrefix(token.Address, "0x"))
		if err != nil {
			utils.LogFatal(err, "decoding string to hex error", 0)
		}
		logrus.Infof("processing token %v at address %x", token.Name, address)

		meta := &types.ERC20Metadata{}
		meta.Decimals = big.NewInt(token.Decimals).Bytes()
		meta.Description = token.Extensions.Description
		if len(token.LogoURI) > 0 {
			resp, err := client.Get(token.LogoURI)

			if err == nil && resp.StatusCode == 200 {
				body, err := io.ReadAll(resp.Body)

				if err != nil {
					utils.LogFatal(err, "reading body for ERC20 token logo URI error", 0)
				}

				meta.Logo = body
				meta.LogoFormat = token.LogoURI
			}
		}
		meta.Name = token.Name
		meta.OfficialSite = token.Extensions.Link
		meta.Symbol = token.Symbol

		err = bt.SaveERC20Metadata(address, meta)
		if err != nil {
			utils.LogFatal(err, "error while saving ERC20 metadata", 0)
		}
		time.Sleep(time.Millisecond * 250)
	}

}

func ImportNameLabels(bt *db.Bigtable) {
	type NameEntry struct {
		Name string
	}

	res := make(map[string]*NameEntry)

	data, err := os.ReadFile("")

	if err != nil {
		utils.LogFatal(err, "reading file error", 0)
	}

	err = json.Unmarshal(data, &res)

	if err != nil {
		utils.LogFatal(err, "unmarshal json error", 0)
	}

	logrus.Infof("retrieved %v names", len(res))

	for address, name := range res {
		if name.Name == "" {
			continue
		}
		logrus.Infof("%v: %v", address, name.Name)
		bt.SaveAddressName(common.FromHex(strings.TrimPrefix(address, "0x")), name.Name)
	}
}
