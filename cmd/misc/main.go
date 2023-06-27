package main

import (
	"bytes"
	"context"
	"eth2-exporter/db"
	"eth2-exporter/exporter"
	"eth2-exporter/rpc"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/coocood/freecache"
	_ "github.com/jackc/pgx/v4/stdlib"
	utilMath "github.com/protolambda/zrnt/eth2/util/math"
	"golang.org/x/sync/errgroup"

	"flag"

	"github.com/sirupsen/logrus"
)

var opts = struct {
	Command         string
	User            uint64
	TargetVersion   int64
	StartEpoch      uint64
	EndEpoch        uint64
	StartDay        uint64
	EndDay          uint64
	Validator       uint64
	StartBlock      uint64
	EndBlock        uint64
	BatchSize       uint64
	DataConcurrency uint64
	Transformers    string
	Family          string
	Key             string
	DryRun          bool
}{}

func main() {
	configPath := flag.String("config", "config/default.config.yml", "Path to the config file")
	flag.StringVar(&opts.Command, "command", "", "command to run, available: updateAPIKey, applyDbSchema, epoch-export, debug-rewards, clear-bigtable, index-old-eth1-blocks, update-aggregation-bits")
	flag.Uint64Var(&opts.StartEpoch, "start-epoch", 0, "start epoch")
	flag.Uint64Var(&opts.EndEpoch, "end-epoch", 0, "end epoch")
	flag.Uint64Var(&opts.User, "user", 0, "user id")
	flag.Uint64Var(&opts.StartDay, "day-start", 0, "start day to debug")
	flag.Uint64Var(&opts.EndDay, "day-end", 0, "end day to debug")
	flag.Uint64Var(&opts.Validator, "validator", 0, "validator to check for")
	flag.Int64Var(&opts.TargetVersion, "target-version", -2, "Db migration target version, use -2 to apply up to the latest version, -1 to apply only the next version or the specific versions")
	flag.StringVar(&opts.Family, "family", "", "big table family")
	flag.StringVar(&opts.Key, "key", "", "big table key")
	flag.Uint64Var(&opts.StartBlock, "blocks.start", 0, "Block to start indexing")
	flag.Uint64Var(&opts.EndBlock, "blocks.end", 0, "Block to finish indexing")
	flag.Uint64Var(&opts.DataConcurrency, "data.concurrency", 30, "Concurrency to use when indexing data from bigtable")
	flag.Uint64Var(&opts.BatchSize, "data.batchSize", 1000, "Batch size")
	flag.StringVar(&opts.Transformers, "transformers", "", "Comma separated list of transformers used by the eth1 indexer")
	dryRun := flag.String("dry-run", "true", "if 'false' it deletes all rows starting with the key, per default it only logs the rows that would be deleted, but does not really delete them")
	versionFlag := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version.Version)
		return
	}

	opts.DryRun = *dryRun != "false"

	logrus.WithField("config", *configPath).WithField("version", version.Version).Printf("starting")
	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg

	chainIdString := strconv.FormatUint(utils.Config.Chain.Config.DepositChainID, 10)

	bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, chainIdString, utils.Config.RedisCacheEndpoint)
	if err != nil {
		utils.LogFatal(err, "error initializing bigtable", 0)
	}

	chainIDBig := new(big.Int).SetUint64(utils.Config.Chain.Config.DepositChainID)
	rpcClient, err := rpc.NewLighthouseClient("http://"+cfg.Indexer.Node.Host+":"+cfg.Indexer.Node.Port, chainIDBig)
	if err != nil {
		utils.LogFatal(err, "lighthouse client error", 0)
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
	db.MustInitFrontendDB(&types.DatabaseConfig{
		Username:     cfg.Frontend.WriterDatabase.Username,
		Password:     cfg.Frontend.WriterDatabase.Password,
		Name:         cfg.Frontend.WriterDatabase.Name,
		Host:         cfg.Frontend.WriterDatabase.Host,
		Port:         cfg.Frontend.WriterDatabase.Port,
		MaxOpenConns: cfg.Frontend.WriterDatabase.MaxOpenConns,
		MaxIdleConns: cfg.Frontend.WriterDatabase.MaxIdleConns,
	}, &types.DatabaseConfig{
		Username:     cfg.Frontend.ReaderDatabase.Username,
		Password:     cfg.Frontend.ReaderDatabase.Password,
		Name:         cfg.Frontend.ReaderDatabase.Name,
		Host:         cfg.Frontend.ReaderDatabase.Host,
		Port:         cfg.Frontend.ReaderDatabase.Port,
		MaxOpenConns: cfg.Frontend.ReaderDatabase.MaxOpenConns,
		MaxIdleConns: cfg.Frontend.ReaderDatabase.MaxIdleConns,
	})
	defer db.FrontendReaderDB.Close()
	defer db.FrontendWriterDB.Close()

	switch opts.Command {
	case "updateAPIKey":
		err := UpdateAPIKey(opts.User)
		if err != nil {
			logrus.WithError(err).Fatal("error updating API key")
		}
	case "applyDbSchema":
		logrus.Infof("applying db schema")
		err := db.ApplyEmbeddedDbSchema(opts.TargetVersion)
		if err != nil {
			logrus.WithError(err).Fatal("error applying db schema")
		}
		logrus.Infof("db schema applied successfully")
	case "epoch-export":
		logrus.Infof("exporting epochs %v - %v", opts.StartEpoch, opts.EndEpoch)

		err = services.InitLastAttestationCache(utils.Config.LastAttestationCachePath)
		if err != nil {
			logrus.Fatalf("error initializing last attesation cache: %v", err)
		}

		for epoch := opts.StartEpoch; epoch <= opts.EndEpoch; epoch++ {
			err = exporter.ExportEpoch(epoch, rpcClient)

			if err != nil {
				logrus.Errorf("error exporting epoch: %v", err)
			}
			logrus.Printf("finished export for epoch %v", epoch)
		}
	case "debug-rewards":
		CompareRewards(opts.StartDay, opts.EndDay, opts.Validator, bt)
	case "clear-bigtable":
		ClearBigtable(opts.Family, opts.Key, opts.DryRun, bt)
	case "index-old-eth1-blocks":
		IndexOldEth1Blocks(opts.StartBlock, opts.EndBlock, opts.BatchSize, opts.DataConcurrency, opts.Transformers, bt)
	case "update-aggregation-bits":
		updateAggreationBits(rpcClient, opts.StartEpoch, opts.EndEpoch, opts.DataConcurrency)
	default:
		utils.LogFatal(nil, "unknown command", 0)
	}
}

func updateAggreationBits(rpcClient *rpc.LighthouseClient, startEpoch uint64, endEpoch uint64, concurency uint64) {
	logrus.Infof("update-aggregation-bits epochs %v - %v", startEpoch, endEpoch)
	for epoch := startEpoch; epoch <= endEpoch; epoch++ {
		logrus.Infof("Getting data from the node for epoch %v", epoch)
		data, err := rpcClient.GetEpochData(epoch, false)
		if err != nil {
			utils.LogError(err, fmt.Sprintf("Error getting epoch[%v] data from the client", epoch), 0)
			return
		}

		ctx := context.Background()
		g, gCtx := errgroup.WithContext(ctx)
		g.SetLimit(int(concurency))

		for _, bm := range data.Blocks {
			for _, b := range bm {
				block := b
				logrus.Infof("Updating data for slot %v", block.Slot)

				if len(block.Attestations) == 0 {
					logrus.Infof("No Attestations for slot %v", block.Slot)

					g.Go(func() error {
						select {
						case <-gCtx.Done():
							return nil // halt once processing of a slot failed
						default:
						}

						// if we have some obsolete attestations we clean them from the db
						rows, err := db.WriterDb.Exec(`
								DELETE FROM blocks_attestations
								WHERE
									block_slot=$1
							`, block.Slot)
						if err != nil {
							return fmt.Errorf("error deleting obsolete attestations for Slot [%v]:  %v", block.Slot, err)
						}
						if rowsAffected, _ := rows.RowsAffected(); rowsAffected > 0 {
							logrus.Infof("%v obsolete attestations removed for Slot[%v]", rowsAffected, block.Slot)
						} else {
							logrus.Infof("No obsolete attestations found for Slot[%v] so we move on", block.Slot)
						}

						return nil
					})
					continue
				}
				for i, a := range block.Attestations {
					att := a
					index := i
					g.Go(func() error {
						select {
						case <-gCtx.Done():
							return nil // halt once processing of a slot failed
						default:
						}
						var aggregationbits *[]byte

						// block_slot and block_index are already unique, but to be sure we use the correct index we also check the signature
						err := db.ReaderDb.Get(&aggregationbits, `
						SELECT aggregationbits
						FROM blocks_attestations WHERE 
							block_slot=$1 AND
							block_index=$2 AND
							signature = $3`, block.Slot, index, att.Signature)
						if err != nil {
							return fmt.Errorf("error getting aggregationbits on Slot [%v] Index [%v] with Sig [%v]: %v", block.Slot, index, att.Signature, err)
						}

						if !bytes.Equal(*aggregationbits, att.AggregationBits) {
							_, err = db.WriterDb.Exec(`
								UPDATE blocks_attestations
								SET
									aggregationbits=$1
								WHERE
									block_slot=$2 AND
									block_index=$3 AND
									signature = $4
							`, att.AggregationBits, block.Slot, index, att.Signature)
							if err != nil {
								return fmt.Errorf("error updating aggregationbits on Slot [%v] Index [%v] :  %v", block.Slot, index, err)
							}
							logrus.Infof("Update of Slot[%v] Index[%v] complete", block.Slot, index)
						} else {
							logrus.Infof("Slot[%v] Index[%v] was already up to date", block.Slot, index)
						}

						return nil
					})

				}
			}
		}

		err = g.Wait()

		if err != nil {
			utils.LogError(err, fmt.Sprintf("error updating aggregationbits for epoch [%v]", epoch), 0)
			return
		}
		logrus.Infof("Update of Epoch[%v] complete", epoch)
	}
}

// Updates a users API key
func UpdateAPIKey(user uint64) error {
	type User struct {
		PHash  string `db:"password"`
		Email  string `db:"email"`
		OldKey string `db:"api_key"`
	}

	var u User
	err := db.FrontendWriterDB.Get(&u, `SELECT password, email, api_key from users where id = $1`, user)
	if err != nil {
		return fmt.Errorf("error getting current user, err: %w", err)
	}

	apiKey, err := utils.GenerateRandomAPIKey()
	if err != nil {
		return err
	}

	logrus.Infof("updating api key for user %v from old key: %v to new key: %v", user, u.OldKey, apiKey)

	tx, err := db.FrontendWriterDB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`UPDATE api_statistics set apikey = $1 where apikey = $2`, apiKey, u.OldKey)
	if err != nil {
		return err
	}

	rows, err := tx.Exec(`UPDATE users SET api_key = $1 WHERE id = $2`, apiKey, user)
	if err != nil {
		return err
	}

	amount, err := rows.RowsAffected()
	if err != nil {
		return err
	}

	if amount > 1 {
		return fmt.Errorf("error too many rows affected expected 1 but got: %v", amount)
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// Debugging function to compare Rewards from the Statistic Table with the onces from the Big Table
func CompareRewards(dayStart uint64, dayEnd uint64, validator uint64, bt *db.Bigtable) {

	for day := dayStart; day <= dayEnd; day++ {
		startEpoch := day * utils.EpochsPerDay()
		endEpoch := startEpoch + utils.EpochsPerDay() - 1
		hist, err := bt.GetValidatorIncomeDetailsHistory([]uint64{validator}, startEpoch, endEpoch)
		if err != nil {
			logrus.Fatal(err)
		}
		var tot int64
		for _, rew := range hist[validator] {
			tot += rew.TotalClRewards()
		}
		logrus.Infof("Total CL Rewards for day [%v]: %v", day, tot)
		var dbRewards *int64
		err = db.ReaderDb.Get(&dbRewards, `
		SELECT 
		COALESCE(cl_rewards_gwei, 0) AS cl_rewards_gwei
		FROM validator_stats WHERE day = $1 and validatorindex = $2`, day, validator)
		if err != nil {
			logrus.Fatalf("error getting cl_rewards_gwei from db: %v", err)
			return
		}
		if tot != *dbRewards {
			logrus.Errorf("Rewards are not the same on day %v-> big: %v, db: %v", day, tot, *dbRewards)
		}
	}

}

func ClearBigtable(family string, key string, dryRun bool, bt *db.Bigtable) {

	if !dryRun {
		confirmation := utils.CmdPrompt(fmt.Sprintf("Are you sure you want to delete all big table entries starting with [%v] for family [%v]?", key, family))
		if confirmation != "yes" {
			logrus.Infof("Abort!")
			return
		}
	}
	deletedKeys, err := bt.ClearByPrefix(family, key, dryRun)

	if err != nil {
		logrus.Fatalf("error deleting from bigtable: %v", err)
	} else if dryRun {
		logrus.Infof("the following keys would be deleted: %v", deletedKeys)
	} else {
		logrus.Infof("%v keys have been deleted", len(deletedKeys))
	}
}

func IndexOldEth1Blocks(startBlock uint64, endBlock uint64, batchSize uint64, concurrency uint64, transformerFlag string, bt *db.Bigtable) {
	if endBlock > 0 && endBlock < startBlock {
		utils.LogError(nil, fmt.Sprintf("endBlock [%v] < startBlock [%v]", endBlock, startBlock), 0)
		return
	}
	if concurrency == 0 {
		utils.LogError(nil, "concurrency must be greater than 0", 0)
		return
	}
	if bt == nil {
		utils.LogError(nil, "no bigtable provided", 0)
		return
	}

	client, err := rpc.NewErigonClient(utils.Config.Eth1ErigonEndpoint)
	if err != nil {
		logrus.Fatalf("error initializing erigon client: %v", err)
	}

	transforms := make([]func(blk *types.Eth1Block, cache *freecache.Cache) (*types.BulkMutations, *types.BulkMutations, error), 0)

	logrus.Infof("transformerFlag: %v", transformerFlag)
	transformerList := strings.Split(transformerFlag, ",")
	if len(transformerList) == 0 {
		utils.LogError(nil, "no transformer functions provided", 0)
		return
	}
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

	if startBlock == 0 && endBlock == 0 {
		utils.LogFatal(nil, "no start+end block defined", 0)
		return
	}

	lastBlockFromBlocksTable, err := bt.GetLastBlockInBlocksTable()
	if err != nil {
		utils.LogError(err, "error retrieving last blocks from blocks table", 0)
		return
	}

	to := uint64(lastBlockFromBlocksTable)
	if endBlock > 0 {
		to = utilMath.MinU64(to, endBlock)
	}
	blockCount := utilMath.MaxU64(1, batchSize)

	logrus.Infof("Starting to index all blocks ranging from %d to %d", startBlock, to)
	for from := startBlock; from <= to; from = from + blockCount {
		toBlock := utilMath.MinU64(to, from+blockCount-1)

		logrus.Infof("indexing blocks %v to %v in data table ...", from, toBlock)
		err = bt.IndexEventsWithTransformers(int64(from), int64(toBlock), transforms, int64(concurrency), cache)
		if err != nil {
			utils.LogError(err, "error indexing from bigtable", 0)
		}
		cache.Clear()

	}

	if importENSChanges {
		if err = bt.ImportEnsUpdates(client.GetNativeClient()); err != nil {
			utils.LogError(err, "error importing ens from events", 0)
			return
		}
	}

	logrus.Infof("index run completed")
}
