package price

import (
	"context"
	"eth2-exporter/contracts/chainlink_feed"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var logger = logrus.New().WithField("module", "price")

var availableCurrencies = []string{}

var runOnce sync.Once
var runOnceWg sync.WaitGroup
var prices = map[string]float64{}
var pricesMu = &sync.Mutex{}
var didInit = uint64(0)
var feeds = map[string]*chainlink_feed.Feed{}
var calcPairs = map[string]bool{}
var clCurrency = "ETH"
var elCurrency = "ETH"

var currencies = map[string]struct {
	Symbol string
	Label  string
}{
	"AUD":  {"A$", "Australian Dollar"},
	"CAD":  {"C$", "Canadian Dollar"},
	"CNY":  {"¥", "Chinese Yuan"},
	"DAI":  {"DAI", "DAI stablecoin"},
	"xDAI": {"xDAI", "xDAI stablecoin"},
	"ETH":  {"ETH", "Ether"},
	"EUR":  {"€", "Euro"},
	"GBP":  {"£", "Pound Sterling"},
	"GNO":  {"GNO", "Gnosis"},
	"mGNO": {"mGNO", "mGnosis"},
	"JPY":  {"¥", "Japanese Yen"},
	"RUB":  {"₽", "Russian Ruble"},
	"USD":  {"$", "United States Dollar"},
}

func init() {
	runOnceWg.Add(1)
}

func Init(chainId uint64, eth1Endpoint, clCurrencyParam, elCurrencyParam string) {
	if atomic.AddUint64(&didInit, 1) > 1 {
		logrus.Warnf("price.Init called multiple times")
		return
	}

	switch chainId {
	case 1, 100:
	default:
		setPrice(elCurrency, elCurrency, 1)
		setPrice(clCurrency, clCurrency, 1)
		availableCurrencies = []string{clCurrency, elCurrency}
		logger.Warnf("chainId not supported for fetching prices: %v", chainId)
		runOnce.Do(func() { runOnceWg.Done() })
		return
	}

	clCurrency = clCurrencyParam
	elCurrency = elCurrencyParam
	if elCurrency == "xDAI" {
		elCurrency = "DAI"
	}
	calcPairs[elCurrency] = true
	calcPairs[clCurrency] = true

	eClient, err := ethclient.Dial(eth1Endpoint)
	if err != nil {
		logger.Errorf("error dialing pricing eth1 endpoint: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	clientChainId, err := eClient.ChainID(ctx)
	if err != nil {
		logger.WithError(err).Fatalf("failed getting chainID")
	}
	if chainId != clientChainId.Uint64() {
		logger.WithError(err).Fatalf("chainId does not match chainId from client (%v != %v)", chainId, clientChainId.Uint64())
	}

	feedAddrs := map[string]string{}
	switch chainId {
	case 1:
		// see: https://docs.chain.link/data-feeds/price-feeds/addresses/
		feedAddrs["ETH/USD"] = "0x5f4ec3df9cbd43714fe2740f5e3616155c5b8419"
		feedAddrs["EUR/USD"] = "0xb49f677943bc038e9857d61e7d053caa2c1734c1"
		feedAddrs["CAD/USD"] = "0xa34317db73e77d453b1b8d04550c44d10e981c8e"
		feedAddrs["CNY/USD"] = "0xef8a4af35cd47424672e3c590abd37fbb7a7759a"
		feedAddrs["JPY/USD"] = "0xbce206cae7f0ec07b545edde332a47c2f75bbeb3"
		feedAddrs["GBP/USD"] = "0x5c0ab2d9b5a7ed9f470386e82bb36a3613cdd4b5"
		feedAddrs["AUD/USD"] = "0x77f9710e7d0a19669a13c055f62cd80d313df022"

		availableCurrencies = []string{"ETH", "USD", "EUR", "GBP", "CNY", "CAD", "AUD", "JPY"}
	case 5:
		// see: https://docs.chain.link/data-feeds/price-feeds/addresses/
		feedAddrs["ETH/USD"] = "0x694AA1769357215DE4FAC081bf1f309aDC325306"
		feedAddrs["EUR/USD"] = "0x1a81afB8146aeFfCFc5E50e8479e826E7D55b910"

		availableCurrencies = []string{"ETH", "USD", "EUR"}
	case 11155111:
		// see: https://docs.chain.link/data-feeds/price-feeds/addresses/
		feedAddrs["ETH/USD"] = "0xD4a33860578De61DBAbDc8BFdb98FD742fA7028e"
		feedAddrs["EUR/USD"] = "0x44390589104C9164407A0E0562a9DBe6C24A0E05"

		availableCurrencies = []string{"ETH", "USD", "EUR"}
	case 100:
		// see: https://docs.chain.link/data-feeds/price-feeds/addresses/?network=gnosis-chain
		feedAddrs["GNO/USD"] = "0x22441d81416430A54336aB28765abd31a792Ad37"
		feedAddrs["DAI/USD"] = "0x678df3415fc31947dA4324eC63212874be5a82f8"
		feedAddrs["EUR/USD"] = "0xab70BCB260073d036d1660201e9d5405F5829b7a"
		feedAddrs["JPY/USD"] = "0x2AfB993C670C01e9dA1550c58e8039C1D8b8A317"
		// feedAddrs["CHFUSD"] = "0xFb00261Af80ADb1629D3869E377ae1EEC7bE659F"
		feedAddrs["ETH/USD"] = "0xa767f745331D267c7751297D982b050c93985627"

		setPrice("mGNO", "GNO", float64(1)/float64(32))
		setPrice("GNO", "mGNO", 32)
		setPrice("mGNO", "mGNO", float64(1)/float64(32))
		setPrice("GNO", "GNO", 1)

		calcPairs["GNO"] = true

		availableCurrencies = []string{"GNO", "mGNO", "DAI", "ETH", "USD", "EUR", "JPY"}
	default:
		logger.Fatalf("unsupported chainId %v", chainId)
	}

	for pair, addrHex := range feedAddrs {
		feed, err := chainlink_feed.NewFeed(common.HexToAddress(addrHex), eClient)
		if err != nil {
			logger.Errorf("failed to initialized chainlink feed for %v (addr: %v): %v", pair, addrHex, err)
			return
		}
		feeds[pair] = feed
	}

	go func() {
		for {
			updatePrices()
			time.Sleep(time.Minute)
		}
	}()
}

func updatePrices() {
	g := &errgroup.Group{}
	for pair, feed := range feeds {
		pair := pair
		feed := feed
		g.Go(func() error {
			price, err := getPriceFromFeed(feed)
			if err != nil {
				return fmt.Errorf("error getting price from feed for %v: %w", pair, err)
			}
			pricesMu.Lock()
			defer pricesMu.Unlock()
			prices[pair] = price
			if pair == "GNO/USD" {
				prices["mGNO/USD"] = price / 32
			}
			return nil
		})
	}
	err := g.Wait()
	if err != nil {
		logger.WithError(err).Errorf("error upating prices")
		return
	}
	for p := range calcPairs {
		if err = calcPricePairs(p); err != nil {
			logger.WithError(err).Errorf("error calculating price pairs for %v", p)
			return
		}
	}
	setPrice(elCurrency, elCurrency, 1)
	setPrice(clCurrency, clCurrency, 1)

	runOnce.Do(func() { runOnceWg.Done() })
}

func calcPricePairs(currency string) error {
	pricesMu.Lock()
	defer pricesMu.Unlock()
	pricesCopy := prices
	currencyUsdPrice, exists := prices[currency+"/USD"]
	if !exists {
		return fmt.Errorf("failed updating prices: cant find %v pair %+v", currency+"/USD", prices)
	}
	for pair, price := range pricesCopy {
		s := strings.Split(pair, "/")
		if len(s) < 2 || s[1] != "USD" {
			continue
		}
		// availableCurrencies = append(availableCurrencies, s[0])
		prices[currency+"/"+s[0]] = currencyUsdPrice / price
	}
	return nil
}

func setPrice(a, b string, v float64) {
	pricesMu.Lock()
	defer pricesMu.Unlock()
	prices[a+"/"+b] = v
}

func GetPrice(a, b string) float64 {
	if didInit < 1 {
		logger.Fatal("using GetPrice without calling price.Init once")
	}
	runOnceWg.Wait()
	pricesMu.Lock()
	defer pricesMu.Unlock()
	if a == "xDAI" {
		a = "DAI"
	}
	if b == "xDAI" {
		b = "DAI"
	}
	price, exists := prices[a+"/"+b]
	if !exists {
		logrus.WithFields(logrus.Fields{"pair": a + "/" + b}).Warnf("price pair not found")
		return 1
	}
	return price
}

func getPriceFromFeed(feed *chainlink_feed.Feed) (float64, error) {
	decimals := decimal.NewFromInt(1e8) // 8 decimal places for the Chainlink feeds
	res, err := feed.LatestRoundData(&bind.CallOpts{})
	if err != nil {
		return 0, fmt.Errorf("failed to fetch latest chainlink eth/usd price feed data: %w", err)
	}
	return decimal.NewFromBigInt(res.Answer, 0).Div(decimals).InexactFloat64(), nil
}

func GetAvailableCurrencies() []string {
	return availableCurrencies
}

func IsAvailableCurrency(currency string) bool {
	for _, c := range availableCurrencies {
		if c == currency {
			return true
		}
	}
	return false
}

func GetCurrencyLabel(currency string) string {
	x, exists := currencies[currency]
	if !exists {
		return ""
	}
	return x.Label
}

func GetCurrencySymbol(currency string) string {
	x, exists := currencies[currency]
	if !exists {
		return ""
	}
	return x.Symbol
}
