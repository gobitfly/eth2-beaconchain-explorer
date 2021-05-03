package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/lib/pq"
)

var validatorRewardsServicesTemplate = template.Must(template.New("validatorRewards").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/validatorRewards.html"))

type rewardsResp struct {
	Currencies []string
	CsrfField  template.HTML
}

func ValidatorRewards(w http.ResponseWriter, r *http.Request) {
	var err error

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "services", "/rewards", "Ethereum Validator rewards")

	var supportedCurrencies []string
	err = db.DB.Select(&supportedCurrencies,
		`select column_name 
			from information_schema.columns 
			where table_name = 'price'`)
	if err != nil {
		logger.Errorf("error getting eth1-deposits-distribution for stake pools: %w", err)
	}

	data.Data = rewardsResp{Currencies: supportedCurrencies, CsrfField: csrf.TemplateField(r)}

	err = validatorRewardsServicesTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
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
		// year, month, day := item.TS.Date()
		// date := fmt.Sprintf("%d-%d-%d", day, month, year)
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
		case "jyp":
			prices[date] = item.JPY
		case "rub":
			prices[date] = item.RUB
		default:
			prices[date] = item.USD
		}
	}

	totalIncomePerDay := map[string][2]int64{}
	for _, item := range income {
		// year, month, day := utils.DayToTime(item.Day).Date()
		// date := fmt.Sprintf("%d-%d-%d", day, month, year)
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

	// type resp struct {
	// 	Date         string  `json:"date"`
	// 	EndBalance   int64   `json:"end_balance"`
	// 	StartBalance int64   `json:"start_balance"`
	// 	Price        float64 `json:"price"`
	// 	Currency     string  `json:"currency"`
	// }

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

	return data
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
	if currency == "" {
		currency = "usd"
	}

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
