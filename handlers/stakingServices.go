package handlers

import (
	"eth2-exporter/mail"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
)

var stakingServicesTemplate = template.Must(template.New("stakingServices").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/stakingServices.html", "templates/components/bannerStakingServices.html"))

func StakingServices(w http.ResponseWriter, r *http.Request) {
	var err error

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "services", "/stakingServices", "Ethereum 2.0 Staking Services Overview")

	pageData := &types.StakeWithUsPageData{}
	pageData.FlashMessage, err = utils.GetFlash(w, r, "stake_flash")
	if err != nil {
		logger.Errorf("error retrieving flashes for advertisewithusform %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	data.Data = pageData

	err = stakingServicesTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
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

	err = mail.SendMail("support@beaconcha.in", "New staking inquiry", msg)
	if err != nil {
		logger.Errorf("error sending ad form: %v", err)
		utils.SetFlash(w, r, "stake_flash", "Error: unable to submit ad request")
		http.Redirect(w, r, "/stakingServices", http.StatusSeeOther)
		return
	}

	utils.SetFlash(w, r, "stake_flash", "Thank you for your inquiry, we will get back to you as soon as possible.")
	http.Redirect(w, r, "/stakingServices", http.StatusSeeOther)
}
