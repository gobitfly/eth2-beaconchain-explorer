package handlers

import (
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

	"github.com/gorilla/mux"
)

var epochTemplate = template.Must(template.New("epoch").Funcs(template.FuncMap{"formatBlockStatus": utils.FormatBlockStatus}).ParseFiles("templates/layout.html", "templates/epoch.html"))
var epochNotFoundTemplate = template.Must(template.New("epochnotfound").ParseFiles("templates/layout.html", "templates/epochnotfound.html"))

// Epoch will show the epoch using a go template
func Epoch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	vars := mux.Vars(r)
	epochString := strings.Replace(vars["epoch"], "0x", "", -1)

	data := &types.PageData{
		Meta: &types.Meta{
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
		},
		ShowSyncingMessage: services.IsSyncing(),
		Active:             "epochs",
		Data:               nil,
		Version:            version.Version,
	}

	epoch, err := strconv.ParseUint(epochString, 10, 64)

	if err != nil {
		data.Meta.Title = fmt.Sprintf("%v - Epoch %v - beaconcha.in - %v", utils.Config.Frontend.SiteName, epochString, time.Now().Year())
		data.Meta.Path = "/epoch/" + epochString
		logger.Printf("Error retrieving block data: %v", err)
		err = epochNotFoundTemplate.ExecuteTemplate(w, "layout", data)

		if err != nil {
			logger.Fatalf("Error executing template for %v route: %v", r.URL.String(), err)
		}
		return
	}

	data.Meta.Title = fmt.Sprintf("%v - Epoch %v - beaconcha.in - %v", utils.Config.Frontend.SiteName, epoch, time.Now().Year())
	data.Meta.Path = fmt.Sprintf("/epoch/%v", epoch)

	epochPageData := types.EpochPageData{}

	err = db.DB.Get(&epochPageData, `SELECT epoch, 
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
		logger.Printf("Error getting epoch data: %v", err)
		err = epochNotFoundTemplate.ExecuteTemplate(w, "layout", data)

		if err != nil {
			logger.Fatalf("Error executing template for %v route: %v", r.URL.String(), err)
		}
		return
	}

	err = db.DB.Select(&epochPageData.Blocks, `SELECT blocks.slot, 
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
										WHERE epoch = $1
										ORDER BY blocks.slot DESC`, epoch)

	if err != nil {
		logger.Printf("Error epoch blocks data: %v", err)
		err = epochNotFoundTemplate.ExecuteTemplate(w, "layout", data)

		if err != nil {
			logger.Fatalf("Error executing template for %v route: %v", r.URL.String(), err)
		}
		return
	}

	for _, block := range epochPageData.Blocks {
		block.Ts = utils.SlotToTime(block.Slot)
	}

	epochPageData.Ts = utils.EpochToTime(epochPageData.Epoch)
	epochPageData.NextEpoch = epochPageData.Epoch + 1
	epochPageData.PreviousEpoch = epochPageData.Epoch - 1

	epochPageData.VotedEtherFormatted = fmt.Sprintf("%.2f ETH", float64(epochPageData.VotedEther)/float64(1000000000))
	epochPageData.EligibleEtherFormatted = fmt.Sprintf("%.2f ETH", float64(epochPageData.EligibleEther)/float64(1000000000))
	epochPageData.GlobalParticipationRateFormatted = fmt.Sprintf("%.0f", epochPageData.GlobalParticipationRate*float64(100))

	epochs := types.EpochPageMinMaxSlot{}
	err = db.DB.Get(&epochs, "SELECT MAX(epoch) AS maxepoch, MIN(epoch) as minepoch FROM epochs")
	if err != nil {
		logger.Printf("Error retrieving block data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	if epochPageData.NextEpoch > epochs.MaxEpoch {
		epochPageData.NextEpoch = 0
	}

	if epochPageData.PreviousEpoch < epochs.MinEpoch {
		epochPageData.PreviousEpoch = 0
	}

	data.Data = epochPageData

	err = epochTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Fatalf("Error executing template for %v route: %v", r.URL.String(), err)
	}
}
