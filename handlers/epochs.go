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
	templateFiles := append(layoutTemplateFiles, "epochs.html")
	var epochsTemplate = templates.GetTemplate(templateFiles...)

	currency := GetCurrency(r)

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "blockchain", "/epochs", "Epochs", templateFiles)

	var epochs []*types.EpochsPageData

	epochsCount := services.LatestEpoch()

	user, session, err := getUserSession(r)
	if err != nil {
		logger.WithError(err).Error("error getting user session")
	}

	state := GetDataTableState(user, session, "epochs")
	length := uint64(50)
	start := uint64(0)
	var startEpoch uint64
	var endEpoch uint64

	// set start and end epoch from saved state
	if state != nil && state.Length != 0 {
		length = state.Length
		start = state.Start
		startEpoch = epochsCount - state.Start
		endEpoch = epochsCount - state.Start - state.Length + 1
	} else {
		startEpoch = epochsCount
		endEpoch = epochsCount - 50 + 1
	}

	if length < 10 {
		length = 10
	}

	if length > 100 {
		length = 100
	}

	if startEpoch > 9223372036854775807 {
		startEpoch = epochsCount
	}
	if endEpoch > 9223372036854775807 {
		endEpoch = 0
	}

	// logger.Infof("state: %+v", state)
	// logger.Infof("start: %v end: %v", startEpoch, endEpoch)

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
		finalized,
		eligibleether,
		globalparticipationrate,
		votedether
	FROM epochs 
	WHERE epoch >= $1 AND epoch <= $2
	ORDER BY epoch DESC`, endEpoch, startEpoch)
	if err != nil {
		logger.Errorf("error retrieving epoch data: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	tableData := make([][]interface{}, len(epochs))
	for i, b := range epochs {
		// logger.Info("debug", b.Epoch, b.EligibleEther, b.VotedEther, b.GlobalParticipationRate, currency, utils.FormatBalance(b.EligibleEther, currency))
		tableData[i] = []interface{}{
			utils.FormatEpoch(b.Epoch),
			utils.FormatTimestamp(utils.EpochToTime(b.Epoch).Unix()),
			b.AttestationsCount,
			fmt.Sprintf("%v / %v", b.DepositsCount, b.WithdrawalCount),
			fmt.Sprintf("%v / %v", b.ProposerSlashingsCount, b.AttesterSlashingsCount),
			utils.FormatYesNo(b.Finalized),
			utils.FormatBalance(b.EligibleEther, currency),
			utils.FormatGlobalParticipationRate(b.VotedEther, b.GlobalParticipationRate, currency),
		}
	}

	data.Data = &types.DataTableResponse{
		Draw:            0,
		RecordsTotal:    epochsCount,
		RecordsFiltered: epochsCount,
		Data:            tableData,
		PageLength:      length,
		DisplayStart:    start,
	}

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
				finalized,
				eligibleether,
				globalparticipationrate,
				votedether
			FROM epochs 
			WHERE epoch >= $1 AND epoch <= $2
			ORDER BY epoch DESC`, endEpoch, startEpoch)
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
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	tableData := make([][]interface{}, len(epochs))
	for i, b := range epochs {
		// logger.Info("debug", b.Epoch, b.EligibleEther, b.VotedEther, b.GlobalParticipationRate, currency, utils.FormatBalance(b.EligibleEther, currency))
		tableData[i] = []interface{}{
			utils.FormatEpoch(b.Epoch),
			utils.FormatTimestamp(utils.EpochToTime(b.Epoch).Unix()),
			b.AttestationsCount,
			fmt.Sprintf("%v / %v", b.DepositsCount, b.WithdrawalCount),
			fmt.Sprintf("%v / %v", b.ProposerSlashingsCount, b.AttesterSlashingsCount),
			utils.FormatYesNo(b.Finalized),
			utils.FormatBalance(b.EligibleEther, currency),
			utils.FormatGlobalParticipationRate(b.VotedEther, b.GlobalParticipationRate, currency),
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
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

}
