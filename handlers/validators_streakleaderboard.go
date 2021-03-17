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

var validatorsStreakLeaderboardTemplate = template.Must(template.New("validators").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/validators_streakleaderboard.html"))

// ValidatorsStreaksLeaderboard returns the attestation-streak-leaderboard using a go template
func ValidatorsStreakLeaderboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "validators", "/validators/streaksleaderboard", "Validator Streaks Leaderboard")
	data.HeaderAd = true

	err := validatorsStreakLeaderboardTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
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
		err = db.DB.Select(&sqlData, `
			with 
				longeststreaks as (
					select 
						validatorindex, start, length, rank() over (order by length desc),
						rank() over (partition by validatorindex order by length desc) as vrank
					from validator_attestation_streaks
					where status = 1
				),
				currentstreaks as (
					select validatorindex, start, length, rank() over (order by length desc)
					from validator_attestation_streaks
					where status = 1 and start+length = (select max(start+length) from validator_attestation_streaks)
				)
			select 
				ls.validatorindex,
				COALESCE(validator_names.name, '') AS name,
				cnt.totalcount,
				ls.rank lrank, 
				ls.start lstart, 
				ls.length llength, 
				coalesce(cs.rank,0) crank, 
				coalesce(cs.start,0) cstart, 
				coalesce(cs.length,0) clength 
			from longeststreaks ls
			inner join validators v on ls.validatorindex = v.validatorindex
			left join validator_names on v.pubkey = validator_names.publickey
			left join (select count(*) from longeststreaks) cnt(totalcount) on true
			left join currentstreaks cs on cs.validatorindex = ls.validatorindex
			where vrank = 1
			order by `+orderBy+` `+orderDir+` limit $1 offset $2`, length, start)
	} else {
		err = db.DB.Select(&sqlData, `
			with 
				matched_validators as (
					SELECT v.validatorindex, v.pubkey, COALESCE(vn.name,'') as name
					FROM validators v
					LEFT JOIN validator_names vn ON vn.publickey = v.pubkey
					WHERE (pubkeyhex LIKE $3
						OR CAST(v.validatorindex AS text) LIKE $3)
						OR LOWER(vn.name) LIKE LOWER($3)
				),
				longeststreaks as (
					select 
						validatorindex, start, length, rank() over (order by length desc),
						rank() over (partition by validatorindex order by length desc) as vrank
					from validator_attestation_streaks
					where status = 1
				),
				currentstreaks as (
					select validatorindex, start, length, rank() over (order by length desc)
					from validator_attestation_streaks
					where status = 1 and start+length = (select max(start+length) from validator_attestation_streaks)
				)
			select 
				ls.validatorindex,
				v.name,
				cnt.totalcount,
				ls.rank lrank, 
				ls.start lstart, 
				ls.length llength, 
				coalesce(cs.rank,0) crank, 
				coalesce(cs.start,0) cstart, 
				coalesce(cs.length,0) clength 
			from longeststreaks ls
			inner join matched_validators v on ls.validatorindex = v.validatorindex
			left join (select count(*) from longeststreaks) cnt(totalcount) on true
			left join currentstreaks cs on cs.validatorindex = ls.validatorindex
			where vrank = 1
			order by `+orderBy+` `+orderDir+` limit $1 offset $2`, length, start, "%"+search+"%")
	}
	if err != nil {
		logger.Errorf("error retrieving streaksData data (search=%v): %v", search != "", err)
		http.Error(w, "Internal server error", 503)
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
		http.Error(w, "Internal server error", 503)
		return
	}
}
