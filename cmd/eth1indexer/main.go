package main

import (
	"eth2-exporter/db"
	"eth2-exporter/rpc"
	"eth2-exporter/types"
	"flag"
	"fmt"
	"sync/atomic"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/karlseguin/ccache/v2"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func main() {
	// localhost:8545
	erigonEndpoint := flag.String("erigon", "", "Erigon archive node enpoint")
	start := flag.Int64("start", 0, "Block to start indexing")
	end := flag.Int64("end", 0, "Block to finish indexing")

	flag.Parse()

	bt, err := db.NewBigtable("etherchain", "etherchain", "1")
	if err != nil {
		logrus.Fatalf("error connecting to bigtable: %v", err)
	}
	defer bt.Close()

	// err = bt.CheckForGapsInDataTable()
	// if err != nil {
	// 	logrus.Fatal(err)
	// }

	// return
	// bt.DeleteRowsWithPrefix("1:b:")
	// return

	if erigonEndpoint != nil && *erigonEndpoint != "" {
		logrus.Infof("indexing from node %v", *erigonEndpoint)
		IndexFromNode(bt, erigonEndpoint, start, end)
		return
	}

	transforms := make([]func(blk *types.Eth1Block, cache *ccache.Cache) (*types.BulkMutations, *types.BulkMutations, error), 0)
	transforms = append(transforms, bt.TransformBlock, bt.TransformTx, bt.TransformItx, bt.TransformERC20, bt.TransformERC721, bt.TransformERC1155, bt.TransformUncle)
	// transforms = append(transforms, bt.TransformTx)

	logrus.Infof("indexing from bigtable")
	err = IndexFromBigtable(bt, start, end, transforms)
	if err != nil {
		logrus.WithError(err).Fatalf("error indexing from bigtable")
	}

	// utils.WaitForCtrlC()

}

func IndexFromNode(bt *db.Bigtable, erigonEndpoint *string, start, end *int64) {
	client, err := rpc.NewErigonClient(*erigonEndpoint)
	if err != nil {
		logrus.Fatal(err)
	}

	g := new(errgroup.Group)
	g.SetLimit(30)

	startTs := time.Now()
	lastTickTs := time.Now()

	processedBlocks := int64(0)

	for i := *end; i >= *start; i-- {

		i := i
		g.Go(func() error {
			blockStartTs := time.Now()
			bc, timings, err := client.GetBlock(i)

			if err != nil {
				logrus.Error(err)
				return err
			}

			dbStart := time.Now()
			err = bt.SaveBlock(bc)
			if err != nil {
				logrus.Error(err)
				return err
			}
			current := atomic.AddInt64(&processedBlocks, 1)
			if current%100 == 0 {
				logrus.Infof("retrieved & saved block %v (0x%x) in %v (header: %v, receipts: %v, traces: %v, db: %v)", bc.Number, bc.Hash, time.Since(blockStartTs), timings.Headers, timings.Receipts, timings.Traces, time.Since(dbStart))
				logrus.Infof("processed %v blocks in %v (%.1f blocks / sec)", current, time.Since(startTs), float64((current))/time.Since(lastTickTs).Seconds())

				lastTickTs = time.Now()
				atomic.StoreInt64(&processedBlocks, 0)
			}
			return nil
		})

	}

	if err := g.Wait(); err == nil {
		logrus.Info("Successfully fetched all blocks")
	}
}

func IndexBlocksFromBigtable(bt *db.Bigtable, start, end *int64, transforms []func(blk []*types.Eth1Block) (bulkData *types.BulkMutations, bulkMetadataUpdates *types.BulkMutations, err error)) error {
	g := new(errgroup.Group)
	g.SetLimit(20)

	startTs := time.Now()
	lastTickTs := time.Now()

	processedBlocks := int64(0)
	logrus.Infof("fetching blocks from %d to %d", *end, *start)
	for i := *end; i >= *start; i-- {
		i := i
		g.Go(func() error {
			var err error

			blocks := make([]*types.Eth1Block, 2)

			blocks[0], err = bt.GetBlockFromBlocksTable(uint64(i))
			if err != nil {
				return err
			}

			blocks[1], err = bt.GetBlockFromBlocksTable(uint64(i - 1))
			if err != nil {
				return err
			}

			bulkMutsData := types.BulkMutations{}
			bulkMutsMetadataUpdate := types.BulkMutations{}
			for _, transform := range transforms {
				mutsData, mutsMetadataUpdate, err := transform(blocks)
				if err != nil {
					logrus.WithError(err).Error("error transforming block")
				}
				bulkMutsData.Keys = append(bulkMutsData.Keys, mutsData.Keys...)
				bulkMutsData.Muts = append(bulkMutsData.Muts, mutsData.Muts...)

				if mutsMetadataUpdate != nil {
					bulkMutsMetadataUpdate.Keys = append(bulkMutsMetadataUpdate.Keys, mutsMetadataUpdate.Keys...)
					bulkMutsMetadataUpdate.Muts = append(bulkMutsMetadataUpdate.Muts, mutsMetadataUpdate.Muts...)
				}
			}

			if len(bulkMutsData.Keys) > 0 {
				err = bt.WriteBulk(&bulkMutsData, bt.GetDataTable())
				if err != nil {
					return fmt.Errorf("error writing to bigtable data table: %w", err)
				}
			}

			if len(bulkMutsMetadataUpdate.Keys) > 0 {
				err = bt.WriteBulk(&bulkMutsMetadataUpdate, bt.GetMetadataUpdatesTable())
				if err != nil {
					return fmt.Errorf("error writing to bigtable metadata updates table: %w", err)
				}
			}

			current := atomic.AddInt64(&processedBlocks, 1)
			if current%100 == 0 {
				logrus.Infof("processed %v blocks in %v (%.1f blocks / sec)", current, time.Since(startTs), float64((current))/time.Since(lastTickTs).Seconds())

				lastTickTs = time.Now()
				atomic.StoreInt64(&processedBlocks, 0)
			}
			return nil
		})

	}

	if err := g.Wait(); err == nil {
		logrus.Info("Successfully fetched all blocks")
	} else {
		return err
	}
	return nil
}

func IndexFromBigtable(bt *db.Bigtable, start, end *int64, transforms []func(blk *types.Eth1Block, cache *ccache.Cache) (bulkData *types.BulkMutations, bulkMetadataUpdates *types.BulkMutations, err error)) error {
	g := new(errgroup.Group)
	g.SetLimit(10)

	startTs := time.Now()
	lastTickTs := time.Now()

	processedBlocks := int64(0)

	cache := ccache.New(ccache.Configure().MaxSize(1000000).ItemsToPrune(500))

	logrus.Infof("fetching blocks from %d to %d", *end, *start)
	for i := *end; i >= *start; i-- {
		i := i
		g.Go(func() error {

			block, err := bt.GetBlockFromBlocksTable(uint64(i))
			if err != nil {
				return err
			}

			bulkMutsData := types.BulkMutations{}
			bulkMutsMetadataUpdate := types.BulkMutations{}
			for _, transform := range transforms {
				mutsData, mutsMetadataUpdate, err := transform(block, cache)
				if err != nil {
					logrus.WithError(err).Error("error transforming block")
				}
				bulkMutsData.Keys = append(bulkMutsData.Keys, mutsData.Keys...)
				bulkMutsData.Muts = append(bulkMutsData.Muts, mutsData.Muts...)

				if mutsMetadataUpdate != nil {
					bulkMutsMetadataUpdate.Keys = append(bulkMutsMetadataUpdate.Keys, mutsMetadataUpdate.Keys...)
					bulkMutsMetadataUpdate.Muts = append(bulkMutsMetadataUpdate.Muts, mutsMetadataUpdate.Muts...)
				}
			}

			if len(bulkMutsData.Keys) > 0 {
				err = bt.WriteBulk(&bulkMutsData, bt.GetDataTable())
				if err != nil {
					return fmt.Errorf("error writing to bigtable data table: %w", err)
				}
			}

			if len(bulkMutsMetadataUpdate.Keys) > 0 {
				err = bt.WriteBulk(&bulkMutsMetadataUpdate, bt.GetMetadataUpdatesTable())
				if err != nil {
					return fmt.Errorf("error writing to bigtable metadata updates table: %w", err)
				}
			}

			current := atomic.AddInt64(&processedBlocks, 1)
			if current%500 == 0 {
				curr := uint64(*end) - block.GetNumber()
				diff := *end - *start
				if curr < 1 {
					curr = 0
				}
				if diff < 1 {
					diff = 1
				}
				perc := float64(curr) / float64(diff)
				logrus.Infof("currently processing block: %v; processed %v blocks in %v (%.1f blocks / sec); sync is %.1f%% complete", block.GetNumber(), current, time.Since(startTs), float64((current))/time.Since(lastTickTs).Seconds(), perc*100)
				lastTickTs = time.Now()
				atomic.StoreInt64(&processedBlocks, 0)
			}
			return nil
		})

	}

	if err := g.Wait(); err == nil {
		logrus.Info("Successfully fetched all blocks")
	} else {
		logrus.Error(err)
		return err
	}

	return nil
}
