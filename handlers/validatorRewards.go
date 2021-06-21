package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
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

var validatorRewardsServicesTemplate = template.Must(template.New("validatorRewards").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/validatorRewards.html"))

// var supportedCurrencies = []string{"eur", "usd", "gbp", "cny", "cad", "jpy", "rub"}

type rewardsResp struct {
	Currencies       []string
	CsrfField        template.HTML
	Subscriptions    [][]string
	MinDateTimestamp uint64
}

func ValidatorRewards(w http.ResponseWriter, r *http.Request) {
	var err error

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "services", "/rewards", "Ethereum Validator Rewards")

	var supportedCurrencies []string
	err = db.DB.Select(&supportedCurrencies,
		`select column_name 
			from information_schema.columns 
			where table_name = 'price'`)
	if err != nil {
		logger.Errorf("error getting eth1-deposits-distribution for stake pools: %w", err)
	}

	var minTime time.Time
	err = db.DB.Get(&minTime,
		`select ts from price order by ts asc limit 1`)
	if err != nil {
		logger.Errorf("error getting min ts: %w", err)
	}

	var subs = [][]string{}

	if data.User.Authenticated {
		subs = getUserRewardSubscriptions(data.User.UserID)
	}

	data.Data = rewardsResp{Currencies: supportedCurrencies, CsrfField: csrf.TemplateField(r), Subscriptions: subs, MinDateTimestamp: uint64(minTime.Unix())}

	err = validatorRewardsServicesTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

func getUserRewardSubscriptions(uid uint64) [][]string {
	var dbResp []types.Subscription
	err := db.DB.Select(&dbResp,
		`select * from users_subscriptions where event_name=$1 AND user_id=$2`, types.TaxReportEventName, uid)
	if err != nil {
		logger.Errorf("error getting prices: %w", err)
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
	err := db.DB.Get(&count,
		`select count(column_name) 
		from information_schema.columns 
		where table_name = 'price' AND column_name=$1;`, currency)
	if err != nil {
		logger.Errorf("error checking currency: %w", err)
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

	validatorArr, err := parseValidatorsFromQueryString(q.Get("validators"))
	if err != nil {
		logger.Errorf("error retrieving active validators %v", err)
		http.Error(w, "Invalid query", 400)
		return
	}

	currency := q.Get("currency")

	var start uint64 = 0
	var end uint64 = 0
	dateRange := strings.Split(q.Get("days"), "-")
	if len(dateRange) == 2 {
		start, err = strconv.ParseUint(dateRange[0], 10, 64)
		end, err = strconv.ParseUint(dateRange[1], 10, 64)
		if err != nil {
			logger.Errorf("error retrieving days range %v", err)
			http.Error(w, "Invalid query", 400)
			return
		}
	}

	// days, err := strconv.ParseUint(q.Get("days"), 10, 64)

	data := services.GetValidatorHist(validatorArr, currency, start, end)

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error encoding json response")
		http.Error(w, "Internal server error", 503)
		return
	}

}

func DownloadRewardsHistoricalData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Disposition", "attachment; filename=beaconcha_in-rewards-history.pdf")
	w.Header().Set("Content-Type", "text/csv")

	q := r.URL.Query()

	validatorArr, err := parseValidatorsFromQueryString(q.Get("validators"))
	if err != nil {
		logger.Errorf("error retrieving active validators %v", err)
		http.Error(w, "Invalid query", 400)
		return
	}

	currency := q.Get("currency")

	var start uint64 = 0
	var end uint64 = 0
	dateRange := strings.Split(q.Get("days"), "-")
	if len(dateRange) == 2 {
		start, err = strconv.ParseUint(dateRange[0], 10, 64)
		end, err = strconv.ParseUint(dateRange[1], 10, 64)
		if err != nil {
			logger.Errorf("error retrieving days range %v", err)
			http.Error(w, "Invalid query", 400)
			return
		}
	}

	hist := services.GetValidatorHist(validatorArr, currency, start, end)
	// data := hist.History

	if len(hist.History) == 0 {
		w.Write([]byte("No data available"))
		return
	}

	// cur := data[0][len(data[0])-1]
	// cur = strings.ToUpper(cur)
	// csv := fmt.Sprintf("Date,End-of-date balance ETH,Income for date ETH,Price of ETH for date %s, Income for date %s", cur, cur)

	// totalIncomeEth := 0.0
	// totalIncomeCur := 0.0

	// for _, item := range data {
	// 	if len(item) < 5 {
	// 		csv += "\n0,0,0,0,0"
	// 		continue
	// 	}
	// 	csv += fmt.Sprintf("\n%s,%s,%s,%s,%s", item[0], item[1], item[2], item[3], item[4])
	// 	tEth, err := strconv.ParseFloat(item[2], 64)
	// 	tCur, err := strconv.ParseFloat(item[4], 64)

	// 	if err != nil {
	// 		continue
	// 	}

	// 	totalIncomeEth += tEth
	// 	totalIncomeCur += tCur
	// }

	// csv += fmt.Sprintf("\nTotal, ,%f, ,%f", totalIncomeEth, totalIncomeCur)

	_, err = w.Write(services.GeneratePdfReport(hist))
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error writing response")
		http.Error(w, "Internal server error", 503)
		return
	}

}

func RewardNotificationSubscribe(w http.ResponseWriter, r *http.Request) {
	SetAutoContentType(w, r)
	user := getUser(w, r)
	if !user.Authenticated {
		logger.WithField("route", r.URL.String()).Error("User not Authenticated")
		http.Error(w, "Internal server error, User Not Authenticated", http.StatusInternalServerError)
		return
	}

	q := r.URL.Query()

	validatorArr := q.Get("validators")
	_, err := parseValidatorsFromQueryString(validatorArr)
	if err != nil {
		http.Error(w, "Invalid query, Invalid Validators", 400)
		return
	}

	currency := q.Get("currency")

	if validatorArr == "" || !isValidCurrency(currency) {
		logger.WithField("route", r.URL.String()).Error("Bad Query")
		http.Error(w, "Internal server error, Bad Query", http.StatusInternalServerError)
		return
	}

	err = db.AddSubscription(user.UserID,
		types.TaxReportEventName,
		fmt.Sprintf("validators=%s&days=30&currency=%s", validatorArr, currency))

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
		http.Error(w, "Internal server error", 503)
		return
	}

}

func RewardNotificationUnsubscribe(w http.ResponseWriter, r *http.Request) {
	SetAutoContentType(w, r)
	user := getUser(w, r)
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
		http.Error(w, "Internal server error", 503)
		return
	}
}
