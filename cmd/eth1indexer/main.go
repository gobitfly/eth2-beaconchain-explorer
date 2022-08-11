package main

import (
	"eth2-exporter/db"
	"eth2-exporter/rpc"
	"eth2-exporter/types"
	"flag"
	"fmt"
	"sync/atomic"
	"time"

	"cloud.google.com/go/bigtable"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func main() {
	// localhost:8545
	// erigonEndpoint := flag.String("erigon", "", "Erigon archive node enpoint")
	start := flag.Int64("start", 0, "Block to start indexing")
	end := flag.Int64("end", 0, "Block to finish indexing")

	flag.Parse()

	bt, err := db.NewBigtable("etherchain", "etherchain", "1")
	if err != nil {
		logrus.Fatalf("error connecting to bigtable: %v", err)
	}
	defer bt.Close()

	// bt.DeleteRowsWithPrefix("1:b:")
	// return

	// if erigonEndpoint != nil && *erigonEndpoint != "" {
	// 	logrus.Infof("indexing from node %v", erigonEndpoint)
	// 	go IndexFromNode(bt, erigonEndpoint, start, end)
	// }

	transforms := make([]func(blk *types.Eth1Block) (*types.BulkMutations, error), 0)
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
				return err
			}

			dbStart := time.Now()
			err = bt.SaveBlock(bc)
			if err != nil {
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

func IndexBlocksFromBigtable(bt *db.Bigtable, start, end *int64, transforms []func(blk []*types.Eth1Block) (*types.BulkMutations, error)) error {
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

			blocks[0], err = bt.GetBlock(uint64(i))
			if err != nil {
				return err
			}

			blocks[1], err = bt.GetBlock(uint64(i - 1))
			if err != nil {
				return err
			}

			bulkMuts := types.BulkMutations{}
			for _, transform := range transforms {
				muts, err := transform(blocks)
				if err != nil {
					logrus.WithError(err).Error("error transforming block")
				}
				bulkMuts.Keys = append(bulkMuts.Keys, muts.Keys...)
				bulkMuts.Muts = append(bulkMuts.Muts, muts.Muts...)
			}

			err = bt.WriteBulk(&bulkMuts)
			if err != nil {
				return fmt.Errorf("error writing to bigtable err: %w", err)
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

func IndexFromBigtable(bt *db.Bigtable, start, end *int64, transforms []func(blk *types.Eth1Block) (*types.BulkMutations, error)) error {
	g := new(errgroup.Group)
	g.SetLimit(10)

	startTs := time.Now()
	lastTickTs := time.Now()

	processedBlocks := int64(0)
	logrus.Infof("fetching blocks from %d to %d", *end, *start)
	for i := *end; i >= *start; i-- {
		i := i
		g.Go(func() error {

			block, err := bt.GetBlock(uint64(i))
			if err != nil {
				return err
			}

			bulkMuts := types.BulkMutations{
				Keys: make([]string, 0),
				Muts: make([]*bigtable.Mutation, 0),
			}
			for _, transform := range transforms {
				muts, err := transform(block)
				if err != nil {
					logrus.WithError(err).Error("error transforming block")
				}
				bulkMuts.Keys = append(bulkMuts.Keys, muts.Keys...)
				bulkMuts.Muts = append(bulkMuts.Muts, muts.Muts...)
			}
			// logrus.Infof("writing: %v", len(bulkMuts.Keys))
			err = bt.WriteBulk(&bulkMuts)
			if err != nil {
				return fmt.Errorf("error writing to bigtable err: %w", err)
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
		return err
	}

	return nil
}
