package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

var epochsTemplate = template.Must(template.New("epochs").ParseFiles("templates/layout.html", "templates/epochs.html"))

func Epochs(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html")

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("Epochs - beaconcha.in - Ethereum 2.0 beacon chain explorer - %v", time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/epochs",
		},
		Active: "epochs",
		Data:   nil,
	}

	err := epochsTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Fatalf("Error executing template for %v route: %v", r.URL.String(), err)
	}
}

func EpochsData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Printf("Error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	if err != nil {
		logger.Printf("Error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Printf("Error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
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

	if err != nil {
		logger.Printf("Error retrieving epoch data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	tableData := make([][]string, len(epochs))
	for i, b := range epochs {
		tableData[i] = []string{
			fmt.Sprintf("%v", b.Epoch),
			fmt.Sprintf("%v", utils.EpochToTime(b.Epoch).Unix()),
			fmt.Sprintf("%v", b.BlocksCount),
			fmt.Sprintf("%v", b.AttestationsCount),
			fmt.Sprintf("%v", b.DepositsCount),
			fmt.Sprintf("%v / %v", b.ProposerSlashingsCount, b.AttesterSlashingsCount),
			fmt.Sprintf("%v", b.Finalized),
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
		logger.Fatalf("Error enconding json response for %v route: %v", r.URL.String(), err)
	}

}
