package handlers

import (
	"encoding/hex"
	"encoding/json"
	"eth2-exporter/eth1data"
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
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
	hasError := false

	txHash, err := hex.DecodeString(strings.ReplaceAll(txHashString, "0x", ""))
	if err != nil {
		logger.Errorf("error parsing tx hash %v: %v", txHashString, err)
		data = InitPageData(w, r, "blockchain", "/tx", "Transaction", txNotFoundTemplateFiles)
		txTemplate = txNotFoundTemplate
	}

	if !hasError {
		txData, err := eth1data.GetEth1Transaction(common.BytesToHash(txHash))
		if err != nil {
			mempool := services.LatestMempoolTransactions()
			mempoolTx := mempool.FindTxByHash(txHashString)
			if mempoolTx != nil {

				data = InitPageData(w, r, "blockchain", "/tx", "Transaction", mempoolTxTemplateFiles)
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
				logger.Errorf("error getting eth1 transaction data: %v", err)
				data = InitPageData(w, r, "blockchain", "/tx", "Transaction", txNotFoundTemplateFiles)
				txTemplate = txNotFoundTemplate
			}
		} else {
			data = InitPageData(w, r, "blockchain", "/tx", "Transaction", txTemplateFiles)
			data.Data = txData
		}
	}
	data.HeaderAd = true
	SetPageDataTitle(data, fmt.Sprintf("Transaction %v", txHashString))
	data.Meta.Path = "/tx/" + txHashString

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
