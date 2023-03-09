package handlers

import (
	"eth2-exporter/templates"
	"net/http"
)

// Will return the slot finder page
func SlotFinder(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles,
		"slot/slotfinder.html",
		"slot/components/slotfinder.html",
		"slot/components/upgradescheduler.html")
	var template = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "blockchain", "/slotFinder", "Slot Finder", templateFiles)

	if handleTemplateError(w, r, "slot_finder.go", "Slot Finder", "", template.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}
