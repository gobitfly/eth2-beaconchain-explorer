package handlers

import (
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"net/http"

	"github.com/gorilla/mux"
)

// Will return the EnsPage
func EnsSearch(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "ensSearch.html")
	var ensSearchTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "enssearch", "/ens", "Ens search", templateFiles)

	vars := mux.Vars(r)
	search := vars["search"]

	result, err := GetEnsDomain(search)

	var pageData types.EnsSearchPageData
	if err != nil {
		pageData.Error = "No matching ENS registration found"
	} else {
		pageData.Result = result
	}
	pageData.Search = search

	data.Data = pageData

	if handleTemplateError(w, r, "ensSearch.go", "EnsSearch", "", ensSearchTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}
