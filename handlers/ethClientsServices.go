package handlers

import (
	"eth2-exporter/db"
	ethclients "eth2-exporter/ethClients"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"
)

var ethClientsServicesTemplate = template.Must(template.New("ethClientsServices").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/ethClientsServices.html"))

func EthClientsServices(w http.ResponseWriter, r *http.Request) {
	var err error

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "services", "/ethClientsServices", "Ethereum Clients Services Overview")

	pageData := ethclients.GetEthClientData()
	pageData.CsrfField = csrf.TemplateField(r)
	// pageData.Banner = ethclients.GetBannerClients()
	if data.User.Authenticated {
		var dbData []string
		err = db.FrontendWriterDB.Select(&dbData,
			`select event_filter
			 from users_subscriptions 
			 where user_id = $1 AND event_name=$2
			`, data.User.UserID, string(types.EthClientUpdateEventName))
		if err != nil {
			logger.Errorf("error getting user subscriptions: %v route: %v", r.URL.String(), err)
		}

		for _, item := range dbData {
			switch item {
			case "geth":
				pageData.Geth.IsUserSubscribed = true
			case "openethereum":
				pageData.OpenEthereum.IsUserSubscribed = true
			case "nethermind":
				pageData.Nethermind.IsUserSubscribed = true
			case "besu":
				pageData.Besu.IsUserSubscribed = true
			case "lighthouse":
				pageData.Lighthouse.IsUserSubscribed = true
			case "prysm":
				pageData.Prysm.IsUserSubscribed = true
			case "teku":
				pageData.Teku.IsUserSubscribed = true
			case "nimbus":
				pageData.Nimbus.IsUserSubscribed = true
			case "erigon":
				pageData.Erigon.IsUserSubscribed = true
			case "rocketpool":
				pageData.RocketpoolSmartnode.IsUserSubscribed = true
			default:
				continue
			}
		}

	}

	data.Data = pageData

	err = ethClientsServicesTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}
