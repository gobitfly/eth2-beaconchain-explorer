package main

import (
	"encoding/hex"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/rpc"
	"eth2-exporter/types"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/karlseguin/ccache/v2"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func main() {
	// localhost:8545
	erigonEndpoint := flag.String("erigon", "", "Erigon archive node enpoint")

	block := flag.Int64("block", 0, "Index a specific block")

	concurrencyBlocks := flag.Int64("blocks.concurrency", 30, "Concurrency to use when indexing blocks from erigon")
	startBlocks := flag.Int64("blocks.start", 0, "Block to start indexing")
	endBlocks := flag.Int64("blocks.end", 0, "Block to finish indexing")
	offsetBlocks := flag.Int64("blocks.offset", 100, "Blocks offset")
	checkBlocksGaps := flag.Bool("blocks.gaps", false, "Check for gaps in the blocks table")
	checkBlocksGapsLookback := flag.Int("blocks.gaps.lookback", 1000000, "Lookback for gaps check of the blocks table")

	concurrencyData := flag.Int64("data.concurrency", 30, "Concurrency to use when indexing data from bigtable")
	startData := flag.Int64("data.start", 0, "Block to start indexing")
	endData := flag.Int64("data.end", 0, "Block to finish indexing")
	offsetData := flag.Int64("data.offset", 1000, "Data offset")
	checkDataGaps := flag.Bool("data.gaps", false, "Check for gaps in the data table")
	checkDataGapsLookback := flag.Int("data.gaps.lookback", 1000000, "Lookback for gaps check of the blocks table")

	enableBalanceUpdater := flag.Bool("balances.enabled", false, "Enable balance update process")
	balanceUpdaterPrefix := flag.String("balances.prefix", "", "Prefix to use for fetching balance updates")
	balanceUpdaterBatchSize := flag.Int("balances.batch", 1000, "Batch size for balance updates")

	flag.Parse()

	if erigonEndpoint == nil || *erigonEndpoint == "" {
		logrus.Fatal("no erigon node url provided")
	}

	logrus.Infof("using erigon node at %v", *erigonEndpoint)
	client, err := rpc.NewErigonClient(*erigonEndpoint)
	if err != nil {
		logrus.Fatal(err)
	}

	bt, err := db.NewBigtable("etherchain", "etherchain", "1")
	if err != nil {
		logrus.Fatalf("error connecting to bigtable: %v", err)
	}
	defer bt.Close()

	if *enableBalanceUpdater {
		ProcessMetadataUpdates(bt, client, *balanceUpdaterPrefix, *balanceUpdaterBatchSize)
		return
	}

	transforms := make([]func(blk *types.Eth1Block, cache *ccache.Cache) (*types.BulkMutations, *types.BulkMutations, error), 0)
	transforms = append(transforms, bt.TransformBlock, bt.TransformTx, bt.TransformItx, bt.TransformERC20, bt.TransformERC721, bt.TransformERC1155, bt.TransformUncle)

	if *block != 0 {
		err = IndexFromNode(bt, client, *block, *block, *concurrencyBlocks)
		if err != nil {
			logrus.WithError(err).Fatalf("error indexing from node")
		}
		err = IndexFromBigtable(bt, *block, *block, transforms, *concurrencyData)
		if err != nil {
			logrus.WithError(err).Fatalf("error indexing from bigtable")
		}

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
		err = IndexFromNode(bt, client, *startBlocks, *endBlocks, *concurrencyBlocks)
		if err != nil {
			logrus.WithError(err).Fatalf("error indexing from node")
		}
		return
	}

	if *endData != 0 && *startData < *endData {
		err = IndexFromBigtable(bt, int64(*startData), int64(*endData), transforms, *concurrencyData)
		if err != nil {
			logrus.WithError(err).Fatalf("error indexing from bigtable")
		}
		return
	}

	// return
	// bt.DeleteRowsWithPrefix("1:b:")
	// return

	for {
		lastBlockFromNode, err := client.GetLatestEth1BlockNumber()
		if err != nil {
			logrus.Fatal(err)
		}
		lastBlockFromNode = lastBlockFromNode - 100

		lastBlockFromBlocksTable, err := bt.GetLastBlockInBlocksTable()
		if err != nil {
			logrus.Fatal(err)
		}

		lastBlockFromDataTable, err := bt.GetLastBlockInDataTable()
		if err != nil {
			logrus.Fatal(err)
		}

		logrus.WithFields(
			logrus.Fields{
				"node":   lastBlockFromNode,
				"blocks": lastBlockFromBlocksTable,
				"data":   lastBlockFromDataTable,
			},
		).Infof("last blocks")

		if lastBlockFromBlocksTable < int(lastBlockFromNode) {
			logrus.Infof("missing blocks %v to %v in blocks table, indexing ...", lastBlockFromBlocksTable, lastBlockFromNode)

			err = IndexFromNode(bt, client, int64(lastBlockFromBlocksTable)-*offsetBlocks, int64(lastBlockFromNode), *concurrencyBlocks)
			if err != nil {
				logrus.WithError(err).Fatalf("error indexing from node")
			}
		}

		if lastBlockFromDataTable < int(lastBlockFromNode) {
			// transforms = append(transforms, bt.TransformTx)

			logrus.Infof("missing blocks %v to %v in data table, indexing ...", lastBlockFromDataTable, lastBlockFromNode)
			err = IndexFromBigtable(bt, int64(lastBlockFromDataTable)-*offsetData, int64(lastBlockFromNode), transforms, *concurrencyData)
			if err != nil {
				logrus.WithError(err).Fatalf("error indexing from bigtable")
			}
		}

		logrus.Infof("index run completed")

		if *enableBalanceUpdater {
			ProcessMetadataUpdates(bt, client, *balanceUpdaterPrefix, *balanceUpdaterBatchSize)
		}
		time.Sleep(time.Second * 14)
	}

	// utils.WaitForCtrlC()

}

func ProcessMetadataUpdates(bt *db.Bigtable, client *rpc.ErigonClient, prefix string, batchSize int) {
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

	for {
		start := time.Now()
		updates, err := bt.GetMetadataUpdates(lastKey, batchSize)
		if err != nil {
			logrus.Fatal(err)
		}

		balances, err := client.GetBalances(updates)

		if err != nil {
			logrus.Fatalf("error retrieving balances from node: %v", err)
		}

		err = bt.SaveBalances(balances, updates)
		if err != nil {
			logrus.Fatalf("error saving balances to bigtable: %v", err)
		}

		// for i, b := range balances {

		// 	if len(b) > 0 {
		// 		logrus.Infof("balance for key %v is %x", updates[i], b)
		// 	}
		// }
		logrus.Infof("retrieved %v balances in %v, currently at %v", len(balances), time.Since(start), lastKey)

		lastKey = updates[len(updates)-1]
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

func IndexFromNode(bt *db.Bigtable, client *rpc.ErigonClient, start, end, concurrency int64) error {

	g := new(errgroup.Group)
	g.SetLimit(int(concurrency))

	startTs := time.Now()
	lastTickTs := time.Now()

	processedBlocks := int64(0)

	for i := start; i <= end; i++ {

		i := i
		g.Go(func() error {
			blockStartTs := time.Now()
			bc, timings, err := client.GetBlock(i)

			if err != nil {
				logrus.Error(err)
				return err
			}

			dbStart := time.Now()
			err = bt.SaveBlock(bc)
			if err != nil {
				logrus.Error(err)
				return err
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

	return g.Wait()
}

func IndexFromBigtable(bt *db.Bigtable, start, end int64, transforms []func(blk *types.Eth1Block, cache *ccache.Cache) (bulkData *types.BulkMutations, bulkMetadataUpdates *types.BulkMutations, err error), concurrency int64) error {
	g := new(errgroup.Group)
	g.SetLimit(int(concurrency))

	startTs := time.Now()
	lastTickTs := time.Now()

	processedBlocks := int64(0)

	cache := ccache.New(ccache.Configure().MaxSize(1000000).ItemsToPrune(500))

	logrus.Infof("fetching blocks from %d to %d", start, end)
	for i := start; i <= end; i++ {
		i := i
		g.Go(func() error {

			block, err := bt.GetBlockFromBlocksTable(uint64(i))
			if err != nil {
				logrus.Fatal(err)
				return err
			}

			bulkMutsData := types.BulkMutations{}
			bulkMutsMetadataUpdate := types.BulkMutations{}
			for _, transform := range transforms {
				mutsData, mutsMetadataUpdate, err := transform(block, cache)
				if err != nil {
					logrus.WithError(err).Error("error transforming block")
				}
				bulkMutsData.Keys = append(bulkMutsData.Keys, mutsData.Keys...)
				bulkMutsData.Muts = append(bulkMutsData.Muts, mutsData.Muts...)

				if mutsMetadataUpdate != nil {
					bulkMutsMetadataUpdate.Keys = append(bulkMutsMetadataUpdate.Keys, mutsMetadataUpdate.Keys...)
					bulkMutsMetadataUpdate.Muts = append(bulkMutsMetadataUpdate.Muts, mutsMetadataUpdate.Muts...)
				}
			}

			if len(bulkMutsData.Keys) > 0 {
				err = bt.WriteBulk(&bulkMutsData, bt.GetDataTable())
				if err != nil {
					return fmt.Errorf("error writing to bigtable data table: %w", err)
				}
			}

			if len(bulkMutsMetadataUpdate.Keys) > 0 {
				err = bt.WriteBulk(&bulkMutsMetadataUpdate, bt.GetMetadataUpdatesTable())
				if err != nil {
					return fmt.Errorf("error writing to bigtable metadata updates table: %w", err)
				}
			}

			current := atomic.AddInt64(&processedBlocks, 1)
			if current%500 == 0 {
				r := end - start
				if r == 0 {
					r = 1
				}
				perc := float64(i-start) * 100 / float64(r)
				logrus.Infof("currently processing block: %v; processed %v blocks in %v (%.1f blocks / sec); sync is %.1f%% complete", block.GetNumber(), current, time.Since(startTs), float64((current))/time.Since(lastTickTs).Seconds(), perc)
				lastTickTs = time.Now()
				atomic.StoreInt64(&processedBlocks, 0)
			}
			return nil
		})

	}

	if err := g.Wait(); err == nil {
		logrus.Info("Successfully fetched all blocks")
	} else {
		logrus.Error(err)
		return err
	}

	return nil
}

func ImportMainnetERC20TokenMetadataFromTokenDirectory(bt *db.Bigtable) {

	client := &http.Client{Timeout: time.Second * 10}

	resp, err := client.Get("<INSERT_TOKENLIST_URL>")

	if err != nil {
		logrus.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		logrus.Fatal(err)
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
		logrus.Fatal(err)
	}

	for _, token := range td.Tokens {

		address, err := hex.DecodeString(strings.TrimPrefix(token.Address, "0x"))
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.Infof("processing token %v at address %x", token.Name, address)

		meta := &types.ERC20Metadata{}
		meta.Decimals = big.NewInt(token.Decimals).Bytes()
		meta.Description = token.Extensions.Description
		if len(token.LogoURI) > 0 {
			resp, err := client.Get(token.LogoURI)

			if err == nil && resp.StatusCode == 200 {
				body, err := ioutil.ReadAll(resp.Body)

				if err != nil {
					logrus.Fatal(err)
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
			logrus.Fatal(err)
		}
		time.Sleep(time.Millisecond * 250)
	}

}

func ImportNameLabels(bt *db.Bigtable) {
	type NameEntry struct {
		Name string
	}

	res := make(map[string]*NameEntry)

	data, err := ioutil.ReadFile("")

	if err != nil {
		logrus.Fatal(err)
	}

	err = json.Unmarshal(data, &res)

	if err != nil {
		logrus.Fatal(err)
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
