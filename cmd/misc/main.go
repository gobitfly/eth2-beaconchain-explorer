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
	"strings"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"

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
}{}

func main() {
	configPath := flag.String("config", "config/default.config.yml", "Path to the config file")
	flag.StringVar(&opts.Command, "command", "", "command to run, available: updateAPIKey, applyDbSchema, epoch-export, debug-rewards")
	flag.Uint64Var(&opts.StartEpoch, "start-epoch", 0, "start epoch")
	flag.Uint64Var(&opts.EndEpoch, "end-epoch", 0, "end epoch")
	flag.Uint64Var(&opts.User, "user", 0, "user id")
	flag.Uint64Var(&opts.StartDay, "day-start", 0, "start day to debug")
	flag.Uint64Var(&opts.EndDay, "day-end", 0, "end day to debug")
	flag.Uint64Var(&opts.Validator, "validator", 0, "validator to check for")
	flag.Int64Var(&opts.TargetVersion, "target-version", -2, "Db migration target version, use -2 to apply up to the latest version, -1 to apply only the next version or the specific versions")
	flag.Parse()

	logrus.WithField("config", *configPath).WithField("version", version.Version).Printf("starting")
	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg

	chainIdString := strconv.FormatUint(utils.Config.Chain.Config.DepositChainID, 10)

	_, err = db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Project, chainIdString, utils.Config.RedisCacheEndpoint)
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
		CompareRewards(opts.StartDay, opts.EndDay, opts.Validator)
	case "update-orphaned-statistics":
		UpdateOrphanedStatistics(opts.StartDay, opts.EndDay, db.WriterDb)

	default:
		utils.LogFatal(nil, "unknown command", 0)
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
func CompareRewards(dayStart uint64, dayEnd uint64, validator uint64) {
	bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.Config.DepositChainID), utils.Config.RedisCacheEndpoint)

	if err != nil {
		logrus.Fatalf("error connecting to bigtable: %v", err)
	}
	defer bt.Close()

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

// Update Orphaned statistics for Sync / Attestations
func UpdateOrphanedStatistics(dayStart uint64, dayEnd uint64, WriterDb *sqlx.DB) {
	bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.Config.DepositChainID), utils.Config.RedisCacheEndpoint)
	if err != nil {
		logrus.Fatalf("error connecting to bigtable: %v", err)
	}
	defer bt.Close()

	for day := dayStart; day <= dayEnd; day++ {
		tx, err := WriterDb.Beginx()
		if err != nil {
			logrus.Errorf("error WriterDb.Beginx %v", err)
			return
		}
		defer tx.Rollback()
		startEpoch := day * utils.EpochsPerDay()
		endEpoch := startEpoch + utils.EpochsPerDay() - 1

		if err != nil {
			logrus.Errorf("error getting validator Count %v", err)
			return
		}

		logrus.Infof("exporting failed attestations statistics lastEpoch: %v firstEpoch: %v", startEpoch, endEpoch)
		ma, err := bt.GetValidatorFailedAttestationsCount([]uint64{}, startEpoch, endEpoch)
		if err != nil {
			logrus.Errorf("error getting failed attestations %v", err)
			return
		}
		maArr := make([]*types.ValidatorFailedAttestationsStatistic, 0, len(ma))
		for _, stat := range ma {
			maArr = append(maArr, stat)
		}

		batchSize := 16000 // max parameters: 65535
		for b := 0; b < len(maArr); b += batchSize {
			start := b
			end := b + batchSize
			if len(maArr) < end {
				end = len(maArr)
			}

			numArgs := 4
			valueStrings := make([]string, 0, batchSize)
			valueArgs := make([]interface{}, 0, batchSize*numArgs)
			for i, stat := range maArr[start:end] {
				valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", i*numArgs+1, i*numArgs+2, i*numArgs+3, i*numArgs+4))
				valueArgs = append(valueArgs, stat.Index)
				valueArgs = append(valueArgs, day)
				valueArgs = append(valueArgs, stat.MissedAttestations)
				valueArgs = append(valueArgs, stat.OrphanedAttestations)
			}
			stmt := fmt.Sprintf(`
			insert into validator_stats (validatorindex, day, missed_attestations, orphaned_attestations) VALUES
			%s
			on conflict (validatorindex, day) do update set missed_attestations = excluded.missed_attestations, orphaned_attestations = excluded.orphaned_attestations;`,
				strings.Join(valueStrings, ","))
			_, err := tx.Exec(stmt, valueArgs...)
			if err != nil {
				logrus.Errorf("Error inserting failed attestations %v", err)
				return
			}

			logrus.Infof("saving failed attestations batch %v completed", b)
		}

		logrus.Infof("Update Orphaned for day [%v] epoch %v -> %v", day, startEpoch, endEpoch)
		syncStats, err := bt.GetValidatorSyncDutiesStatistics([]uint64{}, startEpoch, endEpoch)
		if err != nil {
			logrus.Errorf("error getting GetValidatorSyncDutiesStatistics %v", err)
			return
		}

		syncStatsArr := make([]*types.ValidatorSyncDutiesStatistic, 0, len(syncStats))
		for _, stat := range syncStats {
			syncStatsArr = append(syncStatsArr, stat)
		}

		batchSize = 13000 // max parameters: 65535
		for b := 0; b < len(syncStatsArr); b += batchSize {
			start := b
			end := b + batchSize
			if len(syncStatsArr) < end {
				end = len(syncStatsArr)
			}

			numArgs := 5
			valueStrings := make([]string, 0, batchSize)
			valueArgs := make([]interface{}, 0, batchSize*numArgs)
			for i, stat := range syncStatsArr[start:end] {
				valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", i*numArgs+1, i*numArgs+2, i*numArgs+3, i*numArgs+4, i*numArgs+5))
				valueArgs = append(valueArgs, stat.Index)
				valueArgs = append(valueArgs, day)
				valueArgs = append(valueArgs, stat.ParticipatedSync)
				valueArgs = append(valueArgs, stat.MissedSync)
				valueArgs = append(valueArgs, stat.OrphanedSync)
			}
			stmt := fmt.Sprintf(`
				insert into validator_stats (validatorindex, day, participated_sync, missed_sync, orphaned_sync)  VALUES
				%s
				on conflict (validatorindex, day) do update set participated_sync = excluded.participated_sync, missed_sync = excluded.missed_sync, orphaned_sync = excluded.orphaned_sync;`,
				strings.Join(valueStrings, ","))
			_, err := tx.Exec(stmt, valueArgs...)
			if err != nil {
				logrus.Errorf("error inserting into validator_stats %v", err)
				return
			}

			logrus.Infof("saving sync statistics batch %v completed", b)
		}

		err = tx.Commit()
		if err != nil {
			logrus.Errorf("error commiting tx for validator_stats %v", err)
			return
		}
		logrus.Infof("Update Orphaned for day [%v] completed", day)
	}

}
