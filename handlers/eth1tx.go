package handlers

import (
	"encoding/hex"
	"encoding/json"
	"eth2-exporter/eth1data"
	"eth2-exporter/templates"
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

	w.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(r)
	txHashString := strings.Replace(vars["hash"], "0x", "", -1)

	data := InitPageData(w, r, "blockchain", "/tx", "Transaction")
	data.HeaderAd = true

	txHash, err := hex.DecodeString(strings.ReplaceAll(txHashString, "0x", ""))

	if err != nil {

		SetPageDataTitle(data, fmt.Sprintf("Transaction %v", txHashString))
		data.Meta.Path = "/tx/" + txHashString
		logger.Errorf("error parsing tx hash %v: %v", txHashString, err)

		if handleTemplateError(w, r, txNotFoundTemplate.ExecuteTemplate(w, "layout", data)) != nil {
			return // an error has occurred and was processed
		}
		return
	}

	SetPageDataTitle(data, fmt.Sprintf("Transaction 0x%x", txHash))
	data.Meta.Path = fmt.Sprintf("/tx/0x%x", txHash)

	txData, err := eth1data.GetEth1Transaction(common.BytesToHash(txHash))

	if err != nil {
		SetPageDataTitle(data, fmt.Sprintf("Transaction 0x%v", txHashString))
		data.Meta.Path = "/tx/" + txHashString
		logger.Errorf("error getting eth1 transaction data: %v", err)

		if handleTemplateError(w, r, txNotFoundTemplate.ExecuteTemplate(w, "layout", data)) != nil {
			return // an error has occurred and was processed
		}
		return
	}

	data.Data = txData

	if utils.IsApiRequest(r) {
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(data.Data)
	} else {
		err = txTemplate.ExecuteTemplate(w, "layout", data)
	}

	if handleTemplateError(w, r, err) != nil {
		return // an error has occurred and was processed
	}
}
