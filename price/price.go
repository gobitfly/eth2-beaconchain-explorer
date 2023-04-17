package price

import (
	"eth2-exporter/price/chainlink_feed"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var logger = logrus.New().WithField("module", "price")

type EthPrice struct {
	Ethereum struct {
		Cad float64 `json:"cad"`
		Cny float64 `json:"cny"`
		Eur float64 `json:"eur"`
		Jpy float64 `json:"jpy"`
		Usd float64 `json:"usd"`
		Gbp float64 `json:"gbp"`
		Aud float64 `json:"aud"`
	} `json:"ethereum"`
}

var availableCurrencies = []string{"ETH", "USD", "EUR", "GBP", "CNY", "CAD", "AUD", "JPY"}
var ethPrice = new(EthPrice)
var ethPriceMux = &sync.RWMutex{}

var ethUSDFeed *chainlink_feed.Feed
var eurUSDFeed *chainlink_feed.Feed
var cadUSDFeed *chainlink_feed.Feed
var cnyUSDFeed *chainlink_feed.Feed
var jpyUSDFeed *chainlink_feed.Feed
var gbpUSDFeed *chainlink_feed.Feed
var audUSDFeed *chainlink_feed.Feed

func Init(chainId uint64, eth1Endpoint string) {

	eClient, err := ethclient.Dial(eth1Endpoint)
	if err != nil {
		logger.Errorf("error dialing pricing eth1 endpoint: %v", err)
		return
	}

	ethUSDFeed, err = chainlink_feed.NewFeed(common.HexToAddress("0x5f4ec3df9cbd43714fe2740f5e3616155c5b8419"), eClient)
	if err != nil {
		logger.Errorf("failed to initialized chainlink eth/usd feed contract: %v", err)
		return
	}

	eurUSDFeed, err = chainlink_feed.NewFeed(common.HexToAddress("0xb49f677943bc038e9857d61e7d053caa2c1734c1"), eClient)
	if err != nil {
		logger.Errorf("failed to initialized chainlink eur/usd feed contract: %v", err)
		return
	}

	cadUSDFeed, err = chainlink_feed.NewFeed(common.HexToAddress("0xa34317db73e77d453b1b8d04550c44d10e981c8e"), eClient)
	if err != nil {
		logger.Errorf("failed to initialized chainlink eur/usd feed contract: %v", err)
		return
	}

	cnyUSDFeed, err = chainlink_feed.NewFeed(common.HexToAddress("0xef8a4af35cd47424672e3c590abd37fbb7a7759a"), eClient)
	if err != nil {
		logger.Errorf("failed to initialized chainlink eur/usd feed contract: %v", err)
		return
	}

	jpyUSDFeed, err = chainlink_feed.NewFeed(common.HexToAddress("0xbce206cae7f0ec07b545edde332a47c2f75bbeb3"), eClient)
	if err != nil {
		logger.Errorf("failed to initialized chainlink eur/usd feed contract: %v", err)
		return
	}

	gbpUSDFeed, err = chainlink_feed.NewFeed(common.HexToAddress("0x5c0ab2d9b5a7ed9f470386e82bb36a3613cdd4b5"), eClient)
	if err != nil {
		logger.Errorf("failed to initialized chainlink eur/usd feed contract: %v", err)
		return
	}

	audUSDFeed, err = chainlink_feed.NewFeed(common.HexToAddress("0x77f9710e7d0a19669a13c055f62cd80d313df022"), eClient)
	if err != nil {
		logger.Errorf("failed to initialized chainlink eur/usd feed contract: %v", err)
		return
	}

	go updateEthPrice(chainId, eth1Endpoint)
}

func updateEthPrice(chainId uint64, eth1Endpoint string) {

	for {
		fetchChainlinkFeed(chainId)
		time.Sleep(time.Minute)
	}
}

func fetchChainlinkFeed(chainId uint64) {
	if chainId != 1 {
		ethPrice = &EthPrice{
			Ethereum: struct {
				Cad float64 "json:\"cad\""
				Cny float64 "json:\"cny\""
				Eur float64 "json:\"eur\""
				Jpy float64 "json:\"jpy\""
				Usd float64 "json:\"usd\""
				Gbp float64 "json:\"gbp\""
				Aud float64 "json:\"aud\""
			}{
				Cad: 0,
				Cny: 0,
				Eur: 0,
				Jpy: 0,
				Usd: 0,
				Gbp: 0,
				Aud: 0,
			},
		}
		return
	}

	var ethUSDPrice float64
	var eurUSDPrice float64
	var cadUSDPrice float64
	var cnyUSDPrice float64
	var jpyUSDPrice float64
	var gbpUSDPrice float64
	var audUSDPrice float64

	g := &errgroup.Group{}

	g.Go(func() error {
		var err error
		ethUSDPrice, err = getPriceFromFeed(ethUSDFeed)
		if err != nil {
			return fmt.Errorf("error fetching price from EUR/USD feed: %v", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		eurUSDPrice, err = getPriceFromFeed(eurUSDFeed)
		if err != nil {
			return fmt.Errorf("error fetching price from EUR/USD feed: %v", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		cadUSDPrice, err = getPriceFromFeed(cadUSDFeed)
		if err != nil {
			return fmt.Errorf("error fetching price from CAD/USD feed: %v", err)
		}
		return nil
	})

	g.Go(func() error {
		var err error
		cnyUSDPrice, err = getPriceFromFeed(cnyUSDFeed)
		if err != nil {
			return fmt.Errorf("error fetching price from CNY/USD feed: %v", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		jpyUSDPrice, err = getPriceFromFeed(jpyUSDFeed)
		if err != nil {
			return fmt.Errorf("error fetching price from JPY/USD feed: %v", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		gbpUSDPrice, err = getPriceFromFeed(gbpUSDFeed)
		if err != nil {
			return fmt.Errorf("error fetching price from GPY/USD feed: %v", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		audUSDPrice, err = getPriceFromFeed(audUSDFeed)
		if err != nil {
			return fmt.Errorf("error fetching price from AUD/USD feed: %v", err)
		}
		return nil
	})

	err := g.Wait()

	if err != nil {
		logger.Error(err)
		return
	}

	ethPrice = &EthPrice{
		Ethereum: struct {
			Cad float64 "json:\"cad\""
			Cny float64 "json:\"cny\""
			Eur float64 "json:\"eur\""
			Jpy float64 "json:\"jpy\""
			Usd float64 "json:\"usd\""
			Gbp float64 "json:\"gbp\""
			Aud float64 "json:\"aud\""
		}{
			Cad: ethUSDPrice / cadUSDPrice,
			Cny: ethUSDPrice / cnyUSDPrice,
			Eur: ethUSDPrice / eurUSDPrice,
			Jpy: ethUSDPrice / jpyUSDPrice,
			Usd: ethUSDPrice,
			Gbp: ethUSDPrice / gbpUSDPrice,
			Aud: ethUSDPrice / audUSDPrice,
		},
	}
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

func GetEthPrice(currency string) float64 {
	ethPriceMux.RLock()
	defer ethPriceMux.RUnlock()

	switch currency {
	case "EUR":
		return ethPrice.Ethereum.Eur
	case "USD":
		return ethPrice.Ethereum.Usd
	case "CNY":
		return ethPrice.Ethereum.Cny
	case "CAD":
		return ethPrice.Ethereum.Cad
	case "AUD":
		return ethPrice.Ethereum.Aud
	case "JPY":
		return ethPrice.Ethereum.Jpy
	case "GBP":
		return ethPrice.Ethereum.Gbp
	default:
		return 1
	}
}

func GetAvailableCurrencies() []string {
	return availableCurrencies
}

func GetCurrencyLabel(currency string) string {
	switch currency {
	case "ETH":
		return "Ether"
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
	default:
		return ""
	}
}

func GetEthRoundPrice(currency float64) uint64 {
	ethRoundPrice := uint64(currency)
	return ethRoundPrice
}
