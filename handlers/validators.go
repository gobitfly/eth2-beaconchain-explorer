package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/price"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

var validatorsTemplate = template.Must(template.New("validators").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/validators.html"))

// Validators returns the validators using a go template
func Validators(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	validatorsPageData := types.ValidatorsPageData{}
	var validators []*types.ValidatorsPageDataValidators

	err := db.DB.Select(&validators, `SELECT activationepoch, exitepoch, lastattestationslot, slashed FROM validators ORDER BY validatorindex`)

	if err != nil {
		logger.Errorf("error retrieving validators data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	latestEpoch := services.LatestEpoch()

	validatorOnlineThresholdSlot := GetValidatorOnlineThresholdSlot()
	for _, validator := range validators {
		validatorsPageData.TotalCount++
		if latestEpoch > validator.ExitEpoch {
			validatorsPageData.ExitedCount++
		} else if latestEpoch < validator.ActivationEpoch {
			validatorsPageData.PendingCount++
		} else if validator.Slashed {
			// offline validators did not attest in the last 2 epochs (and are active for >1 epochs)
			if validator.ActivationEpoch < latestEpoch && (validator.LastAttestationSlot == nil || uint64(*validator.LastAttestationSlot) < validatorOnlineThresholdSlot) {
				validatorsPageData.SlashingOfflineCount++
			} else {
				validatorsPageData.SlashingOnlineCount++
			}
		} else {
			// offline validators did not attest in the last 2 epochs (and are active for >1 epochs)
			if validator.ActivationEpoch < latestEpoch && (validator.LastAttestationSlot == nil || uint64(*validator.LastAttestationSlot) < validatorOnlineThresholdSlot) {
				validatorsPageData.ActiveOfflineCount++
			} else {
				validatorsPageData.ActiveOnlineCount++
			}
		}
	}
	validatorsPageData.ActiveCount = validatorsPageData.ActiveOnlineCount + validatorsPageData.ActiveOfflineCount
	validatorsPageData.SlashingCount = validatorsPageData.SlashingOnlineCount + validatorsPageData.SlashingOfflineCount
	validatorsPageData.ExitingCount = validatorsPageData.ExitingOnlineCount + validatorsPageData.ExitingOfflineCount

	data := InitPageData(w, r, "validators", "/validators", "Validators")
	data.HeaderAd = true
	data.Data = validatorsPageData

	err = validatorsTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

type ValidatorsDataQueryParams struct {
	Search      string
	OrderBy     string
	OrderDir    string
	Draw        uint64
	Start       uint64
	Length      int64
	StateFilter string
}

func parseValidatorsDataQueryParams(r *http.Request) (*ValidatorsDataQueryParams, error) {
	q := r.URL.Query()

	search := strings.Replace(q.Get("search[value]"), "0x", "", -1)
	if len(search) > 128 {
		search = search[:128]
	}

	filterByState := q.Get("filterByState")
	var qryStateFilter string
	switch filterByState {
	case "pending":
		qryStateFilter = "AND validators.status LIKE 'pending%'"
	case "active":
		qryStateFilter = "AND validators.status LIKE 'active%'"
	case "active_online":
		qryStateFilter = "AND validators.status = 'active_online'"
	case "active_offline":
		qryStateFilter = "AND validators.status = 'active_offline'"
	case "slashing":
		qryStateFilter = "AND validators.status LIKE 'slashing%'"
	case "slashing_online":
		qryStateFilter = "AND validators.status = 'slashing_online'"
	case "slashing_offline":
		qryStateFilter = "AND validators.status = 'slashing_offline'"
	case "slashed":
		qryStateFilter = "AND validators.status = 'slashed'"
	case "exiting":
		qryStateFilter = "AND validators.status LIKE 'exiting%'"
	case "exiting_online":
		qryStateFilter = "AND validators.status = 'exiting_online'"
	case "exiting_offline":
		qryStateFilter = "AND validators.status = 'exiting_offline'"
	case "exited":
		qryStateFilter = "AND (validators.status = 'exited' OR validators.status = 'slashed')"
	default:
		qryStateFilter = ""
	}

	orderColumn := q.Get("order[0][column]")
	orderByMap := map[string]string{
		"0": "pubkey",
		"1": "validatorindex",
		"2": "balance",
		"3": "state",
		"4": "activationepoch",
		"5": "exitepoch",
		"6": "withdrawableepoch",
		"7": "lastattestationslot",
		"8": "slashed",
	}
	orderBy, exists := orderByMap[orderColumn]
	if !exists {
		orderBy = "validatorindex"
	}

	orderDir := q.Get("order[0][dir]")
	if orderDir != "desc" && orderDir != "asc" {
		orderDir = "desc"
	}

	if orderBy == "lastattestationslot" {
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
		qryStateFilter,
	}

	return res, nil
}

// ValidatorsData returns all validators and their balances
func ValidatorsData(w http.ResponseWriter, r *http.Request) {
	currency := GetCurrency(r)

	w.Header().Set("Content-Type", "application/json")

	dataQuery, err := parseValidatorsDataQueryParams(r)
	if err != nil {
		logger.Errorf("error parsing query-data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	totalCount, err := db.GetTotalValidatorsCount()
	if err != nil {
		logger.Errorf("error retrieving ejected validator count: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	qry := fmt.Sprintf(`
		SELECT
			validators.validatorindex,
			validators.pubkey,
			validators.withdrawableepoch,
			validators.balance,
			validators.effectivebalance,
			validators.slashed,
			validators.activationepoch,
			validators.exitepoch,
			validators.lastattestationslot,
			COALESCE(validator_names.name, '') AS name,
			validators.status AS state
		FROM validators
		LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
		WHERE (pubkeyhex LIKE $1
			OR CAST(validators.validatorindex AS text) LIKE $1)
		%s
		ORDER BY %s %s
		LIMIT $2 OFFSET $3`, dataQuery.StateFilter, dataQuery.OrderBy, dataQuery.OrderDir)

	var validators []*types.ValidatorsPageDataValidators
	err = db.DB.Select(&validators, qry, "%"+dataQuery.Search+"%", dataQuery.Length, dataQuery.Start)
	if err != nil {
		logger.Errorf("error retrieving validators data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	tableData := make([][]interface{}, len(validators))
	for i, v := range validators {
		tableData[i] = []interface{}{
			fmt.Sprintf("%x", v.PublicKey),
			fmt.Sprintf("%v", v.ValidatorIndex),
			[]interface{}{
				fmt.Sprintf("%.4f %v", float64(v.CurrentBalance)/float64(1e9)*price.GetEthPrice(currency), currency),
				fmt.Sprintf("%.1f %v", float64(v.EffectiveBalance)/float64(1e9)*price.GetEthPrice(currency), currency),
			},
			v.State,
			[]interface{}{
				v.ActivationEpoch,
				utils.EpochToTime(v.ActivationEpoch).Unix(),
			},
		}

		if v.ExitEpoch != 9223372036854775807 {
			tableData[i] = append(tableData[i], []interface{}{
				v.ExitEpoch,
				utils.EpochToTime(v.ExitEpoch).Unix(),
			})
		} else {
			tableData[i] = append(tableData[i], nil)
		}

		if v.WithdrawableEpoch != 9223372036854775807 {
			tableData[i] = append(tableData[i], []interface{}{
				v.WithdrawableEpoch,
				utils.EpochToTime(v.WithdrawableEpoch).Unix(),
			})
		} else {
			tableData[i] = append(tableData[i], nil)
		}

		if v.LastAttestationSlot != nil {
			tableData[i] = append(tableData[i], []interface{}{
				*v.LastAttestationSlot,
				utils.SlotToTime(uint64(*v.LastAttestationSlot)).Unix(),
			})
		} else {
			tableData[i] = append(tableData[i], nil)
		}

		tableData[i] = append(tableData[i], v.Slashed)

		tableData[i] = append(tableData[i], html.EscapeString(v.Name))
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
