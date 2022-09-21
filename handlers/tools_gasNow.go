package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/templates"
	"eth2-exporter/utils"
	"html/template"
	"net/http"
	"time"
)

// Blocks will return information about blocks using a go template
func ToolsGasNow(w http.ResponseWriter, r *http.Request) {
	var toolsGasNowTemplate = template.Must(template.New("gasnow").Funcs(utils.GetTemplateFuncs()).ParseFS(templates.Files, "layout.html", "tools_gasNow.html", "components.html"))

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "unitConverter", "/unitConerter", "Unit Converter")

	res := []*struct {
		Ts      time.Time
		AvgFast float64
	}{}

	err := db.ReaderDb.Select(&res, ";")

	if err != nil {
		logger.Errorf("error retrieving gas price histors: %v", err)
	}

	resRet := []*struct {
		Ts      int64   `json:"ts"`
		AvgFast float64 `json:"fast"`
	}{}

	for _, r := range res {
		resRet = append(resRet, &struct {
			Ts      int64   `json:"ts"`
			AvgFast float64 `json:"fast"`
		}{
			Ts:      r.Ts.Unix(),
			AvgFast: r.AvgFast,
		})
	}
	data.Data = resRet

	err = toolsGasNowTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		return
	}
}
