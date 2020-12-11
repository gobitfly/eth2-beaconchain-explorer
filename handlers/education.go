package handlers

import (
	"eth2-exporter/utils"
	"html/template"
	"net/http"
)

var educationServicesTemplate = template.Must(template.New("educationServices").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/educationServices.html"))

func EducationServices(w http.ResponseWriter, r *http.Request) {
	var err error

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "educationServices", "/educationServices", "Ethereum 2.0 Education Services Overview")

	// pageData := &types.StakeWithUsPageData{}
	// pageData.FlashMessage, err = utils.GetFlash(w, r, "stake_flash")
	// if err != nil {
	// 	logger.Errorf("error retrieving flashes for advertisewithusform %v", err)
	// 	http.Error(w, "Internal server error", 503)
	// 	return
	// }
	// data.Data = pageData

	err = educationServicesTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
