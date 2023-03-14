package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/csrf"
)

// Load Ad Configuration page
func AdConfiguration(w http.ResponseWriter, r *http.Request) {
	if isAdmin, _ := handleAdminPermissions(w, r); !isAdmin {
		return
	}

	templateFiles := append(layoutTemplateFiles, "user/ad_configuration.html")
	var userTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	configs, err := db.GetAdConfigurations()

	if err != nil {
		utils.LogError(err, "error loading the ad configuration", 0)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := InitPageData(w, r, "user", "/user/ad_configuration", "Ad Configuration", templateFiles)
	pageData := types.AdConfigurationPageData{}
	pageData.CsrfField = csrf.TemplateField(r)
	pageData.Configurations = configs
	pageData.TemplateNames = templates.GetTemplateNames()
	pageData.New = types.AdConfig{
		InsertMode:     "replace",
		TemplateId:     "index/index.html",
		JQuerySelector: "#r-banner",
		Enabled:        true,
	}
	data.Data = pageData

	if handleTemplateError(w, r, "ad_configuration.go", "AdConfiguration", "", userTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

// Insert / Update Ad configuration
func AdConfigurationPost(w http.ResponseWriter, r *http.Request) {
	if isAdmin, _ := handleAdminPermissions(w, r); !isAdmin {
		return
	}

	err := r.ParseForm()
	if err != nil {
		utils.LogError(err, "error parsing form", 0)
		http.Redirect(w, r, "/user/ad_configuration?error=parsingForm", http.StatusSeeOther)
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
			utils.LogError(nil, "error with provided html content", 0)
			http.Redirect(w, r, "/user/ad_configuration?error=noHtmlContent", http.StatusSeeOther)
			return
		}
	} else {
		bannerId, err = strconv.ParseUint(r.FormValue(`bannerId`), 0, 64)
		if err != nil || bannerId == 0 {
			utils.LogError(err, "error no bannerId provided", 0)
			http.Redirect(w, r, "/user/ad_configuration?error=noBannerId", http.StatusSeeOther)
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
		ForAllUsers:     len(r.FormValue(`forAllUsers`)) > 0,
		BannerId:        bannerId,
		HtmlContent:     htmlContent,
	}

	if len(adConfig.Id) == 0 {
		adConfig.Id = uuid.New().String()
		err = db.InsertAdConfigurations(adConfig)
		if err != nil {
			utils.LogError(err, "error inserting new ad config", 0)
			http.Redirect(w, r, "/user/ad_configuration?error=insertingConfig", http.StatusSeeOther)
			return
		}

	} else {
		err = db.UpdateAdConfiguration(adConfig)
		if err != nil {
			utils.LogError(err, "error updating ad config", 0)
			http.Redirect(w, r, "/user/ad_configuration?error=updatingConfig", http.StatusSeeOther)
			return
		}
	}

	http.Redirect(w, r, "/user/ad_configuration", http.StatusSeeOther)
}

// Delete Ad configuration
func AdConfigurationDeletePost(w http.ResponseWriter, r *http.Request) {
	if isAdmin, _ := handleAdminPermissions(w, r); !isAdmin {
		return
	}

	err := r.ParseForm()
	if err != nil {
		utils.LogError(err, "error parsing form", 0)
		http.Redirect(w, r, "/user/ad_configuration?error=parsingForm", http.StatusSeeOther)
		return
	}
	id := r.FormValue(`id`)
	if len(id) == 0 {
		utils.LogError(err, "error no id provided", 0)
		http.Redirect(w, r, "/user/ad_configuration?error=noTemplateId", http.StatusSeeOther)
		return
	}

	err = db.DeleteAdConfiguration(id)
	if err != nil {
		utils.LogError(err, "error deleting ad config", 0)
		http.Redirect(w, r, "/user/ad_configuration?error=notDeleted", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/user/ad_configuration", http.StatusSeeOther)
}
