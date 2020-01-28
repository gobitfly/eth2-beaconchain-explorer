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

type ValidatorsDataQueryParams struct {
	Search   string
	OrderBy  string
	OrderDir string
	Draw     uint64
	Start    uint64
	Length   int64
}

func parseValidatorsDataQueryParams(r *http.Request) (*ValidatorsDataQueryParams, error) {
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

	orderByRouteMap := make(map[string]map[string]string)
	orderByRouteMap["/validators/data/pending"] = map[string]string{"0": "pubkey", "1": "validatorindex", "2": "balance", "3": "effectivebalance", "4": "activationeligibilityepoch", "5": "activationepoch"}
	orderByRouteMap["/validators/data/active"] = map[string]string{"0": "pubkey", "1": "validatorindex", "2": "balance", "3": "effectivebalance", "4": "slashed", "5": "activationepoch"}
	orderByRouteMap["/validators/data/ejected"] = map[string]string{"0": "pubkey", "1": "validatorindex", "2": "balance", "3": "effectivebalance", "4": "slashed", "5": "activationepoch", "6": "exitepoch", "7": "withdrawableepoch"}
	orderByRouteMap["/validators/data/offline"] = map[string]string{"0": "pubkey", "1": "validatorindex", "2": "balance", "3": "effectivebalance", "4": "slashed", "5": "activationepoch", "6": "lastattestationslot"}

	orderByRoute, exists := orderByRouteMap[r.URL.Path]
	if !exists {
		return nil, fmt.Errorf("invalid request-path")
	}

	orderBy, exists := orderByRoute[orderColumn]
	if !exists {
		orderBy = "validatorindex"
	} else if orderBy == "lastattestationslot" {
		if orderDir == "desc" {
			orderDir = "desc nulls last"
		} else {
			orderDir = "asc nulls first"
		}
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

	res := &ValidatorsDataQueryParams{
		search,
		orderBy,
		orderDir,
		draw,
		start,
		length,
	}

	return res, nil
}

func prepareValidatorsDataTable(validators *[]*types.ValidatorsPageDataValidators) [][]interface{} {
	tableData := make([][]interface{}, len(*validators))
	for i, v := range *validators {
		var lastAttestation interface{}
		if v.LastAttestationSlot == nil {
			lastAttestation = nil
		} else {
			lastAttestation = []interface{}{
				fmt.Sprintf("%v", *v.LastAttestationSlot),
				fmt.Sprintf("%v", utils.SlotToTime(uint64(*v.LastAttestationSlot)).Unix()),
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
			[]interface{}{
				fmt.Sprintf("%v", v.ExitEpoch),
				fmt.Sprintf("%v", utils.EpochToTime(v.ExitEpoch).Unix()),
			},
			[]interface{}{
				fmt.Sprintf("%v", v.WithdrawableEpoch),
				fmt.Sprintf("%v", utils.EpochToTime(v.WithdrawableEpoch).Unix()),
			},
			lastAttestation,
		}
	}
	return tableData
}

var validatorsTemplate = template.Must(template.New("validators").ParseFiles("templates/layout.html", "templates/validators.html"))

// Validators returns the validators using a go template
func Validators(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	validatorsPageData := types.ValidatorsPageData{}
	var validators []*types.ValidatorsPageDataValidators

	err := db.DB.Select(&validators, `SELECT activationepoch, exitepoch, lastattestationslot FROM validators ORDER BY validatorindex`)

	if err != nil {
		logger.Errorf("error retrieving validators data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	latestEpoch := services.LatestEpoch()

	var firstSlotOfPreviousEpoch uint64
	if services.LatestEpoch() < 1 {
		firstSlotOfPreviousEpoch = 0
	} else {
		firstSlotOfPreviousEpoch = (services.LatestEpoch() - 1) * utils.Config.Chain.SlotsPerEpoch
	}

	for _, validator := range validators {
		if latestEpoch > validator.ExitEpoch {
			validatorsPageData.EjectedCount++
		} else if latestEpoch < validator.ActivationEpoch {
			validatorsPageData.PendingCount++
		} else {
			validatorsPageData.ActiveCount++
			// offline validators did not attest in the last 2 epochs (and are active for >1 epochs)
			if validator.ActivationEpoch < latestEpoch && (validator.LastAttestationSlot == nil || uint64(*validator.LastAttestationSlot) < firstSlotOfPreviousEpoch) {
				validatorsPageData.OfflineCount++
			}
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

	dataQuery, err := parseValidatorsDataQueryParams(r)
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
				validators.activationepoch,
				validators.exitepoch,
				validators.lastattestationslot,
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

	data := &types.DataTableResponse{
		Draw:            dataQuery.Draw,
		RecordsTotal:    totalCount,
		RecordsFiltered: totalCount,
		Data:            prepareValidatorsDataTable(&validators),
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

	dataQuery, err := parseValidatorsDataQueryParams(r)
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
				validators.activationepoch,
				validators.exitepoch,
				validators.lastattestationslot,
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

	data := &types.DataTableResponse{
		Draw:            dataQuery.Draw,
		RecordsTotal:    totalCount,
		RecordsFiltered: totalCount,
		Data:            prepareValidatorsDataTable(&validators),
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

	dataQuery, err := parseValidatorsDataQueryParams(r)
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
				validators.activationepoch,
				validators.exitepoch,
				validators.lastattestationslot,
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

	data := &types.DataTableResponse{
		Draw:            dataQuery.Draw,
		RecordsTotal:    totalCount,
		RecordsFiltered: totalCount,
		Data:            prepareValidatorsDataTable(&validators),
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

// ValidatorsDataOffline returns the validators that have not attested in the last 2 epochs
func ValidatorsDataOffline(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dataQuery, err := parseValidatorsDataQueryParams(r)
	if err != nil {
		http.Error(w, "Internal server error", 503)
		return
	}

	var firstSlotOfPreviousEpoch uint64
	if services.LatestEpoch() < 1 {
		firstSlotOfPreviousEpoch = 0
	} else {
		firstSlotOfPreviousEpoch = (services.LatestEpoch() - 1) * utils.Config.Chain.SlotsPerEpoch
	}

	var totalCount uint64
	err = db.DB.Get(&totalCount,
		`SELECT COUNT(*) FROM validators 
		WHERE $1 > activationepoch 
			AND $1 < exitepoch 
			AND (lastattestationslot < $2 OR lastattestationslot is null)`,
		services.LatestEpoch(), firstSlotOfPreviousEpoch)
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
				validators.activationepoch,
				validators.exitepoch,
				validators.lastattestationslot,
				COALESCE(validator_balances.balance, 0) AS balance
			FROM validators
			LEFT JOIN validator_balances
				ON validator_balances.epoch = $1
				AND validator_balances.validatorindex = validators.validatorindex
			WHERE $1 > activationepoch
				AND $1 < exitepoch
				AND encode(validators.pubkey::bytea, 'hex') LIKE $3
				AND ( lastattestationslot < $2 OR lastattestationslot is null )
			ORDER BY %s %s
			LIMIT $4 OFFSET $5`, dataQuery.OrderBy, dataQuery.OrderDir),
		services.LatestEpoch(), firstSlotOfPreviousEpoch, "%"+dataQuery.Search+"%", dataQuery.Length, dataQuery.Start)

	if err != nil {
		logger.Errorf("error retrieving offline validators data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	data := &types.DataTableResponse{
		Draw:            dataQuery.Draw,
		RecordsTotal:    totalCount,
		RecordsFiltered: totalCount,
		Data:            prepareValidatorsDataTable(&validators),
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
