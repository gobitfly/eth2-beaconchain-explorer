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

type ValidatorDataQueryParams struct {
	Search   string
	OrderBy  string
	OrderDir string
	Draw     uint64
	Start    uint64
	Length   uint64
}

func parseDataQueryParams(r *http.Request) (*ValidatorDataQueryParams, error) {
	q := r.URL.Query()

	search := strings.Replace(q.Get("search[value]"), "0x", "", -1)
	if len(search) > 128 {
		search = search[:128]
	}

	orderDir := q.Get("order[0][dir]")
	if orderDir != "desc" && orderDir != "asc" {
		orderDir = "desc"
	}
	orderColumn := q.Get("order[0][column]")
	var orderBy string
	switch orderColumn {
	case "0":
		orderBy = "pubkey"
	case "1":
		orderBy = "validatorindex"
	case "2":
		orderBy = "balance"
	case "3":
		orderBy = "effectivebalance"
	case "4":
		orderBy = "slashed"
	case "5":
		orderBy = "activationeligibilityepoch"
	case "6":
		orderBy = "activationepoch"
	default:
		orderBy = "validatorindex"
	}

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("Error converting datatables data parameter from string to int: %v", err)
		return nil, err
	}

	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Errorf("Error converting datatables start parameter from string to int: %v", err)
		return nil, err
	}

	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("Error converting datatables length parameter from string to int: %v", err)
		return nil, err
	}
	if length > 100 {
		length = 100
	}

	res := &ValidatorDataQueryParams{
		search,
		orderBy,
		orderDir,
		draw,
		start,
		length,
	}

	return res, nil
}

var validatorsTemplate = template.Must(template.New("validators").ParseFiles("templates/layout.html", "templates/validators.html"))

// Validators returns the validators using a go template
func Validators(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	validatorsPageData := types.ValidatorsPageData{}
	var validators []*types.ValidatorsPageDataValidators

	err := db.DB.Select(&validators, `SELECT 
												   epoch, 
												   activationepoch, 
												   exitepoch 
											FROM validator_set 
											WHERE epoch = $1 
											ORDER BY validatorindex`, services.LatestEpoch())

	if err != nil {
		logger.Printf("Error retrieving validators data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	for _, validator := range validators {
		if validator.Epoch > validator.ExitEpoch {
			validatorsPageData.EjectedCount++
		} else if validator.Epoch < validator.ActivationEpoch {
			validatorsPageData.PendingCount++
		} else {
			validatorsPageData.ActiveCount++
		}
	}

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - Validators - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/validators",
		},
		ShowSyncingMessage: services.IsSyncing(),
		Active:             "validators",
		Data:               validatorsPageData,
		Version:            version.Version,
	}

	err = validatorsTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Fatalf("Error executing template for %v route: %v", r.URL.String(), err)
	}
}

// ValidatorsDataPending returns the validators that have data pending in json
func ValidatorsDataPending(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dataQuery, err := parseDataQueryParams(r)
	if err != nil {
		http.Error(w, "Internal server error", 503)
		return
	}

	var totalCount uint64

	err = db.DB.Get(&totalCount, "SELECT COUNT(*) FROM validator_set WHERE epoch = $1 AND epoch < activationepoch", services.LatestEpoch())
	if err != nil {
		logger.Printf("Error retrieving pending validator count: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	var validators []*types.ValidatorsPageDataValidators
	err = db.DB.Select(&validators, fmt.Sprintf(`SELECT 
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
										WHERE validator_set.epoch = $1 AND validator_set.epoch < activationepoch
										ORDER BY %s %s 
										LIMIT $2 OFFSET $3`, dataQuery.OrderBy, dataQuery.OrderDir), services.LatestEpoch(), dataQuery.Length, dataQuery.Start)

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
		}
	}

	data := &types.DataTableResponse{
		Draw:            dataQuery.Draw,
		RecordsTotal:    totalCount,
		RecordsFiltered: totalCount,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Fatalf("Error enconding json response for %v route: %v", r.URL.String(), err)
	}
}

// ValidatorsDataActive will return the validators with active data in json
func ValidatorsDataActive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dataQuery, err := parseDataQueryParams(r)
	if err != nil {
		http.Error(w, "Internal server error", 503)
		return
	}

	var totalCount uint64

	err = db.DB.Get(&totalCount, "SELECT COUNT(*) FROM validator_set WHERE epoch = $1 AND epoch >= activationepoch AND epoch < exitepoch", services.LatestEpoch())
	if err != nil {
		logger.Printf("Error retrieving active validator count: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	var validators []*types.ValidatorsPageDataValidators
	err = db.DB.Select(&validators, fmt.Sprintf(`SELECT 
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
										  AND validator_set.epoch >= activationepoch 
										  AND validator_set.epoch < exitepoch 
										  AND encode(validators.pubkey::bytea, 'hex') LIKE $2
										ORDER BY %s %s 
										LIMIT $3 OFFSET $4`, dataQuery.OrderBy, dataQuery.OrderDir), services.LatestEpoch(), "%"+dataQuery.Search+"%", dataQuery.Length, dataQuery.Start)

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
		Draw:            dataQuery.Draw,
		RecordsTotal:    totalCount,
		RecordsFiltered: totalCount,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Fatalf("Error enconding json response for %v route: %v", r.URL.String(), err)
	}
}

// ValidatorsDataEjected returns the validators that have data ejected in json
func ValidatorsDataEjected(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dataQuery, err := parseDataQueryParams(r)
	if err != nil {
		http.Error(w, "Internal server error", 503)
		return
	}

	var totalCount uint64

	err = db.DB.Get(&totalCount, "SELECT COUNT(*) FROM validator_set WHERE epoch = $1 AND epoch >= exitepoch", services.LatestEpoch())
	if err != nil {
		logger.Printf("Error retrieving ejected validator count: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	var validators []*types.ValidatorsPageDataValidators
	err = db.DB.Select(&validators, fmt.Sprintf(`SELECT 
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
										  AND validator_set.epoch >= exitepoch
										  AND encode(validators.pubkey::bytea, 'hex') LIKE $2
										ORDER BY %s %s 
										LIMIT $3 OFFSET $4`, dataQuery.OrderBy, dataQuery.OrderDir), services.LatestEpoch(), "%"+dataQuery.Search+"%", dataQuery.Length, dataQuery.Start)

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
		Draw:            dataQuery.Draw,
		RecordsTotal:    totalCount,
		RecordsFiltered: totalCount,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Fatalf("Error enconding json response for %v route: %v", r.URL.String(), err)
	}
}
