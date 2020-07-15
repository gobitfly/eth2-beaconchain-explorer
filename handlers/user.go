package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"html/template"
	"net/http"
)

var userTemplate = template.Must(template.New("user").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/user/settings.html"))

// User renders the user-template
func UserSettings(w http.ResponseWriter, r *http.Request) {
	userSettingsData := &types.UserSettingsPageData{}

	// TODO: remove before production
	userTemplate = template.Must(template.New("user").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/user/settings.html"))

	w.Header().Set("Content-Type", "text/html")
	authData := getAuthData(w, r)

	email, errQ := db.GetUserEmailById(authData.User.UserID)
	if errQ != nil {
		logger.Errorf("error executing query GetUserEmailById for %v route: %v", r.URL.String(), errQ)
	}

	userSettingsData.Email = email
	userSettingsData.AuthData = authData

	data := &types.PageData{
		Meta: &types.Meta{
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/user",
		},
		Active:                "user",
		Data:                  userSettingsData,
		User:                  authData.User,
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
		CurrentEpoch:          services.LatestEpoch(),
		CurrentSlot:           services.LatestSlot(),
		FinalizationDelay:     services.FinalizationDelay(),
	}
	err := userTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
