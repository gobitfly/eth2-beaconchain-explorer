package handlers

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"eth2-exporter/db"
	"eth2-exporter/price"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"math/big"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	utilMath "github.com/protolambda/zrnt/eth2/util/math"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var pkeyRegex = regexp.MustCompile("[^0-9A-Fa-f]+")

func GetValidatorOnlineThresholdSlot() uint64 {
	latestProposedSlot := services.LatestProposedSlot()
	threshold := utils.Config.Chain.Config.SlotsPerEpoch * 2

	var validatorOnlineThresholdSlot uint64
	if latestProposedSlot < 1 || latestProposedSlot < threshold {
		validatorOnlineThresholdSlot = 0
	} else {
		validatorOnlineThresholdSlot = latestProposedSlot - threshold
	}

	return validatorOnlineThresholdSlot
}

// GetValidatorEarnings will return the earnings (last day, week, month and total) of selected validators, including proposal and statisic information - infused with data from the current b. day
func GetValidatorEarnings(validators []uint64, currency string) (*types.ValidatorEarnings, map[uint64]*types.Validator, error) {
	if len(validators) == 0 {
		return nil, nil, errors.New("no validators provided")
	}
	latestFinalizedEpoch := services.LatestFinalizedEpoch()
	lastStatsDay, err := db.GetLastExportedStatisticDay()
	if err != nil {
		return nil, nil, err
	}
	firstEpoch := (lastStatsDay + 1) * utils.EpochsPerDay()

	balancesMap := make(map[uint64]*types.Validator, 0)
	totalBalance := uint64(0)

	g := errgroup.Group{}
	g.Go(func() error {
		latestBalances, err := db.BigtableClient.GetValidatorBalanceHistory(validators, latestFinalizedEpoch, latestFinalizedEpoch)
		if err != nil {
			logger.Errorf("error getting validator balance data in GetValidatorEarnings: %v", err)
			return err
		}

		for balanceIndex, balance := range latestBalances {
			if len(balance) == 0 {
				continue
			}

			if balancesMap[balanceIndex] == nil {
				balancesMap[balanceIndex] = &types.Validator{}
			}
			balancesMap[balanceIndex].Balance = balance[0].Balance
			balancesMap[balanceIndex].EffectiveBalance = balance[0].EffectiveBalance

			totalBalance += balance[0].Balance
		}
		return nil
	})

	income := types.ValidatorIncomePerformance{}
	g.Go(func() error {
		return db.GetValidatorIncomePerforamance(validators, &income)
	})

	var totalDeposits uint64
	g.Go(func() error {
		return db.GetTotalValidatorDeposits(validators, &totalDeposits)
	})

	var totalWithdrawals uint64
	g.Go(func() error {
		return db.GetTotalValidatorWithdrawals(validators, &totalWithdrawals)
	})

	var lastDeposits uint64
	g.Go(func() error {
		return db.GetValidatorDepositsForEpochs(validators, firstEpoch, latestFinalizedEpoch, &lastDeposits)
	})

	var lastWithdrawals uint64
	g.Go(func() error {
		return db.GetValidatorWithdrawalsForEpochs(validators, firstEpoch, latestFinalizedEpoch, &lastWithdrawals)
	})

	var lastBalance uint64
	g.Go(func() error {
		return db.GetValidatorBalanceForDay(validators, lastStatsDay, &lastBalance)
	})

	proposals := []types.ValidatorProposalInfo{}
	g.Go(func() error {
		return db.GetValidatorPropsosals(validators, &proposals)
	})

	err = g.Wait()
	if err != nil {
		return nil, nil, err
	}
	currentDayClIncome := int64(totalBalance - lastBalance - lastDeposits + lastWithdrawals)

	elClPrice := price.GetPrice(utils.Config.Frontend.ElCurrencySymbol, utils.Config.Frontend.ClCurrencySymbol)

	// calculate combined el and cl earnings
	earnings1d := float64(income.ClIncome1d) + elClPrice*float64(income.ElIncome1d)
	earnings7d := float64(income.ClIncome7d) + elClPrice*float64(income.ElIncome7d)
	earnings31d := float64(income.ClIncome31d) + elClPrice*float64(income.ElIncome31d)

	if totalDeposits == 0 {
		totalDeposits = utils.Config.Chain.Config.MaxEffectiveBalance * uint64(len(validators))
	}

	clApr7d := ((float64(income.ClIncome7d) / float64(totalDeposits)) * 365) / 7
	if clApr7d < float64(-1) {
		clApr7d = float64(-1)
	}

	elApr7d := ((float64(income.ElIncome7d) / float64(totalDeposits)) * 365) / 7
	if elApr7d < float64(-1) {
		elApr7d = float64(-1)
	}

	clApr31d := ((float64(income.ClIncome31d) / float64(totalDeposits)) * 365) / 31
	if clApr31d < float64(-1) {
		clApr31d = float64(-1)
	}

	elApr31d := ((float64(income.ElIncome31d) / float64(totalDeposits)) * 365) / 31
	if elApr31d < float64(-1) {
		elApr31d = float64(-1)
	}
	clApr365d := (float64(income.ClIncome365d) / float64(totalDeposits))
	if clApr365d < float64(-1) {
		clApr365d = float64(-1)
	}

	elApr365d := (float64(income.ElIncome365d) / float64(totalDeposits))
	if elApr365d < float64(-1) {
		elApr365d = float64(-1)
	}

	// retrieve cl income not yet in stats
	currentDayProposerIncome, err := db.GetCurrentDayProposerIncomeTotal(validators)
	if err != nil {
		return nil, nil, err
	}

	incomeToday := types.ClElInt64{
		El:    0,
		Cl:    currentDayClIncome,
		Total: float64(currentDayClIncome),
	}

	proposedToday := []uint64{}
	todayStartEpoch := uint64(lastStatsDay+1) * utils.EpochsPerDay()
	validatorProposalData := types.ValidatorProposalData{}
	validatorProposalData.Proposals = make([][]uint64, len(proposals))
	for i, b := range proposals {
		validatorProposalData.Proposals[i] = []uint64{
			uint64(utils.SlotToTime(b.Slot).Unix()),
			b.Status,
		}
		if b.Status == 0 {
			validatorProposalData.LastScheduledSlot = utilMath.MaxU64(validatorProposalData.LastScheduledSlot, b.Slot)
			validatorProposalData.ScheduledBlocksCount++
		} else if b.Status == 1 {
			validatorProposalData.ProposedBlocksCount++
			// add to list of blocks proposed today if epoch hasn't been exported into stats yet
			if utils.EpochOfSlot(b.Slot) >= todayStartEpoch && b.ExecBlockNumber.Int64 > 0 {
				proposedToday = append(proposedToday, uint64(b.ExecBlockNumber.Int64))
			}
		} else if b.Status == 2 {
			validatorProposalData.MissedBlocksCount++
		} else if b.Status == 3 {
			validatorProposalData.OrphanedBlocksCount++
		}
	}

	validatorProposalData.BlocksCount = uint64(len(proposals))
	if validatorProposalData.BlocksCount > 0 {
		validatorProposalData.UnmissedBlocksPercentage = float64(validatorProposalData.BlocksCount-validatorProposalData.MissedBlocksCount-validatorProposalData.OrphanedBlocksCount) / float64(validatorProposalData.BlocksCount)
	} else {
		validatorProposalData.UnmissedBlocksPercentage = 1.0
	}

	var slots []uint64
	for _, p := range proposals {
		if p.ExecBlockNumber.Int64 > 0 {
			slots = append(slots, p.Slot)
		}
	}

	lookbackAmount := getProposalLuckBlockLookbackAmount(1)
	startPeriod := len(slots) - lookbackAmount
	if startPeriod < 0 {
		startPeriod = 0
	}

	validatorProposalData.ProposalLuck = getProposalLuck(slots[startPeriod:], 1)
	avgSlotInterval := uint64(getAvgSlotInterval(1))
	avgSlotIntervalAsDuration := time.Duration(utils.Config.Chain.Config.SecondsPerSlot*avgSlotInterval) * time.Second
	validatorProposalData.AvgSlotInterval = &avgSlotIntervalAsDuration
	if len(slots) > 0 {
		nextSlotEstimate := utils.SlotToTime(slots[len(slots)-1] + avgSlotInterval)
		validatorProposalData.ProposalEstimate = &nextSlotEstimate
	}

	if len(proposedToday) > 0 {
		// get el data
		execBlocks, err := db.BigtableClient.GetBlocksIndexedMultiple(proposedToday, 10000)
		if err != nil {
			return nil, nil, fmt.Errorf("error retrieving execution blocks data from bigtable: %v", err)
		}

		// get mev data
		relaysData, err := db.GetRelayDataForIndexedBlocks(execBlocks)
		if err != nil {
			return nil, nil, fmt.Errorf("error retrieving mev bribe data: %v", err)
		}

		incomeTodayEl := new(big.Int)
		for _, execBlock := range execBlocks {

			blockEpoch := utils.TimeToEpoch(execBlock.Time.AsTime())
			if blockEpoch > int64(latestFinalizedEpoch) {
				continue
			}
			// add mev bribe if present
			if relaysDatum, hasMevBribes := relaysData[common.BytesToHash(execBlock.Hash)]; hasMevBribes {
				incomeTodayEl = new(big.Int).Add(incomeTodayEl, relaysDatum.MevBribe.Int)
			} else {
				incomeTodayEl = new(big.Int).Add(incomeTodayEl, new(big.Int).SetBytes(execBlock.GetTxReward()))
			}
		}
		incomeToday.El = int64(eth.WeiToGwei(incomeTodayEl))
		incomeToday.Total += float64(incomeToday.El) * elClPrice
	}

	incomeTotal := types.ClElInt64{
		El:    income.ElIncomeTotal + incomeToday.El,
		Cl:    income.ClIncomeTotal + incomeToday.Cl,
		Total: float64(income.ClIncomeTotal+incomeToday.Cl) + elClPrice*float64(income.ElIncomeTotal+incomeToday.El),
	}

	incomeTotalProposer := types.ClElInt64{
		El:    income.ElIncomeTotal,
		Cl:    income.ClProposerIncomeTotal + currentDayProposerIncome,
		Total: float64(income.ClProposerIncomeTotal) + elClPrice*float64(income.ElIncomeTotal) + float64(currentDayProposerIncome),
	}

	return &types.ValidatorEarnings{
		Income1d: types.ClElInt64{
			El:    income.ElIncome1d,
			Cl:    income.ClIncome1d,
			Total: earnings1d,
		},
		Income7d: types.ClElInt64{
			El:    income.ElIncome7d,
			Cl:    income.ClIncome7d,
			Total: earnings7d,
		},
		Income31d: types.ClElInt64{
			El:    income.ElIncome31d,
			Cl:    income.ClIncome31d,
			Total: earnings31d,
		},
		IncomeToday: incomeToday,
		IncomeTotal: incomeTotal,
		Apr7d: types.ClElFloat64{
			El:    elApr7d,
			Cl:    clApr7d,
			Total: clApr7d + elApr7d,
		},
		Apr31d: types.ClElFloat64{
			El:    elApr31d,
			Cl:    clApr31d,
			Total: clApr31d + elApr31d,
		},
		Apr365d: types.ClElFloat64{
			El:    elApr365d,
			Cl:    clApr365d,
			Total: clApr365d + elApr365d,
		},
		TotalDeposits:          int64(totalDeposits),
		LastDayFormatted:       utils.FormatIncome(int64(earnings1d), currency, true),
		LastWeekFormatted:      utils.FormatIncome(int64(earnings7d), currency, true),
		LastMonthFormatted:     utils.FormatIncome(int64(earnings31d), currency, true),
		TotalFormatted:         utils.FormatIncomeClElInt64(incomeTotal, currency),
		ProposerTotalFormatted: utils.FormatIncomeClElInt64(incomeTotalProposer, currency),
		TotalChangeFormatted:   utils.FormatIncome(income.ClIncomeTotal+currentDayClIncome+int64(totalDeposits), currency, true),
		TotalBalance:           utils.FormatIncome(int64(totalBalance), currency, true),
		ProposalData:           validatorProposalData,
	}, balancesMap, nil
}

func getProposalLuckBlockLookbackAmount(validatorCount int) int {
	switch {
	case validatorCount <= 4:
		return 10
	case validatorCount <= 10:
		return 15
	case validatorCount <= 20:
		return 20
	case validatorCount <= 50:
		return 30
	case validatorCount <= 100:
		return 50
	case validatorCount <= 200:
		return 65
	default:
		return 75
	}
}

// getProposalLuck calculates the luck of a given set of proposed blocks for a certain number of validators
// given the blocks proposed by the validators and the number of validators
//
// precondition: slots is sorted by ascending block number
func getProposalLuck(slots []uint64, validatorsCount int) float64 {
	// Return 0 if there are no proposed blocks or no validators
	if len(slots) == 0 || validatorsCount == 0 {
		return 0
	}
	// Timeframe constants
	fiveDays := utils.Day * 5
	oneWeek := utils.Week
	oneMonth := utils.Month
	sixWeeks := utils.Day * 45
	twoMonths := utils.Month * 2
	threeMonths := utils.Month * 3
	fourMonths := utils.Month * 4
	fiveMonths := utils.Month * 5
	sixMonths := utils.Month * 6
	year := utils.Year

	activeValidatorsCount := *services.GetLatestStats().ActiveValidatorCount
	// Calculate the expected number of slot proposals for 30 days
	expectedSlotProposals := calcExpectedSlotProposals(oneMonth, validatorsCount, activeValidatorsCount)

	// Get the timeframe for which we should consider qualified proposals
	var proposalTimeframe time.Duration
	// Time since the first block in the proposed block slice
	timeSinceFirstBlock := time.Since(utils.SlotToTime(slots[0]))

	targetBlocks := 8.0

	// Determine the appropriate timeframe based on the time since the first block and the expected slot proposals
	switch {
	case timeSinceFirstBlock < fiveDays:
		proposalTimeframe = fiveDays
	case timeSinceFirstBlock < oneWeek:
		proposalTimeframe = oneWeek
	case timeSinceFirstBlock < oneMonth:
		proposalTimeframe = oneMonth
	case timeSinceFirstBlock > year && expectedSlotProposals <= targetBlocks/12:
		proposalTimeframe = year
	case timeSinceFirstBlock > sixMonths && expectedSlotProposals <= targetBlocks/6:
		proposalTimeframe = sixMonths
	case timeSinceFirstBlock > fiveMonths && expectedSlotProposals <= targetBlocks/5:
		proposalTimeframe = fiveMonths
	case timeSinceFirstBlock > fourMonths && expectedSlotProposals <= targetBlocks/4:
		proposalTimeframe = fourMonths
	case timeSinceFirstBlock > threeMonths && expectedSlotProposals <= targetBlocks/3:
		proposalTimeframe = threeMonths
	case timeSinceFirstBlock > twoMonths && expectedSlotProposals <= targetBlocks/2:
		proposalTimeframe = twoMonths
	case timeSinceFirstBlock > sixWeeks && expectedSlotProposals <= targetBlocks/1.5:
		proposalTimeframe = sixWeeks
	default:
		proposalTimeframe = oneMonth
	}

	// Recalculate expected slot proposals for the new timeframe
	expectedSlotProposals = calcExpectedSlotProposals(proposalTimeframe, validatorsCount, activeValidatorsCount)
	if expectedSlotProposals == 0 {
		return 0
	}
	// Cutoff time for proposals to be considered qualified
	blockProposalCutoffTime := time.Now().Add(-proposalTimeframe)

	// Count the number of qualified proposals
	qualifiedProposalCount := 0
	for _, slot := range slots {
		if utils.SlotToTime(slot).After(blockProposalCutoffTime) {
			qualifiedProposalCount++
		}
	}
	// Return the luck as the ratio of qualified proposals to expected slot proposals
	return float64(qualifiedProposalCount) / expectedSlotProposals
}

// calcExpectedSlotProposals calculates the expected number of slot proposals for a certain time frame and validator count
func calcExpectedSlotProposals(timeframe time.Duration, validatorCount int, activeValidatorsCount uint64) float64 {
	if validatorCount == 0 || activeValidatorsCount == 0 {
		return 0
	}
	slotsInTimeframe := timeframe.Seconds() / float64(utils.Config.Chain.Config.SecondsPerSlot)
	return (slotsInTimeframe / float64(activeValidatorsCount)) * float64(validatorCount)
}

// getAvgSlotInterval will return the average block interval for a certain number of validators
//
// result of the function should be interpreted as "1 in every X slots will be proposed by this amount of validators on avg."
func getAvgSlotInterval(validatorsCount int) float64 {
	// don't estimate if there are no proposed blocks or no validators
	activeValidatorsCount := *services.GetLatestStats().ActiveValidatorCount
	if activeValidatorsCount == 0 {
		return 0
	}

	probability := float64(validatorsCount) / float64(activeValidatorsCount)
	// in a geometric distribution, the expected value of the number of trials needed until first success is 1/p
	// you can think of this as the average interval of blocks until you get a proposal
	return 1 / probability
}

// getAvgSyncCommitteeInterval will return the average sync committee interval for a certain number of validators
//
// result of the function should be interpreted as "there will be one validator included in every X committees, on average"
func getAvgSyncCommitteeInterval(validatorsCount int) float64 {
	activeValidatorsCount := *services.GetLatestStats().ActiveValidatorCount
	if activeValidatorsCount == 0 {
		return 0
	}

	probability := (float64(utils.Config.Chain.Config.SyncCommitteeSize) / float64(activeValidatorsCount)) * float64(validatorsCount)
	// in a geometric distribution, the expected value of the number of trials needed until first success is 1/p
	// you can think of this as the average interval of sync committees until you expect to have been part of one
	return 1 / probability
}

// LatestState will return common information that about the current state of the eth2 chain
func LatestState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", utils.Config.Chain.Config.SecondsPerSlot)) // set local cache to the seconds per slot interval
	currency := GetCurrency(r)
	data := services.LatestState()
	// data.Currency = currency
	data.EthPrice = price.GetPrice(utils.Config.Frontend.ClCurrencySymbol, currency)
	data.EthRoundPrice = uint64(data.EthPrice)
	data.EthTruncPrice = utils.KFormatterEthPrice(data.EthRoundPrice)

	err := json.NewEncoder(w).Encode(data)

	if err != nil {
		logger.Errorf("error sending latest index page data: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func GetCurrency(r *http.Request) string {
	if cookie, err := r.Cookie("currency"); err == nil {
		return cookie.Value
	}

	return utils.Config.Frontend.MainCurrencySymbol
}

func GetTickerCurrency(r *http.Request) string {
	cookie, err := r.Cookie("currency")
	if err != nil {
		return "USD"
	}
	if cookie.Value == utils.Config.Frontend.MainCurrencySymbol {
		return "USD"
	}
	return cookie.Value
}

func GetCurrencySymbol(r *http.Request) string {
	cookie, err := r.Cookie("currency")
	if err != nil {
		return "$"
	}
	return price.GetCurrencySymbol(cookie.Value)
}

func GetCurrentPrice(r *http.Request) uint64 {
	cookie, err := r.Cookie("currency")
	if err != nil {
		return uint64(price.GetPrice(utils.Config.Frontend.MainCurrencySymbol, "USD"))
	}

	if cookie.Value == utils.Config.Frontend.MainCurrencySymbol {
		return uint64(price.GetPrice(utils.Config.Frontend.MainCurrencySymbol, "USD"))
	}
	return uint64(price.GetPrice(utils.Config.Frontend.MainCurrencySymbol, cookie.Value))
}

func GetCurrentElPrice(r *http.Request) uint64 {
	cookie, err := r.Cookie("currency")
	if err != nil {
		return uint64(price.GetPrice(utils.Config.Frontend.ElCurrencySymbol, "USD"))
	}

	if cookie.Value == utils.Config.Frontend.ElCurrencySymbol {
		return uint64(price.GetPrice(utils.Config.Frontend.ElCurrencySymbol, "USD"))
	}
	return uint64(price.GetPrice(utils.Config.Frontend.ElCurrencySymbol, cookie.Value))
}

func GetCurrentPriceFormatted(r *http.Request) template.HTML {
	userAgent := r.Header.Get("User-Agent")
	userAgent = strings.ToLower(userAgent)
	price := GetCurrentPrice(r)
	if strings.Contains(userAgent, "android") || strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "windows phone") {
		return utils.KFormatterEthPrice(price)
	}
	return utils.FormatAddCommas(uint64(price))
}

func GetCurrentPriceKFormatted(r *http.Request) template.HTML {
	return utils.KFormatterEthPrice(GetCurrentPrice(r))
}

func GetTruncCurrentPriceFormatted(r *http.Request) string {
	price := GetCurrentPrice(r)
	symbol := GetCurrencySymbol(r)
	return fmt.Sprintf("%s %s", symbol, utils.KFormatterEthPrice(price))
}

// GetValidatorIndexFrom gets the validator index from users input
func GetValidatorIndexFrom(userInput string) (pubKey []byte, validatorIndex uint64, err error) {
	validatorIndex, err = strconv.ParseUint(userInput, 10, 64)
	if err == nil {
		pubKey, err = db.GetValidatorPublicKey(validatorIndex)
		return
	}

	pubKey, err = hex.DecodeString(strings.Replace(userInput, "0x", "", -1))
	if err == nil {
		validatorIndex, err = db.GetValidatorIndex(pubKey)
		return
	}
	return
}

func DataTableStateChanges(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user, session, err := getUserSession(r)
	if err != nil {
		logger.Errorf("error retrieving session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := &types.ApiResponse{}
	response.Status = "ERROR"

	defer json.NewEncoder(w).Encode(response)

	settings := types.DataTableSaveState{}
	err = json.NewDecoder(r.Body).Decode(&settings)
	if err != nil {
		logger.Errorf("error saving data table state could not parse body: %v", err)
		response.Status = "error saving table state"
		return
	}

	// never store the page number
	settings.Start = 0

	key := settings.Key
	if len(key) == 0 {
		logger.Errorf("no key provided")
		response.Status = "error saving table state"
		return
	}

	if !user.Authenticated {
		dataTableStatePrefix := "table:state:" + utils.GetNetwork() + ":"
		key = dataTableStatePrefix + key
		count := 0
		for k := range session.Values() {
			k, ok := k.(string)
			if ok && strings.HasPrefix(k, dataTableStatePrefix) {
				count += 1
			}
		}
		if count > 50 {
			_, ok := session.Values()[key]
			if !ok {
				logger.Errorf("error maximum number of datatable states stored in session")
				return
			}
		}
		session.Values()[key] = settings

		err := session.Save(r, w)
		if err != nil {
			logger.WithError(err).Errorf("error updating session with key: %v and value: %v", key, settings)
		}

	} else {
		err = db.SaveDataTableState(user.UserID, settings.Key, settings)
		if err != nil {
			logger.Errorf("error saving data table state could save values to db: %v", err)
			response.Status = "error saving table state"
			return
		}
	}

	response.Status = "OK"
	response.Data = ""
}

func GetDataTableState(user *types.User, session *utils.CustomSession, tableKey string) *types.DataTableSaveState {
	state := types.DataTableSaveState{
		Start: 0,
	}
	if user.Authenticated {
		state, err := db.GetDataTablesState(user.UserID, tableKey)
		if err != nil && err != sql.ErrNoRows {
			logger.Errorf("error getting data table state from db: %v", err)
			return state
		}
		return state
	}
	stateRaw, exists := session.Values()["table:state:"+utils.GetNetwork()+":"+tableKey]
	if !exists {
		return &state
	}
	state, ok := stateRaw.(types.DataTableSaveState)
	if !ok {
		logger.Errorf("error getting state from session: %+v", stateRaw)
		return &state
	}
	return &state
}

// used to handle errors constructed by Template.ExecuteTemplate correctly
func handleTemplateError(w http.ResponseWriter, r *http.Request, fileIdentifier string, functionIdentifier string, infoIdentifier string, err error) error {
	// ignore network related errors
	if err != nil && !errors.Is(err, syscall.EPIPE) && !errors.Is(err, syscall.ETIMEDOUT) {
		logger.WithFields(logrus.Fields{
			"file":       fileIdentifier,
			"function":   functionIdentifier,
			"info":       infoIdentifier,
			"error type": fmt.Sprintf("%T", err),
			"route":      r.URL.String(),
		}).WithError(err).Error("error executing template")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
	}
	return err
}

func GetWithdrawableCountFromCursor(epoch uint64, validatorindex uint64, cursor uint64) (uint64, error) {
	// the validators' balance will not be checked here as this is only a rough estimation
	// checking the balance for hundreds of thousands of validators is too expensive

	var maxValidatorIndex uint64
	err := db.WriterDb.Get(&maxValidatorIndex, "SELECT COALESCE(MAX(validatorindex), 0) FROM validators")
	if err != nil {
		return 0, fmt.Errorf("error getting withdrawable validator count from cursor: %w", err)
	}

	if maxValidatorIndex == 0 {
		return 0, nil
	}

	activeValidators := services.LatestIndexPageData().ActiveValidators
	if activeValidators == 0 {
		activeValidators = maxValidatorIndex
	}

	if validatorindex > cursor {
		// if the validatorindex is after the cursor, simply return the number of validators between the cursor and the validatorindex
		// the returned data is then scaled using the number of currently active validators in order to account for exited / entering validators
		return (validatorindex - cursor) * activeValidators / maxValidatorIndex, nil
	} else if validatorindex < cursor {
		// if the validatorindex is before the cursor (wraparound case) return the number of validators between the cursor and the most recent validator plus the amount of validators from the validator 0 to the validatorindex
		// the returned data is then scaled using the number of currently active validators in order to account for exited / entering validators
		return (maxValidatorIndex - cursor + validatorindex) * activeValidators / maxValidatorIndex, nil
	} else {
		return 0, nil
	}
}
