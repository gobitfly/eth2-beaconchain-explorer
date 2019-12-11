package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"html/template"
	"net/http"
)

var dashboardTemplate = template.Must(template.New("dashboard").ParseFiles("templates/layout.html", "templates/dashboard.html"))
var dashboardNotFoundTemplate = template.Must(template.New("dashboardnotfound").ParseFiles("templates/layout.html", "templates/dashboardnotfound.html"))

func Dashboard(w http.ResponseWriter, r *http.Request) {
	dashboardTemplate = template.Must(template.New("dashboard").ParseFiles("templates/layout.html", "templates/dashboard.html"))
	w.Header().Set("Content-Type", "text/html")
	// validatorsQuery := [...]string{"1", "2", "3", "4", "5"}

	// validators := r.URL.Query().Get("v")
	dashboardPageData := types.DashboardPageData{}

	var err error
	var validators []*types.ValidatorsPageDataValidators

	err = db.DB.Select(&validators, `SELECT 
	epoch, 
	activationepoch, 
	exitepoch 
	FROM validator_set 
	WHERE epoch = $1 and validatorindex in ('1', '2', '3', '4', '5')
	ORDER BY validatorindex`, services.LatestEpoch())

	if err != nil {
		logger.Printf("Error retrieving validators data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	for _, validator := range validators {
		if validator.Epoch > validator.ExitEpoch {
			dashboardPageData.EjectedCount++
		} else if validator.Epoch < validator.ActivationEpoch {
			dashboardPageData.PendingCount++
		} else {
			dashboardPageData.ActiveCount++
		}
	}

	dashboardPageData.Validators = validators

	dashboardPageData.Title = "Hello, World"

	data := &types.PageData{
		Meta: &types.Meta{
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
		},
		ShowSyncingMessage: services.IsSyncing(),
		Active:             "dashboard",
		Data:               nil,
	}

	data.Data = dashboardPageData

	err = dashboardTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Fatalf("Error executing template for %v route: %v", r.URL.String(), err)
	}

}
