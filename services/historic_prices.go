package services

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/metrics"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"net/http"
	"time"
)

func StartHistoricPriceService() {
	for true {
		updateHistoricPrices()
		time.Sleep(time.Hour)
	}
}
func updateHistoricPrices() error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("service_historic_prices").Observe(time.Since(start).Seconds())
	}()
	var dates []time.Time

	err := db.DB.Select(&dates, "SELECT ts FROM price")

	if err != nil {
		return err
	}

	datesMap := make(map[string]bool)
	for _, date := range dates {
		datesMap[date.Format("01-02-2006")] = true
	}

	currentDay := time.Unix(int64(utils.Config.Chain.GenesisTimestamp), 0)

	for currentDay.Before(time.Now()) {
		currentDayTrunc := currentDay.Truncate(time.Hour * 24)
		if !datesMap[currentDayTrunc.Format("01-02-2006")] {
			historicPrice, err := fetchHistoricPrice(currentDayTrunc)

			if err != nil {
				logger.Errorf("error retrieving historic eth prices for day %v: %v", currentDayTrunc, err)
				currentDay = currentDay.Add(time.Hour * 24)
				continue
			}
			_, err = db.DB.Exec("INSERT INTO price (ts, eur, usd, rub, cny, cad, jpy, gbp, aud) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
				currentDayTrunc,
				historicPrice.MarketData.CurrentPrice.Eur,
				historicPrice.MarketData.CurrentPrice.Usd,
				historicPrice.MarketData.CurrentPrice.Rub,
				historicPrice.MarketData.CurrentPrice.Cny,
				historicPrice.MarketData.CurrentPrice.Cad,
				historicPrice.MarketData.CurrentPrice.Jpy,
				historicPrice.MarketData.CurrentPrice.Gbp,
				historicPrice.MarketData.CurrentPrice.Aud,
			)
			if err != nil {
				logger.Errorf("error saving historic eth prices for day %v: %v", currentDayTrunc, err)
				currentDay = currentDay.Add(time.Hour * 24)
				continue
			}
		}
		currentDay = currentDay.Add(time.Hour * 24)
	}
	return nil
}

func fetchHistoricPrice(ts time.Time) (*types.HistoricEthPrice, error) {
	logger.Infof("fetching historic prices for day %v", ts)
	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Get(fmt.Sprintf("https://api.coingecko.com/api/v3/coins/ethereum/history?date=%s", ts.Truncate(time.Hour*24).Format("02-01-2006")))

	if err != nil {
		return nil, err
	}

	priceData := &types.HistoricEthPrice{}
	err = json.NewDecoder(resp.Body).Decode(&priceData)
	return priceData, err
}
