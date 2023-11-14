package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/csrf"
)

// var supportedCurrencies = []string{"eur", "usd", "gbp", "cny", "cad", "jpy", "rub", "aud"}
const USER_SUBSCRIPTION_LIMIT = 5

type rewardsResp struct {
	Currencies        []string
	CsrfField         template.HTML
	ShowSubscriptions bool
	MinDateTimestamp  uint64
}

func ValidatorRewards(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "validatorRewards.html")
	var validatorRewardsServicesTemplate = templates.GetTemplate(templateFiles...)

	var err error

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "services", "/rewards", "Ethereum Validator Rewards", templateFiles)

	var supportedCurrencies []string
	err = db.ReaderDb.Select(&supportedCurrencies,
		`select column_name 
			from information_schema.columns 
			where table_name = 'price'`)
	if err != nil {
		logger.Errorf("error getting eth1-deposits-distribution for stake pools: %v", err)
	}

	var minTime time.Time
	err = db.ReaderDb.Get(&minTime,
		`select ts from price order by ts asc limit 1`)
	if err != nil {
		logger.Errorf("error getting min ts: %v", err)
	}

	data.Data = rewardsResp{Currencies: supportedCurrencies, CsrfField: csrf.TemplateField(r), MinDateTimestamp: uint64(minTime.Unix()), ShowSubscriptions: data.User.Authenticated}

	if handleTemplateError(w, r, "validatorRewards.go", "ValidatorRewards", "", validatorRewardsServicesTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

func getUserRewardSubscriptions(uid uint64) [][]string {
	var dbResp []types.Subscription

	err := db.FrontendWriterDB.Select(&dbResp,
		`select id, user_id, event_name, event_filter, last_sent_ts, last_sent_epoch, created_ts, created_epoch, event_threshold, unsubscribe_hash, internal_state from users_subscriptions where event_name=$1 AND user_id=$2`, strings.ToLower(utils.GetNetwork())+":"+string(types.TaxReportEventName), uid)
	if err != nil {
		logger.Errorf("error getting prices: %v", err)
	}

	res := make([][]string, len(dbResp))
	for i, item := range dbResp {
		q, err := url.ParseQuery(item.EventFilter)
		if err != nil {
			continue
		}
		res[i] = []string{
			fmt.Sprintf("%v", item.CreatedTime),
			q.Get("currency"),
			q.Get("validators"),
			item.EventFilter,
		}
	}

	return res
}

func isValidCurrency(currency string) bool {
	var count uint64
	err := db.ReaderDb.Get(&count,
		`select count(column_name) 
		from information_schema.columns 
		where table_name = 'price' AND column_name=$1;`, currency)
	if err != nil {
		logger.Errorf("error checking currency: %v", err)
		return false
	}

	if count > 0 {
		return true
	}

	return false
}

func RewardsHistoricalData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	validatorIndexArr, _, redirect, err := handleValidatorsQuery(w, r, true)
	if err != nil || redirect {
		return
	}

	currency := q.Get("currency")

	// Set the default start and end time to the first day
	t := time.Unix(int64(utils.Config.Chain.GenesisTimestamp), 0)
	startGenesisDay := uint64(time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC).Unix())
	var start uint64 = startGenesisDay
	var end uint64 = startGenesisDay

	dateRange := strings.Split(q.Get("days"), "-")
	if len(dateRange) == 2 {
		start, err = strconv.ParseUint(dateRange[0], 10, 32) //Limit to uint32 for postgres
		if err != nil || start < startGenesisDay {
			logger.Warnf("error parsing days range: %v", err)
			http.Error(w, "Error: Invalid parameter days.", http.StatusBadRequest)
			return
		}
		end, err = strconv.ParseUint(dateRange[1], 10, 32) //Limit to uint32 for postgres
		if err != nil || end < startGenesisDay {
			logger.Warnf("error parsing days range: %v", err)
			http.Error(w, "Error: Invalid parameter days.", http.StatusBadRequest)
			return
		}
	}

	data := services.GetValidatorHist(validatorIndexArr, currency, start, end)

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error encoding json response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

}

func DownloadRewardsHistoricalData(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	validatorIndexArr, _, redirect, err := handleValidatorsQuery(w, r, true)
	if err != nil || redirect {
		return
	}

	currency := q.Get("currency")

	// Set the default start and end time to the first day
	t := time.Unix(int64(utils.Config.Chain.GenesisTimestamp), 0)
	startGenesisDay := uint64(time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC).Unix())
	var start uint64 = startGenesisDay
	var end uint64 = startGenesisDay

	dateRange := strings.Split(q.Get("days"), "-")
	if len(dateRange) == 2 {
		start, err = strconv.ParseUint(dateRange[0], 10, 32) //Limit to uint32 for postgres
		if err != nil || start < startGenesisDay {
			logger.Warnf("error parsing days range: %v", err)
			http.Error(w, "Error: Invalid parameter days.", http.StatusBadRequest)
			return
		}
		end, err = strconv.ParseUint(dateRange[1], 10, 32) //Limit to uint32 for postgres
		if err != nil || end < startGenesisDay {
			logger.Warnf("error parsing days range: %v", err)
			http.Error(w, "Error: Invalid parameter days.", http.StatusBadRequest)
			return
		}
	}

	hist := services.GetValidatorHist(validatorIndexArr, currency, start, end)

	if len(hist.History) == 0 {
		w.Write([]byte("No data available"))
		return
	}

	s := time.Unix(int64(start), 0)
	e := time.Unix(int64(end), 0)

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=income_history_%v_%v.pdf", s.Format("20060102"), e.Format("20060102")))
	w.Header().Set("Content-Type", "text/csv")

	_, err = w.Write(services.GeneratePdfReport(hist, currency))
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error writing response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

}

func RewardNotificationSubscribe(w http.ResponseWriter, r *http.Request) {
	SetAutoContentType(w, r)
	user := getUser(r)
	if !user.Authenticated {
		http.Error(w, "User Not Authenticated", http.StatusUnauthorized)
		return
	}
	q := r.URL.Query()

	validatorArr := q.Get("validators")
	currency := q.Get("currency")
	validatorLimit := getUserPremium(r).MaxValidators

	errFields := map[string]interface{}{
		"route":            r.URL.String(),
		"user_id":          user.UserID,
		"validators_query": validatorArr,
		"currency":         currency,
		"validator_limit":  validatorLimit,
	}

	var count uint64
	err := db.FrontendWriterDB.Get(&count,
		`select count(event_name) 
		from users_subscriptions 
		where user_id=$1 AND event_name=$2;`, user.UserID, strings.ToLower(utils.GetNetwork())+":"+string(types.TaxReportEventName))
	if err != nil {
		utils.LogError(err, "Failed to get User Subscriptions Count", 0, errFields)
		http.Error(w, "Internal Server Error: failed to get user subscriptions count", http.StatusInternalServerError)
		return
	}

	if count >= USER_SUBSCRIPTION_LIMIT {
		http.Error(w, "Conflicting Request: user subscription limit reached", http.StatusConflict)
		return
	}

	// don't allow passing validator pubkeys in the query string
	_, queryValidatorPubkeys, err := parseValidatorsFromQueryString(validatorArr, validatorLimit)
	if err != nil || len(queryValidatorPubkeys) > 0 {
		http.Error(w, "Bad Request: validators could not be parsed or should be specified using Indices", http.StatusBadRequest)
		return
	}

	if validatorArr == "" || !isValidCurrency(currency) {
		http.Error(w, "Bad Request: no validators or invalid currency given", http.StatusBadRequest)
		return
	}

	err = db.AddSubscription(user.UserID,
		utils.Config.Chain.ClConfig.ConfigName,
		types.TaxReportEventName,
		fmt.Sprintf("validators=%s&days=30&currency=%s", validatorArr, currency), 0)

	if err != nil {
		utils.LogError(err, "Failed to add entry to user subscriptions", 0, errFields)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(struct {
		Msg string `json:"msg"`
	}{Msg: "Subscription Updated"})

	if err != nil {
		utils.LogError(err, "error encoding json response", 0, errFields)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

}

func RewardNotificationUnsubscribe(w http.ResponseWriter, r *http.Request) {
	SetAutoContentType(w, r)
	user := getUser(r)
	if !user.Authenticated {
		http.Error(w, "User Not Authenticated", http.StatusUnauthorized)
		return
	}

	q := r.URL.Query()

	validatorArr := q.Get("validators")
	currency := q.Get("currency")
	validatorLimit := getUserPremium(r).MaxValidators

	errFields := map[string]interface{}{
		"route":            r.URL.String(),
		"user_id":          user.UserID,
		"validators_query": validatorArr,
		"currency":         currency,
		"validator_limit":  validatorLimit,
	}

	// don't allow passing validator pubkeys in the query string
	_, queryValidatorPubkeys, err := parseValidatorsFromQueryString(validatorArr, validatorLimit)
	if err != nil || len(queryValidatorPubkeys) > 0 {
		http.Error(w, "Bad Request: validators could not be parsed or should be specified using Indices", http.StatusBadRequest)
		return
	}

	if validatorArr == "" || !isValidCurrency(currency) {
		http.Error(w, "Bad Request: no validators or invalid currency given", http.StatusBadRequest)
		return
	}

	err = db.DeleteSubscription(user.UserID,
		utils.GetNetwork(),
		types.TaxReportEventName,
		fmt.Sprintf("validators=%s&days=30&currency=%s", validatorArr, currency))

	if err != nil {
		utils.LogError(err, "Failed to delete entry from user subscriptions", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(struct {
		Msg string `json:"msg"`
	}{Msg: "Subscription Deleted"})

	if err != nil {
		utils.LogError(err, "error encoding json response", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func RewardGetUserSubscriptions(w http.ResponseWriter, r *http.Request) {
	SetAutoContentType(w, r)
	user := getUser(r)
	if !user.Authenticated {
		logger.WithField("route", r.URL.String()).Error("User not Authenticated")
		http.Error(w, "Internal server error, User Not Authenticated", http.StatusUnauthorized)
		return
	}

	var count uint64
	err := db.FrontendWriterDB.Get(&count,
		`select count(event_name) 
		from users_subscriptions 
		where user_id=$1 AND event_name=$2;`, user.UserID, strings.ToLower(utils.GetNetwork())+":"+string(types.TaxReportEventName))

	if err != nil {
		logger.WithField("route", r.URL.String()).Error("Failed to get User Subscriptions Count")
		http.Error(w, "Internal server error, Failed to get User Subscriptions Count", http.StatusInternalServerError)
		return
	}

	data := getUserRewardSubscriptions(user.UserID)

	err = json.NewEncoder(w).Encode(struct {
		Data  [][]string `json:"data"`
		Count uint64     `json:"count"`
	}{Data: data, Count: count})

	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error encoding json response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
