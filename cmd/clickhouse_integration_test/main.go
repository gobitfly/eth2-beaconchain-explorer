package main

import (
	"flag"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"

	itypes "github.com/gobitfly/eth-rewards/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
	"github.com/gobitfly/eth2-beaconchain-explorer/version"
	"github.com/sirupsen/logrus"
)

func main() {
	configPath := flag.String("config", "config/default.config.yml", "Path to the config file")

	flag.Parse()

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg
	logrus.WithFields(logrus.Fields{
		"config":    *configPath,
		"version":   version.Version,
		"chainName": utils.Config.Chain.ClConfig.ConfigName}).Printf("starting")

	db.MustInitDB(&types.DatabaseConfig{
		Username:     cfg.WriterDatabase.Username,
		Password:     cfg.WriterDatabase.Password,
		Name:         cfg.WriterDatabase.Name,
		Host:         cfg.WriterDatabase.Host,
		Port:         cfg.WriterDatabase.Port,
		MaxOpenConns: cfg.WriterDatabase.MaxOpenConns,
		MaxIdleConns: cfg.WriterDatabase.MaxIdleConns,
		SSL:          cfg.WriterDatabase.SSL,
	}, &types.DatabaseConfig{
		Username:     cfg.ReaderDatabase.Username,
		Password:     cfg.ReaderDatabase.Password,
		Name:         cfg.ReaderDatabase.Name,
		Host:         cfg.ReaderDatabase.Host,
		Port:         cfg.ReaderDatabase.Port,
		MaxOpenConns: cfg.ReaderDatabase.MaxOpenConns,
		MaxIdleConns: cfg.ReaderDatabase.MaxIdleConns,
		SSL:          cfg.ReaderDatabase.SSL,
	}, "pgx", "postgres")

	db.MustInitClickhouseDB(nil, &types.DatabaseConfig{
		Username:     cfg.ClickHouse.ReaderDatabase.Username,
		Password:     cfg.ClickHouse.ReaderDatabase.Password,
		Name:         cfg.ClickHouse.ReaderDatabase.Name,
		Host:         cfg.ClickHouse.ReaderDatabase.Host,
		Port:         cfg.ClickHouse.ReaderDatabase.Port,
		MaxOpenConns: cfg.ClickHouse.ReaderDatabase.MaxOpenConns,
		MaxIdleConns: cfg.ClickHouse.ReaderDatabase.MaxIdleConns,
		SSL:          true,
	}, "clickhouse", "clickhouse")

	bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.ClConfig.DepositChainID), utils.Config.RedisCacheEndpoint)
	if err != nil {
		logrus.Fatalf("error connecting to bigtable: %v", err)
	}
	db.BigtableClient = bt

	utils.Config.ClickhouseDelay = 0

	// verification funcs are to be run against mainnet
	// normal attestation
	logrus.Infof("verifying history for normal attestations")
	verifyHistory([]uint64{653161}, uint64(323260), uint64(323270), true, true)

	// income details for sync rewards (this currently fails !!!)
	logrus.Infof("verifying history for sync rewards")
	verifyHistory([]uint64{653162}, uint64(323260), uint64(323261), true, true)

	// block proposed
	logrus.Infof("verifying history for proposer rewards at end of an epoch")
	verifyHistory([]uint64{388033}, uint64(323268), uint64(323270), true, true)

	logrus.Infof("verifying history for proposer rewards at start of an epoch")
	verifyHistory([]uint64{1284148}, uint64(323268), uint64(323270), true, true)

	logrus.Infof("verifying history for proposer rewards at during an epoch")
	verifyHistory([]uint64{1208852}, uint64(323268), uint64(323270), true, true)

	// missed attestations
	logrus.Infof("verifying history for missed attestations")
	verifyHistory([]uint64{76040}, uint64(323260), uint64(323270), true, true)

	// missed slots
	logrus.Infof("verifying history for missed slots")
	verifyHistory([]uint64{858473}, uint64(323250), uint64(323260), true, true)

	// slashing (5902 was slashed by 792015)
	logrus.Infof("verifying history for a slashed validator")
	verifyHistory([]uint64{5902}, uint64(314635), uint64(314645), true, true)
	logrus.Infof("verifying history for a slashing validator")
	verifyHistory([]uint64{792015}, uint64(314635), uint64(314645), true, true)

	// validator during activation
	logrus.Infof("verifying history for a validator during activation")
	verifyHistory([]uint64{894572}, uint64(266960), uint64(266970), true, true)

	// validator during exit
	logrus.Infof("verifying history for a validator during exit")
	verifyHistory([]uint64{1646687}, uint64(323090), uint64(323110), true, true)

}

func verifyHistory(validatorIndices []uint64, epochStart, epochEnd uint64, income, balance bool) {
	slotsPerEpoch := utils.Config.Chain.ClConfig.SlotsPerEpoch
	logrus.Infof("verifying history for validator indices %v from epoch %d to %d", validatorIndices, epochStart, epochEnd)
	if income {
		compare(db.BigtableClient.GetValidatorIncomeDetailsHistory, validatorIndices, epochStart, epochEnd)
	}
	if balance {
		compare(db.BigtableClient.GetValidatorBalanceHistory, validatorIndices, epochStart, epochEnd)
	}
	compare(db.BigtableClient.GetValidatorAttestationHistory, validatorIndices, epochStart, epochEnd)
	compare(db.BigtableClient.GetValidatorMissedAttestationHistory, validatorIndices, epochStart, epochEnd)
	compare(db.BigtableClient.GetValidatorSyncDutiesHistory, validatorIndices, epochStart*slotsPerEpoch, epochEnd*slotsPerEpoch)
}

type GenericFunc[T any] func([]uint64, uint64, uint64) (T, error)

func compare[T any](compareFunc GenericFunc[T], validatorIndices []uint64, epochStart, epochEnd uint64) {
	utils.Config.ClickHouseEnabled = false
	bigtableData, err := compareFunc(validatorIndices, epochStart, epochEnd)
	if err != nil {
		logrus.Fatalf("error getting validator income details history from bigtable: %v", err)
	}
	utils.Config.ClickHouseEnabled = true
	clickhouseData, err := compareFunc(validatorIndices, epochStart, epochEnd)
	if err != nil {
		logrus.Fatalf("error getting validator income details history from clickhouse: %v", err)
	}
	diff := cmp.Diff(bigtableData, clickhouseData, cmpopts.IgnoreUnexported(itypes.ValidatorEpochIncome{}))
	if diff != "" {
		logrus.Infof("bigtable")
		spew.Dump(bigtableData)
		logrus.Infof("clickhouse")
		spew.Dump(clickhouseData)
		logrus.Info(diff)
		logrus.Fatalf("bigtable and clickhouse data do not match")
	}
}
