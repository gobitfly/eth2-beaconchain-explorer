package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/price"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

var stakingCalculatorTemplate = template.Must(template.New("calculator").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/calculator.html"))

// StakingCalculator renders stakingCalculatorTemplate
func StakingCalculator(w http.ResponseWriter, r *http.Request) {

	calculatorPageData := types.StakingCalculatorPageData{}

	total, err := db.GetTotalEligableEther()
	if err != nil {
		logger.WithError(err).Error("error getting total staked ether")
		http.Error(w, "Internal server error", 503)
		return
	}

	calculatorPageData.TotalStaked = total

	w.Header().Set("Content-Type", "text/html")
	stakingCalculatorTemplate = template.Must(template.New("calculator").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/calculator.html"))
	data := &types.PageData{
		HeaderAd: true,
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - Staking calculator - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/calculator",
			GATag:       utils.Config.Frontend.GATag,
		},
		ShowSyncingMessage:    services.IsSyncing(),
		Active:                "stats",
		Data:                  calculatorPageData,
		User:                  getUser(w, r),
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
		CurrentEpoch:          services.LatestEpoch(),
		CurrentSlot:           services.LatestSlot(),
		FinalizationDelay:     services.FinalizationDelay(),
		EthPrice:              price.GetEthPrice(),
		Mainnet:               utils.Config.Chain.Mainnet,
		DepositContract:       utils.Config.Indexer.Eth1DepositContractAddress,
	}

	// stakingCalculatorTemplate = template.Must(template.New("staking_estimator").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/calculator.html"))
	err = stakingCalculatorTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

func estimatedValidatorIncomeChartData() ([][]float64, error) {
	rows := []struct {
		Epoch                   uint64
		Eligibleether           uint64
		Votedether              uint64
		Validatorscount         uint64
		Finalitydelay           uint64
		Globalparticipationrate float64
	}{}
	err := db.DB.Select(&rows, `
		SELECT 
			epoch, eligibleether, votedether, validatorscount, globalparticipationrate,
			coalesce(nl.headepoch-nl.finalizedepoch,2) as finalitydelay
		FROM epochs
			LEFT JOIN network_liveness nl ON epochs.epoch = nl.headepoch
		ORDER BY epoch`)
	if err != nil {
		return nil, err
	}

	seriesData := [][]float64{}

	for _, row := range rows {
		if row.Eligibleether == 0 {
			continue
		}
		seriesData = append(seriesData, []float64{
			float64(row.Epoch),
			float64(row.Eligibleether),
			float64(row.Votedether),
			float64(row.Validatorscount),
			float64(row.Finalitydelay),
			row.Globalparticipationrate,
		})
	}

	return seriesData, nil
}
