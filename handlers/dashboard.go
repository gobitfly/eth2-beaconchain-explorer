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
	"time"

	"strconv"
	"strings"

	"github.com/lib/pq"
)

var dashboardTemplate = template.Must(template.New("dashboard").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/dashboard.html"))

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

	dashboardData := types.DashboardData{}

	data := InitPageData(w, r, "dashboard", "/dashboard", "Dashboard")
	data.HeaderAd = true
	data.Data = dashboardData

	err := dashboardTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error executing template")
		http.Error(w, "Internal server error", 503)
		return
	}
}

func DashboardDataBalance(w http.ResponseWriter, r *http.Request) {
	currency := GetCurrency(r)

	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	queryValidators, err := parseValidatorsFromQueryString(q.Get("validators"))
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
	oneWeekEpochs := uint64(3600 * 24 * 7 / float64(utils.Config.Chain.SecondsPerSlot*utils.Config.Chain.SlotsPerEpoch))
	queryOffsetEpoch := uint64(0)
	if latestEpoch > oneWeekEpochs {
		queryOffsetEpoch = latestEpoch - oneWeekEpochs
	}

	query := `
		SELECT
			epoch,
			COALESCE(SUM(effectivebalance),0) AS effectivebalance,
			COALESCE(SUM(balance),0) AS balance,
			COUNT(*) AS validatorcount
		FROM validator_balances_p
		WHERE validatorindex = ANY($1) AND epoch > $2 AND week >= $2 / 1575
		GROUP BY epoch
		ORDER BY epoch ASC`

	data := []*types.DashboardValidatorBalanceHistory{}
	err = db.DB.Select(&data, query, queryValidatorsArr, queryOffsetEpoch)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Errorf("error retrieving validator balance history")
		http.Error(w, "Internal server error", 503)
		return
	}

	balanceHistoryChartData := make([][4]float64, len(data))
	for i, item := range data {
		balanceHistoryChartData[i][0] = float64(utils.EpochToTime(item.Epoch).Unix() * 1000)
		balanceHistoryChartData[i][1] = item.ValidatorCount
		balanceHistoryChartData[i][2] = float64(item.Balance) / 1e9 * price.GetEthPrice(currency)
		balanceHistoryChartData[i][3] = float64(item.EffectiveBalance) / 1e9 * price.GetEthPrice(currency)
	}

	err = json.NewEncoder(w).Encode(balanceHistoryChartData)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error enconding json response")
		http.Error(w, "Internal server error", 503)
		return
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
		Slot   uint64
		Status uint64
	}{}

	err = db.DB.Select(&proposals, `
		SELECT slot, status
		FROM blocks
		WHERE proposer = ANY($1)
		ORDER BY slot`, filter)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error retrieving block-proposals")
		http.Error(w, "Internal server error", 503)
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
		http.Error(w, "Internal server error", 503)
		return
	}
}

func DashboardDataMissedAttestations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	filterArr, err := parseValidatorsFromQueryString(q.Get("validators"))
	if err != nil {
		http.Error(w, "Invalid query", 400)
		return
	}
	filter := pq.Array(filterArr)

	missedAttestations := []struct {
		Epoch          uint64
		Validatorindex uint64
	}{}

	maxEpoch := services.LatestEpoch() - 1
	minEpoch := utils.TimeToEpoch(time.Now().Add(time.Hour * 24 * -7))

	err = db.DB.Select(&missedAttestations, `
		SELECT epoch, validatorindex
		FROM attestation_assignments_p
		WHERE 
			validatorindex = ANY($1) 
			AND epoch <= $2 
			AND epoch >= $3 
			AND week <= $2 / 1575
			AND week >= $3 / 1575
			AND status = 0`, filter, maxEpoch, minEpoch)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error retrieving daily proposed blocks blocks count")
		http.Error(w, "Internal server error", 503)
		return
	}

	result := make(map[int64][]uint64)

	for _, ma := range missedAttestations {
		ts := utils.EpochToTime(ma.Epoch).Unix()
		if _, exists := result[ts]; !exists {
			result[ts] = []uint64{ma.Validatorindex}
		} else {
			result[ts] = append(result[ts], ma.Validatorindex)
		}
	}

	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error enconding json response")
		http.Error(w, "Internal server error", 503)
		return
	}
}

func DashboardDataValidators(w http.ResponseWriter, r *http.Request) {
	currency := GetCurrency(r)

	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	filterArr, err := parseValidatorsFromQueryString(q.Get("validators"))
	if err != nil {
		http.Error(w, "Invalid query", 400)
		return
	}
	filter := pq.Array(filterArr)

	var validators []*types.ValidatorsPageDataValidators
	err = db.DB.Select(&validators, `
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
		LIMIT 100`, filter)

	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Errorf("error retrieving validator data")
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
		http.Error(w, "Internal server error", 503)
		return
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

	earnings, err := GetValidatorEarnings(queryValidators)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Errorf("error retrieving validator earnings")
		http.Error(w, "Internal server error", 503)
	}

	err = json.NewEncoder(w).Encode(earnings)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Errorf("error enconding json response")
		http.Error(w, "Internal server error", 503)
		return
	}
}
