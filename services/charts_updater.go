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
	"deposits_distribution":          {13, depositsDistributionChartData},
	"graffiti_wordcloud":             {14, graffitiCloudChartData},
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
		if prevEpoch >= latestEpoch && latestEpoch != 0 {
			time.Sleep(sleepDuration)
			continue
		}
		now := time.Now()

		if now.Add(time.Minute * -20).After(utils.EpochToTime(latestEpoch)) {
			logger.Info("skipping chartsPageDataUpdater because the explorer is syncing")
			time.Sleep(time.Second * 60)
			continue
		}

		data, err := getChartsPageData()
		if err != nil {
			logger.WithField("epoch", latestEpoch).Errorf("error updating chartPageData: %v", err)
			time.Sleep(sleepDuration)
			continue
		}
		logger.WithField("epoch", latestEpoch).WithField("duration", time.Since(now)).Info("chartPageData update completed")
		chartsPageData.Store(&data)
		prevEpoch = latestEpoch
		if latestEpoch == 0 {
			time.Sleep(time.Second * 60 * 10)
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
				Data:  dailyProposedBlocks,
			},
			{
				Name:  "Missed",
				Color: "#f7a35c",
				Data:  dailyMissedBlocks,
			},
			{
				Name:  "Orphaned",
				Color: "#adadad",
				Data:  dailyOrphanedBlocks,
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
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
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
		Subtitle:     "History of daily staked Ether, which is the sum of all Effective Balances.",
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
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
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
			dailyAverageBalance = append(dailyAverageBalance, []float64{day, utils.RoundDecimals(float64(row.AverageValidatorBalance)/1e9, 4)})
		}
	}

	chartData := &types.GenericChartData{
		Title:        "Validator Balance",
		Subtitle:     "Average Daily Validator Balance.",
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
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

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
			utils.RoundDecimals(row.Globalparticipationrate*100, 2),
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

func inclusionDistanceChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	latestEpoch := LatestEpoch()
	epochOffset := uint64(0)
	maxEpochs := 1 * 24 * 3600 / (utils.Config.Chain.SlotsPerEpoch * utils.Config.Chain.SecondsPerSlot)
	if latestEpoch > maxEpochs {
		epochOffset = latestEpoch - maxEpochs
	}

	rows := []struct {
		Epoch             uint64
		Inclusiondistance float64
	}{}

	err := db.DB.Select(&rows, `
		select a.epoch, avg(a.inclusionslot - a.attesterslot) as inclusiondistance
		from attestation_assignments_p a
		inner join blocks b on b.slot = a.attesterslot and b.status = '1'
		where a.week >= $1 / 1575 a.epoch > $1 and a.inclusionslot > 0
		group by a.epoch
		order by a.epoch asc`, epochOffset)
	if err != nil {
		return nil, err
	}

	seriesData := [][]float64{}

	for _, row := range rows {
		seriesData = append(seriesData, []float64{
			float64(utils.EpochToTime(row.Epoch).Unix() * 1000),
			utils.RoundDecimals(row.Inclusiondistance, 2),
		})
	}

	chartData := &types.GenericChartData{
		Title:        "Average Inclusion Distance (last 24h)",
		Subtitle:     "Inclusion Distance measures how long it took to include attestations in slots.",
		XAxisTitle:   "",
		YAxisTitle:   "Average Inclusion Distance [slots]",
		StackingMode: "false",
		Type:         "line",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Average Inclusion Distance",
				Data: seriesData,
			},
		},
	}

	return chartData, nil
}

func votingDistributionChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	latestEpoch := LatestEpoch()
	epochOffset := uint64(0)
	maxEpochs := 7 * 3600 * 24 / (utils.Config.Chain.SlotsPerEpoch * utils.Config.Chain.SecondsPerSlot)
	if latestEpoch > maxEpochs {
		epochOffset = latestEpoch - maxEpochs
	}

	rows := []struct {
		Epoch             uint64
		Inclusiondistance float64
	}{}

	err := db.DB.Select(&rows, `
		select a.epoch, avg(a.inclusionslot - a.attesterslot) as inclusiondistance
		from attestation_assignments_p a
		inner join blocks b on b.slot = a.attesterslot and b.status = '1'
		where a.inclusionslot > 0 and a.epoch > $1and a.week >= $1 / 1575
		group by a.epoch
		order by a.epoch asc`, epochOffset)
	if err != nil {
		return nil, err
	}

	seriesData := [][]float64{}

	for _, row := range rows {
		seriesData = append(seriesData, []float64{
			float64(utils.EpochToTime(row.Epoch).Unix() * 1000),
			utils.RoundDecimals(row.Inclusiondistance, 2),
		})
	}

	chartData := &types.GenericChartData{
		Title:        "Average Inclusion Distance (last 7 days)",
		Subtitle:     "Inclusion Distance measures how long it took to include attestations in slots.",
		XAxisTitle:   "",
		YAxisTitle:   "Average Inclusion Distance [slots]",
		StackingMode: "false",
		Type:         "line",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Average Inclusion Distance",
				Data: seriesData,
			},
		},
	}

	return chartData, nil
}

func averageDailyValidatorIncomeChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	rows := []struct {
		Epoch           uint64
		Validatorscount uint64
		Rewards         int64
	}{}

	err := db.DB.Select(&rows, `
		with
			firstdeposits as (
				select distinct
					vb.epoch,
					sum(coalesce(vb.balance,32e9)) over (order by v.activationepoch asc) as amount
				from validators v
					left join validator_balances_p vb
						on vb.validatorindex = v.validatorindex
						and vb.epoch = v.activationepoch
						and vb.week = v.activationepoch / 1575
				order by vb.epoch
			),
			extradeposits as (
				select distinct
					(d.block_slot/32)-1 AS epoch,
					sum(d.amount) over (
						order by d.block_slot/32 asc
					) as amount
				from validators
					inner join blocks_deposits d
						on d.publickey = validators.pubkey
						and d.block_slot/32 > validators.activationepoch
				order by epoch
			)
		select 
			e.epoch,
			e.validatorscount,
			e.totalvalidatorbalance-coalesce(fd.amount,0)-coalesce(ed.amount,0) as rewards
		from epochs e
			left join firstdeposits fd on fd.epoch = (
				select epoch from firstdeposits where epoch <= e.epoch order by epoch desc limit 1
			)
			left join extradeposits ed on fd.epoch = (
				select epoch from extradeposits where epoch <= e.epoch order by epoch desc limit 1
			)
		order by epoch`)
	if err != nil {
		return nil, err
	}

	seriesData := [][]float64{}

	var rewards int64
	var day float64
	validatorsCount := uint64(0)
	prevDayRewards := int64(0)
	prevDay := float64(utils.EpochToTime(0).Truncate(time.Hour*24).Unix() * 1000)
	for _, row := range rows {
		validatorsCount = row.Validatorscount
		rewards = row.Rewards
		day = float64(utils.EpochToTime(row.Epoch).Truncate(time.Hour*24).Unix() * 1000)
		if day != prevDay {
			// data for previous day
			seriesData = append(seriesData, []float64{
				prevDay,
				utils.RoundDecimals(float64(rewards-prevDayRewards)/float64(validatorsCount)/1e9, 4),
			})
			prevDayRewards = row.Rewards
			prevDay = day
		}
	}
	// data for current day
	seriesData = append(seriesData, []float64{
		day,
		utils.RoundDecimals(float64(rewards-prevDayRewards)/float64(validatorsCount)/1e9, 4),
	})

	chartData := &types.GenericChartData{
		Title:        "Validator Income",
		Subtitle:     "Average Daily Validator Income.",
		XAxisTitle:   "",
		YAxisTitle:   "Average Daily Validator Income [ETH/day]",
		StackingMode: "false",
		Type:         "column",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Average Daily Validator Income",
				Data: seriesData,
			},
		},
	}

	return chartData, nil
}

func stakingRewardsChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	rows := []struct {
		Epoch   uint64
		Rewards int64
	}{}

	err := db.DB.Select(&rows, `
		with
			firstdeposits as (
				select distinct
					vb.epoch,
					sum(coalesce(vb.balance,32e9)) over (order by v.activationepoch asc) as amount
				from validators v
					left join validator_balances_p vb
						on vb.validatorindex = v.validatorindex
						and vb.epoch = v.activationepoch
						and vb.week = v.activationepoch / 1575
				order by vb.epoch
			),
			extradeposits as (
				select distinct
					(d.block_slot/32)-1 AS epoch,
					sum(d.amount) over (
						order by d.block_slot/32 asc
					) as amount
				from validators
					inner join blocks_deposits d
						on d.publickey = validators.pubkey
						and d.block_slot/32 > validators.activationepoch
				order by epoch
			)
		select 
			e.epoch,
			e.totalvalidatorbalance-coalesce(fd.amount,0)-coalesce(ed.amount,0) as rewards
		from epochs e
			left join firstdeposits fd on fd.epoch = (
				select epoch from firstdeposits where epoch <= e.epoch order by epoch desc limit 1
			)
			left join extradeposits ed on ed.epoch = (
				select epoch from extradeposits where epoch <= e.epoch order by epoch desc limit 1
			)
		order by epoch`)
	if err != nil {
		return nil, err
	}

	seriesData := [][]float64{}

	var rewards float64
	var day float64
	prevDay := float64(utils.EpochToTime(0).Truncate(time.Hour*24).Unix() * 1000)
	for _, row := range rows {
		rewards = utils.RoundDecimals(float64(row.Rewards)/1e9, 4)
		day = float64(utils.EpochToTime(row.Epoch).Truncate(time.Hour*24).Unix() * 1000)
		if day != prevDay {
			// data for previous day
			seriesData = append(seriesData, []float64{
				prevDay,
				rewards,
			})
			prevDay = day
		}
	}
	// data for current day
	seriesData = append(seriesData, []float64{
		day,
		rewards,
	})

	chartData := &types.GenericChartData{
		Title:        "Staking Rewards",
		Subtitle:     "Total Accumulated Staking Rewards",
		XAxisTitle:   "",
		YAxisTitle:   "Staking Rewards [ETH]",
		StackingMode: "false",
		Type:         "line",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Staking Rewards",
				Data: seriesData,
			},
		},
	}

	return chartData, nil
}

func estimatedValidatorIncomeChartData() (*types.GenericChartData, error) {
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

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
	maxEffectiveBalance := uint64(32e9)
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
		Title:        "Average Daily Validator Income",
		Subtitle:     "",
		XAxisTitle:   "",
		YAxisTitle:   "Average Daily Validator Income [ETH/day]",
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
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

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
			utils.RoundDecimals(100*float64(row.Eligibleether)/float64(row.Totalvalidatorbalance), 2),
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
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	tx, err := db.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var currentEpoch uint64
	err = tx.Get(&currentEpoch, "select max(epoch) from epochs")
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
				from validators 
			),
			balances as (
				select balance
				from validators
			),
			histogram as (
				select 
					case
						when min = max then 0
						else width_bucket(balance, min, max, 999) 
					end as bucket,
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
		ShowGapHider:         true,
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
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

	tx, err := db.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var currentEpoch uint64
	err = tx.Get(&currentEpoch, "select max(epoch) from epochs")
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
				from validators
			),
			balances as (
				select effectivebalance
				from validators
			),
			histogram as (
				select 
					case
						when min = max then 0
						else width_bucket(effectivebalance, min, max, 999) 
					end as bucket,
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
		ShowGapHider:         true,
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
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

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
		ShowGapHider:  true,
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
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

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
		ShowGapHider:  true,
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
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

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
		ShowGapHider:  true,
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
	if LatestEpoch() == 0 {
		return nil, fmt.Errorf("chart-data not available pre-genesis")
	}

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
					case
						when min = max then 0
						else width_bucket(performance365d, min, max, 999) 
					end as bucket,
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
		ShowGapHider:  true,
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

func depositsChartData() (*types.GenericChartData, error) {
	var err error

	eth1Rows := []struct {
		Timestamp int64
		Amount    uint64
		Valid     bool
	}{}

	err = db.DB.Select(&eth1Rows, `
		select
			extract(epoch from block_ts)::int as timestamp,
			amount,
			valid_signature as valid
		from eth1_deposits
		order by timestamp`)
	if err != nil {
		return nil, fmt.Errorf("error getting eth1-deposits: %w", err)
	}

	eth2Rows := []struct {
		Slot   uint64
		Amount uint64
	}{}

	err = db.DB.Select(&eth2Rows, `
		select block_slot as slot, amount 
		from blocks_deposits
		order by slot`)
	if err != nil {
		return nil, fmt.Errorf("error getting eth2-deposits: %w", err)
	}

	dailySuccessfulEth1Deposits := [][]float64{}
	dailyFailedEth1Deposits := [][]float64{}
	dailyEth2Deposits := [][]float64{}

	for _, row := range eth1Rows {
		day := float64(time.Unix(row.Timestamp, 0).Truncate(time.Hour*24).Unix() * 1000)

		if row.Valid {
			if len(dailySuccessfulEth1Deposits) == 0 || dailySuccessfulEth1Deposits[len(dailySuccessfulEth1Deposits)-1][0] != day {
				dailySuccessfulEth1Deposits = append(dailySuccessfulEth1Deposits, []float64{day, float64(row.Amount / 1e9)})
			} else {
				dailySuccessfulEth1Deposits[len(dailySuccessfulEth1Deposits)-1][1] += float64(row.Amount / 1e9)
			}
		} else {
			if len(dailyFailedEth1Deposits) == 0 || dailyFailedEth1Deposits[len(dailyFailedEth1Deposits)-1][0] != day {
				dailyFailedEth1Deposits = append(dailyFailedEth1Deposits, []float64{day, float64(row.Amount / 1e9)})
			} else {
				dailyFailedEth1Deposits[len(dailyFailedEth1Deposits)-1][1] += float64(row.Amount / 1e9)
			}
		}
	}

	for _, row := range eth2Rows {
		day := float64(utils.SlotToTime(row.Slot).Truncate(time.Hour*24).Unix() * 1000)

		if len(dailyEth2Deposits) == 0 || dailyEth2Deposits[len(dailyEth2Deposits)-1][0] != day {
			dailyEth2Deposits = append(dailyEth2Deposits, []float64{day, float64(row.Amount / 1e9)})
		} else {
			dailyEth2Deposits[len(dailyEth2Deposits)-1][1] += float64(row.Amount / 1e9)
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
				Name:  "ETH2",
				Data:  dailyEth2Deposits,
				Stack: "eth2",
			},
			{
				Name:  "ETH1 (success)",
				Data:  dailySuccessfulEth1Deposits,
				Stack: "eth1",
			},
			{
				Name:  "ETH1 (failed)",
				Data:  dailyFailedEth1Deposits,
				Stack: "eth1",
			},
		},
	}

	return chartData, nil
}

func depositsDistributionChartData() (*types.GenericChartData, error) {
	var err error

	rows := []struct {
		Address []byte
		Count   uint64
	}{}

	err = db.DB.Select(&rows, `
		select from_address as address, count(*) as count
		from (
			select publickey, from_address
			from eth1_deposits
			where valid_signature = true
			group by publickey, from_address
			having sum(amount) >= 32e9
		) a
		group by from_address
		order by count desc`)
	if err != nil {
		return nil, fmt.Errorf("error getting eth1-deposits-distribution: %w", err)
	}

	type seriesDataItem struct {
		Name string `json:"name"`
		Y    uint64 `json:"y"`
	}
	seriesData := []seriesDataItem{}
	othersItem := seriesDataItem{
		Name: "others",
		Y:    0,
	}
	for i := range rows {
		if i > 20 {
			othersItem.Y += rows[i].Count
			continue
		}
		seriesData = append(seriesData, seriesDataItem{
			Name: string(utils.FormatEth1AddressString(rows[i].Address)),
			Y:    rows[i].Count,
		})
	}
	if othersItem.Y > 0 {
		seriesData = append(seriesData, othersItem)
	}

	chartData := &types.GenericChartData{
		IsNormalChart:    true,
		Type:             "pie",
		Title:            "Eth1 Deposit Addresses",
		Subtitle:         "Validator distribution by Eth1 deposit address.",
		TooltipFormatter: `function(){ return '<b>'+this.point.name+'</b><br\>Percentage: '+this.point.percentage.toFixed(2)+'%<br\>Validators: '+this.point.y }`,
		PlotOptionsPie: `{
			borderWidth: 1,
			borderColor: null, 
			dataLabels: { 
				enabled:true, 
				formatter: function() { 
					var name = this.point.name.length > 8 ? this.point.name.substring(0,8) : this.point.name;
					return '<span style="stroke:none; fill: var(--font-color)"><b style="stroke:none; fill: var(--font-color)">'+name+'…</b><span style="stroke:none; fill: var(--font-color)">: '+this.point.y+' ('+this.point.percentage.toFixed(2)+'%)</span></span>' 
				} 
			} 
		}`,
		PlotOptionsSeriesCursor: "pointer",
		PlotOptionsSeriesEventsClick: `function(event){ 
			if (event.point.name == 'others') { window.location.href = '/validators/eth1deposits' }
			else { window.location.href = '/validators/eth1deposits?q='+encodeURIComponent(event.point.name) } }`,
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Deposits Distribution",
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

	// \x are missed blocks
	// \x0000000000000000000000000000000000000000000000000000000000000000 are empty graffities
	err := db.DB.Select(&rows, `
		with 
			graffities as (
				select count(*), graffiti
				from blocks 
				where graffiti <> '\x' and graffiti <> '\x0000000000000000000000000000000000000000000000000000000000000000'
				group by graffiti order by count desc limit 25
			)
		select count(distinct blocks.proposer) as validators, graffities.graffiti as name, graffities.count as weight
		from blocks 
			inner join graffities on blocks.graffiti = graffities.graffiti 
		group by graffities.graffiti, graffities.count
		order by weight desc`)
	if err != nil {
		return nil, fmt.Errorf("error getting graffiti-occurences: %w", err)
	}

	for i := range rows {
		rows[i].Name = utils.FormatGraffitiString(rows[i].Name)
	}

	chartData := &types.GenericChartData{
		IsNormalChart:                true,
		Type:                         "wordcloud",
		Title:                        "Graffiti Word Cloud",
		Subtitle:                     "Word Cloud of the 25 most occuring graffities.",
		TooltipFormatter:             `function(){ return '<b>'+this.point.name+'</b><br\>Occurences: '+this.point.weight+'<br\>Validators: '+this.point.validators }`,
		PlotOptionsSeriesEventsClick: `function(event){ window.location.href = '/blocks?q='+encodeURIComponent(event.point.name) }`,
		PlotOptionsSeriesCursor:      "pointer",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Occurences",
				Data: rows,
				Type: "wordcloud",
			},
		},
	}

	return chartData, nil
}
