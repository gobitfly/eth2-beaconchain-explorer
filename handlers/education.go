package handlers

import (
	"eth2-exporter/templates"
	"eth2-exporter/utils"
	"html/template"
	"net/http"
)

var educationServicesTemplate = template.Must(template.New("educationServices").Funcs(utils.GetTemplateFuncs()).ParseFS(templates.Files, "layout.html", "educationServices.html"))

func EducationServices(w http.ResponseWriter, r *http.Request) {
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
