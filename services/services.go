package services

import (
	"database/sql"
	"encoding/json"
	"eth2-exporter/cache"
	"eth2-exporter/db"
	"eth2-exporter/price"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"

	geth_types "github.com/ethereum/go-ethereum/core/types"
	geth_rpc "github.com/ethereum/go-ethereum/rpc"
)

var eth1BlockDepositReached atomic.Value
var depositThresholdReached atomic.Value

var logger = logrus.New().WithField("module", "services")

// Init will initialize the services
func Init() {
	ready := &sync.WaitGroup{}
	ready.Add(1)
	go epochUpdater(ready)

	ready.Add(1)
	go slotUpdater(ready)

	ready.Add(1)
	go latestProposedSlotUpdater(ready)

	ready.Add(1)
	go latestBlockUpdater(ready)

	// ready.Add(1)
	// go gasNowUpdater()

	ready.Add(1)
	go slotVizUpdater(ready)

	ready.Add(1)
	go indexPageDataUpdater(ready)

	ready.Add(1)
	go poolsUpdater(ready)

	ready.Add(1)
	go chartsPageDataUpdater(ready)

	ready.Add(1)
	go statsUpdater(ready)

	ready.Wait()
}

func InitNotifications() {
	logger.Infof("starting notifications-sender")
	go notificationsSender()
}

func epochUpdater(wg *sync.WaitGroup) {
	firstRun := true
	for {
		var latestFinalized uint64
		err := db.WriterDb.Get(&latestFinalized, "SELECT COALESCE(MAX(epoch), 0) FROM epochs where finalized is true")
		if err != nil {
			logger.Errorf("error retrieving latest finalized epoch from the database: %v", err)
		} else {
			cacheKey := fmt.Sprintf("%d:frontend:latestFinalized", utils.Config.Chain.Config.DepositChainID)
			err := cache.TieredCache.SetUint64(cacheKey, latestFinalized, time.Hour*24)
			if err != nil {
				logger.Errorf("error caching latestFinalizedEpoch: %v", err)
			}
		}

		var epoch uint64
		err = db.WriterDb.Get(&epoch, "SELECT COALESCE(MAX(epoch), 0) FROM blocks")
		if err != nil {
			logger.Errorf("error retrieving latest epoch from the database: %v", err)
		} else {
			cacheKey := fmt.Sprintf("%d:frontend:latestEpoch", utils.Config.Chain.Config.DepositChainID)
			err := cache.TieredCache.SetUint64(cacheKey, epoch, time.Hour*24)
			if err != nil {
				logger.Errorf("error caching latestEpoch: %v", err)
			}
			if firstRun {
				logger.Info("initialized epoch updater")
				wg.Done()
				firstRun = false
			}
		}
		time.Sleep(time.Second)
	}
}

func slotUpdater(wg *sync.WaitGroup) {
	firstRun := true

	for {
		var slot uint64
		err := db.WriterDb.Get(&slot, "SELECT COALESCE(MAX(slot), 0) FROM blocks where slot < $1", utils.TimeToSlot(uint64(time.Now().Add(time.Second*10).Unix())))

		if err != nil {
			logger.Errorf("error retrieving latest slot from the database: %v", err)

			if err.Error() == "sql: database is closed" {
				logger.Fatalf("error retrieving latest slot from the database: %v", err)
			}
		} else {
			cacheKey := fmt.Sprintf("%d:frontend:slot", utils.Config.Chain.Config.DepositChainID)
			err := cache.TieredCache.SetUint64(cacheKey, slot, time.Hour*24)
			if err != nil {
				logger.Errorf("error caching slot: %v", err)
			}
			if firstRun {
				logger.Info("initialized slot updater")
				wg.Done()
				firstRun = false
			}
		}
		time.Sleep(time.Second)
	}
}

func poolsUpdater(wg *sync.WaitGroup) {
	firstRun := true

	for {
		data, err := getPoolsPageData()
		if err != nil {
			logger.Errorf("error retrieving pools page data: %v", err)
			time.Sleep(time.Second * 10)
			continue
		}

		cacheKey := fmt.Sprintf("%d:frontend:poolsData", utils.Config.Chain.Config.DepositChainID)
		err = cache.TieredCache.Set(cacheKey, data, time.Hour*24)
		if err != nil {
			logger.Errorf("error caching poolsData: %v", err)
		}
		if firstRun {
			logger.Info("initialized pools page updater")
			wg.Done()
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

func latestProposedSlotUpdater(wg *sync.WaitGroup) {
	firstRun := true

	for {
		var slot uint64
		err := db.WriterDb.Get(&slot, "SELECT COALESCE(MAX(slot), 0) FROM blocks WHERE status = '1'")

		if err != nil {
			logger.Errorf("error retrieving latest proposed slot from the database: %v", err)
		} else {

			cacheKey := fmt.Sprintf("%d:frontend:latestProposedSlot", utils.Config.Chain.Config.DepositChainID)
			err = cache.TieredCache.SetUint64(cacheKey, slot, time.Hour*24)
			if err != nil {
				logger.Errorf("error caching latestProposedSlot: %v", err)
			}
			if firstRun {
				logger.Info("initialized last proposed slot updater")
				wg.Done()
				firstRun = false
			}
		}
		time.Sleep(time.Second)
	}
}

func indexPageDataUpdater(wg *sync.WaitGroup) {
	firstRun := true

	for {
		logger.Infof("updating index page data")
		start := time.Now()
		data, err := getIndexPageData()
		if err != nil {
			logger.Errorf("error retrieving index page data: %v", err)
			time.Sleep(time.Second * 10)
			continue
		}
		logger.Infof("index page data update completed in %v", time.Since(start))

		cacheKey := fmt.Sprintf("%d:frontend:indexPageData", utils.Config.Chain.Config.DepositChainID)
		err = cache.TieredCache.Set(cacheKey, data, time.Hour*24)
		if err != nil {
			logger.Errorf("error caching indexPageData: %v", err)
		}
		if firstRun {
			logger.Info("initialized index page updater")
			wg.Done()
			firstRun = false
		}
		time.Sleep(time.Second * 10)
	}
}

func slotVizUpdater(wg *sync.WaitGroup) {
	firstRun := true

	for {
		latestEpoch := LatestEpoch()
		if latestEpoch > 0 {
			epochData, err := db.GetSlotVizData(latestEpoch)
			if err != nil {
				logger.Errorf("error retrieving slot viz data from database: %v latest epoch: %v", err, latestEpoch)
			} else {
				cacheKey := fmt.Sprintf("%d:frontend:slotVizMetrics", utils.Config.Chain.Config.DepositChainID)
				err = cache.TieredCache.Set(cacheKey, epochData, time.Hour*24)
				if err != nil {
					logger.Errorf("error caching slotVizMetrics: %v", err)
				}
				if firstRun {
					logger.Info("initialized slotViz metrics")
					wg.Done()
					firstRun = false
				}
			}
		}
		time.Sleep(time.Second)
	}
}

func getIndexPageData() (*types.IndexPageData, error) {
	currency := "ETH"

	data := &types.IndexPageData{}
	data.Mainnet = utils.Config.Chain.Config.ConfigName == "mainnet"
	data.NetworkName = utils.Config.Chain.Config.ConfigName
	data.DepositContract = utils.Config.Indexer.Eth1DepositContractAddress

	var epoch uint64
	err := db.ReaderDb.Get(&epoch, "SELECT COALESCE(MAX(epoch), 0) FROM epochs")
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
	if now.Before(genesisTransition) || startSlotTime == time.Unix(0, 0) {
		if cutoffSlot < 15 {
			cutoffSlot = 15
		}
		type Deposit struct {
			Total   uint64    `db:"total"`
			BlockTs time.Time `db:"block_ts"`
		}

		deposit := Deposit{}
		err = db.ReaderDb.Get(&deposit, `
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

	var scheduledCount uint8
	err = db.WriterDb.Get(&scheduledCount, `
		select count(*) from blocks where status = '0' and epoch = $1;
	`, epoch)
	if err != nil {
		return nil, fmt.Errorf("error retrieving scheduledCount from blocks: %v", err)
	}
	data.ScheduledCount = scheduledCount

	var epochs []*types.IndexPageDataEpochs
	err = db.ReaderDb.Select(&epochs, `SELECT epoch, finalized , eligibleether, globalparticipationrate, votedether FROM epochs ORDER BY epochs DESC LIMIT 15`)
	if err != nil {
		return nil, fmt.Errorf("error retrieving index epoch data: %v", err)
	}
	epochsMap := make(map[uint64]bool)
	for _, epoch := range epochs {
		epoch.Ts = utils.EpochToTime(epoch.Epoch)
		epoch.FinalizedFormatted = utils.FormatYesNo(epoch.Finalized)
		epoch.VotedEtherFormatted = utils.FormatBalance(epoch.VotedEther, currency)
		epoch.EligibleEtherFormatted = utils.FormatEligibleBalance(epoch.EligibleEther, currency)
		epoch.GlobalParticipationRateFormatted = utils.FormatGlobalParticipationRate(epoch.VotedEther, epoch.GlobalParticipationRate, currency)
		epochsMap[epoch.Epoch] = true
	}

	var blocks []*types.IndexPageDataBlocks
	err = db.ReaderDb.Select(&blocks, `
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
			COALESCE(blocks.exec_block_number, 0) AS exec_block_number,
			COALESCE(validator_names.name, '') AS name
		FROM blocks 
		LEFT JOIN validators ON blocks.proposer = validators.validatorindex
		LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
		WHERE blocks.slot < $1
		ORDER BY blocks.slot DESC LIMIT 20`, cutoffSlot)
	if err != nil {
		return nil, fmt.Errorf("error retrieving index block data: %v", err)
	}

	blocksMap := make(map[uint64]*types.IndexPageDataBlocks)
	for _, block := range blocks {
		if blocksMap[block.Slot] == nil || len(block.BlockRoot) > len(blocksMap[block.Slot].BlockRoot) {
			blocksMap[block.Slot] = block
		}
	}
	blocks = make([]*types.IndexPageDataBlocks, 0, len(blocks))
	for _, b := range blocksMap {
		blocks = append(blocks, b)
	}
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].Slot > blocks[j].Slot
	})
	data.Blocks = blocks

	if len(data.Blocks) > 15 {
		data.Blocks = data.Blocks[:15]
	}

	for _, block := range data.Blocks {
		block.StatusFormatted = utils.FormatBlockStatus(block.Status)
		block.ProposerFormatted = utils.FormatValidatorWithName(block.Proposer, block.ProposerName)
		block.BlockRootFormatted = fmt.Sprintf("%x", block.BlockRoot)

		if !epochsMap[block.Epoch] {
			epochs = append(epochs, &types.IndexPageDataEpochs{
				Epoch:                            block.Epoch,
				Ts:                               utils.EpochToTime(block.Epoch),
				Finalized:                        false,
				FinalizedFormatted:               utils.FormatYesNo(false),
				EligibleEther:                    0,
				EligibleEtherFormatted:           utils.FormatEligibleBalance(0, "ETH"),
				GlobalParticipationRate:          0,
				GlobalParticipationRateFormatted: utils.FormatGlobalParticipationRate(0, 1, ""),
				VotedEther:                       0,
				VotedEtherFormatted:              "",
			})
			epochsMap[block.Epoch] = true
		}
	}
	sort.Slice(epochs, func(i, j int) bool {
		return epochs[i].Epoch > epochs[j].Epoch
	})

	data.Epochs = epochs

	if len(data.Epochs) > 15 {
		data.Epochs = data.Epochs[:15]
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
	err = db.ReaderDb.Get(&queueCount, "SELECT entering_validators_count, exiting_validators_count FROM queue ORDER BY ts DESC LIMIT 1")
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error retrieving validator queue count: %v", err)
	}
	data.EnteringValidators = queueCount.EnteringValidators
	data.ExitingValidators = queueCount.ExitingValidators

	var epochLowerBound uint64
	if epochLowerBound = 0; epoch > 1600 {
		epochLowerBound = epoch - 1600
	}
	var epochHistory []*types.IndexPageEpochHistory
	err = db.WriterDb.Select(&epochHistory, "SELECT epoch, eligibleether, validatorscount, finalized, averagevalidatorbalance FROM epochs WHERE epoch < $1 and epoch > $2 ORDER BY epoch", epoch, epochLowerBound)
	if err != nil {
		return nil, fmt.Errorf("error retrieving staked ether history: %v", err)
	}

	if len(epochHistory) > 0 {
		for i := len(epochHistory) - 1; i >= 0; i-- {
			if epochHistory[i].Finalized {
				data.CurrentFinalizedEpoch = epochHistory[i].Epoch
				data.FinalityDelay = data.CurrentEpoch - epoch
				data.AverageBalance = string(utils.FormatBalance(uint64(epochHistory[i].AverageValidatorBalance), currency))
				break
			}
		}

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
	cacheKey := fmt.Sprintf("%d:frontend:latestEpoch", utils.Config.Chain.Config.DepositChainID)

	if wanted, err := cache.TieredCache.GetUint64WithLocalTimeout(cacheKey, time.Second*5); err == nil {
		return wanted
	} else {
		logger.Errorf("error retrieving latestEpoch from cache: %v", err)
	}

	return 0
}

// LatestFinalizedEpoch will return the most recent epoch that has been finalized.
func LatestFinalizedEpoch() uint64 {
	cacheKey := fmt.Sprintf("%d:frontend:latestFinalized", utils.Config.Chain.Config.DepositChainID)

	if wanted, err := cache.TieredCache.GetUint64WithLocalTimeout(cacheKey, time.Second*5); err == nil {
		return wanted
	} else {
		logger.Errorf("error retrieving latestFinalized from cache: %v", err)
	}
	return 0
}

// LatestSlot will return the latest slot
func LatestSlot() uint64 {
	cacheKey := fmt.Sprintf("%d:frontend:slot", utils.Config.Chain.Config.DepositChainID)

	if wanted, err := cache.TieredCache.GetUint64WithLocalTimeout(cacheKey, time.Second*5); err == nil {
		return wanted
	} else {
		logger.Errorf("error retrieving slot from cache: %v", err)
	}
	return 0
}

//FinalizationDelay will return the current Finalization Delay
func FinalizationDelay() uint64 {
	return LatestEpoch() - LatestFinalizedEpoch()
}

// LatestProposedSlot will return the latest proposed slot
func LatestProposedSlot() uint64 {
	cacheKey := fmt.Sprintf("%d:frontend:latestProposedSlot", utils.Config.Chain.Config.DepositChainID)

	if wanted, err := cache.TieredCache.GetUint64WithLocalTimeout(cacheKey, time.Second*5); err == nil {
		return wanted
	} else {
		logger.Errorf("error retrieving latestProposedSlot from cache: %v", err)
	}
	return 0
}

// LatestIndexPageData returns the latest index page data
func LatestIndexPageData() *types.IndexPageData {
	var wanted *types.IndexPageData
	cacheKey := fmt.Sprintf("%d:frontend:indexPageData", utils.Config.Chain.Config.DepositChainID)

	if err := cache.TieredCache.GetWithLocalTimeout(cacheKey, time.Second*5, &wanted); err == nil {
		return wanted
	} else {
		logger.Errorf("error retrieving indexPageData from cache: %v", err)
	}
	return &types.IndexPageData{
		NetworkName:               "",
		DepositContract:           "",
		ShowSyncingMessage:        false,
		CurrentEpoch:              0,
		CurrentFinalizedEpoch:     0,
		CurrentSlot:               0,
		ScheduledCount:            0,
		FinalityDelay:             0,
		ActiveValidators:          0,
		EnteringValidators:        0,
		ExitingValidators:         0,
		StakedEther:               "",
		AverageBalance:            "",
		DepositedTotal:            0,
		DepositThreshold:          0,
		ValidatorsRemaining:       0,
		NetworkStartTs:            0,
		MinGenesisTime:            0,
		Blocks:                    []*types.IndexPageDataBlocks{},
		Epochs:                    []*types.IndexPageDataEpochs{},
		StakedEtherChartData:      [][]float64{},
		ActiveValidatorsChartData: [][]float64{},
		Subtitle:                  "",
		Genesis:                   false,
		GenesisPeriod:             false,
		Mainnet:                   false,
		DepositChart:              &types.ChartsPageDataChart{},
		DepositDistribution:       &types.ChartsPageDataChart{},
		Countdown:                 nil,
		SlotVizData:               types.SlotVizPageData{},
	}
}

// LatestPoolsPageData returns the latest pools page data
func LatestPoolsPageData() *types.PoolsResp {

	var wanted *types.PoolsResp
	cacheKey := fmt.Sprintf("%d:frontend:poolsData", utils.Config.Chain.Config.DepositChainID)

	if err := cache.TieredCache.GetWithLocalTimeout(cacheKey, time.Second*5, &wanted); err == nil {
		return wanted
	} else {
		logger.Errorf("error retrieving poolsData from cache: %v", err)
	}

	return &types.PoolsResp{
		PoolsDistribution:       types.ChartsPageDataChart{},
		HistoricPoolPerformance: types.ChartsPageDataChart{},
		PoolInfos:               []*types.PoolInfo{},
	}
}

func LatestGasNowData() *types.GasNowPageData {
	var wanted *types.GasNowPageData
	cacheKey := fmt.Sprintf("%d:frontend:gasNow", utils.Config.Chain.Config.DepositChainID)

	if err := cache.TieredCache.GetWithLocalTimeout(cacheKey, time.Second*5, &wanted); err == nil {
		return wanted
	} else {
		logger.Errorf("error retrieving gasNow from cache: %v", err)
	}

	return nil
}

func LatestSlotVizMetrics() []*types.SlotVizEpochs {

	var wanted []*types.SlotVizEpochs
	cacheKey := fmt.Sprintf("%d:frontend:slotVizMetrics", utils.Config.Chain.Config.DepositChainID)

	if err := cache.TieredCache.GetWithLocalTimeout(cacheKey, time.Second*5, &wanted); err == nil {
		return wanted
	} else {
		logger.Errorf("error retrieving slotVizMetrics from cache: %v", err)
	}

	return []*types.SlotVizEpochs{}
}

// LatestState returns statistics about the current eth2 state
func LatestState() *types.LatestState {
	data := &types.LatestState{}
	data.CurrentEpoch = LatestEpoch()
	data.CurrentSlot = LatestSlot()
	data.CurrentFinalizedEpoch = LatestFinalizedEpoch()
	data.LastProposedSlot = LatestProposedSlot()
	data.FinalityDelay = data.CurrentEpoch - data.CurrentFinalizedEpoch
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
	var wanted *types.Stats
	cacheKey := fmt.Sprintf("%d:frontend:latestStats", utils.Config.Chain.Config.DepositChainID)

	if err := cache.TieredCache.GetWithLocalTimeout(cacheKey, time.Second*5, &wanted); err == nil {
		return wanted
	} else {
		logger.Errorf("error retrieving slotVizMetrics from cache: %v", err)
	}

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
}

// IsSyncing returns true if the chain is still syncing
func IsSyncing() bool {
	return time.Now().Add(time.Minute * -10).After(utils.EpochToTime(LatestEpoch()))
}

func gasNowUpdater(wg *sync.WaitGroup) {
	firstRun := true

	for {
		data, err := getGasNowData()
		if err != nil {
			logger.Errorf("error retrieving gas now data: %v", err)
			time.Sleep(time.Second * 5)
			continue
		}

		cacheKey := fmt.Sprintf("%d:frontend:gasNow", utils.Config.Chain.Config.DepositChainID)
		err = cache.TieredCache.Set(cacheKey, data, time.Hour*24)
		if err != nil {
			logger.Errorf("error caching latestFinalizedEpoch: %v", err)
		}
		if firstRun {
			wg.Done()
			firstRun = false
		}
		time.Sleep(time.Second * 5)
	}
}

func getGasNowData() (*types.GasNowPageData, error) {
	gpoData := &types.GasNowPageData{}
	gpoData.Code = 200
	gpoData.Data.Timestamp = time.Now().UnixNano() / 1e6

	client, err := geth_rpc.Dial(utils.Config.Eth1GethEndpoint)
	if err != nil {
		return nil, err
	}
	var raw json.RawMessage
	var body rpcBlock

	err = client.Call(&raw, "eth_getBlockByNumber", "pending", true)

	if err != nil {
		return nil, fmt.Errorf("error retrieving pending block data: %v", err)
	}

	err = json.Unmarshal(raw, &body)
	if err != nil {
		return nil, err
	}

	logger.Infof("pending block has %v tx", len(body.Transactions))

	sort.Slice(body.Transactions, func(i, j int) bool {
		return body.Transactions[i].tx.GasPrice().Cmp(body.Transactions[j].tx.GasPrice()) > 0
	})
	if len(body.Transactions) > 1 {
		medianGasPrice := body.Transactions[len(body.Transactions)/2].tx.GasPrice()
		tailGasPrice := body.Transactions[len(body.Transactions)-1].tx.GasPrice()

		gpoData.Data.Rapid = medianGasPrice
		gpoData.Data.Fast = tailGasPrice
	} else {
		return nil, fmt.Errorf("current pending block contains no tx")
	}

	err = client.Call(&raw, "txpool_content")
	if err != nil {
		logrus.Fatal(err)
	}

	txPoolContent := &TxPoolContent{}
	err = json.Unmarshal(raw, txPoolContent)
	if err != nil {
		logrus.Fatal(err)
	}

	pendingTxs := make([]*geth_types.Transaction, 0, len(txPoolContent.Pending))

	for _, account := range txPoolContent.Pending {
		lowestNonce := 9223372036854775807
		for n := range account {
			if n < int(lowestNonce) {
				lowestNonce = n
			}
		}

		pendingTxs = append(pendingTxs, account[lowestNonce])
	}
	sort.Slice(pendingTxs, func(i, j int) bool {
		return pendingTxs[i].GasPrice().Cmp(pendingTxs[j].GasPrice()) > 0
	})

	standardIndex := int(math.Max(float64(2*len(body.Transactions)), 500))
	slowIndex := int(math.Max(float64(5*len(body.Transactions)), 1000))
	if standardIndex > len(pendingTxs)-1 {
		standardIndex = len(pendingTxs) - 1
	}
	if slowIndex > len(pendingTxs)-1 {
		slowIndex = len(pendingTxs) - 1
	}

	gpoData.Data.Standard = pendingTxs[standardIndex].GasPrice()
	gpoData.Data.Slow = pendingTxs[slowIndex].GasPrice()

	// gpoData.RapidUSD = gpoData.Rapid * 21000 * params.GWei / params.Ether * usd
	// gpoData.FastUSD = gpoData.Fast * 21000 * params.GWei / params.Ether * usd
	// gpoData.StandardUSD = gpoData.Standard * 21000 * params.GWei / params.Ether * usd
	// gpoData.SlowUSD = gpoData.Slow * 21000 * params.GWei / params.Ether * usd
	return gpoData, nil
}

type TxPoolContent struct {
	Pending map[string]map[int]*geth_types.Transaction
}

type rpcTransaction struct {
	tx *geth_types.Transaction
	txExtraInfo
}

type txExtraInfo struct {
	BlockNumber *string         `json:"blockNumber,omitempty"`
	BlockHash   *common.Hash    `json:"blockHash,omitempty"`
	From        *common.Address `json:"from,omitempty"`
}

type rpcBlock struct {
	Transactions []rpcTransaction `json:"transactions"`
}

func (tx *rpcTransaction) UnmarshalJSON(msg []byte) error {
	if err := json.Unmarshal(msg, &tx.tx); err != nil {
		return err
	}
	return json.Unmarshal(msg, &tx.txExtraInfo)
}
