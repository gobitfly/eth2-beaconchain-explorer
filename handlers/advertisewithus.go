package handlers

import (
	"eth2-exporter/mail"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
)

var advertisewithusTemplate = template.Must(template.New("advertisewithus").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/advertisewithus.html"))

func AdvertiseWithUs(w http.ResponseWriter, r *http.Request) {
	var err error

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "advertisewithus", "/advertisewithus", "Adverstise With Us")

	pageData := &types.AdvertiseWithUsPageData{}
	pageData.RecaptchaKey = utils.Config.Frontend.RecaptchaSiteKey

	pageData.FlashMessage, err = utils.GetFlash(w, r, "ad_flash")
	if err != nil {
		logger.Errorf("error retrieving flashes for advertisewithusform %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	data.Data = pageData

	err = advertisewithusTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
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

	err = mail.SendTextMail("support@beaconcha.in", "New ad inquiry", msg, []types.EmailAttachment{})
	if err != nil {
		logger.Errorf("error sending ad form: %v", err)
		utils.SetFlash(w, r, "ad_flash", "Error: unable to submit ad request")
		http.Redirect(w, r, "/advertisewithus", http.StatusSeeOther)
		return
	}

	utils.SetFlash(w, r, "ad_flash", "Thank you for your inquiry, we will get back to you as soon as possible.")
	http.Redirect(w, r, "/advertisewithus", http.StatusSeeOther)
}
