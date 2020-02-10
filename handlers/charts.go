package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/prysmaticlabs/prysm/shared/mathutil"
)

var chartsTemplate = template.Must(template.New("charts").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/charts.html"))
var genericChartTemplate = template.Must(template.New("chart").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/genericchart.html"))

type chartHandler struct {
	Order    int
	DataFunc func() (*types.GenericChartData, error)
}

var chartHandlers = map[string]chartHandler{
	"blocks":                     chartHandler{1, blocksChartData},
	"validators":                 chartHandler{2, activeValidatorsChartData},
	"staked_ether":               chartHandler{3, stakedEtherChartData},
	"average_balance":            chartHandler{4, averageBalanceChartData},
	"network_liveness":           chartHandler{5, networkLivenessChartData},
	"participation_rate":         chartHandler{6, participationRateChartData},
	"estimated_validator_return": chartHandler{7, estimatedValidatorReturnChartData},
	"stake_effectiveness":        chartHandler{8, stakeEffectivenessChartData},
}

// Charts uses a go template for presenting the page to show charts
func Charts(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html")

	type chartHandlerRes struct {
		Order int
		Path  string
		Data  *types.GenericChartData
		Error error
	}

	wg := sync.WaitGroup{}
	wg.Add(len(chartHandlers))

	chartHandlerResChan := make(chan chartHandlerRes, len(chartHandlers))

	for i, ch := range chartHandlers {
		go func(i string, ch chartHandler) {
			defer wg.Done()
			data, err := ch.DataFunc()
			if err != nil {
				logger.Errorf("error getting chart data for %v: %w", i, err)
			}
			chartHandlerResChan <- chartHandlerRes{ch.Order, i, data, err}
		}(i, ch)
	}

	go func() {
		wg.Wait()
		close(chartHandlerResChan)
	}()

	pageCharts := []chartHandlerRes{}

	for chart := range chartHandlerResChan {
		if chart.Error != nil {
			http.Error(w, "Internal server error", 503)
			return
		}
		pageCharts = append(pageCharts, chart)
	}

	sort.Slice(pageCharts, func(i, j int) bool {
		return pageCharts[i].Order < pageCharts[j].Order
	})

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - Charts - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/charts",
		},
		ShowSyncingMessage:    services.IsSyncing(),
		Active:                "charts",
		Data:                  nil,
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
	}

	err := chartsTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

func GenericChart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	vars := mux.Vars(r)
	chartVar := vars["chart"]
	chartHandler, exists := chartHandlers[chartVar]
	if !exists {
		http.Error(w, "Internal server error", 503)
		return
	}

	chartData, err := chartHandler.DataFunc()
	if err != nil {
		logger.Errorf("error retrieving chart data for route %v: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - %v Chart - beaconcha.in - %v", chartData.Title, utils.Config.Frontend.SiteName, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/charts/" + chartVar,
		},
		ShowSyncingMessage:    services.IsSyncing(),
		Active:                "charts",
		Data:                  nil,
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
	}

	err = genericChartTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

func blocksChartData() (*types.GenericChartData, error) {
	rows := []struct {
		Epoch     uint64
		Status    uint64
		NbrBlocks uint64
	}{}

	err := db.DB.Select(&rows, "SELECT epoch, status, count(*) as nbrBlocks FROM blocks GROUP BY epoch, status ORDER BY epoch")
	if err != nil {
		return nil, err
	}

	dailyProposedBlocks := [][]float64{}
	dailyMissedBlocks := [][]float64{}
	dailyOrphanedBlocks := [][]float64{}

	for _, row := range rows {
		day := float64(utils.EpochToTime(row.Epoch).Truncate(time.Hour*24).Unix() * 1000)

		if row.Status == 1 {
			if len(dailyProposedBlocks) == 0 || dailyProposedBlocks[len(dailyProposedBlocks)-1][0] != day {
				dailyProposedBlocks = append(dailyProposedBlocks, []float64{day, float64(row.NbrBlocks)})
			} else {
				dailyProposedBlocks[len(dailyProposedBlocks)-1][1] += float64(row.NbrBlocks)
			}
		}

		if row.Status == 2 {
			if len(dailyMissedBlocks) == 0 || dailyMissedBlocks[len(dailyMissedBlocks)-1][0] != day {
				dailyMissedBlocks = append(dailyMissedBlocks, []float64{day, float64(row.NbrBlocks)})
			} else {
				dailyMissedBlocks[len(dailyMissedBlocks)-1][1] += float64(row.NbrBlocks)
			}
		}

		if row.Status == 3 {
			if len(dailyOrphanedBlocks) == 0 || dailyOrphanedBlocks[len(dailyOrphanedBlocks)-1][0] != day {
				dailyOrphanedBlocks = append(dailyOrphanedBlocks, []float64{day, float64(row.NbrBlocks)})
			} else {
				dailyOrphanedBlocks[len(dailyOrphanedBlocks)-1][1] += float64(row.NbrBlocks)
			}
		}
	}

	chartData := &types.GenericChartData{
		Title:        "Blocks",
		Subtitle:     "History of daily blocks proposed",
		XAxisTitle:   "",
		YAxisTitle:   "# of Blocks",
		StackingMode: "normal",
		Type:         "column",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Proposed",
				Data: dailyProposedBlocks,
			},
			{
				Name: "Missed",
				Data: dailyMissedBlocks,
			},
			{
				Name: "Orphaned",
				Data: dailyOrphanedBlocks,
			},
		},
	}

	return chartData, nil
}

// ActiveValidatorChart will show the Active Validators Chart
func ActiveValidatorChart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - Active Validators Chart - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/charts",
		},
		ShowSyncingMessage:    services.IsSyncing(),
		Active:                "charts",
		Data:                  nil,
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
	}

	rows := []struct {
		Epoch           uint64
		ValidatorsCount uint64
	}{}

	err := db.DB.Select(&rows, "SELECT epoch, validatorscount FROM epochs ORDER BY epoch")
	if err != nil {
		return nil, err
	}

	dailyActiveValidators := [][]float64{}

	for _, row := range rows {
		day := float64(utils.EpochToTime(row.Epoch).Truncate(time.Hour*24).Unix() * 1000)

		if len(dailyActiveValidators) == 0 || dailyActiveValidators[len(dailyActiveValidators)-1][0] != day {
			dailyActiveValidators = append(dailyActiveValidators, []float64{day, float64(row.ValidatorsCount)})
		}
	}

	chartData := &types.GenericChartData{
		Title:        "Validators",
		Subtitle:     "History of daily active validators",
		XAxisTitle:   "",
		YAxisTitle:   "# of Validators",
		StackingMode: "false",
		Type:         "column",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Validators",
				Data: dailyActiveValidators,
			},
		},
	}

	return chartData, nil
}

// StakedEtherChart will show the Staked Ether Chart
func StakedEtherChart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - Staked Ether Chart - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/charts",
		},
		ShowSyncingMessage:    services.IsSyncing(),
		Active:                "charts",
		Data:                  nil,
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
	}

	rows := []struct {
		Epoch         uint64
		EligibleEther uint64
	}{}

	err := db.DB.Select(&rows, "SELECT epoch, eligibleether FROM epochs ORDER BY epoch")
	if err != nil {
		return nil, err
	}

	dailyStakedEther := [][]float64{}

	for _, row := range rows {
		day := float64(utils.EpochToTime(row.Epoch).Truncate(time.Hour*24).Unix() * 1000)

		if len(dailyStakedEther) == 0 || dailyStakedEther[len(dailyStakedEther)-1][0] != day {
			dailyStakedEther = append(dailyStakedEther, []float64{day, float64(row.EligibleEther) / 1000000000})
		}
	}

	chartData := &types.GenericChartData{
		Title:        "Staked Ether",
		Subtitle:     "History of daily staked Ether",
		XAxisTitle:   "",
		YAxisTitle:   "Ether",
		StackingMode: "false",
		Type:         "column",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Staked Ether",
				Data: dailyStakedEther,
			},
		},
	}

	return chartData, nil
}

// AverageBalanceChart will show the Average Validator Balance Chart
func AverageBalanceChart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - Average Validator Balance Chart - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/charts",
		},
		ShowSyncingMessage:    services.IsSyncing(),
		Active:                "charts",
		Data:                  nil,
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
	}

	rows := []struct {
		Epoch                   uint64
		AverageValidatorBalance uint64
	}{}

	err := db.DB.Select(&rows, "SELECT epoch, averagevalidatorbalance FROM epochs ORDER BY epoch")
	if err != nil {
		return nil, err
	}

	dailyAverageBalance := [][]float64{}

	for _, row := range rows {
		day := float64(utils.EpochToTime(row.Epoch).Truncate(time.Hour*24).Unix() * 1000)

		if len(dailyAverageBalance) == 0 || dailyAverageBalance[len(dailyAverageBalance)-1][0] != day {
			dailyAverageBalance = append(dailyAverageBalance, []float64{day, float64(row.AverageValidatorBalance) / 1000000000})
		}
	}

	chartData := &types.GenericChartData{
		Title:        "Validator Balance",
		Subtitle:     "History of the daily average validator balance",
		XAxisTitle:   "",
		YAxisTitle:   "Ether",
		StackingMode: "false",
		Type:         "column",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Average Balance [ETH]",
				Data: dailyAverageBalance,
			},
		},
	}

	return chartData, nil
}

func networkLivenessChartData() (*types.GenericChartData, error) {
	rows := []struct {
		Timestamp      uint64
		HeadEpoch      uint64
		FinalizedEpoch uint64
	}{}

	err := db.DB.Select(&rows, "SELECT EXTRACT(epoch FROM ts)::INT AS timestamp, headepoch, finalizedepoch FROM network_liveness ORDER BY ts")
	if err != nil {
		return nil, err
	}

	seriesData := [][]float64{}

	for _, row := range rows {
		// networkliveness := (1 - 4*float64(row.HeadEpoch-2-row.FinalizedEpoch)/100)
		// if networkliveness < 0 {
		// 	networkliveness = 0
		// }
		seriesData = append(seriesData, []float64{
			float64(row.Timestamp * 1000),
			float64(row.HeadEpoch - row.FinalizedEpoch),
		})
	}

	chartData := &types.GenericChartData{
		Title:        "Network Liveness",
		Subtitle:     "History of how far the last Finalized Epoch is behind the Head Epoch",
		XAxisTitle:   "",
		YAxisTitle:   "Network Liveness [epochs]",
		StackingMode: "false",
		Type:         "column",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Network Liveness",
				Data: seriesData,
			},
		},
	}

	return chartData, nil
}

func participationRateChartData() (*types.GenericChartData, error) {
	rows := []struct {
		Epoch                   uint64
		Globalparticipationrate float64
	}{}

	err := db.DB.Select(&rows, "SELECT epoch, globalparticipationrate FROM epochs WHERE epoch < $1 ORDER BY epoch", services.LatestEpoch())
	if err != nil {
		return nil, err
	}

	seriesData := [][]float64{}

	for _, row := range rows {
		seriesData = append(seriesData, []float64{
			float64(utils.EpochToTime(row.Epoch).Unix() * 1000),
			row.Globalparticipationrate * 100,
		})
	}

	chartData := &types.GenericChartData{
		Title:        "Participation Rate",
		Subtitle:     "History of the Participation Rate",
		XAxisTitle:   "",
		YAxisTitle:   "Participation Rate [%]",
		StackingMode: "false",
		Type:         "line",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Participation Rate",
				Data: seriesData,
			},
		},
	}

	return chartData, nil
}

func estimatedValidatorReturnChartData() (*types.GenericChartData, error) {
	rows := []struct {
		Epoch           uint64
		Eligibleether   uint64
		Votedether      uint64
		Validatorscount uint64
	}{}

	// note: eligibleether might not be correct, need to check what exactly the node returns
	// for the reward-calculation we need the sum of all effective balances
	err := db.DB.Select(&rows, `SELECT epoch, eligibleether, votedether, validatorscount FROM epochs ORDER BY epoch`)
	if err != nil {
		return nil, err
	}

	seriesData := [][]float64{}

	// see: https://github.com/ethereum/eth2.0-specs/blob/dev/specs/phase0/beacon-chain.md#rewards-and-penalties-1
	maxEffectiveBalance := uint64(32e8)
	baseRewardFactor := uint64(64)
	baseRewardPerEpoch := uint64(4)
	proposerRewardQuotient := uint64(8)
	slotsPerDay := 3600 * 24 / utils.Config.Chain.SecondsPerSlot
	epochsPerDay := slotsPerDay / utils.Config.Chain.SlotsPerEpoch

	for _, row := range rows {
		if row.Eligibleether == 0 {
			continue
		}

		baseReward := maxEffectiveBalance * baseRewardFactor / mathutil.IntegerSquareRoot(row.Eligibleether) / baseRewardPerEpoch
		// Micro-incentives for matching FFG source, FFG target, and head
		estimatedRewardPerDay := epochsPerDay * 3 * baseReward * row.Votedether / row.Eligibleether
		// Proposer and inclusion delay micro-rewards
		proposerReward := baseReward / proposerRewardQuotient
		estimatedRewardPerDay += epochsPerDay * (baseReward - proposerReward)
		proposalsPerDay := slotsPerDay / row.Validatorscount
		estimatedRewardPerDay += proposalsPerDay * proposerReward

		seriesData = append(seriesData, []float64{
			float64(utils.EpochToTime(row.Epoch).Unix() * 1000),
			float64(estimatedRewardPerDay) / 1e9,
		})
	}

	chartData := &types.GenericChartData{
		Title:        "Estimated Validator Return",
		Subtitle:     "History of the Estimated Validator Return",
		XAxisTitle:   "",
		YAxisTitle:   "Estimated Validator Return [ETH/day]",
		StackingMode: "false",
		Type:         "line",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Estimated Validator Return",
				Data: seriesData,
			},
		},
	}

	return chartData, nil
}

func stakeEffectivenessChartData() (*types.GenericChartData, error) {
	rows := []struct {
		Epoch                 uint64
		Totalvalidatorbalance uint64
		Eligibleether         uint64
	}{}

	err := db.DB.Select(&rows, `
		SELECT
			epoch, 
			COALESCE(totalvalidatorbalance,0) as totalvalidatorbalance,
			COALESCE(eligibleether,0) as eligibleether
		FROM epochs ORDER BY epoch`)
	if err != nil {
		return nil, err
	}

	seriesData := [][]float64{}

	for _, row := range rows {
		if row.Eligibleether == 0 {
			continue
		}
		if row.Totalvalidatorbalance == 0 {
			continue
		}
		seriesData = append(seriesData, []float64{
			float64(utils.EpochToTime(row.Epoch).Unix() * 1000),
			100 * float64(row.Eligibleether) / float64(row.Totalvalidatorbalance),
		})
	}

	chartData := &types.GenericChartData{
		Title:        "Stake Effectiveness",
		Subtitle:     "History of the Stake Effectiveness",
		XAxisTitle:   "",
		YAxisTitle:   "Stake Effectiveness [%]",
		StackingMode: "false",
		Type:         "line",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Stake Effectiveness",
				Data: seriesData,
			},
		},
	}

	return chartData, nil
}
