package exporter

import (
	"bytes"
	"eth2-exporter/db"
	"eth2-exporter/rpc"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New().WithField("module", "exporter")

// If exporting an epoch fails for 10 consecutive times exporting this epoch will be disabled
// This is a workaround for a bug in the prysm archive node that causes epochs without blocks
// to not be archived properly (see https://github.com/prysmaticlabs/prysm/issues/4165)
var epochBlacklist = make(map[uint64]uint64)

// Start will start the export of data from rpc into the database
func Start(client rpc.Client) error {

	if utils.Config.Indexer.FullIndexOnStartup {
		logger.Printf("Performing one time full db reindex")
		head, err := client.GetChainHead()
		if err != nil {
			logger.Fatal(err)
		}

		for epoch := uint64(1); epoch <= head.HeadEpoch; epoch++ {
			err := ExportEpoch(epoch, client)

			if err != nil {
				logger.Error(err)
			}
		}
	}

	if utils.Config.Indexer.IndexMissingEpochsOnStartup {
		// Add any missing epoch to the export set (might happen if the indexer was stopped for a long period of time)
		epochs, err := db.GetAllEpochs()
		if err != nil {
			logger.Fatal(err)
		}

		for i := 0; i < len(epochs)-1; i++ {
			if epochs[i] != epochs[i+1]-1 && epochs[i] != epochs[i+1] {
				logger.Println("Epochs between", epochs[i], "and", epochs[i+1], "are missing!")

				for epoch := epochs[i]; epoch <= epochs[i+1]; epoch++ {
					err := ExportEpoch(epoch, client)
					if err != nil {
						logger.Error(err)
					}
					logger.Printf("Finished export for epoch %v", epoch)
				}
			}
		}
	}

	if utils.Config.Indexer.CheckAllBlocksOnStartup {
		// Make sure that all blocks are correct by comparing all block hashes in the database to the ones we have in the node
		head, err := client.GetChainHead()
		if err != nil {
			logger.Fatal(err)
		}

		dbBlocks, err := db.GetLastPendingAndProposedBlocks(1, head.HeadEpoch)
		if err != nil {
			logger.Fatal(err)
		}

		nodeBlocks, err := GetLastBlocks(1, head.HeadEpoch, client)
		if err != nil {
			logger.Fatal(err)
		}

		blocksMap := make(map[string]*types.BlockComparisonContainer)

		for _, block := range dbBlocks {
			key := fmt.Sprintf("%v-%x", block.Slot, block.BlockRoot)
			_, found := blocksMap[key]

			if !found {
				blocksMap[key] = &types.BlockComparisonContainer{Epoch: block.Epoch}
			}

			blocksMap[key].Db = block
		}
		for _, block := range nodeBlocks {
			key := fmt.Sprintf("%v-%x", block.Slot, block.BlockRoot)
			_, found := blocksMap[key]

			if !found {
				blocksMap[key] = &types.BlockComparisonContainer{Epoch: block.Epoch}
			}

			blocksMap[key].Node = block
		}

		epochsToExport := make(map[uint64]bool)

		for key, block := range blocksMap {
			if block.Db == nil {
				logger.Printf("Queuing epoch %v for export as block %v is present on the node but missing in the db", block.Epoch, key)
				epochsToExport[block.Epoch] = true
			} else if block.Node == nil {
				logger.Printf("Queuing epoch %v for export as block %v is present on the db but missing in the node", block.Epoch, key)
				epochsToExport[block.Epoch] = true
			} else if bytes.Compare(block.Db.BlockRoot, block.Node.BlockRoot) != 0 {
				logger.Printf("Queuing epoch %v for export as block %v has a different hash in the db as on the node", block.Epoch, key)
				epochsToExport[block.Epoch] = true
			}
		}

		logger.Printf("Exporting %v epochs.", len(epochsToExport))

		keys := make([]uint64, 0)
		for k := range epochsToExport {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i] < keys[j]
		})

		for _, epoch := range keys {
			logger.Printf("Exporting epoch %v", epoch)

			err = ExportEpoch(epoch, client)

			if err != nil {
				logger.Errorf("error exporting epoch: %v", err)
				if utils.EpochToTime(epoch).Before(time.Now().Add(time.Hour * -24)) {
					epochBlacklist[epoch]++
				}
			}
			logger.Printf("Finished export for epoch %v", epoch)
		}
	}

	for true {

		head, err := client.GetChainHead()
		if err != nil {
			logger.Fatal(err)
		}

		dbBlocks, err := db.GetLastPendingAndProposedBlocks(head.FinalizedEpoch-1, head.HeadEpoch)
		if err != nil {
			logger.Fatal(err)
		}

		nodeBlocks, err := GetLastBlocks(head.FinalizedEpoch-1, head.HeadEpoch, client)
		if err != nil {
			logger.Fatal(err)
		}

		blocksMap := make(map[string]*types.BlockComparisonContainer)

		for _, block := range dbBlocks {
			key := fmt.Sprintf("%v-%x", block.Slot, block.BlockRoot)
			_, found := blocksMap[key]

			if !found {
				blocksMap[key] = &types.BlockComparisonContainer{Epoch: block.Epoch}
			}

			blocksMap[key].Db = block
		}
		for _, block := range nodeBlocks {
			key := fmt.Sprintf("%v-%x", block.Slot, block.BlockRoot)
			_, found := blocksMap[key]

			if !found {
				blocksMap[key] = &types.BlockComparisonContainer{Epoch: block.Epoch}
			}

			blocksMap[key].Node = block
		}

		epochsToExport := make(map[uint64]bool)

		for key, block := range blocksMap {
			if block.Db == nil {
				logger.Printf("Queuing epoch %v for export as block %v is present on the node but missing in the db", block.Epoch, key)
				epochsToExport[block.Epoch] = true
			} else if block.Node == nil {
				logger.Printf("Queuing epoch %v for export as block %v is present on the db but missing in the node", block.Epoch, key)
				epochsToExport[block.Epoch] = true
			} else if bytes.Compare(block.Db.BlockRoot, block.Node.BlockRoot) != 0 {
				logger.Printf("Queuing epoch %v for export as block %v has a different hash in the db as on the node", block.Epoch, key)
				epochsToExport[block.Epoch] = true
			}
		}

		// Add any missing epoch to the export set (might happen if the indexer was stopped for a long period of time)
		epochs, err := db.GetAllEpochs()
		if err != nil {
			logger.Fatal(err)
		}

		// Add not yet exported epochs to the export set (for example during the initial sync)
		if len(epochs) > 0 && epochs[len(epochs)-1] < head.HeadEpoch {
			for i := epochs[len(epochs)-1]; i <= head.HeadEpoch; i++ {
				epochsToExport[i] = true
			}
		} else if len(epochs) > 0 && epochs[0] != 0 { // Export the genesis epoch if not yet present in the db
			epochsToExport[0] = true
		} else if len(epochs) == 0 { // No epochs are present int the db
			for i := uint64(0); i <= head.HeadEpoch; i++ {
				epochsToExport[i] = true
			}
		}

		logger.Printf("Exporting %v epochs.", len(epochsToExport))

		keys := make([]uint64, 0)
		for k := range epochsToExport {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i] < keys[j]
		})

		for _, epoch := range keys {
			if epochBlacklist[epoch] > 3 {
				logger.Printf("Skipping export of epoch %v as it has errored %d times", epoch, epochBlacklist[epoch])
				continue
			}

			logger.Printf("Exporting epoch %v", epoch)

			err = ExportEpoch(epoch, client)

			if err != nil {
				logger.Errorf("error exporting epoch: %v", err)
				if utils.EpochToTime(epoch).Before(time.Now().Add(time.Hour * -24)) {
					epochBlacklist[epoch]++
				}
			}
			logger.Printf("Finished export for epoch %v", epoch)
		}

		// Update epoch statistics up to 10 epochs after the last finalized epoch
		startEpoch := uint64(0)
		if head.FinalizedEpoch > 10 {
			startEpoch = head.FinalizedEpoch - 10
		}
		err = updateEpochStatus(client, startEpoch, head.HeadEpoch)
		if err != nil {
			logger.Fatal(err)
		}

		err = exportAttestationPool(client)
		if err != nil {
			logger.Fatal(err)
		}

		err = exportValidatorQueue(client)
		if err != nil {
			logger.Error(err)
		}

		err = MarkOrphanedBlocks(head.FinalizedEpoch-1, head.HeadEpoch, nodeBlocks)
		if err != nil {
			logger.Fatal(err)
		}

		time.Sleep(time.Second * 10)
	}

	return nil
}

// MarkOrphanedBlocks will mark the orphaned blocks in the database
func MarkOrphanedBlocks(startEpoch, endEpoch uint64, blocks []*types.MinimalBlock) error {
	blocksMap := make(map[string]bool)

	for _, block := range blocks {
		blocksMap[fmt.Sprintf("%x", block.BlockRoot)] = false
	}

	orphanedBlocks := make([][]byte, 0)
	parentRoot := ""
	for i := len(blocks) - 1; i >= 0; i-- {
		blockRoot := fmt.Sprintf("%x", blocks[i].BlockRoot)

		if i == len(blocks)-1 { // First block is always canon
			parentRoot = fmt.Sprintf("%x", blocks[i].ParentRoot)
			blocksMap[blockRoot] = true
			continue
		}
		if parentRoot != blockRoot { // Block is not part of the canonical chain
			logger.Errorf("Block %x at slot %v in epoch %v has been orphaned", blocks[i].BlockRoot, blocks[i].Slot, blocks[i].Epoch)
			orphanedBlocks = append(orphanedBlocks, blocks[i].BlockRoot)
			continue
		}
		blocksMap[blockRoot] = true
		parentRoot = fmt.Sprintf("%x", blocks[i].ParentRoot)
	}

	return db.UpdateCanonicalBlocks(startEpoch, endEpoch, orphanedBlocks)
}

// GetLastBlocks will get all blocks for a range of epochs
func GetLastBlocks(startEpoch, endEpoch uint64, client rpc.Client) ([]*types.MinimalBlock, error) {
	wrappedBlocks := make([]*types.MinimalBlock, 0)

	for epoch := startEpoch; epoch <= endEpoch; epoch++ {
		startSlot := epoch * utils.Config.Chain.SlotsPerEpoch
		endSlot := (epoch+1)*utils.Config.Chain.SlotsPerEpoch - 1
		for slot := startSlot; slot <= endSlot; slot++ {
			blocks, err := client.GetBlocksBySlot(slot)
			if err != nil {
				logger.Fatal(err)
			}

			for _, block := range blocks {
				wrappedBlocks = append(wrappedBlocks, &types.MinimalBlock{
					Epoch:      epoch,
					Slot:       block.Slot,
					BlockRoot:  block.BlockRoot,
					ParentRoot: block.ParentRoot,
				})
			}
		}

		logger.Printf("Retrieving all blocks for epoch %v. %v epochs remaining", epoch, endEpoch-epoch)
	}

	return wrappedBlocks, nil
}

// ExportEpoch will export an epoch from rpc into the database
func ExportEpoch(epoch uint64, client rpc.Client) error {
	start := time.Now()

	logger.Printf("Retrieving data for epoch %v", epoch)
	data, err := client.GetEpochData(epoch)

	if err != nil {
		return fmt.Errorf("error retrieving epoch data: %v", err)
	}

	logger.Printf("Data for epoch %v retrieved, took %v", epoch, time.Since(start))

	return db.SaveEpoch(data)
}

func exportAttestationPool(client rpc.Client) error {
	attestations, err := client.GetAttestationPool()

	if err != nil {
		return fmt.Errorf("error retrieving attestation pool data: %v", err)
	}

	return db.SaveAttestationPool(attestations)
}

func exportValidatorQueue(client rpc.Client) error {

	validators, validatorIndices, err := client.GetValidatorQueue()
	if err != nil {
		return fmt.Errorf("error retrieving validator queue data: %v", err)
	}

	return db.SaveValidatorQueue(validators, validatorIndices)
}

func updateEpochStatus(client rpc.Client, startEpoch, endEpoch uint64) error {
	for epoch := startEpoch; epoch <= endEpoch; epoch++ {
		epochParticipationStats, err := client.GetValidatorParticipation(epoch)
		if err != nil {
			logger.Printf("error retrieving epoch participation statistics: %v", err)
		} else {
			logger.Printf("Updating epoch %v with status finalized = %v", epoch, epochParticipationStats.Finalized)
			err := db.UpdateEpochStatus(epochParticipationStats)

			if err != nil {
				return err
			}
		}
	}
	return nil
}
