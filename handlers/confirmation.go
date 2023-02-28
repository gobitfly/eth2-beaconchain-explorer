package handlers

import (
	"eth2-exporter/templates"
	"eth2-exporter/utils"
	"net/http"
)

// Will return the confirmation page
func Confirmation(w http.ResponseWriter, r *http.Request) {

	var confirmationTemplate = templates.GetTemplate("layout.html", "confirmation.html")

	w.Header().Set("Content-Type", "text/html")

	type confirmationPageData struct {
		Flashes []interface{}
	}

	// _, session, err := getUserSession(w, r)
	// if err != nil {
	// 	logger.Errorf("error retrieving session: %v", err)
	// 	http.Error(w, "Internal server error", http.StatusInternalServerError)
	// 	return
	// }
	// session.AddFlash("this is something")
	// session.AddFlash("Error: this is something else")
	// session.Save(r, w)

	pageData := confirmationPageData{}
	pageData.Flashes = utils.GetFlashes(w, r, authSessionName)

	if len(pageData.Flashes) <= 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := InitPageData(w, r, "confirmation", "/blocks", "Blocks")
	data.Data = pageData
	data.Meta.NoTrack = true

	if handleTemplateError(w, r, "confirmation.go", "Confirmation", "", confirmationTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}
