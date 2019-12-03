package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

var blocksTemplate = template.Must(template.New("blocks").ParseFiles("templates/layout.html", "templates/blocks.html"))

func Blocks(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html")

	var blocksTreeData []*types.BlocksTreeData
	err := db.DB.Select(&blocksTreeData, "select slot, blockroot, parentroot from blocks where status = '1' order by slot desc limit 25;")

	if err != nil {
		logger.Printf("Error retrieving block tree data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("Blocks - beaconcha.in - Ethereum 2.0 beacon chain explorer - %v", time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/blocks",
		},
		ShowSyncingMessage: services.IsSyncing(),
		Active:             "blocks",
		Data:               blocksTreeData,
	}

	err = blocksTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Fatalf("Error executing template for %v route: %v", r.URL.String(), err)
	}
}

func BlocksData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Printf("Error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Printf("Error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Printf("Error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	if length > 100 {
		length = 100
	}

	var blocksCount uint64

	err = db.DB.Get(&blocksCount, "SELECT MAX(slot) + 1 FROM blocks")
	if err != nil {
		logger.Printf("Error retrieving max slot number: %v", err)
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

	var blocks []*types.IndexPageDataBlocks
	err = db.DB.Select(&blocks, `SELECT blocks.epoch, 
											    blocks.slot, 
											    blocks.proposer, 
											    blocks.blockroot, 
											    blocks.parentroot, 
											    blocks.attestationscount, 
											    blocks.depositscount, 
											    blocks.voluntaryexitscount, 
											    blocks.proposerslashingscount, 
											    blocks.attesterslashingscount, 
											    blocks.status 
										FROM blocks 
										WHERE blocks.slot >= $1 AND blocks.slot <= $2
										ORDER BY blocks.slot DESC`, endSlot, startSlot)

	if err != nil {
		logger.Printf("Error retrieving block data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	tableData := make([][]string, len(blocks))
	for i, b := range blocks {
		tableData[i] = []string{
			fmt.Sprintf("%v", b.Epoch),
			fmt.Sprintf("%v", b.Slot),
			fmt.Sprintf("%v", utils.FormatBlockStatus(b.Status)),
			fmt.Sprintf("%v", utils.SlotToTime(b.Slot).Unix()),
			fmt.Sprintf("%v", b.Proposer),
			fmt.Sprintf("%x", b.BlockRoot),
			fmt.Sprintf("%v", b.Attestations),
			fmt.Sprintf("%v", b.Deposits),
			fmt.Sprintf("%v / %v", b.Proposerslashings, b.Attesterslashings),
			fmt.Sprintf("%v", b.Exits),
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
		logger.Fatalf("Error enconding json response for %v route: %v", r.URL.String(), err)
	}
}
