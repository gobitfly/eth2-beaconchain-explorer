package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var ethTemplates = template.Must(template.New("deposits").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/ethOneDeposit.html"))

// Blocks will return information about deposits using a go template
func EthOneDeposits(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html")

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - Eth1 Deposits - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/deposits/eth1",
		},
		ShowSyncingMessage:    services.IsSyncing(),
		Active:                "ethOneDeposit",
		Data:                  nil,
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
		CurrentEpoch:          services.LatestEpoch(),
		CurrentSlot:           services.LatestSlot(),
	}

	err := ethTemplates.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

// BlocksData will return information about blocks
func EthOneData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	search := q.Get("search[value]")
	search = strings.Replace(search, "0x", "", -1)

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	if length > 100 {
		length = 100
	}

	depositCount, err := db.GetEth1DepositsCount()
	if err != nil {
		logger.Errorf("GetEth1DepositsCount error retrieving eth1_deposit data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	deposits, err := db.GetEth1DepositsJoinEth2Deposits("", length, start)
	if err != nil {
		logger.Errorf("GetEth1Deposits error retrieving eth1_deposit data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	logger.Printf("found %d results", len(deposits))

	tableData := make([][]interface{}, len(deposits))
	for i, d := range deposits {
		tableData[i] = []interface{}{
			fmt.Sprintf("%#x", d.FromAddress),
			fmt.Sprintf("%#x", d.PublicKey),
			fmt.Sprintf("%g ETH", float64(d.Amount)/float64(1000000000)),
			fmt.Sprintf("%#x", d.TxHash),
			d.BlockTs,
			d.BlockNumber,
			d.Activated,
			// d.TxIndex,
			// d.TxInput,
			// utils.FormatPublicKey(d.WithdrawalCredentials),
			// d.MerkletreeIndex,
		}
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    depositCount,
		RecordsFiltered: depositCount,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
