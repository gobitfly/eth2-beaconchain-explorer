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
	currency := GetCurrency(r)
	templateFiles := append(layoutTemplateFiles, "withdrawals.html", "validator/withdrawalOverviewRow.html", "components/charts.html")
	var withdrawalsTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	pageData := &types.WithdrawalsPageData{}
	pageData.Stats = services.GetLatestStats()

	data := InitPageData(w, r, "validators", "/withdrawals", "Validator Withdrawals", templateFiles)

	latestChartsPageData := services.LatestChartsPageData()
	if len(latestChartsPageData) != 0 {
		for _, c := range latestChartsPageData {
			if c.Path == "withdrawals" {
				pageData.WithdrawalChart = c
				break
			}
		}
	}

	// withdrawalChartData, err := services.WithdrawalsChartData()
	// if err != nil {
	// 	logger.Errorf("error getting withdrawal chart data: %v", err)
	// 	http.Error(w, "Internal server error", http.StatusServiceUnavailable)
	// 	return
	// }
	// pageData.WithdrawalChart = &types.ChartsPageDataChart{
	// 	Data:   withdrawalChartData,
	// 	Order:  17,
	// 	Path:   "withdrawals",
	// 	Height: 300,
	// }

	user, session, err := getUserSession(r)
	if err != nil {
		logger.WithError(err).Error("error getting user session")
	}

	state := GetDataTableState(user, session, "withdrawals")
	if state.Length == 0 {
		state.Length = 10
	}

	withdrawals, err := WithdrawalsTableData(1, state.Search.Search, state.Length, state.Start, "", "", currency)
	if err != nil {
		logger.Errorf("error getting withdrawals table data: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	pageData.Withdrawals = withdrawals

	blsChange, err := BLSTableData(1, state.Search.Search, state.Length, state.Start, "", "")
	if err != nil {
		logger.Errorf("error getting bls table data: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	pageData.BlsChanges = blsChange

	data.Data = pageData

	err = withdrawalsTemplate.ExecuteTemplate(w, "layout", data)
	if handleTemplateError(w, r, "withdrawals.go", "withdrawals", "", err) != nil {
		return // an error has occurred and was processed
	}
}

// WithdrawalsData will return eth1-deposits as json
func WithdrawalsData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	currency := GetCurrency(r)
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

	orderBy := q.Get("order[0][column]")
	orderDir := q.Get("order[0][dir]")

	data, err := WithdrawalsTableData(draw, search, length, start, orderBy, orderDir, currency)
	if err != nil {
		logger.Errorf("error getting withdrawal table data: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func WithdrawalsTableData(draw uint64, search string, length, start uint64, orderBy, orderDir string, currency string) (*types.DataTableResponse, error) {
	orderByMap := map[string]string{
		"0": "epoch",
		"1": "slot",
		"2": "index",
		"3": "validator",
		"4": "address",
		"5": "amount",
	}
	orderColumn, exists := orderByMap[orderBy]
	if !exists {
		orderBy = "index"
	}

	if orderDir != "asc" {
		orderDir = "desc"
	}

	withdrawalCount, err := db.GetTotalWithdrawals()
	if err != nil {
		return nil, fmt.Errorf("error getting total withdrawals: %w", err)
	}

	withdrawals, err := db.GetWithdrawals(search, length, start, orderColumn, orderDir)
	if err != nil {
		return nil, fmt.Errorf("error getting withdrawals: %w", err)
	}

	formatCurrency := currency
	if currency == "ETH" {
		formatCurrency = "Ether"
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
			template.HTML(fmt.Sprintf("%v", utils.FormatAmount(new(big.Int).Mul(new(big.Int).SetUint64(w.Amount), big.NewInt(1e9)), formatCurrency, 6))),
		}
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    withdrawalCount,
		RecordsFiltered: withdrawalCount,
		Data:            tableData,
		PageLength:      length,
		DisplayStart:    start,
	}
	return data, nil
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

	orderBy := q.Get("order[0][column]")
	orderDir := q.Get("order[0][dir]")

	data, err := BLSTableData(draw, search, length, start, orderBy, orderDir)
	if err != nil {
		logger.Errorf("Error getting bls changes: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func BLSTableData(draw uint64, search string, length, start uint64, orderBy, orderDir string) (*types.DataTableResponse, error) {

	orderByMap := map[string]string{
		"0": "block_slot",
		"1": "block_slot",
		"2": "validatorindex",
	}
	orderVar, exists := orderByMap[orderBy]
	if !exists {
		orderBy = "block_slot"
	}

	if orderDir != "asc" {
		orderDir = "desc"
	}

	total, err := db.GetTotalBLSChanges()
	if err != nil {
		return nil, fmt.Errorf("error getting total bls changes: %w", err)
	}

	blsChange, err := db.GetBLSChanges(search, length, start, orderVar, orderDir)
	if err != nil {
		return nil, fmt.Errorf("error getting bls changes: %w", err)
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
		PageLength:      length,
		DisplayStart:    start,
	}
	return data, nil
}
