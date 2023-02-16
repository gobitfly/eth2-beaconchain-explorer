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

func Broadcast(w http.ResponseWriter, r *http.Request) {
	var tpl = templates.GetTemplate("layout.html", "components/bannerGeneric.html", "broadcast.html")
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "tools", "/tools/broadcast", "Change Withdrawal Credentials")
	pageData := &types.BroadcastPageData{}
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
	if handleTemplateError(w, r, "broadcast.go", "broadcast", "", err) != nil {
		return // an error has occurred and was processed
	}
}

func BroadcastPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		logger.Errorf("error parsing form: %v", err)
		utils.SetFlash(w, r, "info_flash", "Error: invalid form submitted")
		http.Redirect(w, r, "/tools/broadcast", http.StatusSeeOther)
		return
	}

	if len(utils.Config.Frontend.RecaptchaSecretKey) > 0 && len(utils.Config.Frontend.RecaptchaSiteKey) > 0 {
		if len(r.FormValue("g-recaptcha-response")) == 0 {
			utils.SetFlash(w, r, "info_flash", "Error: Failed to create request")
			logger.Errorf("error no recaptca response present %v route: %v", r.URL.String(), r.FormValue("g-recaptcha-response"))
			http.Redirect(w, r, "/tools/broadcast", http.StatusSeeOther)
			return
		}

		valid, err := utils.ValidateReCAPTCHA(r.FormValue("g-recaptcha-response"))
		if err != nil || !valid {
			utils.SetFlash(w, r, "info_flash", "Error: Failed to create request")
			logger.Errorf("error validating recaptcha %v route: %v", r.URL.String(), err)
			http.Redirect(w, r, "/tools/broadcast", http.StatusSeeOther)
			return
		}
	}

	jobData := r.FormValue("inputSignatures")

	job, err := db.CreateBLSToExecutionChangesNodeJob([]byte(jobData))
	if err != nil {
		logger.Errorf("error creating a node-job: %v", err)
		utils.SetFlash(w, r, "info_flash", fmt.Sprintf("Error: %s", err))
		http.Redirect(w, r, "/tools/broadcast", http.StatusSeeOther)
		return
	}

	url := fmt.Sprintf("/tools/broadcast/status/%s", job.ID)
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func BroadcastStatus(w http.ResponseWriter, r *http.Request) {
	var tpl = templates.GetTemplate("layout.html", "components/bannerGeneric.html", "broadcaststatus.html")
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "tools", "/tools/broadcast/status", "Change Withdrawal Credentials Job")

	vars := mux.Vars(r)

	job, err := db.GetNodeJob(vars["jobID"])
	if err != nil {
		logger.Errorf("error retrieving job %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	data.Data = job
	err = tpl.ExecuteTemplate(w, "layout", data)
	if handleTemplateError(w, r, "broadcast.go", "broadcast", "", err) != nil {
		return // an error has occurred and was processed
	}
}
