package handlers

import (
	"encoding/hex"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"strings"
	"time"
)

var blockTemplate = template.Must(template.New("block").ParseFiles("templates/layout.html", "templates/block.html"))
var blockNotFoundTemplate = template.Must(template.New("blocknotfound").ParseFiles("templates/layout.html", "templates/blocknotfound.html"))

func Block(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	vars := mux.Vars(r)
	slotOrHash := strings.Replace(vars["slotOrHash"], "0x", "", -1)

	blockPageData := types.BlockPageData{}

	blockRootHash, err := hex.DecodeString(slotOrHash)

	if err == nil && len(slotOrHash) == 64 {
		slotOrHash = "-1"
	}

	err = db.DB.Get(&blockPageData, `
	SELECT 
			epoch, 
			slot, 
			blockroot, 
			parentroot, 
			stateroot, 
			signature, 
			randaoreveal, 
			graffiti, 
			eth1data_depositroot, 
			eth1data_depositcount, 
			eth1data_blockhash, 
			proposerslashingscount, 
			attesterslashingscount,
			attestationscount, 
			depositscount, 
			voluntaryexitscount, 
			proposer,
			status   
	FROM blocks 
	WHERE slot = $1 OR blockroot = $2`,
		slotOrHash, blockRootHash)

	logger.Println(slotOrHash, blockRootHash)
	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("Slot %v - beaconcha.in - Ethereum 2.0 beacon chain explorer - %v", slotOrHash, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/block/" + slotOrHash,
		},
		ShowSyncingMessage: services.IsSyncing(),
		Active:             "blocks",
		Data:               nil,
	}

	if err != nil {
		logger.Printf("Error retrieving block data: %v", err)
		err = blockNotFoundTemplate.ExecuteTemplate(w, "layout", data)

		if err != nil {
			logger.Fatalf("Error executing template for %v route: %v", r.URL.String(), err)
		}
		return
	}

	blockPageData.Ts = utils.SlotToTime(blockPageData.Slot)
	blockPageData.NextSlot = blockPageData.Slot + 1
	blockPageData.PreviousSlot = blockPageData.Slot - 1
	blockPageData.SlashingsCount = blockPageData.AttesterSlashingsCount + blockPageData.ProposerSlashingsCount

	slots := types.BlockPageMinMaxSlot{}
	err = db.DB.Get(&slots, "SELECT MAX(slot) AS maxslot, MIN(slot) as minslot FROM blocks")
	if err != nil {
		logger.Printf("Error retrieving block data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	if blockPageData.NextSlot > slots.MaxSlot {
		blockPageData.NextSlot = 0
	}

	if blockPageData.PreviousSlot < slots.MinSlot {
		blockPageData.PreviousSlot = 0
	}

	var attestations []*types.BlockPageAttestation
	err = db.DB.Select(&attestations, `SELECT aggregationbits, 
												     signature, 
												     slot, 
												     index, 
												     beaconblockroot, 
												     source_epoch, 
												     source_root, 
												     target_epoch, 
												     target_root 
												FROM blocks_attestations 
												WHERE block_slot = $1 
												ORDER BY slot, index`,
		blockPageData.Slot)
	if err != nil {
		logger.Printf("Error retrieving block attestation data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	blockPageData.Attestations = attestations

	var deposits []*types.BlockPageDeposit
	err = db.DB.Select(&deposits, `SELECT publickey, 
												     withdrawalcredentials, 
												     amount, 
												     signature
												FROM blocks_deposits 
												WHERE block_slot = $1 
												ORDER BY block_index`,
		blockPageData.Slot)
	if err != nil {
		logger.Printf("Error retrieving block deposit data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	for _, d := range deposits {
		d.AmountFormatted = utils.FormatBalance(d.Amount)
	}
	blockPageData.Deposits = deposits

	data.Data = blockPageData

	err = blockTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Fatalf("Error executing template for %v route: %v", r.URL.String(), err)
	}
}
