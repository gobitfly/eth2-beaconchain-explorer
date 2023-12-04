package handlers

import (
	"eth2-exporter/templates"
	"eth2-exporter/utils"
	"net/http"
)

// Will return the confirmation page
func Confirmation(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "confirmation.html")
	var confirmationTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	type confirmationPageData struct {
		Flashes []interface{}
	}

	pageData := confirmationPageData{}
	pageData.Flashes = utils.GetFlashes(w, r, authSessionName)

	if len(pageData.Flashes) <= 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := InitPageData(w, r, "confirmation", "/blocks", "Blocks", templateFiles)
	data.Data = pageData
	data.Meta.NoTrack = true

	if handleTemplateError(w, r, "confirmation.go", "Confirmation", "", confirmationTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}
