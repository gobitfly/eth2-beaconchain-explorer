package price

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
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
	} `json:"ethereum"`
}

var ethPrice = new(EthPrice)
var ethPriceMux = &sync.RWMutex{}

func Init() {
	go updateEthPrice()
}

func updateEthPrice() {
	for true {
		fetchPrice()
		time.Sleep(time.Minute)
	}
}

func fetchPrice() {
	resp, err := http.Get("https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd%2Ceur%2Crub%2Ccny%2Ccad%2Cjpy")

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
	case "JPY":
		return ethPrice.Ethereum.Jpy
	default:
		return 1
	}
}
