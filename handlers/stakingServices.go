package handlers

import (
	"eth2-exporter/mail"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

var stakingServicesTemplate = template.Must(template.New("stakingServices").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/stakingServices.html"))

func StakingServices(w http.ResponseWriter, r *http.Request) {
	var err error

	w.Header().Set("Content-Type", "text/html")
	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - Ethereum 2.0 Staking Services Overview - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/stakingServices",
			GATag:       utils.Config.Frontend.GATag,
		},
		ShowSyncingMessage:    services.IsSyncing(),
		Active:                "stakingServices",
		User:                  getUser(w, r),
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
		CurrentEpoch:          services.LatestEpoch(),
		CurrentSlot:           services.LatestSlot(),
		FinalizationDelay:     services.FinalizationDelay(),
	}

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

	msg := fmt.Sprintf(`Add new Staking Service:
								Name: %s
								Url: %s
								Custodial: %s
								Stake: %s
								Fee: %s
								Open Source: %s
								Social Links: %s
								Comments: %s`, name, url, custodial, stake, fee, open, links, comments)
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
