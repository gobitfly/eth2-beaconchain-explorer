package handlers

import (
	"eth2-exporter/utils"
	"html/template"
	"net/http"
)

var eth1AccountTemplate = template.Must(template.New("account").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/execution/account.html"))

func Eth1Account(w http.ResponseWriter, r *http.Request) {

}

func Eth1AccountTransactions(w http.ResponseWriter, r *http.Request) {

}
