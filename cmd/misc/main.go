package main

import (
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

	_ "github.com/jackc/pgx/v4/stdlib"

	"flag"

	"github.com/sirupsen/logrus"
)

var opts = struct {
	Command       string
	User          uint64
	TargetVersion int64
	StartEpoch    uint64
	EndEpoch      uint64
	StartDay      uint64
	EndDay        uint64
	Validator     uint64
	Validators    string
	Address       string
	Family        string
	Key           string
	DryRun        bool
}{}

func main() {
	configPath := flag.String("config", "config/default.config.yml", "Path to the config file")
	flag.StringVar(&opts.Command, "command", "", "command to run, available: updateAPIKey, applyDbSchema, epoch-export, debug-rewards, clear-bigtable")
	flag.Uint64Var(&opts.StartEpoch, "start-epoch", 0, "start epoch")
	flag.Uint64Var(&opts.EndEpoch, "end-epoch", 0, "end epoch")
	flag.Uint64Var(&opts.User, "user", 0, "user id")
	flag.Uint64Var(&opts.StartDay, "day-start", 0, "start day to debug")
	flag.Uint64Var(&opts.EndDay, "day-end", 0, "end day to debug")
	flag.Uint64Var(&opts.Validator, "validator", 0, "validator to check for")
	flag.StringVar(&opts.Validators, "validators", "", "comma separated list of validators (pubkey of index)")
	flag.Int64Var(&opts.TargetVersion, "target-version", -2, "Db migration target version, use -2 to apply up to the latest version, -1 to apply only the next version or the specific versions")
	flag.StringVar(&opts.Address, "address", "0xf00", "address")
	flag.StringVar(&opts.Family, "family", "", "big table family")
	flag.StringVar(&opts.Key, "key", "", "big table key")
	dryRun := flag.String("dry-run", "true", "if 'false' it deletes all rows starting with the key, per default it only logs the rows that would be deleted, but does not really delete them")
	flag.Parse()

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
	db.MustInitFrontendDB(&types.DatabaseConfig{
		Username: cfg.Frontend.WriterDatabase.Username,
		Password: cfg.Frontend.WriterDatabase.Password,
		Name:     cfg.Frontend.WriterDatabase.Name,
		Host:     cfg.Frontend.WriterDatabase.Host,
		Port:     cfg.Frontend.WriterDatabase.Port,
	}, &types.DatabaseConfig{
		Username: cfg.Frontend.ReaderDatabase.Username,
		Password: cfg.Frontend.ReaderDatabase.Password,
		Name:     cfg.Frontend.ReaderDatabase.Name,
		Host:     cfg.Frontend.ReaderDatabase.Host,
		Port:     cfg.Frontend.ReaderDatabase.Port,
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
	case "debug-balances":
		DebugBalances(opts.StartDay, opts.EndDay, opts.Validator)
	case "get-token-metadata":
		err = GetTokenMetadata()
		if err != nil {
			logrus.Fatal(err)
		}
		CompareRewards(opts.StartDay, opts.EndDay, opts.Validator, bt)
	case "clear-bigtable":
		ClearBigtable(opts.Family, opts.Key, opts.DryRun, bt)

	default:
		utils.LogFatal(nil, "unknown command", 0)
	}
}

func GetTokenMetadata() error {
	gethClient, err := rpc.NewErigonClient(utils.Config.Eth1GethEndpoint)
	if err != nil {
		return err
	}
	d, err := gethClient.GetERC20TokenMetadata(utils.MustParseHex(opts.Address))
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", d)
	return nil
}

func DebugBalances(startDay, endDay, validator uint64) {
	bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.Config.DepositChainID), utils.Config.RedisCacheEndpoint)
	if err != nil {
		logrus.Fatalf("error connecting to bigtable: %v", err)
	}
	defer bt.Close()

	logrus.Infof("--- bal history")
	balHist, err := bt.GetValidatorBalanceHistory([]uint64{validator}, opts.StartEpoch, opts.EndEpoch)
	if err != nil {
		logrus.Fatal(err)
	}
	for _, v := range balHist {
		for _, b := range v {
			fmt.Printf("%+v\n", b)
		}
	}

	logrus.Infof("--- rew history")
	rewHist, err := bt.GetValidatorIncomeDetailsHistory([]uint64{validator}, opts.StartEpoch, opts.EndEpoch)
	if err != nil {
		logrus.Fatal(err)
	}
	for _, v := range rewHist {
		for _, b := range v {
			fmt.Printf("%+v\n", b)
		}
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
