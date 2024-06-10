package handlers

import (
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"net/http"
)

func MobilePage(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "mobilepage.html")
	var mobileTemplate = templates.GetTemplate(templateFiles...)

	var err error
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "more", "/mobile", "Beaconchain Dashboard", templateFiles)
	pageData := &types.AdvertiseWithUsPageData{}
	pageData.RecaptchaKey = utils.Config.Frontend.RecaptchaSiteKey

	pageData.FlashMessage, err = utils.GetFlash(w, r, "ad_flash")
	if err != nil {
		logger.Errorf("error retrieving flashes for mobile page %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	data.Data = pageData

	if handleTemplateError(w, r, "mobilepage.go", "MobilePage", "", mobileTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}
