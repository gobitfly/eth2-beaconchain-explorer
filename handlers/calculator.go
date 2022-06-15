package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"html/template"
	"net/http"
)

var stakingCalculatorTemplate = template.Must(template.New("calculator").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/calculator.html"))

// StakingCalculator renders stakingCalculatorTemplate
func StakingCalculator(w http.ResponseWriter, r *http.Request) {

	calculatorPageData := types.StakingCalculatorPageData{}

	total, err := db.GetTotalEligibleEther()
	if err != nil {
		logger.WithError(err).Error("error getting total staked ether")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	calculatorPageData.TotalStaked = total

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "stats", "/calculator", "Staking calculator")
	data.Data = calculatorPageData

	// stakingCalculatorTemplate = template.Must(template.New("staking_estimator").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/calculator.html"))
	err = stakingCalculatorTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
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
	err := db.ReaderDb.Select(&rows, `
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
