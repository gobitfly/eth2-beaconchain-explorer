package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"html/template"
	"net/http"
)

var graffitiwallTemplate = template.Must(template.New("vis").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/graffitiwall.html"))

func Graffitiwall(w http.ResponseWriter, r *http.Request) {

	var err error

	w.Header().Set("Content-Type", "text/html")

	var graffitiwallData []*types.GraffitiwallData

	err = db.ReaderDb.Select(&graffitiwallData, "select x, y, color, slot, validator from graffitiwall")

	if err != nil {
		logger.Errorf("error retrieving block tree data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	data := InitPageData(w, r, "more", "/graffitiwall", "Graffitiwall")
	data.HeaderAd = true
	data.Data = graffitiwallData

	err = graffitiwallTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
