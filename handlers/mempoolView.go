package handlers

import (
	"encoding/json"
	"eth2-exporter/utils"
	"html/template"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/sirupsen/logrus"

	gethRPC "github.com/ethereum/go-ethereum/rpc"
)

func MempoolView(w http.ResponseWriter, r *http.Request) {

	rawMempoolData := _fetchRawMempoolData()
	formatedData := formatToHtml(rawMempoolData)

	var err error
	var mempoolViewTemplate = template.Must(template.New("mempoolview").Funcs(utils.GetTemplateFuncs()).ParseFiles(
		"templates/layout.html",
		"templates/mempoolview.html",
	))
	w.Header().Set("Content-Type", "text/html")
	data := InitPageData(w, r, "services", "/mempool", "Pending Mempool Transactions")

	data.Data = formatedData

	err = mempoolViewTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}

}

// this function connects to RPC or to local cache and requests raw memory pool data.
func _fetchRawMempoolData() json.RawMessage {

	client, err := gethRPC.Dial(utils.Config.Eth1GethEndpoint)
	var raw json.RawMessage

	err = client.Call(&raw, "txpool_content")
	if err != nil {
		logrus.Error("node rpc command error: ", err)
	}
	return raw
}

// This is a helper function. If receiver Address is nil or empty(in case of a new contract creation).
// This function catches the Nil exception
func isContractCreation(tx *common.Address) string {
	if tx == nil {
		return "Contract Creation"
	} else {
		return string(utils.FormatAddressAll(tx.Bytes(), "", false, "address", "", int(12), int(12), true))
	}

}

// This Function formats each Transaction into Html strings.
// This should make all calculations faster, reducing browser's rendering time.
func formatToHtml(content json.RawMessage) []formatedTx {

	rawMempoolData := RawMempoolResponse{}
	err := json.Unmarshal(content, &rawMempoolData)
	if err != nil {
		logrus.Error("JSON Unmarshalling failed: ", err)
	}

	var htmlFormatedData []formatedTx
	for _, pendingData := range rawMempoolData.Pending {
		for _, tx := range pendingData {
			htmlFormatedData = append(htmlFormatedData, formatedTx{Hash: template.HTML(tx.Hash.String()),
				From:  utils.FormatAddressAll(tx.From.Bytes(), "", false, "address", "", int(12), int(12), true),
				To:    template.HTML(isContractCreation(tx.To)),
				Value: utils.FormatAmount((*big.Int)(tx.Value), "ETH", 5),
				Gas:   utils.FormatAmountFormated(tx.Gas.ToInt(), "GWei", 5, 0, true, true, false)})
		}
	}
	return htmlFormatedData
}

type RawMempoolResponse struct {
	Pending map[string]map[int]rawTransaction
}

type rawTransaction struct {
	Hash      common.Hash     `json:"hash"`
	From      *common.Address `json:"from"`
	To        *common.Address `json:"to"`
	Value     *hexutil.Big    `json:"value"`
	Gas       *hexutil.Big    `json:"gas"`
	GasFeeCap *hexutil.Big    `json:"maxFeePerGas,omitempty"`
}

type formatedTx struct {
	Hash      template.HTML `json:"hash"`
	From      template.HTML `json:"from"`
	To        template.HTML `default:"Empty address"`
	Value     template.HTML `json:"value"`
	Gas       template.HTML `json:"gas"`
	GasFeeCap template.HTML `json:"maxFeePerGas,omitempty"`
}
