package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"net/http"
	"strconv"
	"strings"
)

// Eth2Deposits will return information about deposits using a go template
func Eth2Deposits(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/validators/deposits", http.StatusMovedPermanently)
}

// Eth2DepositsData will return information eth1-deposits in json
func Eth2DepositsData(w http.ResponseWriter, r *http.Request) {
	currency := GetCurrency(r)
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	search := ReplaceEnsNameWithAddress(q.Get("search[value]"))
	search = strings.Replace(search, "0x", "", -1)

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables draw parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter draw", http.StatusBadRequest)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter start", http.StatusBadRequest)
		return
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter length", http.StatusBadRequest)
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

	deposits, depositCount, err := db.GetEth2Deposits(search, length, start, orderBy, orderDir)
	if err != nil {
		logger.Errorf("error retrieving eth2_deposit data or count: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tableData := make([][]interface{}, len(deposits))
	for i, d := range deposits {
		tableData[i] = []interface{}{
			utils.FormatBlockSlot(d.BlockSlot),
			utils.FormatPublicKey(d.Publickey),
			utils.FormatDepositAmount(d.Amount, currency),
			utils.FormatWithdawalCredentials(d.Withdrawalcredentials, false),
			utils.FormatHash(d.Signature),
			utils.FormatHash(d.Withdrawalcredentials, false),
			utils.FormatHash(d.Signature, false),
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
