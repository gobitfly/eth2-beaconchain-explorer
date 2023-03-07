package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/csrf"
)

// Load Ad Configuration page
func AdConfiguration(w http.ResponseWriter, r *http.Request) {
	var userTemplate = templates.GetTemplate("layout.html", "user/ad_configuration.html")

	w.Header().Set("Content-Type", "text/html")

	user, _, err := getUserSession(r)
	if err != nil {
		logger.Errorf("error retrieving session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user.UserGroup != "ADMIN" {
		http.Error(w, "Insufficient privilleges", http.StatusUnauthorized)
		return
	}

	type notificationConfig struct {
		Target  string
		Content string
		Enabled bool
	}

	var configs []*types.AdConfig

	configs, err = db.GetAdConfigurations()

	if err != nil {
		logger.Errorf("error loading the ad configuration: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := InitPageData(w, r, "user", "/user/ad_configuration", "Ad Configuration")
	pageData := types.AdConfigurationPageData{}
	pageData.CsrfField = csrf.TemplateField(r)
	pageData.Configurations = configs
	pageData.New = types.AdConfig{
		InsertMode: "insert",
		Enabled:    true,
	}
	data.Data = pageData

	if handleTemplateError(w, r, "ad_configuration.go", "AdConfiguration", "", userTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

// Insert / Update Ad configuration
func AdConfigurationPost(w http.ResponseWriter, r *http.Request) {
	user, _, err := getUserSession(r)
	if err != nil {
		logger.Errorf("error retrieving session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user.UserGroup != "ADMIN" {
		http.Error(w, "Insufficient privilleges", http.StatusUnauthorized)
		return
	}

	err = r.ParseForm()
	if err != nil {
		logger.Errorf("error parsing form: %v", err)
		http.Redirect(w, r, "/user/ad_configuration", http.StatusSeeOther)
		return
	}
	id := r.FormValue(`id`)

	refreshInterval, err := strconv.ParseUint(r.FormValue(`refreshInterval`), 0, 64)
	if err != nil {
		refreshInterval = 0
	}

	var bannerId uint64
	var htmlContent = ""
	if len(r.FormValue(`useHtmlContent`)) > 0 {
		htmlContent = r.FormValue(`htmlContent`)

		if len(htmlContent) == 0 {
			logger.Errorf("error no html content provided: %v", err)
			http.Redirect(w, r, "/user/ad_configuration", http.StatusSeeOther)
			return
		}
	} else {
		bannerId, err = strconv.ParseUint(r.FormValue(`bannerId`), 0, 64)
		if err != nil || bannerId == 0 {
			logger.Errorf("error no bannerId provided: %v", err)
			http.Redirect(w, r, "/user/ad_configuration", http.StatusSeeOther)
			return
		}
	}

	adConfig := types.AdConfig{
		Id:              id,
		TemplateId:      r.FormValue(`templateId`),
		JQuerySelector:  r.FormValue(`jQuerySelector`),
		InsertMode:      r.FormValue(`insertMode`),
		RefreshInterval: refreshInterval,
		Enabled:         len(r.FormValue(`enabled`)) > 0,
		BannerId:        bannerId,
		HtmlContent:     htmlContent,
	}
	logger.Infof(`adConfig %v`, adConfig)
	if len(adConfig.Id) == 0 {
		adConfig.Id = uuid.New().String()
		err = db.InsertAdConfigurations(adConfig)
		if err != nil {
			logger.Errorf("error inserting new ad config: %v", err)
		}

	} else {
		err = db.UpdateAdConfiguration(adConfig)
		if err != nil {
			logger.Errorf("error updating ad config: %v", err)
		}
	}

	http.Redirect(w, r, "/user/ad_configuration", http.StatusSeeOther)
}

// Delete Ad configuration
func AdConfigurationDeletePost(w http.ResponseWriter, r *http.Request) {
	user, _, err := getUserSession(r)
	if err != nil {
		logger.Errorf("error retrieving session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user.UserGroup != "ADMIN" {
		http.Error(w, "Insufficient privilleges", http.StatusUnauthorized)
		return
	}

	err = r.ParseForm()
	if err != nil {
		logger.Errorf("error parsing form: %v", err)
		http.Redirect(w, r, "/user/ad_configuration", http.StatusSeeOther)
		return
	}
	id := r.FormValue(`id`)
	if len(id) == 0 {
		logger.Errorf("error no id provided: %v", err)
		http.Redirect(w, r, "/user/ad_configuration", http.StatusSeeOther)
		return
	}

	err = db.DeleteAdConfiguration(id)
	if err != nil {
		logger.Errorf("error deleting ad config: %v", err)
	}

	http.Redirect(w, r, "/user/ad_configuration", http.StatusSeeOther)
}
