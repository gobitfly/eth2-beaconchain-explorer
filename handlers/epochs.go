package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

var epochsTemplate = template.Must(template.New("epochs").ParseFiles("templates/layout.html", "templates/epochs.html"))

// Epochs will return the epochs using a go template
func Epochs(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html")

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - Epochs - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/epochs",
		},
		ShowSyncingMessage:    services.IsSyncing(),
		Active:                "epochs",
		Data:                  nil,
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
	}

	err := epochsTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

// EpochsData will return the epoch data using a go template
func EpochsData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	search, err := strconv.ParseInt(q.Get("search[value]"), 10, 64)
	if err != nil {
		search = -1
	}

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

	epochsCount := services.LatestEpoch()

	startEpoch := epochsCount - start
	endEpoch := epochsCount - start - length + 1

	if startEpoch > 9223372036854775807 {
		startEpoch = epochsCount
	}
	if endEpoch > 9223372036854775807 {
		endEpoch = epochsCount
	}

	var epochs []*types.EpochsPageData

	if search == -1 {
		err = db.DB.Select(&epochs, `SELECT epoch, 
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
										WHERE epoch >= $1 AND epoch <= $2
										ORDER BY epoch DESC`, endEpoch, startEpoch)
	} else {
		err = db.DB.Select(&epochs, `SELECT epoch, 
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
										WHERE epoch = $1
										ORDER BY epoch DESC`, search)
	}
	if err != nil {
		logger.Errorf("error retrieving epoch data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	tableData := make([][]interface{}, len(epochs))
	for i, b := range epochs {
		tableData[i] = []interface{}{
			b.Epoch,
			utils.EpochToTime(b.Epoch).Unix(),
			b.BlocksCount,
			b.AttestationsCount,
			b.DepositsCount,
			fmt.Sprintf("%v / %v", b.ProposerSlashingsCount, b.AttesterSlashingsCount),
			b.Finalized,
			utils.FormatBalance(b.EligibleEther),
			utils.FormatBalance(b.VotedEther),
			fmt.Sprintf("%.0f%%", b.GlobalParticipationRate*100),
		}
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    epochsCount,
		RecordsFiltered: epochsCount,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}

}
