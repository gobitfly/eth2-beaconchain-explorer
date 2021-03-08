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

var validatorsStreaksLeaderboardTemplate = template.Must(template.New("validators").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/validators_streaksleaderboard.html"))

// ValidatorsLeaderboard returns the validator-leaderboard using a go template
func ValidatorsStreaksLeaderboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "validators", "/validators/streaksleaderboard", "Validator Streaks Leaderboard")
	data.HeaderAd = true

	// #TODO:patrick remove this line
	validatorsStreaksLeaderboardTemplate = template.Must(template.New("validators").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/validators_streaksleaderboard.html"))
	err := validatorsStreaksLeaderboardTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

// ValidatorsLeaderboardData returns the leaderboard of validators according to their income in json
func ValidatorsStreaksLeaderboardData(w http.ResponseWriter, r *http.Request) {

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
		"1": "lrank",
		"4": "crank",
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

	var sqlData []struct {
		Totalcount     int
		Validatorindex uint64
		Name           string
		Lrank          int
		Lstart         int
		Llength        int
		Crank          int
		Cstart         int
		Clength        int
	}

	if search == "" {
		err = db.DB.Select(&sqlData, `
			with 
				longeststreaks as (
					select validatorindex, start, length, rank() over (order by length)
					from validator_attestation_streaks
					where status = 1
				),
				currentstreaks as (
					select validatorindex, start, length, rank() over (order by length)
					from validator_attestation_streaks
					where status = 1 and start+length = (select max(start+length) from validator_attestation_streaks)
				)
			select 
				COALESCE(validator_names.name, '') AS name,
				cnt.totalcount,
				ls.rank lrank, 
				ls.start lstart, 
				ls.length llength, 
				cs.rank crank, 
				cs.start cstart, 
				cs.length clength, 
			from higheststreaks ls
			left join validators on higheststreaks.validatorindex = validators.validatorindex
			LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
			LEFT JOIN (SELECT COUNT(*) FROM higheststreaks) cnt(totlacount) ON true
			left join currentstreaks cs on cs.valdiatorindex = hs.validatorindex
			order by `+orderBy+` `+orderDir+`limit $1 offset $2`, length, start)
		if err != nil {
			logger.Errorf("error retrieving streaksData data without search: %v", err)
			http.Error(w, "Internal server error", 503)
			return
		}
	} else {
		http.Error(w, "not implemented yet", 503)
		return
	}

	tableData := make([][]interface{}, len(sqlData))
	for i, d := range sqlData {
		tableData[i] = []interface{}{
			utils.FormatValidatorWithName(d.Validatorindex, d.Name),
			fmt.Sprintf("%v", d.Lrank),
			fmt.Sprintf("%v", d.Lstart),
			fmt.Sprintf("%v", d.Llength),
			fmt.Sprintf("%v", d.Crank),
			fmt.Sprintf("%v", d.Cstart),
			fmt.Sprintf("%v", d.Clength),
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
