package handlers

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"eth2-exporter/db"
	"eth2-exporter/price"
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"time"

	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lib/pq"
	"golang.org/x/sync/errgroup"
)

var ErrTooManyValidators = errors.New("too many validators")

func handleValidatorsQuery(w http.ResponseWriter, r *http.Request, checkValidatorLimit bool) ([]uint64, [][]byte, bool, error) {
	q := r.URL.Query()
	validatorLimit := getUserPremium(r).MaxValidators

	errFieldMap := map[string]interface{}{"route": r.URL.String()}

	// Parse all the validator indices and pubkeys from the query string
	queryValidatorIndices, queryValidatorPubkeys, err := parseValidatorsFromQueryString(q.Get("validators"), validatorLimit)
	if err != nil && (checkValidatorLimit || err != ErrTooManyValidators) {
		logger.Warnf("could not parse validators from query string: %v; Route: %v", err, r.URL.String())
		http.Error(w, "Invalid query", http.StatusBadRequest)
		return nil, nil, false, err
	}

	// Check whether pubkeys can be converted to indices and redirect if necessary
	redirect, err := updateValidatorsQueryString(w, r, queryValidatorIndices, queryValidatorPubkeys)
	if err != nil {
		utils.LogError(err, fmt.Errorf("error finding validators in database for dashboard query update"), 0, errFieldMap)
		http.Error(w, "Not found", http.StatusNotFound)
		return nil, nil, false, err
	}

	if !redirect {
		// Check after the redirect whether all validators are correct
		err = checkValidatorsQuery(queryValidatorIndices, queryValidatorPubkeys)
		if err != nil {
			logger.Warnf("could not find validators in database from query string: %v; Route: %v", err, r.URL.String())
			http.Error(w, "Not found", http.StatusNotFound)
			return nil, nil, false, err
		}
	}

	return queryValidatorIndices, queryValidatorPubkeys, redirect, nil
}

// parseValidatorsFromQueryString returns a slice of validator indices and a slice of validator pubkeys from a parsed query string
func parseValidatorsFromQueryString(str string, validatorLimit int) ([]uint64, [][]byte, error) {
	if str == "" {
		return []uint64{}, [][]byte{}, nil
	}

	strSplit := strings.Split(str, ",")
	strSplitLen := len(strSplit)

	// we only support up to [validatorLimit] validators
	if strSplitLen > validatorLimit {
		return []uint64{}, [][]byte{}, ErrTooManyValidators
	}

	var validatorIndices []uint64
	var validatorPubkeys [][]byte
	keys := make(map[interface{}]bool, strSplitLen)

	// Find all pubkeys
	for _, vStr := range strSplit {
		if !searchPubkeyExactRE.MatchString(vStr) {
			continue
		}
		if !strings.HasPrefix(vStr, "0x") {
			// Query string public keys have to have 0x prefix
			return []uint64{}, [][]byte{}, fmt.Errorf("invalid pubkey")
		}
		// make sure keys are unique
		if exists := keys[vStr]; exists {
			continue
		}
		keys[vStr] = true
		validatorPubkeys = append(validatorPubkeys, common.FromHex(vStr))

	}

	// Find all indices
	for _, vStr := range strSplit {
		if searchPubkeyExactRE.MatchString(vStr) {
			continue
		}
		v, err := strconv.ParseUint(vStr, 10, 64)
		if err != nil {
			return []uint64{}, [][]byte{}, err
		}
		// make sure keys are unique
		if exists := keys[v]; exists {
			continue
		}
		keys[v] = true
		validatorIndices = append(validatorIndices, v)
	}

	return validatorIndices, validatorPubkeys, nil
}

func updateValidatorsQueryString(w http.ResponseWriter, r *http.Request, validatorIndices []uint64, validatorPubkeys [][]byte) (bool, error) {
	validatorsCount := len(validatorIndices) + len(validatorPubkeys)
	if validatorsCount == 0 {
		return false, nil
	}

	// Convert pubkeys to indices if possible
	// validatorsCount stays the same after conversion
	redirect := false
	if len(validatorPubkeys) > 0 {
		validatorInfos := []struct {
			Index  uint64
			Pubkey []byte
		}{}
		err := db.ReaderDb.Select(&validatorInfos, `SELECT validatorindex as index, pubkey FROM validators WHERE pubkey = ANY($1)`, validatorPubkeys)
		if err != nil {
			return false, err
		}

		for _, info := range validatorInfos {
			// Having duplicates of validator indices is not a problem so we don't need to check for that
			validatorIndices = append(validatorIndices, info.Index)

			redirect = true
			for idx, pubkey := range validatorPubkeys {
				if bytes.Contains(pubkey, info.Pubkey) {
					validatorPubkeys = append(validatorPubkeys[:idx], validatorPubkeys[idx+1:]...)
					break
				}
			}
		}
	}

	if redirect {
		strValidators := make([]string, validatorsCount)
		for i, n := range validatorIndices {
			strValidators[i] = fmt.Sprintf("%v", n)
		}
		for i, n := range validatorPubkeys {
			strValidators[i+len(validatorIndices)] = fmt.Sprintf("%#x", n)
		}

		q := r.URL.Query()
		q.Set("validators", strings.Join(strValidators, ","))
		r.URL.RawQuery = q.Encode()

		http.Redirect(w, r, r.URL.String(), http.StatusSeeOther)
	}
	return redirect, nil
}

func checkValidatorsQuery(validatorIndices []uint64, validatorPubkeys [][]byte) error {
	validatorCount := 0

	if len(validatorIndices) > 0 {
		err := db.ReaderDb.Get(&validatorCount, `SELECT COUNT(*) FROM validators WHERE validatorindex = ANY($1)`, validatorIndices)
		if err != nil {
			return err
		}
		if validatorCount != len(validatorIndices) {
			return fmt.Errorf("invalid validator index")
		}
	}

	if len(validatorPubkeys) > 0 {
		err := db.ReaderDb.Get(&validatorCount, `SELECT COUNT(DISTINCT publickey) AS distinct_count FROM eth1_deposits WHERE publickey = ANY($1)`, validatorPubkeys)
		if err != nil {
			return err
		}
		if validatorCount != len(validatorPubkeys) {
			return fmt.Errorf("invalid validator public key")
		}
	}

	return nil
}

func Heatmap(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "heatmap.html")
	var heatmapTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")
	validatorLimit := getUserPremium(r).MaxValidators

	heatmapData := types.HeatmapData{}
	heatmapData.ValidatorLimit = validatorLimit

	min := 1
	max := 400000

	validatorCount := 100
	count, err := strconv.Atoi(r.URL.Query().Get("count"))
	if err == nil && count > 0 && count <= 1000 {
		validatorCount = count
	}

	validatorMap := make(map[uint64]bool)
	for len(validatorMap) < validatorCount {
		validatorMap[uint64(rand.Intn(max-min)+min)] = true
	}
	validators := make([]uint64, 0, len(validatorMap))
	for key := range validatorMap {
		validators = append(validators, key)
	}
	sort.Slice(validators, func(i, j int) bool { return validators[i] < validators[j] })

	validatorsCatagoryMap := make(map[uint64]int)
	for index, validator := range validators {
		validatorsCatagoryMap[validator] = index
	}
	heatmapData.Validators = validators

	endEpoch := services.LatestFinalizedEpoch()
	epochs := make([]uint64, 0, 100)
	epochsCatagoryMap := make(map[uint64]int)
	for e := endEpoch - 99; e <= endEpoch; e++ {
		epochs = append(epochs, e)
		epochsCatagoryMap[e] = len(epochs) - 1

	}
	heatmapData.Epochs = epochs

	errFieldMap := map[string]interface{}{"route": r.URL.String()}

	start := time.Now()
	if len(validators) == 0 {
		http.Error(w, "Error: No validators provided", http.StatusBadRequest)
		return
	}
	incomeData, err := db.BigtableClient.GetValidatorIncomeDetailsHistory(validators, endEpoch-100, endEpoch)
	if err != nil {
		utils.LogError(err, "error loading validator income history data", 0, errFieldMap)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	heatmapData.IncomeData = make([][3]int64, 0, validatorCount*100)
	for validator, epochs := range incomeData {
		for epoch, income := range epochs {
			income := int64(income.AttestationHeadReward+income.AttestationSourceReward+income.AttestationTargetReward) - int64(income.AttestationSourcePenalty+income.AttestationTargetPenalty)
			if income > heatmapData.MaxIncome {
				heatmapData.MaxIncome = income
			}
			if income < heatmapData.MinIncome {
				heatmapData.MinIncome = income
			}
			heatmapData.IncomeData = append(heatmapData.IncomeData, [3]int64{int64(epochsCatagoryMap[epoch]), int64(validatorsCatagoryMap[validator]), income})
		}
	}
	sort.Slice(heatmapData.IncomeData, func(i, j int) bool {
		if heatmapData.IncomeData[i][0] != heatmapData.IncomeData[j][0] {
			return heatmapData.IncomeData[i][0] < heatmapData.IncomeData[j][0]
		}
		return heatmapData.IncomeData[i][1] < heatmapData.IncomeData[j][1]
	})

	logger.Infof("retrieved income history of %v validators in %v", len(incomeData), time.Since(start))

	data := InitPageData(w, r, "dashboard", "/heatmap", "Validator Heatmap", templateFiles)
	data.Data = heatmapData

	if handleTemplateError(w, r, "dashboard.go", "Heatmap", "", heatmapTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

func Dashboard(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "dashboard.html", "dashboard/tables.html")
	var dashboardTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	_, _, redirect, err := handleValidatorsQuery(w, r, false)
	if err != nil || redirect {
		return
	}

	dashboardData := types.DashboardData{}
	dashboardData.ValidatorLimit = getUserPremium(r).MaxValidators

	epoch := services.LatestEpoch()
	dashboardData.CappellaHasHappened = epoch >= (utils.Config.Chain.ClConfig.CappellaForkEpoch)

	data := InitPageData(w, r, "dashboard", "/dashboard", "Dashboard", templateFiles)
	data.Data = dashboardData

	if handleTemplateError(w, r, "dashboard.go", "Dashboard", "", dashboardTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

func getNextWithdrawalRow(queryValidators []uint64, currency string) ([][]interface{}, error) {
	if len(queryValidators) == 0 {
		return nil, nil
	}

	stats := services.GetLatestStats()
	if stats == nil || stats.LatestValidatorWithdrawalIndex == nil || stats.TotalValidatorCount == nil {
		return nil, errors.New("stats not available")
	}

	epoch := services.LatestEpoch()

	// find subscribed validators that are active and have valid withdrawal credentials (balance will be checked later as it will be queried from bigtable)
	// order by validator index to ensure that "last withdrawal" cursor handling works
	var validatorsDb []*types.Validator
	err := db.ReaderDb.Select(&validatorsDb, `
			SELECT
				validatorindex,
				withdrawalcredentials,
				withdrawableepoch
			FROM validators
			WHERE
				activationepoch <= $1 AND exitepoch > $1 AND
				withdrawalcredentials LIKE '\x01' || '%'::bytea AND
				validatorindex = ANY($2)
			ORDER BY validatorindex ASC`, epoch, pq.Array(queryValidators))

	if err != nil {
		return nil, err
	}

	if len(validatorsDb) == 0 {
		return nil, nil
	}

	// GetValidatorBalanceHistory only takes uint64 slice
	var validatorIds = make([]uint64, 0, len(validatorsDb))
	for _, v := range validatorsDb {
		validatorIds = append(validatorIds, v.Index)
	}

	// retrieve up2date balances for all valid validators from bigtable
	balances, err := db.BigtableClient.GetValidatorBalanceHistory(validatorIds, epoch, epoch)
	if err != nil {
		return nil, err
	}

	// find the first withdrawable validator by matching validators and balances
	var nextValidator *types.Validator
	for _, v := range validatorsDb {
		balance, ok := balances[v.Index]
		if !ok {
			continue
		}
		if len(balance) == 0 {
			continue
		}

		if (balance[0].Balance > 0 && v.WithdrawableEpoch <= epoch) ||
			(balance[0].EffectiveBalance == utils.Config.Chain.ClConfig.MaxEffectiveBalance && balance[0].Balance > utils.Config.Chain.ClConfig.MaxEffectiveBalance) {
			// this validator is eligible for withdrawal, check if it is the next one
			if nextValidator == nil || v.Index > *stats.LatestValidatorWithdrawalIndex {
				nextValidator = v
				nextValidator.Balance = balance[0].Balance
				if nextValidator.Index > *stats.LatestValidatorWithdrawalIndex {
					// the first validator after the cursor has to be the next validator
					break
				}
			}
		}
	}

	if nextValidator == nil {
		return nil, nil
	}

	lastWithdrawnEpochs, err := db.GetLastWithdrawalEpoch([]uint64{nextValidator.Index})
	if err != nil {
		return nil, err
	}
	lastWithdrawnEpoch := lastWithdrawnEpochs[nextValidator.Index]

	distance, err := GetWithdrawableCountFromCursor(epoch, nextValidator.Index, *stats.LatestValidatorWithdrawalIndex)
	if err != nil {
		return nil, err
	}

	timeToWithdrawal := utils.GetTimeToNextWithdrawal(distance)

	// it normally takes two epochs to finalize
	latestFinalized := services.LatestFinalizedEpoch()
	if timeToWithdrawal.Before(utils.EpochToTime(epoch + (epoch - latestFinalized))) {
		return nil, nil
	}

	var withdrawalCredentialsTemplate template.HTML
	address, err := utils.WithdrawalCredentialsToAddress(nextValidator.WithdrawalCredentials)
	if err != nil {
		// warning only as "N/A" will be displayed
		logger.Warn("invalid withdrawal credentials")
	}
	if address != nil {
		withdrawalCredentialsTemplate = template.HTML(fmt.Sprintf(`<a href="/address/0x%x"><span class="text-muted">%s</span></a>`, address, utils.FormatAddress(address, nil, "", false, false, true)))
	} else {
		withdrawalCredentialsTemplate = `<span class="text-muted">N/A</span>`
	}

	var withdrawalAmount uint64
	if nextValidator.WithdrawableEpoch <= epoch {
		// full withdrawal
		withdrawalAmount = nextValidator.Balance
	} else {
		// partial withdrawal
		withdrawalAmount = nextValidator.Balance - utils.Config.Chain.ClConfig.MaxEffectiveBalance
	}

	if lastWithdrawnEpoch == epoch || nextValidator.Balance < utils.Config.Chain.ClConfig.MaxEffectiveBalance {
		withdrawalAmount = 0
	}

	nextData := make([][]interface{}, 0, 1)
	nextData = append(nextData, []interface{}{
		utils.FormatValidator(nextValidator.Index),
		template.HTML(fmt.Sprintf(`<span class="text-muted">~ %s</span>`, utils.FormatEpoch(uint64(utils.TimeToEpoch(timeToWithdrawal))))),
		template.HTML(fmt.Sprintf(`<span class="text-muted">~ %s</span>`, utils.FormatBlockSlot(utils.TimeToSlot(uint64(timeToWithdrawal.Unix()))))),
		template.HTML(fmt.Sprintf(`<span class="">~ %s</span>`, utils.FormatTimestamp(timeToWithdrawal.Unix()))),
		withdrawalCredentialsTemplate,
		template.HTML(fmt.Sprintf(`<span class="text-muted"><span data-toggle="tooltip" title="If the withdrawal were to be processed at this very moment, this amount would be withdrawn"><i class="far ml-1 fa-question-circle" style="margin-left: 0px !important;"></i></span> %s</span>`, utils.FormatClCurrency(withdrawalAmount, currency, 6, true, false, false, true))),
	})

	return nextData, nil
}

// Dashboard Chart that combines balance data and
func DashboardDataBalanceCombined(w http.ResponseWriter, r *http.Request) {
	var lowerBoundDay uint64
	param := r.URL.Query().Get("days")
	if len(param) != 0 {
		days, err := strconv.ParseUint(param, 10, 32)
		if err != nil {
			logger.Warnf("error parsing days: %v", err)
			http.Error(w, "Error: invalid parameter days", http.StatusBadRequest)
			return
		}
		lastStatsDay, err := services.LatestExportedStatisticDay()
		if days < lastStatsDay && err == nil {
			lowerBoundDay = lastStatsDay - days + 1
		}
	}

	currency := GetCurrency(r)
	errFieldMap := map[string]interface{}{"route": r.URL.String()}

	w.Header().Set("Content-Type", "application/json")

	queryValidatorIndices, _, redirect, err := handleValidatorsQuery(w, r, true)
	if err != nil || redirect {
		return
	}

	if len(queryValidatorIndices) < 1 {
		http.Error(w, "Invalid query", http.StatusBadRequest)
		return
	}

	g, _ := errgroup.WithContext(context.Background())
	var incomeHistoryChartData []*types.ChartDataPoint
	var executionChartData []*types.ChartDataPoint
	g.Go(func() error {
		incomeHistoryChartData, err = db.GetValidatorIncomeHistoryChart(queryValidatorIndices, currency, services.LatestFinalizedEpoch(), lowerBoundDay)
		if err != nil {
			return fmt.Errorf("error in GetValidatorIncomeHistoryChart: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		executionChartData, err = getExecutionChartData(queryValidatorIndices, currency, lowerBoundDay)
		if err != nil {
			return fmt.Errorf("error in getExecutionChartData: %w", err)
		}
		return nil
	})

	err = g.Wait()
	if err != nil {
		utils.LogError(err, "error while combining balance chart", 0, errFieldMap)
		SendBadRequestResponse(w, r.URL.String(), err.Error())
		return
	}

	var response struct {
		ConsensusChartData []*types.ChartDataPoint `json:"consensusChartData"`
		ExecutionChartData []*types.ChartDataPoint `json:"executionChartData"`
	}
	response.ConsensusChartData = incomeHistoryChartData
	response.ExecutionChartData = executionChartData

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.LogError(err, "error enconding json response", 0, errFieldMap)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// DashboardDataBalance retrieves the income history of a set of validators
func DashboardDataBalance(w http.ResponseWriter, r *http.Request) {
	currency := GetCurrency(r)
	errFieldMap := map[string]interface{}{"route": r.URL.String()}

	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	validatorLimit := getUserPremium(r).MaxValidators
	queryValidatorIndices, queryValidatorPubkeys, err := parseValidatorsFromQueryString(q.Get("validators"), validatorLimit)
	if err != nil || len(queryValidatorPubkeys) > 0 {
		utils.LogError(err, "error parsing validators from query string", 0, errFieldMap)
		http.Error(w, "Invalid query", http.StatusBadRequest)
		return
	}
	if len(queryValidatorIndices) < 1 {
		http.Error(w, "Invalid query", http.StatusBadRequest)
		return
	}

	incomeHistoryChartData, err := db.GetValidatorIncomeHistoryChart(queryValidatorIndices, currency, services.LatestFinalizedEpoch(), 0)
	if err != nil {
		utils.LogError(err, "failed to genereate income history chart data for dashboard view", 0, errFieldMap)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(incomeHistoryChartData)
	if err != nil {
		utils.LogError(err, "error enconding json response", 0, errFieldMap)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func DashboardDataProposals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	filterArr, _, redirect, err := handleValidatorsQuery(w, r, true)
	if err != nil || redirect {
		return
	}

	filter := pq.Array(filterArr)

	proposals := []struct {
		Slot   uint64
		Status uint64
	}{}

	errFieldMap := map[string]interface{}{"route": r.URL.String()}

	err = db.ReaderDb.Select(&proposals, `
		SELECT slot, status
		FROM blocks
		WHERE proposer = ANY($1)
		ORDER BY slot`, filter)
	if err != nil {
		utils.LogError(err, "error retrieving block-proposals", 0, errFieldMap)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
		utils.LogError(err, "error enconding json response", 0, errFieldMap)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func DashboardDataWithdrawals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	reqCurrency := GetCurrency(r)
	q := r.URL.Query()

	validatorIndices, _, redirect, err := handleValidatorsQuery(w, r, true)
	if err != nil || redirect {
		return
	}

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables draw parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter draw", http.StatusBadRequest)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter start", http.StatusBadRequest)
		return
	}

	orderColumn := q.Get("order[0][column]")
	orderByMap := map[string]string{
		"0": "validatorindex",
		"1": "block_slot",
		"2": "block_slot",
		"3": "withdrawalindex",
		"4": "address",
		"5": "amount",
	}
	orderBy, exists := orderByMap[orderColumn]
	if !exists {
		orderBy = "validatorindex"
	}
	orderDir := q.Get("order[0][dir]")
	if orderDir != "asc" {
		orderDir = "desc"
	}

	length := uint64(10)

	errFieldMap := map[string]interface{}{"route": r.URL.String()}

	withdrawalCount, err := db.GetTotalWithdrawalsCount(validatorIndices)
	if err != nil {
		utils.LogError(err, "error retrieving dashboard validator withdrawals count", 0, errFieldMap)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	withdrawals, err := db.GetDashboardWithdrawals(validatorIndices, length, start, orderBy, orderDir)
	if err != nil {
		utils.LogError(err, "error retrieving validator withdrawals", 0, errFieldMap)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var tableData [][]interface{}

	// check if there is a NextWithdrawal and append
	NextWithdrawalRow, err := getNextWithdrawalRow(validatorIndices, reqCurrency)
	if err != nil {
		utils.LogError(err, "error calculating next withdrawal row", 0, errFieldMap)
		tableData = make([][]interface{}, 0, len(withdrawals))
	} else {
		if NextWithdrawalRow == nil {
			tableData = make([][]interface{}, 0, len(withdrawals))
		} else {
			// make the array +1 larger to append the NextWithdrawal row
			tableData = make([][]interface{}, 0, len(withdrawals)+1)
			tableData = append(NextWithdrawalRow, tableData...)
		}
	}

	for _, w := range withdrawals {
		tableData = append(tableData, []interface{}{
			utils.FormatValidator(w.ValidatorIndex),
			utils.FormatEpoch(utils.EpochOfSlot(w.Slot)),
			utils.FormatBlockSlot(w.Slot),
			utils.FormatTimestamp(utils.SlotToTime(w.Slot).Unix()),
			utils.FormatAddress(w.Address, nil, "", false, false, true),
			utils.FormatClCurrency(w.Amount, reqCurrency, 6, true, false, false, true),
		})
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    withdrawalCount,
		RecordsFiltered: withdrawalCount,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		utils.LogError(err, "error enconding json response", 0, errFieldMap)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func DashboardDataValidators(w http.ResponseWriter, r *http.Request) {
	currency := GetCurrency(r)

	w.Header().Set("Content-Type", "application/json")

	validatorIndexArr, validatorPubkeyArr, redirect, err := handleValidatorsQuery(w, r, true)
	if err != nil || redirect {
		return
	}

	errFieldMap := map[string]interface{}{"route": r.URL.String()}

	filter := pq.Array(validatorIndexArr)
	validatorLimit := getUserPremium(r).MaxValidators

	var validatorsByIndex []*types.ValidatorsData
	err = db.ReaderDb.Select(&validatorsByIndex, `
		SELECT
			validators.validatorindex,
			validators.pubkey,
			validators.withdrawableepoch,
			validators.slashed,
			validators.activationeligibilityepoch,
			validators.activationepoch,
			validators.exitepoch,
			(SELECT COUNT(*) FROM blocks WHERE proposer = validators.validatorindex AND status = '1') as executedproposals,
			(SELECT COUNT(*) FROM blocks WHERE proposer = validators.validatorindex AND status = '2') as missedproposals,
			COALESCE(validator_performance.cl_performance_7d, 0) as performance7d,
			COALESCE(validator_names.name, '') AS name,
		    validators.status AS state
		FROM validators
		LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
		LEFT JOIN validator_performance ON validators.validatorindex = validator_performance.validatorindex
		WHERE validators.validatorindex = ANY($1)
		LIMIT $2`, filter, validatorLimit)

	if err != nil {
		utils.LogError(err, "error retrieving validator data", 0, errFieldMap)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	validatorsByIndexPubKeys := make([][]byte, len(validatorsByIndex))
	for idx := range validatorsByIndex {
		validatorsByIndexPubKeys[idx] = validatorsByIndex[idx].PublicKey
	}
	pubkeyFilter := pq.ByteaArray(validatorsByIndexPubKeys)

	validatorsDeposits := []struct {
		Pubkey  []byte `db:"publickey"`
		Address []byte `db:"from_address"`
	}{}
	err = db.ReaderDb.Select(&validatorsDeposits, `
		SELECT
			publickey,
			from_address
		FROM eth1_deposits
		WHERE publickey = ANY($1)`, pubkeyFilter)
	if err != nil {
		utils.LogError(err, "error retrieving validator deposists", 0, errFieldMap)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	validatorsDepositsMap := make(map[string][]string)
	for _, deposit := range validatorsDeposits {
		key := hex.EncodeToString(deposit.Pubkey)
		if _, ok := validatorsDepositsMap[key]; !ok {
			validatorsDepositsMap[key] = make([]string, 0)
		}
		validatorsDepositsMap[key] = append(validatorsDepositsMap[key], fmt.Sprintf("%#x", deposit.Address))
	}

	latestEpoch := services.LatestEpoch()

	stats := services.GetLatestStats()
	churnRate := stats.ValidatorChurnLimit

	if len(validatorIndexArr) > 0 {
		balances, err := db.BigtableClient.GetValidatorBalanceHistory(validatorIndexArr, latestEpoch, latestEpoch)
		if err != nil {
			utils.LogError(err, "error retrieving validator balance data", 0, errFieldMap)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		for _, validator := range validatorsByIndex {
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

		lastAttestationSlots, err := db.BigtableClient.GetLastAttestationSlots(validatorIndexArr)
		if err != nil {
			utils.LogError(err, "error retrieving validator last attestation slot data", 0, errFieldMap)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		for _, validator := range validatorsByIndex {
			validator.LastAttestationSlot = int64(lastAttestationSlots[validator.ValidatorIndex])
		}
	}

	validatorsByPubkey := make([]*types.ValidatorsData, len(validatorPubkeyArr))
	for i := range validatorsByPubkey {
		// Validators without an index don't have  activation, exit and withdrawable epochs yet.
		// Show them as pending even if they are still in the state "Deposited".
		validatorsByPubkey[i] = &types.ValidatorsData{
			PublicKey:         validatorPubkeyArr[i],
			ActivationEpoch:   math.MaxInt64,
			ExitEpoch:         math.MaxInt64,
			WithdrawableEpoch: math.MaxInt64,
			State:             "pending_deposited",
		}
	}

	validators := append(validatorsByIndex, validatorsByPubkey...)

	tableData := make([][]interface{}, len(validators))
	for i, v := range validators {
		indexInfo := fmt.Sprintf("%v", v.ValidatorIndex)
		if i >= len(validatorsByIndex) {
			// If the validator does not have an index yet show custom text that is like the state
			indexInfo = "Pending"
		}
		var queueAhead uint64
		var estimatedActivationTs time.Time
		if v.State == "pending" {
			if v.ActivationEpoch > 100_000_000 {
				queueAhead, err = db.GetQueueAheadOfValidator(v.ValidatorIndex)
				if err != nil {
					utils.LogError(err, fmt.Sprintf("failed to retrieve queue ahead of validator %v for dashboard", v.ValidatorIndex), 0, errFieldMap)
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
				epochsToWait := queueAhead / *churnRate
				// calculate dequeue epoch
				estimatedActivationEpoch := latestEpoch + epochsToWait + 1
				// add activation offset
				estimatedActivationEpoch += utils.Config.Chain.ClConfig.MaxSeedLookahead + 1
				estimatedActivationTs = utils.EpochToTime(estimatedActivationEpoch)
			} else {
				queueAhead = 0
				estimatedActivationTs = utils.EpochToTime(v.ActivationEpoch)
			}
		}

		tableData[i] = []interface{}{
			fmt.Sprintf("%x", v.PublicKey),
			indexInfo,
			[]interface{}{
				fmt.Sprintf("%.4f %v", float64(v.CurrentBalance)/float64(1e9)*price.GetPrice(utils.Config.Frontend.ClCurrency, currency), currency),
				fmt.Sprintf("%.1f %v", float64(v.EffectiveBalance)/float64(1e9)*price.GetPrice(utils.Config.Frontend.ClCurrency, currency), currency),
			},
			[]interface{}{
				v.ValidatorIndex,
				v.State,
				queueAhead + 1,
				estimatedActivationTs.Unix()},
		}

		if v.ActivationEpoch != math.MaxInt64 {
			tableData[i] = append(tableData[i], []interface{}{
				v.ActivationEpoch,
				utils.EpochToTime(v.ActivationEpoch).Unix(),
			})
		} else {
			tableData[i] = append(tableData[i], nil)
		}

		if v.ExitEpoch != math.MaxInt64 {
			tableData[i] = append(tableData[i], []interface{}{
				v.ExitEpoch,
				utils.EpochToTime(v.ExitEpoch).Unix(),
			})
		} else {
			tableData[i] = append(tableData[i], nil)
		}

		if v.WithdrawableEpoch != math.MaxInt64 {
			tableData[i] = append(tableData[i], []interface{}{
				v.WithdrawableEpoch,
				utils.EpochToTime(v.WithdrawableEpoch).Unix(),
			})
		} else {
			tableData[i] = append(tableData[i], nil)
		}

		if v.LastAttestationSlot != 0 {
			tableData[i] = append(tableData[i], []interface{}{
				v.LastAttestationSlot,
				utils.FormatTimestamp(utils.SlotToTime(uint64(v.LastAttestationSlot)).Unix()),
				//utils.SlotToTime(uint64(*v.LastAttestationSlot)).Unix(),
			})
		} else {
			tableData[i] = append(tableData[i], nil)
		}

		tableData[i] = append(tableData[i], []interface{}{
			v.ExecutedProposals,
			v.MissedProposals,
		})

		tableData[i] = append(tableData[i], utils.FormatIncome(v.Performance7d, currency, true))

		validatorDeposits := validatorsDepositsMap[hex.EncodeToString(v.PublicKey)]
		if validatorDeposits != nil {
			tableData[i] = append(tableData[i], validatorDeposits)
		} else {
			tableData[i] = append(tableData[i], nil)
		}

	}

	type dataType struct {
		LatestEpoch uint64          `json:"latestEpoch"`
		Data        [][]interface{} `json:"data"`
	}
	data := &dataType{
		LatestEpoch: latestEpoch,
		Data:        tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		utils.LogError(err, "error enconding json response", 0, errFieldMap)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func DashboardDataEarnings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	errFieldMap := map[string]interface{}{"route": r.URL.String()}

	queryValidatorIndices, _, redirect, err := handleValidatorsQuery(w, r, true)
	if err != nil || redirect {
		return
	}

	earnings, _, err := GetValidatorEarnings(queryValidatorIndices, GetCurrency(r))
	if err != nil {
		utils.LogError(err, "error retrieving validator earnings", 0, errFieldMap)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}

	if earnings == nil {
		earnings = &types.ValidatorEarnings{}
	}

	err = json.NewEncoder(w).Encode(earnings)
	if err != nil {
		utils.LogError(err, "error enconding json response", 0, errFieldMap)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func DashboardDataEffectiveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	filterArr, _, redirect, err := handleValidatorsQuery(w, r, true)
	if err != nil || redirect {
		return
	}

	errFieldMap := map[string]interface{}{"route": r.URL.String()}

	filter := pq.Array(filterArr)

	var activeValidators []uint64
	err = db.ReaderDb.Select(&activeValidators, `
		SELECT validatorindex FROM validators where validatorindex = ANY($1) and activationepoch < $2 AND exitepoch > $2
	`, filter, services.LatestEpoch())
	if err != nil {
		utils.LogError(err, "error retrieving active validators", 0, errFieldMap)
	}

	if len(activeValidators) == 0 {
		// valid 200 response with empty data
		w.Write([]byte(`{}`))
		return
	}

	var avgIncDistance []float64

	epoch := services.LatestEpoch()
	if epoch > 0 {
		epoch = epoch - 1
	}

	effectiveness, err := db.BigtableClient.GetValidatorEffectiveness(activeValidators, epoch)
	if err != nil {
		utils.LogError(err, "error retrieving validator effectiveness", 0, errFieldMap)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	for _, e := range effectiveness {
		avgIncDistance = append(avgIncDistance, e.AttestationEfficiency)
	}

	err = json.NewEncoder(w).Encode(avgIncDistance)
	if err != nil {
		utils.LogError(err, "error enconding json response", 0, errFieldMap)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func DashboardDataProposalsHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	errFieldMap := map[string]interface{}{"route": r.URL.String()}

	lastDay, err := db.GetLastExportedStatisticDay()
	if err != nil && err != db.ErrNoStats {
		utils.LogError(err, "error retrieving last exported statistic day", 0, errFieldMap)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	dayStart, err := strconv.Atoi(r.URL.Query().Get("start_day"))
	if err != nil {
		dayStart = 0
	}
	dayEnd, err := strconv.Atoi(r.URL.Query().Get("end_day"))
	if err != nil {
		dayEnd = int(lastDay)
	}

	filterArr, _, redirect, err := handleValidatorsQuery(w, r, true)
	if err != nil || redirect {
		return
	}

	allowedDayRange := utils.GetMaxAllowedDayRangeValidatorStats(len(filterArr))

	if dayEnd < dayStart {
		http.Error(w, "Error: Invalid day range", http.StatusBadRequest)
		return
	}

	if dayEnd-dayStart > allowedDayRange {
		dayStart = dayEnd - allowedDayRange
	}

	filter := pq.Array(filterArr)

	proposals := []struct {
		ValidatorIndex uint64  `db:"validatorindex"`
		Day            int64   `db:"day"`
		Proposed       *uint64 `db:"proposed_blocks"`
		Missed         *uint64 `db:"missed_blocks"`
		Orphaned       *uint64 `db:"orphaned_blocks"`
	}{}
	todaysProposals := proposals

	dayFilter := "day >= $2 AND day <= $3"
	args := []interface{}{filter, dayStart, dayEnd}
	if allowedDayRange == 0 {
		dayFilter = "day = $2"
		args = []interface{}{filter, dayStart}
	}

	err = db.ReaderDb.Select(&proposals, fmt.Sprintf(`
		SELECT validatorindex, day, proposed_blocks, missed_blocks, orphaned_blocks
		FROM validator_stats
		WHERE validatorindex = ANY($1) 
		AND (proposed_blocks > 0 OR missed_blocks > 0 OR orphaned_blocks > 0)
		AND %v
		ORDER BY day DESC`, dayFilter), args...)
	if err != nil {
		utils.LogError(err, "error retrieving validator_stats", 0, errFieldMap)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if uint64(dayEnd) > lastDay {
		_, lastExportedEpoch := utils.GetFirstAndLastEpochForDay(lastDay)

		err = db.ReaderDb.Select(&todaysProposals, `
		SELECT
			proposer as validatorindex,
			SUM(CASE WHEN status = '1' THEN 1 ELSE 0 END) as proposed_blocks,
			SUM(CASE WHEN status = '2' THEN 1 ELSE 0 END) as missed_blocks,
			SUM(CASE WHEN status = '3' THEN 1 ELSE 0 END) as orphaned_blocks
		FROM blocks
		WHERE proposer = ANY($1) AND epoch > $2
		group by proposer`, filter, lastExportedEpoch)
		if err != nil {
			utils.LogError(err, "error retrieving validator_stats", 0, errFieldMap)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		for i := range todaysProposals {
			todaysProposals[i].Day = int64(lastDay + 1)
		}

		proposals = append(todaysProposals, proposals...)
	}

	proposalsHistResult := make([][]uint64, len(proposals))
	for i, proposal := range proposals {
		var proposed, missed, orphaned uint64 = 0, 0, 0
		if proposal.Proposed != nil {
			proposed = *proposal.Proposed
		}
		if proposal.Missed != nil {
			missed = *proposal.Missed
		}
		if proposal.Orphaned != nil {
			orphaned = *proposal.Orphaned
		}
		proposalsHistResult[i] = []uint64{
			proposal.ValidatorIndex,
			uint64(utils.DayToTime(proposal.Day).Unix()),
			proposed,
			missed,
			orphaned,
		}
	}

	responseStruct := struct {
		StartDay int64      `json:"start_day"`
		EndDay   int64      `json:"end_day"`
		Data     [][]uint64 `json:"data"`
	}{int64(dayStart), int64(dayEnd), proposalsHistResult}

	err = json.NewEncoder(w).Encode(responseStruct)
	if err != nil {
		utils.LogError(err, "error enconding json response", 0, errFieldMap)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
