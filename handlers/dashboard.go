package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/price"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"

	"strconv"
	"strings"

	"github.com/lib/pq"
)

var dashboardTemplate = template.Must(template.New("dashboard").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/dashboard.html"))

func parseValidatorsFromQueryString(str string, validatorLimit int) ([]uint64, error) {
	if str == "" {
		return []uint64{}, nil
	}

	strSplit := strings.Split(str, ",")
	strSplitLen := len(strSplit)

	// we only support up to 200 validators
	if strSplitLen > validatorLimit {
		return []uint64{}, fmt.Errorf("too much validators")
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
	validatorLimit := getUserPremium(r).MaxValidators

	dashboardData := types.DashboardData{}
	dashboardData.ValidatorLimit = validatorLimit

	data := InitPageData(w, r, "dashboard", "/dashboard", "Dashboard")
	data.HeaderAd = true
	data.Data = dashboardData

	err := dashboardTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error executing template")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

// DashboardDataBalance retrieves the income history of a set of validators
func DashboardDataBalance(w http.ResponseWriter, r *http.Request) {
	currency := GetCurrency(r)

	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	validatorLimit := getUserPremium(r).MaxValidators
	queryValidators, err := parseValidatorsFromQueryString(q.Get("validators"), validatorLimit)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error parsing validators from query string")
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

	var incomeHistory []*types.ValidatorIncomeHistory
	err = db.ReaderDb.Select(&incomeHistory, "SELECT day, COALESCE(SUM(start_balance),0) AS start_balance, COALESCE(SUM(end_balance),0) AS end_balance, COALESCE(SUM(deposits_amount), 0) AS deposits_amount FROM validator_stats WHERE validatorindex = ANY($1) GROUP BY day ORDER BY day;", queryValidatorsArr)
	if err != nil {
		logger.Errorf("error retrieving validator balance history: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	var currentBalance uint64
	err = db.ReaderDb.Get(&currentBalance, "SELECT SUM(balance) as balance FROM validators WHERE validatorindex = ANY($1) AND status <> 'deposited'", queryValidatorsArr)
	if err != nil {
		logger.Errorf("error retrieving validator current balance: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	incomeHistoryChartData := make([]*types.ChartDataPoint, len(incomeHistory)+1)

	if len(incomeHistory) > 0 {
		for i := 0; i < len(incomeHistory); i++ {
			var income int64
			if i == len(incomeHistory)-1 {
				income = incomeHistory[i].EndBalance - incomeHistory[i].StartBalance - incomeHistory[i].Deposits
			} else {
				income = incomeHistory[i+1].StartBalance - incomeHistory[i].StartBalance - incomeHistory[i].Deposits
			}
			color := "#7cb5ec"
			if income < 0 {
				color = "#f7a35c"
			}
			change := utils.ExchangeRateForCurrency(currency) * (float64(income) / 1000000000)
			balanceTs := utils.DayToTime(incomeHistory[i].Day)
			incomeHistoryChartData[i] = &types.ChartDataPoint{X: float64(balanceTs.Unix() * 1000), Y: change, Color: color}
		}

		lastDayBalance := incomeHistory[len(incomeHistory)-1].EndBalance
		lastDayIncome := int64(currentBalance) - lastDayBalance
		lastDayIncomeColor := "#7cb5ec"
		if lastDayIncome < 0 {
			lastDayIncomeColor = "#f7a35c"
		}

		currentDay := latestEpoch / ((24 * 60 * 60) / utils.Config.Chain.Config.SlotsPerEpoch / utils.Config.Chain.Config.SecondsPerSlot)

		incomeHistoryChartData[len(incomeHistoryChartData)-1] = &types.ChartDataPoint{X: float64(utils.DayToTime(int64(currentDay)).Unix() * 1000), Y: utils.ExchangeRateForCurrency(currency) * (float64(lastDayIncome) / 1000000000), Color: lastDayIncomeColor}
	}

	err = json.NewEncoder(w).Encode(incomeHistoryChartData)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error enconding json response")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func DashboardDataProposals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	validatorLimit := getUserPremium(r).MaxValidators
	filterArr, err := parseValidatorsFromQueryString(q.Get("validators"), validatorLimit)
	if err != nil {
		http.Error(w, "Invalid query", 400)
		return
	}
	filter := pq.Array(filterArr)

	proposals := []struct {
		Slot   uint64
		Status uint64
	}{}

	err = db.ReaderDb.Select(&proposals, `
		SELECT slot, status
		FROM blocks
		WHERE proposer = ANY($1)
		ORDER BY slot`, filter)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error retrieving block-proposals")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	proposalsResult := make([][]uint64, len(proposals))
	for i, b := range proposals {
		proposalsResult[i] = []uint64{
			uint64(utils.SlotToTime(b.Slot).Unix()),
			b.Status,
		}
	}

	err = json.NewEncoder(w).Encode(proposalsResult)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error enconding json response")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func DashboardDataValidators(w http.ResponseWriter, r *http.Request) {
	currency := GetCurrency(r)

	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	validatorLimit := getUserPremium(r).MaxValidators
	filterArr, err := parseValidatorsFromQueryString(q.Get("validators"), validatorLimit)
	if err != nil {
		http.Error(w, "Invalid query", 400)
		return
	}
	filter := pq.Array(filterArr)

	var validators []*types.ValidatorsPageDataValidators
	err = db.ReaderDb.Select(&validators, `
		WITH
			proposals AS (
				SELECT validatorindex, pa.status, count(*)
				FROM proposal_assignments pa
				INNER JOIN blocks b ON pa.proposerslot = b.slot AND b.status <> '3'
				WHERE validatorindex = ANY($1)
				GROUP BY validatorindex, pa.status
			)
		SELECT
			validators.validatorindex,
			validators.pubkey,
			validators.withdrawableepoch,
			validators.balance,
			validators.effectivebalance,
			validators.slashed,
			validators.activationeligibilityepoch,
			validators.lastattestationslot,
			validators.activationepoch,
			validators.exitepoch,
			COALESCE(p1.count, 0) as executedproposals,
			COALESCE(p2.count, 0) as missedproposals,
			COALESCE(validator_performance.performance7d, 0) as performance7d,
			COALESCE(validator_names.name, '') AS name,
		    validators.status AS state
		FROM validators
		LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
		LEFT JOIN proposals p1 ON validators.validatorindex = p1.validatorindex AND p1.status = 1
		LEFT JOIN proposals p2 ON validators.validatorindex = p2.validatorindex AND p2.status = 2
		LEFT JOIN validator_performance ON validators.validatorindex = validator_performance.validatorindex
		WHERE validators.validatorindex = ANY($1)
		LIMIT $2`, filter, validatorLimit)

	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Errorf("error retrieving validator data")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
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
		}

		if v.ActivationEpoch != 9223372036854775807 {
			tableData[i] = append(tableData[i], []interface{}{
				v.ActivationEpoch,
				utils.EpochToTime(v.ActivationEpoch).Unix(),
			})
		} else {
			tableData[i] = append(tableData[i], nil)
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

		// tableData[i] = append(tableData[i], []interface{}{
		// 	v.ExecutedAttestations,
		// 	v.MissedAttestations,
		// })

		// tableData[i] = append(tableData[i], fmt.Sprintf("%.4f ETH", float64(v.Performance7d)/float64(1e9)))
		tableData[i] = append(tableData[i], utils.FormatIncome(v.Performance7d, currency))
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
		logger.WithError(err).WithField("route", r.URL.String()).Errorf("error enconding json response")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func DashboardDataEarnings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	validatorLimit := getUserPremium(r).MaxValidators
	queryValidators, err := parseValidatorsFromQueryString(q.Get("validators"), validatorLimit)
	if err != nil {
		http.Error(w, "Invalid query", 400)
		return
	}

	earnings, err := GetValidatorEarnings(queryValidators, GetCurrency(r))
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Errorf("error retrieving validator earnings")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
	}

	if earnings == nil {
		earnings = &types.ValidatorEarnings{}
	}

	err = json.NewEncoder(w).Encode(earnings)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Errorf("error enconding json response")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func DashboardDataEffectiveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	validatorLimit := getUserPremium(r).MaxValidators
	filterArr, err := parseValidatorsFromQueryString(q.Get("validators"), validatorLimit)
	if err != nil {
		logger.Errorf("error retrieving active validators %v", err)
		http.Error(w, "Invalid query", 400)
		return
	}
	filter := pq.Array(filterArr)

	var activeValidators []uint64
	err = db.ReaderDb.Select(&activeValidators, `
		SELECT validatorindex FROM validators where validatorindex = ANY($1) and activationepoch < $2 AND exitepoch > $2
	`, filter, services.LatestEpoch())
	if err != nil {
		logger.Errorf("error retrieving active validators")
	}

	var avgIncDistance []float64

	effectiveness, err := db.BigtableClient.GetValidatorEffectiveness(activeValidators, services.LatestEpoch()-1)
	for _, e := range effectiveness {
		avgIncDistance = append(avgIncDistance, 100-((1+e.AttestationEfficiency)/32*100))
	}
	if err != nil {
		logger.Errorf("error retrieving AverageAttestationInclusionDistance: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	err = json.NewEncoder(w).Encode(avgIncDistance)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func DashboardDataProposalsHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	validatorLimit := getUserPremium(r).MaxValidators
	filterArr, err := parseValidatorsFromQueryString(q.Get("validators"), validatorLimit)
	if err != nil {
		http.Error(w, "Invalid query", 400)
		return
	}
	filter := pq.Array(filterArr)

	proposals := []struct {
		ValidatorIndex uint64  `db:"validatorindex"`
		Day            int64   `db:"day"`
		Proposed       *uint64 `db:"proposed_blocks"`
		Missed         *uint64 `db:"missed_blocks"`
		Orphaned       *uint64 `db:"orphaned_blocks"`
	}{}

	err = db.ReaderDb.Select(&proposals, `
		SELECT validatorindex, day, proposed_blocks, missed_blocks, orphaned_blocks
		FROM validator_stats
		WHERE validatorindex = ANY($1) AND (proposed_blocks IS NOT NULL OR missed_blocks IS NOT NULL OR orphaned_blocks IS NOT NULL)
		ORDER BY day DESC`, filter)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error retrieving validator_stats")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	proposalsHistResult := make([][]uint64, len(proposals))
	for i, b := range proposals {
		var proposed, missed, orphaned uint64 = 0, 0, 0
		if b.Proposed != nil {
			proposed = *b.Proposed
		}
		if b.Missed != nil {
			missed = *b.Missed
		}
		if b.Orphaned != nil {
			orphaned = *b.Orphaned
		}
		proposalsHistResult[i] = []uint64{
			b.ValidatorIndex,
			uint64(utils.DayToTime(b.Day).Unix()),
			proposed,
			missed,
			orphaned,
		}
	}

	err = json.NewEncoder(w).Encode(proposalsHistResult)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error enconding json response")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}
