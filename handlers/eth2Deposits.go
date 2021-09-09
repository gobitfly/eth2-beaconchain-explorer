package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

var eth2DepositsTemplate = template.Must(template.New("eth2Deposits").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/eth2Deposits.html"))

// Eth2Deposits will return information about deposits using a go template
func Eth2Deposits(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "eth2Deposits", "/deposits/eth2", "Eth2 Deposits")
	data.HeaderAd = true

	err := eth2DepositsTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

// Eth2DepositsData will return information eth1-deposits in json
func Eth2DepositsData(w http.ResponseWriter, r *http.Request) {
	currency := GetCurrency(r)
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	search := q.Get("search[value]")
	search = strings.Replace(search, "0x", "", -1)

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	if length > 100 {
		length = 100
	}

	orderColumn := q.Get("order[0][column]")
	orderByMap := map[string]string{
		"0": "block_slot",
		// "1": "validatorindex",
		"1": "publickey",
		"2": "amount",
		"3": "withdrawalcredentials",
		"4": "signature",
	}
	orderBy, exists := orderByMap[orderColumn]
	if !exists {
		orderBy = "block_ts"
	}

	orderDir := q.Get("order[0][dir]")

	depositCount, err := db.GetEth2DepositsCount()
	if err != nil {
		logger.Errorf("error retrieving eth2_deposit count: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	deposits, err := db.GetEth2Deposits(search, length, start, orderBy, orderDir)
	if err != nil {
		logger.Errorf("error retrieving eth2_deposit data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	tableData := make([][]interface{}, len(deposits))
	for i, d := range deposits {
		tableData[i] = []interface{}{
			utils.FormatBlockSlot(d.BlockSlot),
			utils.FormatPublicKey(d.Publickey),
			utils.FormatDepositAmount(d.Amount, currency),
			utils.FormatHash(d.Withdrawalcredentials),
			utils.FormatHash(d.Signature),
		}
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    depositCount,
		RecordsFiltered: depositCount,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
