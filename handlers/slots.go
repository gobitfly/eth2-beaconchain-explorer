package handlers

import (
	"context"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

// Will return the slots page
func Slots(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "slots.html")
	var blocksTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "blockchain", "/slots", "Slots", templateFiles)

	if handleTemplateError(w, r, "blocks.go", "Blocks", "", blocksTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

// SlotsData will return information about slots
func SlotsData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	search := q.Get("search[value]")

	if q.Get("q") != "" {
		search = q.Get("q")
	}

	search = strings.Replace(search, "0x", "", -1)

	searchForEmpty := (len(search) == 0 && q.Get("columns[11][search][value]") == "#showonlyemptygraffiti")

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables draw parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter draw", http.StatusBadRequest)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter start", http.StatusBadRequest)
		return
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Warnf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Error: Missing or invalid parameter length", http.StatusBadRequest)
		return
	}

	tableData, err := GetSlotsTableData(draw, start, length, search, searchForEmpty)
	if err != nil {
		logger.Errorf("error rendering blocks table data: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(tableData)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func GetSlotsTableData(draw, start, length uint64, search string, searchForEmpty bool) (*types.DataTableResponse, error) {
	var filteredCount uint64
	var blocks []*types.BlocksPageDataBlocks

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	latestSlot := services.LatestSlot()

	if length > 100 {
		length = 100
	}

	if search == "" && !searchForEmpty {
		filteredCount = latestSlot
		startSlot := latestSlot - start
		endSlot := latestSlot - start - length + 1

		if startSlot > 9223372036854775807 {
			startSlot = latestSlot
		}
		if endSlot > 9223372036854775807 {
			endSlot = 0
		}
		err := db.ReaderDb.Select(&blocks, `
			SELECT 
				blocks.epoch, 
				blocks.slot, 
				blocks.proposer, 
				blocks.blockroot, 
				blocks.parentroot, 
				blocks.attestationscount, 
				blocks.depositscount,
				blocks.withdrawalcount, 
				blocks.voluntaryexitscount, 
				blocks.proposerslashingscount, 
				blocks.attesterslashingscount, 
				blocks.syncaggregate_participation, 
				blocks.status, 
				COALESCE((SELECT SUM(ARRAY_LENGTH(validators, 1)) FROM blocks_attestations WHERE beaconblockroot = blocks.blockroot), 0) AS votes,
				blocks.graffiti,
				COALESCE(validator_names.name, '') AS name
			FROM blocks 
			LEFT JOIN validators ON blocks.proposer = validators.validatorindex
			LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
			WHERE blocks.slot >= $1 AND blocks.slot <= $2
			ORDER BY blocks.slot DESC`, endSlot, startSlot)
		if err != nil {
			return nil, fmt.Errorf("error retrieving block data: %w", err)
		}
	} else {
		// we search for blocks matching the search-string:
		//
		// - block-slot (exact when number)
		// - block-graffiti (infix)
		// - proposer-index (exact when number)
		// - proposer-publickey (prefix when hex, exact when hex and 96 chars)
		// - proposer-name (infix)
		//
		// the resulting query will look like this:
		//
		// 		$blocksQry1
		// 		union $blocksQry2
		// 		union $blocksQryN
		// 		union select slot from blocks where proposer in (
		// 			$proposersQry1
		// 			union $proposersQry2
		// 			union $proposersQryN
		// 		)
		//
		// note: we use union instead of disjunct or-queries for performance-reasons
		args := []interface{}{}

		searchLimit := 1000

		searchBlocksQry := ""
		if searchForEmpty {
			searchBlocksQry = fmt.Sprintf(`(select slot from blocks where blocks.graffiti_text is NULL OR blocks.graffiti_text = '' order by slot desc limit %d)`, searchLimit)
		} else {
			searchBlocksQrys := []string{}
			searchProposersQrys := []string{}

			ilikeSearch := "%" + search + "%"
			var graffiti [][]byte
			statsQry := `SELECT DISTINCT(graffiti) FROM graffiti_stats WHERE graffiti_text ILIKE $1`
			err := db.ReaderDb.SelectContext(ctx, &graffiti, statsQry, ilikeSearch)
			if err != nil {
				return nil, fmt.Errorf("error retrieving graffiti stats data (with search): %w", err)
			}

			args = append(args, pq.ByteaArray(graffiti))
			searchBlocksQrys = append(searchBlocksQrys, fmt.Sprintf(`(select slot from blocks where graffiti = ANY($%d) order by slot desc limit %d)`, len(args), searchLimit))

			args = append(args, ilikeSearch)
			searchProposersQrys = append(searchProposersQrys, fmt.Sprintf(`(select publickey as pubkey from validator_names where name ilike $%d limit %d)`, len(args), searchLimit))

			searchNumber, err := strconv.ParseUint(search, 10, 64)
			if err == nil {
				// if the search-string is a number we can look for exact matchings
				args = append(args, searchNumber)
				searchBlocksQrys = append(searchBlocksQrys, fmt.Sprintf(`(select slot from blocks where slot = $%d order by slot desc limit %d)`, len(args), searchLimit))
				searchProposersQrys = append(searchProposersQrys, fmt.Sprintf(`(select pubkey from validators where validatorindex = $%d limit %d)`, len(args), searchLimit))
			}
			if searchPubkeyExactRE.MatchString(search) {
				// if the search-string looks like a publickey we look for exact match
				pubkey := strings.ToLower(strings.Replace(search, "0x", "", -1))
				args = append(args, pubkey)
				searchProposersQrys = append(searchProposersQrys, fmt.Sprintf(`(select pubkey from validators where pubkeyhex = $%d limit %d)`, len(args), searchLimit))
			} else if len(search) > 2 && searchPubkeyLikeRE.MatchString(search) {
				// if the search-string is a valid hex-string but not long enough for a full publickey we look for prefix
				pubkey := strings.ToLower(strings.Replace(search, "0x", "", -1))
				args = append(args, pubkey+"%")
				searchProposersQrys = append(searchProposersQrys, fmt.Sprintf(`(select pubkey from validators where pubkeyhex like $%d limit %d)`, len(args), searchLimit))
			}

			// join proposer-queries and append to block-queries looking for proposers
			searchBlocksQrys = append(searchBlocksQrys, fmt.Sprintf(`select slot from blocks where proposer in (select v.validatorindex from (%v) a left join validators v on v.pubkey = a.pubkey) order by slot desc`, strings.Join(searchProposersQrys, " union ")))
			searchBlocksQry = strings.Join(searchBlocksQrys, " union ")
		}

		args = append(args, length)
		args = append(args, start)

		qry := fmt.Sprintf(`
			WITH matched_slots as (%v)
			SELECT 
				blocks.epoch, 
				blocks.slot, 
				blocks.proposer, 
				blocks.blockroot, 
				blocks.parentroot, 
				blocks.attestationscount, 
				blocks.depositscount,
				blocks.withdrawalcount,
				blocks.voluntaryexitscount, 
				blocks.proposerslashingscount, 
				blocks.attesterslashingscount, 
				blocks.syncaggregate_participation, 
				blocks.status, 
				COALESCE((SELECT SUM(ARRAY_LENGTH(validators, 1)) FROM blocks_attestations WHERE beaconblockroot = blocks.blockroot), 0) AS votes, 
				blocks.graffiti,
				COALESCE(validator_names.name, '') AS name,
				cnt.total_count
			FROM matched_slots
			INNER JOIN blocks on blocks.slot = matched_slots.slot 
			LEFT JOIN validators ON blocks.proposer = validators.validatorindex
			LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
			LEFT JOIN (select count(*) from matched_slots) cnt(total_count) ON true
			ORDER BY slot DESC LIMIT $%v OFFSET $%v`, searchBlocksQry, len(args)-1, len(args))

		err := db.ReaderDb.SelectContext(ctx, &blocks, qry, args...)
		if err != nil {
			return nil, fmt.Errorf("error retrieving block data (with search): %w", err)
		}

		filteredCount = 0
		if len(blocks) > 0 {
			filteredCount = blocks[0].TotalCount
		}
	}

	tableData := make([][]interface{}, len(blocks))
	for i, b := range blocks {
		validatorName := template.HTML("N/A")
		if b.Slot > 0 {
			validatorName = utils.FormatValidatorWithName(b.Proposer, b.ProposerName)
		}
		tableData[i] = []interface{}{
			utils.FormatEpoch(b.Epoch),
			utils.FormatBlockSlot(b.Slot),
			utils.FormatBlockStatus(b.Status, b.Slot),
			utils.FormatTimestamp(utils.SlotToTime(b.Slot).Unix()),
			validatorName,
			b.Attestations,
			template.HTML(fmt.Sprintf("%v / %v", b.Deposits, b.Withdrawals)),
			fmt.Sprintf("%v / %v", b.Proposerslashings, b.Attesterslashings),
			b.Exits,
			b.Votes,
			fmt.Sprintf("%.2f", b.SyncAggParticipation*100.0),
			utils.FormatGraffitiAsLink(b.Graffiti),
		}
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    latestSlot,
		RecordsFiltered: filteredCount,
		Data:            tableData,
		DisplayStart:    start,
		PageLength:      length,
	}
	return data, nil
}
