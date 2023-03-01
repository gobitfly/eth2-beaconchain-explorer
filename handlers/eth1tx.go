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

	var txNotFoundTemplate = templates.GetTemplate("layout.html", "eth1txnotfound.html")
	var txTemplate = templates.GetTemplate("layout.html", "eth1tx.html")
	var mempoolTxTemplate = templates.GetTemplate("layout.html", "mempoolTx.html")

	w.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(r)
	txHashString := vars["hash"]

	data := InitPageData(w, r, "blockchain", "/tx", "Transaction")
	data.HeaderAd = true

	SetPageDataTitle(data, fmt.Sprintf("Transaction %v", txHashString))
	data.Meta.Path = "/tx/" + txHashString

	txHash, err := hex.DecodeString(strings.ReplaceAll(txHashString, "0x", ""))
	if err != nil {
		logger.Errorf("error parsing tx hash %v: %v", txHashString, err)

		if handleTemplateError(w, r, "eth1tx.go", "Eth1TransactionTx", "decodeString", txNotFoundTemplate.ExecuteTemplate(w, "layout", data)) != nil {
			return // an error has occurred and was processed
		}
		return
	}

	txData, err := eth1data.GetEth1Transaction(common.BytesToHash(txHash))
	if err != nil {
		mempool := services.LatestMempoolTransactions()
		mempoolTx := mempool.FindTxByHash(txHashString)
		if mempoolTx != nil {
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
			if handleTemplateError(w, r, "eth1tx.go", "Eth1TransactionTx", "GetEth1Transaction", txNotFoundTemplate.ExecuteTemplate(w, "layout", data)) != nil {
				return // an error has occurred and was processed
			}
			return
		}
	} else {
		data.Data = txData
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
