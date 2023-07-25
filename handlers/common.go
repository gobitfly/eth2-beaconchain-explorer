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
	"github.com/gorilla/mux"
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
	firstSlot := utils.GetLastBalanceInfoSlotForDay(lastStatsDay) + 1
	lastSlot := latestFinalizedEpoch * utils.Config.Chain.Config.SlotsPerEpoch

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

	var firstActivationEpoch uint64
	g.Go(func() error {
		return db.GetFirstActivationEpoch(validators, &firstActivationEpoch)
	})

	var totalWithdrawals uint64
	g.Go(func() error {
		return db.GetTotalValidatorWithdrawals(validators, &totalWithdrawals)
	})

	var lastDeposits uint64
	g.Go(func() error {
		return db.GetValidatorDepositsForSlots(validators, firstSlot, lastSlot, &lastDeposits)
	})

	var lastWithdrawals uint64
	g.Go(func() error {
		return db.GetValidatorWithdrawalsForSlots(validators, firstSlot, lastSlot, &lastWithdrawals)
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

	// calculate combined el and cl earnings
	earnings1d := income.ClIncome1d + income.ElIncome1d
	earnings7d := income.ClIncome7d + income.ElIncome7d
	earnings31d := income.ClIncome31d + income.ElIncome31d

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
		Total: currentDayClIncome,
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

	validatorProposalData.ProposalLuck = getProposalLuck(slots, len(validators), firstActivationEpoch)
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
		incomeToday.Total += incomeToday.El
	}

	incomeTotal := types.ClElInt64{
		El:    income.ElIncomeTotal + incomeToday.El,
		Cl:    income.ClIncomeTotal + incomeToday.Cl,
		Total: income.ClIncomeTotal + income.ElIncomeTotal + incomeToday.Total,
	}

	incomeTotalProposer := types.ClElInt64{
		El:    income.ElIncomeTotal,
		Cl:    income.ClProposerIncomeTotal + currentDayProposerIncome,
		Total: income.ClProposerIncomeTotal + income.ElIncomeTotal + currentDayProposerIncome,
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
		LastDayFormatted:       utils.FormatIncome(earnings1d, currency),
		LastWeekFormatted:      utils.FormatIncome(earnings7d, currency),
		LastMonthFormatted:     utils.FormatIncome(earnings31d, currency),
		TotalFormatted:         utils.FormatIncomeClElInt64(incomeTotal, currency),
		ProposerTotalFormatted: utils.FormatIncomeClElInt64(incomeTotalProposer, currency),
		TotalChangeFormatted:   utils.FormatIncome(income.ClIncomeTotal+currentDayClIncome+int64(totalDeposits), currency),
		TotalBalance:           utils.FormatIncome(int64(totalBalance), currency),
		ProposalData:           validatorProposalData,
	}, balancesMap, nil
}

// getProposalLuck calculates the luck of a given set of proposed blocks for a certain number of validators
// given the blocks proposed by the validators and the number of validators
//
// precondition: slots is sorted by ascending block number
func getProposalLuck(slots []uint64, validatorsCount int, fromEpoch uint64) float64 {
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
	// Time since the first epoch of the related validators
	timeSinceFirstEpoch := time.Since(utils.EpochToTime(fromEpoch))

	targetBlocks := 8.0

	// Determine the appropriate timeframe based on the time since the first block and the expected slot proposals
	switch {
	case timeSinceFirstEpoch < fiveDays:
		proposalTimeframe = fiveDays
	case timeSinceFirstEpoch < oneWeek:
		proposalTimeframe = oneWeek
	case timeSinceFirstEpoch < oneMonth:
		proposalTimeframe = oneMonth
	case timeSinceFirstEpoch > year && expectedSlotProposals <= targetBlocks/12:
		proposalTimeframe = year
	case timeSinceFirstEpoch > sixMonths && expectedSlotProposals <= targetBlocks/6:
		proposalTimeframe = sixMonths
	case timeSinceFirstEpoch > fiveMonths && expectedSlotProposals <= targetBlocks/5:
		proposalTimeframe = fiveMonths
	case timeSinceFirstEpoch > fourMonths && expectedSlotProposals <= targetBlocks/4:
		proposalTimeframe = fourMonths
	case timeSinceFirstEpoch > threeMonths && expectedSlotProposals <= targetBlocks/3:
		proposalTimeframe = threeMonths
	case timeSinceFirstEpoch > twoMonths && expectedSlotProposals <= targetBlocks/2:
		proposalTimeframe = twoMonths
	case timeSinceFirstEpoch > sixWeeks && expectedSlotProposals <= targetBlocks/1.5:
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
	data.EthPrice = price.GetEthPrice(currency)
	data.EthRoundPrice = price.GetEthRoundPrice(data.EthPrice)
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

	return "ETH"
}

func GetCurrencySymbol(r *http.Request) string {

	cookie, err := r.Cookie("currency")
	if err != nil {
		return "$"
	}

	switch cookie.Value {
	case "AUD":
		return "A$"
	case "CAD":
		return "C$"
	case "CNY":
		return "¥"
	case "EUR":
		return "€"
	case "GBP":
		return "£"
	case "JPY":
		return "¥"
	case "RUB":
		return "₽"
	default:
		return "$"
	}
}

func GetCurrentPrice(r *http.Request) uint64 {
	cookie, err := r.Cookie("currency")
	if err != nil {
		return price.GetEthRoundPrice(price.GetEthPrice("USD"))
	}

	if cookie.Value == "ETH" {
		return price.GetEthRoundPrice(price.GetEthPrice("USD"))
	}
	return price.GetEthRoundPrice(price.GetEthPrice(cookie.Value))
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
	validatorIndex, err = strconv.ParseUint(userInput, 10, 32)
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
		utils.LogError(err, errMsgPrefix+", could not parse body", 0, errFields)
		w.WriteHeader(http.StatusInternalServerError)
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
