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

// ValidatorsStreaksLeaderboard returns the attestation-streak-leaderboard using a go template
func ValidatorsStreakLeaderboard(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "validators_streakleaderboard.html")
	var validatorsStreakLeaderboardTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "validators", "/validators/streaksleaderboard", "Validator Streaks Leaderboard", templateFiles)

	if handleTemplateError(w, r, "validators_streakLeaderboard.go", "ValidatorsStreakLeaderboard", "", validatorsStreakLeaderboardTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

// ValidatorsStreakLeaderboardData returns the leaderboard of attestation-streaks
func ValidatorsStreakLeaderboardData(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	search := strings.Replace(q.Get("search[value]"), "0x", "", -1)
	if len(search) > 128 {
		search = search[:128]
	}

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
		"1": "cs.rank",
		"4": "ls.rank",
	}
	orderBy, exists := orderByMap[orderColumn]
	if !exists {
		orderBy = "ls.rank"
	}

	orderDir := q.Get("order[0][dir]")
	if orderDir != "desc" && orderDir != "asc" {
		orderDir = "asc"
	}

	var totalCount uint64

	var sqlData []struct {
		Totalcount     uint64
		Validatorindex uint64
		Name           string
		Lrank          int
		Lstart         uint64
		Llength        int
		Crank          int
		Cstart         uint64
		Clength        int
	}

	if search == "" {
		err = db.ReaderDb.Select(&sqlData, `
			with
				longeststreaks as (
					select validatorindex, start, length, rank() over(order by length desc)
					from validator_attestation_streaks where longest = 't' and status = 1
				),
				currentstreaks as (
					select validatorindex, start, length, rank() over(order by length desc)
					from validator_attestation_streaks where current = 't' and status = 1
				)
			select 
				v.validatorindex,
				coalesce(vn.name, '') as name,
				cnt.totalcount,
				coalesce(ls.rank, 0) lrank,
				coalesce(ls.start, 0) lstart,
				coalesce(ls.length, 0) llength,
				coalesce(cs.rank, 0) crank,
				coalesce(cs.start, 0) cstart,
				coalesce(cs.length, 0) clength
			from longeststreaks ls
			inner join validators v on ls.validatorindex = v.validatorindex
			left join currentstreaks cs on cs.validatorindex = v.validatorindex
			left join validator_names vn on v.pubkey = vn.publickey
			left join (select count(*) from longeststreaks) cnt(totalcount) on true
			order by `+orderBy+` `+orderDir+` limit $1 offset $2`, length, start)
	} else {
		err = db.ReaderDb.Select(&sqlData, `
			with 
				matched_validators as (
					select v.validatorindex, v.pubkey, coalesce(vn.name,'') as name
					from validators v
					left join validator_names vn ON vn.publickey = v.pubkey
					where (pubkeyhex like $4
						or cast(v.validatorindex as text) like $3)
						or vn.name ilike $3
				),
				longeststreaks as (
					select validatorindex, start, length, rank() over(order by length desc)
					from validator_attestation_streaks where longest = 't' and status = 1
				),
				currentstreaks as (
					select validatorindex, start, length, rank() over(order by length desc)
					from validator_attestation_streaks where current = 't' and status = 1
				)
			select 
				v.validatorindex,
				coalesce(vn.name, '') as name,
				cnt.totalcount,
				coalesce(ls.rank, 0) lrank,
				coalesce(ls.start, 0) lstart,
				coalesce(ls.length, 0) llength,
				coalesce(cs.rank, 0) crank,
				coalesce(cs.start, 0) cstart,
				coalesce(cs.length, 0) clength
			from longeststreaks ls
			inner join matched_validators mv on ls.validatorindex = mv.validatorindex
			inner join validators v on ls.validatorindex = v.validatorindex
			left join currentstreaks cs on cs.validatorindex = v.validatorindex
			left join validator_names vn on v.pubkey = vn.publickey
			left join (select count(*) from matched_validators) cnt(totalcount) on true
			order by `+orderBy+` `+orderDir+` limit $1 offset $2`, length, start, "%"+search+"%", search+"%")
	}
	if err != nil {
		logger.Errorf("error retrieving streaksData data (search=%v): %v", search != "", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	if len(sqlData) > 0 {
		totalCount = sqlData[0].Totalcount
	}

	tableData := make([][]interface{}, len(sqlData))
	for i, d := range sqlData {
		tableData[i] = []interface{}{
			utils.FormatValidatorWithName(d.Validatorindex, d.Name),
			fmt.Sprintf("%v", d.Crank),
			utils.FormatEpoch(d.Cstart),
			fmt.Sprintf("%v", d.Clength),
			fmt.Sprintf("%v", d.Lrank),
			utils.FormatEpoch(d.Lstart),
			fmt.Sprintf("%v", d.Llength),
		}
		// current streak is missed
		if d.Crank == 0 {
			tableData[i][1] = `<span data-toggle="tooltip" title="Last attestation is missed">-</span>`
			tableData[i][2] = `<span data-toggle="tooltip" title="Last attestation is missed">-</span>`
			tableData[i][3] = `<span data-toggle="tooltip" title="Last attestation is missed">-</span>`
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
