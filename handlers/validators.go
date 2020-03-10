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

var validatorsTemplate = template.Must(template.New("validators").ParseFiles("templates/layout.html", "templates/validators.html"))

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

	var firstSlotOfPreviousEpoch uint64
	if latestEpoch < 1 {
		firstSlotOfPreviousEpoch = 0
	} else {
		firstSlotOfPreviousEpoch = (latestEpoch - 1) * utils.Config.Chain.SlotsPerEpoch
	}

	for _, validator := range validators {
		validatorsPageData.TotalCount++
		if latestEpoch > validator.ExitEpoch {
			validatorsPageData.ExitedCount++
		} else if latestEpoch < validator.ActivationEpoch {
			validatorsPageData.PendingCount++
		} else if validator.Slashed {
			// offline validators did not attest in the last 2 epochs (and are active for >1 epochs)
			if validator.ActivationEpoch < latestEpoch && (validator.LastAttestationSlot == nil || uint64(*validator.LastAttestationSlot) < firstSlotOfPreviousEpoch) {
				validatorsPageData.SlashingOfflineCount++
			} else {
				validatorsPageData.SlashingOnlineCount++
			}
		} else {
			// offline validators did not attest in the last 2 epochs (and are active for >1 epochs)
			if validator.ActivationEpoch < latestEpoch && (validator.LastAttestationSlot == nil || uint64(*validator.LastAttestationSlot) < firstSlotOfPreviousEpoch) {
				validatorsPageData.ActiveOfflineCount++
			} else {
				validatorsPageData.ActiveOnlineCount++
			}
		}
	}
	validatorsPageData.ActiveCount = validatorsPageData.ActiveOnlineCount + validatorsPageData.ActiveOfflineCount
	validatorsPageData.SlashingCount = validatorsPageData.SlashingOnlineCount + validatorsPageData.SlashingOfflineCount
	validatorsPageData.ExitingCount = validatorsPageData.ExitingOnlineCount + validatorsPageData.ExitingOfflineCount

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - Validators - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/validators",
		},
		ShowSyncingMessage:    services.IsSyncing(),
		Active:                "validators",
		Data:                  validatorsPageData,
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
	}

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
		qryStateFilter = "AND a.state LIKE 'pending%'"
	case "active":
		qryStateFilter = "AND a.state LIKE 'active%'"
	case "active_online":
		qryStateFilter = "AND a.state = 'active_online'"
	case "active_offline":
		qryStateFilter = "AND a.state = 'active_offline'"
	case "slashing":
		qryStateFilter = "AND a.state LIKE 'slashing%'"
	case "slashing_online":
		qryStateFilter = "AND a.state = 'slashing_online'"
	case "slashing_offline":
		qryStateFilter = "AND a.state = 'slashing_offline'"
	case "exiting":
		qryStateFilter = "AND a.state LIKE 'exiting%'"
	case "exiting_online":
		qryStateFilter = "AND a.state = 'exiting_online'"
	case "exiting_offline":
		qryStateFilter = "AND a.state = 'exiting_offline'"
	case "exited":
		qryStateFilter = "AND a.state = 'exited'"
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
		"8": "executedproposals",
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
	w.Header().Set("Content-Type", "application/json")

	dataQuery, err := parseValidatorsDataQueryParams(r)
	if err != nil {
		logger.Errorf("error parsing query-data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	var totalCount uint64
	err = db.DB.Get(&totalCount, `SELECT COUNT(*) FROM validators`)
	if err != nil {
		logger.Errorf("error retrieving ejected validator count: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	lastestEpoch := services.LatestEpoch()
	var firstSlotOfPreviousEpoch uint64
	if lastestEpoch < 1 {
		firstSlotOfPreviousEpoch = 0
	} else {
		firstSlotOfPreviousEpoch = (lastestEpoch - 1) * utils.Config.Chain.SlotsPerEpoch
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
			a.state,
			COALESCE(p1.c,0) as executedproposals,
			COALESCE(p2.c,0) as missedproposals
		FROM validators
		INNER JOIN (
			SELECT validatorindex,
			CASE 
				WHEN exitepoch <= $1 then 'exited'
				WHEN activationepoch > $1 then 'pending'
				WHEN slashed and activationepoch < $1 and (lastattestationslot < $2 OR lastattestationslot is null) then 'slashing_offline'
				WHEN slashed then 'slashing_online'
				WHEN activationepoch < $1 and (lastattestationslot < $2 OR lastattestationslot is null) then 'active_offline' 
				ELSE 'active_online'
			END AS state
			FROM validators
		) a ON a.validatorindex = validators.validatorindex
		LEFT JOIN (
			select validatorindex, count(*) as c 
			from proposal_assignments
			where status = 1
			group by validatorindex
		) p1 ON validators.validatorindex = p1.validatorindex
		LEFT JOIN (
			select validatorindex, count(*) as c 
			from proposal_assignments
			where status = 2
			group by validatorindex
		) p2 ON validators.validatorindex = p2.validatorindex
		WHERE (encode(validators.pubkey::bytea, 'hex') LIKE $3
			OR CAST(validators.validatorindex AS text) LIKE $3)
		%s
		ORDER BY %s %s
		LIMIT $4 OFFSET $5`, dataQuery.StateFilter, dataQuery.OrderBy, dataQuery.OrderDir)

	var validators []*types.ValidatorsPageDataValidators
	err = db.DB.Select(&validators, qry, lastestEpoch, firstSlotOfPreviousEpoch, "%"+dataQuery.Search+"%", dataQuery.Length, dataQuery.Start)
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
				// utils.FormatBalance(v.CurrentBalance),
				// utils.FormatBalance(v.EffectiveBalance),
				fmt.Sprintf("%.4f ETH", float64(v.CurrentBalance)/float64(1e9)),
				fmt.Sprintf("%.1f ETH", float64(v.EffectiveBalance)/float64(1e9)),
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

		tableData[i] = append(tableData[i], []interface{}{
			v.ExecutedProposals,
			v.MissedProposals,
		})
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
