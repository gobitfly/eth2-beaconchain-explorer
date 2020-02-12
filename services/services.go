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
	go performanceDataUpdater()

	ready.Add(2)
	go epochUpdater()
	go indexPageDataUpdater()
	ready.Wait()
}

func performanceDataUpdater() {
	for true {
		logger.Info("updating validator performance data")
		err := UpdateValidatorPerformance()

		if err != nil {
			logger.Printf("error updating validator performance data: %v", err)
		} else {
			logger.Info("validator performance data update completed")
		}
		time.Sleep(time.Hour)
	}
}

func epochUpdater() {
	firstRun := true

	for true {
		var epoch uint64
		err := db.DB.Get(&epoch, "SELECT COALESCE(MAX(epoch), 0) FROM epochs")

		if err != nil {
			logger.Printf("error retrieving latest epoch from the database: %v", err)
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
										WHERE blocks.slot < $1
										ORDER BY blocks.slot DESC LIMIT 20`, utils.TimeToSlot(uint64(time.Now().Add(time.Second*10).Unix())))

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

func UpdateValidatorPerformance() error {
	tx, err := db.DB.Beginx()
	if err != nil {
		return fmt.Errorf("error starting db transaction: %w")
	}
	defer tx.Rollback()

	_, err = tx.Exec("TRUNCATE validator_performance")
	if err != nil {
		return fmt.Errorf("error truncating validator performance table: %w")
	}

	var currentEpoch uint64

	err = tx.Get(&currentEpoch, "SELECT MAX(epoch) FROM validator_balances")
	if err != nil {
		return fmt.Errorf("error retrieving latest epoch from validator_balances table: %w")
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
			       validators.validatorindex, 
			       amount AS balance 
			FROM blocks_deposits 
			    LEFT JOIN validators ON validators.pubkey = blocks_deposits.publickey
			WHERE validators.validatorindex IS NOT NULL;
			`)
	if err != nil {
		return fmt.Errorf("error retrieving initial validator balances data: %w", err)
	}

	startBalanceMap := make(map[uint64]uint64)
	for _, balance := range startBalances {
		startBalanceMap[balance.Index] += balance.Balance
	}

	var balances []*types.ValidatorBalance

	err = tx.Select(&balances, `SELECT 
											   validator_balances.epoch, 
											   validator_balances.validatorindex, 
											   validator_balances.balance
										FROM validator_balances 
										WHERE validator_balances.epoch IN ($1, $2, $3, $4, $5)`, currentEpoch, epoch1d, epoch7d, epoch31d, epoch365d)
	if err != nil {
		return fmt.Errorf("error retrieving validator performance data: %w", err)
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

		_, err := tx.Exec("INSERT INTO validator_performance (validatorindex, balance, performance1d, performance7d, performance31d, performance365d) VALUES ($1, $2, $3, $4, $5, $6)",
			validator, currentBalance, performance1d, performance7d, performance31d, performance365d)

		if err != nil {
			return fmt.Errorf("error saving validator performance data: %w", err)
		}
	}

	return tx.Commit()
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
