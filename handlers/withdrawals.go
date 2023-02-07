package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"math/big"
	"net/http"
	"strconv"
	"strings"
)

// Withdrawals will return information about recent withdrawals
func Withdrawals(w http.ResponseWriter, r *http.Request) {

	var withdrawalsTemplate = templates.GetTemplate("layout.html", "withdrawals.html")

	w.Header().Set("Content-Type", "text/html")

	pageData := &types.WithdrawalsPageData{}
	pageData.Stats = services.GetLatestStats()

	latestChartsPageData := services.LatestChartsPageData()
	if len(latestChartsPageData) != 0 {
		for _, c := range latestChartsPageData {
			if c.Path == "withdrawals" {
				pageData.WithdrawalChart = c
				break
			}
		}
	}
	// var err error
	// pageData.Stats, err = services.CalculateStats()
	// if err != nil {
	// 	logger.Errorf("error getting latest stats: %v", err)
	// }

	data := InitPageData(w, r, "validators", "/withdrawals", "Validator Withdrawals")
	data.HeaderAd = true
	data.Data = pageData

	err := withdrawalsTemplate.ExecuteTemplate(w, "layout", data)
	if handleTemplateError(w, r, "withdrawals.go", "withdrawals", "", err) != nil {
		return // an error has occurred and was processed
	}
}

// WithdrawalsData will return eth1-deposits as json
func WithdrawalsData(w http.ResponseWriter, r *http.Request) {
	currency := GetCurrency(r)

	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	search := q.Get("search[value]")
	search = strings.Replace(search, "0x", "", -1)

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	if length > 100 {
		length = 100
	}

	orderColumn := q.Get("order[0][column]")
	orderByMap := map[string]string{
		"0": "epoch",
		"1": "slot",
		"2": "index",
		"3": "validator",
		"4": "address",
		"5": "amount",
	}
	orderBy, exists := orderByMap[orderColumn]
	if !exists {
		orderBy = "index"
	}

	orderDir := q.Get("order[0][dir]")

	withdrawalCount, err := db.GetTotalWithdrawals()
	if err != nil {
		logger.Errorf("error getting total withdrawal count: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	withdrawals, err := db.GetWithdrawals(search, length, start, orderBy, orderDir)
	if err != nil {
		logger.Errorf("error getting withdrawals: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	tableData := make([][]interface{}, len(withdrawals))
	for i, w := range withdrawals {
		tableData[i] = []interface{}{
			template.HTML(fmt.Sprintf("%v", utils.FormatEpoch(utils.EpochOfSlot(w.Slot)))),
			template.HTML(fmt.Sprintf("%v", utils.FormatBlockSlot(w.Slot))),
			template.HTML(fmt.Sprintf("%v", w.Index)),
			template.HTML(fmt.Sprintf("%v", utils.FormatValidator(w.ValidatorIndex))),
			template.HTML(fmt.Sprintf("%v", utils.FormatTimeFromNow(utils.SlotToTime(w.Slot)))),
			template.HTML(fmt.Sprintf("%v", utils.FormatAddress(w.Address, nil, "", false, false, true))),
			template.HTML(fmt.Sprintf("%v", utils.FormatAmount(new(big.Int).Mul(new(big.Int).SetUint64(w.Amount), big.NewInt(1e9)), currency, 6))),
		}
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    withdrawalCount,
		RecordsFiltered: withdrawalCount,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

// Eth1DepositsData will return eth1-deposits as json
func BLSChangeData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	q := r.URL.Query()

	search := q.Get("search[value]")
	search = strings.Replace(search, "0x", "", -1)

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	if length > 100 {
		length = 100
	}

	orderColumn := q.Get("order[0][column]")
	orderByMap := map[string]string{
		"0": "block_slot",
		"1": "block_slot",
		"2": "validatorindex",
	}
	orderBy, exists := orderByMap[orderColumn]
	if !exists {
		orderBy = "block_slot"
	}

	orderDir := q.Get("order[0][dir]")

	// latestEpoch := services.LatestEpoch()

	total, err := db.GetTotalBLSChanges()
	if err != nil {
		logger.Errorf("Error getting total bls changes: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	blsChange, err := db.GetBLSChanges(search, length, start, orderBy, orderDir)
	if err != nil {
		logger.Errorf("Error getting bls changes: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	tableData := make([][]interface{}, len(blsChange))
	for i, bls := range blsChange {
		tableData[i] = []interface{}{
			template.HTML(fmt.Sprintf("%v", utils.FormatEpoch(utils.EpochOfSlot(bls.Slot)))),
			template.HTML(fmt.Sprintf("%v", utils.FormatBlockSlot(bls.Slot))),
			template.HTML(fmt.Sprintf("%v", utils.FormatValidator(bls.Validatorindex))),
			template.HTML(fmt.Sprintf("%v", utils.FormatHashWithCopy(bls.Signature))),
			template.HTML(fmt.Sprintf("%v", utils.FormatHashWithCopy(bls.BlsPubkey))),
			template.HTML(fmt.Sprintf("%v", utils.FormatAddress(bls.Address, nil, "", false, false, true))),
		}
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    total,
		RecordsFiltered: total,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}
