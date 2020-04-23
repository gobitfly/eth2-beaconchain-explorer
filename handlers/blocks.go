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

var blocksTemplate = template.Must(template.New("blocks").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/blocks.html"))

// Blocks will return information about blocks using a go template
func Blocks(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html")

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - Blocks - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/blocks",
		},
		ShowSyncingMessage:    services.IsSyncing(),
		Active:                "blocks",
		Data:                  nil,
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
	}

	err := blocksTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

// BlocksData will return information about blocks
func BlocksData(w http.ResponseWriter, r *http.Request) {
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

	var blocksCount uint64
	var blocks []*types.IndexPageDataBlocks
	if search == "" {
		err = db.DB.Get(&blocksCount, "SELECT MAX(slot) + 1 FROM blocks")
		if err != nil {
			logger.Errorf("error retrieving max slot number: %v", err)
			http.Error(w, "Internal server error", 503)
			return
		}

		startSlot := blocksCount - start
		endSlot := blocksCount - start - length + 1

		if startSlot > 9223372036854775807 {
			startSlot = blocksCount
		}
		if endSlot > 9223372036854775807 {
			endSlot = 0
		}
		err = db.DB.Select(&blocks, `
			SELECT 
				blocks.epoch, 
				blocks.slot, 
				blocks.proposer, 
				blocks.blockroot, 
				blocks.parentroot, 
				blocks.attestationscount, 
				blocks.depositscount, 
				blocks.voluntaryexitscount, 
				blocks.proposerslashingscount, 
				blocks.attesterslashingscount, 
				blocks.status, 
				COALESCE((SELECT SUM(ARRAY_LENGTH(validators, 1)) FROM blocks_attestations WHERE beaconblockroot = blocks.blockroot), 0) AS votes,
				blocks.graffiti
			FROM blocks 
			WHERE blocks.slot >= $1 AND blocks.slot <= $2 
			ORDER BY blocks.slot DESC`, endSlot, startSlot)
	} else {
		err = db.DB.Get(&blocksCount, "SELECT count(*) FROM blocks WHERE CAST(blocks.slot as text) LIKE $1 OR graffiti LIKE convert_to($2, $3)", search+"%", "%"+search+"%", "UTF-8")
		if err != nil {
			logger.Errorf("error retrieving max slot number: %v", err)
			http.Error(w, "Internal server error", 503)
			return
		}
		err = db.DB.Select(&blocks, `
			SELECT 
				blocks.epoch, 
				blocks.slot, 
				blocks.proposer, 
				blocks.blockroot, 
				blocks.parentroot, 
				blocks.attestationscount, 
				blocks.depositscount, 
				blocks.voluntaryexitscount, 
				blocks.proposerslashingscount, 
				blocks.attesterslashingscount, 
				blocks.status, 
				COALESCE((SELECT SUM(ARRAY_LENGTH(validators, 1)) FROM blocks_attestations WHERE beaconblockroot = blocks.blockroot), 0) AS votes, 
				blocks.graffiti 
			FROM blocks
			WHERE slot IN (
				SELECT slot 
				FROM blocks
				WHERE CAST(blocks.slot as text) LIKE $1 OR graffiti LIKE convert_to($2, $3) 
				ORDER BY blocks.slot ASC 
				LIMIT $4 
				OFFSET $5
			) ORDER BY blocks.slot DESC
			`, search+"%", "%"+search+"%", "UTF-8", length, start)
	}

	if err != nil {
		logger.Errorf("error retrieving block data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	tableData := make([][]interface{}, len(blocks))
	for i, b := range blocks {
		tableData[i] = []interface{}{
			b.Epoch,
			b.Slot,
			utils.FormatBlockStatus(b.Status),
			utils.SlotToTime(b.Slot).Unix(),
			utils.FormatValidator(b.Proposer),
			fmt.Sprintf("%x", b.BlockRoot),
			b.Attestations,
			b.Deposits,
			fmt.Sprintf("%v / %v", b.Proposerslashings, b.Attesterslashings),
			b.Exits,
			b.Votes,
			fmt.Sprintf("%x", b.Graffiti),
		}
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    blocksCount,
		RecordsFiltered: blocksCount,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
