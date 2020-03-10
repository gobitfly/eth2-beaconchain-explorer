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
	"time"
)

var chartsTemplate = template.Must(template.New("charts").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/charts.html"))
var genericChartTemplate = template.Must(template.New("chart").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/genericchart.html"))

// Charts uses a go template for presenting the page to show charts
func Charts(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html")

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

// BlocksChart will show the history of daily blocks proposed chart
func BlocksChart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - Blocks Chart - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year()),
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
		Epoch     uint64
		Status    uint64
		NbrBlocks uint64
	}{}

	err := db.DB.Select(&rows, "SELECT epoch, status, count(*) as nbrBlocks FROM blocks GROUP BY epoch, status ORDER BY epoch")
	if err != nil {
		logger.Errorf("error retrieving chart data for route %v: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
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
		Title:        "History of daily blocks proposed",
		XAxisTitle:   "",
		YAxisTitle:   "# of Blocks",
		StackingMode: "normal",
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

	data.Data = chartData

	err = genericChartTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
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
		logger.Errorf("error retrieving chart data for route %v: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}

	dailyActiveValidators := [][]float64{}

	for _, row := range rows {
		day := float64(utils.EpochToTime(row.Epoch).Truncate(time.Hour*24).Unix() * 1000)

		if len(dailyActiveValidators) == 0 || dailyActiveValidators[len(dailyActiveValidators)-1][0] != day {
			dailyActiveValidators = append(dailyActiveValidators, []float64{day, float64(row.ValidatorsCount)})
		}
	}

	chartData := &types.GenericChartData{
		Title:        "History of daily active validators",
		XAxisTitle:   "",
		YAxisTitle:   "# of Validators",
		StackingMode: "false",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Validators",
				Data: dailyActiveValidators,
			},
		},
	}

	data.Data = chartData

	err = genericChartTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
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
		logger.Errorf("error retrieving chart data for route %v: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}

	dailyStakedEther := [][]float64{}

	for _, row := range rows {
		day := float64(utils.EpochToTime(row.Epoch).Truncate(time.Hour*24).Unix() * 1000)

		if len(dailyStakedEther) == 0 || dailyStakedEther[len(dailyStakedEther)-1][0] != day {
			dailyStakedEther = append(dailyStakedEther, []float64{day, float64(row.EligibleEther) / 1000000000})
		}
	}

	chartData := &types.GenericChartData{
		Title:        "History of daily staked Ether",
		Subtitle:     "Ethereum 2.0 Beacon Chain Chart",
		XAxisTitle:   "",
		YAxisTitle:   "Ether",
		StackingMode: "false",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Staked Ether",
				Data: dailyStakedEther,
			},
		},
	}

	data.Data = chartData

	err = genericChartTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
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
		logger.Errorf("error retrieving chart data for route %v: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}

	dailyAverageBalance := [][]float64{}

	for _, row := range rows {
		day := float64(utils.EpochToTime(row.Epoch).Truncate(time.Hour*24).Unix() * 1000)

		if len(dailyAverageBalance) == 0 || dailyAverageBalance[len(dailyAverageBalance)-1][0] != day {
			dailyAverageBalance = append(dailyAverageBalance, []float64{day, float64(row.AverageValidatorBalance) / 1000000000})
		}
	}

	chartData := &types.GenericChartData{
		Title:        "History of the daily average validator balance",
		Subtitle:     "Ethereum 2.0 Beacon Chain Chart",
		XAxisTitle:   "",
		YAxisTitle:   "Ether",
		StackingMode: "false",
		Series: []*types.GenericChartDataSeries{
			{
				Name: "Average Balance [ETH]",
				Data: dailyAverageBalance,
			},
		},
	}

	data.Data = chartData

	err = genericChartTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
