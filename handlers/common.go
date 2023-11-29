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
	"math"
	"math/big"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	utilMath "github.com/protolambda/zrnt/eth2/util/math"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func GetValidatorOnlineThresholdSlot() uint64 {
	latestProposedSlot := services.LatestProposedSlot()
	threshold := utils.Config.Chain.ClConfig.SlotsPerEpoch * 2

	var validatorOnlineThresholdSlot uint64
	if latestProposedSlot < 1 || latestProposedSlot < threshold {
		validatorOnlineThresholdSlot = 0
	} else {
		validatorOnlineThresholdSlot = latestProposedSlot - threshold
	}

	return validatorOnlineThresholdSlot
}

// GetValidatorEarnings will return the earnings (last day, week, month and total) of selected validators, including proposal and statisic information - infused with data from the current day. all values are
func GetValidatorEarnings(validators []uint64, currency string) (*types.ValidatorEarnings, map[uint64]*types.Validator, error) {
	if len(validators) == 0 {
		return nil, nil, errors.New("no validators provided")
	}
	latestFinalizedEpoch := services.LatestFinalizedEpoch()

	firstSlot := uint64(0)
	lastStatsDay, lastExportedStatsErr := services.LatestExportedStatisticDay()
	if lastExportedStatsErr == nil {
		firstSlot = utils.GetLastBalanceInfoSlotForDay(lastStatsDay) + 1
	}

	lastSlot := latestFinalizedEpoch * utils.Config.Chain.ClConfig.SlotsPerEpoch

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
		return db.GetValidatorIncomePerformance(validators, &income)
	})

	var totalDeposits uint64
	g.Go(func() error {
		return db.GetTotalValidatorDeposits(validators, &totalDeposits)
	})

	var firstActivationEpoch uint64
	g.Go(func() error {
		return db.GetFirstActivationEpoch(validators, &firstActivationEpoch)
	})

	var lastDeposits uint64
	var lastWithdrawals uint64
	var lastBalance uint64
	g.Go(func() error {
		if lastExportedStatsErr == db.ErrNoStats {
			err := db.GetValidatorActivationBalance(validators, &lastBalance)
			if err != nil {
				return err
			}
		} else {
			err := db.GetValidatorBalanceForDay(validators, lastStatsDay, &lastBalance)
			if err != nil {
				return err
			}
		}
		err := db.GetValidatorDepositsForSlots(validators, firstSlot, lastSlot, &lastDeposits)
		if err != nil {
			return err
		}
		return db.GetValidatorWithdrawalsForSlots(validators, firstSlot, lastSlot, &lastWithdrawals)
	})

	proposals := []types.ValidatorProposalInfo{}
	g.Go(func() error {
		return db.GetValidatorPropsosals(validators, &proposals)
	})

	err := g.Wait()
	if err != nil {
		return nil, nil, err
	}

	clElPrice := price.GetPrice(utils.Config.Frontend.ClCurrency, utils.Config.Frontend.ElCurrency)

	if totalDeposits == 0 {
		totalDeposits = utils.Config.Chain.ClConfig.MaxEffectiveBalance * uint64(len(validators))
	}

	clApr7d := income.ClIncomeWei7d.DivRound(decimal.NewFromInt(1e9), 18).DivRound(decimal.NewFromInt(int64(totalDeposits)), 18).Mul(decimal.NewFromInt(365)).Div(decimal.NewFromInt(7)).InexactFloat64()
	if clApr7d < float64(-1) {
		clApr7d = float64(-1)
	}
	if math.IsNaN(clApr7d) {
		clApr7d = float64(0)
	}

	elApr7d := income.ElIncomeWei7d.DivRound(decimal.NewFromInt(1e9), 18).DivRound(decimal.NewFromInt(int64(totalDeposits)), 18).Mul(decimal.NewFromInt(365)).Div(decimal.NewFromInt(7)).InexactFloat64()
	if elApr7d < float64(-1) {
		elApr7d = float64(-1)
	}
	if math.IsNaN(elApr7d) {
		elApr7d = float64(0)
	}

	clApr31d := income.ClIncomeWei31d.DivRound(decimal.NewFromInt(1e9), 18).DivRound(decimal.NewFromInt(int64(totalDeposits)), 18).Mul(decimal.NewFromInt(365)).Div(decimal.NewFromInt(31)).InexactFloat64()
	if clApr31d < float64(-1) {
		clApr31d = float64(-1)
	}
	if math.IsNaN(clApr31d) {
		clApr31d = float64(0)
	}

	elApr31d := income.ElIncomeWei31d.DivRound(decimal.NewFromInt(1e9), 18).DivRound(decimal.NewFromInt(int64(totalDeposits)), 18).Mul(decimal.NewFromInt(365)).Div(decimal.NewFromInt(31)).InexactFloat64()
	if elApr31d < float64(-1) {
		elApr31d = float64(-1)
	}
	if math.IsNaN(elApr31d) {
		elApr31d = float64(0)
	}

	clApr365d := income.ClIncomeWei365d.DivRound(decimal.NewFromInt(1e9), 18).DivRound(decimal.NewFromInt(int64(totalDeposits)), 18).InexactFloat64()
	if clApr365d < float64(-1) {
		clApr365d = float64(-1)
	}
	if math.IsNaN(clApr365d) {
		clApr365d = float64(0)
	}

	elApr365d := income.ElIncomeWei365d.DivRound(decimal.NewFromInt(1e9), 18).DivRound(decimal.NewFromInt(int64(totalDeposits)), 18).InexactFloat64()
	if elApr365d < float64(-1) {
		elApr365d = float64(-1)
	}
	if math.IsNaN(elApr365d) {
		elApr365d = float64(0)
	}

	proposedToday := []uint64{}
	todayStartEpoch := uint64(0)
	if lastExportedStatsErr == nil {
		todayStartEpoch = uint64(lastStatsDay+1) * utils.EpochsPerDay()
	}
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

	validatorProposalData.ProposalLuck, _ = getProposalLuck(slots, len(validators), firstActivationEpoch)
	avgSlotInterval := uint64(getAvgSlotInterval(len(validators)))
	avgSlotIntervalAsDuration := time.Duration(utils.Config.Chain.ClConfig.SecondsPerSlot*avgSlotInterval) * time.Second
	validatorProposalData.AvgSlotInterval = &avgSlotIntervalAsDuration
	if len(slots) > 0 {
		nextSlotEstimate := utils.SlotToTime(slots[len(slots)-1] + avgSlotInterval)
		validatorProposalData.ProposalEstimate = &nextSlotEstimate
	}

	currentDayClIncome := decimal.NewFromInt(int64(totalBalance - lastBalance - lastDeposits + lastWithdrawals)).Mul(decimal.NewFromInt(1e9))
	incomeToday := types.ClEl{
		El:    decimal.NewFromInt(0),
		Cl:    currentDayClIncome.Mul(decimal.NewFromFloat(clElPrice)),
		Total: currentDayClIncome.Mul(decimal.NewFromFloat(clElPrice)),
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
		incomeToday.El = decimal.NewFromBigInt(incomeTodayEl, 0)
		incomeToday.Total = incomeToday.Total.Add(incomeToday.El)
	}

	earnings := &types.ValidatorEarnings{
		IncomeToday: incomeToday,
		Income1d: types.ClEl{
			El:    income.ElIncomeWei1d,
			Cl:    income.ClIncomeWei1d.Mul(decimal.NewFromFloat(clElPrice)),
			Total: income.ElIncomeWei1d.Add(income.ClIncomeWei1d.Mul(decimal.NewFromFloat(clElPrice))),
		},
		Income7d: types.ClEl{
			El:    income.ElIncomeWei7d,
			Cl:    income.ClIncomeWei7d.Mul(decimal.NewFromFloat(clElPrice)),
			Total: income.ElIncomeWei7d.Add(income.ClIncomeWei7d.Mul(decimal.NewFromFloat(clElPrice))),
		},
		Income31d: types.ClEl{
			El:    income.ElIncomeWei31d,
			Cl:    income.ClIncomeWei31d.Mul(decimal.NewFromFloat(clElPrice)),
			Total: income.ElIncomeWei31d.Add(income.ClIncomeWei31d.Mul(decimal.NewFromFloat(clElPrice))),
		},
		IncomeTotal: types.ClEl{
			El:    income.ElIncomeWeiTotal.Add(incomeToday.El),
			Cl:    income.ClIncomeWeiTotal.Add(incomeToday.Cl).Mul(decimal.NewFromFloat(clElPrice)),
			Total: income.ElIncomeWeiTotal.Add(incomeToday.El).Add(income.ClIncomeWeiTotal.Add(incomeToday.Cl).Mul(decimal.NewFromFloat(clElPrice))),
		},
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
		TotalDeposits: int64(totalDeposits),
		ProposalData:  validatorProposalData,
	}
	earnings.LastDayFormatted = utils.FormatIncomeClEl(earnings.Income1d, currency)
	earnings.LastWeekFormatted = utils.FormatIncomeClEl(earnings.Income7d, currency)
	earnings.LastMonthFormatted = utils.FormatIncomeClEl(earnings.Income31d, currency)
	earnings.TotalFormatted = utils.FormatIncomeClEl(earnings.IncomeTotal, currency)
	earnings.TotalBalance = "<b>" + utils.FormatClCurrency(totalBalance, currency, 5, true, false, false, false) + "</b>"
	return earnings, balancesMap, nil
}

// Timeframe constants
const fiveDays = utils.Day * 5
const oneWeek = utils.Week
const oneMonth = utils.Month
const sixWeeks = utils.Day * 45
const twoMonths = utils.Month * 2
const threeMonths = utils.Month * 3
const fourMonths = utils.Month * 4
const fiveMonths = utils.Month * 5
const sixMonths = utils.Month * 6
const year = utils.Year

// getProposalLuck calculates the luck of a given set of proposed blocks for a certain number of validators
// given the blocks proposed by the validators and the number of validators
//
// precondition: slots is sorted by ascending block number
func getProposalLuck(slots []uint64, validatorsCount int, fromEpoch uint64) (float64, time.Duration) {
	// Return 0 if there are no proposed blocks or no validators
	if len(slots) == 0 || validatorsCount == 0 {
		return 0, 0
	}

	activeValidatorsCount := *services.GetLatestStats().ActiveValidatorCount
	// Calculate the expected number of slot proposals for 30 days
	expectedSlotProposals := calcExpectedSlotProposals(oneMonth, validatorsCount, activeValidatorsCount)

	// Get the timeframe for which we should consider qualified proposals
	var proposalTimeFrame time.Duration
	// Time since the first epoch of the related validators
	timeSinceFirstEpoch := time.Since(utils.EpochToTime(fromEpoch))

	targetBlocks := 8.0

	// Determine the appropriate timeframe based on the time since the first block and the expected slot proposals
	switch {
	case timeSinceFirstEpoch < fiveDays:
		proposalTimeFrame = fiveDays
	case timeSinceFirstEpoch < oneWeek:
		proposalTimeFrame = oneWeek
	case timeSinceFirstEpoch < oneMonth:
		proposalTimeFrame = oneMonth
	case timeSinceFirstEpoch > year && expectedSlotProposals <= targetBlocks/12:
		proposalTimeFrame = year
	case timeSinceFirstEpoch > sixMonths && expectedSlotProposals <= targetBlocks/6:
		proposalTimeFrame = sixMonths
	case timeSinceFirstEpoch > fiveMonths && expectedSlotProposals <= targetBlocks/5:
		proposalTimeFrame = fiveMonths
	case timeSinceFirstEpoch > fourMonths && expectedSlotProposals <= targetBlocks/4:
		proposalTimeFrame = fourMonths
	case timeSinceFirstEpoch > threeMonths && expectedSlotProposals <= targetBlocks/3:
		proposalTimeFrame = threeMonths
	case timeSinceFirstEpoch > twoMonths && expectedSlotProposals <= targetBlocks/2:
		proposalTimeFrame = twoMonths
	case timeSinceFirstEpoch > sixWeeks && expectedSlotProposals <= targetBlocks/1.5:
		proposalTimeFrame = sixWeeks
	default:
		proposalTimeFrame = oneMonth
	}

	// Recalculate expected slot proposals for the new timeframe
	expectedSlotProposals = calcExpectedSlotProposals(proposalTimeFrame, validatorsCount, activeValidatorsCount)
	if expectedSlotProposals == 0 {
		return 0, 0
	}
	// Cutoff time for proposals to be considered qualified
	blockProposalCutoffTime := time.Now().Add(-proposalTimeFrame)

	// Count the number of qualified proposals
	qualifiedProposalCount := 0
	for _, slot := range slots {
		if utils.SlotToTime(slot).After(blockProposalCutoffTime) {
			qualifiedProposalCount++
		}
	}
	// Return the luck as the ratio of qualified proposals to expected slot proposals
	return float64(qualifiedProposalCount) / expectedSlotProposals, proposalTimeFrame
}

func getProposalTimeframeName(proposalTimeframe time.Duration) string {
	switch {
	case proposalTimeframe == fiveDays:
		return "5 days"
	case proposalTimeframe == oneWeek:
		return "week"
	case proposalTimeframe == oneMonth:
		return "month"
	case proposalTimeframe == sixWeeks:
		return "6 weeks"
	case proposalTimeframe == twoMonths:
		return "2 months"
	case proposalTimeframe == threeMonths:
		return "3 months"
	case proposalTimeframe == fourMonths:
		return "4 months"
	case proposalTimeframe == fiveMonths:
		return "5 months"
	case proposalTimeframe == sixMonths:
		return "6 months"
	case proposalTimeframe == year:
		return "year"
	default:
		return "month"
	}
}

// calcExpectedSlotProposals calculates the expected number of slot proposals for a certain time frame and validator count
func calcExpectedSlotProposals(timeframe time.Duration, validatorCount int, activeValidatorsCount uint64) float64 {
	if validatorCount == 0 || activeValidatorsCount == 0 {
		return 0
	}
	slotsInTimeframe := timeframe.Seconds() / float64(utils.Config.Chain.ClConfig.SecondsPerSlot)
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

	probability := (float64(utils.Config.Chain.ClConfig.SyncCommitteeSize) / float64(activeValidatorsCount)) * float64(validatorsCount)
	// in a geometric distribution, the expected value of the number of trials needed until first success is 1/p
	// you can think of this as the average interval of sync committees until you expect to have been part of one
	return 1 / probability
}

// LatestState will return common information that about the current state of the eth2 chain
func LatestState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", utils.Config.Chain.ClConfig.SecondsPerSlot)) // set local cache to the seconds per slot interval

	data := services.LatestState()
	data.Rates = services.GetRates(GetCurrency(r))
	userAgent := r.Header.Get("User-Agent")
	userAgent = strings.ToLower(userAgent)
	if strings.Contains(userAgent, "android") || strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "windows phone") {
		data.Rates.MainCurrencyPriceFormatted = utils.KFormatterEthPrice(uint64(data.Rates.MainCurrencyPrice))
	}

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error sending latest index page data: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func GetCurrency(r *http.Request) string {
	if cookie, err := r.Cookie("currency"); err == nil {
		if price.IsAvailableCurrency(cookie.Value) {
			return cookie.Value
		}
	}
	return utils.Config.Frontend.MainCurrency
}

func GetCurrencySymbol(r *http.Request) string {
	cookie, err := r.Cookie("currency")
	if err != nil {
		logger.WithError(err).Tracef("error in handlers.GetCurrencySymbol")
		return "$"
	}
	if cookie.Value == utils.Config.Frontend.MainCurrency {
		return "USD"
	}
	return price.GetCurrencySymbol(cookie.Value)
}

func GetCurrentPrice(r *http.Request) uint64 {
	cookie, err := r.Cookie("currency")
	if err != nil {
		return uint64(price.GetPrice(utils.Config.Frontend.MainCurrency, "USD"))
	}
	if cookie.Value == utils.Config.Frontend.MainCurrency {
		return uint64(price.GetPrice(utils.Config.Frontend.MainCurrency, "USD"))
	}
	return uint64(price.GetPrice(utils.Config.Frontend.MainCurrency, cookie.Value))
}

func GetCurrentElPrice(r *http.Request) uint64 {
	cookie, err := r.Cookie("currency")
	if err != nil {
		return uint64(price.GetPrice(utils.Config.Frontend.ElCurrency, "USD"))
	}
	if cookie.Value == utils.Config.Frontend.ElCurrency {
		return uint64(price.GetPrice(utils.Config.Frontend.ElCurrency, "USD"))
	}
	return uint64(price.GetPrice(utils.Config.Frontend.ElCurrency, cookie.Value))
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

func GetCurrentElPriceFormatted(r *http.Request) template.HTML {
	userAgent := r.Header.Get("User-Agent")
	userAgent = strings.ToLower(userAgent)
	price := GetCurrentElPrice(r)
	if strings.Contains(userAgent, "android") || strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "windows phone") {
		return utils.KFormatterEthPrice(price)
	}
	return utils.FormatAddCommas(uint64(price))
}

func GetCurrentPriceKFormatted(r *http.Request) template.HTML {
	return utils.KFormatterEthPrice(GetCurrentPrice(r))
}

func GetCurrentElPriceKFormatted(r *http.Request) template.HTML {
	return utils.KFormatterEthPrice(GetCurrentElPrice(r))
}

func GetTruncCurrentPriceFormatted(r *http.Request) string {
	price := GetCurrentPrice(r)
	symbol := GetCurrencySymbol(r)
	return fmt.Sprintf("%s %s", symbol, utils.KFormatterEthPrice(price))
}

// GetValidatorKeysFrom gets the validator keys from users input
func GetValidatorKeysFrom(userInput []string) (pubKeys [][]byte, err error) {
	indexList := []uint64{}
	keyList := [][]byte{}
	for _, input := range userInput {

		validatorIndex, err := strconv.ParseUint(input, 10, 32)
		if err == nil {
			indexList = append(indexList, validatorIndex)
		}

		pubKey, err := hex.DecodeString(strings.Replace(input, "0x", "", -1))
		if err == nil {
			keyList = append(keyList, pubKey)
		}
	}

	pubKeys, err = db.GetValidatorPublicKeys(indexList, keyList)
	if len(pubKeys) != len(userInput) {
		err = fmt.Errorf("not all validators found in db")
	}
	return
}

func GetDataTableStateChanges(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	tableKey := vars["tableId"]

	errMsgPrefix := "error loading data table state"
	errFields := map[string]interface{}{
		"tableKey": tableKey}

	response := &types.ApiResponse{}
	response.Status = errMsgPrefix
	response.Data = ""

	defer json.NewEncoder(w).Encode(response)

	user, _, err := getUserSession(r)
	if err != nil {
		utils.LogError(err, errMsgPrefix+", could not retrieve user session", 0, errFields)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if user.Authenticated {
		state, err := db.GetDataTablesState(user.UserID, tableKey)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				utils.LogError(err, errMsgPrefix+", could not load values from db", 0, errFields)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			// the time for the state load of a data table must not be older than 2 hours so set it to the current time
			state.Time = uint64(time.Now().Unix() * 1000)

			response.Data = state
		}
	}

	response.Status = "OK"
}

func SetDataTableStateChanges(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	tableKey := vars["tableId"]

	errMsgPrefix := "error saving data table state"
	errFields := map[string]interface{}{
		"tableKey": tableKey}

	response := &types.ApiResponse{}
	response.Status = errMsgPrefix
	response.Data = ""

	defer json.NewEncoder(w).Encode(response)

	user, _, err := getUserSession(r)
	if err != nil {
		utils.LogError(err, errMsgPrefix+", could not retrieve user session", 0, errFields)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	settings := types.DataTableSaveState{}
	err = json.NewDecoder(r.Body).Decode(&settings)
	if err != nil {
		logger.Warnf(errMsgPrefix+", could not parse body for tableKey %v: %v", tableKey, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	settings.Key = tableKey

	// never store the page number
	settings.Start = 0

	if user.Authenticated {
		err = db.SaveDataTableState(user.UserID, settings.Key, settings)
		if err != nil {
			utils.LogError(err, errMsgPrefix+", could no save values to db", 0, errFields)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		response.Data = settings
	}

	response.Status = "OK"
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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

func getExecutionChartData(indices []uint64, currency string, lowerBoundDay uint64) ([]*types.ChartDataPoint, error) {
	var limit uint64 = 300
	blockList, consMap, err := findExecBlockNumbersByProposerIndex(indices, 0, limit, false, true, lowerBoundDay)
	if err != nil {
		return nil, err
	}

	blocks, err := db.BigtableClient.GetBlocksIndexedMultiple(blockList, limit)
	if err != nil {
		return nil, err
	}
	relaysData, err := db.GetRelayDataForIndexedBlocks(blocks)
	if err != nil {
		return nil, err
	}

	var chartData = []*types.ChartDataPoint{}
	epochsPerDay := utils.EpochsPerDay()
	color := "#90ed7d"

	// Map to keep track of the cumulative reward for each day
	dayRewardMap := make(map[int64]float64)

	for _, block := range blocks {
		consData := consMap[block.Number]
		day := int64(consData.Epoch / epochsPerDay)

		var totalReward float64
		if relayData, ok := relaysData[common.BytesToHash(block.Hash)]; ok {
			totalReward = utils.WeiToEther(relayData.MevBribe.BigInt()).InexactFloat64()
		} else {
			totalReward = utils.WeiToEther(utils.Eth1TotalReward(block)).InexactFloat64()
		}

		// Add the reward to the existing reward for the day or set it if not previously set
		dayRewardMap[day] += totalReward
	}

	// Now populate the chartData array using the dayRewardMap
	exchangeRate := price.GetPrice(utils.Config.Frontend.ElCurrency, currency)
	for day, reward := range dayRewardMap {
		ts := float64(utils.DayToTime(day).Unix() * 1000)
		chartData = append(chartData, &types.ChartDataPoint{
			X:     ts,
			Y:     exchangeRate * reward,
			Color: color,
		})
	}

	// If needed, sort chartData based on X values
	sort.Slice(chartData, func(i, j int) bool {
		return chartData[i].X < chartData[j].X
	})

	return chartData, nil
}
