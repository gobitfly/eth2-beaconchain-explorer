package handlers

// import (
// 	"encoding/json"
// 	"eth2-exporter/db"
// 	"eth2-exporter/services"
// 	"eth2-exporter/templates"
// 	"eth2-exporter/types"
// 	"eth2-exporter/utils"
// 	"fmt"
// 	"html/template"
// 	"math/big"
// 	"net/http"
// 	"strconv"
// 	"strings"
// )

// // Withdrawals will return information about recent withdrawals
// func Withdrawals(w http.ResponseWriter, r *http.Request) {

// 	var withdrawalsTemplate = templates.GetTemplate("layout.html", "withdrawals.html")

// 	w.Header().Set("Content-Type", "text/html")

// 	pageData := &types.WithdrawalsPageData{}

// 	pageData.Stats = services.GetLatestStats()

// 	data := InitPageData(w, r, "validators", "/withdrawals", "Validator Withdrawals")
// 	data.HeaderAd = true
// 	data.Data = pageData

// 	if handleTemplateError(w, r, withdrawalsTemplate.ExecuteTemplate(w, "layout", data)) != nil {
// 		return // an error has occurred and was processed
// 	}
// }

// // WithdrawalsData will return eth1-deposits as json
// func WithdrawalsData(w http.ResponseWriter, r *http.Request) {
// 	currency := GetCurrency(r)

// 	w.Header().Set("Content-Type", "application/json")

// 	q := r.URL.Query()

// 	search := q.Get("search[value]")
// 	search = strings.Replace(search, "0x", "", -1)

// 	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
// 	if err != nil {
// 		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
// 		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
// 		return
// 	}

// 	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
// 	if err != nil {
// 		logger.Errorf("error converting datatables start parameter from string to int: %v", err)
// 		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
// 		return
// 	}
// 	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
// 	if err != nil {
// 		logger.Errorf("error converting datatables length parameter from string to int: %v", err)
// 		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
// 		return
// 	}
// 	if length > 100 {
// 		length = 100
// 	}

// 	orderColumn := q.Get("order[0][column]")
// 	orderByMap := map[string]string{
// 		"0": "epoch",
// 		"1": "slot",
// 		"2": "index",
// 		"3": "validator",
// 		"4": "address",
// 		"5": "amount",
// 	}
// 	orderBy, exists := orderByMap[orderColumn]
// 	if !exists {
// 		orderBy = "index"
// 	}

// 	orderDir := q.Get("order[0][dir]")

// 	withdrawalCount, err := db.GetTotalWithdrawals()
// 	if err != nil {
// 		logger.Errorf("error getting total withdrawal count: %v", err)
// 		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
// 		return
// 	}

// 	withdrawals, err := db.GetWithdrawals(search, length, start, orderBy, orderDir)
// 	if err != nil {
// 		logger.Errorf("error getting withdrawals: %v", err)
// 		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
// 		return
// 	}

// 	tableData := make([][]interface{}, len(withdrawals))
// 	for i, w := range withdrawals {
// 		tableData[i] = []interface{}{
// 			template.HTML(fmt.Sprintf("%v", utils.FormatEpoch(utils.EpochOfSlot(w.Slot)))),
// 			template.HTML(fmt.Sprintf("%v", utils.FormatBlockSlot(w.Slot))),
// 			template.HTML(fmt.Sprintf("%v", w.Index)),
// 			template.HTML(fmt.Sprintf("%v", utils.FormatTimeFromNow(utils.SlotToTime(w.Slot)))),
// 			template.HTML(fmt.Sprintf("%v", utils.FormatAddress(w.Address, nil, "", false, false, true))),
// 			template.HTML(fmt.Sprintf("%v", utils.FormatAmount(new(big.Int).Mul(new(big.Int).SetUint64(w.Amount), big.NewInt(1e9)), currency, 6))),
// 		}
// 	}

// 	data := &types.DataTableResponse{
// 		Draw:            draw,
// 		RecordsTotal:    withdrawalCount,
// 		RecordsFiltered: withdrawalCount,
// 		Data:            tableData,
// 	}

// 	err = json.NewEncoder(w).Encode(data)
// 	if err != nil {
// 		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
// 		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
// 		return
// 	}
// }

// // Eth1DepositsData will return eth1-deposits as json
// func BLSChangeData(w http.ResponseWriter, r *http.Request) {
// 	currency := GetCurrency(r)
// 	w.Header().Set("Content-Type", "application/json")
// 	q := r.URL.Query()

// 	search := q.Get("search[value]")
// 	search = strings.Replace(search, "0x", "", -1)

// 	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
// 	if err != nil {
// 		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
// 		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
// 		return
// 	}
// 	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
// 	if err != nil {
// 		logger.Errorf("error converting datatables start parameter from string to int: %v", err)
// 		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
// 		return
// 	}
// 	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
// 	if err != nil {
// 		logger.Errorf("error converting datatables length parameter from string to int: %v", err)
// 		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
// 		return
// 	}
// 	if length > 100 {
// 		length = 100
// 	}

// 	orderColumn := q.Get("order[0][column]")
// 	orderByMap := map[string]string{
// 		"0": "from_address",
// 		"1": "amount",
// 		"2": "validcount",
// 		"3": "invalidcount",
// 		"4": "pendingcount",
// 		"5": "activecount",
// 		"6": "slashedcount",
// 		"7": "voluntary_exit_count",
// 		"8": "totalcount",
// 	}
// 	orderBy, exists := orderByMap[orderColumn]
// 	if !exists {
// 		orderBy = "amount"
// 	}

// 	orderDir := q.Get("order[0][dir]")

// 	latestEpoch := services.LatestEpoch()

// 	deposits, depositCount, err := db.GetEth1DepositsLeaderboard(search, length, start, orderBy, orderDir, latestEpoch)
// 	if err != nil {
// 		logger.Errorf("GetEth1Deposits error retrieving eth1_deposit leaderboard data: %v", err)
// 		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
// 		return
// 	}

// 	tableData := make([][]interface{}, len(deposits))
// 	for i, d := range deposits {
// 		tableData[i] = []interface{}{
// 			utils.FormatEth1Address(d.FromAddress),
// 			utils.FormatBalance(d.Amount, currency),
// 			d.ValidCount,
// 			d.InvalidCount,
// 			d.PendingCount,
// 			d.ActiveCount,
// 			d.SlashedCount,
// 			d.VoluntaryExitCount,
// 			d.TotalCount,
// 			// utils.FormatPublicKey(d.PublicKey),
// 			// utils.FormatDepositAmount(d.Amount),
// 			// utils.FormatEth1TxHash(d.TxHash),
// 			// utils.FormatTimestamp(d.BlockTs.Unix()),
// 			// utils.FormatEth1Block(d.BlockNumber),
// 			// utils.FormatValidatorStatus(d.State),
// 			// d.ValidSignature,
// 		}
// 	}

// 	data := &types.DataTableResponse{
// 		Draw:            draw,
// 		RecordsTotal:    depositCount,
// 		RecordsFiltered: depositCount,
// 		Data:            tableData,
// 	}

// 	err = json.NewEncoder(w).Encode(data)
// 	if err != nil {
// 		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
// 		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
// 		return
// 	}
// }
