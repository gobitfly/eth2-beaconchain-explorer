package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"net/http"
	"strconv"
)

// Epochs will return the epochs using a go template
func Epochs(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles,
		"epochs.html",
		"components/timestamp.html")
	var epochsTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "blockchain", "/epochs", "Epochs", templateFiles)

	if handleTemplateError(w, r, "epochs.go", "Epochs", "Done", epochsTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

// EpochsData will return the epoch data using a go template
func EpochsData(w http.ResponseWriter, r *http.Request) {
	currency := GetCurrency(r)
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	search, err := strconv.ParseInt(q.Get("search[value]"), 10, 64)
	if err != nil {
		search = -1
	}

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
		endEpoch = 0
	}

	var epochs []*types.EpochsPageData

	latestFinalizedEpoch := services.LatestFinalizedEpoch()
	if search == -1 {
		err = db.ReaderDb.Select(&epochs, `
			SELECT epoch, 
				blockscount, 
				proposerslashingscount, 
				attesterslashingscount, 
				attestationscount, 
				depositscount, 
				withdrawalcount,
				voluntaryexitscount, 
				validatorscount, 
				averagevalidatorbalance, 
				(epoch <= $3) AS finalized,
				eligibleether,
				globalparticipationrate,
				votedether
			FROM epochs 
			WHERE epoch >= $1 AND epoch <= $2
			ORDER BY epoch DESC`, endEpoch, startEpoch, latestFinalizedEpoch)
	} else {
		err = db.ReaderDb.Select(&epochs, `
			SELECT epoch, 
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
			WHERE epoch = $1
			ORDER BY epoch DESC`, search, latestFinalizedEpoch)
	}
	if err != nil {
		logger.Errorf("error retrieving epoch data: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tableData := make([][]interface{}, len(epochs))
	for i, b := range epochs {
		// logger.Info("debug", b.Epoch, b.EligibleEther, b.VotedEther, b.GlobalParticipationRate, currency, utils.FormatBalance(b.EligibleEther, currency))
		tableData[i] = []interface{}{
			utils.FormatEpoch(b.Epoch),
			utils.FormatTimestamp(utils.EpochToTime(b.Epoch).Unix()),
			utils.FormatCount(b.AttestationsCount, b.Finalized, false),
			fmt.Sprintf("%v / %v", utils.FormatCount(b.DepositsCount, b.Finalized, true), utils.FormatCount(b.WithdrawalCount, b.Finalized, true)),
			fmt.Sprintf("%v / %v", utils.FormatCount(b.ProposerSlashingsCount, b.Finalized, true), utils.FormatCount(b.AttesterSlashingsCount, b.Finalized, true)),
			utils.FormatYesNo(b.Finalized),
			utils.FormatBalance(b.EligibleEther, currency),
			utils.FormatGlobalParticipationRate(b.VotedEther, b.GlobalParticipationRate, currency),
		}
	}

	filteredCount := epochsCount
	if search > -1 {
		filteredCount = uint64(len(epochs))
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    epochsCount,
		RecordsFiltered: filteredCount,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

}
