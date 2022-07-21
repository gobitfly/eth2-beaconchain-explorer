package handlers

import (
	"database/sql"
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var epochTemplate = template.Must(template.New("epoch").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/epoch.html"))
var epochFutureTemplate = template.Must(template.New("epochFuture").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/epochFuture.html"))
var epochNotFoundTemplate = template.Must(template.New("epochnotfound").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/epochnotfound.html"))

// Epoch will show the epoch using a go template
func Epoch(w http.ResponseWriter, r *http.Request) {
	const MaxEpochValue = 4294967296 // we only render a page for epochs up to this value

	w.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(r)
	epochString := strings.Replace(vars["epoch"], "0x", "", -1)

	data := InitPageData(w, r, "epochs", "/epochs", "Epoch")
	data.HeaderAd = true

	epoch, err := strconv.ParseUint(epochString, 10, 64)

	if err != nil {
		data.Meta.Title = fmt.Sprintf("%v - Epoch %v - beaconcha.in - %v", utils.Config.Frontend.SiteName, epochString, time.Now().Year())
		data.Meta.Path = "/epoch/" + epochString
		logger.Errorf("error parsing epoch index %v: %v", epochString, err)
		err = epochNotFoundTemplate.ExecuteTemplate(w, "layout", data)

		if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	data.Meta.Title = fmt.Sprintf("%v - Epoch %v - beaconcha.in - %v", utils.Config.Frontend.SiteName, epoch, time.Now().Year())
	data.Meta.Path = fmt.Sprintf("/epoch/%v", epoch)

	epochPageData := types.EpochPageData{}

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
			finalized,
			eligibleether,
			globalparticipationrate,
			votedether
		FROM epochs 
		WHERE epoch = $1`, epoch)
	if err != nil {
		//Epoch not in database -> Show future epoch
		if epoch > MaxEpochValue {
			err = epochNotFoundTemplate.ExecuteTemplate(w, "layout", data)

			if err != nil {
				logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			return
		}

		//Create placeholder structs
		blocks := make([]*types.IndexPageDataBlocks, 32)
		for i := range blocks {
			slot := uint64(i) + epoch*32
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
			BlocksCount:   32,
			PreviousEpoch: epoch - 1,
			NextEpoch:     epoch + 1,
			Ts:            utils.EpochToTime(epoch),
			Blocks:        blocks,
		}

		//Render template
		data.Data = epochPageData
		err = epochFutureTemplate.ExecuteTemplate(w, "layout", data)
		if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", http.StatusServiceUnavailable)
			return
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
			blocks.voluntaryexitscount, 
			blocks.proposerslashingscount, 
			blocks.attesterslashingscount,
       		blocks.status,
			blocks.syncaggregate_participation
		FROM blocks 
		WHERE epoch = $1
		ORDER BY blocks.slot DESC`, epoch)
	if err != nil {
		logger.Errorf("error epoch blocks data: %v", err)
		err = epochNotFoundTemplate.ExecuteTemplate(w, "layout", data)

		if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	for _, block := range epochPageData.Blocks {
		block.Ts = utils.SlotToTime(block.Slot)

		switch block.Status {
		case 0:
			epochPageData.ScheduledCount += 1
		case 1:
			epochPageData.ProposedCount += 1
			epochPageData.SyncParticipationRate += block.SyncAggParticipation
		case 2:
			epochPageData.MissedCount += 1
		case 3:
			epochPageData.OrphanedCount += 1
		}
	}
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
	err = db.ReaderDb.Get(&epochPageData.PreviousEpoch, "SELECT epoch FROM epochs WHERE epoch < $1 ORDER BY epoch DESC LIMIT 1", epochPageData.Epoch)
	if err != nil {
		logger.Errorf("error retrieving previous epoch for epoch %v: %v", epochPageData.Epoch, err)
		epochPageData.PreviousEpoch = 0
	}

	data.Data = epochPageData

	if utils.IsApiRequest(r) {
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(data.Data)
	} else {
		err = epochTemplate.ExecuteTemplate(w, "layout", data)
	}

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
