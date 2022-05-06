package handlers

import (
	"eth2-exporter/mail"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
)

var mobileTemplate = template.Must(template.New("mobilepage").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/mobilepage.html"))

func MobilePage(w http.ResponseWriter, r *http.Request) {
	var err error
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "more", "/mobile", "Beaconchain Dashboard")
	pageData := &types.AdvertiseWithUsPageData{}
	pageData.RecaptchaKey = utils.Config.Frontend.RecaptchaSiteKey

	pageData.FlashMessage, err = utils.GetFlash(w, r, "ad_flash")
	if err != nil {
		logger.Errorf("error retrieving flashes for mobile page %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	data.Data = pageData
	data.HeaderAd = true

	err2 := mobileTemplate.ExecuteTemplate(w, "layout", data)
	if err2 != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err2)
		http.Error(w, "Internal server error", 503)
		return
	}
}

func MobilePagePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		logger.Errorf("error parsing form: %v", err)
		utils.SetFlash(w, r, "ad_flash", "Error: invalid form submitted")
		http.Redirect(w, r, "/mobile", http.StatusSeeOther)
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
			logger.Errorf("error validating recaptcha %v route: %v", r.URL.String(), err)
			http.Redirect(w, r, "/pricing", http.StatusSeeOther)
			return
		}
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
