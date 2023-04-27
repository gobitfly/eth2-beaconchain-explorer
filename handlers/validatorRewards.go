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
	validatorLimit := getUserPremium(r).MaxValidators
	validatorArr, err := parseValidatorsFromQueryString(q.Get("validators"), validatorLimit)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error parsing validators from query string")
		http.Error(w, "Invalid query", 400)
		return
	}

	currency := q.Get("currency")

	var start uint64 = 0
	var end uint64 = 0
	dateRange := strings.Split(q.Get("days"), "-")
	if len(dateRange) == 2 {
		start, err = strconv.ParseUint(dateRange[0], 10, 64)
		if err != nil {
			logger.Errorf("error retrieving days range %v", err)
			http.Error(w, "Invalid query", 400)
			return
		}
		end, err = strconv.ParseUint(dateRange[1], 10, 64)
		if err != nil {
			logger.Errorf("error retrieving days range %v", err)
			http.Error(w, "Invalid query", 400)
			return
		}
	}

	data := services.GetValidatorHist(validatorArr, currency, start, end)

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error encoding json response")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

}

func DownloadRewardsHistoricalData(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	validatorLimit := getUserPremium(r).MaxValidators
	validatorArr, err := parseValidatorsFromQueryString(q.Get("validators"), validatorLimit)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error parsing validators from query string")
		http.Error(w, "Invalid query", 400)
		return
	}

	currency := q.Get("currency")

	var start uint64 = 0
	var end uint64 = 0
	dateRange := strings.Split(q.Get("days"), "-")
	if len(dateRange) == 2 {
		start, err = strconv.ParseUint(dateRange[0], 10, 64)
		if err != nil {
			logger.Errorf("error retrieving days range %v", err)
			http.Error(w, "Invalid query", 400)
			return
		}
		end, err = strconv.ParseUint(dateRange[1], 10, 64)
		if err != nil {
			logger.Errorf("error retrieving days range %v", err)
			http.Error(w, "Invalid query", 400)
			return
		}
	}

	hist := services.GetValidatorHist(validatorArr, currency, start, end)

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
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

}

func RewardNotificationSubscribe(w http.ResponseWriter, r *http.Request) {
	SetAutoContentType(w, r)
	user := getUser(r)
	if !user.Authenticated {
		logger.WithField("route", r.URL.String()).Error("User not Authenticated")
		http.Error(w, "Internal server error, User Not Authenticated", http.StatusInternalServerError)
		return
	}

	var count uint64
	err := db.FrontendWriterDB.Get(&count,
		`select count(event_name) 
		from users_subscriptions 
		where user_id=$1 AND event_name=$2;`, user.UserID, strings.ToLower(utils.GetNetwork())+":"+string(types.TaxReportEventName))

	if err != nil || count >= 5 {
		logger.WithField("route", r.URL.String()).Info(fmt.Sprintf("User Subscription limit (%v) reached %v", count, err))
		http.Error(w, "Internal server error, User Subscription limit reached", http.StatusInternalServerError)
		return
	}

	q := r.URL.Query()

	validatorArr := q.Get("validators")
	validatorLimit := getUserPremium(r).MaxValidators
	_, err = parseValidatorsFromQueryString(validatorArr, validatorLimit)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error parsing validators from query string")
		http.Error(w, "Invalid query", 400)
		return
	}

	currency := q.Get("currency")

	if validatorArr == "" || !isValidCurrency(currency) {
		logger.WithField("route", r.URL.String()).Error("Bad Query")
		http.Error(w, "Internal server error, Bad Query", http.StatusInternalServerError)
		return
	}

	err = db.AddSubscription(user.UserID,
		utils.Config.Chain.Config.ConfigName,
		types.TaxReportEventName,
		fmt.Sprintf("validators=%s&days=30&currency=%s", validatorArr, currency), 0)

	if err != nil {
		logger.Errorf("error updating user subscriptions: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(struct {
		Msg string `json:"msg"`
	}{Msg: "Subscription Updated"})

	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error encoding json response")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

}

func RewardNotificationUnsubscribe(w http.ResponseWriter, r *http.Request) {
	SetAutoContentType(w, r)
	user := getUser(r)
	if !user.Authenticated {
		logger.WithField("route", r.URL.String()).Error("User not Authenticated")
		http.Error(w, "Internal server error, User Not Authenticated", http.StatusInternalServerError)
		return
	}

	q := r.URL.Query()

	validatorArr := q.Get("validators")

	currency := q.Get("currency")

	if validatorArr == "" || !isValidCurrency(currency) {
		logger.WithField("route", r.URL.String()).Error("Bad Query")
		http.Error(w, "Internal server error, Bad Query", http.StatusInternalServerError)
		return
	}

	err := db.DeleteSubscription(user.UserID,
		utils.GetNetwork(),
		types.TaxReportEventName,
		fmt.Sprintf("validators=%s&days=30&currency=%s", validatorArr, currency))

	if err != nil {
		logger.Errorf("error deleting entry from user subscriptions: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(struct {
		Msg string `json:"msg"`
	}{Msg: "Subscription Deleted"})

	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error encoding json response")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func RewardGetUserSubscriptions(w http.ResponseWriter, r *http.Request) {
	SetAutoContentType(w, r)
	user := getUser(r)
	if !user.Authenticated {
		logger.WithField("route", r.URL.String()).Error("User not Authenticated")
		http.Error(w, "Internal server error, User Not Authenticated", http.StatusInternalServerError)
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
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}
