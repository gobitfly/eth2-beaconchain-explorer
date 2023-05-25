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

	"github.com/coocood/freecache"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/sirupsen/logrus"

	_ "net/http/pprof"
)

func main() {
	erigonEndpoint := flag.String("erigon", "", "Erigon archive node enpoint")
	block := flag.Int64("block", 0, "Index a specific block")

	startBlocks := flag.Int64("blocks.start", 0, "Block to start indexing")
	endBlocks := flag.Int64("blocks.end", 0, "Block to finish indexing")

	concurrencyData := flag.Int64("data.concurrency", 30, "Concurrency to use when indexing data from bigtable")
	batchSize := flag.Int64("data.batchSize", 1000, "Batch size")

	bigtableProject := flag.String("bigtable.project", "", "Bigtable project")
	bigtableInstance := flag.String("bigtable.instance", "", "Bigtable instance")

	transformerFlag := flag.String("transformers", "", "Comma separated list of transformer functions")

	versionFlag := flag.Bool("version", false, "Print version and exit")

	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")

	flag.Parse()

	if *versionFlag {
		fmt.Println(version.Version)
		return
	}

	logrus.Infof("transformerFlag: %v %s", *transformerFlag, *transformerFlag)
	transformerList := strings.Split(*transformerFlag, ",")
	if len(transformerList) == 0 {
		utils.LogError(nil, "no transformer functions provided", 0)
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
		logrus.Infof("Starting to index a single block: %d", *block)
		err = bt.IndexEventsWithTransformers(*block, *block, transforms, *concurrencyData, cache)
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
	blockCount := utils.Int64Max(1, *batchSize)

	logrus.Infof("Starting to index all blocks ranging from %d to %d", *startBlocks, to)
	for from := *startBlocks; from <= to; from = from + blockCount {
		toBlock := utils.Int64Min(to, from+blockCount-1)

		logrus.Infof("indexing missing ens blocks %v to %v in data table ...", from, toBlock)
		err = bt.IndexEventsWithTransformers(int64(from), int64(toBlock), transforms, *concurrencyData, cache)
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
