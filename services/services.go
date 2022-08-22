package services

import (
	"database/sql"
	"eth2-exporter/db"
	"eth2-exporter/price"
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
var latestNodeEpoch uint64
var latestFinalizedEpoch uint64
var latestNodeFinalizedEpoch uint64
var latestSlot uint64
var latestProposedSlot uint64
var latestValidatorCount uint64
var indexPageData atomic.Value
var chartsPageData atomic.Value
var poolsData atomic.Value
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

	if utils.Config.Frontend.OnlyAPI {
		ready.Done()
	} else {
		go poolsUpdater()
	}
	ready.Wait()
	if utils.Config.Frontend.OnlyAPI {
		return
	}
	// we do this after the rest has readied up to ensure we get a complete index page
	ready.Add(1)
	go indexPageDataUpdater()
	ready.Wait()

	if !utils.Config.Frontend.DisableCharts {
		go chartsPageDataUpdater()
	}

	go statsUpdater()
}

func InitNotifications() {
	logger.Infof("starting notifications-sender")
	go notificationsSender()
}

func epochUpdater() {
	firstRun := true

	for {
		// latest epoch acording to the node
		var epochNode uint64
		err := db.WriterDb.Get(&epochNode, "SELECT headepoch FROM network_liveness order by headepoch desc LIMIT 1")
		if err != nil {
			logger.Errorf("error retrieving latest node epoch from the database: %v", err)
		} else {
			atomic.StoreUint64(&latestNodeEpoch, epochNode)
		}

		// latest finalized epoch acording to the node
		var latestNodeFinalized uint64
		err = db.WriterDb.Get(&latestNodeFinalized, "SELECT finalizedepoch FROM network_liveness order by headepoch desc LIMIT 1")
		if err != nil {
			logger.Errorf("error retrieving latest node finalized epoch from the database: %v", err)
		} else {
			atomic.StoreUint64(&latestNodeFinalizedEpoch, latestNodeFinalized)
		}

		// latest exported epoch
		var epoch uint64
		err = db.WriterDb.Get(&epoch, "SELECT COALESCE(MAX(epoch), 0) FROM epochs")
		if err != nil {
			logger.Errorf("error retrieving latest exported epoch from the database: %v", err)
		} else {
			atomic.StoreUint64(&latestEpoch, epoch)
		}

		// latest exportered finalized epoch
		var latestFinalized uint64
		err = db.WriterDb.Get(&latestFinalized, "SELECT COALESCE(MAX(epoch), 0) FROM epochs where epoch <= (select finalizedepoch from network_liveness order by headepoch desc limit 1)")
		if err != nil {
			logger.Errorf("error retrieving latest exported finalized epoch from the database: %v", err)
		} else {
			atomic.StoreUint64(&latestFinalizedEpoch, latestFinalized)
			if firstRun {
				logger.Info("initialized epoch updater")
				ready.Done()
				firstRun = false
			}
		}
		time.Sleep(time.Second)
	}
}

func slotUpdater() {
	firstRun := true

	for {
		var slot uint64
		err := db.WriterDb.Get(&slot, "SELECT COALESCE(MAX(slot), 0) FROM blocks where slot < $1", utils.TimeToSlot(uint64(time.Now().Add(time.Second*10).Unix())))

		if err != nil {
			logger.Errorf("error retrieving latest slot from the database: %v", err)
		} else {
			atomic.StoreUint64(&latestSlot, slot)
			if firstRun {
				logger.Info("initialized slot updater")
				ready.Done()
				firstRun = false
			}
		}
		time.Sleep(time.Second)
	}
}

func poolsUpdater() {
	firstRun := true

	for {
		data, err := getPoolsPageData()
		if err != nil {
			logger.Errorf("error retrieving pools page data: %v", err)
			time.Sleep(time.Second * 10)
			continue
		}
		poolsData.Store(data)
		if firstRun {
			logger.Info("initialized pools page updater")
			ready.Done()
			firstRun = false
		}
		time.Sleep(time.Minute * 10)
	}
}

func getPoolsPageData() (*types.PoolsResp, error) {
	var poolData types.PoolsResp

	err := db.ReaderDb.Select(&poolData.PoolInfos, `
	select 
		coalesce(pool, 'Unknown') as name, 
		count(*) as count, 
		avg(performance31d)::integer as avg_performance_31d, 
		avg(performance7d)::integer as avg_performance_7d, 
		avg(performance1d)::integer as avg_performance_1d 
	from validators 
		left outer join validator_pool on validators.pubkey = validator_pool.publickey 
		left outer join validator_performance on validators.validatorindex = validator_performance.validatorindex 
	where validators.status in ('active_online', 'active_offline') 
	group by pool 
	order by count(*) desc;`)

	if err != nil {
		return nil, err
	}

	return &poolData, nil
}

func latestProposedSlotUpdater() {
	firstRun := true

	for {
		var slot uint64
		err := db.WriterDb.Get(&slot, "SELECT COALESCE(MAX(slot), 0) FROM blocks WHERE status = '1'")

		if err != nil {
			logger.Errorf("error retrieving latest proposed slot from the database: %v", err)
		} else {
			atomic.StoreUint64(&latestProposedSlot, slot)
			if firstRun {
				logger.Info("initialized last proposed slot updater")
				ready.Done()
				firstRun = false
			}
		}
		time.Sleep(time.Second)
	}
}

func indexPageDataUpdater() {
	firstRun := true

	for {
		data, err := getIndexPageData()
		if err != nil {
			logger.Errorf("error retrieving index page data: %v", err)
			time.Sleep(time.Second * 10)
			continue
		}
		indexPageData.Store(data)
		if firstRun {
			logger.Info("initialized index page updater")
			ready.Done()
			firstRun = false
		}
		time.Sleep(time.Second * 10)
	}
}

func getIndexPageData() (*types.IndexPageData, error) {
	var err error
	currency := "ETH"

	data := &types.IndexPageData{}
	data.Mainnet = utils.Config.Chain.Config.ConfigName == "mainnet"
	data.NetworkName = utils.Config.Chain.Config.ConfigName
	data.DepositContract = utils.Config.Indexer.Eth1DepositContractAddress
	data.ShowSyncingMessage = IsSyncing()

	data.CurrentEpoch = LatestEpoch()

	cutoffSlot := utils.TimeToSlot(uint64(time.Now().Add(time.Second * 10).Unix()))

	// If we are before the genesis block show the first 20 slots by default
	startSlotTime := utils.SlotToTime(0)
	genesisTransition := utils.SlotToTime(160)
	now := time.Now()

	// run deposit query until the Genesis period is over
	if now.Before(genesisTransition) || startSlotTime == time.Unix(0, 0) {
		if cutoffSlot < 15 {
			cutoffSlot = 15
		}
		type Deposit struct {
			Total   uint64    `db:"total"`
			BlockTs time.Time `db:"block_ts"`
		}

		deposit := Deposit{}
		err = db.WriterDb.Get(&deposit, `
			SELECT COUNT(*) as total, COALESCE(MAX(block_ts),NOW()) AS block_ts
			FROM (
				SELECT publickey, SUM(amount) AS amount, MAX(block_ts) as block_ts
				FROM eth1_deposits
				WHERE valid_signature = true
				GROUP BY publickey
				HAVING SUM(amount) >= 32e9
			) a`)
		if err != nil {
			return nil, fmt.Errorf("error retrieving eth1 deposits: %v", err)
		}

		threshold, err := db.GetDepositThresholdTime()
		if err != nil {
			logger.WithError(err).Error("error could not calcualte threshold time")
		}
		if threshold == nil {
			threshold = &deposit.BlockTs
		}

		data.DepositThreshold = float64(utils.Config.Chain.Config.MinGenesisActiveValidatorCount) * 32
		data.DepositedTotal = float64(deposit.Total) * 32

		data.ValidatorsRemaining = (data.DepositThreshold - data.DepositedTotal) / 32
		genesisDelay := time.Duration(int64(utils.Config.Chain.Config.GenesisDelay) * 1000 * 1000 * 1000) // convert seconds to nanoseconds

		minGenesisTime := time.Unix(int64(utils.Config.Chain.GenesisTimestamp), 0)

		data.MinGenesisTime = minGenesisTime.Unix()
		data.NetworkStartTs = minGenesisTime.Add(genesisDelay).Unix()

		if minGenesisTime.Before(time.Now()) {
			minGenesisTime = time.Now()
		}

		// enough deposits
		if data.DepositedTotal > data.DepositThreshold {
			if depositThresholdReached.Load() == nil {
				eth1BlockDepositReached.Store(*threshold)
				depositThresholdReached.Store(true)
			}
			eth1Block := eth1BlockDepositReached.Load().(time.Time)

			if !(startSlotTime == time.Unix(0, 0)) && eth1Block.Add(genesisDelay).After(minGenesisTime) {
				// Network starts after min genesis time
				data.NetworkStartTs = eth1Block.Add(genesisDelay).Unix()
			} else {
				data.NetworkStartTs = minGenesisTime.Unix()
			}
		}

		latestChartsPageData := LatestChartsPageData()
		if latestChartsPageData != nil {
			for _, c := range *latestChartsPageData {
				if c.Path == "deposits" {
					data.DepositChart = c
				} else if c.Path == "deposits_distribution" {
					data.DepositDistribution = c
				}
			}
		}
	}
	if data.DepositChart != nil && data.DepositChart.Data != nil && data.DepositChart.Data.Series != nil {
		series := data.DepositChart.Data.Series
		if len(series) > 2 {
			points := series[1].Data.([][]float64)
			periodDays := float64(len(points))
			avgDepositPerDay := data.DepositedTotal / periodDays
			daysUntilThreshold := (data.DepositThreshold - data.DepositedTotal) / avgDepositPerDay
			estimatedTimeToThreshold := time.Now().Add(time.Hour * 24 * time.Duration(daysUntilThreshold))
			if estimatedTimeToThreshold.After(time.Unix(data.NetworkStartTs, 0)) {
				data.NetworkStartTs = estimatedTimeToThreshold.Add(time.Duration(int64(utils.Config.Chain.Config.GenesisDelay) * 1000 * 1000 * 1000)).Unix()
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
	if now.Add(time.Hour*24).After(startSlotTime) && now.Before(genesisTransition) {
		data.GenesisPeriod = true
	} else {
		data.GenesisPeriod = false
	}

	if startSlotTime == time.Unix(0, 0) {
		data.Genesis = false
	}

	var epochs []*types.IndexPageDataEpochs
	currentNodeEpoch := LatestNodeEpoch()
	finalizedNodeEpoch := LatestNodeFinalizedEpoch()

	err = db.WriterDb.Select(&epochs, `
	SELECT 
		epoch, 
		eligibleether, 
		globalparticipationrate, 
		votedether,
		completeparticipationstats
	FROM epochs
	where
		epoch > $1 - 15 and 
		epoch <= $1
	ORDER BY epochs DESC`, currentNodeEpoch)
	if err != nil {
		return nil, fmt.Errorf("error retrieving index epoch data: %v", err)
	}

	if len(epochs) < 15 && currentNodeEpoch >= 15 {
		// fill in projected epochs
		for e := currentNodeEpoch - (15 - uint64(len(epochs)) - 1); e <= currentNodeEpoch; e++ {
			tmp := types.IndexPageDataEpochs{
				Epoch:                   e,
				GlobalParticipationRate: 1,
			}
			epochs = append([]*types.IndexPageDataEpochs{&tmp}, epochs...)
		}
	}

	for _, epoch := range epochs {
		epoch.Ts = utils.EpochToTime(epoch.Epoch)
		epoch.Finalized = epoch.Epoch <= finalizedNodeEpoch
		epoch.FinalizedFormatted = utils.FormatYesNo(epoch.Finalized)
		epoch.VotedEtherFormatted = utils.FormatBalance(epoch.VotedEther, currency)
		epoch.EligibleEtherFormatted = utils.FormatBalanceShort(epoch.EligibleEther, currency)
		epoch.GlobalParticipationRateFormatted = utils.FormatGlobalParticipationRate(epoch.VotedEther, epoch.GlobalParticipationRate, currency, epoch.ParticipationStatsCompelete)
	}
	data.Epochs = epochs

	var scheduledCount uint8
	err = db.WriterDb.Get(&scheduledCount, `
		select count(*) from blocks where status = '0' and epoch = $1;
	`, data.CurrentEpoch)
	if err != nil {
		return nil, fmt.Errorf("error retrieving scheduledCount from blocks: %v", err)
	}
	data.ScheduledCount = scheduledCount

	var blocks []*types.IndexPageDataBlocks
	err = db.WriterDb.Select(&blocks, `
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
			COALESCE(validator_names.name, '') AS name
		FROM blocks 
		LEFT JOIN validators ON blocks.proposer = validators.validatorindex
		LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
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
	err = db.WriterDb.Get(&queueCount, "SELECT entering_validators_count, exiting_validators_count FROM queue ORDER BY ts DESC LIMIT 1")
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error retrieving validator queue count: %v", err)
	}
	data.EnteringValidators = queueCount.EnteringValidators
	data.ExitingValidators = queueCount.ExitingValidators

	var averageBalance float64
	err = db.WriterDb.Get(&averageBalance, "SELECT COALESCE(AVG(balance), 0) FROM validators")
	if err != nil {
		return nil, fmt.Errorf("error retrieving validator balance: %v", err)
	}
	data.AverageBalance = string(utils.FormatBalance(uint64(averageBalance), currency))

	var epochLowerBound uint64
	if epochLowerBound = 0; data.CurrentEpoch > 1600 {
		epochLowerBound = data.CurrentEpoch - 1600
	}
	var epochHistory []*types.IndexPageEpochHistory
	err = db.WriterDb.Select(&epochHistory, "SELECT epoch, eligibleether, validatorscount FROM epochs WHERE epoch < $1 and epoch > $2 and completeparticipationstats = true ORDER BY epoch", data.CurrentEpoch, epochLowerBound)
	if err != nil {
		return nil, fmt.Errorf("error retrieving staked ether history: %v", err)
	}

	if len(epochHistory) > 0 {
		data.CurrentFinalizedEpoch = LatestFinalizedEpoch()
		data.FinalityDelay = currentNodeEpoch - finalizedNodeEpoch

		data.StakedEther = string(utils.FormatBalance(epochHistory[len(epochHistory)-1].EligibleEther, currency))
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

// LatestNodeEpoch will return the latest epoch acording to the node
func LatestNodeEpoch() uint64 {
	return atomic.LoadUint64(&latestNodeEpoch)
}

// LatestFinalizedEpoch will return the most recent epoch that has been finalized.
func LatestFinalizedEpoch() uint64 {
	return atomic.LoadUint64(&latestFinalizedEpoch)
}

// LatestNodeFinalizedEpoch will return the most recent epoch that has been finalized acording to the node
func LatestNodeFinalizedEpoch() uint64 {
	return atomic.LoadUint64(&latestNodeFinalizedEpoch)
}

// LatestSlot will return the latest slot
func LatestSlot() uint64 {
	return atomic.LoadUint64(&latestSlot)
}

//FinalizationDelay will return the current Finalization Delay
func FinalizationDelay() uint64 {
	return LatestNodeEpoch() - LatestNodeFinalizedEpoch()
}

// LatestProposedSlot will return the latest proposed slot
func LatestProposedSlot() uint64 {
	return atomic.LoadUint64(&latestProposedSlot)
}

// LatestIndexPageData returns the latest index page data
func LatestIndexPageData() *types.IndexPageData {
	return indexPageData.Load().(*types.IndexPageData)
}

// LatestPoolsPageData returns the latest pools page data
func LatestPoolsPageData() *types.PoolsResp {
	return poolsData.Load().(*types.PoolsResp)
}

func LatestValidatorCount() uint64 {
	return atomic.LoadUint64(&latestValidatorCount)
}

// LatestState returns statistics about the current eth2 state
func LatestState() *types.LatestState {
	data := &types.LatestState{}
	data.CurrentEpoch = LatestEpoch()
	data.CurrentSlot = LatestSlot()
	data.CurrentFinalizedEpoch = LatestFinalizedEpoch()
	data.LastProposedSlot = atomic.LoadUint64(&latestProposedSlot)
	data.FinalityDelay = FinalizationDelay()
	data.IsSyncing = IsSyncing()
	data.UsdRoundPrice = price.GetEthRoundPrice(price.GetEthPrice("USD"))
	data.UsdTruncPrice = utils.KFormatterEthPrice(data.UsdRoundPrice)
	data.EurRoundPrice = price.GetEthRoundPrice(price.GetEthPrice("EUR"))
	data.EurTruncPrice = utils.KFormatterEthPrice(data.EurRoundPrice)
	data.GbpRoundPrice = price.GetEthRoundPrice(price.GetEthPrice("GBP"))
	data.GbpTruncPrice = utils.KFormatterEthPrice(data.GbpRoundPrice)
	data.CnyRoundPrice = price.GetEthRoundPrice(price.GetEthPrice("CNY"))
	data.CnyTruncPrice = utils.KFormatterEthPrice(data.CnyRoundPrice)
	data.RubRoundPrice = price.GetEthRoundPrice(price.GetEthPrice("RUB"))
	data.RubTruncPrice = utils.KFormatterEthPrice(data.RubRoundPrice)
	data.CadRoundPrice = price.GetEthRoundPrice(price.GetEthPrice("CAD"))
	data.CadTruncPrice = utils.KFormatterEthPrice(data.CadRoundPrice)
	data.AudRoundPrice = price.GetEthRoundPrice(price.GetEthPrice("AUD"))
	data.AudTruncPrice = utils.KFormatterEthPrice(data.AudRoundPrice)
	data.JpyRoundPrice = price.GetEthRoundPrice(price.GetEthPrice("JPY"))
	data.JpyTruncPrice = utils.KFormatterEthPrice(data.JpyRoundPrice)

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
			InvalidDepositCount:   new(uint64),
			UniqueValidatorCount:  new(uint64),
			TotalValidatorCount:   new(uint64),
			ActiveValidatorCount:  new(uint64),
			PendingValidatorCount: new(uint64),
			ValidatorChurnLimit:   new(uint64),
		}
	} else if stats.(*types.Stats).TopDepositors != nil && len(*stats.(*types.Stats).TopDepositors) == 1 {
		*stats.(*types.Stats).TopDepositors = append(*stats.(*types.Stats).TopDepositors, types.StatsTopDepositors{
			Address:      "000",
			DepositCount: 0,
		})
	}
	return stats.(*types.Stats)
}

// IsSyncing returns true if the chain is still syncing
func IsSyncing() bool {
	return time.Now().Add(time.Minute * -10).After(utils.EpochToTime(LatestEpoch()))
}
