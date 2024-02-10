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
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
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

	errFields := map[string]interface{}{
		"route":        r.URL.String(),
		"txHashString": txHashString,
	}

	txHash, err := hex.DecodeString(strings.ReplaceAll(txHashString, "0x", ""))
	if err != nil {
		logger.Warnf("error parsing tx hash %v: %v", txHashString, err)
		data = InitPageData(w, r, "blockchain", path, title, txNotFoundTemplateFiles)
		txTemplate = txNotFoundTemplate
	} else {
		txData, err := eth1data.GetEth1Transaction(common.BytesToHash(txHash), "ETH")
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
				if !errors.Is(err, ethereum.NotFound) && !errors.Is(err, eth1data.ErrTxIsPending) {
					utils.LogError(err, "error getting eth1 transaction data", 0, errFields)
				}
				data = InitPageData(w, r, "blockchain", path, title, txNotFoundTemplateFiles)
				txTemplate = txNotFoundTemplate
			}
		} else {
			p := message.NewPrinter(language.English)

			symbol := GetCurrencySymbol(r)
			etherValue := utils.WeiBytesToEther(txData.Value)

			currentPrice := GetCurrentPrice(r)
			currentEthPrice := etherValue.Mul(decimal.NewFromInt(int64(currentPrice)))
			txData.CurrentEtherPrice = template.HTML(p.Sprintf(`<span>%s%.2f</span>`, symbol, currentEthPrice.InexactFloat64()))

			txData.HistoricalEtherPrice = ""
			if txData.Timestamp.Unix() >= int64(utils.Config.Chain.GenesisTimestamp) {
				txDay := utils.TimeToDay(uint64(txData.Timestamp.Unix()))
				errFields["txDay"] = txDay
				latestEpoch, err := db.GetLatestEpoch()
				if err != nil {
					utils.LogError(err, "error retrieving latest epoch from db", 0, errFields)
				}

				currentDay := latestEpoch / utils.EpochsPerDay()

				if txDay < currentDay {
					// Do not show the historical price if it is the current day
					currency := GetCurrency(r)
					price, err := db.GetHistoricalPrice(utils.Config.Chain.ClConfig.DepositChainID, currency, txDay)
					if err != nil {
						errFields["currency"] = currency
						utils.LogError(err, "error retrieving historical prices", 0, errFields)
					} else {
						historicalEthPrice := etherValue.Mul(decimal.NewFromFloat(price))
						txData.HistoricalEtherPrice = template.HTML(p.Sprintf(`<span>%s%.2f <i class="far fa-clock"></i></span>`, symbol, historicalEthPrice.InexactFloat64()))
					}
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

func Eth1TransactionTxData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	txHashString := vars["hash"]
	currency := GetCurrency(r)
	err := json.NewEncoder(w).Encode(getEth1TransactionTxData(txHashString, currency))
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func getEth1TransactionTxData(txhash, currency string) *types.DataTableResponse {
	tableData := make([][]interface{}, 0, minimumTransactionsPerUpdate)
	txHash, err := hex.DecodeString(strings.ReplaceAll(txhash, "0x", ""))
	if err != nil {
		logger.Warnf("error parsing tx hash %v: %v", txhash, err)
	} else {
		txData, err := eth1data.GetEth1Transaction(common.BytesToHash(txHash), currency)
		its := txData.InternalTxns
		if err != nil {
			utils.LogError(err, "error getting transaction data", 0, map[string]interface{}{"txhash": txHash})
		} else {
			for _, i := range its {
				tableData = append(tableData, []interface{}{
					i.TracePath,
					i.From,
					i.To,
					i.Amount,
					i.Gas.Limit,
				})
			}
		}
	}

	return &types.DataTableResponse{
		Data: tableData,
	}
}
