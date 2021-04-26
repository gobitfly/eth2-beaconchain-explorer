package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/utils"
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"
)

var validatorRewardsServicesTemplate = template.Must(template.New("validatorRewards").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/validatorRewards.html"))

type rewardsResp struct {
	Currencies []string
	CsrfField  template.HTML
}

func ValidatorRewards(w http.ResponseWriter, r *http.Request) {
	var err error

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "services", "/rewards", "Ethereum Validator rewards")

	var supportedCurrencies []string
	err = db.DB.Select(&supportedCurrencies,
		`select column_name 
			from information_schema.columns 
			where table_name = 'price'`)
	if err != nil {
		logger.Errorf("error getting eth1-deposits-distribution for stake pools: %w", err)
	}

	data.Data = rewardsResp{Currencies: supportedCurrencies, CsrfField: csrf.TemplateField(r)}

	err = validatorRewardsServicesTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
