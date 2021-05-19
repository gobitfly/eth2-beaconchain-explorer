package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/csrf"
	"github.com/lib/pq"
)

var validatorRewardsServicesTemplate = template.Must(template.New("validatorRewards").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/validatorRewards.html"))

// var supportedCurrencies = []string{"eur", "usd", "gbp", "cny", "cad", "jpy", "rub"}

type rewardsResp struct {
	Currencies    []string
	CsrfField     template.HTML
	Subscriptions [][]string
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

	var subs = [][]string{}

	if data.User.Authenticated {
		subs = getUserRewardSubscriptions(data.User.UserID)
	}

	data.Data = rewardsResp{Currencies: supportedCurrencies, CsrfField: csrf.TemplateField(r), Subscriptions: subs}

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

func getValidatorHist(validatorArr []uint64, currency string, days uint64) [][]string {
	var err error
	validatorFilter := pq.Array(validatorArr)

	var pricesDb []types.Price
	err = db.DB.Select(&pricesDb,
		`select * from price order by ts desc limit $1`, days)
	if err != nil {
		logger.Errorf("error getting prices: %w", err)
	}

	var maxDay uint64
	err = db.DB.Get(&maxDay,
		`select MAX(day) from validator_stats`)
	if err != nil {
		logger.Errorf("error getting max day: %w", err)
	}

	if days > 365 {
		days = 365
	}
	var lowerBound uint64 = 0
	if maxDay > days {
		lowerBound = maxDay - days
	}
	var income []types.ValidatorStatsTableRow
	err = db.DB.Select(&income,
		`select day, start_balance, end_balance
		 from validator_stats 
		 where validatorindex=ANY($1) AND day > $2
		 order by day desc`, validatorFilter, lowerBound)
	if err != nil {
		logger.Errorf("error getting incomes: %w", err)
	}

	prices := map[string]float64{}
	for _, item := range pricesDb {
		date := fmt.Sprintf("%v", item.TS)
		date = strings.Split(date, " ")[0]
		switch currency {
		case "eur":
			prices[date] = item.EUR
		case "usd":
			prices[date] = item.USD
		case "gbp":
			prices[date] = item.GBP
		case "cad":
			prices[date] = item.CAD
		case "cny":
			prices[date] = item.CNY
		case "jpy":
			prices[date] = item.JPY
		case "rub":
			prices[date] = item.RUB
		default:
			prices[date] = item.USD
			currency = "usd"
		}
	}

	totalIncomePerDay := map[string][2]int64{}
	for _, item := range income {
		date := fmt.Sprintf("%v", utils.DayToTime(item.Day))
		date = strings.Split(date, " ")[0]
		if _, exist := totalIncomePerDay[date]; !exist {
			totalIncomePerDay[date] = [2]int64{item.StartBalance.Int64, item.EndBalance.Int64}
			continue
		}
		state := totalIncomePerDay[date]
		state[0] += item.StartBalance.Int64
		state[1] += item.EndBalance.Int64
		totalIncomePerDay[date] = state
	}

	data := make([][]string, len(totalIncomePerDay))
	i := 0
	for key, item := range totalIncomePerDay {
		if len(item) < 2 {
			continue
		}
		data[i] = []string{
			key,
			fmt.Sprintf("%f", float64(item[1])/1e9), // end of day balance
			fmt.Sprintf("%f", (float64(item[1])/1e9)-(float64(item[0])/1e9)),               // income of day ETH
			fmt.Sprintf("%f", prices[key]),                                                 //price will default to 0 if key does not exist
			fmt.Sprintf("%f", ((float64(item[1])/1e9)-(float64(item[0])/1e9))*prices[key]), // income of day Currency
			currency,
		}
		i++
	}

	sort.Slice(data, func(p, q int) bool {
		i, err := time.Parse("2006-01-02", data[p][0])
		i2, err := time.Parse("2006-01-02", data[q][0])
		if err != nil {
			return false
		}
		return i2.Before(i)
	})

	return data
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

	days, err := strconv.ParseUint(q.Get("days"), 10, 64)
	if err != nil {
		logger.Errorf("error retrieving days %v", err)
		http.Error(w, "Invalid query", 400)
		return
	}

	data := getValidatorHist(validatorArr, currency, days)

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error encoding json response")
		http.Error(w, "Internal server error", 503)
		return
	}

}

func DownloadRewardsHistoricalData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Disposition", "attachment; filename=beaconcah_in-income-report.csv")
	w.Header().Set("Content-Type", "text/csv")

	q := r.URL.Query()

	validatorArr, err := parseValidatorsFromQueryString(q.Get("validators"))
	if err != nil {
		logger.Errorf("error retrieving active validators %v", err)
		http.Error(w, "Invalid query", 400)
		return
	}

	currency := q.Get("currency")

	days, err := strconv.ParseUint(q.Get("days"), 10, 64)
	if err != nil {
		logger.Errorf("error retrieving days %v", err)
		http.Error(w, "Invalid query", 400)
		return
	}

	data := getValidatorHist(validatorArr, currency, days)

	if len(data) == 0 {
		w.Write([]byte("No data available"))
		return
	}
	cur := data[0][len(data[0])-1]
	cur = strings.ToUpper(cur)
	csv := fmt.Sprintf("Date,End-of-date balance ETH,Income for date ETH,Price of ETH for date %s, Income for date %s", cur, cur)

	totalIncomeEth := 0.0
	totalIncomeCur := 0.0

	for _, item := range data {
		if len(item) < 5 {
			csv += "\n0,0,0,0,0"
			continue
		}
		csv += fmt.Sprintf("\n%s,%s,%s,%s,%s", item[0], item[1], item[2], item[3], item[4])
		tEth, err := strconv.ParseFloat(item[2], 64)
		tCur, err := strconv.ParseFloat(item[4], 64)

		if err != nil {
			continue
		}

		totalIncomeEth += tEth
		totalIncomeCur += tCur
	}

	csv += fmt.Sprintf("\nTotal, ,%f, ,%f", totalIncomeEth, totalIncomeCur)

	_, err = w.Write([]byte(csv))
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
