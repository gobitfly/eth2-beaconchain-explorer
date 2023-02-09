package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/mail"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
)

func ChangeWithdrawalCredentials(w http.ResponseWriter, r *http.Request) {
	var tpl = templates.GetTemplate("layout.html", "changewithdrawalcredentials.html")
	w.Header().Set("Content-Type", "text/html")
	pageData := &types.ChangeWithdrawalCredentialsPageData{}
	pageData.RecaptchaKey = utils.Config.Frontend.RecaptchaSiteKey

	data := InitPageData(w, r, "tools", "/changewithdrawalcredentials", "Change Withdrawal Credentials")
	err := tpl.ExecuteTemplate(w, "layout", data)
	if handleTemplateError(w, r, "changewithdrawalcredentials.go", "changewithdrawalcredentials", "", err) != nil {
		return // an error has occurred and was processed
	}
}

func ChangeWithdrawalCredentialsPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		logger.Errorf("error parsing form: %v", err)
		utils.SetFlash(w, r, "ad_flash", "Error: invalid form submitted")
		http.Redirect(w, r, "/changewithdrawalcredentials", http.StatusSeeOther)
		return
	}

	if len(utils.Config.Frontend.RecaptchaSecretKey) > 0 && len(utils.Config.Frontend.RecaptchaSiteKey) > 0 {
		if len(r.FormValue("g-recaptcha-response")) == 0 {
			utils.SetFlash(w, r, "pricing_flash", "Error: Failed to create request")
			logger.Errorf("error no recaptca response present %v route: %v", r.URL.String(), r.FormValue("g-recaptcha-response"))
			http.Redirect(w, r, "/changewithdrawalcredentials", http.StatusSeeOther)
			return
		}

		valid, err := utils.ValidateReCAPTCHA(r.FormValue("g-recaptcha-response"))
		if err != nil || !valid {
			utils.SetFlash(w, r, "pricing_flash", "Error: Failed to create request")
			logger.Errorf("error validating recaptcha %v route: %v", r.URL.String(), err)
			http.Redirect(w, r, "/changewithdrawalcredentials", http.StatusSeeOther)
			return
		}
	}

	jobData := []byte{}
	job, err := db.CreateBLSToExecutionChangesNodeJob(jobData)
	if err != nil {
		logger.Errorf("error creating a node-job: %v", err)
		utils.SetFlash(w, r, "ad_flash", "Error: something went wrong")
		http.Redirect(w, r, "/changewithdrawalcredentials", http.StatusSeeOther)
		return
	}
	_ = job

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
