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
	"sync"

	"strconv"
	"strings"

	"github.com/lib/pq"
)

var dashboardTemplate = template.Must(template.New("dashboard").ParseFiles("templates/layout.html", "templates/dashboard.html"))

func parseValidatorsFromQueryString(str string) ([]uint64, error) {
	if str == "" {
		return []uint64{}, nil
	}

	strSplit := strings.Split(str, ",")
	strSplitLen := len(strSplit)

	// we only support up to 100 validators
	if strSplitLen > 100 {
		return []uint64{}, fmt.Errorf("Too much validators")
	}

	validators := make([]uint64, strSplitLen)
	keys := make(map[uint64]bool, strSplitLen)

	for i, vStr := range strSplit {
		v, err := strconv.ParseUint(vStr, 10, 64)
		if err != nil {
			return []uint64{}, err
		}
		// make sure keys are uniq
		if exists := keys[v]; exists {
			continue
		}
		keys[v] = true
		validators[i] = v
	}

	return validators, nil
}

func Dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := &types.PageData{
		Meta: &types.Meta{
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
		},
		ShowSyncingMessage: services.IsSyncing(),
		Active:             "dashboard",
		Data:               nil,
	}

	err := dashboardTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Fatalf("Error executing template for %v route: %v", r.URL.String(), err)
	}
}

func DashboardDataBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	queryValidators, err := parseValidatorsFromQueryString(q.Get("validators"))
	if err != nil {
		logger.WithError(err).Error("Failed parsing validators from query string")
		http.Error(w, "Invalid query", 400)
		return
	}
	if err != nil {
		http.Error(w, "Invalid query", 400)
		return
	}
	if len(queryValidators) < 1 {
		http.Error(w, "Invalid query", 400)
		return
	}
	queryValidatorsArr := pq.Array(queryValidators)

	// get data from one week before latest epoch
	latestEpoch := services.LatestEpoch()
	oneWeekEpochs := uint64(3600 * 24 * 7 / float64(utils.Config.Chain.SecondsPerSlot*utils.Config.Chain.SlotsPerEpoch))
	queryOffsetEpoch := uint64(0)
	if latestEpoch > oneWeekEpochs {
		queryOffsetEpoch = latestEpoch - oneWeekEpochs
	}

	query := `SELECT
			validator_balances.epoch,
			SUM(validator_balances.balance) AS balance,
			SUM(v.effectivebalance) AS effectivebalance,
			COUNT(*) AS validatorcount
		FROM (
			SELECT
				validatorindex,
				epoch,
				effectivebalance
			FROM validator_set
			WHERE
				validatorindex = ANY($1)
				AND epoch > $2
				AND epoch > activationepoch 
				AND epoch < exitepoch
		) AS v
		LEFT JOIN validator_balances
			ON validator_balances.validatorindex = v.validatorindex
			AND validator_balances.epoch = v.epoch
		GROUP BY validator_balances.epoch
		ORDER BY validator_balances.epoch ASC`

	data := []*types.DashboardValidatorBalanceHistory{}
	err = db.DB.Select(&data, query, queryValidatorsArr, queryOffsetEpoch)
	if err != nil {
		logger.Errorf("Error retrieving validator balance history: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	balanceHistoryChartData := make([][4]float64, len(data))
	for i, item := range data {
		balanceHistoryChartData[i][0] = float64(utils.EpochToTime(item.Epoch).Unix() * 1000)
		balanceHistoryChartData[i][1] = item.ValidatorCount
		balanceHistoryChartData[i][2] = float64(item.Balance) / 1e9
		balanceHistoryChartData[i][3] = float64(item.EffectiveBalance) / 1e9
	}

	err = json.NewEncoder(w).Encode(balanceHistoryChartData)
	if err != nil {
		logger.Fatalf("Error enconding json response for %v route: %v", r.URL.String(), err)
	}
}

func DashboardDataBalance2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	queryValidators, err := parseValidatorsFromQueryString(q.Get("validators"))
	if err != nil {
		logger.WithError(err).Error("Failed parsing validators from query string")
		http.Error(w, "Invalid query", 400)
		return
	}
	if err != nil {
		http.Error(w, "Invalid query", 400)
		return
	}
	if len(queryValidators) < 1 {
		http.Error(w, "Invalid query", 400)
		return
	}
	queryValidatorsArr := pq.Array(queryValidators)

	// get data from one week before latest epoch
	latestEpoch := services.LatestEpoch()
	oneWeekEpochs := uint64(3600 * 24 * 7 / float64(utils.Config.Chain.SecondsPerSlot*utils.Config.Chain.SlotsPerEpoch))
	queryOffsetEpoch := uint64(0)
	if latestEpoch > oneWeekEpochs {
		queryOffsetEpoch = latestEpoch - oneWeekEpochs
	}

	type data1Type struct {
		Epoch            uint64 `db:"epoch"`
		EffectiveBalance uint64 `db:"effectivebalance"`
		ValidatorCount   uint64 `db:"count"`
	}

	type data2Type struct {
		Epoch   uint64 `db:"epoch"`
		Balance uint64 `db:"balance"`
	}

	data1 := []*data1Type{}
	data2 := []*data2Type{}
	errs := make(chan error, 1)
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		query := `SELECT epoch, SUM(effectivebalance) as effectivebalance, count(*)
		FROM validator_set
		WHERE validatorindex = ANY($1)
		AND epoch > $2 AND epoch > activationepoch AND epoch < exitepoch
		GROUP BY epoch 
		ORDER BY epoch DESC`
		err := db.DB.Select(&data1, query, queryValidatorsArr, queryOffsetEpoch)
		if err != nil {
			logger.Errorf("Error retrieving validator effectivebalance history: %v", err)
		}
		errs <- err
		wg.Done()
	}()

	go func() {
		query := `SELECT epoch, SUM(balance) as balance
		FROM validator_balances
		WHERE validatorindex = ANY($1) AND epoch > $2
		GROUP BY epoch 
		ORDER BY epoch DESC`
		err := db.DB.Select(&data2, query, queryValidatorsArr, queryOffsetEpoch)
		if err != nil {
			logger.Errorf("Error retrieving validator balance history: %v", err)
		}
		errs <- err
		wg.Done()
	}()

	go func() {
		wg.Wait()
		close(errs)
	}()

	for err := range errs {
		if err != nil {
			http.Error(w, "Internal server error", 503)
			return
		}
	}

	balanceHistoryChartData := make([][4]float64, len(data1))

	for i, item := range data1 {
		if item.Epoch != data2[i].Epoch {
			logger.Errorf("Epoch of the results does not match: %v", i)
			http.Error(w, "Internal server error", 503)
			return
		}
		balanceHistoryChartData[i][0] = float64(utils.EpochToTime(item.Epoch).Unix() * 1000)
		balanceHistoryChartData[i][1] = float64(item.ValidatorCount)
		balanceHistoryChartData[i][3] = float64(item.EffectiveBalance) / 1e9
		balanceHistoryChartData[i][2] = float64(data2[i].Balance) / 1e9
	}

	err = json.NewEncoder(w).Encode(balanceHistoryChartData)
	if err != nil {
		logger.Fatalf("Error enconding json response for %v route: %v", r.URL.String(), err)
	}
}

func DashboardDataProposals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	filterArr, err := parseValidatorsFromQueryString(q.Get("validators"))
	if err != nil {
		http.Error(w, "Invalid query", 400)
		return
	}
	filter := pq.Array(filterArr)

	proposals := []struct {
		Day    uint64
		Status uint64
		Count  uint
	}{}

	err = db.DB.Select(&proposals, "select slot / 7200 as day, status, count(*) FROM blocks WHERE proposer = ANY($1) group by day, status order by day;", filter)
	if err != nil {
		logger.WithError(err).Error("Error retrieving Daily Proposed Blocks blocks count")
		http.Error(w, "Internal server error", 503)
		return
	}

	dailyProposalCount := []types.DailyProposalCount{}

	for i := 0; i < len(proposals); i++ {
		if i == len(proposals)-1 {
			if proposals[i].Status == 1 {
				dailyProposalCount = append(dailyProposalCount, types.DailyProposalCount{
					Day:      utils.SlotToTime(proposals[i].Day * 7200).Unix(),
					Proposed: proposals[i].Count,
					Missed:   0,
					Orphaned: 0,
				})
			} else if proposals[i].Status == 2 {
				dailyProposalCount = append(dailyProposalCount, types.DailyProposalCount{
					Day:      utils.SlotToTime(proposals[i].Day * 7200).Unix(),
					Proposed: 0,
					Missed:   proposals[i].Count,
					Orphaned: 0,
				})
			} else if proposals[i].Status == 3 {
				dailyProposalCount = append(dailyProposalCount, types.DailyProposalCount{
					Day:      utils.SlotToTime(proposals[i].Day * 7200).Unix(),
					Proposed: 0,
					Missed:   0,
					Orphaned: proposals[i].Count,
				})
			} else {
				logger.WithError(err).Error("Error parsing Daily Proposed Blocks unkown status")
			}
		} else {
			if proposals[i].Day == proposals[i+1].Day {
				dailyProposalCount = append(dailyProposalCount, types.DailyProposalCount{
					Day:      utils.SlotToTime(proposals[i].Day * 7200).Unix(),
					Proposed: proposals[i].Count,
					Missed:   proposals[i+1].Count,
					Orphaned: proposals[i+1].Count,
				})
				i++
			} else if proposals[i].Status == 1 {
				dailyProposalCount = append(dailyProposalCount, types.DailyProposalCount{
					Day:      utils.SlotToTime(proposals[i].Day * 7200).Unix(),
					Proposed: proposals[i].Count,
					Missed:   0,
					Orphaned: 0,
				})
			} else if proposals[i].Status == 2 {
				dailyProposalCount = append(dailyProposalCount, types.DailyProposalCount{
					Day:      utils.SlotToTime(proposals[i].Day * 7200).Unix(),
					Proposed: 0,
					Missed:   proposals[i].Count,
					Orphaned: 0,
				})
			} else if proposals[i].Status == 3 {
				dailyProposalCount = append(dailyProposalCount, types.DailyProposalCount{
					Day:      utils.SlotToTime(proposals[i].Day * 7200).Unix(),
					Proposed: 0,
					Missed:   0,
					Orphaned: proposals[i].Count,
				})
			} else {
				logger.WithError(err).Error("Error parsing Daily Proposed Blocks unkown status")
			}
		}
	}

	err = json.NewEncoder(w).Encode(dailyProposalCount)
	if err != nil {
		logger.Fatalf("Error enconding json response for %v route: %v", r.URL.String(), err)
	}
}

func DashboardDataValidators(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	filterArr, err := parseValidatorsFromQueryString(q.Get("validators"))
	if err != nil {
		http.Error(w, "Invalid query", 400)
		return
	}
	filter := pq.Array(filterArr)

	// dataQuery, err := parseDataQueryParams(r)
	// if err != nil {
	// 	http.Error(w, "Invalid query", 400)
	// 	return
	// }

	var totalCount uint64

	err = db.DB.Get(&totalCount, "SELECT COUNT(*) FROM validator_set WHERE epoch = $1 AND epoch < activationepoch AND validator_set.validatorindex = ANY($2)", services.LatestEpoch(), filter)
	if err != nil {
		logger.Errorf("Error retrieving validator count: %v", err)
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
			validator_balances.balance,
			(select max(epoch) from attestation_assignments where validator_set.validatorindex = attestation_assignments.validatorindex and status = 1) as lastattested,
			(select max(epoch) from proposal_assignments where validator_set.validatorindex = proposal_assignments.validatorindex and status = 1) as lastproposed
		FROM validator_set
		LEFT JOIN validator_balances
			ON validator_set.epoch = validator_balances.epoch
			AND validator_set.validatorindex = validator_balances.validatorindex
		LEFT JOIN validators
			ON validator_set.validatorindex = validators.validatorindex
		WHERE validator_set.validatorindex = ANY($1)
		LIMIT 100`, filter)

	if err != nil {
		logger.Errorf("Error retrieving validator data: %v", err)
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

	type dataType struct {
		LatestEpoch uint64          `json:"latestEpoch"`
		Data        [][]interface{} `json:"data"`
	}
	data := &dataType{
		LatestEpoch: services.LatestEpoch(),
		Data:        tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Fatalf("Error enconding json response for %v route: %v", r.URL.String(), err)
	}
}

func DashboardDataValidatorsPending(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	filterArr, err := parseValidatorsFromQueryString(q.Get("validators"))
	if err != nil {
		http.Error(w, "Invalid query", 400)
		return
	}
	filter := pq.Array(filterArr)

	dataQuery, err := parseDataQueryParams(r)
	if err != nil {
		http.Error(w, "Invalid query", 400)
		return
	}

	var totalCount uint64

	err = db.DB.Get(&totalCount, "SELECT COUNT(*) FROM validator_set WHERE epoch = $1 AND epoch < activationepoch AND validator_set.validatorindex = ANY($2)", services.LatestEpoch(), filter)
	if err != nil {
		logger.Errorf("Error retrieving pending validator count: %v", err)
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
		LEFT JOIN validator_balances 
			ON validator_set.epoch = validator_balances.epoch
			AND validator_set.validatorindex = validator_balances.validatorindex
		LEFT JOIN validators 
			ON validator_set.validatorindex = validators.validatorindex
		WHERE validator_set.epoch = $1 
			AND validator_set.epoch < activationepoch
			AND encode(validators.pubkey::bytea, 'hex') LIKE $2
			AND validator_set.validatorindex = ANY($4)
		ORDER BY %s %s 
		LIMIT 100 OFFSET $3`, dataQuery.OrderBy, dataQuery.OrderDir), services.LatestEpoch(), "%"+dataQuery.Search+"%", dataQuery.Start, filter)

	if err != nil {
		logger.Errorf("Error retrieving pending validator data: %v", err)
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

func DashboardDataValidatorsActive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	filterArr, err := parseValidatorsFromQueryString(q.Get("validators"))
	if err != nil {
		http.Error(w, "Invalid query", 400)
		return
	}
	filter := pq.Array(filterArr)

	dataQuery, err := parseDataQueryParams(r)
	if err != nil {
		http.Error(w, "Invalid query: "+err.Error(), 400)
		return
	}

	var totalCount uint64

	err = db.DB.Get(&totalCount, "SELECT COUNT(*) FROM validator_set WHERE epoch = $1 AND epoch > activationepoch AND epoch < exitepoch AND validator_set.validatorindex = ANY($2)", services.LatestEpoch(), filter)
	if err != nil {
		logger.Errorf("Error retrieving active validator count: %v", err)
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
			validator_balances.balance,
			(select max(epoch) from attestation_assignments where validator_set.validatorindex = attestation_assignments.validatorindex and status = 1) as lastattested,
			(select max(epoch) from proposal_assignments where validator_set.validatorindex = proposal_assignments.validatorindex and status = 1) as lastproposed
		FROM validator_set
		LEFT JOIN validator_balances 
			ON validator_set.epoch = validator_balances.epoch
			AND validator_set.validatorindex = validator_balances.validatorindex
		LEFT JOIN validators 
			ON validator_set.validatorindex = validators.validatorindex
		WHERE validator_set.epoch = $1 
			AND validator_set.epoch > activationepoch 
			AND validator_set.epoch < exitepoch 
			AND encode(validators.pubkey::bytea, 'hex') LIKE $2
			AND validator_set.validatorindex = ANY($4)
		ORDER BY %s %s 
		LIMIT 100 OFFSET $3`, dataQuery.OrderBy, dataQuery.OrderDir), services.LatestEpoch(), "%"+dataQuery.Search+"%", dataQuery.Start, filter)

	if err != nil {
		logger.Errorf("Error retrieving active validators data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	tableData := make([][]interface{}, len(validators))
	for i, v := range validators {
		var proposed *int64
		var attested *int64
		if v.LastAttested == nil {
			attested = new(int64)
		} else {
			att := utils.EpochToTime(uint64(*v.LastAttested)).Unix()
			attested = &att
		}
		if v.LastProposed == nil {
			proposed = new(int64)
		} else {
			pr := utils.EpochToTime(uint64(*v.LastProposed)).Unix()
			proposed = &pr
		}
		tableData[i] = []interface{}{
			fmt.Sprintf("%x", v.PublicKey),
			fmt.Sprintf("%v", v.ValidatorIndex),
			utils.FormatBalance(v.CurrentBalance),
			utils.FormatBalance(v.EffectiveBalance),
			fmt.Sprintf("%v", v.Slashed),
			fmt.Sprintf("%v", v.ActivationEligibilityEpoch),
			fmt.Sprintf("%v", v.ActivationEpoch),
			attested,
			proposed,
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

func DashboardDataValidatorsEjected(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	filterArr, err := parseValidatorsFromQueryString(q.Get("validators"))
	if err != nil {
		http.Error(w, "Invalid query", 400)
		return
	}
	filter := pq.Array(filterArr)

	dataQuery, err := parseDataQueryParams(r)
	if err != nil {
		http.Error(w, "Invalid query", 400)
		return
	}

	var totalCount uint64

	err = db.DB.Get(&totalCount, "SELECT COUNT(*) FROM validator_set WHERE epoch = $1 AND epoch > exitepoch AND validator_set.validatorindex = ANY($2)", services.LatestEpoch(), filter)
	if err != nil {
		logger.Errorf("Error retrieving ejected validator count: %v", err)
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
		LEFT JOIN validator_balances 
			ON validator_set.epoch = validator_balances.epoch
			AND validator_set.validatorindex = validator_balances.validatorindex
		LEFT JOIN validators 
			ON validator_set.validatorindex = validators.validatorindex
		WHERE validator_set.epoch = $1 
			AND validator_set.epoch > exitepoch
			AND encode(validators.pubkey::bytea, 'hex') LIKE $2
			AND validator_set.validatorindex = ANY($4)
		ORDER BY %s %s 
		LIMIT 100 OFFSET $3`, dataQuery.OrderBy, dataQuery.OrderDir), services.LatestEpoch(), "%"+dataQuery.Search+"%", dataQuery.Start, filter)

	if err != nil {
		logger.Errorf("Error retrieving ejected validators data: %v", err)
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
