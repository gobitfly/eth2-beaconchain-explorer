package services

import (
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/prysmaticlabs/prysm/shared/mathutil"
)

type chartHandler struct {
	Order    int
	DataFunc func() (*types.GenericChartData, error)
}

var chartHandlers = map[string]chartHandler{
	"blocks":                         chartHandler{1, blocksChartData},
	"validators":                     chartHandler{2, activeValidatorsChartData},
	"staked_ether":                   chartHandler{3, stakedEtherChartData},
	"average_balance":                chartHandler{4, averageBalanceChartData},
	"network_liveness":               chartHandler{5, networkLivenessChartData},
	"participation_rate":             chartHandler{6, participationRateChartData},
	"validator_income":               chartHandler{7, averageDailyValidatorIncomeChartData},
	"stake_effectiveness":            chartHandler{8, stakeEffectivenessChartData},
	"balance_distribution":           chartHandler{9, balanceDistributionChartData},
	"effective_balance_distribution": chartHandler{10, effectiveBalanceDistributionChartData},
	// performance_distribution_1d":    chartHandler{12, performanceDistribution1dChartData},
	// performance_distribution_7d":    chartHandler{13, performanceDistribution7dChartData},
	// performance_distribution_31d":   chartHandler{14, performanceDistribution31dChartData},
	"performance_distribution_365d": chartHandler{15, performanceDistribution365dChartData},
}

// LatestChartsPageData returns the latest chart page data
func LatestChartsPageData() *[]*types.ChartsPageDataChart {
	data, ok := chartsPageData.Load().(*[]*types.ChartsPageDataChart)
	if !ok {
		return nil
	}
	return data
}

func chartsPageDataUpdater() {
	sleepDuration := time.Second * time.Duration(utils.Config.Chain.SecondsPerSlot)
	var prevEpoch uint64

	for {
		latestEpoch := LatestEpoch()
		if prevEpoch >= latestEpoch {
			time.Sleep(sleepDuration)
			continue
		}
		logger.WithField("epoch", latestEpoch).Info("updating chartPageData")
		now := time.Now()
		data, err := getChartsPageData()
		if err != nil {
			logger.Errorf("error updating chartPageData: %w", err)
			time.Sleep(sleepDuration)
			continue
		}
		logger.WithField("epoch", latestEpoch).WithField("duration", time.Since(now)).Info("chartPageData update completed")
		chartsPageData.Store(&data)
		prevEpoch = latestEpoch
	}
}

func getChartsPageData() ([]*types.ChartsPageDataChart, error) {
	type chartHandlerRes struct {
		Order int
		Path  string
		Data  *types.GenericChartData
		Error error
	}

	wg := sync.WaitGroup{}
	wg.Add(len(chartHandlers))

	chartHandlerResChan := make(chan *chartHandlerRes, len(chartHandlers))

	for i, ch := range chartHandlers {
		go func(i string, ch chartHandler) {
			defer wg.Done()
			data, err := ch.DataFunc()
			if err != nil {
				logger.Errorf("error getting chart data for %v: %w", i, err)
			}
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
			return nil, chart.Error
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
		Subtitle:     "History of daily blocks proposed.",
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

func activeValidatorsChartData() (*types.GenericChartData, error) {
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
		Subtitle:     "History of daily active validators.",
		XAxisTitle:   "",
		YAxisTitle:   "# of Validators",
		StackingMode: "false",
		Type:         "column",
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
		Subtitle:     "History of daily staked Ether, which is the sum of all effctive balances.",
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

func averageBalanceChartData() (*types.GenericChartData, error) {
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
		Subtitle:     "Average Daily Validator Balance",
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
		Title:                           "Network Liveness",
		Subtitle:                        "Network Liveness measures how far the last Finalized Epoch is behind the Head Epoch. The protocol allows epochs to be finalized after 2 epochs. If the last Finalized Epoch is more than 4 epochs behind the Head Epoch all validators will get penalized.",
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
	rows := []struct {
		Epoch                   uint64
		Globalparticipationrate float64
	}{}

	err := db.DB.Select(&rows, "SELECT epoch, globalparticipationrate FROM epochs WHERE epoch < $1 ORDER BY epoch", LatestEpoch())
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
		Subtitle:     "Participation Rate measures how many of the validators expected to attest to blocks are actually doing so.",
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

func averageDailyValidatorIncomeChartData() (*types.GenericChartData, error) {
	rows := []struct {
		Epoch                 uint64
		Validatorscount       uint64
		Totalvalidatorbalance int64
	}{}

	err := db.DB.Select(&rows, `
	with
	extradeposits as (
		SELECT distinct
            (d.block_slot/32)-1 AS epoch,
            sum(d.amount) over (
                order by d.block_slot/32 asc
            ) as amount
        FROM validators
            INNER JOIN blocks_deposits d
                ON d.publickey = validators.pubkey
                AND (d.block_slot/32) > validators.activationepoch
        ORDER BY epoch
	)
select 
	epochs.epoch, validatorscount,
    coalesce(totalvalidatorbalance - coalesce(ed.amount,0),0) as totalvalidatorbalance
from epochs
	left join extradeposits ed on ed.epoch = (
        select epoch from extradeposits where epoch <= epochs.epoch limit 1
    )
    left join network_liveness nl on epochs.epoch = nl.headepoch
order by epoch;`)
	if err != nil {
		return nil, err
	}

	seriesData := [][]float64{}

	var prevTotalvalidatorbalance int64
	var prevDay float64
	for _, row := range rows {
		day := float64(utils.EpochToTime(row.Epoch).Truncate(time.Hour*24).Unix() * 1000)
		if prevDay != day && prevTotalvalidatorbalance != 0 && row.Totalvalidatorbalance != 0 {
			seriesData = append(seriesData, []float64{
				day,
				float64(int64(prevTotalvalidatorbalance)-int64(row.Totalvalidatorbalance)) / float64(row.Validatorscount) / 1e9,
			})
		}
		if prevDay != day {
			prevDay = day
			prevTotalvalidatorbalance = row.Totalvalidatorbalance
		}
	}

	chartData := &types.GenericChartData{
		Title:        "Average daily validator income",
		Subtitle:     "",
		XAxisTitle:   "",
		YAxisTitle:   "Average daily Validator Income [ETH/day]",
		StackingMode: "false",
		Type:         "line",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Average Daily Validator Income",
				Data: seriesData,
			},
		},
	}

	return chartData, nil
}

func estimatedValidatorIncomeChartData() (*types.GenericChartData, error) {
	rows := []struct {
		Epoch                   uint64
		Eligibleether           uint64
		Votedether              uint64
		Validatorscount         uint64
		Finalitydelay           uint64
		Globalparticipationrate float64
		Totalvalidatorbalance   uint64
	}{}

	// note: eligibleether might not be correct, need to check what exactly the node returns
	// for the reward-calculation we need the sum of all effective balances
	err := db.DB.Select(&rows, `
		with
			extradeposits as (
				select
					(d.block_slot/32) as epoch,
					sum(d.amount) as amount
					from validators
				inner join blocks_deposits d 
					on d.publickey = validators.pubkey
					and (d.block_slot/32) > validators.activationepoch
				group by epoch
			)
		select 
			epochs.epoch, eligibleether, votedether, validatorscount, globalparticipationrate,
			coalesce(totalvalidatorbalance - coalesce(ed.amount,0),0) as totalvalidatorbalance
		from epochs
			left join extradeposits ed on epochs.epoch = ed.epoch
			left join network_liveness nl on epochs.epoch = nl.headepoch
		order by epoch;`)
	if err != nil {
		return nil, err
	}

	seriesData := [][]float64{}
	avgDailyValidatorIncomeSeries := [][]float64{}

	// see: https://github.com/ethereum/eth2.0-specs/blob/dev/specs/phase0/beacon-chain.md#rewards-and-penalties-1
	maxEffectiveBalance := uint64(32e8)
	baseRewardFactor := uint64(64)
	baseRewardPerEpoch := uint64(4)
	proposerRewardQuotient := uint64(8)
	slotsPerDay := 3600 * 24 / utils.Config.Chain.SecondsPerSlot
	epochsPerDay := slotsPerDay / utils.Config.Chain.SlotsPerEpoch
	minAttestationInclusionDelay := uint64(1) // epochs
	minEpochsToInactivityPenalty := uint64(4) // epochs
	// inactivityPenaltyQuotient := uint6(33554432) // 2**25

	var prevTotalvalidatorbalance uint64
	var prevDay float64
	for _, row := range rows {
		if row.Eligibleether == 0 {
			continue
		}
		baseReward := maxEffectiveBalance * baseRewardFactor / mathutil.IntegerSquareRoot(row.Eligibleether) / baseRewardPerEpoch
		// Micro-incentives for matching FFG source, FFG target, and head
		rewardPerEpoch := int64(3 * baseReward * row.Votedether / row.Eligibleether)
		// Proposer and inclusion delay micro-rewards
		proposerReward := baseReward / proposerRewardQuotient
		attesters := float64(row.Validatorscount/32) * row.Globalparticipationrate
		rewardPerEpoch += int64(attesters * float64(proposerReward*(utils.Config.Chain.SlotsPerEpoch/row.Validatorscount)))
		rewardPerEpoch += int64((baseReward - proposerReward) / minAttestationInclusionDelay)

		// inactivity-penalty
		if row.Finalitydelay > minEpochsToInactivityPenalty {
			rewardPerEpoch -= int64(baseReward * baseRewardPerEpoch)
			// if the validator is slashed
			// rewardPerEpoch -=  maxEffectiveBalance*finality_delay/inactivityPenaltyQuotient
		}

		ts := float64(utils.EpochToTime(row.Epoch).Unix() * 1000)
		rewardPerDay := rewardPerEpoch * int64(epochsPerDay)
		seriesData = append(seriesData, []float64{
			ts,
			float64(rewardPerDay) / 1e9,
		})

		day := float64(utils.EpochToTime(row.Epoch).Truncate(time.Hour*24).Unix() * 1000)
		if prevDay != day && prevTotalvalidatorbalance != 0 {
			avgDailyValidatorIncomeSeries = append(avgDailyValidatorIncomeSeries, []float64{
				day,
				float64(int64(prevTotalvalidatorbalance)-int64(row.Totalvalidatorbalance)) / float64(row.Validatorscount) / 1e9,
			})
		}
		if prevDay != day {
			prevDay = day
			prevTotalvalidatorbalance = row.Totalvalidatorbalance
		}
	}

	chartData := &types.GenericChartData{
		Title:        "Average daily validator income",
		Subtitle:     "",
		XAxisTitle:   "",
		YAxisTitle:   "Average daily Validator Income [ETH/day]",
		StackingMode: "false",
		Type:         "line",
		Series: []*types.GenericChartDataSeries{
			// {
			// 	Name: "Estimated Daily Validator Income",
			// 	Data: seriesData,
			// },
			{
				Name: "Average Daily Validator Income",
				Data: avgDailyValidatorIncomeSeries,
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
			COALESCE(totalvalidatorbalance, 0) AS totalvalidatorbalance,
			COALESCE(eligibleether, 0) AS eligibleether
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
		Subtitle:     "Stake Effectiveness measures the relation between the sum of all effective balances and the sum of all balances. 100% Stake Effectiveness means that 100% of the locked Ether is used for staking.",
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

func balanceDistributionChartData() (*types.GenericChartData, error) {
	tx, err := db.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var currentEpoch uint64
	err = tx.Get(&currentEpoch, "select max(epoch) from validator_balances")
	if err != nil {
		return nil, err
	}

	rows := []struct {
		MaxBalance float64
		Count      float64
	}{}

	err = tx.Select(&rows, `
		with
			stats as (
				select 
					min(balance) as min,
					max(balance) as max
				from validator_balances where epoch = (select max(epoch) as maxepoch from validator_balances) 
			),
			balances as (
				select balance
				from validator_balances where epoch = (select max(epoch) as maxepoch from validator_balances)
			),
			histogram as (
				select 
					width_bucket(balance, min, max, 999) as bucket,
					max(balance) as max,
					count(*)
				from  balances, stats
				group by bucket
				order by bucket
			)
		select max/1e9 as maxbalance, count
		from histogram`)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	seriesData := make([][]float64, len(rows))

	for i, row := range rows {
		seriesData[i] = []float64{row.MaxBalance, row.Count}
	}

	chartData := &types.GenericChartData{
		IsNormalChart:        true,
		Title:                "Balance Distribution",
		Subtitle:             fmt.Sprintf("Histogram of Balances at epoch %d.", currentEpoch),
		XAxisTitle:           "Balance",
		YAxisTitle:           "# of Validators",
		XAxisLabelsFormatter: `function(){ return this.value+'ETH' }`,
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
	tx, err := db.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var currentEpoch uint64
	err = tx.Get(&currentEpoch, "select max(epoch) from validator_balances")
	if err != nil {
		return nil, err
	}

	rows := []struct {
		MaxBalance float64
		Count      float64
	}{}

	err = tx.Select(&rows, `
		with
			stats as (
				select 
					min(effectivebalance) as min,
					max(effectivebalance) as max
				from validator_balances where epoch = (select max(epoch) as maxepoch from validator_balances) 
			),
			balances as (
				select effectivebalance
				from validator_balances where epoch = (select max(epoch) as maxepoch from validator_balances)
			),
			histogram as (
				select 
					width_bucket(effectivebalance, min, max, 999) as bucket,
					max(effectivebalance) as max,
					count(*)
				from  balances, stats
				group by bucket
				order by bucket
			)
		select max/1e9 as maxbalance, count
		from histogram`)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	seriesData := make([][]float64, len(rows))

	for i, row := range rows {
		seriesData[i] = []float64{row.MaxBalance, row.Count}
	}

	chartData := &types.GenericChartData{
		IsNormalChart:        true,
		Title:                "Effective Balance Distribution",
		Subtitle:             fmt.Sprintf("Histogram of Effective Balances at epoch %d.", currentEpoch),
		XAxisTitle:           "Effective Balance",
		YAxisTitle:           "# of Validators",
		XAxisLabelsFormatter: `function(){ return this.value+'ETH' }`,
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

func performanceDistribution1dChartData() (*types.GenericChartData, error) {
	var err error

	rows := []struct {
		MaxPerformance float64
		Count          float64
	}{}

	err = db.DB.Select(&rows, `
		with
			stats as (
				select 
					min(performance1d) as min,
					max(performance1d) as max
				from validator_performance
			),
			histogram as (
				select 
					width_bucket(performance1d, min, max, 9999) as bucket,
					max(performance1d) as max,
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
		Title:         "Income Distribution (1 day)",
		Subtitle:      fmt.Sprintf("Histogram of income-performances of the last day at epoch %d.", LatestEpoch()),
		XAxisTitle:    "Income",
		XAxisLabelsFormatter: `function(){
  if (this.value < 0) return '<span style="color:var(--danger)">'+this.value+'ETH<span>'
  return '<span style="color:var(--success)">'+this.value+'ETH<span>'
}
`,
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

func performanceDistribution7dChartData() (*types.GenericChartData, error) {
	var err error

	rows := []struct {
		MaxPerformance float64
		Count          float64
	}{}

	err = db.DB.Select(&rows, `
		with
			stats as (
				select 
					min(performance7d) as min,
					max(performance7d) as max
				from validator_performance
			),
			histogram as (
				select 
					width_bucket(performance7d, min, max, 9999) as bucket,
					max(performance7d) as max,
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
		Title:         "Income Distribution (7 days)",
		Subtitle:      fmt.Sprintf("Histogram of income-performances of the last 7 days at epoch %d.", LatestEpoch()),
		XAxisTitle:    "Income",
		XAxisLabelsFormatter: `function(){
  if (this.value < 0) return '<span style="color:var(--danger)">'+this.value+'ETH<span>'
  return '<span style="color:var(--success)">'+this.value+'ETH<span>'
}
`,
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

func performanceDistribution31dChartData() (*types.GenericChartData, error) {
	var err error

	rows := []struct {
		MaxPerformance float64
		Count          float64
	}{}

	err = db.DB.Select(&rows, `
		with
			stats as (
				select 
					min(performance31d) as min,
					max(performance31d) as max
				from validator_performance
			),
			histogram as (
				select 
					width_bucket(performance31d, min, max, 9999) as bucket,
					max(performance31d) as max,
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
		Title:         "Income Distribution (31 days)",
		Subtitle:      fmt.Sprintf("Histogram of income-performances of the last 31 days at epoch %d.", LatestEpoch()),
		XAxisTitle:    "Income",
		XAxisLabelsFormatter: `function(){
  if (this.value < 0) return '<span style="color:var(--danger)">'+this.value+'ETH<span>'
  return '<span style="color:var(--success)">'+this.value+'ETH<span>'
}
`,
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

func performanceDistribution365dChartData() (*types.GenericChartData, error) {
	var err error

	rows := []struct {
		MaxPerformance float64
		Count          float64
	}{}

	err = db.DB.Select(&rows, `
		with
			stats as (
				select 
					min(performance365d) as min,
					max(performance365d) as max
				from validator_performance
			),
			histogram as (
				select 
					width_bucket(performance365d, min, max, 9999) as bucket,
					max(performance365d) as max,
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
		Title:         "Income Distribution (365 days)",
		Subtitle:      fmt.Sprintf("Histogram of income-performances of the last 365 days at epoch %d.", LatestEpoch()),
		XAxisTitle:    "Income",
		XAxisLabelsFormatter: `function(){
  if (this.value < 0) return '<span style="color:var(--danger)">'+this.value+'ETH<span>'
  return '<span style="color:var(--success)">'+this.value+'ETH<span>'
}
`,
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
