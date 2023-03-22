package handlers

import (
	"context"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Will return the slots page
func Slots(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "slots.html")
	var blocksTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")
	q := r.URL.Query()

	data := InitPageData(w, r, "blockchain", "/slots", "Slots", templateFiles)

	user, session, err := getUserSession(r)
	if err != nil {
		logger.WithError(err).Error("error getting user session")
	}

	state := GetDataTableState(user, session, "slots")

	length := uint64(50)
	start := uint64(0)
	search := ""
	searchForEmpty := false

	if state.Length == 0 {
		length = 50
	}
	if state.Search.Search != "" {
		search = state.Search.Search
	}

	if q.Get("search[value]") != "" {
		search = q.Get("search[value]")
	}

	if q.Get("q") != "" {
		search = q.Get("q")
	}

	search = strings.Replace(search, "0x", "", -1)

	tableData, err := GetSlotsTableData(0, start, length, search, searchForEmpty)
	if err != nil {
		logger.Errorf("error rendering blocks table data: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	data.Data = tableData
	if handleTemplateError(w, r, "blocks.go", "Blocks", "", blocksTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

// BlocksData will return information about blocks
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
		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	tableData, err := GetSlotsTableData(draw, start, length, search, searchForEmpty)
	if err != nil {
		logger.Errorf("error rendering blocks table data: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	err = json.NewEncoder(w).Encode(tableData)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func GetSlotsTableData(draw, start, length uint64, search string, searchForEmpty bool) (*types.DataTableResponse, error) {
	var totalCount uint64
	var filteredCount uint64
	var blocks []*types.BlocksPageDataBlocks

	err := db.ReaderDb.Get(&totalCount, "SELECT COALESCE(MAX(slot),0) FROM blocks")
	if err != nil {
		return nil, err
	}

	if length > 100 {
		length = 100
	}

	if search == "" && !searchForEmpty {
		filteredCount = totalCount
		startSlot := totalCount - start
		endSlot := totalCount - start - length + 1

		if startSlot > 9223372036854775807 {
			startSlot = totalCount
		}
		if endSlot > 9223372036854775807 {
			endSlot = 0
		}
		err = db.ReaderDb.Select(&blocks, `
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

			args = append(args, "%"+search+"%")
			searchBlocksQrys = append(searchBlocksQrys, fmt.Sprintf(`(select slot from blocks where graffiti_text ilike $%d order by slot desc limit %d)`, len(args), searchLimit))
			searchProposersQrys = append(searchProposersQrys, fmt.Sprintf(`(select publickey as pubkey from validator_names where name ilike $%d limit %d)`, len(args), searchLimit))

			searchNumber, err := strconv.ParseUint(search, 10, 64)
			if err == nil {
				// if the search-string is a number we can look for exact matchings
				args = append(args, searchNumber)
				searchBlocksQrys = append(searchBlocksQrys, fmt.Sprintf(`(select slot from blocks where slot = $%d order by slot desc limit %d)`, len(args), searchLimit))
				searchProposersQrys = append(searchProposersQrys, fmt.Sprintf(`(select pubkey from validators where validatorindex = $%d limit %d)`, len(args), searchLimit))
			}
			if searchPubkeyExactRE.MatchString(search) {
				// if the search-string is a valid hex-string but not long enough for a full publickey we look for prefix
				pubkey := strings.ToLower(strings.Replace(search, "0x", "", -1))
				args = append(args, pubkey)
				searchProposersQrys = append(searchProposersQrys, fmt.Sprintf(`(select pubkey from validators where pubkeyhex = $%d limit %d)`, len(args), searchLimit))
			} else if len(search) > 2 && searchPubkeyLikeRE.MatchString(search) {
				// if the search-string looks like a publickey we look for exact match
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

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err = db.ReaderDb.SelectContext(ctx, &blocks, qry, args...)
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
		if b.Slot == 0 {
			tableData[i] = []interface{}{
				utils.FormatEpoch(b.Epoch),
				utils.FormatBlockSlot(b.Slot),
				template.HTML("<span class=\"badge text-dark\" style=\"background: rgba(179, 159, 70, 0.8) none repeat scroll 0% 0%;\">Genesis</span>"),
				utils.FormatTimestamp(utils.SlotToTime(b.Slot).Unix()),
				template.HTML("N/A"),
				b.Attestations,
				template.HTML(fmt.Sprintf("%v / %v", b.Deposits, b.Withdrawals)),
				fmt.Sprintf("%v / %v", b.Proposerslashings, b.Attesterslashings),
				b.Exits,
				b.Votes,
				fmt.Sprintf("%.2f", b.SyncAggParticipation*100.0),
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
				template.HTML(fmt.Sprintf("%v / %v", b.Deposits, b.Withdrawals)),
				fmt.Sprintf("%v / %v", b.Proposerslashings, b.Attesterslashings),
				b.Exits,
				b.Votes,
				fmt.Sprintf("%.2f", b.SyncAggParticipation*100.0),
				utils.FormatGraffitiAsLink(b.Graffiti),
			}
		}
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    totalCount,
		RecordsFiltered: filteredCount,
		Data:            tableData,
		DisplayStart:    start,
		PageLength:      length,
	}
	return data, nil
}
