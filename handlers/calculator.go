package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"html/template"
	"net/http"
)

var stakingCalculatorTemplate = template.Must(template.New("calculator").Funcs(utils.GetTemplateFuncs()).ParseFS(templates.Files, "layout.html", "calculator.html"))

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

	err = stakingCalculatorTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}
