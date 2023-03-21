package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"net/http"

	"github.com/gorilla/csrf"
)

// Load Explorer Configuration page
func ExplorerConfiguration(w http.ResponseWriter, r *http.Request) {
	if isAdmin, _ := handleAdminPermissions(w, r); !isAdmin {
		return
	}

	templateFiles := append(layoutTemplateFiles, "user/explorer_configuration.html")
	var userTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	configs, err := services.GetExplorerConfigurationsWithDefaults()
	if err != nil {
		utils.LogError(err, "error loading the explorer configuration", 0)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := InitPageData(w, r, "user", "/user/explorer_configuration", "Explorer Configuration", templateFiles)

	pageData := types.ExplorerConfigurationPageData{}
	pageData.Configurations = configs
	pageData.CsrfField = csrf.TemplateField(r)
	data.Data = pageData

	if handleTemplateError(w, r, "ad_configuration.go", "AdConfiguration", "", userTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

// Insert / Update Ad configuration
func ExplorerConfigurationPost(w http.ResponseWriter, r *http.Request) {
	if isAdmin, _ := handleAdminPermissions(w, r); !isAdmin {
		return
	}

	err := r.ParseForm()
	if err != nil {
		utils.LogError(err, "error parsing form", 0)
		http.Redirect(w, r, "/user/explorer_configuration?error=parsingForm", http.StatusSeeOther)
		return
	}

	configs := []types.ExplorerConfig{}

	for category, keyMap := range services.DefaultExplorerConfiguration {

		for key, configValue := range keyMap {
			value := r.FormValue(fmt.Sprintf(`%v-%v`, category, key))
			config := types.ExplorerConfig{Category: category, Key: key}
			config.DataType = configValue.DataType
			config.Value = value
			configs = append(configs, config)
		}
	}

	err = db.SaveExplorerConfiguration(configs)

	if err != nil {
		utils.LogError(err, "error saving explorer configuration", 0)
		http.Redirect(w, r, "/user/explorer_configuration?error=saveFailed", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/user/explorer_configuration", http.StatusSeeOther)
}
