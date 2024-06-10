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
	for {
		updateHistoricPrices()
		time.Sleep(time.Hour)
	}
}

func WriteHistoricPricesForDay(ts time.Time) error {
	tsFormatted := ts.Format("01-02-2006")

	historicPrice, err := fetchHistoricPrice(ts)
	if err != nil {
		return fmt.Errorf("error retrieving historic eth prices for %v: %w", tsFormatted, err)
	}

	if historicPrice.MarketData.CurrentPrice.Eth == 0.0 ||
		historicPrice.MarketData.CurrentPrice.Eur == 0.0 ||
		historicPrice.MarketData.CurrentPrice.Usd == 0.0 ||
		historicPrice.MarketData.CurrentPrice.Rub == 0.0 ||
		historicPrice.MarketData.CurrentPrice.Cny == 0.0 ||
		historicPrice.MarketData.CurrentPrice.Cad == 0.0 ||
		historicPrice.MarketData.CurrentPrice.Jpy == 0.0 ||
		historicPrice.MarketData.CurrentPrice.Gbp == 0.0 ||
		historicPrice.MarketData.CurrentPrice.Aud == 0.0 {
		return fmt.Errorf("incomplete historic eth prices for %v", tsFormatted)
	}

	_, err = db.WriterDb.Exec(`
		INSERT INTO price (ts, eur, usd, rub, cny, cad, jpy, gbp, aud)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (ts) DO UPDATE SET
			eur = excluded.eur,
			usd = excluded.usd,
			rub = excluded.rub,
			cny = excluded.cny,
			cad = excluded.cad,
			jpy = excluded.jpy,
			gbp = excluded.gbp,
			aud = excluded.aud`,
		ts,
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
		return fmt.Errorf("error saving historic eth prices for %v: %w", tsFormatted, err)
	}
	return nil
}

func updateHistoricPrices() error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("service_historic_prices").Observe(time.Since(start).Seconds())
	}()
	var dates []time.Time

	err := db.WriterDb.Select(&dates, "SELECT ts FROM price")

	if err != nil {
		return err
	}

	datesMap := make(map[string]bool)
	for _, date := range dates {
		datesMap[date.Format("01-02-2006")] = true
	}

	currentDay := time.Unix(int64(utils.Config.Chain.GenesisTimestamp), 0)

	for currentDay.Before(time.Now()) {
		currentDayTrunc := currentDay.Truncate(utils.Day)
		if !datesMap[currentDayTrunc.Format("01-02-2006")] {
			err = WriteHistoricPricesForDay(currentDayTrunc)
			if err != nil {
				utils.LogError(err, "error writing historic price", 0)
			}

			// Wait to not overload the API
			time.Sleep(5 * time.Second)
		}
		currentDay = currentDay.Add(utils.Day)
	}
	return nil
}

func fetchHistoricPrice(ts time.Time) (*types.HistoricEthPrice, error) {
	logger.Infof("fetching historic prices for day %v", ts)
	client := &http.Client{Timeout: time.Second * 10}

	chain := "ethereum"

	if utils.Config.Chain.Name == "gnosis" {
		chain = "gnosis"
	}
	resp, err := client.Get(fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%s/history?date=%s", chain, ts.Truncate(utils.Day).Format("02-01-2006")))

	if err != nil {
		return nil, err
	}

	priceData := &types.HistoricEthPrice{}
	err = json.NewDecoder(resp.Body).Decode(&priceData)
	return priceData, err
}
