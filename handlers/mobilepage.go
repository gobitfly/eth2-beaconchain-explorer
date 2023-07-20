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
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	data.Data = pageData

	if handleTemplateError(w, r, "mobilepage.go", "MobilePage", "", mobileTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

func MobilePagePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		utils.LogError(err, "error parsing form", 0)
		utils.SetFlash(w, r, "ad_flash", "Error: invalid form submitted")
		http.Redirect(w, r, "/mobile", http.StatusSeeOther)
		return
	}

	if err := utils.HandleRecaptcha(w, r, "/pricing"); err != nil {
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	url := r.FormValue("url")
	company := r.FormValue("company")
	text := r.FormValue("text")

	msg := fmt.Sprintf(`New app pool support inquiry:
								Name: %s
								Email: %s
								Url: %s
								Company: %s
								Message: %s`,
		name, email, url, company, text)
	// escape html
	msg = template.HTMLEscapeString(msg)

	err = mail.SendTextMail("support@beaconcha.in", "New app pool support inquiry", msg, []types.EmailAttachment{})
	if err != nil {
		logger.Errorf("error sending app pool form: %v", err)
		utils.SetFlash(w, r, "ad_flash", "Error: unable to submit app pool request")
		http.Redirect(w, r, "/mobile", http.StatusSeeOther)
		return
	}

	utils.SetFlash(w, r, "ad_flash", "Thank you for your inquiry, we will get back to you as soon as possible.")
	http.Redirect(w, r, "/mobile", http.StatusSeeOther)
}
