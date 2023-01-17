package handlers

import (
	"eth2-exporter/templates"
	"eth2-exporter/utils"
	"html/template"
	"net/http"
	"os"
	"path"
)

// Imprint will show the imprint data using a go template
func Imprint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "imprint", "/imprint", "Imprint")
	data.HeaderAd = true

	if handleTemplateError(w, r, getImprintTemplate(getImprintPath()).ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

func CheckAndPreloadImprint() error {
	imprintPath := getImprintPath()
	if len(imprintPath) > 0 {
		_, err := os.Stat(imprintPath) // check file exists
		if err != nil {
			return err
		}
	}

	getImprintTemplate(imprintPath) // preload
	return nil
}

func getImprintPath() string {
	if utils.Config.Frontend.LegalDir == "" {
		return utils.Config.Frontend.Imprint
	}
	return path.Join(utils.Config.Frontend.LegalDir, "index.html")
}

func getImprintTemplate(path string) *template.Template {
	if len(path) == 0 {
		return templates.GetTemplate("layout.html", "imprint.example.html")
	}

	var imprintTemplate = templates.GetTemplate("layout.html")
	imprintTemplate = templates.AddTemplateFile(imprintTemplate, path)
	return imprintTemplate
}
