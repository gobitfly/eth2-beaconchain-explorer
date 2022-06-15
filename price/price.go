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

func Init() {
	go updateEthPrice()
}

func updateEthPrice() {
	for {
		fetchPrice()
		time.Sleep(time.Minute)
	}
}

func fetchPrice() {
	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Get("https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd%2Ceur%2Crub%2Ccny%2Ccad%2Cjpy%2Cgbp%2Caud")

	if err != nil {
		logger.Errorf("error retrieving ETH price: %v", err)
		return
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

func GetEthRoundPrice(currency float64) uint64 {
	ethRoundPrice := uint64(currency)
	return ethRoundPrice
}
