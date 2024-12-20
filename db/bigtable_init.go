package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
	"github.com/sirupsen/logrus"

	gcp_bigtable "cloud.google.com/go/bigtable"
)

func InitBigtableSchema() error {

	tables := make(map[string]map[string]gcp_bigtable.GCPolicy)

	tables["beaconchain_validators"] = map[string]gcp_bigtable.GCPolicy{
		ATTESTATIONS_FAMILY: gcp_bigtable.MaxVersionsGCPolicy(1),
	}
	tables["beaconchain_validators_history"] = map[string]gcp_bigtable.GCPolicy{
		VALIDATOR_BALANCES_FAMILY:             nil,
		VALIDATOR_HIGHEST_ACTIVE_INDEX_FAMILY: nil,
		ATTESTATIONS_FAMILY:                   nil,
		SYNC_COMMITTEES_FAMILY:                nil,
		INCOME_DETAILS_COLUMN_FAMILY:          nil,
		STATS_COLUMN_FAMILY:                   nil,
	}
	tables["blocks"] = map[string]gcp_bigtable.GCPolicy{
		DEFAULT_FAMILY_BLOCKS: gcp_bigtable.MaxVersionsGCPolicy(1),
	}
	tables["data"] = map[string]gcp_bigtable.GCPolicy{
		CONTRACT_METADATA_FAMILY: gcp_bigtable.MaxAgeGCPolicy(utils.Day),
		DEFAULT_FAMILY:           nil,
	}
	tables["machine_metrics"] = map[string]gcp_bigtable.GCPolicy{
		MACHINE_METRICS_COLUMN_FAMILY: gcp_bigtable.MaxAgeGCPolicy(utils.Day * 31),
	}
	tables["metadata"] = map[string]gcp_bigtable.GCPolicy{
		ACCOUNT_METADATA_FAMILY:  nil,
		CONTRACT_METADATA_FAMILY: nil,
		ERC20_METADATA_FAMILY:    nil,
		ERC721_METADATA_FAMILY:   nil,
		ERC1155_METADATA_FAMILY:  nil,
		SERIES_FAMILY:            gcp_bigtable.MaxVersionsGCPolicy(1),
	}
	tables["metadata_updates"] = map[string]gcp_bigtable.GCPolicy{
		METADATA_UPDATES_FAMILY_BLOCKS: gcp_bigtable.MaxAgeGCPolicy(utils.Day),
		DEFAULT_FAMILY:                 nil,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	if utils.Config.Bigtable.Emulator {
		if utils.Config.Bigtable.EmulatorHost == "" {
			utils.Config.Bigtable.EmulatorHost = "127.0.0.1"
		}
		logrus.Infof("using emulated local bigtable environment, setting BIGTABLE_EMULATOR_HOST env variable to %s:%d", utils.Config.Bigtable.EmulatorHost, utils.Config.Bigtable.EmulatorPort)
		err := os.Setenv("BIGTABLE_EMULATOR_HOST", fmt.Sprintf("%s:%d", utils.Config.Bigtable.EmulatorHost, utils.Config.Bigtable.EmulatorPort))

		if err != nil {
			logrus.Fatal(err, "unable to set bigtable emulator environment variable", 0)
		}
	}

	admin, err := gcp_bigtable.NewAdminClient(ctx, utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance)
	if err != nil {
		return err
	}

	existingTables, err := admin.Tables(ctx)
	if err != nil {
		return err
	}

	if len(existingTables) > 0 {
		return fmt.Errorf("aborting bigtable schema init as tables are already present")
	}

	for name, definition := range tables {
		err := admin.CreateTable(ctx, name)
		if err != nil {
			return err
		}

		for columnFamily, gcPolicy := range definition {
			err := admin.CreateColumnFamily(ctx, name, columnFamily)
			if err != nil {
				return err
			}

			if gcPolicy != nil {
				err := admin.SetGCPolicy(ctx, name, columnFamily, gcPolicy)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
