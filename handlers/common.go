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
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
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

// GetValidatorEarnings will return the earnings (last day, week, month and total) of selected validators
func GetValidatorEarnings(validators []uint64, currency string) (*types.ValidatorEarnings, map[uint64]*types.Validator, error) {
	validatorsPQArray := pq.Array(validators)
	latestEpoch := int64(services.LatestFinalizedEpoch())

	balancesMap := make(map[uint64]*types.Validator, 0)

	latestBalances, err := db.BigtableClient.GetValidatorBalanceHistory(validators, uint64(latestEpoch), uint64(latestEpoch))
	if err != nil {
		logger.Errorf("error getting validator balance data in GetValidatorEarnings: %v", err)
		return nil, nil, err
	}

	totalBalance := uint64(0)
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

	var income struct {
		ClIncome1d            int64 `db:"cl_performance_1d"`
		ClIncome7d            int64 `db:"cl_performance_7d"`
		ClIncome31d           int64 `db:"cl_performance_31d"`
		ClIncome365d          int64 `db:"cl_performance_365d"`
		ClIncomeTotal         int64 `db:"cl_performance_total"`
		ClProposerIncomeTotal int64 `db:"cl_proposer_performance_total"`
		ElIncome1d            int64 `db:"el_performance_1d"`
		ElIncome7d            int64 `db:"el_performance_7d"`
		ElIncome31d           int64 `db:"el_performance_31d"`
		ElIncome365d          int64 `db:"el_performance_365d"`
		ElIncomeTotal         int64 `db:"el_performance_total"`
		ClIncomeToday         int64
	}

	// el rewards are converted from wei to gwei
	err = db.ReaderDb.Get(&income, `
		SELECT 
		COALESCE(SUM(cl_performance_1d), 0) AS cl_performance_1d,
		COALESCE(SUM(cl_performance_7d), 0) AS cl_performance_7d,
		COALESCE(SUM(cl_performance_31d), 0) AS cl_performance_31d,
		COALESCE(SUM(cl_performance_365d), 0) AS cl_performance_365d,
		COALESCE(SUM(cl_performance_total), 0) AS cl_performance_total,
		COALESCE(SUM(cl_proposer_performance_total), 0) AS cl_proposer_performance_total,
		CAST(COALESCE(SUM(mev_performance_1d), 0) / 1e9 AS bigint) AS el_performance_1d,
		CAST(COALESCE(SUM(mev_performance_7d), 0) / 1e9 AS bigint) AS el_performance_7d,
		CAST(COALESCE(SUM(mev_performance_31d), 0) / 1e9 AS bigint) AS el_performance_31d,
		CAST(COALESCE(SUM(mev_performance_365d), 0) / 1e9 AS bigint) AS el_performance_365d,
		CAST(COALESCE(SUM(mev_performance_total), 0) / 1e9 AS bigint) AS el_performance_total
		FROM validator_performance WHERE validatorindex = ANY($1)`, validatorsPQArray)
	if err != nil {
		return nil, nil, err
	}

	var totalDeposits uint64

	err = db.ReaderDb.Get(&totalDeposits, `
	SELECT 
		COALESCE(SUM(amount), 0) 
	FROM blocks_deposits d
	INNER JOIN blocks b ON b.blockroot = d.block_root AND b.status = '1' 
	WHERE publickey IN (SELECT pubkey FROM validators WHERE validatorindex = ANY($1))`, validatorsPQArray)
	if err != nil {
		return nil, nil, err
	}
	if totalDeposits == 0 {
		totalDeposits = utils.Config.Chain.Config.MaxEffectiveBalance
	}

	var totalWithdrawals uint64

	err = db.ReaderDb.Get(&totalWithdrawals, `
	SELECT 
		COALESCE(sum(w.amount), 0)
	FROM blocks_withdrawals w
	INNER JOIN blocks b ON b.blockroot = w.block_root AND b.status = '1'
	WHERE validatorindex = ANY($1)
	`, validatorsPQArray)
	if err != nil {
		return nil, nil, err
	}

	// calculate combined el and cl earnings
	earnings1d := income.ClIncome1d + income.ElIncome1d
	earnings7d := income.ClIncome7d + income.ElIncome7d
	earnings31d := income.ClIncome31d + income.ElIncome31d

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
	currentDayIncome, currentDayProposerIncome, err := db.GetCurrentDayClIncomeTotal(validators)
	if err != nil {
		return nil, nil, err
	}

	incomeTotal := types.ClElInt64{
		El:    income.ElIncomeTotal,
		Cl:    income.ClIncomeTotal + currentDayIncome,
		Total: income.ClIncomeTotal + income.ElIncomeTotal + currentDayIncome,
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
		TotalChangeFormatted:   utils.FormatIncome(income.ClIncomeTotal+currentDayIncome+int64(totalDeposits), currency),
		TotalBalance:           utils.FormatIncome(int64(totalBalance), currency),
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
