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
	latestEpoch := int64(services.LatestEpoch())

	balances := []*types.Validator{}

	balancesMap := make(map[uint64]*types.Validator, len(balances))

	for _, balance := range balances {
		balancesMap[balance.Index] = balance
	}

	latestBalances, err := db.BigtableClient.GetValidatorBalanceHistory(validators, uint64(latestEpoch), uint64(latestEpoch))
	if err != nil {
		logger.Errorf("error getting validator balance data in GetValidatorEarnings: %v", err)
		return nil, nil, err
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
	}

	var income struct {
		ClIncome1d    int64 `db:"cl_performance_1d"`
		ClIncome7d    int64 `db:"cl_performance_7d"`
		ClIncome31d   int64 `db:"cl_performance_31d"`
		ClIncome365d  int64 `db:"cl_performance_365d"`
		ClIncomeTotal int64 `db:"cl_performance_total"`
		ElIncome1d    int64 `db:"el_performance_1d"`
		ElIncome7d    int64 `db:"el_performance_7d"`
		ElIncome31d   int64 `db:"el_performance_31d"`
		ElIncome365d  int64 `db:"el_performance_365d"`
		ElIncomeTotal int64 `db:"el_performance_total"`
	}

	err = db.ReaderDb.Get(&income, `
		SELECT 
			SUM(cl_performance_1d) AS cl_performance_1d,
			SUM(cl_performance_7d) AS cl_performance_7d,
			SUM(cl_performance_31d) AS cl_performance_31d,
			SUM(cl_performance_365d) AS cl_performance_365d,
			SUM(cl_performance_total) AS cl_performance_total,
			SUM(mev_performance_1d) AS el_performance_1d,
			SUM(mev_performance_7d) AS el_performance_7d,
			SUM(mev_performance_31d) AS el_performance_31d,
			SUM(mev_performance_365d) AS el_performance_365d,
			SUM(mev_performance_total) AS el_performance_total
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
	earnings1d := ((income.ClIncome1d * 1e9) + income.ElIncome1d) / 1e9
	earnings7d := ((income.ClIncome7d * 1e9) + income.ElIncome7d) / 1e9
	earnings31d := ((income.ClIncome31d * 1e9) + income.ElIncome31d) / 1e9

	// since only the first 5 digits are shown in the frontend, the lost precision is probably negligible
	income.ElIncome1d /= 1e9
	income.ElIncome7d /= 1e9
	income.ElIncome31d /= 1e9
	income.ElIncome365d /= 1e9
	income.ElIncomeTotal /= 1e9

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
		ElIncomeTotal:        income.ElIncomeTotal,
		TotalDeposits:        int64(totalDeposits),
		LastDayFormatted:     utils.FormatIncome(earnings1d, currency),
		LastWeekFormatted:    utils.FormatIncome(earnings7d, currency),
		LastMonthFormatted:   utils.FormatIncome(earnings31d, currency),
		TotalFormatted:       utils.FormatIncome(income.ClIncomeTotal, currency),
		TotalChangeFormatted: utils.FormatIncome(income.ClIncomeTotal+int64(totalDeposits), currency),
	}, balancesMap, nil
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
	fiveDays := time.Hour * 24 * 5
	oneWeek := time.Hour * 24 * 7
	oneMonth := time.Hour * 24 * 30
	sixWeeks := time.Hour * 24 * 45
	twoMonths := time.Hour * 24 * 60
	threeMonths := time.Hour * 24 * 90
	fourMonths := time.Hour * 24 * 120
	fiveMonths := time.Hour * 24 * 150

	activeValidatorsCount := *services.GetLatestStats().ActiveValidatorCount
	// Calculate the expected number of slot proposals for 30 days
	expectedSlotProposals := calcExpectedSlotProposals(oneMonth, validatorsCount, activeValidatorsCount)

	// Get the timeframe for which we should consider qualified proposals
	var proposalTimeframe time.Duration
	// Time since the first block in the proposed block slice
	timeSinceFirstBlock := time.Since(utils.SlotToTime(slots[0]))

	// Determine the appropriate timeframe based on the time since the first block and the expected slot proposals
	switch {
	case timeSinceFirstBlock < fiveDays:
		proposalTimeframe = fiveDays
	case timeSinceFirstBlock < oneWeek:
		proposalTimeframe = oneWeek
	case timeSinceFirstBlock < oneMonth:
		proposalTimeframe = oneMonth
	case timeSinceFirstBlock > fiveMonths && expectedSlotProposals <= 0.75:
		proposalTimeframe = fiveMonths
	case timeSinceFirstBlock > fourMonths && expectedSlotProposals <= 1:
		proposalTimeframe = fourMonths
	case timeSinceFirstBlock > threeMonths && expectedSlotProposals <= 1.4:
		proposalTimeframe = threeMonths
	case timeSinceFirstBlock > twoMonths && expectedSlotProposals <= 2.1:
		proposalTimeframe = twoMonths
	case timeSinceFirstBlock > sixWeeks && expectedSlotProposals <= 2.8:
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
	slotsPerTimeframe := timeframe.Seconds() / float64(utils.Config.Chain.Config.SecondsPerSlot)
	return (slotsPerTimeframe / float64(activeValidatorsCount)) * float64(validatorCount)
}

// getNextBlockEstimateTimestamp will return the estimated timestamp of the next block proposal
// given the blocks proposed by the validators and the number of validators
//
// precondition: proposedBlocks is sorted by ascending block number
func getNextBlockEstimateTimestamp(slots []uint64, validatorsCount int) *time.Time {
	// don't estimate if there are no proposed blocks or no validators
	if len(slots) == 0 || validatorsCount == 0 {
		return nil
	}
	activeValidatorsCount := *services.GetLatestStats().ActiveValidatorCount
	if activeValidatorsCount == 0 {
		return nil
	}

	probability := float64(validatorsCount) / float64(activeValidatorsCount)
	// in a geometric distribution, the expected value of the number of trials needed until first success is 1/p
	// you can think of this as the average interval of blocks until you get a proposal
	expectedValue := 1 / probability

	// return the timestamp of the last proposed block plus the average interval
	nextExpectedSlot := slots[len(slots)-1] + uint64(expectedValue)
	estimate := utils.SlotToTime(nextExpectedSlot)
	return &estimate
}

// getNextSyncEstimateTimestamp will return the estimated timestamp of the next sync committee
// given the maximum sync period the validators have peen part of and the number of validators
func getNextSyncEstimateTimestamp(maxPeriod uint64, validatorsCount int) *time.Time {
	// don't estimate if there are no validators or no sync committees
	if maxPeriod == 0 || validatorsCount == 0 {
		return nil
	}
	activeValidatorsCount := *services.GetLatestStats().ActiveValidatorCount
	if activeValidatorsCount == 0 {
		return nil
	}

	probability := (float64(utils.Config.Chain.Config.SyncCommitteeSize) / float64(activeValidatorsCount)) * float64(validatorsCount)
	// in a geometric distribution, the expected value of the number of trials needed until first success is 1/p
	// you can think of this as the average interval of sync committees until you expect to have been part of one
	expectedValue := 1 / probability

	// return the timestamp of the last sync committee plus the average interval
	nextExpectedSyncPeriod := maxPeriod + uint64(expectedValue)
	estimate := utils.EpochToTime(utils.FirstEpochOfSyncPeriod(nextExpectedSyncPeriod))
	return &estimate
}

// LatestState will return common information that about the current state of the eth2 chain
func LatestState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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
