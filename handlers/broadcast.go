package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func Broadcast(w http.ResponseWriter, r *http.Request) {
	var tpl = templates.GetTemplate("layout.html", "components/bannerGeneric.html", "broadcast.html")
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "tools", "/broadcast", "Change Withdrawal Credentials")
	pageData := &types.BroadcastPageData{}
	pageData.RecaptchaKey = utils.Config.Frontend.RecaptchaSiteKey
	data.Data = pageData
	err := tpl.ExecuteTemplate(w, "layout", data)
	if handleTemplateError(w, r, "broadcast.go", "Broadcast", "", err) != nil {
		return // an error has occurred and was processed
	}
}

func BroadcastPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		logger.Errorf("error parsing form: %v", err)
		utils.SetFlash(w, r, "broadcast_flash", "Error: invalid form submitted")
		http.Redirect(w, r, "/tools/broadcast", http.StatusSeeOther)
		return
	}
	d := r.FormValue("bls_to_execution_changes")
	job, err := db.CreateBLSToExecutionChangesNodeJob([]byte(d))
	if err != nil {
		logger.WithFields(logrus.Fields{"error": err}).Warnf("failed creating job")
		utils.SetFlash(w, r, "broadcast_flash", "Error: invalid form submitted")
		http.Redirect(w, r, "/tools/broadcast", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/tools/broadcast/status/%v", job.ID), http.StatusSeeOther)
}

func BroadcastStatus(w http.ResponseWriter, r *http.Request) {
	var tpl = templates.GetTemplate("layout.html", "components/bannerGeneric.html", "broadcaststatus.html")
	w.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(r)
	jobID := vars["id"]
	job, err := db.GetNodeJob(jobID)
	if err != nil {
		logger.Errorf("error getting nodeJob: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	data := InitPageData(w, r, "tools", "/broadcast", "Change Withdrawal Credentials")
	data.Data = types.BroadcastStatusPageData{
		Job: job,
	}
	err = tpl.ExecuteTemplate(w, "layout", data)
	if handleTemplateError(w, r, "broadcast.go", "Broadcast", "", err) != nil {
		return // an error has occurred and was processed
	}
}
