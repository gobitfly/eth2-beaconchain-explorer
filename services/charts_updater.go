package services

import (
	"eth2-exporter/cache"
	"eth2-exporter/db"
	"eth2-exporter/metrics"
	"eth2-exporter/rpc"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"hash/fnv"
	"html/template"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/aybabtme/uniplot/histogram"
)

type chartHandler struct {
	Order    int
	DataFunc func() (*types.GenericChartData, error)
}

var ChartHandlers = map[string]chartHandler{
	"blocks":             {1, blocksChartData},
	"validators":         {2, activeValidatorsChartData},
	"staked_ether":       {3, stakedEtherChartData},
	"average_balance":    {4, averageBalanceChartData},
	"network_liveness":   {5, networkLivenessChartData},
	"participation_rate": {6, participationRateChartData},

	// "inclusion_distance":             {7, inclusionDistanceChartData},
	// "incorrect_attestations":         {6, incorrectAttestationsChartData},
	// "validator_income":               {7, averageDailyValidatorIncomeChartData},
	// "staking_rewards":                {8, stakingRewardsChartData},

	"stake_effectiveness":            {9, stakeEffectivenessChartData},
	"balance_distribution":           {10, balanceDistributionChartData},
	"effective_balance_distribution": {11, effectiveBalanceDistributionChartData},
	"performance_distribution_365d":  {12, performanceDistribution365dChartData},
	"deposits":                       {13, depositsChartData},
	"withdrawals":                    {17, withdrawalsChartData},
	"graffiti_wordcloud":             {14, graffitiCloudChartData},
	"pools_distribution":             {15, poolsDistributionChartData},
	"historic_pool_performance":      {16, historicPoolPerformanceData},

	// execution charts start with 20+

	"avg_gas_used_chart_data": {22, AvgGasUsedChartData},
	"execution_burned_fees":   {23, BurnedFeesChartData},
	"block_gas_used":          {25, TotalGasUsedChartData},
	// "non_failed_tx_gas_usage_chart_data": {21, NonFailedTxGasUsageChartData},
	"block_count_chart_data":    {26, BlockCountChartData},
	"block_time_avg_chart_data": {27, BlockTimeAvgChartData},
	// "avg_gas_price":                      {25, AvgGasPrice},
	"avg_gas_limit_chart_data":  {28, AvgGasLimitChartData},
	"avg_block_util_chart_data": {29, AvgBlockUtilChartData},
	"tx_count_chart_data":       {31, TxCountChartData},
	// "avg_block_size_chart_data":          {32, AvgBlockSizeChartData},
}

// LatestChartsPageData returns the latest chart page data
func LatestChartsPageData() []*types.ChartsPageDataChart {
	wanted := &[]*types.ChartsPageDataChart{}
	cacheKey := fmt.Sprintf("%d:frontend:chartsPageData", utils.Config.Chain.ClConfig.DepositChainID)

	if wanted, err := cache.TieredCache.GetWithLocalTimeout(cacheKey, time.Hour, wanted); err == nil {
		return *wanted.(*[]*types.ChartsPageDataChart)
	} else {
		logger.Errorf("error retrieving chartsPageData from cache: %v", err)
	}

	return nil
}

func chartsPageDataUpdater(wg *sync.WaitGroup) {
	sleepDuration := time.Hour // only update charts once per hour
	var prevEpoch uint64

	firstun := true
	for {
		latestEpoch := LatestEpoch()
		if prevEpoch >= latestEpoch && latestEpoch != 0 {
			time.Sleep(sleepDuration)
			continue
		}
		start := time.Now()

		// if start.Add(time.Minute * -20).After(utils.EpochToTime(latestEpoch)) {
		// 	logger.Info("skipping chartsPageDataUpdater because the explorer is syncing")
		// 	time.Sleep(time.Minute)
		// 	continue
		// }

		data, err := getChartsPageData()
		if err != nil {
			logger.WithField("epoch", latestEpoch).Errorf("error updating chartPageData: %v", err)
			time.Sleep(sleepDuration)
			continue
		}
		metrics.TaskDuration.WithLabelValues("service_charts_updater").Observe(time.Since(start).Seconds())
		logger.WithField("epoch", latestEpoch).WithField("duration", time.Since(start)).Info("chartPageData update completed")

		cacheKey := fmt.Sprintf("%d:frontend:chartsPageData", utils.Config.Chain.ClConfig.DepositChainID)
		cache.TieredCache.Set(cacheKey, data, utils.Day)

		prevEpoch = latestEpoch

		if firstun {
			wg.Done()
			firstun = false
		}
		if latestEpoch == 0 {
			ReportStatus("chartsPageDataUpdater", "Running", nil)
			time.Sleep(time.Minute * 10)
		}
	}
}

func getChartsPageData() ([]*types.ChartsPageDataChart, error) {
	type chartHandlerRes struct {
		Order int
		Path  string
		Data  *types.GenericChartData
		Error error
	}

	// add charts if it is mainnet
	if utils.Config.Chain.ClConfig.DepositChainID == 1 {
		ChartHandlers["total_supply"] = chartHandler{20, TotalEmissionChartData}
		ChartHandlers["market_cap_chart_data"] = chartHandler{21, MarketCapChartData}
	}

	wg := sync.WaitGroup{}
	wg.Add(len(ChartHandlers))

	chartHandlerResChan := make(chan *chartHandlerRes, len(ChartHandlers))

	for i, ch := range ChartHandlers {
		go func(i string, ch chartHandler) {
			defer wg.Done()
			start := time.Now()
			data, err := ch.DataFunc()
			logger.WithField("chart", i).WithField("duration", time.Since(start)).WithField("error", err).Debug("generated chart")
			chartHandlerResChan <- &chartHandlerRes{ch.Order, i, data, err}
		}(i, ch)
	}

	go func() {
		wg.Wait()
		close(chartHandlerResChan)
	}()

	pageCharts := []*types.ChartsPageDataChart{}

	for chart := range chartHandlerResChan {
		if chart.Error != nil {
			logger.Errorf("error getting chart data for %v: %v", chart.Path, chart.Error)
			continue
		}
		pageCharts = append(pageCharts, &types.ChartsPageDataChart{
			Order: chart.Order,
			Path:  chart.Path,
			Data:  chart.Data,
		})
	}

	sort.Slice(pageCharts, func(i, j int) bool {
		return pageCharts[i].Order < pageCharts[j].Order
	})

	return pageCharts, nil
}

func blocksChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	data := []struct {
		Indicator string    `db:"indicator"`
		Time      time.Time `db:"time"`
		Value     float64   `db:"value"`
	}{}

	err := db.ReaderDb.Select(&data, "SELECT time, value, indicator FROM chart_series WHERE indicator = any('{PROPOSED_BLOCKS, MISSED_BLOCKS, ORPHANED_BLOCKS}') ORDER BY time")
	if err != nil {
		return nil, err
	}

	proposedBlocksSeries := [][]float64{}
	missedBlocksSeries := [][]float64{}
	orphanedBlocksSeries := [][]float64{}

	for _, d := range data {
		switch d.Indicator {
		case "PROPOSED_BLOCKS":
			proposedBlocksSeries = append(proposedBlocksSeries, []float64{float64(d.Time.UnixMilli()), d.Value})
		case "MISSED_BLOCKS":
			missedBlocksSeries = append(missedBlocksSeries, []float64{float64(d.Time.UnixMilli()), d.Value})
		case "ORPHANED_BLOCKS":
			orphanedBlocksSeries = append(orphanedBlocksSeries, []float64{float64(d.Time.UnixMilli()), d.Value})
		default:
			return nil, fmt.Errorf("unexpected indicator %v when generating blocksChartData", d.Indicator)
		}
	}

	chartData := &types.GenericChartData{
		Title:                "Blocks",
		Subtitle:             "History of daily blocks proposed.",
		XAxisTitle:           "",
		YAxisTitle:           "% of Blocks",
		Type:                 "column",
		StackingMode:         "percent",
		DataLabelsEnabled:    true,
		DataLabelsFormatter:  `function(){ return this.point.percentage.toFixed(2)+'%' }`,
		TooltipShared:        true,
		TooltipUseHTML:       true,
		TooltipFollowPointer: true,
		TooltipFormatter: `function(tooltip){
	let header = '<div style="font-weight:bold; text-align:center;">' + Highcharts.dateFormat("%Y-%m-%d %H:%M", this.x) + '</div><table>'
	this.points.sort((a, b) => b.y - a.y)
	let total = 0
	return this.points.reduce(function (s, point) {
		total += point.y
		return s +
			'<tr><td>' +
			'<span style="color:' + point.series.color + ';">\u25CF </span>' +
			'<span style="font-weight:bold;">' + point.series.name + ':</span></td><td>' +
			point.percentage.toFixed(2)+'% ('+point.y+' blocks)'
			'</td></tr>'
	}, header) + 
	'<tr><td>' + 
	'<span>\u25CF </span><span style="font-weight:bold;">Total:</span></td><td>' + total + ' blocks'
	'</td></tr>' +
	'</table>'
}`,
		Series: []*types.GenericChartDataSeries{
			{
				Name:  "Proposed",
				Color: "#90ed7d",
				Data:  proposedBlocksSeries,
			},
			{
				Name:  "Missed",
				Color: "#f7a35c",
				Data:  missedBlocksSeries,
			},
			{
				Name:  "Missed (Orphaned)",
				Color: "#adadad",
				Data:  orphanedBlocksSeries,
			},
		},
	}

	return chartData, nil
}

func activeValidatorsChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	rows := []struct {
		Epoch           uint64
		ValidatorsCount uint64
	}{}

	err := db.ReaderDb.Select(&rows, "SELECT epoch, validatorscount FROM epochs ORDER BY epoch")
	if err != nil {
		return nil, err
	}

	dailyActiveValidators := [][]float64{}

	for _, row := range rows {
		day := float64(utils.EpochToTime(row.Epoch).Truncate(utils.Day).Unix() * 1000)

		if len(dailyActiveValidators) == 0 || dailyActiveValidators[len(dailyActiveValidators)-1][0] != day {
			dailyActiveValidators = append(dailyActiveValidators, []float64{day, float64(row.ValidatorsCount)})
		}
	}

	chartData := &types.GenericChartData{
		Title:                           "Validators",
		Subtitle:                        "History of daily active validators.",
		XAxisTitle:                      "",
		YAxisTitle:                      "# of Validators",
		StackingMode:                    "false",
		Type:                            "column",
		ColumnDataGroupingApproximation: "close",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "# of Validators",
				Data: dailyActiveValidators,
			},
		},
	}

	return chartData, nil
}

func stakedEtherChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	data := []struct {
		Indicator string    `db:"indicator"`
		Time      time.Time `db:"time"`
		Value     float64   `db:"value"`
	}{}

	err := db.ReaderDb.Select(&data, "SELECT time, value, indicator FROM chart_series WHERE indicator = 'STAKED_ETH' ORDER BY time")
	if err != nil {
		return nil, err
	}

	series := [][]float64{}
	for _, d := range data {
		series = append(series, []float64{float64(d.Time.UnixMilli()), d.Value})
	}

	chartData := &types.GenericChartData{
		Title:                           fmt.Sprintf("Staked %v", utils.Config.Frontend.ClCurrency),
		Subtitle:                        fmt.Sprintf("History of daily staked %v, which is the sum of all Effective Balances.", utils.Config.Frontend.ClCurrency),
		XAxisTitle:                      "",
		YAxisTitle:                      utils.Config.Frontend.ClCurrency,
		StackingMode:                    "false",
		Type:                            "column",
		ColumnDataGroupingApproximation: "close",
		Series: []*types.GenericChartDataSeries{
			{
				Name: fmt.Sprintf("Staked %v", utils.Config.Frontend.ClCurrency),
				Data: series,
			},
		},
	}

	return chartData, nil
}

func averageBalanceChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	data := []struct {
		Indicator string    `db:"indicator"`
		Time      time.Time `db:"time"`
		Value     float64   `db:"value"`
	}{}

	err := db.ReaderDb.Select(&data, "SELECT time, value, indicator FROM chart_series WHERE indicator = 'AVG_VALIDATOR_BALANCE_ETH' ORDER BY time")
	if err != nil {
		return nil, err
	}

	series := [][]float64{}
	for _, d := range data {
		series = append(series, []float64{float64(d.Time.UnixMilli()), d.Value})
	}

	chartData := &types.GenericChartData{
		Title:                           "Validator Balance",
		Subtitle:                        "Average Daily Validator Balance.",
		XAxisTitle:                      "",
		YAxisTitle:                      utils.Config.Frontend.ClCurrency,
		StackingMode:                    "false",
		Type:                            "column",
		ColumnDataGroupingApproximation: "average",
		Series: []*types.GenericChartDataSeries{
			{
				Name: fmt.Sprintf("Average Balance [%s]", utils.Config.Frontend.ClCurrency),
				Data: series,
			},
		},
	}

	return chartData, nil
}

func networkLivenessChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	rows := []struct {
		Day  uint64
		Diff uint64
		// FinalizedEpoch uint64
	}{}

	err := db.ReaderDb.Select(&rows, "SELECT EXTRACT(epoch FROM date_trunc('day', ts))::bigint as day, max(headepoch-finalizedepoch) as diff FROM network_liveness group by day ORDER BY day;")
	if err != nil {
		return nil, err
	}

	seriesData := [][]float64{}

	for _, row := range rows {
		seriesData = append(seriesData, []float64{
			float64(row.Day * 1000),
			float64(row.Diff),
		})
	}

	chartData := &types.GenericChartData{
		Title:                           "Network Liveness",
		Subtitle:                        "Network Liveness measures how far the last Finalized Epoch is behind the Head Epoch. The protocol allows epochs to be finalized after 2 epochs.",
		XAxisTitle:                      "",
		YAxisTitle:                      "Network Liveness [epochs]",
		StackingMode:                    "false",
		ColumnDataGroupingApproximation: "high",
		Type:                            "column",
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
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	data := []struct {
		Indicator string    `db:"indicator"`
		Time      time.Time `db:"time"`
		Value     float64   `db:"value"`
	}{}

	err := db.ReaderDb.Select(&data, "SELECT time, value, indicator FROM chart_series WHERE indicator = 'AVG_PARTICIPATION_RATE' ORDER BY time")
	if err != nil {
		return nil, err
	}

	series := [][]float64{}
	for _, d := range data {
		series = append(series, []float64{float64(d.Time.UnixMilli()), d.Value})
	}

	chartData := &types.GenericChartData{
		Title:        "Participation Rate",
		Subtitle:     "Participation Rate measures how many of the validators expected to attest to blocks are actually doing so.",
		XAxisTitle:   "",
		YAxisTitle:   "Participation Rate [%]",
		StackingMode: "false",
		Type:         "line",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Participation Rate",
				Data: series,
			},
		},
	}

	return chartData, nil
}

func historicPoolPerformanceData() (*types.GenericChartData, error) {
	// retrieve pool performance from db
	var performanceDays []types.EthStoreDay
	err := db.ReaderDb.Select(&performanceDays, `
		SELECT pool, day, max(effective_balances_sum_wei) as effective_balances_sum_wei, min(start_balances_sum_wei) as start_balances_sum_wei, max(end_balances_sum_wei) as end_balances_sum_wei, max(deposits_sum_wei) as deposits_sum_wei, AVG(apr) as apr
		FROM historical_pool_performance
		where pool IN (select pool from historical_pool_performance group by pool, day, validators order by day desc, validators desc limit 10)
		GROUP BY pool, day
		ORDER BY day, pool ASC;`)
	if err != nil {
		return nil, fmt.Errorf("error getting historical pool performance: %w", err)
	}

	// generate pool performance series datapoints
	poolSeriesData := map[string][][2]float64{}
	var timestamp float64
	for _, poolPerfDay := range performanceDays {
		timestamp = float64(utils.DayToTime(int64(poolPerfDay.Day)).Unix() * 1000)
		poolSeriesData[poolPerfDay.Pool] = append(poolSeriesData[poolPerfDay.Pool], [2]float64{
			timestamp,
			poolPerfDay.APR.InexactFloat64() * 100,
		})
	}

	// create pool performance series
	var colors = [...]string{
		"#7fa6d4", "#90c978", "#e6a467", "#cc8398", "#bebdbe", "#928b8b", "#a5e5e1", "#ca5c58",
		"#939b58", "#594f9d", "#7d81dc", "#d9cd66", "#d9cd66"}

	chartSeries := []*types.GenericChartDataSeries{}
	hash := fnv.New32()
	var index int

	for poolName, poolData := range poolSeriesData {
		// generate hash from poolname for deterministic way of getting color index
		hash.Write([]byte(poolName))
		index = int(hash.Sum32()) % len(colors)
		hash.Reset()

		poolSeries := types.GenericChartDataSeries{
			Name:  poolName,
			Data:  poolData,
			Color: colors[index],
		}
		chartSeries = append(chartSeries, &poolSeries)
	}

	// retrieve eth.store data from db
	performanceDays = nil
	err = db.ReaderDb.Select(&performanceDays, `
		SELECT	day, effective_balances_sum_wei, start_balances_sum_wei, end_balances_sum_wei, deposits_sum_wei, apr
		FROM	eth_store_stats WHERE validator = -1 
		ORDER BY day ASC`)
	if err != nil {
		return nil, fmt.Errorf("error getting eth store days: %w", err)
	}
	if len(performanceDays) > 0 {
		// generate eth store series datapoints
		for _, ethStoreDay := range performanceDays {
			timestamp = float64(utils.DayToTime(int64(ethStoreDay.Day)).Unix() * 1000)
			poolSeriesData["ETH.STORE"] = append(poolSeriesData["ETH.STORE"], [2]float64{
				timestamp,
				ethStoreDay.APR.InexactFloat64() * 100,
			})
		}
		// create eth store series
		ethStoreSeries := types.GenericChartDataSeries{
			Name:  "ETH.STORE®",
			Data:  poolSeriesData["ETH.STORE"],
			Color: "#ed1c24",
		}
		chartSeries = append([]*types.GenericChartDataSeries{&ethStoreSeries}, chartSeries...)
	}

	//create chart struct, hypertext color is hardcoded into subtitle text
	chartData := &types.GenericChartData{
		Title:         "Historical Pool Performance",
		Subtitle:      "Uses a neutral & verifiable formula <a href=\"https://github.com/gobitfly/eth.store\">ETH.STORE®</a><sup>1</sup> to measure pool performance for consensus & execution rewards.",
		XAxisTitle:    "",
		YAxisTitle:    "APR [%] (Logarithmic)",
		StackingMode:  "false",
		Type:          "line",
		TooltipShared: false,
		Series:        chartSeries,
		Footer:        EthStoreDisclaimer(),
	}

	return chartData, nil
}

func stakeEffectivenessChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	data := []struct {
		Indicator string    `db:"indicator"`
		Time      time.Time `db:"time"`
		Value     float64   `db:"value"`
	}{}

	err := db.ReaderDb.Select(&data, "SELECT time, value, indicator FROM chart_series WHERE indicator = 'AVG_STAKE_EFFECTIVENESS' ORDER BY time")
	if err != nil {
		return nil, err
	}

	series := [][]float64{}
	for _, d := range data {
		series = append(series, []float64{float64(d.Time.UnixMilli()), d.Value})
	}

	chartData := &types.GenericChartData{
		Title:        "Stake Effectiveness",
		Subtitle:     "Stake Effectiveness measures the relation between the sum of all effective balances and the sum of all balances. 100% Stake Effectiveness means that 100% of the locked Ether is used for staking.",
		XAxisTitle:   "",
		YAxisTitle:   "Stake Effectiveness [%]",
		StackingMode: "false",
		Type:         "line",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Stake Effectiveness",
				Data: series,
			},
		},
	}

	return chartData, nil
}

func balanceDistributionChartData() (*types.GenericChartData, error) {
	epoch := LatestEpoch()
	if epoch == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	validators, err := rpc.CurrentClient.GetValidatorState(epoch)
	if err != nil {
		return nil, err
	}

	if validators.Data == nil {
		return nil, fmt.Errorf("GetValidatorState returned empty validator set for epoch %v", epoch)
	}

	currentBalances := make([]float64, 0, len(validators.Data))
	for _, entry := range validators.Data {
		currentBalances = append(currentBalances, float64(entry.Balance)/1e9)
	}

	bins := int(math.Sqrt(float64(len(currentBalances)))) + 1
	hist := histogram.Hist(bins, currentBalances)

	seriesData := make([][]float64, len(hist.Buckets))

	for i, row := range hist.Buckets {
		seriesData[i] = []float64{row.Max, float64(row.Count)}
	}

	chartData := &types.GenericChartData{
		IsNormalChart:        true,
		ShowGapHider:         true,
		Title:                "Balance Distribution",
		Subtitle:             fmt.Sprintf("Histogram of Balances at epoch %d.", epoch),
		XAxisTitle:           "Balance",
		YAxisTitle:           "# of Validators",
		XAxisLabelsFormatter: template.JS(fmt.Sprintf(`function(){ return this.value+' %s' }`, utils.Config.Frontend.ClCurrency)),
		StackingMode:         "false",
		Type:                 "column",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "# of Validators",
				Data: seriesData,
			},
		},
	}

	return chartData, nil
}

func effectiveBalanceDistributionChartData() (*types.GenericChartData, error) {
	epoch := LatestEpoch()
	if epoch == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	validators, err := rpc.CurrentClient.GetValidatorState(epoch)
	if err != nil {
		return nil, err
	}

	if validators.Data == nil {
		return nil, fmt.Errorf("GetValidatorState returned empty validator set for epoch %v", epoch)
	}

	effectiveBalances := make([]float64, 0, len(validators.Data))

	for _, entry := range validators.Data {
		effectiveBalances = append(effectiveBalances, float64(entry.Validator.EffectiveBalance)/1e9)
	}

	bins := int(math.Sqrt(float64(len(effectiveBalances)))) + 1
	hist := histogram.Hist(bins, effectiveBalances)

	seriesData := make([][]float64, len(hist.Buckets))

	for i, row := range hist.Buckets {
		seriesData[i] = []float64{row.Max, float64(row.Count)}
	}

	chartData := &types.GenericChartData{
		IsNormalChart:        true,
		ShowGapHider:         true,
		Title:                "Effective Balance Distribution",
		Subtitle:             fmt.Sprintf("Histogram of Effective Balances at epoch %d.", epoch),
		XAxisTitle:           "Effective Balance",
		YAxisTitle:           "# of Validators",
		XAxisLabelsFormatter: template.JS(fmt.Sprintf(`function(){ return this.value+' %s' }`, utils.Config.Frontend.ClCurrency)),
		StackingMode:         "false",
		Type:                 "column",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "# of Validators",
				Data: seriesData,
			},
		},
	}

	return chartData, nil
}

func performanceDistribution365dChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	var err error

	rows := []struct {
		MaxPerformance float64
		Count          float64
	}{}

	err = db.ReaderDb.Select(&rows, `
		with
			stats as (
				select 
					min(cl_performance_365d) as min,
					max(cl_performance_365d) as max
				from validator_performance
			),
			histogram as (
				select 
					case
						when min = max then 0
						else width_bucket(cl_performance_365d, min, max, 999) 
					end as bucket,
					max(cl_performance_365d) as max,
					count(*) as cnt
				from  validator_performance, stats
				group by bucket
				order by bucket
			)
		select max/1e9 as maxperformance, cnt as count
		from histogram`)
	if err != nil {
		return nil, err
	}

	seriesData := make([][]float64, len(rows))

	for i, row := range rows {
		seriesData[i] = []float64{row.MaxPerformance, row.Count}
	}

	chartData := &types.GenericChartData{
		IsNormalChart: true,
		ShowGapHider:  true,
		Title:         "Income Distribution (365 days)",
		Subtitle:      fmt.Sprintf("Histogram of income-performances of the last 365 days at epoch %d.", LatestEpoch()),
		XAxisTitle:    "Income",
		XAxisLabelsFormatter: template.JS(fmt.Sprintf(`function(){
  if (this.value < 0) return '<span style="color:var(--danger)">'+this.value+' %[1]v<span>'
  return '<span style="color:var(--success)">'+this.value+' %[1]v<span>'
}`, utils.Config.Frontend.ClCurrency)),
		YAxisTitle:   "# of Validators",
		StackingMode: "false",
		Type:         "column",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "# of Validators",
				Data: seriesData,
			},
		},
	}

	return chartData, nil
}

func depositsChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	data := []struct {
		Indicator string    `db:"indicator"`
		Time      time.Time `db:"time"`
		Value     float64   `db:"value"`
	}{}

	err := db.ReaderDb.Select(&data, "SELECT time, value, indicator FROM chart_series WHERE indicator = any('{EL_VALID_DEPOSITS_ETH, EL_INVALID_DEPOSITS_ETH, CL_DEPOSITS_ETH}') ORDER BY time")
	if err != nil {
		return nil, err
	}

	elValidSeries := [][]float64{}
	elInvalidSeries := [][]float64{}
	clSeries := [][]float64{}

	for _, d := range data {
		switch d.Indicator {
		case "EL_VALID_DEPOSITS_ETH":
			elValidSeries = append(elValidSeries, []float64{float64(d.Time.UnixMilli()), d.Value})
		case "EL_INVALID_DEPOSITS_ETH":
			elInvalidSeries = append(elInvalidSeries, []float64{float64(d.Time.UnixMilli()), d.Value})
		case "CL_DEPOSITS_ETH":
			clSeries = append(clSeries, []float64{float64(d.Time.UnixMilli()), d.Value})
		default:
			return nil, fmt.Errorf("unexpected indicator %v when generating depositsChartData", d.Indicator)
		}
	}

	chartData := &types.GenericChartData{
		Title:        "Deposits",
		Subtitle:     "Daily Amount of deposited ETH.",
		XAxisTitle:   "Income",
		YAxisTitle:   "Deposited ETH",
		StackingMode: "normal",
		Type:         "column",
		Series: []*types.GenericChartDataSeries{
			{
				Name:  "Consensus",
				Data:  clSeries,
				Stack: "eth2",
				Color: "#66bce9",
			},
			{
				Name:  "Execution (success)",
				Data:  elValidSeries,
				Stack: "eth1",
				Color: "#7dc382",
			},
			{
				Name:  "Execution (failed)",
				Data:  elInvalidSeries,
				Stack: "eth1",
				Color: "#f3454a",
			},
		},
	}

	return chartData, nil
}

func withdrawalsChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	rows := []struct {
		Time  time.Time `db:"time"`
		Value float64   `db:"value"`
	}{}

	err := db.ReaderDb.Select(&rows, "SELECT time, value FROM chart_series WHERE indicator = 'WITHDRAWALS_ETH' ORDER BY time")
	if err != nil {
		return nil, err
	}

	seriesData := [][]float64{}

	for _, row := range rows {
		seriesData = append(seriesData, []float64{
			float64(row.Time.UnixMilli()),
			utils.ClToMainCurrency(row.Value).InexactFloat64(),
		})
	}

	chartData := &types.GenericChartData{
		Title:        "Withdrawals",
		Subtitle:     fmt.Sprintf("Daily Amount of withdrawals in %s.", utils.Config.Frontend.ClCurrency),
		XAxisTitle:   "",
		YAxisTitle:   fmt.Sprintf("Withdrawals %s", utils.Config.Frontend.ClCurrency),
		StackingMode: "normal",
		Type:         "column",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Withdrawals",
				Data: seriesData,
			},
		},
	}

	return chartData, nil
}

func poolsDistributionChartData() (*types.GenericChartData, error) {

	type seriesDataItem struct {
		Name      string `json:"name"`
		Address   string `json:"address"`
		Y         int64  `json:"y"`
		Drilldown string `json:"drilldown"`
	}

	poolsPageData := LatestPoolsPageData()
	poolData := []*types.PoolInfo{}
	if poolsPageData == nil {
		utils.LogError(nil, "got nil for LatestPoolsPageData", 0)
	} else {
		poolData = poolsPageData.PoolInfos
	}
	if len(poolData) > 1 {
		poolData = poolData[1:]
	}

	seriesData := make([]seriesDataItem, 0, len(poolData))

	for _, row := range poolData {
		seriesData = append(seriesData, seriesDataItem{
			Name: row.Name,
			Y:    row.Count,
		})
	}

	chartData := &types.GenericChartData{
		IsNormalChart:    true,
		Type:             "pie",
		Title:            "Pool Distribution",
		Subtitle:         "Validator distribution by staking pool.",
		TooltipFormatter: `function(){ return '<b>'+this.point.name+'</b><br\>Percentage: '+this.point.percentage.toFixed(2)+'%<br\>Validators: '+this.point.y }`,
		PlotOptionsPie: `{
			borderWidth: 1,
			borderColor: null, 
			dataLabels: { 
				enabled:true, 
				formatter: function() { 
					var name = this.point.name.length > 20 ? this.point.name.substring(0,20)+'...' : this.point.name;
					return '<span style="stroke:none; fill: var(--font-color)"><b style="stroke:none; fill: var(--font-color)">'+name+'</b></span>' 
				} 
			} 
		}`,
		PlotOptionsSeriesCursor: "pointer",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Pool Distribution",
				Type: "pie",
				Data: seriesData,
			},
		},
	}

	return chartData, nil
}

func graffitiCloudChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	rows := []struct {
		Name       string `json:"name"`
		Weight     uint64 `json:"weight"`
		Validators uint64 `json:"validators"`
	}{}

	err := db.ReaderDb.Select(&rows, `select graffiti_text as name, count(*) as weight, sum(proposer_count) as validators from graffiti_stats group by graffiti_text order by weight desc limit 25`)
	if err != nil {
		return nil, fmt.Errorf("error getting graffiti-occurrences: %w", err)
	}

	for i := range rows {
		rows[i].Name = utils.FormatGraffitiString(rows[i].Name)
	}

	chartData := &types.GenericChartData{
		IsNormalChart:                true,
		Type:                         "wordcloud",
		Title:                        "Graffiti Word Cloud",
		Subtitle:                     "Word Cloud of the 25 most occurring graffities.",
		TooltipFormatter:             `function(){ return '<b>'+this.point.name+'</b><br\>Occurrences: '+this.point.weight+'<br\>Validators: '+this.point.validators }`,
		PlotOptionsSeriesEventsClick: `function(event){ window.location.href = '/slots?q='+encodeURIComponent(event.point.name) }`,
		PlotOptionsSeriesCursor:      "pointer",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Occurrences",
				Data: rows,
				Type: "wordcloud",
			},
		},
	}

	return chartData, nil
}

func BurnedFeesChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	rows := []struct {
		Day        time.Time `db:"time"`
		BurnedFees float64   `db:"value"`
	}{}

	err := db.ReaderDb.Select(&rows, "SELECT time, Round(value / 1e18, 2) as value FROM chart_series WHERE indicator = 'BURNED_FEES' ORDER BY time")
	if err != nil {
		return nil, err
	}

	seriesData := [][]float64{}

	for _, row := range rows {
		seriesData = append(seriesData, []float64{
			float64(row.Day.UnixMilli()),
			row.BurnedFees,
		})
	}

	chartData := &types.GenericChartData{
		Title:                           "Burned Fees",
		Subtitle:                        "Evolution of the total number of Ether burned with EIP 1559",
		XAxisTitle:                      "",
		YAxisTitle:                      "Burned Fees [ETH]",
		StackingMode:                    "false",
		Type:                            "area",
		ColumnDataGroupingApproximation: "average",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Burned Fees",
				Data: seriesData,
			},
		},
	}

	return chartData, nil
}

func NonFailedTxGasUsageChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	rows := []struct {
		Day        time.Time `db:"time"`
		BurnedFees float64   `db:"value"`
	}{}

	err := db.ReaderDb.Select(&rows, "SELECT time, ROUND(value, 0) as value FROM chart_series WHERE indicator = 'NON_FAILED_TX_GAS_USAGE' ORDER BY time")
	if err != nil {
		return nil, err
	}

	seriesData := [][]float64{}

	for _, row := range rows {
		seriesData = append(seriesData, []float64{
			float64(row.Day.UnixMilli()),
			row.BurnedFees,
		})
	}

	chartData := &types.GenericChartData{
		// IsNormalChart: true,
		Title:                           "Gas Usage - Successful Tx",
		Subtitle:                        "Gas usage of successful transactions that are not reverted.",
		XAxisTitle:                      "",
		YAxisTitle:                      "Gas Usage [Gas]",
		StackingMode:                    "false",
		Type:                            "area",
		ColumnDataGroupingApproximation: "average",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Gas Usage",
				Data: seriesData,
			},
		},
	}

	return chartData, nil
}

func BlockCountChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	rows := []struct {
		Day   time.Time `db:"time"`
		Value float64   `db:"value"`
	}{}

	err := db.ReaderDb.Select(&rows, "SELECT time, ROUND(value, 0) as value FROM chart_series WHERE indicator = 'BLOCK_COUNT' ORDER BY time")
	if err != nil {
		return nil, err
	}

	seriesData := [][]float64{}

	for _, row := range rows {
		seriesData = append(seriesData, []float64{
			float64(row.Day.UnixMilli()),
			row.Value,
		})
	}

	chartData := &types.GenericChartData{
		Title:                           "Daily Block Count",
		Subtitle:                        "Number of blocks produced (daily)",
		XAxisTitle:                      "",
		YAxisTitle:                      "Block Count [#]",
		StackingMode:                    "false",
		Type:                            "area",
		ColumnDataGroupingApproximation: "average",
		TooltipFormatter: `
		function (tooltip) {
			this.point.y = Math.round(this.point.y)
			var orig = tooltip.defaultFormatter.call(this, tooltip)
			var epoch = timeToEpoch(this.x)
			if (epoch > 0) {
				orig[0] = orig[0] + '<span style="font-size:10px">Epoch ' + epoch + '</span>'
			}
			return orig
		}
		`,
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Block Count",
				Data: seriesData,
			},
		},
	}

	return chartData, nil
}

func BlockTimeAvgChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	rows := []struct {
		Day   time.Time `db:"time"`
		Value float64   `db:"value"`
	}{}

	err := db.ReaderDb.Select(&rows, "SELECT time, ROUND(value, 2) as value FROM chart_series WHERE indicator = 'BLOCK_TIME_AVG' ORDER BY time")
	if err != nil {
		return nil, err
	}

	seriesData := [][]float64{}

	for _, row := range rows {
		seriesData = append(seriesData, []float64{
			float64(row.Day.UnixMilli()),
			row.Value,
		})
	}

	chartData := &types.GenericChartData{
		Title:                           "Block Time (Avg)",
		Subtitle:                        "Average time between blocks over the last 24 hours",
		XAxisTitle:                      "",
		YAxisTitle:                      "Block Time [seconds]",
		StackingMode:                    "false",
		Type:                            "area",
		ColumnDataGroupingApproximation: "average",
		TooltipFormatter: `
		function (tooltip) {
			this.point.y = Math.round(this.point.y * 100) / 100
			var orig = tooltip.defaultFormatter.call(this, tooltip)
			var epoch = timeToEpoch(this.x)
			if (epoch > 0) {
				orig[0] = orig[0] + '<span style="font-size:10px">Epoch ' + epoch + '</span>'
			}
			return orig
		}
		`,
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Block Time (s)",
				Data: seriesData,
			},
		},
	}

	return chartData, nil
}

func TotalEmissionChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	rows := []struct {
		Day   time.Time `db:"time"`
		Value float64   `db:"value"`
	}{}

	err := db.ReaderDb.Select(&rows, "SELECT time, ROUND(value / 1e18, 5) as value FROM chart_series WHERE indicator = 'TOTAL_EMISSION' ORDER BY time")
	if err != nil {
		return nil, err
	}

	seriesData := [][]float64{}

	for _, row := range rows {
		seriesData = append(seriesData, []float64{
			float64(row.Day.UnixMilli()),
			72009990.50 + row.Value,
		})
	}

	chartData := &types.GenericChartData{
		// IsNormalChart: true,
		Title:                           "Total Ether Supply",
		Subtitle:                        "Evolution of the total Ether supply",
		XAxisTitle:                      "",
		YAxisTitle:                      "Total Supply [ETH]",
		StackingMode:                    "false",
		Type:                            "area",
		ColumnDataGroupingApproximation: "average",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Total Supply",
				Data: seriesData,
			},
		},
		TooltipFormatter: `
		function (tooltip) {
			this.point.y = Math.round(this.point.y * 100000) / 100000
			var orig = tooltip.defaultFormatter.call(this, tooltip)
			var epoch = timeToEpoch(this.x)
			if (epoch > 0) {
				orig[0] = '<span style="font-size:10px">Epoch ' + epoch + '</span><br />' + orig[0]
			}
			return orig
		}
		`,
	}

	return chartData, nil
}

func AvgGasPrice() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	rows := []struct {
		Day   time.Time `db:"time"`
		Value float64   `db:"value"`
	}{}

	err := db.ReaderDb.Select(&rows, "SELECT time, ROUND(value / 1e9, 2) as value FROM chart_series WHERE indicator = 'AVG_GASPRICE' ORDER BY time")
	if err != nil {
		return nil, err
	}

	seriesData := [][]float64{}

	for _, row := range rows {
		seriesData = append(seriesData, []float64{
			float64(row.Day.UnixMilli()),
			row.Value,
		})
	}

	chartData := &types.GenericChartData{
		Title:                           "Average Gas Price",
		Subtitle:                        "The average gas price for non-EIP1559 transaction.",
		XAxisTitle:                      "",
		YAxisTitle:                      "Gas Price [GWei]",
		StackingMode:                    "false",
		Type:                            "area",
		ColumnDataGroupingApproximation: "average",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Gas Price (avg)",
				Data: seriesData,
			},
		},
		TooltipFormatter: `
		function (tooltip) {
			this.point.y = Math.round(this.point.y * 100000) / 100000
			var orig = tooltip.defaultFormatter.call(this, tooltip)
			var epoch = timeToEpoch(this.x)
			if (epoch > 0) {
				orig[0] = '<span style="font-size:10px">Epoch ' + epoch + '</span><br />' + orig[0]
			}
			return orig
		}
		`,
	}

	return chartData, nil
}

func AvgGasUsedChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	rows := []struct {
		Day        time.Time `db:"time"`
		BurnedFees float64   `db:"value"`
	}{}

	err := db.ReaderDb.Select(&rows, "SELECT time, value FROM chart_series WHERE indicator = 'AVG_GASUSED' ORDER BY time")
	if err != nil {
		return nil, err
	}

	seriesData := [][]float64{}

	for _, row := range rows {
		seriesData = append(seriesData, []float64{
			float64(row.Day.UnixMilli()),
			row.BurnedFees,
		})
	}

	chartData := &types.GenericChartData{
		Title:                           "Block Gas Usage",
		Subtitle:                        "The average amount of gas used by blocks per day.",
		XAxisTitle:                      "",
		YAxisTitle:                      "Block Gas Usage [gas]",
		StackingMode:                    "false",
		Type:                            "area",
		ColumnDataGroupingApproximation: "average",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Average Gas Used",
				Data: seriesData,
			},
		},
	}

	return chartData, nil
}

func TotalGasUsedChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	rows := []struct {
		Day   time.Time `db:"time"`
		Value float64   `db:"value"`
	}{}

	err := db.ReaderDb.Select(&rows, "SELECT time, value FROM chart_series WHERE indicator = 'TOTAL_GASUSED' ORDER BY time")
	if err != nil {
		return nil, err
	}

	seriesData := [][]float64{}

	for _, row := range rows {
		seriesData = append(seriesData, []float64{
			float64(row.Day.UnixMilli()),
			row.Value,
		})
	}

	chartData := &types.GenericChartData{
		Title:                           "Total Gas Usage",
		Subtitle:                        "The total amout of daily gas used.",
		XAxisTitle:                      "",
		YAxisTitle:                      "Total Gas Usage [gas]",
		StackingMode:                    "false",
		Type:                            "area",
		ColumnDataGroupingApproximation: "average",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Total Gas Usage",
				Data: seriesData,
			},
		},
		TooltipFormatter: `
		function (tooltip) {
			this.point.y = Math.round(this.point.y)
			var orig = tooltip.defaultFormatter.call(this, tooltip)
			var epoch = timeToEpoch(this.x)
			if (epoch > 0) {
				orig[0] = '<span style="font-size:10px">Epoch ' + epoch + '</span><br />' + orig[0]
			}
			return orig
		}
		`,
	}

	return chartData, nil
}

func AvgGasLimitChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	rows := []struct {
		Day   time.Time `db:"time"`
		Value float64   `db:"value"`
	}{}

	err := db.ReaderDb.Select(&rows, "SELECT time, value FROM chart_series WHERE indicator = 'AVG_GASLIMIT' ORDER BY time")
	if err != nil {
		return nil, err
	}

	seriesData := [][]float64{}

	for _, row := range rows {
		seriesData = append(seriesData, []float64{
			float64(row.Day.UnixMilli()),
			row.Value,
		})
	}

	chartData := &types.GenericChartData{
		Title:                           "Block Gas Limit",
		Subtitle:                        "Evolution of the average block gas limit",
		XAxisTitle:                      "",
		YAxisTitle:                      "Gas Limit [gas]",
		StackingMode:                    "false",
		Type:                            "area",
		ColumnDataGroupingApproximation: "average",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Gas Limit",
				Data: seriesData,
			},
		},
		TooltipFormatter: `
		function (tooltip) {
			this.point.y = Math.round(this.point.y)
			var orig = tooltip.defaultFormatter.call(this, tooltip)
			var epoch = timeToEpoch(this.x)
			if (epoch > 0) {
				orig[0] = '<span style="font-size:10px">Epoch ' + epoch + '</span><br />' + orig[0]
			}
			return orig
		}
		`,
	}

	return chartData, nil
}

func AvgBlockUtilChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	rows := []struct {
		Day        time.Time `db:"time"`
		BurnedFees float64   `db:"value"`
	}{}

	err := db.ReaderDb.Select(&rows, "SELECT time, value FROM chart_series WHERE indicator = 'AVG_BLOCK_UTIL' ORDER BY time")
	if err != nil {
		return nil, err
	}

	seriesData := [][]float64{}

	for _, row := range rows {
		seriesData = append(seriesData, []float64{
			float64(row.Day.UnixMilli()),
			row.BurnedFees,
		})
	}

	chartData := &types.GenericChartData{
		Title:                           "Average Block Usage",
		Subtitle:                        "Evolution of the average utilization of Ethereum blocks",
		XAxisTitle:                      "",
		YAxisTitle:                      "Block Usage [%]",
		StackingMode:                    "false",
		Type:                            "spline",
		ColumnDataGroupingApproximation: "average",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Block Usage",
				Data: seriesData,
			},
		},
		TooltipFormatter: `
		function (tooltip) {
			this.point.y = Math.round(this.point.y * 100) / 100
			var orig = tooltip.defaultFormatter.call(this, tooltip)
			var epoch = timeToEpoch(this.x)
			if (epoch > 0) {
				orig[0] = '<span style="font-size:10px">Epoch ' + epoch + '</span><br />' + orig[0]
			}
			return orig
		}
		`,
	}

	return chartData, nil
}

func MarketCapChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	rows := []struct {
		Day   time.Time `db:"time"`
		Value float64   `db:"value"`
	}{}

	err := db.ReaderDb.Select(&rows, "SELECT time, value FROM chart_series WHERE indicator = 'MARKET_CAP' ORDER BY time")
	if err != nil {
		return nil, err
	}

	seriesData := [][]float64{}

	for _, row := range rows {
		seriesData = append(seriesData, []float64{
			float64(row.Day.UnixMilli()),
			row.Value,
		})
	}

	chartData := &types.GenericChartData{
		Title:                           "Market Cap",
		Subtitle:                        "The Evolution of the Ethereum Market Cap.",
		XAxisTitle:                      "",
		YAxisTitle:                      "Market Cap [$]",
		StackingMode:                    "false",
		Type:                            "area",
		ColumnDataGroupingApproximation: "average",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Market Cap",
				Data: seriesData,
			},
		},
		TooltipFormatter: `
		function (tooltip) {
			this.point.y = Math.round(this.point.y)
			var orig = tooltip.defaultFormatter.call(this, tooltip)
			var epoch = timeToEpoch(this.x)
			if (epoch > 0) {
				orig[0] = orig[0] + '<span style="font-size:10px">Epoch ' + epoch + '</span>'
			}
			return orig
		}
		`,
	}

	return chartData, nil
}

func TxCountChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	rows := []struct {
		Day        time.Time `db:"time"`
		BurnedFees float64   `db:"value"`
	}{}

	err := db.ReaderDb.Select(&rows, "SELECT time, value FROM chart_series WHERE indicator = 'TX_COUNT' ORDER BY time")
	if err != nil {
		return nil, err
	}

	seriesData := [][]float64{}

	for _, row := range rows {
		seriesData = append(seriesData, []float64{
			float64(row.Day.UnixMilli()),
			row.BurnedFees,
		})
	}

	chartData := &types.GenericChartData{
		Title:                           "Transactions",
		Subtitle:                        "The total number of transactions per day",
		XAxisTitle:                      "",
		YAxisTitle:                      "Tx Count [#]",
		StackingMode:                    "false",
		Type:                            "area",
		ColumnDataGroupingApproximation: "average",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Transactions",
				Data: seriesData,
			},
		},
		TooltipFormatter: `
		function (tooltip) {
			this.point.y = Math.round(this.point.y)
			var orig = tooltip.defaultFormatter.call(this, tooltip)
			var epoch = timeToEpoch(this.x)
			if (epoch > 0) {
				orig[0] = '<span style="font-size:10px">Epoch ' + epoch + '</span><br />' + orig[0]
			}
			return orig
		}
		`,
	}

	return chartData, nil
}

func AvgBlockSizeChartData() (*types.GenericChartData, error) {
	return nil, fmt.Errorf("unimplemented")
}

func PowerConsumptionChartData() (*types.GenericChartData, error) {
	return nil, fmt.Errorf("unimplemented")
}

func NewAccountsChartData() (*types.GenericChartData, error) {
	return nil, fmt.Errorf("unimplemented")
}
