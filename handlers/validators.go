package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/price"
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type states struct {
	Name  string `db:"statename"`
	Count uint64 `db:"statecount"`
}

// Validators returns the validators using a go template
func Validators(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "validators.html")
	var validatorsTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	validatorsPageData := types.ValidatorsPageData{}

	var currentStateCounts []*states

	qry := "SELECT status AS statename, COUNT(*) AS statecount FROM validators GROUP BY status"
	err := db.ReaderDb.Select(&currentStateCounts, qry)
	if err != nil {
		logger.Errorf("error retrieving validators data: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
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

	epoch := services.LatestEpoch()

	validatorsPageData.ActiveCount = validatorsPageData.ActiveOnlineCount + validatorsPageData.ActiveOfflineCount
	validatorsPageData.SlashingCount = validatorsPageData.SlashingOnlineCount + validatorsPageData.SlashingOfflineCount
	validatorsPageData.ExitingCount = validatorsPageData.ExitingOnlineCount + validatorsPageData.ExitingOfflineCount
	validatorsPageData.ExitedCount = validatorsPageData.VoluntaryExitsCount + validatorsPageData.Slashed
	validatorsPageData.TotalCount = validatorsPageData.ActiveCount + validatorsPageData.ExitingCount + validatorsPageData.ExitedCount + validatorsPageData.PendingCount + validatorsPageData.DepositedCount
	validatorsPageData.CappellaHasHappened = epoch >= (utils.Config.Chain.Config.CappellaForkEpoch)

	data := InitPageData(w, r, "validators", "/validators", "Validators", templateFiles)
	data.Data = validatorsPageData

	if handleTemplateError(w, r, "validators.go", "Validators", "", validatorsTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
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

var searchPubkeyExactRE = regexp.MustCompile(`^(0x)?[0-9a-fA-F]{96}`)  // only search for pubkeys if string consists of 96 hex-chars
var searchPubkeyLikeRE = regexp.MustCompile(`^(0x)?[0-9a-fA-F]{2,96}`) // only search for pubkeys if string consists of 96 hex-chars

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
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	var validators []*types.ValidatorsPageDataValidators
	qry := ""
	// if dataQuery.Search == "" && dataQuery.StateFilter == "" {
	qry = fmt.Sprintf(`
			SELECT
				validators.validatorindex,
				validators.pubkey,
				validators.withdrawableepoch,
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

	err = db.ReaderDb.Select(&validators, qry, dataQuery.Length, dataQuery.Start)
	if err != nil {
		logger.Errorf("error retrieving validators data: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	indices := make([]uint64, len(validators))
	for i, validator := range validators {
		indices[i] = validator.ValidatorIndex
	}
	balances, err := db.BigtableClient.GetValidatorBalanceHistory(indices, services.LatestEpoch(), services.LatestEpoch())
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Errorf("error retrieving validator balance data")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	for _, validator := range validators {
		for balanceIndex, balance := range balances {
			if len(balance) == 0 {
				continue
			}
			if validator.ValidatorIndex == balanceIndex {
				validator.CurrentBalance = balance[0].Balance
				validator.EffectiveBalance = balance[0].EffectiveBalance
			}
		}
	}

	isAll := true

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
	qry = "SELECT MAX(validatorindex) + 1 as total FROM validators"
	err = db.ReaderDb.Get(&countTotal, qry)
	if err != nil {
		logger.Errorf("error retrieving validators total count: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
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
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}
