package handlers

import (
	"context"
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
	"math/big"
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

func parseValidatorsFromQueryString(str string, validatorLimit int) ([]uint64, error) {
	if str == "" {
		return []uint64{}, nil
	}

	strSplit := strings.Split(str, ",")
	strSplitLen := len(strSplit)

	// we only support up to [validatorLimit] validators
	if strSplitLen > validatorLimit {
		return []uint64{}, ErrTooManyValidators
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

	start := time.Now()
	if len(validators) == 0 {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error no validators provided")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	incomeData, err := db.BigtableClient.GetValidatorIncomeDetailsHistory(validators, endEpoch-100, endEpoch)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error loading validator income history data")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
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
	validatorLimit := getUserPremium(r).MaxValidators

	q := r.URL.Query()
	queryValidators, err := parseValidatorsFromQueryString(q.Get("validators"), validatorLimit)
	if err != nil && err != ErrTooManyValidators {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error parsing validators from query string")
		http.Error(w, "Invalid query", 400)
		return
	}

	dashboardData := types.DashboardData{}
	dashboardData.ValidatorLimit = validatorLimit

	epoch := services.LatestEpoch()
	dashboardData.CappellaHasHappened = epoch >= (utils.Config.Chain.Config.CappellaForkEpoch)

	dashboardData.NextWithdrawalRow, err = getNextWithdrawalRow(queryValidators)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error calculating next withdrawal row")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := InitPageData(w, r, "dashboard", "/dashboard", "Dashboard", templateFiles)
	data.Data = dashboardData

	if handleTemplateError(w, r, "dashboard.go", "Dashboard", "", dashboardTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

func getNextWithdrawalRow(queryValidators []uint64) ([][]interface{}, error) {
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
			(balance[0].EffectiveBalance == utils.Config.Chain.Config.MaxEffectiveBalance && balance[0].Balance > utils.Config.Chain.Config.MaxEffectiveBalance) {
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

	_, lastWithdrawnEpoch, err := db.GetValidatorWithdrawalsCount(nextValidator.Index)
	if err != nil {
		return nil, err
	}

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
		withdrawalAmount = nextValidator.Balance - utils.Config.Chain.Config.MaxEffectiveBalance
	}

	if lastWithdrawnEpoch == epoch || nextValidator.Balance < utils.Config.Chain.Config.MaxEffectiveBalance {
		withdrawalAmount = 0
	}

	nextData := make([][]interface{}, 0, 1)
	nextData = append(nextData, []interface{}{
		template.HTML(fmt.Sprintf("%v", utils.FormatValidator(nextValidator.Index))),
		template.HTML(fmt.Sprintf(`<span class="text-muted">~ %s</span>`, utils.FormatEpoch(uint64(utils.TimeToEpoch(timeToWithdrawal))))),
		template.HTML(fmt.Sprintf(`<span class="text-muted">~ %s</span>`, utils.FormatBlockSlot(utils.TimeToSlot(uint64(timeToWithdrawal.Unix()))))),
		template.HTML(fmt.Sprintf(`<span class="">~ %s</span>`, utils.FormatTimeFromNow(timeToWithdrawal))),
		withdrawalCredentialsTemplate,
		template.HTML(fmt.Sprintf(`<span class="text-muted"><span data-toggle="tooltip" title="If the withdrawal were to be processed at this very moment, this amount would be withdrawn"><i class="far ml-1 fa-question-circle" style="margin-left: 0px !important;"></i></span> %s</span>`, utils.FormatAmount(new(big.Int).Mul(new(big.Int).SetUint64(withdrawalAmount), big.NewInt(1e9)), "Ether", 6))),
	})

	return nextData, nil
}

// Dashboard Chart that combines balance data and
func DashboardDataBalanceCombined(w http.ResponseWriter, r *http.Request) {
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
	if len(queryValidators) < 1 {
		http.Error(w, "Invalid query", 400)
		return
	}

	g, _ := errgroup.WithContext(context.Background())
	var incomeHistoryChartData []*types.ChartDataPoint
	var executionChartData []*types.ChartDataPoint
	g.Go(func() error {
		incomeHistoryChartData, _, err = db.GetValidatorIncomeHistoryChart(queryValidators, currency)
		return err
	})

	g.Go(func() error {
		executionChartData, err = getExecutionChartData(queryValidators, currency)
		return err
	})

	err = g.Wait()
	if err != nil {
		logger.Errorf("combined balance chart %v", err)
		sendErrorResponse(w, r.URL.String(), err.Error())
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
		logger.WithError(err).WithField("route", r.URL.String()).Error("error enconding json response")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func getExecutionChartData(indices []uint64, currency string) ([]*types.ChartDataPoint, error) {
	var limit uint64 = 300
	blockList, consMap, err := findExecBlockNumbersByProposerIndex(indices, 0, limit, false)
	if err != nil {
		return nil, err
	}

	blocks, err := db.BigtableClient.GetBlocksIndexedMultiple(blockList, limit)
	if err != nil {
		return nil, err
	}
	relaysData, err := db.GetRelayDataForIndexedBlocks(blocks)
	if err != nil {
		return nil, err
	}

	var chartData = make([]*types.ChartDataPoint, len(blocks))
	epochsPerDay := utils.EpochsPerDay()

	for i := len(blocks) - 1; i >= 0; i-- {
		consData := consMap[blocks[i].Number]
		day := int64(consData.Epoch / epochsPerDay)
		color := "#90ed7d"
		totalReward, _ := utils.WeiToEther(utils.Eth1TotalReward(blocks[i])).Float64()
		relayData, ok := relaysData[common.BytesToHash(blocks[i].Hash)]
		if ok {
			totalReward, _ = utils.WeiToEther(relayData.MevBribe.BigInt()).Float64()
		}

		//balanceTs := blocks[i].GetTime().AsTime().Unix()

		chartData[len(blocks)-1-i] = &types.ChartDataPoint{
			X:     float64(utils.DayToTime(day).Unix() * 1000), //float64(balanceTs * 1000),
			Y:     utils.ExchangeRateForCurrency(currency) * totalReward,
			Color: color,
		}
	}
	return chartData, nil
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

	incomeHistoryChartData, _, err := db.GetValidatorIncomeHistoryChart(queryValidators, currency)
	if err != nil {
		logger.Errorf("failed to genereate income history chart data for dashboard view: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
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

func DashboardDataWithdrawals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	validatorLimit := getUserPremium(r).MaxValidators
	validators, err := parseValidatorsFromQueryString(q.Get("validators"), validatorLimit)
	if err != nil {
		http.Error(w, "Invalid query", 400)
		return
	}

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		utils.LogError(err, fmt.Errorf("error converting datatables data parameter from string to int: %v", err), 0)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		utils.LogError(err, fmt.Errorf("error converting datatables data parameter from string to int: %v", err), 0)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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

	withdrawalCount, err := db.GetDashboardWithdrawalsCount(validators)
	if err != nil {
		utils.LogError(err, fmt.Errorf("error retrieving dashboard validator withdrawals count: %v", err), 0)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	withdrawals, err := db.GetDashboardWithdrawals(validators, length, start, orderBy, orderDir)
	if err != nil {
		utils.LogError(err, fmt.Errorf("error retrieving validator withdrawals: %v", err), 0)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tableData := make([][]interface{}, 0, len(withdrawals))

	for _, w := range withdrawals {
		tableData = append(tableData, []interface{}{
			template.HTML(fmt.Sprintf("%v", utils.FormatValidator(w.ValidatorIndex))),
			template.HTML(fmt.Sprintf("%v", utils.FormatEpoch(utils.EpochOfSlot(w.Slot)))),
			template.HTML(fmt.Sprintf("%v", utils.FormatBlockSlot(w.Slot))),
			template.HTML(fmt.Sprintf("%v", utils.FormatTimeFromNow(utils.SlotToTime(w.Slot)))),
			template.HTML(fmt.Sprintf("%v", utils.FormatAddress(w.Address, nil, "", false, false, true))),
			template.HTML(fmt.Sprintf("%v", utils.FormatAmount(new(big.Int).Mul(new(big.Int).SetUint64(w.Amount), big.NewInt(1e9)), "Ether", 6))),
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
		utils.LogError(err, fmt.Errorf("error enconding json response for %v route: %v", r.URL.String(), err), 0)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
		SELECT
			validators.validatorindex,
			validators.pubkey,
			validators.withdrawableepoch,
			validators.slashed,
			validators.activationeligibilityepoch,
			validators.lastattestationslot,
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
		logger.WithError(err).WithField("route", r.URL.String()).Errorf("error retrieving validator data")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	balances, err := db.BigtableClient.GetValidatorBalanceHistory(filterArr, services.LatestEpoch(), services.LatestEpoch())
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

		if v.LastAttestationSlot != nil && *v.LastAttestationSlot != 0 {
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

	earnings, _, err := GetValidatorEarnings(queryValidators, GetCurrency(r))
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

	if len(activeValidators) == 0 {
		http.Error(w, "Invalid query", 400)
		return
	}

	var avgIncDistance []float64

	effectiveness, err := db.BigtableClient.GetValidatorEffectiveness(activeValidators, services.LatestEpoch()-1)
	for _, e := range effectiveness {
		avgIncDistance = append(avgIncDistance, e.AttestationEfficiency)
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
