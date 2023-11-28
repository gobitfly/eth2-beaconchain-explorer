package db

import (
	"context"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"sort"
	"time"

	gcp_bigtable "cloud.google.com/go/bigtable"
	"github.com/sirupsen/logrus"
)

func (bigtable *Bigtable) WriteBulk(mutations *types.BulkMutations, table *gcp_bigtable.Table, batchSize int) error {

	callingFunctionName := utils.GetParentFuncName()

	ctx, done := context.WithTimeout(context.Background(), time.Minute*5)
	defer done()

	numMutations := len(mutations.Muts)
	numKeys := len(mutations.Keys)
	if numKeys != numMutations {
		return fmt.Errorf("error expected same number of keys as mutations keys: %v mutations: %v", numKeys, numMutations)
	}

	// pre-sort mutations for efficient bulk inserts
	sort.Sort(mutations)

	length := batchSize
	if length > MAX_BATCH_MUTATIONS {
		logger.Infof("WriteBulk: capping provided batchSize %v to %v", length, MAX_BATCH_MUTATIONS)
		length = MAX_BATCH_MUTATIONS
	}

	iterations := numKeys / length

	for offset := 0; offset < iterations; offset++ {
		start := offset * length
		end := offset*length + length

		startTime := time.Now()
		errs, err := table.ApplyBulk(ctx, mutations.Keys[start:end], mutations.Muts[start:end])
		for _, e := range errs {
			if e != nil {
				return e
			}
		}
		if err != nil {
			return err
		}
		logger.Infof("%s: wrote from %v to %v rows to bigtable in %.1f s", callingFunctionName, start, end, time.Since(startTime).Seconds())

	}

	if (iterations * length) < numKeys {
		start := iterations * length
		startTime := time.Now()
		errs, err := table.ApplyBulk(ctx, mutations.Keys[start:], mutations.Muts[start:])
		if err != nil {
			return err
		}
		for _, e := range errs {
			if e != nil {
				return e
			}
		}
		logger.Infof("%s: wrote from %v to %v rows to bigtable in %.1fs", callingFunctionName, start, numKeys, time.Since(startTime).Seconds())

		return nil
	}

	return nil
}

func (bigtable *Bigtable) ClearByPrefix(table string, family, prefix string, dryRun bool) error {
	if family == "" || prefix == "" {
		return fmt.Errorf("please provide family [%v] and prefix [%v]", family, prefix)
	}

	rowRange := gcp_bigtable.PrefixRange(prefix)

	var btTable *gcp_bigtable.Table

	switch table {
	case "data":
		btTable = bigtable.tableData
	case "blocks":
		btTable = bigtable.tableBlocks
	case "metadata_updates":
		btTable = bigtable.tableMetadataUpdates
	case "metadata":
		btTable = bigtable.tableMetadata
	case "beaconchain":
		btTable = bigtable.tableBeaconchain
	case "machine_metrics":
		btTable = bigtable.tableMachineMetrics
	case "beaconchain_validators":
		btTable = bigtable.tableValidators
	case "beaconchain_validators_history":
		btTable = bigtable.tableValidatorsHistory
	default:
		return fmt.Errorf("unknown table %v provided", table)
	}

	mutsDelete := types.NewBulkMutations(MAX_BATCH_MUTATIONS)

	keysCount := 0
	err := btTable.ReadRows(context.Background(), rowRange, func(row gcp_bigtable.Row) bool {

		if family == "*" {
			if dryRun {
				logger.Infof("would delete key %v", row.Key())
			}

			mutDelete := gcp_bigtable.NewMutation()
			mutDelete.DeleteRow()
			mutsDelete.Keys = append(mutsDelete.Keys, row.Key())
			mutsDelete.Muts = append(mutsDelete.Muts, mutDelete)
			keysCount++
		} else {
			row_ := row[family][0]
			if dryRun {
				logger.Infof("would delete key %v", row_.Row)
			}

			mutDelete := gcp_bigtable.NewMutation()
			mutDelete.DeleteRow()
			mutsDelete.Keys = append(mutsDelete.Keys, row_.Row)
			mutsDelete.Muts = append(mutsDelete.Muts, mutDelete)
			keysCount++
		}

		// we still need to commit in batches here (instead of just calling WriteBulk only once) as loading all keys to be deleted in memory first is not feasible as the delete function could be used to delete millions of rows
		if mutsDelete.Len() == MAX_BATCH_MUTATIONS {
			logrus.Infof("deleting %v keys (first key %v, last key %v)", len(mutsDelete.Keys), mutsDelete.Keys[0], mutsDelete.Keys[len(mutsDelete.Keys)-1])
			if !dryRun {
				err := bigtable.WriteBulk(mutsDelete, btTable, DEFAULT_BATCH_INSERTS)

				if err != nil {
					logger.Errorf("error writing bulk mutations: %v", err)
					return false
				}
			}
			mutsDelete = types.NewBulkMutations(MAX_BATCH_MUTATIONS)
		}
		return true
	})
	if err != nil {
		return err
	}

	if !dryRun && mutsDelete.Len() > 0 {
		logger.Infof("deleting %v keys (first key %v, last key %v)", len(mutsDelete.Keys), mutsDelete.Keys[0], mutsDelete.Keys[len(mutsDelete.Keys)-1])

		err := bigtable.WriteBulk(mutsDelete, btTable, DEFAULT_BATCH_INSERTS)

		if err != nil {
			return err
		}
	}

	logger.Infof("deleted %v keys", keysCount)

	return nil
}
