package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

var validatorsLeaderboardTemplate = template.Must(template.New("validators").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/validators_leaderboard.html"))

// ValidatorsLeaderboard returns the validator-leaderboard using a go template
func ValidatorsLeaderboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "validators", "/validators/leaderboard", "Validator Staking Leaderboard")
	data.HeaderAd = true

	err := validatorsLeaderboardTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
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
		"4": "performance1d",
		"5": "performance7d",
		"6": "performance31d",
		"7": "performance365d",
	}
	orderBy, exists := orderByMap[orderColumn]
	if !exists {
		orderBy = "performance7d"
	}

	orderDir := q.Get("order[0][dir]")
	if orderDir != "desc" && orderDir != "asc" {
		orderDir = "desc"
	}

	var totalCount uint64
	var performanceData []*types.ValidatorPerformance

	if search == "" {
		err = db.DB.Select(&performanceData, `
			SELECT 
				a.*,
				validators.pubkey,
				COALESCE(validator_names.name, '') AS name,
				cnt.total_count
			FROM (
					SELECT
						ROW_NUMBER() OVER (ORDER BY `+orderBy+` DESC) AS rank,
						validator_performance.*
					FROM validator_performance
					ORDER BY `+orderBy+` `+orderDir+`
					LIMIT $1 OFFSET $2
			) AS a
			LEFT JOIN validators ON validators.validatorindex = a.validatorindex
			LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
			LEFT JOIN (SELECT COUNT(*) FROM validator_performance) cnt(total_count) ON true`, length, start)
	} else {
		err = db.DB.Select(&performanceData, `
			WITH 
				matched_validators AS (
					SELECT v.validatorindex, v.pubkey, COALESCE(vn.name,'') as name
					FROM validators v
					LEFT JOIN validator_names vn ON vn.publickey = v.pubkey
					WHERE (pubkeyhex LIKE $3
						OR CAST(v.validatorindex AS text) LIKE $3)
						OR LOWER(vn.name) LIKE LOWER($3)
				)
			SELECT 
				mv.*, 
				perf.rank, perf.balance, perf.performance1d, perf.performance7d, perf.performance31d, perf.performance365d, 
				cnt.total_count
			FROM matched_validators mv
			LEFT JOIN (SELECT COUNT(*) FROM matched_validators) cnt(total_count) ON true
			LEFT JOIN (
				SELECT
					ROW_NUMBER() OVER (ORDER BY `+orderBy+` DESC) AS rank,
					validator_performance.*
				FROM validator_performance
				ORDER BY `+orderBy+` `+orderDir+`
			) perf ON perf.validatorindex = mv.validatorindex
			limit $1 OFFSET $2`, length, start, "%"+search+"%")
	}
	if err != nil {
		logger.Errorf("error retrieving performanceData data (search=%v): %v", search != "", err)
		http.Error(w, "Internal server error", 503)
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
			utils.FormatIncome(b.Performance1d, currency),
			utils.FormatIncome(b.Performance7d, currency),
			utils.FormatIncome(b.Performance31d, currency),
			utils.FormatIncome(b.Performance365d, currency),
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
		http.Error(w, "Internal server error", 503)
		return
	}
}
