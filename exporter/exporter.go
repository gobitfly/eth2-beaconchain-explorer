package exporter

import (
	"bytes"
	"eth2-exporter/db"
	"eth2-exporter/metrics"
	"eth2-exporter/rpc"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var logger = logrus.New().WithField("module", "exporter")

// If exporting an epoch fails for 10 consecutive times exporting this epoch will be disabled
// This is a workaround for a bug in the prysm archive node that causes epochs without blocks
// to not be archived properly (see https://github.com/prysmaticlabs/prysm/issues/4165)
var epochBlacklist = make(map[uint64]uint64)
var fullCheckRunning = uint64(0)

var Client *rpc.Client

// Start will start the export of data from rpc into the database
func Start(client rpc.Client) error {
	go networkLivenessUpdater(client)
	go eth1DepositsExporter()
	go genesisDepositsExporter(client)
	go checkSubscriptions()
	go syncCommitteesExporter(client)
	go syncCommitteesCountExporter()
	if utils.Config.SSVExporter.Enabled {
		go ssvExporter()
	}
	if utils.Config.RocketpoolExporter.Enabled {
		go rocketpoolExporter()
	}

	if utils.Config.Indexer.PubKeyTagsExporter.Enabled {
		go UpdatePubkeyTag()
	}

	if utils.Config.MevBoostRelayExporter.Enabled {
		go mevBoostRelaysExporter()
	}
	// wait until the beacon-node is available
	for {
		head, err := client.GetChainHead()
		if err == nil {
			logger.Infof("beacon node is available with head slot: %v", head.HeadSlot)
			// if we are still waiting for genesis, export epoch 0 with genesis data
			// epoch 0 will only contain genesis data at this point
			if head.HeadSlot == 0 {
				err := ExportEpoch(0, client)
				if err != nil {
					logger.Errorf("error exporting genesis information for epoch 0 err: %v", err)
				}
			}

			break
		}
		logger.Errorf("beacon-node seems to be unavailable: %v", err)
		time.Sleep(time.Second * 10)
	}

	if utils.Config.Indexer.FullIndexOnStartup {
		logger.Printf("performing one time full db reindex")
		head, err := client.GetChainHead()
		if err != nil {
			utils.LogFatal(err, "getting chain head from client for full db reindex error", 0)
		}

		for epoch := uint64(0); epoch <= head.HeadEpoch; epoch++ {
			err := ExportEpoch(epoch, client)

			if err != nil {
				utils.LogError(err, "exporting all epochs up to head epoch error", 0)
			}
		}
	}

	if utils.Config.Indexer.FixCanonOnStartup {
		logger.Printf("performing one time full canon check")
		head, err := client.GetChainHead()
		if err != nil {
			utils.LogFatal(err, "getting chain head from client for full canon check error", 0)
		}

		for epoch := int64(head.HeadEpoch) - 1; epoch >= 0; epoch-- {
			blocks, err := client.GetBlockStatusByEpoch(uint64(epoch))
			if err != nil {
				logger.Errorf("error retrieving block status: %v", err)
				continue
			}
			err = db.SetBlockStatus(blocks)
			if err != nil {
				logger.Errorf("error saving block status: %v", err)
				continue
			}
		}
	}

	if utils.Config.Indexer.IndexMissingEpochsOnStartup {
		// Add any missing epoch to the export set (might happen if the indexer was stopped for a long period of time)
		logger.Infof("checking for missing epochs")
		epochs, err := db.GetAllEpochs()
		if err != nil {
			utils.LogFatal(err, "getting all epochs from db error", 0)
		}

		if len(epochs) > 0 && epochs[0] != 0 {
			err := ExportEpoch(0, client)
			if err != nil {
				utils.LogError(err, "exporting epoch 0 error", 0)
			}
			logger.Printf("finished export for epoch %v", 0)
			epochs = append([]uint64{0}, epochs...)
		}

		for i := 0; i < len(epochs)-1; i++ {
			if epochs[i] != epochs[i+1]-1 && epochs[i] != epochs[i+1] {
				logger.Println("Epochs between", epochs[i], "and", epochs[i+1], "are missing!")

				for epoch := epochs[i]; epoch <= epochs[i+1]; epoch++ {
					err := ExportEpoch(epoch, client)
					if err != nil {
						utils.LogError(err, fmt.Errorf("exporting epochs between %v and %v error", epochs[i], epochs[i+1]), 0)
					}
					logger.Printf("finished export for epoch %v", epoch)
				}
			}
		}
	}

	if utils.Config.Indexer.CheckAllBlocksOnStartup {
		// Make sure that all blocks are correct by comparing all block hashes in the database to the ones we have in the node
		head, err := client.GetChainHead()
		if err != nil {
			utils.LogFatal(err, "getting chain head from client for full blocks check error", 0)
		}

		dbBlocks, err := db.GetLastPendingAndProposedBlocks(1, head.HeadEpoch)
		if err != nil {
			utils.LogFatal(err, "getting proposed and pending blocks error ", 0)
		}

		nodeBlocks, err := GetLastBlocks(1, head.HeadEpoch, client)
		if err != nil {
			utils.LogFatal(err, "getting blocks for range of epochs error", 0)
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
				logger.Printf("queuing epoch %v for export as block %v is present on the node but missing in the db", block.Epoch, key)
				epochsToExport[block.Epoch] = true
			} else if block.Node == nil && !strings.HasSuffix(key, "-00") { //do not re-export because of missed blocks
				logger.Printf("queuing epoch %v for export as block %v is present on the db but missing in the node", block.Epoch, key)
				epochsToExport[block.Epoch] = true
			} else if !bytes.Equal(block.Db.BlockRoot, block.Node.BlockRoot) {
				logger.Printf("queuing epoch %v for export as block %v has a different hash in the db as on the node", block.Epoch, key)
				epochsToExport[block.Epoch] = true
			}
		}

		logger.Printf("exporting %v epochs.", len(epochsToExport))

		keys := make([]uint64, 0)
		for k := range epochsToExport {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i] < keys[j]
		})

		for _, epoch := range keys {
			err = ExportEpoch(epoch, client)

			if err != nil {
				logger.Errorf("error exporting epoch: %v", err)
				if utils.EpochToTime(epoch).Before(time.Now().Add(time.Hour * -24)) {
					epochBlacklist[epoch]++
				}
			}
		}
	}

	if utils.Config.Indexer.UpdateAllEpochStatistics {
		// Update all epoch statistics
		head, err := client.GetChainHead()
		if err != nil {
			utils.LogFatal(err, "getting chain head from client for updating epoch statistics error", 0)
		}
		startEpoch := uint64(0)
		err = updateEpochStatus(client, startEpoch, head.HeadEpoch)
		if err != nil {
			utils.LogFatal(err, "error while updating epoch status", 0)
		}
	}

	newBlockChan := client.GetNewBlockChan()

	lastExportedSlot := uint64(0)

	// doFullCheck(client)

	logger.Infof("entering monitoring mode")
	for {
		block := <-newBlockChan
		// Do a full check on any epoch transition or after during the first run
		if utils.EpochOfSlot(lastExportedSlot) != utils.EpochOfSlot(block.Slot) || utils.EpochOfSlot(block.Slot) == 0 {
			go func() {
				v := atomic.LoadUint64(&fullCheckRunning)
				if v == 1 {
					logger.Infof("skipping full check as one is already running")
					return
				}
				atomic.StoreUint64(&fullCheckRunning, 1)
				doFullCheck(client, 0)
				atomic.StoreUint64(&fullCheckRunning, 0)
			}()
		}

		blocksMap := make(map[uint64]map[string]*types.Block)
		if blocksMap[block.Slot] == nil {
			blocksMap[block.Slot] = make(map[string]*types.Block)
		}
		blocksMap[block.Slot][fmt.Sprintf("%x", block.BlockRoot)] = block

		syncDuties := make(map[types.Slot]map[types.ValidatorIndex]bool)
		syncDuties[types.Slot(block.Slot)] = make(map[types.ValidatorIndex]bool)

		for validator, duty := range block.SyncDuties {
			syncDuties[types.Slot(block.Slot)][types.ValidatorIndex(validator)] = duty
		}

		attDuties := make(map[types.Slot]map[types.ValidatorIndex][]types.Slot)
		for validator, attestedSlot := range block.AttestationDuties {
			if attDuties[types.Slot(attestedSlot)] == nil {
				attDuties[types.Slot(attestedSlot)] = make(map[types.ValidatorIndex][]types.Slot)
			}
			if attDuties[types.Slot(attestedSlot)][types.ValidatorIndex(validator)] == nil {
				attDuties[types.Slot(attestedSlot)][types.ValidatorIndex(validator)] = []types.Slot{}
			}
			attDuties[types.Slot(attestedSlot)][types.ValidatorIndex(validator)] = append(attDuties[types.Slot(attestedSlot)][types.ValidatorIndex(validator)], types.Slot(block.Slot))
		}

		err := db.BigtableClient.SaveAttestationDuties(attDuties)
		if err != nil {
			logrus.Errorf("error exporting attestations to bigtable for block %v: %v", block.Slot, err)
		}
		err = db.BigtableClient.SaveSyncComitteeDuties(syncDuties)
		if err != nil {
			logrus.Errorf("error exporting sync committee duties to bigtable for block %v: %v", block.Slot, err)
		}

		err = db.SaveBlock(block, false)
		if err != nil {
			logger.Errorf("error saving block: %v", err)
		}

		err = db.UpdateMissedBlocksInEpochWithSlotCutoff(block.Slot)
		if err != nil {
			logger.Errorf("error marking missed blocks: %v", err)
		}

		lastExportedSlot = block.Slot
	}
}

// Will ensure the db is fully in sync with the node
func doFullCheck(client rpc.Client, lookback uint64) {
	logger.Infof("checking for new blocks/epochs to export")

	// Use the chain head as our current point of reference
	head, err := client.GetChainHead()
	if err != nil {
		logger.Errorf("error retrieving chain head: %v", err)
		return
	}

	startEpoch := uint64(0)
	// Set the start epoch to the epoch prior to the last finalized epoch
	if head.FinalizedEpoch > 1 {
		startEpoch = head.FinalizedEpoch - 1
	}

	startEpoch = startEpoch - lookback

	// If the network is experiencing finality issues limit the export to the last 10 epochs
	// Once the network reaches finality again all epochs should be exported again
	if head.HeadEpoch > 10 && head.HeadEpoch-head.FinalizedEpoch > 10 {
		logger.Infof("no finality since %v epochs, limiting lookback to the last 10 epochs", head.HeadEpoch-head.FinalizedEpoch)
		startEpoch = head.HeadEpoch - 10
	}

	// Retrieve the db contents for the epocht that should be exported
	dbBlocks, err := db.GetLastPendingAndProposedBlocks(startEpoch, head.HeadEpoch)
	if err != nil {
		logger.Errorf("error retrieving last pending and proposed blocks from the database: %v", err)
		return
	}

	// For the same epochs retrieve all block data from the node
	nodeBlocks, err := GetLastBlocks(startEpoch, head.HeadEpoch, client)
	if err != nil {
		logger.Errorf("error retrieving last blocks from backend node: %v", err)
		return
	}

	// Compare the blocks in the db with the blocks in the node
	// If a block is missing on the node that has been exported in the db
	// or if a block is missing in the db that is present in the node
	// export this epoch to the db again
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
			logger.Printf("queuing epoch %v for export as block %v is present on the node but missing in the db", block.Epoch, key)
			epochsToExport[block.Epoch] = true
		} else if block.Node == nil {
			if !strings.HasSuffix(key, "-00") {
				logger.Printf("queuing epoch %v for export as block %v is present on the db but missing in the node", block.Epoch, key)
				epochsToExport[block.Epoch] = true
			}
		} else if !bytes.Equal(block.Db.BlockRoot, block.Node.BlockRoot) {
			logger.Printf("queuing epoch %v for export as block %v has a different hash in the db as on the node", block.Epoch, key)
			epochsToExport[block.Epoch] = true
		}
	}

	// Add any missing epoch to the export set (might happen if the indexer was stopped for a long period of time)
	epochs, err := db.GetAllEpochs()
	if err != nil {
		logger.Errorf("error retrieving all epochs from the db: %v", err)
		return
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

	// Check for epoch gaps
	for i := 0; i < len(epochs)-1; i++ {
		currentEpoch := epochs[i]
		nextEpoch := epochs[i+1]

		if currentEpoch != nextEpoch-1 {
			logger.Infof("epoch gap found between epochs %v and %v", currentEpoch, nextEpoch)
			for j := currentEpoch + 1; j <= nextEpoch-1; j++ {
				logger.Printf("queuing epoch %v for export", j)
				epochsToExport[j] = true
			}
		}
	}

	logger.Printf("exporting %v epochs.", len(epochsToExport))

	keys := make([]uint64, 0)
	for k := range epochsToExport {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	for _, epoch := range keys {
		if epochBlacklist[epoch] > 3 {
			logger.Printf("skipping export of epoch %v as it has errored %d times", epoch, epochBlacklist[epoch])
			continue
		}

		logger.Printf("exporting epoch %v", epoch)

		err = ExportEpoch(epoch, client)

		if err != nil {
			logger.Errorf("error exporting epoch: %v", err)
			if utils.EpochToTime(epoch).Before(time.Now().Add(time.Hour * -24)) {
				epochBlacklist[epoch]++
			}
		}
		logger.Printf("finished export for epoch %v", epoch)

		if len(keys) > 10 && epoch%10 == 0 && epoch >= 10 { // this ensures that epoch finalization is properly updated during long running exports
			logger.Infof("updating status of epochs %v-%v", startEpoch, head.HeadEpoch)
			err = updateEpochStatus(client, epoch-10, epoch)
			if err != nil {
				logger.Errorf("error updating epoch stratus: %v", err)
			}
		}
	}

	logger.Infof("marking orphaned blocks of epochs %v-%v", startEpoch, head.HeadEpoch)
	err = MarkOrphanedBlocks(startEpoch, head.HeadEpoch, nodeBlocks)
	if err != nil {
		logger.Errorf("error marking orphaned blocks: %v", err)
	}

	if head.HeadEpoch > 0 {
		logger.Infof("marking missed blocks of epochs %v-%v", startEpoch, head.HeadEpoch-1)
		err = MarkMissedBlocks(startEpoch, head.HeadEpoch-1)
		if err != nil {
			logger.Errorf("error marking missed blocks: %v", err)
		}
	}

	// Update epoch statistics up to 10 epochs after the last finalized epoch
	startEpoch = uint64(0)
	if head.FinalizedEpoch > 10 {
		startEpoch = head.FinalizedEpoch - 10
		if head.HeadEpoch-startEpoch > 10 {
			startEpoch = head.HeadEpoch - 10
		}
	}
	logger.Infof("updating status of epochs %v-%v", startEpoch, head.HeadEpoch)
	err = updateEpochStatus(client, startEpoch, head.HeadEpoch)
	if err != nil {
		logger.Errorf("error updating epoch stratus: %v", err)
	}

	logger.Infof("exporting validation queue")
	err = exportValidatorQueue(client)
	if err != nil {
		logger.Errorf("error exporting validator queue data: %v", err)
	}

	logger.Infof("finished exporting all new blocks/epochs")
}

// MarkOrphanedBlocks will mark the orphaned blocks in the database
func MarkOrphanedBlocks(startEpoch, endEpoch uint64, blocks []*types.MinimalBlock) error {
	return db.UpdateCanonicalBlocks(startEpoch, endEpoch, blocks)
}

// MarkMissedBlocks will mark the missed blocks in the database
func MarkMissedBlocks(startEpoch, endEpoch uint64) error {
	return db.UpdateMissedBlocks(startEpoch, endEpoch)
}

// GetLastBlocks will get all blocks for a range of epochs
func GetLastBlocks(startEpoch, endEpoch uint64, client rpc.Client) ([]*types.MinimalBlock, error) {
	wrappedBlocks := make([]*types.MinimalBlock, 0)

	for epoch := startEpoch; epoch <= endEpoch; epoch++ {
		startSlot := epoch * utils.Config.Chain.Config.SlotsPerEpoch
		endSlot := (epoch+1)*utils.Config.Chain.Config.SlotsPerEpoch - 1
		for slot := startSlot; slot <= endSlot; slot++ {
			blocks, err := client.GetBlocksBySlot(slot)
			if err != nil {
				return nil, err
			}

			for _, block := range blocks {
				wrappedBlocks = append(wrappedBlocks, &types.MinimalBlock{
					Epoch:      epoch,
					Slot:       block.Slot,
					BlockRoot:  block.BlockRoot,
					ParentRoot: block.ParentRoot,
					Canonical:  block.Canonical,
				})
			}
		}

		logger.Printf("retrieving all blocks for epoch %v. %v epochs remaining", epoch, endEpoch-epoch)
	}

	return wrappedBlocks, nil
}

// ExportEpoch will export an epoch from rpc into the database
func ExportEpoch(epoch uint64, client rpc.Client) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("export_epoch").Observe(time.Since(start).Seconds())
		logger.WithFields(logrus.Fields{"duration": time.Since(start), "epoch": epoch}).Info("completed exporting epoch")
	}()

	startGetEpochData := time.Now()
	logger.Printf("retrieving data for epoch %v", epoch)

	data, err := client.GetEpochData(epoch, false)
	if err != nil {
		return fmt.Errorf("error retrieving epoch data: %v", err)
	}
	metrics.TaskDuration.WithLabelValues("rpc_get_epoch_data").Observe(time.Since(startGetEpochData).Seconds())
	logger.WithFields(logrus.Fields{"duration": time.Since(startGetEpochData), "epoch": epoch}).Info("completed getting epoch-data")
	logger.Printf("data for epoch %v retrieved, took %v", epoch, time.Since(start))

	if len(data.Validators) == 0 {
		return fmt.Errorf("error retrieving epoch data: no validators received for epoch")
	}

	// export epoch data to bigtable
	g := new(errgroup.Group)
	g.SetLimit(5)
	g.Go(func() error {
		err = db.BigtableClient.SaveValidatorBalances(epoch, data.Validators)
		if err != nil {
			return fmt.Errorf("error exporting validator balances to bigtable: %v", err)
		}
		return nil
	})
	g.Go(func() error {
		err = db.BigtableClient.SaveProposalAssignments(epoch, data.ValidatorAssignmentes.ProposerAssignments)
		if err != nil {
			return fmt.Errorf("error exporting proposal assignments to bigtable: %v", err)
		}
		return nil
	})
	g.Go(func() error {
		err = db.BigtableClient.SaveAttestationDuties(data.AttestationDuties)
		if err != nil {
			return fmt.Errorf("error exporting attestations to bigtable: %v", err)
		}
		return nil
	})
	g.Go(func() error {
		err = db.BigtableClient.SaveProposals(data.Blocks)
		if err != nil {
			return fmt.Errorf("error exporting proposals to bigtable: %v", err)
		}
		return nil
	})
	g.Go(func() error {
		err = db.BigtableClient.SaveSyncComitteeDuties(data.SyncDuties)
		if err != nil {
			return fmt.Errorf("error exporting sync committee duties to bigtable: %v", err)
		}
		return nil
	})

	err = g.Wait()
	if err != nil {
		return fmt.Errorf("error during bigtable export: %w", err)
	}

	// at this point all epoch data has been written to bigtable
	err = db.SaveEpoch(data, client)
	if err != nil {
		return fmt.Errorf("error saving epoch data: %w", err)
	}

	services.ReportStatus("epochExporter", "Running", nil)
	return nil
}

func exportValidatorQueue(client rpc.Client) error {
	queue, err := client.GetValidatorQueue()
	if err != nil {
		return fmt.Errorf("error retrieving validator queue data: %v", err)
	}

	return db.SaveValidatorQueue(queue)
}

func updateEpochStatus(client rpc.Client, startEpoch, endEpoch uint64) error {
	for epoch := startEpoch; epoch <= endEpoch; epoch++ {
		epochParticipationStats, err := client.GetValidatorParticipation(epoch)
		if err != nil {
			logger.Printf("error retrieving epoch participation statistics: %v", err)
		} else {
			logger.Printf("updating epoch %v with participation rate %v", epoch, epochParticipationStats.GlobalParticipationRate)
			err := db.UpdateEpochStatus(epochParticipationStats)

			if err != nil {
				return err
			}
		}
	}
	return nil
}

func networkLivenessUpdater(client rpc.Client) {
	var prevHeadEpoch uint64
	err := db.WriterDb.Get(&prevHeadEpoch, "SELECT COALESCE(MAX(headepoch), 0) FROM network_liveness")
	if err != nil {
		utils.LogFatal(err, "getting previous head epoch from db error", 0)
	}

	epochDuration := time.Second * time.Duration(utils.Config.Chain.Config.SecondsPerSlot*utils.Config.Chain.Config.SlotsPerEpoch)
	slotDuration := time.Second * time.Duration(utils.Config.Chain.Config.SecondsPerSlot)

	for {
		head, err := client.GetChainHead()
		if err != nil {
			logger.Errorf("error getting chainhead when exporting networkliveness: %v", err)
			time.Sleep(slotDuration)
			continue
		}

		if prevHeadEpoch == head.HeadEpoch {
			time.Sleep(slotDuration)
			continue
		}

		// wait for node to be synced
		if time.Now().Add(-epochDuration).After(utils.EpochToTime(head.HeadEpoch)) {
			time.Sleep(slotDuration)
			continue
		}

		_, err = db.WriterDb.Exec(`
			INSERT INTO network_liveness (ts, headepoch, finalizedepoch, justifiedepoch, previousjustifiedepoch)
			VALUES (NOW(), $1, $2, $3, $4)`,
			head.HeadEpoch, head.FinalizedEpoch, head.JustifiedEpoch, head.PreviousJustifiedEpoch)
		if err != nil {
			logger.Errorf("error saving networkliveness: %v", err)
		} else {
			logger.Printf("updated networkliveness for epoch %v", head.HeadEpoch)
			prevHeadEpoch = head.HeadEpoch
		}

		time.Sleep(slotDuration)
	}
}

func genesisDepositsExporter(client rpc.Client) {
	for {
		// check if the beaconchain has started
		var latestEpoch uint64
		err := db.WriterDb.Get(&latestEpoch, "SELECT COALESCE(MAX(epoch), 0) FROM epochs")
		if err != nil {
			logger.Errorf("error retrieving latest epoch from the database: %v", err)
			time.Sleep(time.Second * 10)
			continue
		}

		if latestEpoch == 0 {
			time.Sleep(time.Second * 60)
			continue
		}

		// check if genesis-deposits have already been exported
		var genesisDepositsCount uint64
		err = db.WriterDb.Get(&genesisDepositsCount, "SELECT COUNT(*) FROM blocks_deposits WHERE block_slot=0")
		if err != nil {
			logger.Errorf("error retrieving genesis-deposits-count when exporting genesis-deposits: %v", err)
			time.Sleep(time.Second * 60)
			continue
		}

		// if genesis-deposits have already been exported exit this go-routine
		if genesisDepositsCount > 0 {
			return
		}

		genesisValidators, err := client.GetValidatorState(0)
		if err != nil {
			logger.Errorf("error retrieving genesis validator data for genesis-epoch when exporting genesis-deposits: %v", err)
			time.Sleep(time.Second * 60)
			continue
		}

		tx, err := db.WriterDb.Beginx()
		if err != nil {
			logger.Errorf("error beginning db-tx when exporting genesis-deposits: %v", err)
			time.Sleep(time.Second * 60)
			continue
		}

		for _, validator := range genesisValidators.Data {
			logger.Infof("exporting deposit data for genesis validator %v", validator.Index)
			_, err = tx.Exec(`INSERT INTO blocks_deposits (block_slot, block_root, block_index, publickey, withdrawalcredentials, amount, signature)
			VALUES (0, '\x01', $1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`,
				validator.Index, utils.MustParseHex(validator.Validator.Pubkey), utils.MustParseHex(validator.Validator.WithdrawalCredentials), validator.Balance, []byte{0x0},
			)
			if err != nil {
				tx.Rollback()
				logger.Errorf("error exporting genesis-deposits: %v", err)
				time.Sleep(time.Second * 60)
				continue
			}
		}

		// hydrate the eth1 deposit signature for all genesis validators that have a corresponding eth1 deposit
		_, err = tx.Exec(`
				INSERT INTO blocks_deposits (block_slot, block_index, publickey, withdrawalcredentials, amount, signature)
				SELECT
					0 as block_slot,
					v.validatorindex as block_index,
					v.pubkey as publickey,
					v.withdrawalcredentials,
					b.balance as amount,
					d.signature as signature
				FROM validators v
				LEFT JOIN validator_balances_recent b 
					ON v.validatorindex = b.validatorindex
					AND b.epoch = 0
				LEFT JOIN ( 
					SELECT DISTINCT ON (publickey) publickey, signature FROM eth1_deposits 
				) d ON d.publickey = v.pubkey
				WHERE v.validatorindex < $1 AND d.signature IS NOT NULL
				ON CONFLICT (block_slot, block_index) DO UPDATE SET signature = EXCLUDED.signature`, len(genesisValidators.Data))
		if err != nil {
			tx.Rollback()
			logger.Errorf("error hydrating eth1 data into genesis-deposits: %v", err)
			time.Sleep(time.Second * 60)
			continue
		}
		// update deposits-count
		_, err = tx.Exec("UPDATE blocks SET depositscount = $1 WHERE slot = 0", len(genesisValidators.Data))
		if err != nil {
			tx.Rollback()
			logger.Errorf("error updating deposit count for the genesis slot: %v", err)
			time.Sleep(time.Second * 60)
			continue
		}

		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			logger.Errorf("error committing db-tx when exporting genesis-deposits: %v", err)
			time.Sleep(time.Second * 60)
			continue
		}

		logger.Infof("exported genesis-deposits for %v genesis-validators", len(genesisValidators.Data))
		return
	}
}
