package handlers

import (
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"net/http"
)

func StakingServices(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "stakingServices.html")
	var stakingServicesTemplate = templates.GetTemplate(templateFiles...)

	var err error

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "services", "/stakingServices", "Ethereum Staking Services Overview", templateFiles)

	pageData := &types.StakeWithUsPageData{}
	pageData.RecaptchaKey = utils.Config.Frontend.RecaptchaSiteKey
	pageData.FlashMessage, err = utils.GetFlash(w, r, "stake_flash")
	if err != nil {
		logger.Errorf("error retrieving flashes for advertisewithusform %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	data.Data = pageData

	if handleTemplateError(w, r, "stakingServices.go", "StakingServices", "", stakingServicesTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}
