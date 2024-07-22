package handlers

import (
	"net/http"

	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/templates"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
)

func Graffitiwall(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "graffitiwall.html")
	var graffitiwallTemplate = templates.GetTemplate(templateFiles...)

	var err error

	w.Header().Set("Content-Type", "text/html")

	var graffitiwallData []*types.GraffitiwallData

	// only fetch latest entry for each pixel
	err = db.ReaderDb.Select(&graffitiwallData, "SELECT DISTINCT ON (x, y) x, y, color, slot, validator from graffitiwall ORDER BY x, y, slot DESC")

	if err != nil {
		utils.LogError(err, "error retrieving graffitiwall data", 0)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := InitPageData(w, r, "more", "/graffitiwall", "Graffitiwall", templateFiles)
	data.Data = graffitiwallData

	if handleTemplateError(w, r, "graffitiwall.go", "Graffitiwall", "", graffitiwallTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}
