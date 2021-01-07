package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"html/template"
	"net/http"
	"strconv"

	"github.com/juliangruber/go-intersect"
)

var validatorsSlashingsTemplate = template.Must(template.New("validators").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/validators_slashings.html"))

// ValidatorsSlashings returns validator slashing using a go template
func ValidatorsSlashings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "validators", "/validators/slashings", "Validator Slashings")
	data.HeaderAd = true

	err := validatorsSlashingsTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

// ValidatorsSlashingsData returns validator slashings in json
func ValidatorsSlashingsData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

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

	var slashings []*types.ValidatorSlashing
	err = db.DB.Select(&slashings, `
		SELECT 
			slot,
			epoch,
			proposer,
			slashedvalidator,
			attestation1_indices,
			attestation2_indices,
			type
		FROM (
			SELECT
				blocks.slot, 
				blocks.epoch, 
				blocks.proposer,
				NULL as slashedvalidator,
				blocks_attesterslashings.attestation1_indices, 
				blocks_attesterslashings.attestation2_indices,
				'Attestation Violation'::varchar as type
			FROM blocks_attesterslashings 
			INNER JOIN blocks on blocks_attesterslashings.block_slot = blocks.slot AND blocks.status = '1'
			UNION ALL
			SELECT
				blocks.slot, 
				blocks.epoch, 
				blocks.proposer, 
				blocks_proposerslashings.proposerindex as slashedvalidator,
				NULL as attestation1_indices,
				NULL as attestation2_indices,
				'Proposer Violation' as type 
			FROM blocks_proposerslashings
			INNER JOIN blocks on blocks_proposerslashings.block_slot = blocks.slot AND blocks.status = '1'
		) as query
		ORDER BY slot desc
		LIMIT $1
		OFFSET $2`, length, start)

	tableData := make([][]interface{}, 0, len(slashings))

	validatorNames, err := db.GetValidatorNames()

	if err != nil {
		logger.Errorf("error retrieving validator names from the database: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	for _, row := range slashings {
		entry := []interface{}{}

		slashedValidators := []uint64{}

		if row.Type == "Attestation Violation" {
			inter := intersect.Simple(row.Attestestation1Indices, row.Attestestation2Indices)
			if len(inter) == 0 {
				logger.Warningf("No intersection found for attestation violation, proposer: %v, slot: %v", row.Proposer, row.Slot)
			}
			for _, v := range inter {
				slashedValidators = append(slashedValidators, uint64(v.(int64)))
			}
			entry = append(entry, utils.FormatSlashedValidatorsWithName(slashedValidators, validatorNames))
		}

		if row.Type == "Proposer Violation" {
			entry = append(entry, utils.FormatSlashedValidatorWithName(*row.SlashedValidator, validatorNames[*row.SlashedValidator]))
		}

		entry = append(entry, utils.FormatValidatorWithName(row.Proposer, validatorNames[row.Proposer]))
		entry = append(entry, utils.FormatTimestamp(utils.SlotToTime(row.Slot).Unix()))
		entry = append(entry, row.Type)
		entry = append(entry, utils.FormatBlockSlot(row.Slot))
		entry = append(entry, utils.FormatEpoch(row.Epoch))

		tableData = append(tableData, entry)
	}
	records, err := db.GetSlashingCount()
	if err != nil {
		logger.Errorf("GetSlashingCount failed to retrieve record count: %v", err)
		http.Error(w, "Internal server error", 503)
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    records,
		RecordsFiltered: records,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
