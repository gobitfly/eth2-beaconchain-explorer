package handlers

import (
	"eth2-exporter/templates"
	"net/http"
)

// Will return the slot finder page
func SlotFinder(w http.ResponseWriter, r *http.Request) {

	var template = templates.GetTemplate("layout.html", "slot/slotfinder.html", "slot/components/slotfinder.html", "slot/components/upgradescheduler.html")

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "blockchain", "/slotFinder", "Slot Finder")

	if handleTemplateError(w, r, "slot_finder.go", "Slot Finder", "", template.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}
