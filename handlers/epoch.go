package handlers

import (
	"database/sql"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// Epoch will show the epoch using a go template
func Epoch(w http.ResponseWriter, r *http.Request) {
	epochTemplateFiles := append(layoutTemplateFiles,
		"epoch.html",
		"components/timestamp.html")
	epochFutureTemplateFiles := append(layoutTemplateFiles,
		"epochFuture.html",
		"components/timestamp.html")
	epochNotFoundTemplateFiles := append(layoutTemplateFiles, "epochnotfound.html")
	var epochTemplate = templates.GetTemplate(epochTemplateFiles...)
	var epochFutureTemplate = templates.GetTemplate(epochFutureTemplateFiles...)
	var epochNotFoundTemplate = templates.GetTemplate(epochNotFoundTemplateFiles...)

	const MaxEpochValue = 4294967296 // we only render a page for epochs up to this value

	w.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(r)
	epochString := strings.Replace(vars["epoch"], "0x", "", -1)
	epochTitle := fmt.Sprintf("Epoch %v", epochString)

	epoch, err := strconv.ParseUint(epochString, 10, 64)
	metaPath := fmt.Sprintf("/epoch/%v", epoch)

	if err != nil {
		data := InitPageData(w, r, "blockchain", metaPath, epochTitle, append(layoutTemplateFiles, epochNotFoundTemplateFiles...))

		if handleTemplateError(w, r, "epoch.go", "Epoch", "parse epochString", epochNotFoundTemplate.ExecuteTemplate(w, "layout", data)) != nil {
			return // an error has occurred and was processed
		}
		return
	}

	epochPageData := types.EpochPageData{}
	latestFinalizedEpoch := services.LatestFinalizedEpoch()

	err = db.ReaderDb.Get(&epochPageData, `
		SELECT 
			epoch, 
			blockscount, 
			proposerslashingscount, 
			attesterslashingscount, 
			attestationscount, 
			depositscount, 
			voluntaryexitscount, 
			validatorscount, 
			averagevalidatorbalance, 
			(epoch <= $2) AS finalized,
			eligibleether,
			globalparticipationrate,
			votedether
		FROM epochs 
		WHERE epoch = $1`, epoch, latestFinalizedEpoch)
	if err != nil {
		//Epoch not in database -> Show future epoch
		if epoch > MaxEpochValue {
			data := InitPageData(w, r, "blockchain", metaPath, epochTitle, append(layoutTemplateFiles, epochNotFoundTemplateFiles...))
			if handleTemplateError(w, r, "epoch.go", "Epoch", ">MaxEpochValue", epochNotFoundTemplate.ExecuteTemplate(w, "layout", data)) != nil {
				return // an error has occurred and was processed
			}
			return
		}

		//Create placeholder structs
		blocks := make([]*types.IndexPageDataBlocks, utils.Config.Chain.Config.SlotsPerEpoch)
		for i := range blocks {
			slot := uint64(i) + (epoch * utils.Config.Chain.Config.SlotsPerEpoch)
			block := types.IndexPageDataBlocks{
				Epoch:  epoch,
				Slot:   slot,
				Ts:     utils.SlotToTime(slot),
				Status: 4,
			}
			blocks[31-i] = &block
		}
		epochPageData = types.EpochPageData{
			Epoch:         epoch,
			BlocksCount:   utils.Config.Chain.Config.SlotsPerEpoch,
			PreviousEpoch: epoch - 1,
			NextEpoch:     epoch + 1,
			Ts:            utils.EpochToTime(epoch),
			Blocks:        blocks,
		}

		//Render template
		data := InitPageData(w, r, "blockchain", metaPath, epochTitle, append(layoutTemplateFiles, epochFutureTemplateFiles...))
		data.Data = epochPageData
		if handleTemplateError(w, r, "epoch.go", "Epoch", "Done (not in Database)", epochFutureTemplate.ExecuteTemplate(w, "layout", data)) != nil {
			return // an error has occurred and was processed
		}
		return
	}

	err = db.ReaderDb.Select(&epochPageData.Blocks, `
		SELECT 
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
       		blocks.status,
			blocks.syncaggregate_participation,
			COALESCE(validator_names.name, '') AS name
		FROM blocks
		LEFT JOIN validators ON blocks.proposer = validators.validatorindex
		LEFT JOIN validator_names ON validators.pubkey = validator_names.publickey
		WHERE epoch = $1
		ORDER BY blocks.slot DESC`, epoch)
	if err != nil {
		logger.Errorf("error epoch blocks data: %v", err)
		data := InitPageData(w, r, "blockchain", metaPath, epochTitle, append(layoutTemplateFiles, epochNotFoundTemplateFiles...))

		if handleTemplateError(w, r, "epoch.go", "Epoch", "read Blocks from db", epochNotFoundTemplate.ExecuteTemplate(w, "layout", data)) != nil {
			return // an error has occurred and was processed
		}
		return
	}

	for _, block := range epochPageData.Blocks {
		block.Ts = utils.SlotToTime(block.Slot)
		block.ProposerFormatted = utils.FormatValidatorWithName(block.Proposer, block.ProposerName)

		switch block.Status {
		case 0:
			epochPageData.ScheduledCount += 1
		case 1:
			epochPageData.ProposedCount += 1
			epochPageData.SyncParticipationRate += block.SyncAggParticipation
			epochPageData.WithdrawalCount += block.Withdrawals
		case 2:
			epochPageData.MissedCount += 1
		case 3:
			epochPageData.OrphanedCount += 1
		}
	}

	withdrawalTotal, err := db.GetEpochWithdrawalsTotal(epoch)
	if err != nil {
		logger.Errorf("error getting epoch withdrawals total: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	epochPageData.WithdrawalTotal = utils.FormatCurrentBalance(withdrawalTotal, GetCurrency(r))

	epochPageData.SyncParticipationRate /= float64(epochPageData.ProposedCount)

	epochPageData.Ts = utils.EpochToTime(epochPageData.Epoch)

	err = db.ReaderDb.Get(&epochPageData.NextEpoch, "SELECT epoch FROM epochs WHERE epoch > $1 ORDER BY epoch LIMIT 1", epochPageData.Epoch)
	if err == sql.ErrNoRows {
		epochPageData.NextEpoch = 0
	} else if err != nil {
		logger.Errorf("error retrieving next epoch for epoch %v: %v", epochPageData.Epoch, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if epochPageData.Epoch == 0 {
		epochPageData.PreviousEpoch = 0
	} else {
		err = db.ReaderDb.Get(&epochPageData.PreviousEpoch, "SELECT epoch FROM epochs WHERE epoch < $1 ORDER BY epoch DESC LIMIT 1", epochPageData.Epoch)
		if err != nil {
			logger.Errorf("error retrieving previous epoch for epoch %v: %v", epochPageData.Epoch, err)
			epochPageData.PreviousEpoch = 0
		}
	}

	data := InitPageData(w, r, "blockchain", metaPath, epochTitle, append(layoutTemplateFiles, epochTemplateFiles...))
	data.Data = epochPageData

	if utils.IsApiRequest(r) {
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(data.Data)
	} else {
		err = epochTemplate.ExecuteTemplate(w, "layout", data)
	}

	if handleTemplateError(w, r, "epoch.go", "Epoch", "Done", err) != nil {
		return // an error has occurred and was processed
	}
}
