package exporter

import (
	"bytes"
	"eth2-exporter/db"
	"eth2-exporter/rpc"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"sort"
	"strings"
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
	go eth1DepositsExporter()
	go genesisDepositsExporter()

	// wait until the beacon-node is available
	for {
		_, err := client.GetChainHead()
		if err == nil {
			break
		}
		logger.Errorf("beacon-node seems to be unavailable: %v", err)
		time.Sleep(time.Second * 10)
	}

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

		if len(epochs) > 0 && epochs[0] != 0 {
			err := ExportEpoch(0, client)
			if err != nil {
				logger.Error(err)
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

	newBlockChan := client.GetNewBlockChan()

	lastExportedSlot := uint64(0)
	for {
		select {
		case block := <-newBlockChan:
			// Do a full check on any epoch transition or after during the first run
			if utils.EpochOfSlot(lastExportedSlot) != utils.EpochOfSlot(block.Slot) || utils.EpochOfSlot(block.Slot) == 0 {
				doFullCheck(client)
			} else { // else just save the in epoch block
				err := db.SaveBlock(block)
				if err != nil {
					logger.Errorf("error saving block: %v", err)
				}
			}
			lastExportedSlot = block.Slot
		}
	}

	return nil
}

// Will ensure the db is fully in sync with the node
func doFullCheck(client rpc.Client) {
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

	logger.Infof("marking orphaned blocks of epochs %v-%v", startEpoch, head.HeadEpoch)
	err = MarkOrphanedBlocks(startEpoch, head.HeadEpoch, nodeBlocks)
	if err != nil {
		logger.Errorf("error marking orphaned blocks: %v", err)
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

	// Check if the partition for the validator_balances and attestation_assignments table for this epoch exists
	var one int
	logger.Printf("checking partition status for epoch %v", epoch)
	week := epoch / 1575
	err := db.DB.Get(&one, fmt.Sprintf("SELECT 1 FROM information_schema.tables WHERE table_name = 'attestation_assignments_%v'", week))
	if err != nil {
		logger.Infof("creating partition attestation_assignments_%v", week)
		_, err := db.DB.Exec(fmt.Sprintf("CREATE TABLE attestation_assignments_%v PARTITION OF attestation_assignments_p FOR VALUES IN (%v);", week, week))
		if err != nil {
			logger.Fatalf("unable to create partition attestation_assignments_%v: %v", week, err)
		}
	}
	err = db.DB.Get(&one, fmt.Sprintf("SELECT 1 FROM information_schema.tables WHERE table_name = 'validator_balances_%v'", week))
	if err != nil {
		logger.Infof("creating partition validator_balances_%v", week)
		_, err := db.DB.Exec(fmt.Sprintf("CREATE TABLE validator_balances_%v PARTITION OF validator_balances_p FOR VALUES IN (%v);", week, week))
		if err != nil {
			logger.Fatalf("unable to create partition validator_balances_%v: %v", week, err)
		}
	}

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
		logger.Info("updating validator performance data")
		err := updateValidatorPerformance()

		if err != nil {
			logger.Errorf("error updating validator performance data: %v", err)
		} else {
			logger.Info("validator performance data update completed")
		}
		time.Sleep(time.Hour)
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

	var currentEpoch int64

	err = tx.Get(&currentEpoch, "SELECT MAX(epoch) FROM epochs")
	if err != nil {
		return fmt.Errorf("error retrieving latest epoch: %w", err)
	}

	lastDayEpoch := currentEpoch - 225
	lastWeekEpoch := currentEpoch - 225*7
	lastMonthEpoch := currentEpoch - 225*31

	if lastDayEpoch < 0 {
		lastDayEpoch = 0
	}
	if lastWeekEpoch < 0 {
		lastWeekEpoch = 0
	}
	if lastMonthEpoch < 0 {
		lastMonthEpoch = 0
	}

	var balances []types.Validator
	err = tx.Select(&balances, `
		SELECT 
			   validatorindex,
			   pubkey,
       		   activationepoch,
		       COALESCE(balance, 0) AS balance, 
			   COALESCE(balanceactivation, 0) AS balanceactivation, 
			   COALESCE(balance1d, 0) AS balance1d, 
			   COALESCE(balance7d, 0) AS balance7d, 
			   COALESCE(balance31d , 0) AS balance31d
		FROM validators`)
	if err != nil {
		return fmt.Errorf("error retrieving validator performance data: %w", err)
	}

	deposits := []struct {
		Publickey []byte
		Epoch     int64
		Amount    int64
	}{}

	err = tx.Select(&deposits, `SELECT block_slot / 32 AS epoch, amount, publickey FROM blocks_deposits`)
	if err != nil {
		return fmt.Errorf("error retrieving validator deposits data: %w", err)
	}

	depositsMap := make(map[string]map[int64]int64)
	for _, d := range deposits {
		if _, exists := depositsMap[fmt.Sprintf("%x", d.Publickey)]; !exists {
			depositsMap[fmt.Sprintf("%x", d.Publickey)] = make(map[int64]int64)
		}
		depositsMap[fmt.Sprintf("%x", d.Publickey)][d.Epoch] += d.Amount
	}

	data := make([]*types.ValidatorPerformance, 0, len(balances))

	for _, balance := range balances {

		var earningsTotal int64
		var earningsLastDay int64
		var earningsLastWeek int64
		var earningsLastMonth int64
		var totalDeposits int64

		if int64(balance.ActivationEpoch) < currentEpoch {
			for epoch, deposit := range depositsMap[fmt.Sprintf("%x", balance.PublicKey)] {
				totalDeposits += deposit

				if epoch > int64(balance.ActivationEpoch) {
					earningsTotal -= deposit
				}
				if epoch > lastDayEpoch && epoch >= int64(balance.ActivationEpoch) {
					earningsLastDay -= deposit
				}
				if epoch > lastWeekEpoch && epoch >= int64(balance.ActivationEpoch) {
					earningsLastWeek -= deposit
				}
				if epoch > lastMonthEpoch && epoch >= int64(balance.ActivationEpoch) {
					earningsLastMonth -= deposit
				}
			}

			if balance.Balance1d == 0 {
				balance.Balance1d = balance.BalanceActivation
			}
			if balance.Balance7d == 0 {
				balance.Balance7d = balance.BalanceActivation
			}
			if balance.Balance31d == 0 {
				balance.Balance31d = balance.BalanceActivation
			}

			earningsTotal += int64(balance.Balance) - int64(balance.BalanceActivation)
			earningsLastDay += int64(balance.Balance) - int64(balance.Balance1d)
			earningsLastWeek += int64(balance.Balance) - int64(balance.Balance7d)
			earningsLastMonth += int64(balance.Balance) - int64(balance.Balance31d)
		}

		data = append(data, &types.ValidatorPerformance{
			Rank:            0,
			Index:           balance.Index,
			PublicKey:       nil,
			Name:            "",
			Balance:         balance.Balance,
			Performance1d:   earningsLastDay,
			Performance7d:   earningsLastWeek,
			Performance31d:  earningsLastMonth,
			Performance365d: earningsTotal,
		})
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].Performance7d > data[j].Performance7d
	})

	batchSize := 5000

	rank7d := 0
	for b := 0; b < len(data); b += batchSize {

		start := b
		end := b + batchSize
		if len(data) < end {
			end = len(data)
		}

		valueStrings := make([]string, 0, batchSize)
		valueArgs := make([]interface{}, 0, batchSize*7)

		for i, d := range data[start:end] {
			rank7d++

			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*7+1, i*7+2, i*7+3, i*7+4, i*7+5, i*7+6, i*7+7))
			valueArgs = append(valueArgs, d.Index)
			valueArgs = append(valueArgs, d.Balance)
			valueArgs = append(valueArgs, d.Performance1d)
			valueArgs = append(valueArgs, d.Performance7d)
			valueArgs = append(valueArgs, d.Performance31d)
			valueArgs = append(valueArgs, d.Performance365d)
			valueArgs = append(valueArgs, rank7d)
		}

		stmt := fmt.Sprintf(`		
			INSERT INTO validator_performance (validatorindex, balance, performance1d, performance7d, performance31d, performance365d, rank7d)
			VALUES %s`, strings.Join(valueStrings, ","))

		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			return err
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

func genesisDepositsExporter() {
	for {
		// check if the beaconchain has started
		var latestEpoch uint64
		err := db.DB.Get(&latestEpoch, "SELECT COALESCE(MAX(epoch), 0) FROM epochs")
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
		err = db.DB.Get(&genesisDepositsCount, "SELECT COUNT(*) FROM blocks_deposits WHERE block_slot=0")
		if err != nil {
			logger.Errorf("error retrieving genesis-deposits-count when exporting genesis-deposits: %v", err)
			time.Sleep(time.Second * 60)
			continue
		}

		// if genesis-deposits have already been exported exit this go-routine
		if genesisDepositsCount > 0 {
			return
		}

		// get genesis-validators-count
		var genesisValidatorsCount uint64
		err = db.DB.Get(&genesisValidatorsCount, "SELECT validatorscount FROM epochs WHERE epoch=0")
		if err != nil {
			logger.Errorf("error retrieving validatorscount for genesis-epoch when exporting genesis-deposits: %v", err)
			time.Sleep(time.Second * 60)
			continue
		}

		// check if eth1-deposits have already been exported
		var missingEth1Deposits uint64
		err = db.DB.Get(&missingEth1Deposits, `
			SELECT COUNT(*)
			FROM validators v
			LEFT JOIN ( 
				SELECT DISTINCT ON (publickey) publickey, signature FROM eth1_deposits 
			) d ON d.publickey = v.pubkey
			WHERE d.publickey IS NULL AND v.validatorindex < $1`, genesisValidatorsCount)
		if err != nil {
			logger.Errorf("error retrieving missing-eth1-deposits-count when exporting genesis-deposits")
			time.Sleep(time.Second * 60)
			continue
		}

		if missingEth1Deposits > 0 {
			logger.Infof("delaying export of genesis-deposits until eth1-deposits have been exported")
			time.Sleep(time.Second * 60)
			continue
		}

		tx, err := db.DB.Beginx()
		if err != nil {
			logger.Errorf("error beginning db-tx when exporting genesis-deposits: %v", err)
			time.Sleep(time.Second * 60)
			continue
		}

		// export genesis-deposits from eth1-deposits and data already gathered from the eth2-client
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
				LEFT JOIN validator_balances_p b 
					ON v.validatorindex = b.validatorindex
					AND b.epoch = 0
					AND b.week = 0
				LEFT JOIN ( 
					SELECT DISTINCT ON (publickey) publickey, signature FROM eth1_deposits 
				) d ON d.publickey = v.pubkey
				WHERE v.validatorindex < $1`, genesisValidatorsCount)
		if err != nil {
			tx.Rollback()
			logger.Errorf("error exporting genesis-deposits: %v", err)
			time.Sleep(time.Second * 60)
			continue
		}

		// update deposits-count
		_, err = tx.Exec("UPDATE blocks SET depositscount = $1 WHERE slot = 0", genesisValidatorsCount)
		if err != nil {
			tx.Rollback()
			logger.Errorf("error exporting genesis-deposits: %v", err)
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

		logger.Infof("exported genesis-deposits for %v genesis-validators", genesisValidatorsCount)
		return
	}
}
