package price

import (
	"context"
	"eth2-exporter/price/chainlink_feed"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var logger = logrus.New().WithField("module", "price")

var availableCurrencies = []string{"ETH", "USD", "EUR", "GBP", "CNY", "CAD", "AUD", "JPY"}

var runOnce sync.Once
var runOnceWg sync.WaitGroup
var prices = map[string]float64{}
var pricesMu = &sync.Mutex{}
var feeds = map[string]*chainlink_feed.Feed{}
var clCurrency = "ETH"
var elCurrency = "ETH"

func init() {
	runOnceWg.Add(1)
}

func Init(chainId uint64, eth1Endpoint, clCurrencyParam, elCurrencyParam string) {
	switch strings.ToLower(clCurrencyParam) {
	case "eth", "gno":
		clCurrency = strings.ToUpper(clCurrencyParam)
	default:
		logger.Fatalf("invalid clCurrency: %v", clCurrencyParam)
	}
	switch strings.ToLower(elCurrencyParam) {
	case "xdai":
		elCurrency = "DAI"
	case "dai", "eth":
		elCurrency = strings.ToUpper(elCurrencyParam)
	default:
		logger.Fatalf("invalid elCurrency: %v", elCurrencyParam)
	}

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
		feedAddrs["ETHUSD"] = "0x5f4ec3df9cbd43714fe2740f5e3616155c5b8419"
		feedAddrs["EURUSD"] = "0xb49f677943bc038e9857d61e7d053caa2c1734c1"
		feedAddrs["CADUSD"] = "0xa34317db73e77d453b1b8d04550c44d10e981c8e"
		feedAddrs["CNYUSD"] = "0xef8a4af35cd47424672e3c590abd37fbb7a7759a"
		feedAddrs["JPYUSD"] = "0xbce206cae7f0ec07b545edde332a47c2f75bbeb3"
		feedAddrs["GBPUSD"] = "0x5c0ab2d9b5a7ed9f470386e82bb36a3613cdd4b5"
		feedAddrs["AUDUSD"] = "0x77f9710e7d0a19669a13c055f62cd80d313df022"
	case 100:
		// see: https://docs.chain.link/data-feeds/price-feeds/addresses/?network=gnosis-chain
		feedAddrs["GNOUSD"] = "0x22441d81416430A54336aB28765abd31a792Ad37"
		feedAddrs["DAIUSD"] = "0x678df3415fc31947dA4324eC63212874be5a82f8"
		feedAddrs["EURUSD"] = "0xab70BCB260073d036d1660201e9d5405F5829b7a"
		feedAddrs["JPYUSD"] = "0x2AfB993C670C01e9dA1550c58e8039C1D8b8A317"
		// feedAddrs["CHFUSD"] = "0xFb00261Af80ADb1629D3869E377ae1EEC7bE659F"
		feedAddrs["ETHUSD"] = "0xa767f745331D267c7751297D982b050c93985627"
	default:
		logger.Fatalf("unsupported chainId %v", chainId)
	}

	availableCurrencies = []string{"USD"}
	for pair := range feedAddrs {
		c, ok := strings.CutSuffix(pair, "USD")
		if ok {
			availableCurrencies = append(availableCurrencies, c)
		}
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
			return nil
		})
	}
	err := g.Wait()
	if err != nil {
		logger.WithError(err).Errorf("error upating prices")
		return
	}
	if err = calcPricePairs(clCurrency); err != nil {
		logger.WithError(err).Errorf("error calculating price pairs")
		return
	}
	if clCurrency != elCurrency {
		if err = calcPricePairs(elCurrency); err != nil {
			logger.WithError(err).Errorf("error calculating price pairs")
			return
		}
	}
	prices[elCurrency+elCurrency] = 1
	prices[clCurrency+clCurrency] = 1
	runOnce.Do(func() { runOnceWg.Done() })
}

func calcPricePairs(currency string) error {
	pricesMu.Lock()
	defer pricesMu.Unlock()
	currency = SanitizeCurrency(currency)
	pricesCopy := prices
	currencyUsdPrice, exists := prices[currency+"USD"]
	if !exists {
		return fmt.Errorf("failed updating prices: cant find %v pair %+v", currency+"USD", prices)
	}
	for pair, price := range pricesCopy {
		prefix, found := strings.CutSuffix(pair, "USD")
		if !found || prefix == currency {
			continue
		}
		prices[currency+prefix] = currencyUsdPrice / price
	}
	return nil
}

func SanitizeCurrency(c string) string {
	if strings.ToLower(c) == "xdai" {
		c = "DAI"
	}
	return strings.ToUpper(c)
}

func GetPrice(a, b string) float64 {
	a = SanitizeCurrency(a)
	b = SanitizeCurrency(b)
	pricesMu.Lock()
	defer pricesMu.Unlock()
	price, exists := prices[a+b]
	if !exists {
		return 0
	}
	return price
}

func getPriceFromFeed(feed *chainlink_feed.Feed) (float64, error) {
	decimals, _ := new(big.Float).SetString("100000000") // 8 decimal places for the Chainlink feeds
	res, err := feed.LatestRoundData(nil)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch latest chainlink eth/usd price feed data: %v", err)
	}
	priceRaw := new(big.Float).SetInt(res.Answer)
	priceRaw.Quo(priceRaw, decimals)
	price, _ := priceRaw.Float64()
	return price, nil
}

func GetAvailableCurrencies() []string {
	return availableCurrencies
}

func GetCurrencyLabel(currency string) string {
	switch currency {
	case "ETH":
		return "Ether"
	case "GNO":
		return "Gnosis"
	case "USD":
		return "United States Dollar"
	case "EUR":
		return "Euro"
	case "GBP":
		return "Pound Sterling"
	case "CNY":
		return "Chinese Yuan"
	case "RUB":
		return "Russian Ruble"
	case "CAD":
		return "Canadian Dollar"
	case "AUD":
		return "Australian Dollar"
	case "JPY":
		return "Japanese Yen"
	case "DAI":
		return "DAI stablecoin"
	default:
		return ""
	}
}

func GetSymbol(currency string) string {
	switch currency {
	case "EUR":
		return "€"
	case "USD":
		return "$"
	case "RUB":
		return "₽"
	case "CNY":
		return "¥"
	case "CAD":
		return "C$"
	case "AUD":
		return "A$"
	case "JPY":
		return "¥"
	case "GBP":
		return "£"
	case "DAI":
		return "D"
	default:
		return ""
	}
}

func GetEthRoundPrice(currency float64) uint64 {
	ethRoundPrice := uint64(currency)
	return ethRoundPrice
}
