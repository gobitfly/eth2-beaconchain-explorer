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
	var validatorsLeaderboardTemplate = templates.GetTemplate("layout.html", "validators_leaderboard.html")

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "validators", "/validators/leaderboard", "Validator Staking Leaderboard")
	data.HeaderAd = true

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
						validator_performance.performance1d, 
						validator_performance.performance7d, 
						validator_performance.performance31d, 
						validator_performance.performance365d, 
						validator_performance.rank7d, 
						validator_performance.validatorindex
					FROM validator_performance
					ORDER BY `+orderBy+` `+orderDir+`
					LIMIT $1 OFFSET $2
			) AS a
			LEFT JOIN validators ON validators.validatorindex = a.validatorindex
			LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
			LEFT JOIN (SELECT COUNT(*) FROM validator_performance) cnt(total_count) ON true`, length, start)
	// } else {
	// 	// for performance-reasons we combine multiple search results with `union`
	// 	args := []interface{}{}
	// 	args = append(args, "%"+strings.ToLower(search)+"%")
	// 	searchQry := fmt.Sprintf(`SELECT publickey AS pubkey FROM validator_names WHERE LOWER(name) LIKE $%d `, len(args))
	// 	if searchIndex != nil {
	// 		args = append(args, *searchIndex)
	// 		searchQry += fmt.Sprintf(`UNION SELECT pubkey FROM validators WHERE validatorindex = $%d `, len(args))
	// 	}

	// 	if searchPubkeyExact != nil {
	// 		args = append(args, *searchPubkeyExact)
	// 		searchQry += fmt.Sprintf(`UNION SELECT pubkey FROM validators WHERE pubkeyhex = $%d `, len(args))
	// 	} else if searchPubkeyLike != nil {
	// 		args = append(args, *searchPubkeyLike+"%")
	// 		searchQry += fmt.Sprintf(`UNION SELECT pubkey FROM validators WHERE pubkeyhex LIKE $%d `, len(args))
	// 	}

	// 	args = append(args, length)
	// 	args = append(args, start)
	// 	qry := fmt.Sprintf(`
	// 		WITH matched_validators AS (%v)
	// 		SELECT
	// 			v.validatorindex, mv.pubkey, COALESCE(vn.name, '') as name,
	// 			perf.rank, perf.balance, perf.performance1d, perf.performance7d, perf.performance31d, perf.performance365d,
	// 			cnt.total_count
	// 		FROM matched_validators mv
	// 		INNER JOIN validators v ON v.pubkey = mv.pubkey
	// 		LEFT JOIN validator_names vn ON vn.publickey = mv.pubkey
	// 		LEFT JOIN (SELECT COUNT(*) FROM matched_validators) cnt(total_count) ON true
	// 		LEFT JOIN (
	// 			SELECT ROW_NUMBER() OVER (ORDER BY `+orderBy+` DESC) AS rank, validator_performance.*
	// 			FROM validator_performance
	// 			ORDER BY `+orderBy+` `+orderDir+`
	// 		) perf ON perf.validatorindex = v.validatorindex
	// 		LIMIT $%d OFFSET $%d`, searchQry, len(args)-1, len(args))
	// 	err = db.ReaderDb.Select(&performanceData, qry, args...)
	// }
	if err != nil {
		logger.Errorf("error retrieving performanceData data (search=%v): %v", search != "", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
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
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}
