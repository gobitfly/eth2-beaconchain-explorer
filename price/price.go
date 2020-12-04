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
	USD float64
	EUR float64
	RUB float64
	CNY float64
}

var ethPrice = new(EthPrice)
var ethPriceMux = &sync.RWMutex{}

func Init() {
	go updateEthPrice()
}

func updateEthPrice() {
	for true {
		fetchPrice()
	}
}

func fetchPrice() {
	resp, err := http.Get("https://min-api.cryptocompare.com/data/price?fsym=ETH&tsyms=USD,EUR,RUB,CNY")

	if err != nil {
		logger.Errorf("error retrieving ETH price: %v", err)
		time.Sleep(time.Second * 10)
		return
	}

	ethPriceMux.Lock()
	defer ethPriceMux.Unlock()
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&ethPrice)

	if err != nil {
		logger.Errorf("error decoding ETH price json response to struct: %v", err)
		time.Sleep(time.Second * 10)
		return
	}
	time.Sleep(time.Minute * 5)
}
func GetEthPrice(currency string) float64 {
	ethPriceMux.RLock()
	defer ethPriceMux.RUnlock()

	switch currency {
	case "EUR":
		return ethPrice.EUR
	case "USD":
		return ethPrice.USD
	case "RUB":
		return ethPrice.RUB
	case "CNY":
		return ethPrice.CNY
	default:
		return 1
	}
}
