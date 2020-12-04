package price

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var logger = logrus.New().WithField("module", "price")

type EthPrice struct {
	USD float64
}

var ethPrice *EthPrice = new(EthPrice)

func Init() {
	go updateEthPrice()
}

func updateEthPrice() {
	for true {
		resp, err := http.Get("https://min-api.cryptocompare.com/data/price?fsym=ETH&tsyms=USD")

		if err != nil {
			logger.Errorf("error retrieving ETH price: %v", err)
			time.Sleep(time.Second * 10)
			continue
		}

		defer resp.Body.Close()

		err = json.NewDecoder(resp.Body).Decode(&ethPrice)

		if err != nil {
			logger.Errorf("error decoding ETH price json response to struct: %v", err)
			time.Sleep(time.Second * 10)
			continue
		}
		time.Sleep(time.Minute)
	}
}

func GetEthPrice() int {
	return int(ethPrice.USD)
}
