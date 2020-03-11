package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var validatorsLeaderboardTemplate = template.Must(template.New("validators").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/validators_leaderboard.html"))

// Validators returns the validators using a go template
func ValidatorsLeaderboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - Validator Staking Leaderboard - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/validators/leaderboard",
		},
		ShowSyncingMessage:    services.IsSyncing(),
		Active:                "validators",
		Data:                  nil,
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
	}

	err := validatorsLeaderboardTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

// ValidatorAttestations returns a validators attestations in json
func ValidatorsLeaderboardData(w http.ResponseWriter, r *http.Request) {
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

	err = db.DB.Get(&totalCount, "SELECT COUNT(*) FROM validator_performance")
	if err != nil {
		logger.Errorf("error retrieving proposed blocks count: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	var performanceData []*types.ValidatorPerformance
	err = db.DB.Select(&performanceData, `SELECT * FROM (SELECT 
											   			ROW_NUMBER() OVER (ORDER BY `+orderBy+` DESC) AS rank,
											   			validator_performance.*,
											   			validators.pubkey 
													FROM validator_performance 
													LEFT JOIN validators ON validators.validatorindex = validator_performance.validatorindex
													ORDER BY `+orderBy+` `+orderDir+`) AS a
													WHERE (encode(a.pubkey::bytea, 'hex') LIKE $3
														OR CAST(a.validatorindex AS text) LIKE $3)
													LIMIT $1 OFFSET $2`, length, start, "%"+search+"%")

	if err != nil {
		logger.Errorf("error retrieving validator attestations data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	tableData := make([][]interface{}, len(performanceData))
	for i, b := range performanceData {

		tableData[i] = []interface{}{
			b.Rank,
			utils.FormatValidator(b.Index),
			fmt.Sprintf("%x", b.PublicKey),
			fmt.Sprintf("%v", b.Balance),
			utils.FormatIncome(b.Performance1d),
			utils.FormatIncome(b.Performance7d),
			utils.FormatIncome(b.Performance31d),
			utils.FormatIncome(b.Performance365d),
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
