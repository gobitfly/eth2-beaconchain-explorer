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

var advertisewithusTemplate = template.Must(template.New("advertisewithus").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/advertisewithus.html"))

func AdvertiseWithUs(w http.ResponseWriter, r *http.Request) {
	var err error
	advertisewithusTemplate = template.Must(template.New("advertisewithus").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/advertisewithus.html"))

	w.Header().Set("Content-Type", "text/html")
	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - Adverstise With Us - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/advertisewithus",
			GATag:       utils.Config.Frontend.GATag,
		},
		ShowSyncingMessage:    services.IsSyncing(),
		Active:                "advertisewithus",
		User:                  getUser(w, r),
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
		CurrentEpoch:          services.LatestEpoch(),
		CurrentSlot:           services.LatestSlot(),
		FinalizationDelay:     services.FinalizationDelay(),
	}

	pageData := &types.AdvertiseWithUsPageData{}
	pageData.FlashMessage, err = utils.GetFlash(w, r, "ad_flash")
	if err != nil {
		logger.Errorf("error retrieving flashes for advertisewithusform %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	data.Data = pageData

	err = advertisewithusTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

func AdvertiseWithUsPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		logger.Errorf("error parsing form: %v", err)
		utils.SetFlash(w, r, "ad_flash", "Error: invalid form submitted")
		http.Redirect(w, r, "/advertisewithus", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	url := r.FormValue("url")
	company := r.FormValue("company")
	ad := r.FormValue("ad")
	comments := r.FormValue("comments")

	msg := fmt.Sprintf(`New ad inquiry:
								Name: %s
								Email: %s
								Url: %s
								Company: %s
								Ad: %s
								Comments: %s`, name, email, url, company, ad, comments)
	// escape html
	msg = template.HTMLEscapeString(msg)

	err = mail.SendMail("support@beaconcha.in", "New ad inquiry", msg)
	if err != nil {
		logger.Errorf("error sending ad form: %v", err)
		utils.SetFlash(w, r, "ad_flash", "Error: unable to submit ad request")
		http.Redirect(w, r, "/advertisewithus", http.StatusSeeOther)
		return
	}

	utils.SetFlash(w, r, "ad_flash", "Thank you for your inquiry, we will get back to you as soon as possible.")
	http.Redirect(w, r, "/advertisewithus", http.StatusSeeOther)
}
