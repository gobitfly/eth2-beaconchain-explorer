package handlers

import (
	"eth2-exporter/mail"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
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
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	data.Data = pageData

	if handleTemplateError(w, r, "stakingServices.go", "StakingServices", "", stakingServicesTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

func AddStakingServicePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		logger.Errorf("error parsing form: %v", err)
		utils.SetFlash(w, r, "stake_flash", "Error: invalid form submitted")
		http.Redirect(w, r, "/stakingServices", http.StatusSeeOther)
		return
	}

	if len(utils.Config.Frontend.RecaptchaSecretKey) > 0 && len(utils.Config.Frontend.RecaptchaSiteKey) > 0 {
		if len(r.FormValue("g-recaptcha-response")) == 0 {
			utils.SetFlash(w, r, "pricing_flash", "Error: Failed to create request")
			logger.Errorf("error no recaptca response present %v route: %v", r.URL.String(), r.FormValue("g-recaptcha-response"))
			http.Redirect(w, r, "/pricing", http.StatusSeeOther)
			return
		}

		valid, err := utils.ValidateReCAPTCHA(r.FormValue("g-recaptcha-response"))
		if err != nil || !valid {
			utils.SetFlash(w, r, "pricing_flash", "Error: Failed to create request")
			logger.Warnf("error validating recaptcha %v route: %v", r.URL.String(), err)
			http.Redirect(w, r, "/pricing", http.StatusSeeOther)
			return
		}
	}

	name := r.FormValue("name")
	url := r.FormValue("url")
	custodial := r.FormValue("custodial")
	stake := r.FormValue("stake")
	fee := r.FormValue("fee")
	open := r.FormValue("open")
	links := r.FormValue("links")
	comments := r.FormValue("comments")
	thirdParty := r.FormValue("3rdPartySoftware")
	pooltoken := r.FormValue("pooltoken")
	validatorKeyOwner := r.FormValue("validator_keyowner")
	withdrawalKeyOwner := r.FormValue("withdrawal_keyowner")

	msg := fmt.Sprintf(`Add new Staking Service:
								Name: %s
								Url: %s
								Custodial: %s
								Stake: %s
								Fee: %s
								Open Source: %s
								Social Links: %s
								thirdParty: %s
								Pool Token: %s
								Validator Key Owner: %s
								Validator Key Owner: %s
								Comments: %s`, name, url, custodial, stake, fee, open, links, thirdParty, pooltoken, validatorKeyOwner, withdrawalKeyOwner, comments)
	// escape html
	msg = template.HTMLEscapeString(msg)

	err = mail.SendTextMail("support@beaconcha.in", "New staking inquiry", msg, []types.EmailAttachment{})
	if err != nil {
		logger.Errorf("error sending ad form: %v", err)
		utils.SetFlash(w, r, "stake_flash", "Error: unable to submit ad request")
		http.Redirect(w, r, "/stakingServices", http.StatusSeeOther)
		return
	}

	utils.SetFlash(w, r, "stake_flash", "Thank you for your inquiry, we will get back to you as soon as possible.")
	http.Redirect(w, r, "/stakingServices", http.StatusSeeOther)
}
