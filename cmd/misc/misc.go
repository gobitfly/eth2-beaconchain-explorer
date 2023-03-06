package main

import (
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"fmt"

	_ "github.com/jackc/pgx/v4/stdlib"

	"flag"

	"github.com/sirupsen/logrus"
)

var opts = struct {
	Command string
	User    uint64
}{}

func main() {
	configPath := flag.String("config", "config/default.config.yml", "Path to the config file")
	flag.StringVar(&opts.Command, "command", "", "command to run, available: updateAPIKey")
	flag.Uint64Var(&opts.User, "user", 0, "user id")
	flag.Parse()

	logrus.WithField("config", *configPath).WithField("version", version.Version).Printf("starting")
	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg

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
	case "checkTransactions":

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
