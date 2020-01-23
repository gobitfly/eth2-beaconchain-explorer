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
	Length   int64
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
	case "7":
		orderBy = "lastattestedslot"
		if orderDir == "desc" {
			orderDir = "desc nulls last"
		} else {
			orderDir = "asc nulls first"
		}
	default:
		orderBy = "validatorindex"
	}

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
		return nil, err
	}

	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables start parameter from string to int: %v", err)
		return nil, err
	}

	length, err := strconv.ParseInt(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables length parameter from string to int: %v", err)
		return nil, err
	}
	if length < 0 {
		length = 100
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

	err := db.DB.Select(&validators, `SELECT activationepoch, exitepoch FROM validators ORDER BY validatorindex`)

	if err != nil {
		logger.Errorf("error retrieving validators data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	latestEpoch := services.LatestEpoch()
	for _, validator := range validators {
		if latestEpoch > validator.ExitEpoch {
			validatorsPageData.EjectedCount++
		} else if latestEpoch < validator.ActivationEpoch {
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
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
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

	err = db.DB.Get(&totalCount, "SELECT COUNT(*) FROM validators WHERE $1 < activationepoch", services.LatestEpoch())
	if err != nil {
		logger.Errorf("error retrieving pending validator count: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	var validators []*types.ValidatorsPageDataValidators
	err = db.DB.Select(&validators,
		fmt.Sprintf(`SELECT 
				validators.validatorindex, 
				validators.pubkey, 
				validators.withdrawableepoch, 
				validators.effectivebalance, 
				validators.slashed, 
				validators.activationeligibilityepoch, 
				validators.activationepoch, 
				validators.exitepoch,
				
				COALESCE(validator_balances.balance, 0) AS balance
			FROM validators
			LEFT JOIN validator_balances 
				ON validator_balances.epoch = $1
				AND validator_balances.validatorindex = validators.validatorindex
			WHERE $1 < activationepoch
				AND encode(validators.pubkey::bytea, 'hex') LIKE $2
			ORDER BY %s %s 
			LIMIT $3 OFFSET $4`, dataQuery.OrderBy, dataQuery.OrderDir),
		services.LatestEpoch(), "%"+dataQuery.Search+"%", dataQuery.Length, dataQuery.Start)

	if err != nil {
		logger.Errorf("error retrieving pending validator data: %v", err)
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
			[]interface{}{
				fmt.Sprintf("%v", v.ActivationEligibilityEpoch),
				fmt.Sprintf("%v", utils.EpochToTime(v.ActivationEligibilityEpoch).Unix()),
			},
			[]interface{}{
				fmt.Sprintf("%v", v.ActivationEpoch),
				fmt.Sprintf("%v", utils.EpochToTime(v.ActivationEpoch).Unix()),
			},
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
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
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

	err = db.DB.Get(&totalCount, "SELECT COUNT(*) FROM validators WHERE $1 >= activationepoch AND $1 < exitepoch", services.LatestEpoch())
	if err != nil {
		logger.Errorf("error retrieving active validator count: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	var validators []*types.ValidatorsPageDataValidators
	err = db.DB.Select(&validators,
		fmt.Sprintf(`SELECT 
				validators.validatorindex, 
				validators.pubkey, 
				validators.withdrawableepoch, 
				validators.effectivebalance, 
				validators.slashed, 
				validators.activationeligibilityepoch, 
				validators.activationepoch, 
				validators.exitepoch,
				COALESCE(validator_balances.balance, 0) AS balance
			FROM validators
			LEFT JOIN validator_balances 
				ON validator_balances.epoch = $1
				AND validator_balances.validatorindex = validators.validatorindex
			WHERE $1 >= activationepoch 
				AND $1 < exitepoch 
				AND encode(validators.pubkey::bytea, 'hex') LIKE $2
			ORDER BY %s %s 
			LIMIT $3 OFFSET $4`, dataQuery.OrderBy, dataQuery.OrderDir),
		services.LatestEpoch(), "%"+dataQuery.Search+"%", dataQuery.Length, dataQuery.Start)

	if err != nil {
		logger.Errorf("error retrieving active validators data: %v", err)
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
			[]interface{}{
				fmt.Sprintf("%v", v.ActivationEligibilityEpoch),
				fmt.Sprintf("%v", utils.EpochToTime(v.ActivationEligibilityEpoch).Unix()),
			},
			[]interface{}{
				fmt.Sprintf("%v", v.ActivationEpoch),
				fmt.Sprintf("%v", utils.EpochToTime(v.ActivationEpoch).Unix()),
			},
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
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
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

	err = db.DB.Get(&totalCount, "SELECT COUNT(*) FROM validators WHERE $1 >= exitepoch", services.LatestEpoch())
	if err != nil {
		logger.Errorf("error retrieving ejected validator count: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	var validators []*types.ValidatorsPageDataValidators
	err = db.DB.Select(&validators,
		fmt.Sprintf(`SELECT 
				validators.validatorindex, 
				validators.pubkey, 
				validators.withdrawableepoch, 
				validators.effectivebalance, 
				validators.slashed, 
				validators.activationeligibilityepoch, 
				validators.activationepoch, 
				validators.exitepoch,
				COALESCE(validator_balances.balance, 0) AS balance
			FROM validators 
			LEFT JOIN validator_balances 
				ON validator_balances.epoch = $1
				AND validator_balances.validatorindex = validators.validatorindex
			WHERE $1 >= exitepoch
				AND encode(validators.pubkey::bytea, 'hex') LIKE $2
			ORDER BY %s %s
			LIMIT $3 OFFSET $4`, dataQuery.OrderBy, dataQuery.OrderDir),
		services.LatestEpoch(), "%"+dataQuery.Search+"%", dataQuery.Length, dataQuery.Start)

	if err != nil {
		logger.Errorf("error retrieving ejected validators data: %v", err)
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
			[]interface{}{
				fmt.Sprintf("%v", v.ActivationEligibilityEpoch),
				fmt.Sprintf("%v", utils.EpochToTime(v.ActivationEligibilityEpoch).Unix()),
			},
			[]interface{}{
				fmt.Sprintf("%v", v.ActivationEpoch),
				fmt.Sprintf("%v", utils.EpochToTime(v.ActivationEpoch).Unix()),
			},
			[]interface{}{
				fmt.Sprintf("%v", v.ExitEpoch),
				fmt.Sprintf("%v", utils.EpochToTime(v.ExitEpoch).Unix()),
			},
			[]interface{}{
				fmt.Sprintf("%v", v.WithdrawableEpoch),
				fmt.Sprintf("%v", utils.EpochToTime(v.WithdrawableEpoch).Unix()),
			},
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
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

// ValidatorsDataOffline returns the validators that have not attested in the current and previous epochs in json
func ValidatorsDataOffline(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dataQuery, err := parseDataQueryParams(r)
	if err != nil {
		http.Error(w, "Internal server error", 503)
		return
	}

	// we are looking for validators that have not attested in the previous epoch and this epoch
	var totalCount uint64
	err = db.DB.Get(&totalCount, `SELECT COUNT(*) 
		FROM (
			SELECT validatorindex, COUNT(*) AS c 
			FROM attestation_assignments 
			WHERE status = 0 AND (epoch = $1 OR epoch = $2)
			GROUP BY validatorindex
		) a WHERE c = 2`, services.LatestEpoch(), services.LatestEpoch()-1)
	if err != nil {
		logger.Errorf("error retrieving ejected validator count: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	var validators []*types.ValidatorsPageDataValidators
	err = db.DB.Select(&validators,
		fmt.Sprintf(`SELECT 
				validators.validatorindex, 
				validators.pubkey, 
				validators.withdrawableepoch, 
				validators.effectivebalance, 
				validators.slashed, 
				validators.activationeligibilityepoch, 
				validators.activationepoch, 
				validators.exitepoch,
				validator_balances.balance,
				(SELECT MAX(attesterslot) FROM attestation_assignments WHERE validators.validatorindex = validatorindex AND status = 1) as lastattestedslot
			FROM validators
			INNER JOIN (
				SELECT validatorindex FROM (
					SELECT validatorindex, COUNT(*) AS c 
					FROM attestation_assignments 
					WHERE status = 0 AND (epoch = $1 OR epoch = $2)
					GROUP BY validatorindex
				) a WHERE c = 2
			) v ON v.validatorindex = validators.validatorindex
			LEFT JOIN validator_balances 
				ON validator_balances.epoch = $1
				AND validator_balances.validatorindex = validators.validatorindex
			WHERE $1 >= activationepoch 
				AND $1 < exitepoch 
				AND encode(validators.pubkey::bytea, 'hex') LIKE $3
			ORDER BY %s %s
			LIMIT $4 OFFSET $5`, dataQuery.OrderBy, dataQuery.OrderDir),
		services.LatestEpoch(), services.LatestEpoch()-1, "%"+dataQuery.Search+"%", dataQuery.Length, dataQuery.Start)

	if err != nil {
		logger.Errorf("error retrieving offline validators data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	tableData := make([][]interface{}, len(validators))
	for i, v := range validators {
		var lastAttested interface{}
		if v.LastAttestedSlot == nil {
			lastAttested = nil
		} else {
			lastAttested = []interface{}{
				fmt.Sprintf("%v", *v.LastAttestedSlot),
				fmt.Sprintf("%v", utils.SlotToTime(uint64(*v.LastAttestedSlot)).Unix()),
			}
		}
		tableData[i] = []interface{}{
			fmt.Sprintf("%x", v.PublicKey),
			fmt.Sprintf("%v", v.ValidatorIndex),
			utils.FormatBalance(v.CurrentBalance),
			utils.FormatBalance(v.EffectiveBalance),
			fmt.Sprintf("%v", v.Slashed),
			[]interface{}{
				fmt.Sprintf("%v", v.ActivationEligibilityEpoch),
				fmt.Sprintf("%v", utils.EpochToTime(v.ActivationEligibilityEpoch).Unix()),
			},
			[]interface{}{
				fmt.Sprintf("%v", v.ActivationEpoch),
				fmt.Sprintf("%v", utils.EpochToTime(v.ActivationEpoch).Unix()),
			},
			lastAttested,
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
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
