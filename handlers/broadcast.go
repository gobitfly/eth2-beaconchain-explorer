package handlers

import (
	"database/sql"
	"errors"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func Broadcast(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "broadcast.html", "validator/withdrawalOverviewRow.html")
	var tpl = templates.GetTemplate(templateFiles...)
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "tools", "/tools/broadcast", "Broadcast", templateFiles)
	pageData := &types.BroadcastPageData{}
	pageData.Stats = services.GetLatestStats()
	pageData.RecaptchaKey = utils.Config.Frontend.RecaptchaSiteKey

	var err error
	pageData.FlashMessage, err = utils.GetFlash(w, r, "info_flash")
	if err != nil {
		logger.Errorf("error retrieving flashes for broadcast %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
		utils.LogError(err, "error parsing form", 0)
		utils.SetFlash(w, r, "info_flash", "Error: invalid form submitted")
		http.Redirect(w, r, "/tools/broadcast", http.StatusSeeOther)
		return
	}

	if len(utils.Config.Frontend.RecaptchaSecretKey) > 0 && len(utils.Config.Frontend.RecaptchaSiteKey) > 0 {
		if len(r.FormValue("g-recaptcha-response")) == 0 {
			logger.Warnf("no recaptca response present %v route: %v", r.URL.String(), r.FormValue("g-recaptcha-response"))
			utils.SetFlash(w, r, "info_flash", "Error: Failed to create request")
			http.Redirect(w, r, "/tools/broadcast", http.StatusSeeOther)
			return
		}

		valid, err := utils.ValidateReCAPTCHA(r.FormValue("g-recaptcha-response"))
		if err != nil || !valid {
			logger.Warnf("failed validating recaptcha %v route: %v", r.URL.String(), err)
			utils.SetFlash(w, r, "info_flash", "Error: Failed to create request")
			http.Redirect(w, r, "/tools/broadcast", http.StatusSeeOther)
			return
		}
	}

	jobData := r.FormValue("message")
	job, err := db.CreateNodeJob([]byte(jobData))
	if err != nil {
		errMsg := fmt.Sprintf("Error: %s", err)
		var userErr types.CreateNodeJobUserError
		if !errors.As(err, &userErr) {
			// only send error-message if its a UserError, otherwise just tell the user that something is wrong without details
			errMsg = "Sorry something went wrong :("
			logger.WithError(err).Errorf("failed creating a node-job")
		}
		utils.SetFlash(w, r, "info_flash", errMsg)
		http.Redirect(w, r, "/tools/broadcast", http.StatusSeeOther)
		return
	}

	url := fmt.Sprintf("/tools/broadcast/status/%s", job.ID)
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func BroadcastStatus(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "broadcaststatus.html")
	var tpl = templates.GetTemplate(templateFiles...)
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "tools", "/tools/broadcast/status", "Broadcast Status", templateFiles)

	vars := mux.Vars(r)

	job, err := db.GetNodeJob(vars["jobID"])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Not found", http.StatusNotFound)
		} else {
			logger.WithError(err).Errorf("error retrieving node-job")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	pageData := &types.BroadcastStatusPageData{}
	pageData.Job = job
	pageData.JobTypeLabel = FormatNodeJobType(job.Type)
	pageData.JobTitle = FormatNodeJobTitle(job.Type)
	pageData.JobJson = fmt.Sprintf("%v", string(job.RawData))

	validators, err := db.GetNodeJobValidatorInfos(job)
	if err != nil {
		logger.WithError(err).Errorf("error retrieving validator infos")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	pageData.Validators = &validators

	data.Data = pageData
	err = tpl.ExecuteTemplate(w, "layout", data)
	if handleTemplateError(w, r, "broadcast.go", "broadcast", "", err) != nil {
		return // an error has occurred and was processed
	}
}

func FormatNodeJobType(nodeJobType types.NodeJobType) string {
	label := "Unknown"
	switch nodeJobType {
	case types.BLSToExecutionChangesNodeJobType:
		label = "Set withdrawal address"
	case types.VoluntaryExitsNodeJobType:
		label = "Voluntary exit"
	}
	return label
}

func FormatNodeJobTitle(nodeJobType types.NodeJobType) string {
	label := "Transaction"
	switch nodeJobType {
	case types.BLSToExecutionChangesNodeJobType:
		label = "Withdrawal Credentials Change Request Job"
	case types.VoluntaryExitsNodeJobType:
		label = "Voluntary Exit Request Job"
	}
	return label
}
