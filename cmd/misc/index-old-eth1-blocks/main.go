package main

import (
	"context"
	"eth2-exporter/db"
	"eth2-exporter/rpc"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/coocood/freecache"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	_ "net/http/pprof"
)

func main() {
	erigonEndpoint := flag.String("erigon", "", "Erigon archive node enpoint")
	block := flag.Int64("block", 0, "Index a specific block")

	startBlocks := flag.Int64("blocks.start", 0, "Block to start indexing")
	endBlocks := flag.Int64("blocks.end", 0, "Block to finish indexing")

	concurrencyData := flag.Int64("data.concurrency", 30, "Concurrency to use when indexing data from bigtable")
	offsetData := flag.Int64("data.offset", 1000, "Data offset")

	bigtableProject := flag.String("bigtable.project", "", "Bigtable project")
	bigtableInstance := flag.String("bigtable.instance", "", "Bigtable instance")

	transformerFlag := flag.String("transformers", "", "Comma separated list of transformer functions")

	versionFlag := flag.Bool("version", false, "Print version and exit")

	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")

	flag.Parse()

	logrus.Infof("transformerFlag: %v %s", *transformerFlag, *transformerFlag)
	transformerList := strings.Split(*transformerFlag, ",")

	if len(transformerList) == 0 {
		utils.LogError(nil, "no transformer functions provided")
		return
	}

	if *versionFlag {
		fmt.Println(version.Version)
		return
	}

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		utils.LogError(err, "error reading config file", 0)
	}
	utils.Config = cfg
	logrus.WithField("config", *configPath).WithField("version", version.Version).WithField("chainName", utils.Config.Chain.Config.ConfigName).Printf("starting")

	// enable pprof endpoint if requested
	if utils.Config.Pprof.Enabled {
		go func() {
			logrus.Infof("starting pprof http server on port %s", utils.Config.Pprof.Port)
			logrus.Info(http.ListenAndServe(fmt.Sprintf("localhost:%s", utils.Config.Pprof.Port), nil))
		}()
	}

	db.MustInitDB(&types.DatabaseConfig{
		Username: cfg.WriterDatabase.Username,
		Password: cfg.WriterDatabase.Password,
		Name:     cfg.WriterDatabase.Name,
		Host:     cfg.WriterDatabase.Host,
		Port:     cfg.WriterDatabase.Port,
	}, &types.DatabaseConfig{
		Username: cfg.ReaderDatabase.Username,
		Password: cfg.ReaderDatabase.Password,
		Name:     cfg.ReaderDatabase.Name,
		Host:     cfg.ReaderDatabase.Host,
		Port:     cfg.ReaderDatabase.Port,
	})
	defer db.ReaderDb.Close()
	defer db.WriterDb.Close()

	if erigonEndpoint == nil || *erigonEndpoint == "" {
		utils.LogFatal(nil, "no erigon node url provided", 0)
	}

	logrus.Infof("using erigon node at %v", *erigonEndpoint)
	client, err := rpc.NewErigonClient(*erigonEndpoint)
	if err != nil {
		utils.LogFatal(err, "erigon client creation error", 0)
	}

	chainId := strconv.FormatUint(utils.Config.Chain.Config.DepositChainID, 10)

	nodeChainId, err := client.GetNativeClient().ChainID(context.Background())
	if err != nil {
		utils.LogFatal(err, "node chain id error", 0)
	}

	if nodeChainId.String() != chainId {
		utils.LogFatal(nil, fmt.Errorf("node chain id mismatch, wanted %v got %v", chainId, nodeChainId.String()), 0)
	}

	bt, err := db.InitBigtable(*bigtableProject, *bigtableInstance, chainId)
	if err != nil {
		utils.LogFatal(nil, fmt.Errorf("error connecting to bigtable: %v", err), 0)
	}
	defer bt.Close()

	transforms := make([]func(blk *types.Eth1Block, cache *freecache.Cache) (*types.BulkMutations, *types.BulkMutations, error), 0)

	logrus.Infof("transformers: %v", transformerList)
	importENSChanges := false
	/**
	* Add additional transformers you want to sync to this switch case
	**/
	for _, t := range transformerList {
		switch t {
		case "TransformBlock":
			transforms = append(transforms, bt.TransformBlock)
		case "TransformTx":
			transforms = append(transforms, bt.TransformTx)
		case "TransformItx":
			transforms = append(transforms, bt.TransformItx)
		case "TransformERC20":
			transforms = append(transforms, bt.TransformERC20)
		case "TransformERC721":
			transforms = append(transforms, bt.TransformERC721)
		case "TransformERC1155":
			transforms = append(transforms, bt.TransformERC1155)
		case "TransformWithdrawals":
			transforms = append(transforms, bt.TransformWithdrawals)
		case "TransformUncle":
			transforms = append(transforms, bt.TransformUncle)
		case "TransformEnsNameRegistered":
			transforms = append(transforms, bt.TransformEnsNameRegistered)
			importENSChanges = true
		default:
			utils.LogError(nil, "Invalid transformer flag %v", 0)
			return
		}
	}

	cache := freecache.NewCache(100 * 1024 * 1024) // 100 MB limit

	if *block != 0 {
		err = IndexFromBigtable(bt, *block, *block, transforms, *concurrencyData, cache)
		if err != nil {
			utils.LogFatal(err, "error indexing from bigtable", 0)
		}
		cache.Clear()

		logrus.Infof("indexing of block %v completed", *block)
		return
	}

	if *startBlocks == 0 {
		utils.LogFatal(err, "no start block defined", 0)
		return
	}

	lastBlockFromBlocksTable, err := bt.GetLastBlockInBlocksTable()
	if err != nil {
		utils.LogError(err, "error retrieving last blocks from blocks table", 0)
		return
	}

	to := int64(lastBlockFromBlocksTable)
	if *endBlocks > 0 {
		to = utils.Int64Min(to, *endBlocks)
	}
	blockCount := utils.Int64Max(1, *offsetData)

	for from := *startBlocks; from <= to; from = from + blockCount {
		toBlock := utils.Int64Min(to, from+blockCount-1)

		logrus.Infof("indexing missing ens blocks %v to %v in data table ...", from, toBlock)
		err = IndexFromBigtable(bt, int64(from), int64(toBlock), transforms, *concurrencyData, cache)
		if err != nil {
			utils.LogError(err, "error indexing from bigtable", 0)
		}
		cache.Clear()

	}

	if importENSChanges {
		err = bt.ImportEnsUpdates(client.GetNativeClient())
		if err != nil {
			utils.LogError(err, "error importing ens from events", 0)
			return
		}
	}

	logrus.Infof("index run completed")
}

func IndexFromBigtable(bt *db.Bigtable, start, end int64, transforms []func(blk *types.Eth1Block, cache *freecache.Cache) (bulkData *types.BulkMutations, bulkMetadataUpdates *types.BulkMutations, err error), concurrency int64, cache *freecache.Cache) error {
	g := new(errgroup.Group)
	g.SetLimit(int(concurrency))

	startTs := time.Now()
	lastTickTs := time.Now()

	processedBlocks := int64(0)

	logrus.Infof("fetching blocks from %d to %d", start, end)
	for i := start; i <= end; i++ {
		i := i
		g.Go(func() error {

			block, err := bt.GetBlockFromBlocksTable(uint64(i))
			if err != nil {
				return fmt.Errorf("error getting block: %v from bigtable blocks table err: %w", i, err)
			}

			bulkMutsData := types.BulkMutations{}
			bulkMutsMetadataUpdate := types.BulkMutations{}
			for _, transform := range transforms {
				mutsData, mutsMetadataUpdate, err := transform(block, cache)
				if err != nil {
					utils.LogError(err, "error transforming block", 0)
				}
				bulkMutsData.Keys = append(bulkMutsData.Keys, mutsData.Keys...)
				bulkMutsData.Muts = append(bulkMutsData.Muts, mutsData.Muts...)

				if mutsMetadataUpdate != nil {
					bulkMutsMetadataUpdate.Keys = append(bulkMutsMetadataUpdate.Keys, mutsMetadataUpdate.Keys...)
					bulkMutsMetadataUpdate.Muts = append(bulkMutsMetadataUpdate.Muts, mutsMetadataUpdate.Muts...)
				}
			}

			if len(bulkMutsData.Keys) > 0 {
				metaKeys := strings.Join(bulkMutsData.Keys, ",") // save block keys in order to be able to handle chain reorgs
				err = bt.SaveBlockKeys(block.Number, block.Hash, metaKeys)
				if err != nil {
					return fmt.Errorf("error saving block keys to bigtable metadata updates table: %w", err)
				}

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
		logrus.Info("data table indexing completed")
	} else {
		utils.LogError(err, "wait group error", 0)
		return err
	}

	return nil
}
