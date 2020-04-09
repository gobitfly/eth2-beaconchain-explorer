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
	go performanceDataUpdater()
	go networkLivenessUpdater(client)

	if utils.Config.Indexer.FullIndexOnStartup {
		logger.Printf("performing one time full db reindex")
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
					logger.Printf("finished export for epoch %v", epoch)
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
				logger.Printf("queuing epoch %v for export as block %v is present on the node but missing in the db", block.Epoch, key)
				epochsToExport[block.Epoch] = true
			} else if block.Node == nil {
				logger.Printf("queuing epoch %v for export as block %v is present on the db but missing in the node", block.Epoch, key)
				epochsToExport[block.Epoch] = true
			} else if bytes.Compare(block.Db.BlockRoot, block.Node.BlockRoot) != 0 {
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
			logger.Fatal(err)
		}
		startEpoch := uint64(0)
		err = updateEpochStatus(client, startEpoch, head.HeadEpoch)
		if err != nil {
			logger.Fatal(err)
		}
	}

	for true {
		time.Sleep(time.Second * 10)
		logger.Infof("checking for new blocks/epochs to export")

		head, err := client.GetChainHead()
		if err != nil {
			logger.Errorf("error retrieving chain head: %v", err)
			continue
		}

		startEpoch := uint64(0)
		if head.FinalizedEpoch > 1 {
			startEpoch = head.FinalizedEpoch - 1
		}

		dbBlocks, err := db.GetLastPendingAndProposedBlocks(startEpoch, head.HeadEpoch)
		if err != nil {
			logger.Errorf("error retrieving last pending and proposed blocks from the database: %v", err)
			continue
		}

		nodeBlocks, err := GetLastBlocks(startEpoch, head.HeadEpoch, client)
		if err != nil {
			logger.Errorf("error retrieving last blocks from backend node: %v", err)
			continue
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
			} else if block.Node == nil {
				logger.Printf("queuing epoch %v for export as block %v is present on the db but missing in the node", block.Epoch, key)
				epochsToExport[block.Epoch] = true
			} else if bytes.Compare(block.Db.BlockRoot, block.Node.BlockRoot) != 0 {
				logger.Printf("queuing epoch %v for export as block %v has a different hash in the db as on the node", block.Epoch, key)
				epochsToExport[block.Epoch] = true
			}
		}

		// Add any missing epoch to the export set (might happen if the indexer was stopped for a long period of time)
		epochs, err := db.GetAllEpochs()
		if err != nil {
			if err != nil {
				logger.Errorf("error retrieving all epochs from the db: %v", err)
				continue
			}
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
		}

		// Update epoch statistics up to 10 epochs after the last finalized epoch
		startEpoch = uint64(0)
		if head.FinalizedEpoch > 10 {
			startEpoch = head.FinalizedEpoch - 10
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

		logger.Infof("marking orphaned blocks of epochs %v-%v", startEpoch, head.HeadEpoch)
		err = MarkOrphanedBlocks(startEpoch, head.HeadEpoch, nodeBlocks)
		if err != nil {
			logger.Errorf("error marking orphaned blocks: %v", err)
		}

		logger.Infof("finished exporting all new blocks/epochs")
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
			logger.Errorf("block %x at slot %v in epoch %v has been orphaned", blocks[i].BlockRoot, blocks[i].Slot, blocks[i].Epoch)
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
				return nil, err
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

		logger.Printf("retrieving all blocks for epoch %v. %v epochs remaining", epoch, endEpoch-epoch)
	}

	return wrappedBlocks, nil
}

// ExportEpoch will export an epoch from rpc into the database
func ExportEpoch(epoch uint64, client rpc.Client) error {
	start := time.Now()

	logger.Printf("retrieving data for epoch %v", epoch)
	data, err := client.GetEpochData(epoch)

	if err != nil {
		return fmt.Errorf("error retrieving epoch data: %v", err)
	}

	logger.Printf("data for epoch %v retrieved, took %v", epoch, time.Since(start))

	if len(data.Validators) == 0 {
		return fmt.Errorf("error retrieving epoch data: no validators received for epoch")
	}

	return db.SaveEpoch(data)
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
			logger.Printf("updating epoch %v with status finalized = %v", epoch, epochParticipationStats.Finalized)
			err := db.UpdateEpochStatus(epochParticipationStats)

			if err != nil {
				return err
			}
		}
	}
	return nil
}

func performanceDataUpdater() {
	for true {
		time.Sleep(time.Hour)
		logger.Info("updating validator performance data")
		err := updateValidatorPerformance()

		if err != nil {
			logger.Errorf("error updating validator performance data: %w", err)
		} else {
			logger.Info("validator performance data update completed")
		}
	}
}

func updateValidatorPerformance() error {
	tx, err := db.DB.Beginx()
	if err != nil {
		return fmt.Errorf("error starting db transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec("TRUNCATE validator_performance")
	if err != nil {
		return fmt.Errorf("error truncating validator performance table: %w", err)
	}

	var currentEpoch uint64

	err = tx.Get(&currentEpoch, "SELECT MAX(epoch) FROM validator_balances")
	if err != nil {
		return fmt.Errorf("error retrieving latest epoch from validator_balances table: %w", err)
	}

	now := utils.EpochToTime(currentEpoch)
	epoch1d := utils.TimeToEpoch(now.Add(time.Hour * 24 * -1))
	epoch7d := utils.TimeToEpoch(now.Add(time.Hour * 24 * 7 * -1))
	epoch31d := utils.TimeToEpoch(now.Add(time.Hour * 24 * 31 * -1))
	epoch365d := utils.TimeToEpoch(now.Add(time.Hour * 24 * 356 * -1))

	if epoch1d < 0 {
		epoch1d = 0
	}
	if epoch7d < 0 {
		epoch7d = 0
	}
	if epoch31d < 0 {
		epoch31d = 0
	}
	if epoch365d < 0 {
		epoch365d = 0
	}

	var startBalances []*types.ValidatorBalance
	err = tx.Select(&startBalances, `
		SELECT 
			validator_balances.validatorindex,
			validator_balances.balance
		FROM validators
			LEFT JOIN validator_balances
				ON validators.activationepoch = validator_balances.epoch
				AND validators.validatorindex = validator_balances.validatorindex
		WHERE validator_balances.validatorindex IS NOT NULL`)
	if err != nil {
		return fmt.Errorf("error retrieving initial validator balances data: %w", err)
	}

	startBalanceMap := make(map[uint64]uint64)
	for _, balance := range startBalances {
		startBalanceMap[balance.Index] = balance.Balance
	}

	var balances []*types.ValidatorBalance
	err = tx.Select(&balances, `
		SELECT
			validator_balances.epoch,
			validator_balances.validatorindex,
			validator_balances.balance
		FROM validator_balances
		WHERE validator_balances.epoch IN ($1, $2, $3, $4, $5)`,
		currentEpoch, epoch1d, epoch7d, epoch31d, epoch365d)
	if err != nil {
		return fmt.Errorf("error retrieving validator performance data: %w", err)
	}

	type depositByEpochRange struct {
		Index        uint64 `db:"validatorindex"`
		EpochRange   uint64 `db:"epochrange"`
		DepositTotal uint64 `db:"deposittotal"`
	}

	// get total deposit-amounts from specific epochs up to the current epoch
	var depositsByEpochRange []*depositByEpochRange
	err = tx.Select(&depositsByEpochRange, `
		SELECT
			validatorindex,
			epochrange,
			MAX(deposittotal) as deposittotal
		FROM 
		(
			SELECT DISTINCT
				validatorindex,
				CASE
					WHEN (d.block_slot/32)-1 <= $5 THEN $5
					WHEN (d.block_slot/32)-1 <= $4 THEN $4
					WHEN (d.block_slot/32)-1 <= $3 THEN $3
					WHEN (d.block_slot/32)-1 <= $2 THEN $2
					ELSE $1
				END AS epochrange,
				SUM(d.amount) OVER (
					PARTITION BY d.publickey 
					ORDER BY d.block_slot DESC
				) AS deposittotal
			FROM validators
				INNER JOIN blocks_deposits d
					ON d.publickey = validators.pubkey
					AND (d.block_slot/32) > validators.activationepoch
		) a
		GROUP BY epochrange, validatorindex`,
		currentEpoch, epoch1d, epoch7d, epoch31d, epoch365d)
	if err != nil {
		return fmt.Errorf("error retrieving validator deposits data: %w", err)
	}

	depositsMap := make(map[uint64]map[int64]int64)
	for _, deposit := range depositsByEpochRange {
		if _, exists := depositsMap[deposit.Index]; !exists {
			depositsMap[deposit.Index] = make(map[int64]int64)
		}
		depositsMap[deposit.Index][int64(deposit.EpochRange)] = int64(deposit.DepositTotal)
	}

	performance := make(map[uint64]map[int64]int64)
	for _, balance := range balances {
		if performance[balance.Index] == nil {
			performance[balance.Index] = make(map[int64]int64)
		}
		performance[balance.Index][int64(balance.Epoch)] = int64(balance.Balance)
	}

	for validator, balances := range performance {

		currentBalance := balances[int64(currentEpoch)]
		startBalance := int64(startBalanceMap[validator])

		if currentBalance == 0 || startBalance == 0 {
			continue
		}

		balance1d := balances[epoch1d]
		if balance1d == 0 {
			balance1d = startBalance
		}
		balance7d := balances[epoch7d]
		if balance7d == 0 {
			balance7d = startBalance
		}
		balance31d := balances[epoch31d]
		if balance31d == 0 {
			balance31d = startBalance
		}
		balance365d := balances[epoch365d]
		if balance365d == 0 {
			balance365d = startBalance
		}

		performance1d := currentBalance - balance1d
		performance7d := currentBalance - balance7d
		performance31d := currentBalance - balance31d
		performance365d := currentBalance - balance365d

		if depositsMap[validator] != nil {
			if d, exists := depositsMap[validator][epoch1d]; exists {
				performance1d -= d
			}

			if d, exists := depositsMap[validator][epoch7d]; exists {
				performance7d -= d
			} else if d, exists := depositsMap[validator][epoch1d]; exists {
				performance7d -= d
			}

			if d, exists := depositsMap[validator][epoch31d]; exists {
				performance31d -= d
			} else if d, exists := depositsMap[validator][epoch7d]; exists {
				performance31d -= d
			} else if d, exists := depositsMap[validator][epoch1d]; exists {
				performance31d -= d
			}

			if d, exists := depositsMap[validator][epoch365d]; exists {
				performance365d -= d
			} else if d, exists := depositsMap[validator][epoch31d]; exists {
				performance365d -= d
			} else if d, exists := depositsMap[validator][epoch7d]; exists {
				performance365d -= d
			} else if d, exists := depositsMap[validator][epoch1d]; exists {
				performance365d -= d
			}
		}

		if performance1d > 10000000 {
			performance1d = 0
		}
		if performance7d > 10000000*7 {
			performance7d = 0
		}
		if performance31d > 10000000*31 {
			performance31d = 0
		}
		if performance365d > 10000000*365 {
			performance365d = 0
		}

		_, err := tx.Exec(`
			INSERT INTO validator_performance (validatorindex, balance, performance1d, performance7d, performance31d, performance365d)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			validator, currentBalance, performance1d, performance7d, performance31d, performance365d)

		if err != nil {
			return fmt.Errorf("error saving validator performance data: %w", err)
		}
	}

	return tx.Commit()
}

func networkLivenessUpdater(client rpc.Client) {
	var prevHeadEpoch uint64
	err := db.DB.Get(&prevHeadEpoch, "SELECT COALESCE(MAX(headepoch), 0) FROM network_liveness")
	if err != nil {
		logger.Fatal(err)
	}

	epochDuration := time.Second * time.Duration(utils.Config.Chain.SecondsPerSlot*utils.Config.Chain.SlotsPerEpoch)
	slotDuration := time.Second * time.Duration(utils.Config.Chain.SecondsPerSlot)

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

		_, err = db.DB.Exec(`
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
