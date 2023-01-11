package handlers

import (
	"eth2-exporter/templates"
	"net/http"
)

func EducationServices(w http.ResponseWriter, r *http.Request) {

	var educationServicesTemplate = templates.GetTemplate("layout.html", "educationServices.html")

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "services", "/educationServices", "Ethereum Education Services Overview")

	if handleTemplateError(w, r, educationServicesTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}
