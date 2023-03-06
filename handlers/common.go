package handlers

import (
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

	lastDay := 0
	err := db.WriterDb.Get(&lastDay, "SELECT COALESCE(MAX(day), 0) FROM validator_stats")

	if err != nil {
		return nil, nil, err
	}
	lastWeek := lastDay - 7
	if lastWeek < 0 {
		lastWeek = 0
	}
	lastMonth := lastDay - 31
	if lastMonth < 0 {
		lastMonth = 0
	}

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

	type ClEarnings struct {
		EarningsTotal     int64 `db:"cl_rewards_gwei_total"`
		EarningsLastDay   int64 `db:"cl_rewards_gwei"`
		EarningsLastWeek  int64 `db:"cl_rewards_gwei_7d"`
		EarningsLastMonth int64 `db:"cl_rewards_gwei_31d"`
	}

	c := &ClEarnings{}

	err = db.ReaderDb.Get(c, `
	SELECT 
		COALESCE(SUM(cl_rewards_gwei), 0) AS cl_rewards_gwei, 
		COALESCE(SUM(cl_rewards_gwei_7d), 0) AS cl_rewards_gwei_7d, 
		COALESCE(SUM(cl_rewards_gwei_31d), 0) AS cl_rewards_gwei_31d, 
		COALESCE(SUM(cl_rewards_gwei_total), 0) AS cl_rewards_gwei_total
	FROM validator_stats WHERE day = $1 AND validatorindex = ANY($2)`, lastDay, validatorsPQArray)
	if err != nil {
		return nil, nil, err
	}

	var apr float64
	var totalDeposits int64

	err = db.ReaderDb.Get(&totalDeposits, `
	SELECT 
		COALESCE(SUM(amount), 0) 
	FROM blocks_deposits d
	INNER JOIN blocks b ON b.blockroot = d.block_root AND b.status = '1' 
	WHERE publickey IN (SELECT pubkey FROM validators WHERE validatorindex = ANY($1))`, validatorsPQArray)
	if err != nil {
		return nil, nil, err
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

	apr = (((float64(c.EarningsLastWeek) / 1e9) / (float64(32))) * 365) / 7
	if apr < float64(-1) {
		apr = float64(-1)
	}

	return &types.ValidatorEarnings{
		Total:                c.EarningsTotal,
		LastDay:              c.EarningsLastDay,
		LastWeek:             c.EarningsLastWeek,
		LastMonth:            c.EarningsLastMonth,
		APR:                  apr,
		TotalDeposits:        totalDeposits,
		TotalWithdrawals:     totalWithdrawals,
		LastDayFormatted:     utils.FormatIncome(c.EarningsLastDay, currency),
		LastWeekFormatted:    utils.FormatIncome(c.EarningsLastWeek, currency),
		LastMonthFormatted:   utils.FormatIncome(c.EarningsLastMonth, currency),
		TotalFormatted:       utils.FormatIncome(c.EarningsTotal, currency),
		TotalChangeFormatted: utils.FormatIncome(c.EarningsTotal+totalDeposits, currency),
	}, balancesMap, nil
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
		if err != nil {
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
