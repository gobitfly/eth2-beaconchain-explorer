package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"golang.org/x/sync/errgroup"
)

var eth1TokenTemplate = template.Must(template.New("token").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/execution/token.html"))

func Eth1Token(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(r)
	address := strings.Replace(vars["token"], "0x", "", -1)
	address = strings.ToLower(address)

	data := InitPageData(w, r, "token", "/token", "token")

	g := new(errgroup.Group)
	g.SetLimit(2)

	var txns *types.DataTableResponse
	// var holders *types.DataTableResponse

	g.Go(func() error {
		var err error
		txns, err = db.BigtableClient.GetTokenTransactionsTableData(address, "", "")
		if err != nil {
			return err
		}
		return nil
	})
	// g.Go(func() error {
	// 	var err error
	// 	holders, err = db.BigtableClient.GetTokenHoldersTableData(address, "", "")
	// 	if err != nil {
	// 		return err
	// 	}
	// 	return nil
	// })

	if err := g.Wait(); err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}

	data.Data = types.Eth1TokenPageData{
		Address:        address,
		TransfersTable: txns,
		// HoldersTable: holders,
	}

	if utils.Config.Frontend.Debug {
		eth1TokenTemplate = template.Must(template.New("address").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/execution/token.html"))
	}

	err := eth1TokenTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}

}

func Eth1TokenTransfers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	vars := mux.Vars(r)
	address := strings.Replace(vars["address"], "0x", "", -1)
	address = strings.ToLower(address)

	pageToken := q.Get("pageToken")

	search := ""
	// logger.Infof("GETTING TRANSACTION table data for address: %v search: %v draw: %v start: %v length: %v", address, search, draw, start, length)
	data, err := db.BigtableClient.GetTokenTransactionsTableData(address, search, pageToken)
	if err != nil {
		logger.WithError(err).Errorf("error getting eth1 block table data")
	}

	// logger.Infof("GOT TX: %+v", data)

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}
