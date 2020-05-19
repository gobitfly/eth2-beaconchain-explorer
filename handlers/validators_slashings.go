package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"fmt"
	"github.com/juliangruber/go-intersect"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

var validatorsSlashingsTemplate = template.Must(template.New("validators").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/validators_slashings.html"))

// Validators returns the validators using a go template
func ValidatorsSlashings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - Validator Slashings - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/validators/slashings",
		},
		ShowSyncingMessage:    services.IsSyncing(),
		Active:                "validators",
		Data:                  nil,
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
	}

	err := validatorsSlashingsTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

// ValidatorAttestations returns a validators attestations in json
func ValidatorsSlashingsData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	var attesterSlashings []*types.ValidatorAttestationSlashing
	err = db.DB.Select(&attesterSlashings, `select 
       blocks.slot, 
       blocks.epoch, 
       blocks.proposer, 
       blocks_attesterslashings.attestation1_indices, 
       blocks_attesterslashings.attestation2_indices 
from blocks_attesterslashings 
    left join blocks on blocks_attesterslashings.block_slot = blocks.slot 
where attestation1_indices is not null and attestation2_indices is not null
order by blocks.slot desc;`)

	if err != nil {
		logger.Errorf("error retrieving validator attestations data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	tableData := make([][]interface{}, len(attesterSlashings))
	for i, b := range attesterSlashings {

		inter := intersect.Simple(b.Attestestation1Indices, b.Attestestation2Indices)

		slashedValidator := uint64(0)
		if len(inter) > 0 {
			slashedValidator = uint64(inter[0].(int64))
		}

		tableData[i] = []interface{}{
			utils.FormatSlashedValidator(slashedValidator),
			utils.FormatValidator(b.Proposer),
			utils.SlotToTime(b.Slot).Unix(),
			"Attestation rule violation",
			b.Slot,
			b.Epoch,
		}
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    uint64(len(attesterSlashings)),
		RecordsFiltered: uint64(len(attesterSlashings)),
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
