package db

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"

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

func (bigtable *Bigtable) ClearByPrefix(table string, family, columns, prefix string, dryRun bool) error {
	if family == "" || prefix == "" || columns == "" {
		return fmt.Errorf("please provide family [%v], columns [%v] and prefix [%v]", family, columns, prefix)
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

	var filter gcp_bigtable.Filter
	columnsSlice := strings.Split(columns, ",")
	if len(columnsSlice) > 1 {
		columnNames := make([]gcp_bigtable.Filter, len(columnsSlice))
		for i, f := range columnsSlice {
			columnNames[i] = gcp_bigtable.ColumnFilter(f)
		}
		filter = gcp_bigtable.InterleaveFilters(columnNames...)
	} else {
		filter = gcp_bigtable.ColumnFilter(columnsSlice[0])
	}

	keysCount := 0
	deleteFunc := func(row gcp_bigtable.Row) bool {
		var row_ string

		if family == "*" {
			row_ = row.Key()
		} else {
			row_ = row[family][0].Row
		}
		if dryRun {
			logger.Infof("would delete key %v", row_)
		}

		mutDelete := gcp_bigtable.NewMutation()
		if columns == "*" {
			mutDelete.DeleteRow()
		} else {
			for _, column := range columnsSlice {
				mutDelete.DeleteCellsInColumn(family, column)
			}
		}

		mutsDelete.Keys = append(mutsDelete.Keys, row_)
		mutsDelete.Muts = append(mutsDelete.Muts, mutDelete)
		keysCount++

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
	}
	var err error
	if columns == "*" {
		err = btTable.ReadRows(context.Background(), rowRange, deleteFunc)
	} else {
		err = btTable.ReadRows(context.Background(), rowRange, deleteFunc, gcp_bigtable.RowFilter(filter))
	}
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

func (bigtable *Bigtable) ClearByRowRange(table string, family, columns string, rowRange gcp_bigtable.RowRangeList, dryRun bool) error {
	if family == "" || len(rowRange) == 0 || columns == "" {
		return fmt.Errorf("please provide family [%v], columns [%v] and prefix [%v]", family, columns, rowRange)
	}

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

	var filter gcp_bigtable.Filter
	columnsSlice := strings.Split(columns, ",")
	if len(columnsSlice) > 1 {
		columnNames := make([]gcp_bigtable.Filter, len(columnsSlice))
		for i, f := range columnsSlice {
			columnNames[i] = gcp_bigtable.ColumnFilter(f)
		}
		filter = gcp_bigtable.InterleaveFilters(columnNames...)
	} else {
		filter = gcp_bigtable.ColumnFilter(columnsSlice[0])
	}

	keysCount := 0
	deleteFunc := func(row gcp_bigtable.Row) bool {
		var row_ string

		if family == "*" {
			row_ = row.Key()
		} else {
			row_ = row[family][0].Row
		}
		if dryRun {
			logger.Infof("would delete key %v", row_)
		}

		mutDelete := gcp_bigtable.NewMutation()
		if columns == "*" {
			mutDelete.DeleteRow()
		} else {
			for _, column := range columnsSlice {
				mutDelete.DeleteCellsInColumn(family, column)
			}
		}

		mutsDelete.Keys = append(mutsDelete.Keys, row_)
		mutsDelete.Muts = append(mutsDelete.Muts, mutDelete)
		keysCount++

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
	}
	var err error
	if columns == "*" {
		err = btTable.ReadRows(context.Background(), rowRange, deleteFunc)
	} else {
		err = btTable.ReadRows(context.Background(), rowRange, deleteFunc, gcp_bigtable.RowFilter(filter))
	}
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
