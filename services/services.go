package services

import (
	"database/sql"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

var latestEpoch uint64
var latestFinalizedEpoch uint64
var latestSlot uint64
var latestProposedSlot uint64
var indexPageData atomic.Value
var chartsPageData atomic.Value
var ready = sync.WaitGroup{}

var latestStats atomic.Value

var eth1BlockDepositReached atomic.Value
var depositThresholdReached atomic.Value

var logger = logrus.New().WithField("module", "services")

// Init will initialize the services
func Init() {
	ready.Add(4)
	go epochUpdater()
	go slotUpdater()
	go latestProposedSlotUpdater()
	go indexPageDataUpdater()
	ready.Wait()

	go chartsPageDataUpdater()
	go statsUpdater()

	if utils.Config.Frontend.Notifications.Enabled {
		logger.Infof("starting notifications-sender")
		go notificationsSender()
	}
}

func epochUpdater() {
	firstRun := true

	for true {

		var latestFinalized uint64
		err := db.DB.Get(&latestFinalized, "SELECT COALESCE(MAX(epoch), 0) FROM epochs where finalized is true")
		if err != nil {
			logger.Errorf("error retrieving latest finalized epoch from the database: %v", err)
		} else {
			atomic.StoreUint64(&latestFinalizedEpoch, latestFinalized)
		}

		var epoch uint64
		err = db.DB.Get(&epoch, "SELECT COALESCE(MAX(epoch), 0) FROM epochs")
		if err != nil {
			logger.Errorf("error retrieving latest epoch from the database: %v", err)
		} else {
			atomic.StoreUint64(&latestEpoch, epoch)
			if firstRun {
				ready.Done()
				firstRun = false
			}
		}
		time.Sleep(time.Second)
	}
}

func slotUpdater() {
	firstRun := true

	for true {
		var slot uint64
		err := db.DB.Get(&slot, "SELECT COALESCE(MAX(slot), 0) FROM blocks where slot < $1", utils.TimeToSlot(uint64(time.Now().Add(time.Second*10).Unix())))

		if err != nil {
			logger.Errorf("error retrieving latest slot from the database: %v", err)
		} else {
			atomic.StoreUint64(&latestSlot, slot)
			if firstRun {
				ready.Done()
				firstRun = false
			}
		}
		time.Sleep(time.Second)
	}
}

func latestProposedSlotUpdater() {
	firstRun := true

	for true {
		var slot uint64
		err := db.DB.Get(&slot, "SELECT COALESCE(MAX(slot), 0) FROM blocks WHERE status = '1'")

		if err != nil {
			logger.Errorf("error retrieving latest proposed slot from the database: %v", err)
		} else {
			atomic.StoreUint64(&latestProposedSlot, slot)
			if firstRun {
				ready.Done()
				firstRun = false
			}
		}
		time.Sleep(time.Second)
	}
}

func indexPageDataUpdater() {
	firstRun := true

	for true {
		data, err := getIndexPageData()
		if err != nil {
			logger.Errorf("error retrieving index page data: %v", err)
			time.Sleep(time.Second * 10)
			continue
		}
		indexPageData.Store(data)
		if firstRun {
			ready.Done()
			firstRun = false
		}
		time.Sleep(time.Second * 10)
	}
}

func getIndexPageData() (*types.IndexPageData, error) {
	data := &types.IndexPageData{}

	data.NetworkName = utils.Config.Chain.Network
	data.DepositContract = utils.Config.Indexer.Eth1DepositContractAddress

	var epoch uint64
	err := db.DB.Get(&epoch, "SELECT COALESCE(MAX(epoch), 0) FROM epochs")
	if err != nil {
		return nil, fmt.Errorf("error retrieving latest epoch from the database: %v", err)
	}
	data.CurrentEpoch = epoch

	cutoffSlot := utils.TimeToSlot(uint64(time.Now().Add(time.Second * 10).Unix()))

	// If we are before the genesis block show the first 20 slots by default
	startSlotTime := utils.SlotToTime(0)
	genesisTransition := utils.SlotToTime(160)
	now := time.Now()

	// run deposit query until the Genesis period is over
	if now.Before(genesisTransition) {
		if cutoffSlot < 15 {
			cutoffSlot = 15
		}
		type Deposit struct {
			Total   uint64    `db:"total"`
			BlockTs time.Time `db:"block_ts"`
		}

		deposit := Deposit{}

		err = db.DB.Get(&deposit, `
			SELECT COUNT(*) as total, COALESCE(MAX(block_ts),NOW()) as block_ts
			FROM 
				eth1_deposits as eth1 
			WHERE 
				eth1.amount >= 32e9 and eth1.valid_signature = true;
		`)
		if err != nil {
			return nil, fmt.Errorf("error retrieving eth1 deposits: %v", err)
		}

		data.DepositThreshold = float64(utils.Config.Chain.MinGenesisActiveValidatorCount) * 32
		data.DepositedTotal = float64(deposit.Total) * 32
		data.ValidatorsRemaining = (data.DepositThreshold - data.DepositedTotal) / 32

		minGenesisTime := time.Unix(int64(utils.Config.Chain.GenesisTimestamp), 0)
		data.NetworkStartTs = minGenesisTime.Unix()

		// enough deposits
		if data.DepositedTotal > data.DepositThreshold {
			if depositThresholdReached.Load() == nil {
				eth1BlockDepositReached.Store(deposit.BlockTs)
				depositThresholdReached.Store(true)
			}
			eth1Block := eth1BlockDepositReached.Load().(time.Time)
			genesisDelay := time.Duration(int64(utils.Config.Chain.GenesisDelay))

			if eth1Block.Add(genesisDelay).After(minGenesisTime) {
				// Network starts after min genesis time
				data.NetworkStartTs = eth1Block.Add(genesisDelay).Unix()
			}
		}
	}
	// has genesis occured
	if now.After(startSlotTime) {
		data.Genesis = true
	} else {
		data.Genesis = false
	}
	// show the transition view one hour before the first slot and until epoch 30 is reached
	if now.Add(time.Hour*2).After(startSlotTime) && now.Before(genesisTransition) {
		data.GenesisPeriod = true
	} else {
		data.GenesisPeriod = false
	}

	var epochs []*types.IndexPageDataEpochs
	err = db.DB.Select(&epochs, `SELECT epoch, finalized , eligibleether, globalparticipationrate, votedether FROM epochs ORDER BY epochs DESC LIMIT 15`)
	if err != nil {
		return nil, fmt.Errorf("error retrieving index epoch data: %v", err)
	}

	for _, epoch := range epochs {
		epoch.Ts = utils.EpochToTime(epoch.Epoch)
		epoch.FinalizedFormatted = utils.FormatYesNo(epoch.Finalized)
		epoch.VotedEtherFormatted = utils.FormatBalance(epoch.VotedEther)
		epoch.EligibleEtherFormatted = utils.FormatBalance(epoch.EligibleEther)
		epoch.GlobalParticipationRateFormatted = utils.FormatGlobalParticipationRate(epoch.VotedEther, epoch.GlobalParticipationRate)
	}
	data.Epochs = epochs

	var scheduledCount uint8
	err = db.DB.Get(&scheduledCount, `
		select count(*) from blocks where status = '0' and epoch = (select max(epoch) from blocks limit 1);
	`)
	if err != nil {
		return nil, fmt.Errorf("error retrieving scheduledCount from blocks: %v", err)
	}
	data.ScheduledCount = scheduledCount

	var blocks []*types.IndexPageDataBlocks
	err = db.DB.Select(&blocks, `
		SELECT
			blocks.epoch,
			blocks.slot,
			blocks.proposer,
			blocks.blockroot,
			blocks.parentroot,
			blocks.attestationscount,
			blocks.depositscount,
			blocks.voluntaryexitscount,
			blocks.proposerslashingscount,
			blocks.attesterslashingscount,
			blocks.status,
			COALESCE(validators.name, '') AS name
		FROM blocks 
		LEFT JOIN validators ON blocks.proposer = validators.validatorindex
		WHERE blocks.slot < $1
		ORDER BY blocks.slot DESC LIMIT 15`, cutoffSlot)
	if err != nil {
		return nil, fmt.Errorf("error retrieving index block data: %v", err)
	}
	data.Blocks = blocks

	for _, block := range data.Blocks {
		block.StatusFormatted = utils.FormatBlockStatus(block.Status)
		block.ProposerFormatted = utils.FormatValidatorWithName(block.Proposer, block.ProposerName)
		block.BlockRootFormatted = fmt.Sprintf("%x", block.BlockRoot)
	}

	// if len(blocks) > 0 {
	// 	data.CurrentSlot = blocks[0].Slot
	// }
	if data.GenesisPeriod {
		for _, blk := range blocks {
			if blk.Status != 0 {
				data.CurrentSlot = blk.Slot
			}
		}
	} else if len(blocks) > 0 {
		data.CurrentSlot = blocks[0].Slot
	}

	for _, block := range data.Blocks {
		block.Ts = utils.SlotToTime(block.Slot)
	}

	queueCount := struct {
		EnteringValidators uint64 `db:"entering_validators_count"`
		ExitingValidators  uint64 `db:"exiting_validators_count"`
	}{}

	err = db.DB.Get(&queueCount, "SELECT entering_validators_count, exiting_validators_count FROM queue ORDER BY ts DESC LIMIT 1")
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error retrieving validator queue count: %v", err)
	}

	data.EnteringValidators = queueCount.EnteringValidators
	data.ExitingValidators = queueCount.ExitingValidators

	var averageBalance float64
	err = db.DB.Get(&averageBalance, "SELECT COALESCE(AVG(balance), 0) FROM validator_balances WHERE epoch = $1", epoch)
	if err != nil {
		return nil, fmt.Errorf("error retrieving validator balance: %v", err)
	}
	data.AverageBalance = string(utils.FormatBalance(uint64(averageBalance)))

	var epochHistory []*types.IndexPageEpochHistory
	err = db.DB.Select(&epochHistory, "SELECT epoch, eligibleether, validatorscount, finalized FROM epochs WHERE epoch < $1 ORDER BY epoch", epoch)
	if err != nil {
		return nil, fmt.Errorf("error retrieving staked ether history: %v", err)
	}

	if len(epochHistory) > 0 {
		for i := len(epochHistory) - 1; i >= 0; i-- {
			if epochHistory[i].Finalized {
				data.CurrentFinalizedEpoch = epochHistory[i].Epoch
				data.FinalityDelay = data.CurrentEpoch - epoch
				break
			}
		}

		data.StakedEther = string(utils.FormatBalance(epochHistory[len(epochHistory)-1].EligibleEther))
		data.ActiveValidators = epochHistory[len(epochHistory)-1].ValidatorsCount
	}

	data.StakedEtherChartData = make([][]float64, len(epochHistory))
	data.ActiveValidatorsChartData = make([][]float64, len(epochHistory))
	for i, history := range epochHistory {
		data.StakedEtherChartData[i] = []float64{float64(utils.EpochToTime(history.Epoch).Unix() * 1000), float64(history.EligibleEther) / 1000000000}
		data.ActiveValidatorsChartData[i] = []float64{float64(utils.EpochToTime(history.Epoch).Unix() * 1000), float64(history.ValidatorsCount)}
	}

	data.Subtitle = template.HTML(utils.Config.Frontend.SiteSubtitle)

	return data, nil
}

// LatestEpoch will return the latest epoch
func LatestEpoch() uint64 {
	return atomic.LoadUint64(&latestEpoch)
}

// LatestFinalizedEpoch will return the most recent epoch that has been finalized.
func LatestFinalizedEpoch() uint64 {
	return atomic.LoadUint64(&latestFinalizedEpoch)
}

// LatestSlot will return the latest slot
func LatestSlot() uint64 {
	return atomic.LoadUint64(&latestSlot)
}

//FinalizationDelay will return the current Finalization Delay
func FinalizationDelay() uint64 {
	return LatestEpoch() - LatestFinalizedEpoch()
}

// LatestProposedSlot will return the latest proposed slot
func LatestProposedSlot() uint64 {
	return atomic.LoadUint64(&latestProposedSlot)
}

// LatestIndexPageData returns the latest index page data
func LatestIndexPageData() *types.IndexPageData {
	return indexPageData.Load().(*types.IndexPageData)
}

// LatestState returns statistics about the current eth2 state
func LatestState() *types.LatestState {
	data := &types.LatestState{}
	data.CurrentEpoch = LatestEpoch()
	data.CurrentSlot = LatestSlot()
	data.CurrentFinalizedEpoch = LatestFinalizedEpoch()
	data.LastProposedSlot = atomic.LoadUint64(&latestProposedSlot)
	data.FinalityDelay = data.CurrentEpoch - data.CurrentFinalizedEpoch
	data.IsSyncing = IsSyncing()
	return data
}

func GetLatestStats() *types.Stats {
	stats := latestStats.Load()
	if stats == nil {
		// create an empty stats object if no stats exist (genesis)
		return &types.Stats{
			TopDepositors: &[]types.StatsTopDepositors{
				{
					Address:      "000",
					DepositCount: 0,
				},
				{
					Address:      "000",
					DepositCount: 0,
				},
			},
			InvalidDepositCount:  new(uint64),
			UniqueValidatorCount: new(uint64),
		}
	}
	return stats.(*types.Stats)
}

// IsSyncing returns true if the chain is still syncing
func IsSyncing() bool {
	return time.Now().Add(time.Minute * -10).After(utils.EpochToTime(LatestEpoch()))
}
