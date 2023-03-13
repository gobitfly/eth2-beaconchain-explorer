package handlers

import (
	"encoding/base64"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"

	eth1common "github.com/ethereum/go-ethereum/common"
)

// do not change existing entries, only append new entries
var poapClients = []string{"Prysm", "Lighthouse", "Teku", "Nimbus", "Lodestar"}
var poapMaxSlot = uint64(300000)

var poapData atomic.Value
var poapDataEpoch uint64

func Poap(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "poap.html")
	var poapTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "more", "/poap", "POAP", templateFiles)
	data.Data = struct {
		PoapClients []string
	}{
		PoapClients: poapClients,
	}

	if handleTemplateError(w, r, "poap.go", "Poap", "", poapTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

func PoapData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	latestEpoch := services.LatestEpoch()
	latestPoapDataEpoch := atomic.LoadUint64(&poapDataEpoch)
	latestPoapData := poapData.Load()

	if latestPoapData != nil && (latestEpoch < latestPoapDataEpoch || latestEpoch == 0 || latestEpoch > utils.EpochOfSlot(poapMaxSlot)) {
		err := json.NewEncoder(w).Encode(latestPoapData.(*types.DataTableResponse))
		if err != nil {
			logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", http.StatusServiceUnavailable)
			return
		}
		return
	}

	sqlRes := []struct {
		Graffiti       string
		Blockcount     uint64
		Validatorcount uint64
	}{}
	err := db.ReaderDb.Select(&sqlRes, `
		select 
			graffiti, 
			count(*) as blockcount,
			count(distinct proposer) as validatorcount
		from blocks
		where slot <= $1 and graffiti like 'poap%'
		group by graffiti`, poapMaxSlot)
	if err != nil {
		logger.Errorf("error retrieving poap data: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
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

	poapData.Store(data)
	atomic.StoreUint64(&poapDataEpoch, latestEpoch)

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
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
