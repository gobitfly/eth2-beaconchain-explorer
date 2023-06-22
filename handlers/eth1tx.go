package handlers

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"eth2-exporter/db"
	"eth2-exporter/eth1data"
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Tx will show the tx using a go template
func Eth1TransactionTx(w http.ResponseWriter, r *http.Request) {
	txNotFoundTemplateFiles := append(layoutTemplateFiles, "eth1txnotfound.html")
	txTemplateFiles := append(layoutTemplateFiles, "eth1tx.html")
	mempoolTxTemplateFiles := append(layoutTemplateFiles, "mempoolTx.html")
	var txNotFoundTemplate = templates.GetTemplate(txNotFoundTemplateFiles...)
	var txTemplate = templates.GetTemplate(txTemplateFiles...)
	var mempoolTxTemplate = templates.GetTemplate(mempoolTxTemplateFiles...)

	w.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(r)
	txHashString := vars["hash"]
	var data *types.PageData
	title := fmt.Sprintf("Transaction %v", txHashString)
	path := fmt.Sprintf("/tx/%v", txHashString)

	txHash, err := hex.DecodeString(strings.ReplaceAll(txHashString, "0x", ""))
	if err != nil {
		logger.Errorf("error parsing tx hash %v: %v", txHashString, err)
		data = InitPageData(w, r, "blockchain", path, title, txNotFoundTemplateFiles)
		txTemplate = txNotFoundTemplate
	} else {
		txData, err := eth1data.GetEth1Transaction(common.BytesToHash(txHash))
		if err != nil {
			mempool := services.LatestMempoolTransactions()
			mempoolTx := mempool.FindTxByHash(txHashString)
			if mempoolTx != nil {

				data = InitPageData(w, r, "blockchain", path, title, mempoolTxTemplateFiles)
				mempoolPageData := &types.MempoolTxPageData{RawMempoolTransaction: *mempoolTx}
				txTemplate = mempoolTxTemplate
				if mempoolTx.To == nil {
					mempoolPageData.IsContractCreation = true
				}
				if mempoolTx.Input != nil {
					mempoolPageData.TargetIsContract = true
				}

				data.Data = mempoolPageData
			} else {
				if !errors.Is(err, ethereum.NotFound) {
					logger.Errorf("error getting eth1 transaction data: %v", err)
				}
				data = InitPageData(w, r, "blockchain", path, title, txNotFoundTemplateFiles)
				txTemplate = txNotFoundTemplate
			}
		} else {
			p := message.NewPrinter(language.English)

			symbol := GetCurrencySymbol(r)
			ef := new(big.Float).SetInt(new(big.Int).SetBytes(txData.Value))
			etherValue := new(big.Float).Quo(ef, big.NewFloat(1e18))

			currentPrice := GetCurrentPrice(r)
			currentEthPrice := new(big.Float).Mul(etherValue, big.NewFloat(float64(currentPrice)))
			cPrice, _ := currentEthPrice.Float64()
			txData.CurrentEtherPrice = template.HTML(p.Sprintf(`<span>%s%.2f</span>`, symbol, cPrice))

			txDay := utils.TimeToDay(uint64(txData.Timestamp.Unix()))
			latestEpoch, err := db.GetLatestEpoch()
			if err != nil {
				logrus.Error(err)
			}

			txData.HistoricalEtherPrice = ""
			currentDay := latestEpoch / utils.EpochsPerDay()

			if txDay < currentDay {
				// Do not show the historical price if it is the current day
				currency := GetCurrency(r)
				price, err := db.GetHistoricalPrice(utils.Config.Chain.Config.DepositChainID, currency, txDay)
				if err != nil {
					utils.LogError(err, "error retrieving historical prices", 0, map[string]interface{}{"txDay": txDay, "currency": currency})
				} else {
					historicalEthPrice := new(big.Float).Mul(etherValue, big.NewFloat(price))
					hPrice, _ := historicalEthPrice.Float64()
					txData.HistoricalEtherPrice = template.HTML(p.Sprintf(`<span>%s%.2f <i class="far fa-clock"></i></span>`, symbol, hPrice))
				}
			}

			data = InitPageData(w, r, "blockchain", path, title, txTemplateFiles)
			data.Data = txData
		}
	}

	if utils.IsApiRequest(r) {
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(data.Data)
	} else {
		err = txTemplate.ExecuteTemplate(w, "layout", data)
	}

	if handleTemplateError(w, r, "eth1tx.go", "Eth1TransactionTx", "Done", err) != nil {
		return // an error has occurred and was processed
	}
}
