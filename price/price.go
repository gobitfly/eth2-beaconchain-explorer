package price

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New().WithField("module", "price")

type EthPrice struct {
	Ethereum struct {
		Cad float64 `json:"cad"`
		Cny float64 `json:"cny"`
		Eur float64 `json:"eur"`
		Jpy float64 `json:"jpy"`
		Rub float64 `json:"rub"`
		Usd float64 `json:"usd"`
		Gbp float64 `json:"gbp"`
		Aud float64 `json:"aud"`
	} `json:"ethereum"`
}

var ethPrice = new(EthPrice)
var ethPriceMux = &sync.RWMutex{}

func Init(chainId uint64) {
	go updateEthPrice(chainId)
}

func updateEthPrice(chainId uint64) {
	errorRetrievingEthPriceCount := 0
	for {
		fetchPrice(chainId, &errorRetrievingEthPriceCount)
		time.Sleep(time.Minute)
	}
}

func fetchPrice(chainId uint64, errorRetrievingEthPriceCount *int) {
	if chainId != 1 {
		ethPrice = &EthPrice{
			Ethereum: struct {
				Cad float64 "json:\"cad\""
				Cny float64 "json:\"cny\""
				Eur float64 "json:\"eur\""
				Jpy float64 "json:\"jpy\""
				Rub float64 "json:\"rub\""
				Usd float64 "json:\"usd\""
				Gbp float64 "json:\"gbp\""
				Aud float64 "json:\"aud\""
			}{
				Cad: 0,
				Cny: 0,
				Eur: 0,
				Jpy: 0,
				Rub: 0,
				Usd: 0,
				Gbp: 0,
				Aud: 0,
			},
		}
		return
	}

	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Get("https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd%2Ceur%2Crub%2Ccny%2Ccad%2Cjpy%2Cgbp%2Caud")
	if err != nil {
		*errorRetrievingEthPriceCount++
		if *errorRetrievingEthPriceCount <= 3 { // warn 3 times, before throwing errors starting with the fourth time
			logger.Warnf("error (%d) retrieving ETH price: %v", *errorRetrievingEthPriceCount, err)
		} else {
			logger.Errorf("error (%d) retrieving ETH price: %v", *errorRetrievingEthPriceCount, err)
		}
		return
	} else {
		*errorRetrievingEthPriceCount = 0
	}

	ethPriceMux.Lock()
	defer ethPriceMux.Unlock()
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&ethPrice)
	if err != nil {
		logger.Errorf("error decoding ETH price json response to struct: %v", err)
		return
	}
}

func GetEthPrice(currency string) float64 {
	ethPriceMux.RLock()
	defer ethPriceMux.RUnlock()

	switch currency {
	case "EUR":
		return ethPrice.Ethereum.Eur
	case "USD":
		return ethPrice.Ethereum.Usd
	case "RUB":
		return ethPrice.Ethereum.Rub
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
