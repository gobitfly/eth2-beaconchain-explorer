package db

import (
	"context"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"sort"
	"strings"
	"time"

	gcp_bigtable "cloud.google.com/go/bigtable"
)

func (bigtable *Bigtable) WriteBulk(mutations *types.BulkMutations, table *gcp_bigtable.Table, batchSize int) error {

	callingFunctionName := utils.GetParentFuncName()

	ctx, done := context.WithTimeout(context.Background(), time.Minute*5)
	defer done()

	// pre-sort mutations for efficient bulk inserts
	sort.Sort(mutations)
	length := batchSize
	numMutations := len(mutations.Muts)
	numKeys := len(mutations.Keys)
	iterations := numKeys / length

	if numKeys != numMutations {
		return fmt.Errorf("error expected same number of keys as mutations keys: %v mutations: %v", numKeys, numMutations)
	}

	for offset := 0; offset < iterations; offset++ {
		start := offset * length
		end := offset*length + length

		startTime := time.Now()
		errs, err := table.ApplyBulk(ctx, mutations.Keys[start:end], mutations.Muts[start:end])
		for _, e := range errs {
			if e != nil {
				return err
			}
		}
		logger.Infof("%s: wrote from %v to %v rows to bigtable in %.1f s", callingFunctionName, start, end, time.Since(startTime).Seconds())
		if err != nil {
			return err
		}
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
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}

func (bigtable *Bigtable) DeleteRowsWithPrefix(prefix string) {

	for {
		ctx, done := context.WithTimeout(context.Background(), time.Second*30)
		defer done()

		rr := gcp_bigtable.InfiniteRange(prefix)

		rowsToDelete := make([]string, 0, 10000)
		err := bigtable.tableData.ReadRows(ctx, rr, func(r gcp_bigtable.Row) bool {
			rowsToDelete = append(rowsToDelete, r.Key())
			return true
		})
		if err != nil {
			logger.WithError(err).WithField("prefix", prefix).Errorf("error reading rows in bigtable_eth1 / DeleteRowsWithPrefix")
		}
		mut := gcp_bigtable.NewMutation()
		mut.DeleteRow()

		muts := make([]*gcp_bigtable.Mutation, 0)
		for j := 0; j < 10000; j++ {
			muts = append(muts, mut)
		}

		l := len(rowsToDelete)
		if l == 0 {
			logger.Infof("all done")
			break
		}
		logger.Infof("deleting %v rows", l)

		for i := 0; i < l; i++ {
			if !strings.HasPrefix(rowsToDelete[i], "1:t:") {
				logger.Infof("wrong prefix: %v", rowsToDelete[i])
			}
			ctx, done := context.WithTimeout(context.Background(), time.Second*30)
			defer done()
			if i%10000 == 0 && i != 0 {
				logger.Infof("deleting rows: %v to %v", i-10000, i)
				errs, err := bigtable.tableData.ApplyBulk(ctx, rowsToDelete[i-10000:i], muts)
				if err != nil {
					logger.WithError(err).Errorf("error deleting row: %v", rowsToDelete[i])
				}
				for _, err := range errs {
					utils.LogError(err, fmt.Errorf("bigtable apply bulk error, deleting rows: %v to %v", i-10000, i), 0)
				}
			}
			if l < 10000 && l > 0 {
				logger.Infof("deleting remainder")
				errs, err := bigtable.tableData.ApplyBulk(ctx, rowsToDelete, muts[:len(rowsToDelete)])
				if err != nil {
					logger.WithError(err).Errorf("error deleting row: %v", rowsToDelete[i])
				}
				for _, err := range errs {
					utils.LogError(err, "bigtable apply bulk error, deleting remainer", 0)
				}
				break
			}
		}
	}

}
