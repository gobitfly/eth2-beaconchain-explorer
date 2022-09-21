package handlers

import (
	"eth2-exporter/templates"
	"net/http"
)

func EducationServices(w http.ResponseWriter, r *http.Request) {

	var educationServicesTemplate = templates.GetTemplate("layout.html", "educationServices.html")

	var err error

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "services", "/educationServices", "Ethereum 2.0 Education Services Overview")

	err = educationServicesTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}
