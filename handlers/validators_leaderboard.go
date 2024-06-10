package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// ValidatorsLeaderboard returns the validator-leaderboard using a go template
func ValidatorsLeaderboard(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "validators_leaderboard.html")
	var validatorsLeaderboardTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "validators", "/validators/leaderboard", "Validator Staking Leaderboard", templateFiles)

	if handleTemplateError(w, r, "validators_leaderboard.go", "ValidatorsLeaderboard", "", validatorsLeaderboardTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

// ValidatorsLeaderboardData returns the leaderboard of validators according to their income in json
func ValidatorsLeaderboardData(w http.ResponseWriter, r *http.Request) {
	currency := GetCurrency(r)

	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	search := strings.Replace(q.Get("search[value]"), "0x", "", -1)
	if len(search) > 128 {
		search = search[:128]
	}
	// var searchIndex *uint64
	// index, err := strconv.ParseUint(search, 10, 64)
	// if err == nil {
	// searchIndex = &index
	// }

	// var searchPubkeyExact *string
	// var searchPubkeyLike *string
	// if searchPubkeyExactRE.MatchString(search) {
	// 	pubkey := strings.ToLower(strings.Replace(search, "0x", "", -1))
	// 	searchPubkeyExact = &pubkey
	// } else if searchPubkeyLikeRE.MatchString(search) {
	// 	pubkey := strings.ToLower(strings.Replace(search, "0x", "", -1))
	// 	searchPubkeyLike = &pubkey
	// }

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
		"4": "cl_performance_1d",
		"5": "cl_performance_7d",
		"6": "cl_performance_31d",
		"7": "cl_performance_365d",
	}
	orderBy, exists := orderByMap[orderColumn]
	if !exists {
		orderBy = "cl_performance_7d"
	}

	orderDir := q.Get("order[0][dir]")
	if orderDir != "desc" && orderDir != "asc" {
		orderDir = "desc"
	}

	var totalCount uint64
	var performanceData []*types.ValidatorPerformance

	// if search == "" {
	err = db.ReaderDb.Select(&performanceData, `
			SELECT 
				a.rank,
				a.balance, 
				a.performance1d, 
				a.performance7d, 
				a.performance31d, 
				a.performance365d, 
				a.rank7d, 
				a.validatorindex,
				validators.pubkey,
				COALESCE(validator_names.name, '') AS name,
				cnt.total_count
			FROM (
					SELECT
						ROW_NUMBER() OVER (ORDER BY `+orderBy+` DESC) AS rank,						
						validator_performance.balance, 
						COALESCE(validator_performance.cl_performance_1d, 0) AS performance1d, 
						COALESCE(validator_performance.cl_performance_7d, 0) AS performance7d, 
						COALESCE(validator_performance.cl_performance_31d, 0) AS performance31d, 
						COALESCE(validator_performance.cl_performance_365d, 0) AS performance365d, 
						COALESCE(validator_performance.cl_performance_total, 0) AS performanceTotal, 
						validator_performance.rank7d, 
						validator_performance.validatorindex
					FROM validator_performance
					ORDER BY `+orderBy+` `+orderDir+`
					LIMIT $1 OFFSET $2
			) AS a
			LEFT JOIN validators ON validators.validatorindex = a.validatorindex
			LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
			LEFT JOIN (SELECT COUNT(*) FROM validator_performance) cnt(total_count) ON true`, length, start)

	if err != nil {
		logger.Errorf("error retrieving performanceData data (search=%v): %v", search != "", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if len(performanceData) > 0 {
		totalCount = performanceData[0].TotalCount
	}

	tableData := make([][]interface{}, len(performanceData))
	for i, b := range performanceData {
		tableData[i] = []interface{}{
			b.Rank,
			utils.FormatValidatorWithName(b.Index, b.Name),
			utils.FormatPublicKey(b.PublicKey),
			fmt.Sprintf("%v", b.Balance),
			utils.FormatClCurrency(b.Performance1d, currency, 5, true, true, true, false),
			utils.FormatClCurrency(b.Performance7d, currency, 5, true, true, true, false),
			utils.FormatClCurrency(b.Performance31d, currency, 5, true, true, true, false),
			utils.FormatClCurrency(b.Performance365d, currency, 5, true, true, true, false),
		}
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    totalCount,
		RecordsFiltered: totalCount,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
