package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"eth2-exporter/cmd/misc/commands"
	"eth2-exporter/db"
	"eth2-exporter/exporter"
	"eth2-exporter/rpc"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/coocood/freecache"
	"github.com/ethereum/go-ethereum/common"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	utilMath "github.com/protolambda/zrnt/eth2/util/math"
	go_ens "github.com/wealdtech/go-ens/v3"
	"golang.org/x/sync/errgroup"

	"flag"

	"github.com/sirupsen/logrus"
)

var opts = struct {
	Command             string
	User                uint64
	Addresses           string
	TargetVersion       int64
	StartEpoch          uint64
	EndEpoch            uint64
	StartDay            uint64
	EndDay              uint64
	Validator           uint64
	StartBlock          uint64
	EndBlock            uint64
	BatchSize           uint64
	DataConcurrency     uint64
	Transformers        string
	Table               string
	Columns             string
	Family              string
	Key                 string
	ValidatorNameRanges string
	DryRun              bool
}{}

func main() {
	statsPartitionCommand := commands.StatsMigratorCommand{}

	configPath := flag.String("config", "config/default.config.yml", "Path to the config file")
	flag.StringVar(&opts.Command, "command", "", "command to run, available: updateAPIKey, applyDbSchema, initBigtableSchema, epoch-export, debug-rewards, debug-blocks, clear-bigtable, index-old-eth1-blocks, update-aggregation-bits, historic-prices-export, index-missing-blocks, export-epoch-missed-slots, migrate-last-attestation-slot-bigtable, export-genesis-validators, update-block-finalization-sequentially, nameValidatorsByRanges, export-stats-totals, export-sync-committee-periods, export-sync-committee-validator-stats, partition-validator-stats")
	flag.Uint64Var(&opts.StartEpoch, "start-epoch", 0, "start epoch")
	flag.Uint64Var(&opts.EndEpoch, "end-epoch", 0, "end epoch")
	flag.Uint64Var(&opts.User, "user", 0, "user id")
	flag.Uint64Var(&opts.StartDay, "day-start", 0, "start day to debug")
	flag.Uint64Var(&opts.EndDay, "day-end", 0, "end day to debug")
	flag.Uint64Var(&opts.Validator, "validator", 0, "validator to check for")
	flag.Int64Var(&opts.TargetVersion, "target-version", -2, "Db migration target version, use -2 to apply up to the latest version, -1 to apply only the next version or the specific versions")
	flag.StringVar(&opts.Table, "table", "", "big table table")
	flag.StringVar(&opts.Family, "family", "", "big table family")
	flag.StringVar(&opts.Key, "key", "", "big table key")
	flag.Uint64Var(&opts.StartBlock, "blocks.start", 0, "Block to start indexing")
	flag.Uint64Var(&opts.EndBlock, "blocks.end", 0, "Block to finish indexing")
	flag.Uint64Var(&opts.DataConcurrency, "data.concurrency", 30, "Concurrency to use when indexing data from bigtable")
	flag.Uint64Var(&opts.BatchSize, "data.batchSize", 1000, "Batch size")
	flag.StringVar(&opts.Transformers, "transformers", "", "Comma separated list of transformers used by the eth1 indexer")
	flag.StringVar(&opts.ValidatorNameRanges, "validator-name-ranges", "https://config.dencun-devnet-8.ethpandaops.io/api/v1/nodes/validator-ranges", "url to or json of validator-ranges (format must be: {'ranges':{'X-Y':'name'}})")
	flag.StringVar(&opts.Addresses, "addresses", "", "Comma separated list of addresses that should be processed by the command")
	flag.StringVar(&opts.Columns, "columns", "", "Comma separated list of columns that should be affected by the command")
	dryRun := flag.String("dry-run", "true", "if 'false' it deletes all rows starting with the key, per default it only logs the rows that would be deleted, but does not really delete them")
	versionFlag := flag.Bool("version", false, "Show version and exit")

	statsPartitionCommand.ParseCommandOptions()
	flag.Parse()

	if *versionFlag {
		fmt.Println(version.Version)
		fmt.Println(version.GoVersion)
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

	chainIdString := strconv.FormatUint(utils.Config.Chain.ClConfig.DepositChainID, 10)

	bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, chainIdString, utils.Config.RedisCacheEndpoint)
	if err != nil {
		utils.LogFatal(err, "error initializing bigtable", 0)
	}

	chainIDBig := new(big.Int).SetUint64(utils.Config.Chain.ClConfig.DepositChainID)
	rpcClient, err := rpc.NewLighthouseClient("http://"+cfg.Indexer.Node.Host+":"+cfg.Indexer.Node.Port, chainIDBig)
	if err != nil {
		utils.LogFatal(err, "lighthouse client error", 0)
	}

	erigonClient, err := rpc.NewErigonClient(utils.Config.Eth1ErigonEndpoint)
	if err != nil {
		logrus.Fatalf("error initializing erigon client: %v", err)
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
	case "nameValidatorsByRanges":
		err := nameValidatorsByRanges(opts.ValidatorNameRanges)
		if err != nil {
			logrus.WithError(err).Fatal("error naming validators by ranges")
		}
	case "updateAPIKey":
		err := updateAPIKey(opts.User)
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
	case "initBigtableSchema":
		logrus.Infof("initializing bigtable schema")
		err := db.InitBigtableSchema()
		if err != nil {
			logrus.WithError(err).Fatal("error initializing bigtable schema")
		}
		logrus.Infof("bigtable schema initialization completed")
	case "epoch-export":
		logrus.Infof("exporting epochs %v - %v", opts.StartEpoch, opts.EndEpoch)
		for epoch := opts.StartEpoch; epoch <= opts.EndEpoch; epoch++ {
			tx, err := db.WriterDb.Beginx()
			if err != nil {
				logrus.Fatalf("error starting tx: %v", err)
			}
			for slot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch; slot < (epoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch; slot++ {
				err = exporter.ExportSlot(rpcClient, slot, false, tx)

				if err != nil {
					tx.Rollback()
					logrus.Fatalf("error exporting slot %v: %v", slot, err)
				}
				logrus.Printf("finished export for slot %v", slot)
			}
			err = tx.Commit()
			if err != nil {
				logrus.Fatalf("error committing tx: %v", err)
			}
		}
	case "export-epoch-missed-slots":
		logrus.Infof("exporting epochs with missed slots")
		latestFinalizedEpoch, err := db.GetLatestFinalizedEpoch()
		if err != nil {
			utils.LogError(err, "error getting latest finalized epoch from db", 0)
		}
		epochs := []uint64{}
		err = db.ReaderDb.Select(&epochs, `
			WITH last_exported_epoch AS (
				SELECT (MAX(epoch)*$1) AS slot 
				FROM epochs 
				WHERE epoch <= $2 
				AND rewards_exported
			)
			SELECT epoch 
			FROM blocks
			WHERE status = '0' 
				AND slot < (SELECT slot FROM last_exported_epoch)
			GROUP BY epoch 
			ORDER BY epoch;
		`, utils.Config.Chain.ClConfig.SlotsPerEpoch, latestFinalizedEpoch)
		if err != nil {
			utils.LogError(err, "Error getting epochs with missing slot status from db", 0)
			return
		} else if len(epochs) == 0 {
			logrus.Infof("No epochs with missing slot status found")
			return
		}

		logrus.Infof("Found %v epochs with missing slot status", len(epochs))
		for _, epoch := range epochs {
			tx, err := db.WriterDb.Beginx()
			if err != nil {
				logrus.Fatalf("error starting tx: %v", err)
			}
			for slot := epoch * utils.Config.Chain.ClConfig.SlotsPerEpoch; slot < (epoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch; slot++ {
				err = exporter.ExportSlot(rpcClient, slot, false, tx)

				if err != nil {
					tx.Rollback()
					logrus.Fatalf("error exporting slot %v: %v", slot, err)
				}
				logrus.Printf("finished export for slot %v", slot)
			}
			err = tx.Commit()
			if err != nil {
				logrus.Fatalf("error committing tx: %v", err)
			}
		}
	case "debug-rewards":
		compareRewards(opts.StartDay, opts.EndDay, opts.Validator, bt)
	case "debug-blocks":
		err = debugBlocks()
	case "clear-bigtable":
		clearBigtable(opts.Table, opts.Family, opts.Key, opts.DryRun, bt)
	case "index-old-eth1-blocks":
		indexOldEth1Blocks(opts.StartBlock, opts.EndBlock, opts.BatchSize, opts.DataConcurrency, opts.Transformers, bt, erigonClient)
	case "update-aggregation-bits":
		updateAggreationBits(rpcClient, opts.StartEpoch, opts.EndEpoch, opts.DataConcurrency)
	case "update-block-finalization-sequentially":
		err = updateBlockFinalizationSequentially()
	case "historic-prices-export":
		exportHistoricPrices(opts.StartDay, opts.EndDay)
	case "index-missing-blocks":
		indexMissingBlocks(opts.StartBlock, opts.EndBlock, bt, erigonClient)
	case "migrate-last-attestation-slot-bigtable":
		migrateLastAttestationSlotToBigtable()
	case "export-genesis-validators":
		logrus.Infof("retrieving genesis validator state")
		validators, err := rpcClient.GetValidatorState(0)
		if err != nil {
			logrus.Fatalf("error retrieving genesis validator state")
		}

		validatorsArr := make([]*types.Validator, 0, len(validators.Data))

		for _, validator := range validators.Data {
			validatorsArr = append(validatorsArr, &types.Validator{
				Index:                      uint64(validator.Index),
				PublicKey:                  utils.MustParseHex(validator.Validator.Pubkey),
				WithdrawalCredentials:      utils.MustParseHex(validator.Validator.WithdrawalCredentials),
				Balance:                    uint64(validator.Balance),
				EffectiveBalance:           uint64(validator.Validator.EffectiveBalance),
				Slashed:                    validator.Validator.Slashed,
				ActivationEligibilityEpoch: uint64(validator.Validator.ActivationEligibilityEpoch),
				ActivationEpoch:            uint64(validator.Validator.ActivationEpoch),
				ExitEpoch:                  uint64(validator.Validator.ExitEpoch),
				WithdrawableEpoch:          uint64(validator.Validator.WithdrawableEpoch),
				Status:                     "active_online",
			})
		}

		tx, err := db.WriterDb.Beginx()
		if err != nil {
			logrus.Fatalf("error starting tx: %v", err)
		}
		defer tx.Rollback()

		batchSize := 10000
		for i := 0; i < len(validatorsArr); i += batchSize {
			data := &types.EpochData{
				SyncDuties:        make(map[types.Slot]map[types.ValidatorIndex]bool),
				AttestationDuties: make(map[types.Slot]map[types.ValidatorIndex][]types.Slot),
				ValidatorAssignmentes: &types.EpochAssignments{
					ProposerAssignments: map[uint64]uint64{},
					AttestorAssignments: map[string]uint64{},
					SyncAssignments:     make([]uint64, 0),
				},
				Blocks:                  make(map[uint64]map[string]*types.Block),
				FutureBlocks:            make(map[uint64]map[string]*types.Block),
				EpochParticipationStats: &types.ValidatorParticipation{},
				Finalized:               false,
			}

			data.Validators = make([]*types.Validator, 0, batchSize)

			start := i
			end := i + batchSize
			if end >= len(validatorsArr) {
				end = len(validatorsArr) - 1
			}
			data.Validators = append(data.Validators, validatorsArr[start:end]...)

			logrus.Infof("saving validators %v-%v", data.Validators[0].Index, data.Validators[len(data.Validators)-1].Index)

			err = db.SaveValidators(0, data.Validators, rpcClient, len(data.Validators), tx)
			if err != nil {
				logrus.Fatal(err)
			}
		}

		logrus.Infof("exporting deposit data for genesis %v validators", len(validators.Data))
		for i, validator := range validators.Data {
			if i%1000 == 0 {
				logrus.Infof("exporting deposit data for genesis validator %v (of %v/%v)", validator.Index, i, len(validators.Data))
			}
			_, err = tx.Exec(`INSERT INTO blocks_deposits (block_slot, block_root, block_index, publickey, withdrawalcredentials, amount, signature)
			VALUES (0, '\x01', $1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`,
				validator.Index, utils.MustParseHex(validator.Validator.Pubkey), utils.MustParseHex(validator.Validator.WithdrawalCredentials), validator.Balance, []byte{0x0},
			)
			if err != nil {
				logrus.Errorf("error exporting genesis-deposits: %v", err)
				time.Sleep(time.Minute)
				continue
			}
		}

		_, err = tx.Exec(`
		INSERT INTO blocks (epoch, slot, blockroot, parentroot, stateroot, signature, syncaggregate_participation, proposerslashingscount, attesterslashingscount, attestationscount, depositscount, withdrawalcount, voluntaryexitscount, proposer, status, exec_transactions_count, eth1data_depositcount)
		VALUES (0, 0, '\x'::bytea, '\x'::bytea, '\x'::bytea, '\x'::bytea, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
		ON CONFLICT (slot, blockroot) DO NOTHING`)
		if err != nil {
			logrus.Fatal(err)
		}

		err = db.BigtableClient.SaveValidatorBalances(0, validatorsArr)
		if err != nil {
			logrus.Fatal(err)
		}

		err = tx.Commit()
		if err != nil {
			logrus.Fatal(err)
		}
	case "export-stats-totals":
		exportStatsTotals(opts.Columns, opts.StartDay, opts.EndDay, opts.DataConcurrency)
	case "export-sync-committee-periods":
		exportSyncCommitteePeriods(rpcClient, opts.StartDay, opts.EndDay, opts.DryRun)
	case "export-sync-committee-validator-stats":
		exportSyncCommitteeValidatorStats(rpcClient, opts.StartDay, opts.EndDay, opts.DryRun, true)
	case "fix-exec-transactions-count":
		err = fixExecTransactionsCount()
	case "partition-validator-stats":
		statsPartitionCommand.Config.DryRun = opts.DryRun
		err = statsPartitionCommand.StartStatsPartitionCommand()
	case "fix-ens":
		err = fixEns(erigonClient)
	case "fix-ens-addresses":
		err = fixEnsAddresses(erigonClient)
	default:
		utils.LogFatal(nil, fmt.Sprintf("unknown command %s", opts.Command), 0)
	}

	if err != nil {
		utils.LogFatal(err, "command returned error", 0)
	} else {
		logrus.Infof("command executed successfully")
	}
}

func fixEns(erigonClient *rpc.ErigonClient) error {
	logrus.Infof("command: fix-ens")
	addrs := []struct {
		Address []byte `db:"address"`
		EnsName string `db:"ens_name"`
	}{}
	err := db.WriterDb.Select(&addrs, `select address, ens_name from ens where is_primary_name = true`)
	if err != nil {
		return err
	}

	logrus.Infof("found %v ens entries", len(addrs))

	g := new(errgroup.Group)
	g.SetLimit(10) // limit load on the node

	batchSize := 100
	total := len(addrs)
	for i := 0; i < total; i += batchSize {
		to := i + batchSize
		if to > total {
			to = total
		}
		batch := addrs[i:to]

		logrus.Infof("processing batch %v-%v / %v", i, to, total)
		for _, addr := range batch {
			addr := addr
			g.Go(func() error {
				ensAddr, err := go_ens.Resolve(erigonClient.GetNativeClient(), addr.EnsName)
				if err != nil {
					if err.Error() == "unregistered name" ||
						err.Error() == "no address" ||
						err.Error() == "no resolver" ||
						err.Error() == "abi: attempting to unmarshall an empty string while arguments are expected" ||
						strings.Contains(err.Error(), "execution reverted") ||
						err.Error() == "invalid jump destination" {
						logrus.WithFields(logrus.Fields{"addr": fmt.Sprintf("%#x", addr.Address), "name": addr.EnsName, "reason": fmt.Sprintf("failed resolve: %v", err.Error())}).Warnf("deleting ens entry")
						if !opts.DryRun {
							_, err = db.WriterDb.Exec(`delete from ens where address = $1 and ens_name = $2`, addr.Address, addr.EnsName)
							if err != nil {
								return err
							}
						}
						return nil
					}
					return err
				}

				dbAddr := common.BytesToAddress(addr.Address)
				if dbAddr.Cmp(ensAddr) != 0 {
					logrus.WithFields(logrus.Fields{"addr": fmt.Sprintf("%#x", addr.Address), "name": addr.EnsName, "reason": fmt.Sprintf("dbAddr != resolved ensAddr: %#x != %#x", addr.Address, ensAddr.Bytes())}).Warnf("deleting ens entry")
					if !opts.DryRun {
						_, err = db.WriterDb.Exec(`delete from ens where address = $1 and ens_name = $2`, addr.Address, addr.EnsName)
						if err != nil {
							return err
						}
					}
				}

				reverseName, err := go_ens.ReverseResolve(erigonClient.GetNativeClient(), dbAddr)
				if err != nil {
					if err.Error() == "not a resolver" || err.Error() == "no resolution" {
						logrus.WithFields(logrus.Fields{"addr": fmt.Sprintf("%#x", addr.Address), "name": addr.EnsName, "reason": fmt.Sprintf("failed reverse-resolve: %v", err.Error())}).Warnf("updating ens entry: is_primary_name = false")
						if !opts.DryRun {
							_, err = db.WriterDb.Exec(`update ens set is_primary_name = false where address = $1 and ens_name = $2`, addr.Address, addr.EnsName)
							if err != nil {
								return err
							}
						}
						return nil
					}
					return err
				}

				if reverseName != addr.EnsName {
					logrus.WithFields(logrus.Fields{"addr": fmt.Sprintf("%#x", addr.Address), "name": addr.EnsName, "reason": fmt.Sprintf("resolved != reverseResolved: %v != %v", addr.EnsName, reverseName)}).Warnf("updating ens entry: is_primary_name = false")
					if !opts.DryRun {
						_, err = db.WriterDb.Exec(`update ens set is_primary_name = false where address = $1 and ens_name = $2`, addr.Address, addr.EnsName)
						if err != nil {
							return err
						}
					}
				}

				return nil
			})
		}

		err = g.Wait()
		if err != nil {
			return err
		}
		time.Sleep(time.Millisecond * 100)
	}
	return nil
}

func fixEnsAddresses(erigonClient *rpc.ErigonClient) error {
	logrus.WithFields(logrus.Fields{"dry": opts.DryRun}).Infof("command: fix-ens-addresses")
	if opts.Addresses == "" {
		return errors.New("no addresses specified")
	}

	type DbEntry struct {
		NameHash      []byte    `db:"name_hash"`
		EnsName       string    `db:"ens_name"`
		Address       []byte    `db:"address"`
		IsPrimaryName bool      `db:"is_primary_name"`
		ValidTo       time.Time `db:"valid_to"`
	}

	for _, addrHex := range strings.Split(opts.Addresses, ",") {
		if !common.IsHexAddress(addrHex) {
			return fmt.Errorf("invalid address: %v", addrHex)
		}

		addr := common.HexToAddress(addrHex)

		dbEntry := &DbEntry{}
		err := db.WriterDb.Get(dbEntry, `select name_hash, ens_name, address, is_primary_name, valid_to from ens where address = $1`, addr.Bytes())
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("error getting ens entry for addr [%v]: %w", addr.Hex(), err)
		}
		if err == sql.ErrNoRows {
			dbEntry = nil
		}

		name, err := go_ens.ReverseResolve(erigonClient.GetNativeClient(), addr)
		if err != nil {
			if err.Error() == "not a resolver" ||
				err.Error() == "no resolution" {
				logrus.WithFields(logrus.Fields{"addr": addr.Hex()}).Warnf("error reverse-resolving name: %v", err)
				if dbEntry != nil {
					logrus.WithFields(logrus.Fields{"addr": fmt.Sprintf("%v", addr.Hex()), "reason": fmt.Sprintf("error reverse-resolving name: %v", err)}).Warnf("deleting ens entry")
					if !opts.DryRun {
						_, err = db.WriterDb.Exec(`delete from ens where address = $1`, addr.Bytes())
						if err != nil {
							return fmt.Errorf("error deleting ens entry: %w", err)
						}
					}
				}
				continue
			} else {
				return fmt.Errorf("error go_ens.ReverseResolve for addr %v: %w", addr.Hex(), err)
			}
		}

		if !strings.HasSuffix(name, ".eth") {
			logrus.Infof("need to add .eth to %v for addr %v", name, addr.Hex())
			name = name + ".eth"
		}

		resolvedAddr, err := go_ens.Resolve(erigonClient.GetNativeClient(), name)
		if err != nil {
			if err.Error() == "unregistered name" ||
				err.Error() == "no address" ||
				err.Error() == "no resolver" ||
				err.Error() == "abi: attempting to unmarshall an empty string while arguments are expected" ||
				strings.Contains(err.Error(), "execution reverted") ||
				err.Error() == "invalid jump destination" {
				if dbEntry != nil {
					logrus.WithFields(logrus.Fields{"addr": fmt.Sprintf("%v", addr.Hex()), "reason": fmt.Sprintf("error resolving name: %v", err)}).Warnf("deleting ens entry")
					if !opts.DryRun {
						_, err = db.WriterDb.Exec(`delete from ens where address = $1`, addr.Bytes())
						if err != nil {
							return fmt.Errorf("error deleting ens entry: %w", err)
						}
					}
				}
			} else {
				return fmt.Errorf("error go_ens.Resolve(%v) for addr %v: %w", name, addr.Hex(), err)
			}
		}

		if !bytes.Equal(resolvedAddr.Bytes(), addr.Bytes()) {
			logrus.WithFields(logrus.Fields{"addr": fmt.Sprintf("%v", addr.Hex()), "reason": fmt.Sprintf("addr != resolvedAddr: %v != %v", addr.Hex(), resolvedAddr.Hex())}).Warnf("deleting ens entry")
			if !opts.DryRun {
				_, err = db.WriterDb.Exec(`delete from ens where address = $1`, addr.Bytes())
				if err != nil {
					return fmt.Errorf("error deleting ens entry: %w", err)
				}
			}
		}

		nameHash, err := go_ens.NameHash(name)
		if err != nil {
			return fmt.Errorf("error go_ens.NameHash(%v) for addr %v: %w", name, addr.Hex(), err)
		}
		parts := strings.Split(name, ".")
		mainName := strings.Join(parts[len(parts)-2:], ".")
		ensName, err := go_ens.NewName(erigonClient.GetNativeClient(), mainName)
		if err != nil {
			return fmt.Errorf("error could not create name via go_ens.NewName for [%v]: %w", name, err)
		}
		expires, err := ensName.Expires()
		if err != nil {
			return fmt.Errorf("error could not get ens expire date for [%v]: %w", name, err)
		}

		if dbEntry == nil || dbEntry.EnsName != name || !bytes.Equal(dbEntry.NameHash, nameHash[:]) || !bytes.Equal(dbEntry.Address, resolvedAddr.Bytes()) || dbEntry.ValidTo != expires {
			logFields := logrus.Fields{"resolvedAddr": resolvedAddr, "addr": addr.Hex(), "name": name, "nameHash": fmt.Sprintf("%#x", nameHash), "expires": expires}
			if dbEntry == nil {
				logFields["db"] = "nil"
				logrus.WithFields(logFields).Warnf("adding ens entry")
			} else {
				logFields["db.name"] = dbEntry.EnsName
				logFields["db.nameHash"] = fmt.Sprintf("%#x", dbEntry.NameHash)
				logFields["db.addr"] = fmt.Sprintf("%#x", dbEntry.Address)
				logFields["db.expire"] = dbEntry.ValidTo
				logrus.WithFields(logFields).Warnf("updating ens entry")
			}

			if !opts.DryRun {
				_, err = db.WriterDb.Exec(`
					INSERT INTO ens (
						name_hash, 
						ens_name, 
						address,
						is_primary_name, 
						valid_to)
					VALUES ($1, $2, $3, $4, $5) 
					ON CONFLICT 
						(name_hash) 
					DO UPDATE SET 
						ens_name = excluded.ens_name,
						address = excluded.address,
						is_primary_name = excluded.is_primary_name,
						valid_to = excluded.valid_to`,
					nameHash[:], name, addr.Bytes(), true, expires)
				if err != nil {
					return fmt.Errorf("error writing ens data for addr [%v]: %w", addr.Hex(), err)
				}
			}
		}
	}
	return nil
}

func fixExecTransactionsCount() error {
	startBlockNumber := uint64(opts.StartBlock)
	endBlockNumber := uint64(opts.EndBlock)

	logrus.WithFields(logrus.Fields{"startBlockNumber": startBlockNumber, "endBlockNumber": endBlockNumber}).Infof("fixExecTransactionsCount")

	batchSize := int64(1000)

	dbUpdates := []struct {
		BlockNumber  uint64
		ExecTxsCount uint64
	}{}

	for i := startBlockNumber; i <= endBlockNumber; i += uint64(batchSize) {
		firstBlock := int64(i)
		lastBlock := firstBlock + batchSize - 1
		if lastBlock > int64(endBlockNumber) {
			lastBlock = int64(endBlockNumber)
		}
		blocksChan := make(chan *types.Eth1Block, batchSize)
		go func(stream chan *types.Eth1Block) {
			high := lastBlock
			low := lastBlock - batchSize + 1
			if int64(firstBlock) > low {
				low = firstBlock
			}

			err := db.BigtableClient.GetFullBlocksDescending(stream, uint64(high), uint64(low))
			if err != nil {
				logrus.Errorf("error getting blocks descending high: %v low: %v err: %v", high, low, err)
			}
			close(stream)
		}(blocksChan)
		totalTxsCount := 0
		for b := range blocksChan {
			if len(b.Transactions) > 0 {
				totalTxsCount += len(b.Transactions)
				dbUpdates = append(dbUpdates, struct {
					BlockNumber  uint64
					ExecTxsCount uint64
				}{b.Number, uint64(len(b.Transactions))})
			}
		}
		logrus.Infof("%v-%v: totalTxsCount: %v", firstBlock, lastBlock, totalTxsCount)
	}

	logrus.Infof("dbUpdates: %v", len(dbUpdates))

	tx, err := db.WriterDb.Begin()
	if err != nil {
		return fmt.Errorf("error starting db transactions: %w", err)
	}
	defer tx.Rollback()

	for b := 0; b < len(dbUpdates); b += int(batchSize) {
		start := b
		end := b + int(batchSize)
		if len(dbUpdates) < end {
			end = len(dbUpdates)
		}

		valueStrings := []string{}
		for _, v := range dbUpdates[start:end] {
			valueStrings = append(valueStrings, fmt.Sprintf("(%v,%v)", v.BlockNumber, v.ExecTxsCount))
		}

		stmt := fmt.Sprintf(`
			update blocks as a set exec_transactions_count = b.exec_transactions_count 
			from (values %s) as b(exec_block_number, exec_transactions_count)
			where a.exec_block_number = b.exec_block_number`, strings.Join(valueStrings, ","))

		_, err = tx.Exec(stmt)
		if err != nil {
			return err
		}

		logrus.Infof("updated %v-%v / %v", start, end, len(dbUpdates))
	}

	return tx.Commit()
}

func updateBlockFinalizationSequentially() error {
	var err error

	var maxSlot uint64
	err = db.WriterDb.Get(&maxSlot, `select max(slot) from blocks`)
	if err != nil {
		return err
	}

	lookback := uint64(0)
	if maxSlot > 1e4 {
		lookback = maxSlot - 1e4
	}
	var minNonFinalizedSlot uint64
	for {
		err = db.WriterDb.Get(&minNonFinalizedSlot, `select coalesce(min(slot),0) from blocks where finalized = false and slot >= $1 and slot <= $1+1e4`, lookback)
		if err != nil {
			return err
		}
		if minNonFinalizedSlot == 0 {
			break
		}
		if minNonFinalizedSlot == lookback && lookback > 1e4 {
			lookback -= 1e4
			continue
		}
		break
	}

	logrus.WithFields(logrus.Fields{"minNonFinalizedSlot": minNonFinalizedSlot}).Infof("updateBlockFinalizationSequentially")
	nextStartEpoch := minNonFinalizedSlot / utils.Config.Chain.ClConfig.SlotsPerEpoch
	stepSize := uint64(100)
	for ; ; time.Sleep(time.Millisecond * 50) {
		t0 := time.Now()
		var finalizedEpoch uint64
		err = db.WriterDb.Get(&finalizedEpoch, `SELECT COALESCE(MAX(epoch) - 3, 0) FROM epochs WHERE finalized`)
		if err != nil {
			return err
		}
		lastEpoch := nextStartEpoch + stepSize - 1
		if lastEpoch > finalizedEpoch {
			lastEpoch = finalizedEpoch
		}
		_, err = db.WriterDb.Exec(`UPDATE blocks SET finalized = true WHERE epoch >= $1 AND epoch <= $2 AND NOT finalized;`, nextStartEpoch, lastEpoch)
		if err != nil {
			return err
		}
		secondsPerEpoch := time.Since(t0).Seconds() / float64(stepSize)
		timeLeft := time.Second * time.Duration(float64(finalizedEpoch-lastEpoch)*time.Since(t0).Seconds()/float64(stepSize))
		logrus.WithFields(logrus.Fields{"finalizedEpoch": finalizedEpoch, "epochs": fmt.Sprintf("%v-%v", nextStartEpoch, lastEpoch), "timeLeft": timeLeft, "secondsPerEpoch": secondsPerEpoch}).Infof("did set blocks to finalized")
		if finalizedEpoch <= lastEpoch {
			logrus.Infof("all relevant blocks have been set to finalized (up to epoch %v)", finalizedEpoch)
			return nil
		}
		nextStartEpoch = nextStartEpoch + stepSize
	}
}

func debugBlocks() error {
	elClient, err := rpc.NewErigonClient(utils.Config.Eth1ErigonEndpoint)
	if err != nil {
		return err
	}

	clClient, err := rpc.NewLighthouseClient(fmt.Sprintf("http://%v:%v", utils.Config.Indexer.Node.Host, utils.Config.Indexer.Node.Port), new(big.Int).SetUint64(utils.Config.Chain.ClConfig.DepositChainID))
	if err != nil {
		return err
	}

	for i := opts.StartBlock; i <= opts.EndBlock; i++ {
		btBlock, err := db.BigtableClient.GetBlockFromBlocksTable(i)
		if err != nil {
			return err
		}
		// logrus.WithFields(logrus.Fields{"block": i, "data": fmt.Sprintf("%+v", b)}).Infof("block from bt")

		elBlock, _, err := elClient.GetBlock(int64(i), "parity/geth")
		if err != nil {
			return err
		}

		slot := utils.TimeToSlot(uint64(elBlock.Time.Seconds))
		clBlock, err := clClient.GetBlockBySlot(slot)
		if err != nil {
			return err
		}
		logFields := logrus.Fields{
			"block":            i,
			"bt.hash":          fmt.Sprintf("%#x", btBlock.Hash),
			"bt.BlobGasUsed":   btBlock.BlobGasUsed,
			"bt.ExcessBlobGas": btBlock.ExcessBlobGas,
			"bt.txs":           len(btBlock.Transactions),
			"el.BlobGasUsed":   elBlock.BlobGasUsed,
			"el.hash":          fmt.Sprintf("%#x", elBlock.Hash),
			"el.ExcessBlobGas": elBlock.ExcessBlobGas,
			"el.txs":           len(elBlock.Transactions),
		}
		if !bytes.Equal(clBlock.ExecutionPayload.BlockHash, elBlock.Hash) {
			logrus.Warnf("clBlock.ExecutionPayload.BlockHash != i: %x != %x", clBlock.ExecutionPayload.BlockHash, elBlock.Hash)
		} else if clBlock.ExecutionPayload.BlockNumber != i {
			logrus.Warnf("clBlock.ExecutionPayload.BlockNumber != i: %v != %v", clBlock.ExecutionPayload.BlockNumber, i)
		} else {
			logFields["cl.txs"] = len(clBlock.ExecutionPayload.Transactions)
		}

		logrus.WithFields(logFields).Infof("debug block")

		for i := range elBlock.Transactions {
			btx := elBlock.Transactions[i]
			ctx := elBlock.Transactions[i]
			btxH := []string{}
			ctxH := []string{}
			for _, h := range btx.BlobVersionedHashes {
				btxH = append(btxH, fmt.Sprintf("%#x", h))
			}
			for _, h := range ctx.BlobVersionedHashes {
				ctxH = append(ctxH, fmt.Sprintf("%#x", h))
			}

			logrus.WithFields(logrus.Fields{
				"b.hash":                 fmt.Sprintf("%#x", btx.Hash),
				"el.hash":                fmt.Sprintf("%#x", ctx.Hash),
				"b.BlobVersionedHashes":  fmt.Sprintf("%+v", btxH),
				"el.BlobVersionedHashes": fmt.Sprintf("%+v", ctxH),
				"b.maxFeePerBlobGas":     btx.MaxFeePerBlobGas,
				"el.maxFeePerBlobGas":    ctx.MaxFeePerBlobGas,
				"b.BlobGasPrice":         btx.BlobGasPrice,
				"el.BlobGasPrice":        ctx.BlobGasPrice,
				"b.BlobGasUsed":          btx.BlobGasUsed,
				"el.BlobGasUsed":         ctx.BlobGasUsed,
			}).Infof("debug tx")
		}
	}
	return nil
}

func nameValidatorsByRanges(rangesUrl string) error {
	ranges := struct {
		Ranges map[string]string `json:"ranges"`
	}{}

	if strings.HasPrefix(rangesUrl, "http") {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		err := utils.HttpReq(ctx, http.MethodGet, rangesUrl, nil, &ranges)
		if err != nil {
			return err
		}
	} else {
		err := json.Unmarshal([]byte(rangesUrl), &ranges)
		if err != nil {
			return err
		}
	}

	for r, n := range ranges.Ranges {
		rs := strings.Split(r, "-")
		if len(rs) != 2 {
			return fmt.Errorf("invalid format, range must be X-Y")
		}
		rFrom, err := strconv.ParseUint(rs[0], 10, 64)
		if err != nil {
			return err
		}
		rTo, err := strconv.ParseUint(rs[1], 10, 64)
		if err != nil {
			return err
		}
		if rTo < rFrom {
			return fmt.Errorf("invalid format, range must be X-Y where X <= Y")
		}

		fmt.Printf("insert into validator_names(publickey, name) select pubkey as publickey, %v as name from validators where validatorindex >= %v and validatorindex <= %v on conflict(publickey) do update set name = excluded.name;\n", n, rFrom, rTo)
		_, err = db.WriterDb.Exec("insert into validator_names(publickey, name) select pubkey as publickey, $1 as name from validators where validatorindex >= $2 and validatorindex <= $3 on conflict(publickey) do update set name = excluded.name;", n, rFrom, rTo)
		if err != nil {
			return err
		}
	}

	return nil
}

// one time migration of the last attestation slot values from postgres to bigtable
// will write the last attestation slot that is currently in postgres to bigtable
// this can safely be done for active validators as bigtable will only keep the most recent
// last attestation slot
func migrateLastAttestationSlotToBigtable() {
	validators := []types.Validator{}

	err := db.WriterDb.Select(&validators, "SELECT validatorindex, lastattestationslot FROM validators WHERE lastattestationslot IS NOT NULL ORDER BY validatorindex")

	if err != nil {
		utils.LogFatal(err, "error retrieving last attestation slot", 0)
	}

	for _, validator := range validators {
		logrus.Infof("setting last attestation slot %v for validator %v", validator.LastAttestationSlot, validator.Index)

		err := db.BigtableClient.SetLastAttestationSlot(validator.Index, uint64(validator.LastAttestationSlot.Int64))
		if err != nil {
			utils.LogFatal(err, "error setting last attestation slot", 0)
		}
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

		tx, err := db.WriterDb.Beginx()
		if err != nil {
			logrus.Fatal(err)
		}
		defer tx.Rollback()

		for _, bm := range data.Blocks {
			for _, b := range bm {
				block := b
				logrus.Infof("Updating data for slot %v", block.Slot)

				if len(block.Attestations) == 0 {
					logrus.Infof("No Attestations for slot %v", block.Slot)

					g.Go(func() error {
						select {
						case <-gCtx.Done():
							return gCtx.Err()
						default:
						}

						// if we have some obsolete attestations we clean them from the db
						rows, err := tx.Exec(`
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

				status := uint64(0)
				err := tx.Get(&status, `
				SELECT status
				FROM blocks WHERE 
					slot=$1`, block.Slot)
				if err != nil {
					utils.LogError(err, fmt.Errorf("error getting Slot [%v] status", block.Slot), 0)
					return
				}
				importWholeBlock := false

				if status != block.Status {
					logrus.Infof("Slot[%v] has the wrong status [%v], but should be [%v]", block.Slot, status, block.Status)
					if block.Status == 1 {
						importWholeBlock = true
					} else {
						utils.LogError(err, fmt.Errorf("error on Slot [%v] - no update process for status [%v]", block.Slot, block.Status), 0)
						return
					}
				} else if len(block.Attestations) > 0 {
					count := 0
					err := tx.Get(&count, `
						SELECT COUNT(*)
						FROM 
							blocks_attestations 
						WHERE 
							block_slot=$1`, block.Slot)
					if err != nil {
						utils.LogError(err, fmt.Errorf("error getting Slot [%v] status", block.Slot), 0)
						return
					}
					// We only know about cases where we have no attestations in the db but the node has one.
					// So we don't handle cases (for now) where there are attestations with different sizes - that would require a different handling
					if count == 0 {
						importWholeBlock = true
					}
				}

				if importWholeBlock {
					err := db.SaveBlock(block, true, tx)
					if err != nil {
						utils.LogError(err, fmt.Errorf("error saving Slot [%v]", block.Slot), 0)
						return
					}
					continue
				}

				for i, a := range block.Attestations {
					att := a
					index := i
					g.Go(func() error {
						select {
						case <-gCtx.Done():
							return gCtx.Err()
						default:
						}
						var aggregationbits *[]byte

						// block_slot and block_index are already unique, but to be sure we use the correct index we also check the signature
						err := tx.Get(&aggregationbits, `
							SELECT aggregationbits
							FROM blocks_attestations WHERE 
								block_slot=$1 AND
								block_index=$2
						`, block.Slot, index)
						if err != nil {
							return fmt.Errorf("error getting aggregationbits on Slot [%v] Index [%v] with Sig [%v]: %v", block.Slot, index, att.Signature, err)
						}

						if !bytes.Equal(*aggregationbits, att.AggregationBits) {
							_, err = tx.Exec(`
								UPDATE blocks_attestations
								SET
									aggregationbits=$1
								WHERE
									block_slot=$2 AND
									block_index=$3
							`, att.AggregationBits, block.Slot, index)
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

		err = tx.Commit()
		if err != nil {
			utils.LogError(err, fmt.Sprintf("error committing tx for epoch [%v]", epoch), 0)
			return
		}
		logrus.Infof("Update of Epoch[%v] complete", epoch)
	}
}

// Updates a users API key
func updateAPIKey(user uint64) error {
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
func compareRewards(dayStart uint64, dayEnd uint64, validator uint64, bt *db.Bigtable) {

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
		FROM validator_stats WHERE validatorindex = $2 AND day = $1`, day, validator)
		if err != nil {
			logrus.Fatalf("error getting cl_rewards_gwei from db: %v", err)
			return
		}
		if tot != *dbRewards {
			logrus.Errorf("Rewards are not the same on day %v-> big: %v, db: %v", day, tot, *dbRewards)
		}
	}

}

func clearBigtable(table string, family string, key string, dryRun bool, bt *db.Bigtable) {

	if !dryRun {
		confirmation := utils.CmdPrompt(fmt.Sprintf("Are you sure you want to delete all big table entries starting with [%v] for family [%v]?", key, family))
		if confirmation != "yes" {
			logrus.Infof("Abort!")
			return
		}
	}

	if !strings.Contains(key, ":") {
		logrus.Fatalf("provided invalid prefix: %s", key)
	}

	// admin, err := gcp_bigtable.NewAdminClient(context.Background(), utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance)
	// if err != nil {
	// 	logrus.Fatal(err)
	// }

	// err = admin.DropRowRange(context.Background(), table, key)
	// if err != nil {
	// 	logrus.Fatal(err)
	// }
	err := bt.ClearByPrefix(table, family, key, dryRun)

	if err != nil {
		logrus.Fatalf("error deleting from bigtable: %v", err)
	}
	logrus.Info("delete completed")
}

// Goes through the tableData table and checks what blocks in the given range from [start] to [end] are missing and exports/indexes the missing ones
//
//	Both [start] and [end] are inclusive
//	Pass math.MaxInt64 as [end] to export from [start] to the last block in the blocks table
func indexMissingBlocks(start uint64, end uint64, bt *db.Bigtable, client *rpc.ErigonClient) {
	if end == math.MaxInt64 {
		lastBlockFromBlocksTable, err := bt.GetLastBlockInBlocksTable()
		if err != nil {
			logrus.Errorf("error retrieving last blocks from blocks table: %v", err)
			return
		}
		end = uint64(lastBlockFromBlocksTable)
	}

	errFields := map[string]interface{}{
		"start": start,
		"end":   end}

	batchSize := uint64(10000)
	for from := start; from <= end; from += batchSize {
		targetCount := batchSize
		if from+targetCount >= end {
			targetCount = end - from + 1
		}
		to := from + targetCount - 1

		errFields["from"] = from
		errFields["to"] = to
		errFields["targetCount"] = targetCount

		list, err := bt.GetBlocksDescending(to, targetCount)
		if err != nil {
			utils.LogError(err, "error retrieving blocks from tableData", 0, errFields)
			return
		}

		receivedLen := uint64(len(list))
		if receivedLen == targetCount {
			logrus.Infof("found all blocks [%v]->[%v], skipping batch", from, to)
			continue
		}

		logrus.Infof("%v blocks are missing from [%v]->[%v]", targetCount-receivedLen, from, to)

		blocksMap := make(map[uint64]bool)
		for _, item := range list {
			blocksMap[item.Number] = true
		}

		for block := from; block <= to; block++ {
			if blocksMap[block] {
				// block already saved, skip
				continue
			}

			logrus.Infof("block [%v] not found, will index it", block)
			if _, err := db.BigtableClient.GetBlockFromBlocksTable(block); err != nil {
				logrus.Infof("could not load [%v] from blocks table, will try to fetch it from the node and save it", block)

				bc, _, err := client.GetBlock(int64(block), "parity/geth")
				if err != nil {
					utils.LogError(err, fmt.Sprintf("error getting block %v from the node", block), 0)
					return
				}

				err = bt.SaveBlock(bc)
				if err != nil {
					utils.LogError(err, fmt.Sprintf("error saving block: %v ", block), 0)
					return
				}
			}

			indexOldEth1Blocks(block, block, 1, 1, "all", bt, client)
		}
	}
}

func indexOldEth1Blocks(startBlock uint64, endBlock uint64, batchSize uint64, concurrency uint64, transformerFlag string, bt *db.Bigtable, client *rpc.ErigonClient) {
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

	transforms := make([]func(blk *types.Eth1Block, cache *freecache.Cache) (*types.BulkMutations, *types.BulkMutations, error), 0)

	logrus.Infof("transformerFlag: %v", transformerFlag)
	transformerList := strings.Split(transformerFlag, ",")
	if transformerFlag == "all" {
		transformerList = []string{"TransformBlock", "TransformTx", "TransformBlobTx", "TransformItx", "TransformERC20", "TransformERC721", "TransformERC1155", "TransformWithdrawals", "TransformUncle", "TransformEnsNameRegistered"}
	} else if len(transformerList) == 0 {
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
		case "TransformBlobTx":
			transforms = append(transforms, bt.TransformBlobTx)
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

	to := endBlock
	if endBlock == math.MaxInt64 {
		lastBlockFromBlocksTable, err := bt.GetLastBlockInBlocksTable()
		if err != nil {
			utils.LogError(err, "error retrieving last blocks from blocks table", 0)
			return
		}

		to = uint64(lastBlockFromBlocksTable)
	}
	blockCount := utilMath.MaxU64(1, batchSize)

	logrus.Infof("Starting to index all blocks ranging from %d to %d", startBlock, to)
	for from := startBlock; from <= to; from = from + blockCount {
		toBlock := utilMath.MinU64(to, from+blockCount-1)

		logrus.Infof("indexing blocks %v to %v in data table ...", from, toBlock)
		err := bt.IndexEventsWithTransformers(int64(from), int64(toBlock), transforms, int64(concurrency), cache)
		if err != nil {
			utils.LogError(err, "error indexing from bigtable", 0)
		}
		cache.Clear()

	}

	if importENSChanges {
		if err := bt.ImportEnsUpdates(client.GetNativeClient(), math.MaxInt64); err != nil {
			utils.LogError(err, "error importing ens from events", 0)
			return
		}
	}

	logrus.Infof("index run completed")
}

func exportHistoricPrices(dayStart uint64, dayEnd uint64) {
	logrus.Infof("exporting historic prices for days %v - %v", dayStart, dayEnd)
	for day := dayStart; day <= dayEnd; day++ {
		timeStart := time.Now()
		ts := utils.DayToTime(int64(day)).UTC().Truncate(utils.Day)
		err := services.WriteHistoricPricesForDay(ts)
		if err != nil {
			errMsg := fmt.Sprintf("error exporting historic prices for day %v", day)
			utils.LogError(err, errMsg, 0)
			return
		}
		logrus.Printf("finished export for day %v, took %v", day, time.Since(timeStart))

		if day < dayEnd {
			// Wait to not overload the API
			time.Sleep(5 * time.Second)
		}
	}

	logrus.Info("historic price update run completed")
}

func exportStatsTotals(columns string, dayStart, dayEnd, concurrency uint64) {
	start := time.Now()
	exportToToday := false
	if dayEnd <= 0 {
		exportToToday = true
		dayEnd = math.MaxInt
	}

	logrus.Infof("exporting stats totals for columns '%v'", columns)

	// validate columns input
	columnsSlice := strings.Split(columns, ",")
	validColumns := []string{
		"cl_rewards_gwei_total",
		"el_rewards_wei_total",
		"mev_rewards_wei_total",
		"missed_attestations_total",
		"participated_sync_total",
		"missed_sync_total",
		"orphaned_sync_total",
		"withdrawals_total",
		"withdrawals_amount_total",
		"deposits_total",
		"deposits_amount_total",
	}

OUTER:
	for _, c := range columnsSlice {
		for _, vc := range validColumns {
			if c == vc {
				// valid column found, continue to next column from input
				continue OUTER
			}
		}
		// no valid column matched, exit with error
		utils.LogFatal(nil, "invalid column provided, please use a valid one", 0, map[string]interface{}{
			"usedColumn":   c,
			"validColumns": validColumns,
		})
	}

	// build insert query from input columns
	var totalClauses []string
	var conflictClauses []string

	for _, col := range columnsSlice {
		totalClause := fmt.Sprintf("COALESCE(vs1.%s, 0) + COALESCE(vs2.%s, 0)", strings.TrimSuffix(col, "_total"), col)
		totalClauses = append(totalClauses, totalClause)

		conflictClause := fmt.Sprintf("%s = excluded.%s", col, col)
		conflictClauses = append(conflictClauses, conflictClause)
	}

	insertQuery := fmt.Sprintf(`
		INSERT INTO validator_stats (validatorindex, day, %s)
		SELECT
			vs1.validatorindex,
			vs1.day,
			%s
		FROM validator_stats vs1
		LEFT JOIN validator_stats vs2
		ON vs2.day = vs1.day - 1 AND vs2.validatorindex = vs1.validatorindex
		WHERE vs1.day = $1 AND vs1.validatorindex >= $2 AND vs1.validatorindex <= $3
		ON CONFLICT (validatorindex, day) DO UPDATE SET %s;`,
		strings.Join(columnsSlice, ",\n\t"),
		strings.Join(totalClauses, ",\n\t\t"),
		strings.Join(conflictClauses, ",\n\t"))

	for day := dayStart; day <= dayEnd; day++ {
		timeDay := time.Now()
		logrus.Infof("exporting total sync and for columns %v for day %v", columns, day)

		// get max validator index for day
		firstEpoch, _ := utils.GetFirstAndLastEpochForDay(day + 1)
		var maxValidatorIndex uint64
		err := db.ReaderDb.Get(&maxValidatorIndex, `SELECT MAX(validatorindex) FROM validator_stats WHERE day = $1`, day)
		if err != nil {
			utils.LogFatal(err, "error: could not get max validator index", 0, map[string]interface{}{
				"epoch": firstEpoch,
			})
		} else if maxValidatorIndex == uint64(0) {
			utils.LogFatal(err, "error: no validator found", 0, map[string]interface{}{
				"epoch": firstEpoch,
			})
		}

		ctx := context.Background()
		g, gCtx := errgroup.WithContext(ctx)
		g.SetLimit(int(concurrency))

		batchSize := 1000

		// insert stats totals for each batch of validators
		for b := 0; b <= int(maxValidatorIndex); b += batchSize {
			start := b
			end := b + batchSize - 1
			if int(maxValidatorIndex) < end {
				end = int(maxValidatorIndex)
			}

			g.Go(func() error {
				select {
				case <-gCtx.Done():
					return gCtx.Err()
				default:
				}

				_, err = db.WriterDb.Exec(insertQuery, day, start, end)
				return err
			})
		}
		if err = g.Wait(); err != nil {
			utils.LogFatal(err, "error exporting stats totals", 0, map[string]interface{}{
				"day":     day,
				"columns": columns,
			})
		}

		if exportToToday {
			dayEnd, err = db.GetLastExportedStatisticDay()
			if err != nil {
				utils.LogError(err, "error getting last exported statistic day", 0)
				return
			}
		}
		logrus.Infof("finished exporting stats totals for columns '%v for day %v, took %v", columns, day, time.Since(timeDay))
	}

	logrus.Infof("finished all exporting stats totals for columns '%v' for days %v - %v, took %v", columns, dayStart, dayEnd, time.Since(start))
}

/*
Instead of deleting entries from the sync_committee table in a prod environment and wait for the exporter to sync back all entries,
this method will replace each sync committee period one by one with the new one. Which is much nicer for a prod environment.
*/
func exportSyncCommitteePeriods(rpcClient rpc.Client, startDay, endDay uint64, dryRun bool) {
	var lastEpoch = uint64(0)

	firstPeriod := utils.SyncPeriodOfEpoch(utils.Config.Chain.ClConfig.AltairForkEpoch)
	if startDay > 0 {
		firstEpoch, _ := utils.GetFirstAndLastEpochForDay(startDay)
		firstPeriod = utils.SyncPeriodOfEpoch(firstEpoch)
	}

	if endDay <= 0 {
		var err error
		lastEpoch, err = db.GetLatestFinalizedEpoch()
		if err != nil {
			utils.LogError(err, "error getting latest finalized epoch", 0)
			return
		}
		if lastEpoch > 0 { // guard against underflows
			lastEpoch = lastEpoch - 1
		}
	} else {
		_, lastEpoch = utils.GetFirstAndLastEpochForDay(endDay)
	}

	lastPeriod := utils.SyncPeriodOfEpoch(uint64(lastEpoch)) + 1 // we can look into the future

	start := time.Now()
	for p := firstPeriod; p <= lastPeriod; p++ {
		t0 := time.Now()

		err := reExportSyncCommittee(rpcClient, p, dryRun)
		if err != nil {
			if strings.Contains(err.Error(), "not found 404") {
				logrus.WithField("period", p).Infof("reached max period, stopping")
				break
			} else {
				utils.LogError(err, "error re-exporting sync_committee", 0, map[string]interface{}{
					"period": p,
				})
				return
			}
		}

		logrus.WithFields(logrus.Fields{
			"period":   p,
			"epoch":    utils.FirstEpochOfSyncPeriod(p),
			"duration": time.Since(t0),
		}).Infof("re-exported sync_committee")
	}

	logrus.Infof("finished all exporting sync_committee for periods %v - %v, took %v", firstPeriod, lastPeriod, time.Since(start))
}

func exportSyncCommitteeValidatorStats(rpcClient rpc.Client, startDay, endDay uint64, dryRun, skipPhase1 bool) {
	if endDay <= 0 {
		lastEpoch, err := db.GetLatestFinalizedEpoch()
		if err != nil {
			utils.LogError(err, "error getting latest finalized epoch", 0)
			return
		}
		if lastEpoch > 0 { // guard against underflows
			lastEpoch = lastEpoch - 1
		}

		_, err = db.GetLastExportedStatisticDay()
		if err != nil {
			logrus.Infof("skipping exporting stats, first day has not been indexed yet")
			return
		}

		epochsPerDay := utils.EpochsPerDay()
		currentDay := lastEpoch / epochsPerDay
		endDay = currentDay - 1 // current day will be picked up by exporter
	}

	start := time.Now()

	for day := startDay; day <= endDay; day++ {
		startDay := time.Now()
		err := UpdateValidatorStatisticsSyncData(day, rpcClient, dryRun)
		if err != nil {
			utils.LogError(err, fmt.Errorf("error exporting stats for day %v", day), 0)
			break
		}

		logrus.Infof("finished updating validators_stats for day %v, took %v", day, time.Since(startDay))
	}

	logrus.Infof("finished all exporting stats for days %v - %v, took %v", startDay, endDay, time.Since(start))
	logrus.Infof("REMEMBER: To execute export-stats-totals now to update the totals")
}

func UpdateValidatorStatisticsSyncData(day uint64, client rpc.Client, dryRun bool) error {
	exportStart := time.Now()
	firstEpoch, lastEpoch := utils.GetFirstAndLastEpochForDay(day)

	logrus.Infof("exporting statistics for day %v (epoch %v to %v)", day, firstEpoch, lastEpoch)

	if err := db.CheckIfDayIsFinalized(day); err != nil && !dryRun {
		return err
	}

	logrus.Infof("getting exported state for day %v", day)

	var err error
	var maxValidatorIndex uint64
	err = db.ReaderDb.Get(&maxValidatorIndex, `SELECT MAX(validatorindex) FROM validator_stats WHERE day = $1`, day)
	if err != nil {
		utils.LogFatal(err, "error: could not get max validator index", 0, map[string]interface{}{
			"epoch": firstEpoch,
		})
	} else if maxValidatorIndex == uint64(0) {
		utils.LogFatal(err, "error: no validator found", 0, map[string]interface{}{
			"epoch": firstEpoch,
		})
	}
	maxValidatorIndex += 10000 // add some buffer, exact number is not important. Should just be bigger than max validators that can join in a day

	validatorData := make([]*types.ValidatorStatsTableDbRow, 0, maxValidatorIndex)
	validatorDataMux := &sync.Mutex{}

	logrus.Infof("processing statistics for validators 0-%d", maxValidatorIndex)
	for i := uint64(0); i <= maxValidatorIndex; i++ {
		validatorData = append(validatorData, &types.ValidatorStatsTableDbRow{
			ValidatorIndex: i,
			Day:            int64(day),
		})
	}

	g := &errgroup.Group{}

	g.Go(func() error {
		if err := db.GatherValidatorSyncDutiesForDay(nil, day, validatorData, validatorDataMux); err != nil {
			return fmt.Errorf("error in GatherValidatorSyncDutiesForDay: %w", err)
		}
		return nil
	})

	err = g.Wait()
	if err != nil {
		return err
	}

	onlySyncCommitteeValidatorData := make([]*types.ValidatorStatsTableDbRow, 0, len(validatorData))
	for index := range validatorData {

		if validatorData[index].ParticipatedSync > 0 || validatorData[index].MissedSync > 0 || validatorData[index].OrphanedSync > 0 {
			onlySyncCommitteeValidatorData = append(onlySyncCommitteeValidatorData, validatorData[index])
		}
	}

	if len(onlySyncCommitteeValidatorData) == 0 {
		return nil // no sync committee yet skip
	}

	logrus.Infof("statistics data collection for day %v completed", day)

	var statisticsDataToday []*types.ValidatorStatsTableDbRow
	if dryRun {
		var err error
		statisticsDataToday, err = db.GatherStatisticsForDay(int64(day)) // convert to int64 to avoid underflows
		if err != nil {
			return fmt.Errorf("error in GatherPreviousDayStatisticsData: %w", err)
		}
	}

	tx, err := db.WriterDb.Beginx()
	if err != nil {
		return fmt.Errorf("error retrieving raw sql connection: %w", err)
	}
	defer tx.Rollback()

	logrus.Infof("updating statistics data into the validator_stats table %v | %v", len(onlySyncCommitteeValidatorData), len(validatorData))

	for _, data := range onlySyncCommitteeValidatorData {
		if dryRun {
			logrus.Infof(
				"validator %v: participated sync: %v -> %v, missed sync: %v -> %v, orphaned sync: %v -> %v",
				data.ValidatorIndex, statisticsDataToday[data.ValidatorIndex].ParticipatedSync, data.ParticipatedSync, statisticsDataToday[data.ValidatorIndex].MissedSync, data.MissedSync, statisticsDataToday[data.ValidatorIndex].OrphanedSync,
				data.OrphanedSync,
			)
		} else {
			tx.Exec(`
				UPDATE validator_stats set
				participated_sync = $1,
				missed_sync = $2,
				orphaned_sync = $3,
				WHERE day = $4 AND validatorindex = $5`,
				data.ParticipatedSync,
				data.MissedSync,
				data.OrphanedSync,
				data.Day, data.ValidatorIndex)
		}
	}

	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error during statistics data insert: %w", err)
	}

	logrus.Infof("statistics sync re-export of day %v completed, took %v", day, time.Since(exportStart))
	return nil
}

func reExportSyncCommittee(rpcClient rpc.Client, p uint64, dryRun bool) error {
	if dryRun {
		var currentData []struct {
			ValidatorIndex uint64 `db:"validatorindex"`
			CommitteeIndex uint64 `db:"committeeindex"`
		}

		err := db.WriterDb.Select(&currentData, `SELECT validatorindex, committeeindex FROM sync_committees WHERE period = $1`, p)
		if err != nil {
			return errors.Wrap(err, "select old entries")
		}

		newData, err := exporter.GetSyncCommitteAtPeriod(rpcClient, p)
		if err != nil {
			return errors.Wrap(err, "export")
		}

		// now we compare currentData with newData and print any difference in committeeindex
		for _, d := range currentData {
			for _, n := range newData {
				if d.ValidatorIndex == n.ValidatorIndex && d.CommitteeIndex != n.CommitteeIndex {
					logrus.Infof("validator %v has different committeeindex: %v -> %v", d.ValidatorIndex, d.CommitteeIndex, n.CommitteeIndex)
				}
			}
		}
		return nil
	} else {
		tx, err := db.WriterDb.Beginx()
		if err != nil {
			return errors.Wrap(err, "tx")
		}

		defer tx.Rollback()
		_, err = tx.Exec(`DELETE FROM sync_committees WHERE period = $1`, p)
		if err != nil {
			return errors.Wrap(err, "delete old entries")
		}

		err = exporter.ExportSyncCommitteeAtPeriod(rpcClient, p, tx)
		if err != nil {
			return errors.Wrap(err, "export")
		}

		return tx.Commit()
	}
}
