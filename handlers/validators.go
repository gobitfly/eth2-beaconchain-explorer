package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/price"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html"
	"html/template"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var validatorsTemplate = template.Must(template.New("validators").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/validators.html"))

type states struct {
	Name  string `db:"statename"`
	Count uint64 `db:"statecount"`
}

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

	validatorsPageData.PendingCount = 0
	validatorsPageData.ActiveOnlineCount = 0
	validatorsPageData.ActiveOfflineCount = 0
	validatorsPageData.ActiveCount = 0
	validatorsPageData.SlashingOnlineCount = 0
	validatorsPageData.SlashingOfflineCount = 0
	validatorsPageData.SlashingCount = 0
	validatorsPageData.ExitingOnlineCount = 0
	validatorsPageData.ExitingOfflineCount = 0
	validatorsPageData.ExitedCount = 0
	validatorsPageData.VoluntaryExitsCount = 0
	validatorsPageData.DepositedCount = 0

	var currentStateCounts []*states

	qry := "SELECT status AS statename, COUNT(*) AS statecount FROM validators GROUP BY status"
	err = db.DB.Select(&currentStateCounts, qry)
	if err != nil {
		logger.Errorf("error retrieving validators data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	for _, state := range currentStateCounts {
		switch state.Name {
		case "pending":
			validatorsPageData.PendingCount = state.Count
		case "active_online":
			validatorsPageData.ActiveOnlineCount = state.Count
		case "active_offline":
			validatorsPageData.ActiveOfflineCount = state.Count
		case "slashing_online":
			validatorsPageData.SlashingOnlineCount = state.Count
		case "slashing_offline":
			validatorsPageData.SlashingOfflineCount = state.Count
		case "slashed":
			validatorsPageData.Slashed = state.Count
		case "exiting_online":
			validatorsPageData.ExitingOnlineCount = state.Count
		case "exiting_offline":
			validatorsPageData.ExitingOfflineCount = state.Count
		case "exited":
			validatorsPageData.VoluntaryExitsCount = state.Count
		case "deposited":
			validatorsPageData.DepositedCount = state.Count
		}
	}

	validatorsPageData.ActiveCount = validatorsPageData.ActiveOnlineCount + validatorsPageData.ActiveOfflineCount
	validatorsPageData.SlashingCount = validatorsPageData.SlashingOnlineCount + validatorsPageData.SlashingOfflineCount
	validatorsPageData.ExitingCount = validatorsPageData.ExitingOnlineCount + validatorsPageData.ExitingOfflineCount
	validatorsPageData.ExitedCount = validatorsPageData.VoluntaryExitsCount + validatorsPageData.Slashed
	validatorsPageData.TotalCount = validatorsPageData.ActiveCount + validatorsPageData.ExitingCount + validatorsPageData.ExitedCount + validatorsPageData.PendingCount + validatorsPageData.DepositedCount

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
	Search            string
	SearchIndex       *uint64
	SearchPubkeyExact *string
	SearchPubkeyLike  *string
	OrderBy           string
	OrderDir          string
	Draw              uint64
	Start             uint64
	Length            int64
	StateFilter       string
}

var searchPubkeyExactRE = regexp.MustCompile(`^0?x?[0-9a-fA-F]{96}`)  // only search for pubkeys if string consists of 96 hex-chars
var searchPubkeyLikeRE = regexp.MustCompile(`^0?x?[0-9a-fA-F]{2,96}`) // only search for pubkeys if string consists of 96 hex-chars

func parseValidatorsDataQueryParams(r *http.Request) (*ValidatorsDataQueryParams, error) {
	q := r.URL.Query()

	search := strings.Replace(q.Get("search[value]"), "0x", "", -1)
	if len(search) > 128 {
		search = search[:128]
	}
	var searchIndex *uint64
	if search == "" {
		index := uint64(0)
		searchIndex = &index
	} else {
		index, err := strconv.ParseUint(search, 10, 64)
		if err == nil {
			searchIndex = &index
		}
	}

	var searchPubkeyExact *string
	var searchPubkeyLike *string
	if searchPubkeyExactRE.MatchString(search) {
		pubkey := strings.ToLower(strings.Replace(search, "0x", "", -1))
		searchPubkeyExact = &pubkey
	} else if searchPubkeyLikeRE.MatchString(search) {
		pubkey := strings.ToLower(strings.Replace(search, "0x", "", -1))
		searchPubkeyLike = &pubkey
	}

	filterByState := q.Get("filterByState")
	var qryStateFilter string
	switch filterByState {
	case "pending":
		qryStateFilter = "WHERE validators.status = 'pending'"
	case "active":
		qryStateFilter = "WHERE validators.status LIKE 'active%'"
	case "active_online":
		qryStateFilter = "WHERE validators.status = 'active_online'"
	case "active_offline":
		qryStateFilter = "WHERE validators.status = 'active_offline'"
	case "slashing":
		qryStateFilter = "WHERE validators.status LIKE 'slashing%'"
	case "slashing_online":
		qryStateFilter = "WHERE validators.status = 'slashing_online'"
	case "slashing_offline":
		qryStateFilter = "WHERE validators.status = 'slashing_offline'"
	case "slashed":
		qryStateFilter = "WHERE validators.status = 'slashed'"
	case "exiting":
		qryStateFilter = "WHERE validators.status LIKE 'exiting%'"
	case "exiting_online":
		qryStateFilter = "WHERE validators.status = 'exiting_online'"
	case "exiting_offline":
		qryStateFilter = "WHERE validators.status = 'exiting_offline'"
	case "exited":
		qryStateFilter = "WHERE (validators.status = 'exited' OR validators.status = 'slashed')"
	case "voluntary":
		qryStateFilter = "WHERE validators.status = 'exited'"
	case "deposited":
		qryStateFilter = "WHERE validators.status = 'deposited'"
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
		Search:            search,
		SearchIndex:       searchIndex,
		SearchPubkeyExact: searchPubkeyExact,
		SearchPubkeyLike:  searchPubkeyLike,
		OrderBy:           orderBy,
		OrderDir:          orderDir,
		Draw:              draw,
		Start:             start,
		Length:            length,
		StateFilter:       qryStateFilter,
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

	var validators []*types.ValidatorsPageDataValidators
	qry := ""
	isAll := bool(false)
	if dataQuery.Search == "" && dataQuery.StateFilter == "" {
		qry = fmt.Sprintf(`
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
			ORDER BY %s %s
			LIMIT $1 OFFSET $2`, dataQuery.OrderBy, dataQuery.OrderDir)

		err = db.DB.Select(&validators, qry, dataQuery.Length, dataQuery.Start)
		if err != nil {
			logger.Errorf("error retrieving validators data: %v", err)
			http.Error(w, "Internal server error", 503)
			return
		}
		isAll = true
	} else {
		// for perfomance-reasons we combine multiple search results with `union`
		args := []interface{}{}
		searchQry := ""
		countWhere := ""

		if dataQuery.Search != "" {
			args = append(args, "%"+strings.ToLower(dataQuery.Search)+"%")
			countWhere += fmt.Sprintf(`LOWER(validator_names.name) LIKE $%d`, len(args))
			searchQry += `SELECT publickey AS pubkey FROM validator_names WHERE ` + countWhere
		} else {
			if searchQry != "" {
				searchQry += " UNION "
			}
			searchQry += "SELECT pubkey FROM validators"
		}
		if dataQuery.SearchIndex != nil && *dataQuery.SearchIndex != 0 {
			if searchQry != "" {
				searchQry += " UNION "
			}
			args = append(args, *dataQuery.SearchIndex)
			searchQry += fmt.Sprintf(`SELECT pubkey FROM validators WHERE validatorindex = $%d`, len(args))
		}
		if dataQuery.SearchPubkeyExact != nil {
			if searchQry != "" {
				searchQry += " UNION "
			}
			args = append(args, *dataQuery.SearchPubkeyExact)
			searchQry += fmt.Sprintf(`SELECT pubkey FROM validators WHERE pubkeyhex = $%d`, len(args))
		} else if dataQuery.SearchPubkeyLike != nil {
			if searchQry != "" {
				searchQry += " UNION "
			}
			args = append(args, *dataQuery.SearchPubkeyLike+"%")
			searchQry += fmt.Sprintf(`SELECT pubkey FROM validators WHERE pubkeyhex LIKE $%d`, len(args))
		}

		args = append(args, dataQuery.Length)
		args = append(args, dataQuery.Start)

		if searchQry == "" {
			logger.Errorf("error sql statement incomplete (without with statement)")
			http.Error(w, "Internal server error", 503)
			return
		}

		addAnd := ""
		if countWhere != "" {
			addAnd = "AND"
		}
		qry = fmt.Sprintf(`
			WITH matched_validators AS (%s)
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
					validators.status AS state,
					COALESCE(cnt.total_count, 0) as total_count
			FROM validators
			INNER JOIN matched_validators ON validators.pubkey = matched_validators.pubkey
			LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
			LEFT JOIN (SELECT count(*)
						FROM validators
						LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
						%s %s
    					%s) cnt(total_count) ON true
			%s
			ORDER BY %s %s
			LIMIT $%d OFFSET $%d`, searchQry, dataQuery.StateFilter, addAnd, countWhere, dataQuery.StateFilter, dataQuery.OrderBy, dataQuery.OrderDir, len(args)-1, len(args))

		err = db.DB.Select(&validators, qry, args...)
		if err != nil {
			logger.Errorf("error retrieving validators data (with search): %v", err)
			http.Error(w, "Internal server error", 503)
			return
		}
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

	countTotal := uint64(0)
	qry = "SELECT count(*) as total FROM validators"
	err = db.DB.Get(&countTotal, qry)
	if err != nil {
		logger.Errorf("error retrieving validators total count: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	data := &types.DataTableResponse{
		Draw:            dataQuery.Draw,
		RecordsTotal:    countTotal,
		RecordsFiltered: countTotal,
		Data:            tableData,
	}

	if !isAll && validators != nil {
		data.RecordsFiltered = validators[0].TotalCount
	}
	if validators == nil {
		data.RecordsFiltered = 0
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
