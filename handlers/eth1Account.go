package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

var eth1AddressTemplate = template.Must(template.New("address").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/execution/address.html"))

func Eth1Address(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(r)
	address := strings.Replace(vars["address"], "0x", "", -1)

	data := InitPageData(w, r, "address", "/address", "address")

	data.Data = types.Eth1AddressPageData{
		Address: fmt.Sprintf("0x%s", address),
	}

	if utils.Config.Frontend.Debug {
		eth1AddressTemplate = template.Must(template.New("address").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/execution/address.html"))
	}

	err := eth1AddressTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		return
	}
}

func Eth1AddressTransactions(w http.ResponseWriter, r *http.Request) {
	logger.Infof("calling eth1 transactions data")
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	vars := mux.Vars(r)
	address := strings.Replace(vars["address"], "0x", "", -1)

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		//logger.Errorf("error converting datatables data parameter from string to int for route %v: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables start parameter from string to int for route %v: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables length parameter from string to int for route %v: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
	if length > 100 {
		length = 100
	}

	search := ""
	logger.Infof("GETTING TRANSACTION table data for address: %v search: %v draw: %v start: %v length: %v", address, search, draw, start, length)
	data, err := GetAddressTransactionsTableData(address, search, draw, start, length)
	if err != nil {
		logger.WithError(err).Errorf("error getting eth1 block table data")
	}

	logger.Infof("GOT TX: %+v", data)

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

func GetAddressTransactionsTableData(address string, search string, draw, start, length uint64) (*types.DataTableResponse, error) {
	transactions, _, err := db.BigtableClient.GetEth1TxForAddress(address, db.FILTER_TIME, int64(length))
	if err != nil {
		return nil, err
	}

	count, err := db.BigtableClient.GetEth1TxForAddressCount(address, db.FILTER_TIME)
	if err != nil {
		return nil, err
	}

	tableData := make([][]interface{}, len(transactions))
	for i, t := range transactions {
		tableData[i] = []interface{}{
			utils.FormatHash(t.Hash),
			utils.FormatBlockNumber(t.BlockNumber),
			utils.FormatHash(t.From),
			utils.FormatHash(t.To),
			utils.FormatAmount(float64(new(big.Int).SetBytes(t.Value).Int64()), "ETH", 6),
			utils.FormatTime(t.Time.AsTime()),
		}
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    count,
		RecordsFiltered: count,
		Data:            tableData,
	}

	return data, nil
}

func GetAddressErc20TableData(address string, search string, draw, start, length uint64) (*types.DataTableResponse, error) {

	return nil, nil
}

func GetAddressErc721TableData(address string, search string, draw, start, length uint64) (*types.DataTableResponse, error) {

	return nil, nil
}

func GetAddressErc1155TableData(address string, search string, draw, start, length uint64) (*types.DataTableResponse, error) {

	return nil, nil
}
