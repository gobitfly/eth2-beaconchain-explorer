package services

import (
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
var indexPageData atomic.Value
var ready = sync.WaitGroup{}

var logger = logrus.New().WithField("module", "services")

// Init will initialize the services
func Init() {
	ready.Add(2)
	go epochUpdater()
	go indexPageDataUpdater()
	ready.Wait()
}

func epochUpdater() {
	firstRun := true

	for true {
		var epoch uint64
		err := db.DB.Get(&epoch, "SELECT COALESCE(MAX(epoch), 0) FROM epochs")

		if err != nil {
			logger.Printf("Error retrieving latest epoch from the database: %v", err)
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

	var epoch uint64
	err := db.DB.Get(&epoch, "SELECT COALESCE(MAX(epoch), 0) FROM epochs")

	if err != nil {
		return nil, fmt.Errorf("error retrieving latest epoch from the database: %v", err)
	}

	var blocks []*types.IndexPageDataBlocks

	err = db.DB.Select(&blocks, `SELECT blocks.epoch, 
											    blocks.slot, 
											    blocks.proposer, 
											    blocks.blockroot, 
											    blocks.parentroot, 
											    blocks.attestationscount, 
											    blocks.depositscount, 
											    blocks.voluntaryexitscount, 
											    blocks.proposerslashingscount, 
											    blocks.attesterslashingscount,
       											blocks.status
										FROM blocks 
										ORDER BY blocks.slot DESC LIMIT 20`)

	if err != nil {
		return nil, fmt.Errorf("error retrieving index block data: %v", err)
	}
	data.Blocks = blocks

	for _, block := range data.Blocks {
		block.StatusFormatted = utils.FormatBlockStatus(block.Status)
		block.ProposerFormatted = utils.FormatValidator(block.Proposer)
		block.BlockRootFormatted = fmt.Sprintf("%x", block.BlockRoot)
	}

	if len(blocks) > 0 {
		data.CurrentSlot = blocks[0].Slot
	}

	for _, block := range data.Blocks {
		block.Ts = utils.SlotToTime(block.Slot)
	}

	err = db.DB.Get(&data.EnteringValidators, "SELECT COUNT(*) FROM validatorqueue_activation")
	if err != nil {
		return nil, fmt.Errorf("error retrieving entering validator count: %v", err)
	}

	err = db.DB.Get(&data.ExitingValidators, "SELECT COUNT(*) FROM validatorqueue_exit")
	if err != nil {
		return nil, fmt.Errorf("error retrieving exiting validator count: %v", err)
	}

	var averageBalance float64
	err = db.DB.Get(&averageBalance, "SELECT COALESCE(AVG(balance), 0) FROM validator_balances WHERE epoch = $1", epoch)
	if err != nil {
		return nil, fmt.Errorf("error retrieving validator balance: %v", err)
	}
	data.AverageBalance = utils.FormatBalance(uint64(averageBalance))

	var epochHistory []*types.IndexPageEpochHistory
	err = db.DB.Select(&epochHistory, "SELECT epoch, eligibleether, validatorscount, finalized FROM epochs WHERE epoch < $1 ORDER BY epoch", epoch)
	if err != nil {
		return nil, fmt.Errorf("error retrieving staked ether history: %v", err)
	}

	if len(epochHistory) > 0 {
		data.CurrentEpoch = epochHistory[len(epochHistory)-1].Epoch

		for i := len(epochHistory) - 1; i >= 0; i-- {
			if epochHistory[i].Finalized {
				data.CurrentFinalizedEpoch = epochHistory[i].Epoch
				data.FinalityDelay = data.CurrentEpoch - data.CurrentFinalizedEpoch
				break
			}
		}

		data.StakedEther = utils.FormatBalance(epochHistory[len(epochHistory)-1].EligibleEther)
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

// LatestIndexPageData returns the latest index page data
func LatestIndexPageData() *types.IndexPageData {
	return indexPageData.Load().(*types.IndexPageData)
}

// IsSyncing returns true if the chain is still syncing
func IsSyncing() bool {
	return time.Now().Add(time.Minute * -10).After(utils.EpochToTime(LatestEpoch()))
}
