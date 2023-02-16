package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func ChangeWithdrawalCredentials(w http.ResponseWriter, r *http.Request) {
	var tpl = templates.GetTemplate("layout.html", "components/bannerGeneric.html", "changewithdrawalcredentials.html")
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "tools", "/tools/changewithdrawalcredentials", "Change Withdrawal Credentials")

	pageData := &types.ChangeWithdrawalCredentialsPageData{}
	pageData.RecaptchaKey = utils.Config.Frontend.RecaptchaSiteKey

	var err error
	pageData.FlashMessage, err = utils.GetFlash(w, r, "info_flash")
	if err != nil {
		logger.Errorf("error retrieving flashes for changewithdrawalcredentials %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	data.Data = pageData
	err = tpl.ExecuteTemplate(w, "layout", data)
	if handleTemplateError(w, r, "changewithdrawalcredentials.go", "changewithdrawalcredentials", "", err) != nil {
		return // an error has occurred and was processed
	}
}

func ChangeWithdrawalCredentialsJob(w http.ResponseWriter, r *http.Request) {
	var tpl = templates.GetTemplate("layout.html", "components/bannerGeneric.html", "changewithdrawalcredentialsjob.html")
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "tools", "/tools/changewithdrawalcredentials", "Change Withdrawal Credentials Job")

	vars := mux.Vars(r)

	job, err := db.GetNodeJob(vars["jobID"])
	if err != nil {
		logger.Errorf("error retrieving job %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	data.Data = job
	err = tpl.ExecuteTemplate(w, "layout", data)
	if handleTemplateError(w, r, "changewithdrawalcredentials.go", "changewithdrawalcredentialsjob", "", err) != nil {
		return // an error has occurred and was processed
	}
}

func ChangeWithdrawalCredentialsPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		logger.Errorf("error parsing form: %v", err)
		utils.SetFlash(w, r, "info_flash", "Error: invalid form submitted")
		http.Redirect(w, r, "/tools/changeWithdrawalCredentials", http.StatusSeeOther)
		return
	}

	if len(utils.Config.Frontend.RecaptchaSecretKey) > 0 && len(utils.Config.Frontend.RecaptchaSiteKey) > 0 {
		if len(r.FormValue("g-recaptcha-response")) == 0 {
			utils.SetFlash(w, r, "info_flash", "Error: Failed to create request")
			logger.Errorf("error no recaptca response present %v route: %v", r.URL.String(), r.FormValue("g-recaptcha-response"))
			http.Redirect(w, r, "/tools/changeWithdrawalCredentials", http.StatusSeeOther)
			return
		}

		valid, err := utils.ValidateReCAPTCHA(r.FormValue("g-recaptcha-response"))
		if err != nil || !valid {
			utils.SetFlash(w, r, "info_flash", "Error: Failed to create request")
			logger.Errorf("error validating recaptcha %v route: %v", r.URL.String(), err)
			http.Redirect(w, r, "/tools/changeWithdrawalCredentials", http.StatusSeeOther)
			return
		}
	}

	jobData := r.FormValue("inputSignatures")

	job, err := db.CreateBLSToExecutionChangesNodeJob([]byte(jobData))
	if err != nil {
		logger.Errorf("error creating a node-job: %v", err)
		utils.SetFlash(w, r, "info_flash", fmt.Sprintf("Error: %s", err))
		http.Redirect(w, r, "/tools/changeWithdrawalCredentials", http.StatusSeeOther)
		return
	}

	url := fmt.Sprintf("/tools/changeWithdrawalCredentials/%s", job.Info.ID)
	http.Redirect(w, r, url, http.StatusSeeOther)
}
