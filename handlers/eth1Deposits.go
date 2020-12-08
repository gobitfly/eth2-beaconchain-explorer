package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

var eth1DepositsTemplate = template.Must(template.New("eth1Deposits").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/eth1Deposits.html", "templates/index/depositChart.html"))
var eth1DepositsLeaderboardTemplate = template.Must(template.New("eth1Deposits").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/eth1DepositsLeaderboard.html"))

// Eth1Deposits will return information about deposits using a go template
func Eth1Deposits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	pageData := &types.EthOneDepositsPageData{}

	latestChartsPageData := services.LatestChartsPageData()
	if latestChartsPageData != nil {
		for _, c := range *latestChartsPageData {
			if c.Path == "deposits" {
				pageData.DepositChart = c
				break
			}
		}
	}

	pageData.Stats = services.GetLatestStats()
	pageData.DepositContract = utils.Config.Indexer.Eth1DepositContractAddress

	data := InitPageData(w, r, "eth1Deposits", "/deposits/eth1", "Eth1 Deposits")
	data.HeaderAd = true
	data.Data = pageData

	err := eth1DepositsTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

// Eth1DepositsData will return eth1-deposits as json
func Eth1DepositsData(w http.ResponseWriter, r *http.Request) {
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
		"0": "from_address",
		"1": "publickey",
		"2": "amount",
		"3": "tx_hash",
		"4": "block_ts",
		"5": "block_number",
		"6": "state",
		"7": "valid_signature",
	}
	orderBy, exists := orderByMap[orderColumn]
	if !exists {
		orderBy = "block_ts"
	}

	orderDir := q.Get("order[0][dir]")

	latestEpoch := services.LatestEpoch()
	validatorOnlineThresholdSlot := GetValidatorOnlineThresholdSlot()

	deposits, depositCount, err := db.GetEth1DepositsJoinEth2Deposits(search, length, start, orderBy, orderDir, latestEpoch, validatorOnlineThresholdSlot)
	if err != nil {
		logger.Errorf("GetEth1Deposits error retrieving eth1_deposit data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	tableData := make([][]interface{}, len(deposits))
	for i, d := range deposits {
		valid := "❌"
		if d.ValidSignature {
			valid = "✅"
		}
		tableData[i] = []interface{}{
			utils.FormatEth1Address(d.FromAddress),
			utils.FormatPublicKey(d.PublicKey),
			utils.FormatDepositAmount(d.Amount, currency),
			utils.FormatEth1TxHash(d.TxHash),
			utils.FormatTimestamp(d.BlockTs.Unix()),
			utils.FormatEth1Block(d.BlockNumber),
			utils.FormatValidatorStatus(d.State),
			valid,
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

// Eth1Deposits will return information about deposits using a go template
func Eth1DepositsLeaderboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "eth1Deposits", "/deposits/eth1", "Eth1 Deposits")
	data.HeaderAd = true

	data.Data = types.EthOneDepositLeaderBoardPageData{
		DepositContract: utils.Config.Indexer.Eth1DepositContractAddress,
	}

	err := eth1DepositsLeaderboardTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

// Eth1DepositsData will return eth1-deposits as json
func Eth1DepositsLeaderboardData(w http.ResponseWriter, r *http.Request) {
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
		"0": "from_address",
		"1": "amount",
		"2": "validcount",
		"3": "invalidcount",
		"4": "pendingcount",
		"5": "activecount",
		"6": "slashedcount",
		"7": "voluntary_exit_count",
		"8": "totalcount",
	}
	orderBy, exists := orderByMap[orderColumn]
	if !exists {
		orderBy = "amount"
	}

	orderDir := q.Get("order[0][dir]")

	latestEpoch := services.LatestEpoch()

	deposits, depositCount, err := db.GetEth1DepositsLeaderboard(search, length, start, orderBy, orderDir, latestEpoch)
	if err != nil {
		logger.Errorf("GetEth1Deposits error retrieving eth1_deposit leaderboard data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	tableData := make([][]interface{}, len(deposits))
	for i, d := range deposits {
		tableData[i] = []interface{}{
			utils.FormatEth1Address(d.FromAddress),
			utils.FormatBalance(d.Amount, currency),
			d.ValidCount,
			d.InvalidCount,
			d.PendingCount,
			d.ActiveCount,
			d.SlashedCount,
			d.VoluntaryExitCount,
			d.TotalCount,
			// utils.FormatPublicKey(d.PublicKey),
			// utils.FormatDepositAmount(d.Amount),
			// utils.FormatEth1TxHash(d.TxHash),
			// utils.FormatTimestamp(d.BlockTs.Unix()),
			// utils.FormatEth1Block(d.BlockNumber),
			// utils.FormatValidatorStatus(d.State),
			// d.ValidSignature,
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
