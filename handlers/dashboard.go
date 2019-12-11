package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"

	"strconv"
	"strings"

	"github.com/lib/pq"
)

var dashboardTemplate = template.Must(template.New("dashboard").ParseFiles("templates/layout.html", "templates/dashboard.html"))
var dashboardNotFoundTemplate = template.Must(template.New("dashboardnotfound").ParseFiles("templates/layout.html", "templates/dashboardnotfound.html"))

var filter = pq.Array([]int{1, 2, 3, 4})

func Dashboard(w http.ResponseWriter, r *http.Request) {
	dashboardTemplate = template.Must(template.New("dashboard").ParseFiles("templates/layout.html", "templates/dashboard.html"))
	w.Header().Set("Content-Type", "text/html")
	// validatorsQuery := [...]string{"1", "2", "3", "4", "5"}
	// validators := r.URL.Query().Get("v")
	dashboardPageData := types.DashboardPageData{}

	var err error
	var validators []*types.ValidatorsPageDataValidators

	err = db.DB.Select(&validators, `SELECT 
	epoch, 
	activationepoch, 
	exitepoch 
	FROM validator_set 
	WHERE epoch = $1 and validatorindex in ('1', '2', '3', '4')
	ORDER BY validatorindex`, services.LatestEpoch())

	if err != nil {
		logger.Printf("Error retrieving validators data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	for _, validator := range validators {
		if validator.Epoch > validator.ExitEpoch {
			dashboardPageData.EjectedCount++
		} else if validator.Epoch < validator.ActivationEpoch {
			dashboardPageData.PendingCount++
		} else {
			dashboardPageData.ActiveCount++
		}
	}

	dashboardPageData.Validators = validators

	dashboardPageData.Title = "Hello, World"

	data := &types.PageData{
		Meta: &types.Meta{
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
		},
		ShowSyncingMessage: services.IsSyncing(),
		Active:             "dashboard",
		Data:               nil,
	}

	data.Data = dashboardPageData

	err = dashboardTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Fatalf("Error executing template for %v route: %v", r.URL.String(), err)
	}

}

func DashboardValidatorsDataPending(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Printf("Error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Printf("Error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Printf("Error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	if length > 100 {
		length = 100
	}

	var totalCount uint64

	err = db.DB.Get(&totalCount, "SELECT COUNT(*) FROM validator_set WHERE epoch = $1 AND epoch < activationepoch AND validator_set.validatorindex = ANY($2)", services.LatestEpoch(), filter)
	if err != nil {
		logger.Printf("Error retrieving pending validator count: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	var validators []*types.ValidatorsPageDataValidators
	err = db.DB.Select(&validators, `SELECT 
											 validator_set.epoch,
       										 validator_set.validatorindex, 
											 validators.pubkey, 
											 validator_set.withdrawableepoch, 
											 validator_set.effectivebalance, 
											 validator_set.slashed, 
											 validator_set.activationeligibilityepoch, 
											 validator_set.activationepoch, 
											 validator_set.exitepoch,
       										 validator_balances.balance
										FROM validator_set
										LEFT JOIN validator_balances ON validator_set.epoch = validator_balances.epoch
											AND validator_set.validatorindex = validator_balances.validatorindex
										LEFT JOIN validators ON validator_set.validatorindex = validators.validatorindex
										WHERE validator_set.epoch = $1 
										AND validator_set.epoch < activationepoch
										AND validator_set.validatorindex = ANY($5)
										ORDER BY activationepoch DESC 
										LIMIT $2 OFFSET $3`, services.LatestEpoch(), length, start, filter)

	if err != nil {
		logger.Printf("Error retrieving pending validator data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	tableData := make([][]interface{}, len(validators))
	for i, v := range validators {
		tableData[i] = []interface{}{
			fmt.Sprintf("%x", v.PublicKey),
			fmt.Sprintf("%v", v.ValidatorIndex),
			utils.FormatBalance(v.CurrentBalance),
			utils.FormatBalance(v.EffectiveBalance),
			fmt.Sprintf("%v", v.Slashed),
			fmt.Sprintf("%v", v.ActivationEligibilityEpoch),
			fmt.Sprintf("%v", v.ActivationEpoch),
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
		logger.Fatalf("Error enconding json response for %v route: %v", r.URL.String(), err)
	}
}

func DashboardValidatorsDataActive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	search := strings.Replace(q.Get("search[value]"), "0x", "", -1)
	if len(search) > 128 {
		search = search[:128]
	}

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Printf("Error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Printf("Error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Printf("Error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	if length > 100 {
		length = 100
	}

	var totalCount uint64

	err = db.DB.Get(&totalCount, "SELECT COUNT(*) FROM validator_set WHERE epoch = $1 AND epoch > activationepoch AND epoch < exitepoch AND validator_set.validatorindex = ANY($2)", services.LatestEpoch(), filter)
	if err != nil {
		logger.Printf("Error retrieving active validator count: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	// filter := make([]string, 4)
	// filter = append(filter, "1")
	// filter = append(filter, "2")
	// filter = append(filter, "3")
	// filter = append(filter, "4")

	var validators []*types.ValidatorsPageDataValidators
	err = db.DB.Select(&validators, `SELECT 
											 validator_set.epoch, 
											 validator_set.validatorindex, 
											 validators.pubkey, 
											 validator_set.withdrawableepoch, 
											 validator_set.effectivebalance, 
											 validator_set.slashed, 
											 validator_set.activationeligibilityepoch, 
											 validator_set.activationepoch, 
											 validator_set.exitepoch,
       										 validator_balances.balance
										FROM validator_set
										LEFT JOIN validator_balances ON validator_set.epoch = validator_balances.epoch
											AND validator_set.validatorindex = validator_balances.validatorindex
										LEFT JOIN validators ON validator_set.validatorindex = validators.validatorindex
										WHERE validator_set.epoch = $1 
										  AND validator_set.epoch > activationepoch 
										  AND validator_set.epoch < exitepoch 
											AND encode(validators.pubkey::bytea, 'hex') LIKE $2
											AND validator_set.validatorindex = ANY($5)
										ORDER BY activationepoch DESC 
										LIMIT $3 OFFSET $4`, services.LatestEpoch(), "%"+search+"%", length, start, filter)

	if err != nil {
		logger.Printf("Error retrieving active validators data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	tableData := make([][]interface{}, len(validators))
	for i, v := range validators {
		tableData[i] = []interface{}{
			fmt.Sprintf("%x", v.PublicKey),
			fmt.Sprintf("%v", v.ValidatorIndex),
			utils.FormatBalance(v.CurrentBalance),
			utils.FormatBalance(v.EffectiveBalance),
			fmt.Sprintf("%v", v.Slashed),
			fmt.Sprintf("%v", v.ActivationEligibilityEpoch),
			fmt.Sprintf("%v", v.ActivationEpoch),
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
		logger.Fatalf("Error enconding json response for %v route: %v", r.URL.String(), err)
	}
}

func DashboardValidatorsDataEjected(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	search := strings.Replace(q.Get("search[value]"), "0x", "", -1)
	if len(search) > 128 {
		search = search[:128]
	}

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Printf("Error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Printf("Error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Printf("Error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	if length > 100 {
		length = 100
	}

	var totalCount uint64

	err = db.DB.Get(&totalCount, "SELECT COUNT(*) FROM validator_set WHERE epoch = $1 AND epoch > exitepoch AND validator_set.validatorindex = ANY($2)", services.LatestEpoch(), filter)
	if err != nil {
		logger.Printf("Error retrieving ejected validator count: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	var validators []*types.ValidatorsPageDataValidators
	err = db.DB.Select(&validators, `SELECT 
											 validator_set.epoch,
       										 validator_set.validatorindex, 
											 validators.pubkey, 
											 validator_set.withdrawableepoch, 
											 validator_set.effectivebalance, 
											 validator_set.slashed, 
											 validator_set.activationeligibilityepoch, 
											 validator_set.activationepoch, 
											 validator_set.exitepoch,
       										 validator_balances.balance
										FROM validator_set 
										LEFT JOIN validator_balances ON validator_set.epoch = validator_balances.epoch
											AND validator_set.validatorindex = validator_balances.validatorindex
										LEFT JOIN validators ON validator_set.validatorindex = validators.validatorindex
										WHERE validator_set.epoch = $1 
										  AND validator_set.epoch > exitepoch
											AND encode(validators.pubkey::bytea, 'hex') LIKE $2
											AND validator_set.validatorindex = ANY($5)
										ORDER BY activationepoch DESC 
										LIMIT $3 OFFSET $4`, services.LatestEpoch(), "%"+search+"%", length, start, filter)

	if err != nil {
		logger.Printf("Error retrieving ejected validators data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	tableData := make([][]interface{}, len(validators))
	for i, v := range validators {
		tableData[i] = []interface{}{
			fmt.Sprintf("%x", v.PublicKey),
			fmt.Sprintf("%v", v.ValidatorIndex),
			utils.FormatBalance(v.CurrentBalance),
			utils.FormatBalance(v.EffectiveBalance),
			fmt.Sprintf("%v", v.Slashed),
			fmt.Sprintf("%v", v.ActivationEligibilityEpoch),
			fmt.Sprintf("%v", v.ActivationEpoch),
			fmt.Sprintf("%v", v.ExitEpoch),
			fmt.Sprintf("%v", v.WithdrawableEpoch),
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
		logger.Fatalf("Error enconding json response for %v route: %v", r.URL.String(), err)
	}
}
