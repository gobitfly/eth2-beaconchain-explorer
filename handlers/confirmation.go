package handlers

import (
	"eth2-exporter/utils"
	"html/template"
	"net/http"
)

var confirmationTemplate = template.Must(template.New("confirmation").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/confirmation.html"))

// Blocks will return information about blocks using a go template
func Confirmation(w http.ResponseWriter, r *http.Request) {
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

	err := confirmationTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}
