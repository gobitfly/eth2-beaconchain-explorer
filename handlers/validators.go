package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/price"
	"github.com/gobitfly/eth2-beaconchain-explorer/services"
	"github.com/gobitfly/eth2-beaconchain-explorer/templates"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
)

// Validators returns the validators using a go template
func Validators(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "validators.html")
	var validatorsTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	validatorsPageData := types.ValidatorsPageData{}

	var currentStateCounts []*types.ValidatorStateCountRow
	err := db.ReaderDb.Select(&currentStateCounts, "SELECT status, validator_count FROM validators_status_counts")
	if err != nil {
		utils.LogError(err, "error retrieving validators state counts", 0, nil)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	for _, status := range currentStateCounts {
		switch status.Name {
		case "pending":
			validatorsPageData.PendingCount = status.Count
		case "active_online":
			validatorsPageData.ActiveOnlineCount = status.Count
		case "active_offline":
			validatorsPageData.ActiveOfflineCount = status.Count
		case "slashing_online":
			validatorsPageData.SlashingOnlineCount = status.Count
		case "slashing_offline":
			validatorsPageData.SlashingOfflineCount = status.Count
		case "slashed":
			validatorsPageData.Slashed = status.Count
		case "exiting_online":
			validatorsPageData.ExitingOnlineCount = status.Count
		case "exiting_offline":
			validatorsPageData.ExitingOfflineCount = status.Count
		case "exited":
			validatorsPageData.VoluntaryExitsCount = status.Count
		case "deposited":
			validatorsPageData.DepositedCount = status.Count
		}
	}

	epoch := services.LatestEpoch()

	validatorsPageData.ActiveCount = validatorsPageData.ActiveOnlineCount + validatorsPageData.ActiveOfflineCount
	validatorsPageData.SlashingCount = validatorsPageData.SlashingOnlineCount + validatorsPageData.SlashingOfflineCount
	validatorsPageData.ExitingCount = validatorsPageData.ExitingOnlineCount + validatorsPageData.ExitingOfflineCount
	validatorsPageData.ExitedCount = validatorsPageData.VoluntaryExitsCount + validatorsPageData.Slashed
	validatorsPageData.TotalCount = validatorsPageData.ActiveCount + validatorsPageData.ExitingCount + validatorsPageData.ExitedCount + validatorsPageData.PendingCount + validatorsPageData.DepositedCount
	validatorsPageData.CappellaHasHappened = epoch >= (utils.Config.Chain.ClConfig.CappellaForkEpoch)

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

var searchPubkeyExactRE = regexp.MustCompile(`^(0x)?[0-9a-fA-F]{96}`) // only search for pubkeys if string consists of 96 hex-chars
var searchPubkeyLikeRE = regexp.MustCompile(`^(0x)?[0-9a-fA-F]{2,96}`)

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
	case "online":
		qryStateFilter = "WHERE validators.status LIKE '%online'"
	case "offline":
		qryStateFilter = "WHERE validators.status LIKE '%offline'"
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
		logger.Warnf("error converting datatables draw parameter from string to int: %v", err)
		return nil, err
	}

	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables start parameter from string to int: %v", err)
		return nil, err
	}
	if start > 10000 {
		// limit offset to 10000, otherwise the query will be too slow
		start = 10000
	}

	length, err := strconv.ParseInt(q.Get("length"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables length parameter from string to int: %v", err)
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

// ValidatorsData returns all validators and basic information about them based on a StateFilter
func ValidatorsData(w http.ResponseWriter, r *http.Request) {
	currency := GetCurrency(r)

	w.Header().Set("Content-Type", "application/json")

	dataQuery, err := parseValidatorsDataQueryParams(r)
	if err != nil {
		logger.Warnf("error parsing query-data: %v", err)
		http.Error(w, "Error: Invalid query-data parameter.", http.StatusBadRequest)
		return
	}

	errFields := map[string]interface{}{
		"route":     r.URL.String(),
		"dataQuery": dataQuery,
	}

	var validators []*types.ValidatorsData
	qry := fmt.Sprintf(`
		SELECT  
		validators.validatorindex,  
		validators.pubkey,  
		validators.withdrawableepoch,  
		validators.slashed,  
		validators.activationepoch,  
		validators.exitepoch,  
		COALESCE(validator_names.name, '') AS name,  
		validators.status AS state  
		FROM validators  
		LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey  
		%s
		ORDER BY %s %s  
		LIMIT $1 OFFSET $2`, dataQuery.StateFilter, dataQuery.OrderBy, dataQuery.OrderDir)

	err = db.ReaderDb.Select(&validators, qry, dataQuery.Length, dataQuery.Start)
	if err != nil {
		utils.LogError(err, "error retrieving validators data", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tableData := make([][]interface{}, len(validators))
	if len(validators) > 0 {
		indices := make([]uint64, len(validators))
		for i, validator := range validators {
			indices[i] = validator.ValidatorIndex
		}
		balances, err := db.BigtableClient.GetValidatorBalanceHistory(indices, services.LatestEpoch(), services.LatestEpoch())
		if err != nil {
			utils.LogError(err, "error retrieving validator balance data", 0, errFields)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
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

		lastAttestationSlots, err := db.BigtableClient.GetLastAttestationSlots(indices)
		if err != nil {
			utils.LogError(err, "error retrieving validator last attestation slot data", 0, errFields)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		for _, validator := range validators {
			validator.LastAttestationSlot = int64(lastAttestationSlots[validator.ValidatorIndex])
		}

		for i, v := range validators {
			tableData[i] = []interface{}{
				fmt.Sprintf("%x", v.PublicKey),
				fmt.Sprintf("%v", v.ValidatorIndex),
				[]interface{}{
					fmt.Sprintf("%.4f %v", float64(v.CurrentBalance)/float64(utils.Config.Frontend.ClCurrencyDivisor)*price.GetPrice(utils.Config.Frontend.ClCurrency, currency), currency),
					fmt.Sprintf("%.1f %v", float64(v.EffectiveBalance)/float64(utils.Config.Frontend.ClCurrencyDivisor)*price.GetPrice(utils.Config.Frontend.ClCurrency, currency), currency),
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

			if v.LastAttestationSlot > 0 {
				tableData[i] = append(tableData[i], []interface{}{
					v.LastAttestationSlot,
					utils.SlotToTime(uint64(v.LastAttestationSlot)).Unix(),
				})
			} else {
				tableData[i] = append(tableData[i], nil)
			}

			tableData[i] = append(tableData[i], v.Slashed)

			tableData[i] = append(tableData[i], html.EscapeString(v.Name))
		}
	}

	countTotal := uint64(0)
	qry = "SELECT MAX(validatorindex) + 1 as total FROM validators"
	err = db.ReaderDb.Get(&countTotal, qry)
	if err != nil {
		utils.LogError(err, "error retrieving validators total count", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	countFiltered := uint64(0)
	if dataQuery.StateFilter != "" {
		qry = fmt.Sprintf(`SELECT SUM(validator_count) FROM validators_status_counts AS validators %s`, dataQuery.StateFilter)
		err = db.ReaderDb.Get(&countFiltered, qry)
		if err != nil {
			utils.LogError(err, "error retrieving validators total count", 0, errFields)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		countFiltered = countTotal
	}

	if countTotal > 10000 {
		countTotal = 10000
	}
	if countFiltered > 10000 {
		countFiltered = 10000
	}

	data := &types.DataTableResponse{
		Draw:            dataQuery.Draw,
		RecordsTotal:    countTotal,
		RecordsFiltered: countFiltered,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		utils.LogError(err, "error encoding json response", 0, errFields)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
