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
	address = strings.ToLower(address)

	data := InitPageData(w, r, "address", "/address", "address")

	txns, err := db.BigtableClient.GetAddressTransactionsTableData(address, "", "")
	if err != nil {
		logger.WithError(err).Errorf("error getting eth1 block table data")
	}

	data.Data = types.Eth1AddressPageData{
		Address:           fmt.Sprintf("0x%s", address),
		TransactionsTable: txns,
	}

	if utils.Config.Frontend.Debug {
		eth1AddressTemplate = template.Must(template.New("address").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/execution/address.html"))
	}

	err = eth1AddressTemplate.ExecuteTemplate(w, "layout", data)
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
	address = strings.ToLower(address)

	// draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	// if err != nil {
	// 	logger.Errorf("error converting datatables data parameter from string to int for route %v: %v", r.URL.String(), err)
	// 	http.Error(w, "Internal server error", 503)
	// 	return
	// }
	// start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	// if err != nil {
	// 	// logger.Errorf("error converting datatables start parameter from string to int for route %v: %v", r.URL.String(), err)
	// 	// http.Error(w, "Internal server error", 503)
	// 	// return
	// 	start = 0
	// }
	// length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	// if err != nil {
	// 	// logger.Errorf("error converting datatables length parameter from string to int for route %v: %v", r.URL.String(), err)
	// 	// http.Error(w, "Internal server error", 503)
	// 	// return
	// 	length = 10
	// }
	// if length > 100 {
	// 	length = 100
	// }

	pageToken := q.Get("pageToken")

	// logger.Infof("PAGETOKEN: %v", pageToken)

	search := ""
	// logger.Infof("GETTING TRANSACTION table data for address: %v search: %v draw: %v start: %v length: %v", address, search, draw, start, length)
	data, err := db.BigtableClient.GetAddressTransactionsTableData(address, search, pageToken)
	if err != nil {
		logger.WithError(err).Errorf("error getting eth1 block table data")
	}

	// logger.Infof("GOT TX: %+v", data)

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

func Eth1AddressInternalTransactions(w http.ResponseWriter, r *http.Request) {
	logger.Infof("calling eth1 internal transactions data")
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	vars := mux.Vars(r)
	address := strings.Replace(vars["address"], "0x", "", -1)
	address = strings.ToLower(address)

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
	// logger.Infof("GETTING TRANSACTION table data for address: %v search: %v draw: %v start: %v length: %v", address, search, draw, start, length)
	data, err := GetAddressInternalTableData(address, search, draw, start, length)
	if err != nil {
		logger.WithError(err).Errorf("error getting eth1 block table data")
	}

	// logger.Infof("GOT TX: %+v", data)

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

func GetAddressInternalTableData(address string, search string, draw, start, length uint64) (*types.DataTableResponse, error) {
	transactions, _, err := db.BigtableClient.GetEth1ItxForAddress(address, db.FILTER_TIME, int64(length))
	if err != nil {
		return nil, err
	}

	count, err := db.BigtableClient.GetEth1InternalTxForAddressCount(address, db.FILTER_TIME)
	if err != nil {
		return nil, err
	}

	tableData := make([][]interface{}, len(transactions))
	for i, t := range transactions {
		from := utils.FormatHash(t.From)
		if fmt.Sprintf("%x", t.From) != address {
			from = utils.FormatAddressAsLink(t.From, "", false, false)
		}
		to := utils.FormatHash(t.To)
		if fmt.Sprintf("%x", t.To) != address {
			to = utils.FormatAddressAsLink(t.To, "", false, false)
		}
		tableData[i] = []interface{}{
			utils.FormatTransactionHash(t.ParentHash),
			utils.FormatTimestamp(t.Time.AsTime().Unix()),
			from,
			to,
			utils.FormatAmount(float64(new(big.Int).SetBytes(t.Value).Int64()), "ETH", 6),
			t.Type,
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

func Eth1AddressErc20Transactions(w http.ResponseWriter, r *http.Request) {
	logger.Infof("calling erc20 transactions data")
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	vars := mux.Vars(r)
	address := strings.Replace(vars["address"], "0x", "", -1)
	address = strings.ToLower(address)

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
	// logger.Infof("GETTING TRANSACTION table data for address: %v search: %v draw: %v start: %v length: %v", address, search, draw, start, length)
	data, err := GetAddressErc20TableData(address, search, draw, start, length)
	if err != nil {
		logger.WithError(err).Errorf("error getting eth1 block table data")
	}

	// logger.Infof("GOT TX: %+v", data)

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

func GetAddressErc20TableData(address string, search string, draw, start, length uint64) (*types.DataTableResponse, error) {
	transactions, _, err := db.BigtableClient.GetEth1ERC20ForAddress(address, db.FILTER_TIME, int64(length))
	if err != nil {
		return nil, err
	}

	count, err := db.BigtableClient.GetEth1ERC20TxForAddressCount(address, db.FILTER_TIME)
	if err != nil {
		return nil, err
	}

	tableData := make([][]interface{}, len(transactions))
	for i, t := range transactions {
		from := utils.FormatHash(t.From)
		if fmt.Sprintf("%x", t.From) != address {
			from = utils.FormatAddressAsLink(t.From, "", false, false)
		}
		to := utils.FormatHash(t.To)
		if fmt.Sprintf("%x", t.To) != address {
			to = utils.FormatAddressAsLink(t.To, "", false, false)
		}
		tableData[i] = []interface{}{
			utils.FormatTransactionHash(t.ParentHash),
			from,
			to,
			new(big.Int).SetBytes(t.Value),
			// utils.FormatAmount(float64(new(big.Int).SetBytes(t.Value).Int64()), "ETH", 6),
			utils.FormatAddressAsLink(t.TokenAddress, "", false, true),
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

func Eth1AddressErc721Transactions(w http.ResponseWriter, r *http.Request) {
	logger.Infof("calling erc721 data")
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	vars := mux.Vars(r)
	address := strings.Replace(vars["address"], "0x", "", -1)
	address = strings.ToLower(address)

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
	// logger.Infof("GETTING TRANSACTION table data for address: %v search: %v draw: %v start: %v length: %v", address, search, draw, start, length)
	data, err := GetAddressErc721TableData(address, search, draw, start, length)
	if err != nil {
		logger.WithError(err).Errorf("error getting eth1 block table data")
	}

	// logger.Infof("GOT TX: %+v", data)

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

func GetAddressErc721TableData(address string, search string, draw, start, length uint64) (*types.DataTableResponse, error) {
	transactions, _, err := db.BigtableClient.GetEth1ERC721ForAddress(address, db.FILTER_TIME, int64(length))
	if err != nil {
		return nil, err
	}

	count, err := db.BigtableClient.GetEth1ERC721TxForAddressCount(address, db.FILTER_TIME)
	if err != nil {
		return nil, err
	}

	tableData := make([][]interface{}, len(transactions))
	for i, t := range transactions {
		from := utils.FormatHash(t.From)
		if fmt.Sprintf("%x", t.From) != address {
			from = utils.FormatAddressAsLink(t.From, "", false, false)
		}
		to := utils.FormatHash(t.To)
		if fmt.Sprintf("%x", t.To) != address {
			to = utils.FormatAddressAsLink(t.To, "", false, false)
		}
		tableData[i] = []interface{}{
			utils.FormatTransactionHash(t.ParentHash),
			utils.FormatTimestamp(t.Time.AsTime().Unix()),
			from,
			to,
			utils.FormatAddressAsLink(t.TokenAddress, "", false, true),
			new(big.Int).SetBytes(t.TokenId).String(),
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

func Eth1AddressErc1155Transactions(w http.ResponseWriter, r *http.Request) {
	logger.Infof("calling eth1 erc1155 data")
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	vars := mux.Vars(r)
	address := strings.Replace(vars["address"], "0x", "", -1)
	address = strings.ToLower(address)

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
	// logger.Infof("GETTING TRANSACTION table data for address: %v search: %v draw: %v start: %v length: %v", address, search, draw, start, length)
	data, err := GetAddressErc1155TableData(address, search, draw, start, length)
	if err != nil {
		logger.WithError(err).Errorf("error getting eth1 block table data")
	}

	// logger.Infof("GOT TX: %+v", data)

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

func GetAddressErc1155TableData(address string, search string, draw, start, length uint64) (*types.DataTableResponse, error) {
	transactions, _, err := db.BigtableClient.GetEth1ERC1155ForAddress(address, db.FILTER_TIME, int64(length))
	if err != nil {
		return nil, err
	}

	count, err := db.BigtableClient.GetEth1ERC1155TxForAddressCount(address, db.FILTER_TIME)
	if err != nil {
		return nil, err
	}

	tableData := make([][]interface{}, len(transactions))
	for i, t := range transactions {
		from := utils.FormatHash(t.From)
		if fmt.Sprintf("%x", t.From) != address {
			from = utils.FormatAddressAsLink(t.From, "", false, false)
		}
		to := utils.FormatHash(t.To)
		if fmt.Sprintf("%x", t.To) != address {
			to = utils.FormatAddressAsLink(t.To, "", false, false)
		}
		tableData[i] = []interface{}{
			utils.FormatTransactionHash(t.ParentHash),
			utils.FormatTimestamp(t.Time.AsTime().Unix()),
			from,
			to,
			utils.FormatAddressAsLink(t.TokenAddress, "", false, true),
			new(big.Int).SetBytes(t.TokenId).String(),
			new(big.Int).SetBytes(t.Value).String(),
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

func Eth1Transaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	transactionHash := vars["hash"]

	http.Redirect(w, r, fmt.Sprintf("https://etherscan.io/tx/%s", transactionHash), http.StatusSeeOther)
}

func Eth1Block(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	block := vars["block"]

	http.Redirect(w, r, fmt.Sprintf("https://etherscan.io/block/%s", block), http.StatusSeeOther)
}
