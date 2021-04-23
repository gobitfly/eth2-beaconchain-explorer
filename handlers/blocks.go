package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

var blocksTemplate = template.Must(template.New("blocks").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/blocks.html"))

// Blocks will return information about blocks using a go template
func Blocks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "blocks", "/blocks", "Blocks")

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
		err = db.DB.Get(&blocksCount, "SELECT COALESCE(MAX(slot) + 1,0) FROM blocks")
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
				blocks.graffiti,
				COALESCE(validator_names.name, '') AS name
			FROM blocks 
			LEFT JOIN validators ON blocks.proposer = validators.validatorindex
			LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
			WHERE blocks.slot >= $1 AND blocks.slot <= $2 
			ORDER BY blocks.slot DESC`, endSlot, startSlot)
	} else {
		err = db.DB.Get(&blocksCount, `
			SELECT count(*) 
			FROM blocks 
			WHERE 
				CAST(blocks.slot as text) LIKE $1 
				OR LOWER(ENCODE(graffiti , 'escape')) LIKE LOWER($2)
				OR ENCODE(graffiti, 'hex') LIKE ($3)
				OR proposer IN (
					SELECT validatorindex
					FROM validators
					INNER JOIN validator_names ON validators.pubkey = validator_names.publickey
					WHERE validator_names.name IS NOT NULL AND LOWER(validator_names.name) LIKE LOWER($2)
				)`, search+"%", "%"+search+"%", fmt.Sprintf("%%%x%%", search))
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
				blocks.graffiti,
				COALESCE(validator_names.name, '') AS name
			FROM blocks 
			LEFT JOIN validators ON blocks.proposer = validators.validatorindex
			LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
			WHERE slot IN (
				SELECT slot 
				FROM blocks
				WHERE 
					CAST(blocks.slot as text) LIKE $1 
					OR LOWER(ENCODE(graffiti , 'escape')) LIKE LOWER($2) 
					OR ENCODE(graffiti, 'hex') LIKE ($3)
					OR proposer IN (
						SELECT validatorindex
						FROM validators
						INNER JOIN validator_names ON validators.pubkey = validator_names.publickey
						WHERE validator_names.name IS NOT NULL AND LOWER(validator_names.name) LIKE LOWER($2)
					)
				ORDER BY blocks.slot DESC 
				LIMIT $4
				OFFSET $5
			) ORDER BY blocks.slot DESC`,
			search+"%", "%"+search+"%", fmt.Sprintf("%%%x%%", search), length, start)
	}

	if err != nil {
		logger.Errorf("error retrieving block data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	tableData := make([][]interface{}, len(blocks))
	for i, b := range blocks {
		if b.Slot == 0 {
			tableData[i] = []interface{}{
				utils.FormatEpoch(b.Epoch),
				utils.FormatBlockSlot(b.Slot),
				template.HTML("<span class=\"badge text-dark\" style=\"background: rgba(179, 159, 70, 0.8) none repeat scroll 0% 0%;\">Genesis</span>"),
				utils.FormatTimestamp(utils.SlotToTime(b.Slot).Unix()),
				template.HTML("N/A"),
				b.Attestations,
				b.Deposits,
				fmt.Sprintf("%v / %v", b.Proposerslashings, b.Attesterslashings),
				b.Exits,
				b.Votes,
				utils.FormatGraffitiAsLink(b.Graffiti),
			}
		} else {
			tableData[i] = []interface{}{
				utils.FormatEpoch(b.Epoch),
				utils.FormatBlockSlot(b.Slot),
				utils.FormatBlockStatus(b.Status),
				utils.FormatTimestamp(utils.SlotToTime(b.Slot).Unix()),
				utils.FormatValidatorWithName(b.Proposer, b.ProposerName),
				b.Attestations,
				b.Deposits,
				fmt.Sprintf("%v / %v", b.Proposerslashings, b.Attesterslashings),
				b.Exits,
				b.Votes,
				utils.FormatGraffitiAsLink(b.Graffiti),
			}
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
