package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"

	eth1common "github.com/ethereum/go-ethereum/common"
)

var poapTemplate = template.Must(template.ParseFiles("templates/layout.html", "templates/poap.html"))

// do not change existing entries, only append new entries
var poapClients = []string{"Prysm", "Lighthouse", "Teku", "Nimbus", "Lodestar"}

func Poap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	data := &types.PageData{
		HeaderAd: true,
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - POAP - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/poap",
			GATag:       utils.Config.Frontend.GATag,
		},
		ShowSyncingMessage: services.IsSyncing(),
		Active:             "more",
		Data: struct {
			PoapClients []string
		}{
			PoapClients: poapClients,
		},
		User:                  getUser(w, r),
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
		CurrentEpoch:          services.LatestEpoch(),
		CurrentSlot:           services.LatestSlot(),
		FinalizationDelay:     services.FinalizationDelay(),
	}
	err := poapTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

func PoapData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sqlRes := []struct {
		Graffiti       string
		Blockcount     uint64
		Validatorcount uint64
	}{}
	err := db.DB.Select(&sqlRes, `
		select 
			graffiti, 
			count(*) as blockcount,
			count(distinct proposer) as validatorcount
		from blocks
		where slot < 300000 and graffiti like 'poap%'
		group by graffiti`)
	if err != nil {
		logger.Errorf("error retrieving poap data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	// map[<eth1-addr>]map[<client>][<block-count>,<validator-count>]
	res := map[string]map[string][]uint64{}

	for _, d := range sqlRes {
		eth1Addr, client, err := decodePoapGraffiti(d.Graffiti)
		if err != nil {
			continue
		}
		_, exists := res[eth1Addr]
		if !exists {
			res[eth1Addr] = map[string][]uint64{}
			for _, name := range poapClients {
				res[eth1Addr][name] = []uint64{0, 0}
			}
		}
		res[eth1Addr][client][0] = d.Blockcount
		res[eth1Addr][client][1] = d.Validatorcount
	}

	// [<eth1-addr>, <total-block-count>, <total-validator-count>, <client1-block-count>, <client1-validator-count>, ..]
	tableData := [][]interface{}{}
	for eth1Addr, d := range res {
		f := []interface{}{eth1common.HexToAddress(eth1Addr).Hex(), uint64(0), uint64(0)}
		totalBlocks := uint64(0)
		totalValidators := uint64(0)
		for _, name := range poapClients {
			totalBlocks += d[name][0]
			totalValidators += d[name][1]
			f = append(f, d[name][0])
			f = append(f, d[name][1])
		}
		f[1] = totalBlocks
		f[2] = totalValidators
		tableData = append(tableData, f)
	}

	data := &types.DataTableResponse{
		Draw:            1,
		RecordsTotal:    1,
		RecordsFiltered: 1,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

func decodePoapGraffiti(graffiti string) (eth1Address, client string, err error) {
	if len(graffiti) != 32 {
		return "", "", fmt.Errorf("invalid graffiti-length")
	}
	b, err := base64.StdEncoding.DecodeString(graffiti[4:])
	if err != nil {
		return "", "", fmt.Errorf("failed decoding base64: %w", err)
	}
	str := fmt.Sprintf("%x", b)
	if len(str) != 42 {
		return "", "", fmt.Errorf("invalid length")
	}
	eth1Address = "0x" + str[:40]
	if !utils.IsValidEth1Address(eth1Address) {
		return "", "", fmt.Errorf("invalid eth1-address: %v", eth1Address)
	}
	clientID, err := strconv.ParseInt(str[40:], 16, 64)
	if err != nil {
		return "", "", fmt.Errorf("invalid clientID: %v: %w", str[40:], err)
	}
	if clientID < 0 || int64(len(poapClients)) < clientID {
		return "", "", fmt.Errorf("invalid clientID: %v", str[40:])
	}
	return eth1Address, poapClients[clientID], nil
}
