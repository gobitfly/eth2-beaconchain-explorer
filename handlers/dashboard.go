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
	"sync"

	"strconv"
	"strings"

	"github.com/lib/pq"
)

var dashboardTemplate = template.Must(template.New("dashboard").ParseFiles("templates/layout.html", "templates/dashboard.html"))

func parseValidatorsFromQueryString(str string) ([]int64, error) {
	if str == "" {
		return []int64{}, nil
	}

	strSplit := strings.Split(str, ",")
	strSplitLen := len(strSplit)

	// we only support up to 100 validators
	if strSplitLen > 100 {
		return []int64{}, fmt.Errorf("Too much validators")
	}

	validators := make([]int64, strSplitLen)
	keys := make(map[int64]bool, strSplitLen)

	for i, vStr := range strSplit {
		v, err := strconv.ParseInt(vStr, 10, 64)
		if err != nil {
			return []int64{}, err
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
		Version:            version.Version,
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
			validator_set.epoch, 
			SUM(effectivebalance) AS effectivebalance, 
			COALESCE(SUM(balance),0) AS balance, 
			COUNT(*) AS validatorcount
		FROM validator_set 
		LEFT JOIN validator_balances 
			ON validator_set.epoch = validator_balances.epoch 
			AND validator_set.validatorindex = validator_balances.validatorindex 
		WHERE validator_set.validatorindex = any($1) 
			AND validator_set.epoch > $2
		GROUP BY validator_set.epoch
		ORDER BY validator_set.epoch ASC`

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

	err = db.DB.Select(&proposals, `SELECT 
			slot / 7200 AS day, 
			status, 
			COUNT(*) 
		FROM blocks 
		WHERE proposer = ANY($1) 
		GROUP BY day, status 
		ORDER BY day`, filter)
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

	logger.Errorf("getting dashboard-validators %v %v", services.LatestEpoch(), services.LatestEpoch()-1)
	q := r.URL.Query()

	filterArr, err := parseValidatorsFromQueryString(q.Get("validators"))
	if err != nil {
		http.Error(w, "Invalid query", 400)
		return
	}
	filter := pq.Array(filterArr)

	var validators []*types.ValidatorsPageDataValidators
	err = db.DB.Select(&validators, `SELECT
			validators.validatorindex,
			validators.pubkey,
			validators.withdrawableepoch,
			validators.effectivebalance,
			validators.slashed,
			validators.activationeligibilityepoch,
			validators.activationepoch,
			validators.exitepoch,
			validator_balances.balance,
			lastattestations.epoch as lastattestedepoch,
			lastproposals.epoch as lastproposedepoch
		FROM validators
		LEFT JOIN validator_balances
			ON validator_balances.epoch = $1
			AND validator_balances.validatorindex = validators.validatorindex
		LEFT JOIN (
			SELECT validatorindex, MAX(epoch) as epoch
			FROM attestation_assignments 
			WHERE validatorindex = ANY($2) AND status = 1
			GROUP BY validatorindex
		) AS lastattestations ON lastattestations.validatorindex = validators.validatorindex
		LEFT JOIN (
			SELECT validatorindex, MAX(epoch) as epoch
			FROM proposal_assignments 
			WHERE validatorindex = ANY($2) AND status = 1
			GROUP BY validatorindex
		) AS lastproposals ON lastproposals.validatorindex = validators.validatorindex
		WHERE validators.validatorindex = ANY($2)
		LIMIT 100`, services.LatestEpoch(), filter)

	if err != nil {
		logger.Errorf("Error retrieving validator data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	tableData := make([][]interface{}, len(validators))
	for i, v := range validators {
		var lastAttested interface{}
		if v.LastAttestedEpoch == nil {
			lastAttested = nil
		} else {
			lastAttested = []interface{}{
				fmt.Sprintf("%v", *v.LastAttestedEpoch),
				fmt.Sprintf("%v", utils.EpochToTime(uint64(*v.LastAttestedEpoch)).Unix()),
			}
		}
		var lastProposed interface{}
		if v.LastProposedEpoch == nil {
			lastProposed = nil
		} else {
			lastProposed = []interface{}{
				fmt.Sprintf("%v", *v.LastProposedEpoch),
				fmt.Sprintf("%v", utils.EpochToTime(uint64(*v.LastProposedEpoch)).Unix()),
			}
		}
		tableData[i] = []interface{}{
			fmt.Sprintf("%x", v.PublicKey),
			fmt.Sprintf("%v", v.ValidatorIndex),
			utils.FormatBalance(v.CurrentBalance),
			utils.FormatBalance(v.EffectiveBalance),
			fmt.Sprintf("%v", v.Slashed),
			[]interface{}{ // 5
				v.ActivationEligibilityEpoch,
				utils.EpochToTime(v.ActivationEligibilityEpoch).Unix(),
			},
			[]interface{}{ // 6
				v.ActivationEpoch,
				utils.EpochToTime(v.ActivationEpoch).Unix(),
			},
			[]interface{}{ // 7
				v.ExitEpoch,
				utils.EpochToTime(v.ExitEpoch).Unix(),
			},
			lastAttested,
			lastProposed,
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

func DashboardDataEarnings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	queryValidators, err := parseValidatorsFromQueryString(q.Get("validators"))
	if err != nil {
		http.Error(w, "Invalid query", 400)
		return
	}
	queryValidatorsArr := pq.Array(queryValidators)

	latestEpoch := services.LatestEpoch()

	oneDayEpochs := uint64(3600 * 24 * 1 / float64(utils.Config.Chain.SecondsPerSlot*utils.Config.Chain.SlotsPerEpoch))
	oneWeekEpochs := oneDayEpochs * 7
	oneMonthEpochs := oneDayEpochs * 30

	lastDayEpoch := uint64(0)
	if latestEpoch > oneDayEpochs {
		lastDayEpoch = latestEpoch - oneDayEpochs
	}

	lastWeekEpoch := uint64(0)
	if latestEpoch > oneWeekEpochs {
		lastWeekEpoch = latestEpoch - oneWeekEpochs
	}

	lastMonthEpoch := uint64(0)
	if latestEpoch > oneMonthEpochs {
		lastWeekEpoch = latestEpoch - oneMonthEpochs
	}

	earningsTotalQuery := `SELECT 
			SUM(last.balance - first.balance) AS earnings
		FROM (
			SELECT 
				validatorindex, 
				MIN(epoch) AS firstepoch, 
				MAX(epoch) AS lastepoch
			FROM validator_balances
			WHERE validatorindex = any($1)
			GROUP by validatorindex
		) minmaxepoch
		INNER JOIN validator_balances first
			ON first.validatorindex = minmaxepoch.validatorindex
			AND first.epoch = minmaxepoch.firstepoch
		INNER JOIN validator_balances last
			ON last.validatorindex = minmaxepoch.validatorindex
			AND last.epoch = minmaxepoch.lastepoch`

	earningsRangeQuery := `SELECT 
			SUM(last.balance - first.balance) AS earnings
		FROM (
			SELECT 
				validatorindex, 
				MIN(epoch) AS firstepoch, 
				MAX(epoch) AS lastepoch
			FROM validator_balances
			WHERE validatorindex = any($1) AND epoch > $2
			GROUP by validatorindex
		) minmaxepoch
		INNER JOIN validator_balances first
			ON first.validatorindex = minmaxepoch.validatorindex
			AND first.epoch = minmaxepoch.firstepoch
		INNER JOIN validator_balances last
			ON last.validatorindex = minmaxepoch.validatorindex
			AND last.epoch = minmaxepoch.lastepoch`

	var earningsTotal int64
	var earningsLastDay int64
	var earningsLastWeek int64
	var earningsLastMonth int64

	wg := sync.WaitGroup{}
	wg.Add(4)
	errs := make(chan error, 4)

	go func() {
		defer wg.Done()
		err := db.DB.Get(&earningsTotal, earningsTotalQuery, queryValidatorsArr)
		if err != nil {
			logger.WithField("route", r.URL.String()).Errorf("error retrieving total earnings: %v", err)
		}
		errs <- err
	}()

	go func() {
		defer wg.Done()
		err := db.DB.Get(&earningsLastDay, earningsRangeQuery, queryValidatorsArr, lastDayEpoch)
		if err != nil {
			logger.WithField("route", r.URL.String()).Errorf("error retrieving earnings of last day: %v", err)
		}
		errs <- err
	}()

	go func() {
		defer wg.Done()
		err := db.DB.Get(&earningsLastWeek, earningsRangeQuery, queryValidatorsArr, lastWeekEpoch)
		if err != nil {
			logger.WithField("route", r.URL.String()).Errorf("error retrieving earnings of last week: %v", err)
		}
		errs <- err
	}()

	go func() {
		defer wg.Done()
		err := db.DB.Get(&earningsLastMonth, earningsRangeQuery, queryValidatorsArr, lastMonthEpoch)
		if err != nil {
			logger.WithField("route", r.URL.String()).Errorf("error retrieving earnings of last month: %v", err)
		}
		errs <- err
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

	earnings := &types.DashboardEarnings{
		Total:     earningsTotal,
		LastDay:   earningsLastDay,
		LastWeek:  earningsLastWeek,
		LastMonth: earningsLastMonth,
	}

	err = json.NewEncoder(w).Encode(earnings)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
