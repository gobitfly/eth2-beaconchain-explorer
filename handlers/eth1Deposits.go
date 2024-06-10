package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"net/http"
	"strconv"
	"strings"
)

// Deposits will return information about deposits using a go template
func Deposits(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "deposits.html", "index/depositChart.html")
	var DepositsTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	pageData := &types.DepositsPageData{}

	latestChartsPageData := services.LatestChartsPageData()
	if len(latestChartsPageData) != 0 {
		for _, c := range latestChartsPageData {
			if c.Path == "deposits" {
				pageData.DepositChart = c
				break
			}
		}
	}

	pageData.Stats = services.GetLatestStats()
	pageData.DepositContract = utils.Config.Chain.ClConfig.DepositContractAddress

	data := InitPageData(w, r, "blockchain", "/deposits", "Deposits", templateFiles)
	data.Data = pageData

	if handleTemplateError(w, r, "eth1Depostis.go", "Deposits", "", DepositsTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

func Eth1Deposits(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/validators/deposits", http.StatusMovedPermanently)
}

// Eth1DepositsData will return eth1-deposits as json
func Eth1DepositsData(w http.ResponseWriter, r *http.Request) {
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
		"0": "from_address",
		"1": "publickey",
		"2": "withdrawal_credential",
		"3": "amount",
		"4": "tx_hash",
		"5": "block_ts",
		"6": "block_number",
		"7": "state",
		"8": "valid_signature",
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
			utils.FormatWithdawalCredentials(d.WithdrawalCredentials, true),
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// Eth1Deposits will return information about deposits using a go template
func Eth1DepositsLeaderboard(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "eth1DepositsLeaderboard.html")
	var eth1DepositsLeaderboardTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "eth1Deposits", "/deposits/eth1", "Initiated Deposits", templateFiles)

	data.Data = types.EthOneDepositLeaderBoardPageData{
		DepositContract: utils.Config.Chain.ClConfig.DepositContractAddress,
	}

	if handleTemplateError(w, r, "eth1Deposits.go", "Eth1DepositsLeaderboard", "", eth1DepositsLeaderboardTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

// Eth1DepositsData will return eth1-deposits as json
func Eth1DepositsLeaderboardData(w http.ResponseWriter, r *http.Request) {
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

	deposits, depositCount, err := db.GetEth1DepositsLeaderboard(search, length, start, orderBy, orderDir)
	if err != nil {
		logger.Errorf("GetEth1Deposits error retrieving eth1_deposit leaderboard data: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
