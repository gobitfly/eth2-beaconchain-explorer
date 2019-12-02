package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

var indexTemplate = template.Must(template.New("index").Funcs(template.FuncMap{"formatBlockStatus": utils.FormatBlockStatus}).ParseFiles("templates/layout.html", "templates/index.html"))

func Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	indexPageData := types.IndexPageData{}

	var blocks []*types.IndexPageDataBlocks

	err := db.DB.Select(&blocks, `SELECT blocks.epoch, 
											    blocks.slot, 
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
										ORDER BY blocks.slot DESC LIMIT 20`)

	if err != nil {
		logger.Printf("Error retrieving index block data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	indexPageData.Blocks = blocks

	if len(blocks) > 0 {
		indexPageData.CurrentSlot = blocks[0].Slot
	}

	for _, block := range indexPageData.Blocks {
		block.Ts = utils.SlotToTime(block.Slot)
	}

	err = db.DB.Get(&indexPageData.EnteringValidators, "SELECT COUNT(*) FROM validatorqueue_activation")
	if err != nil {
		logger.Printf("Error retrieving entering validator count: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	err = db.DB.Get(&indexPageData.ExitingValidators, "SELECT COUNT(*) FROM validatorqueue_exit")
	if err != nil {
		logger.Printf("Error retrieving exiting validator count: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	var averageBalance float64
	err = db.DB.Get(&averageBalance, "SELECT COALESCE(AVG(balance), 0) FROM validator_balances WHERE epoch = $1", services.LatestEpoch())
	if err != nil {
		logger.Printf("Error retrieving validator balance: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	indexPageData.AverageBalance = utils.FormatBalance(uint64(averageBalance))

	var epochHistory []*types.IndexPageEpochHistory
	err = db.DB.Select(&epochHistory, "SELECT epoch, eligibleether, validatorscount, finalized FROM epochs WHERE epoch < $1 ORDER BY epoch", services.LatestEpoch())
	if err != nil {
		logger.Printf("Error retrieving staked ether history: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	if len(epochHistory) > 0 {
		indexPageData.CurrentEpoch = epochHistory[len(epochHistory)-1].Epoch

		for i := len(epochHistory) - 1; i >= 0; i-- {
			if epochHistory[i].Finalized {
				indexPageData.CurrentFinalizedEpoch = epochHistory[i].Epoch
				indexPageData.FinalityDelay = indexPageData.CurrentEpoch - indexPageData.CurrentFinalizedEpoch
				break
			}
		}

		indexPageData.StakedEther = utils.FormatBalance(epochHistory[len(epochHistory)-1].EligibleEther)
		indexPageData.ActiveValidators = epochHistory[len(epochHistory)-1].ValidatorsCount
	}

	indexPageData.StakedEtherChartData = make([][]float64, len(epochHistory))
	indexPageData.ActiveValidatorsChartData = make([][]float64, len(epochHistory))
	for i, history := range epochHistory {
		indexPageData.StakedEtherChartData[i] = []float64{float64(utils.EpochToTime(history.Epoch).Unix() * 1000), float64(history.EligibleEther) / 1000000000}
		indexPageData.ActiveValidatorsChartData[i] = []float64{float64(utils.EpochToTime(history.Epoch).Unix() * 1000), float64(history.ValidatorsCount)}
	}

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("Index - beaconcha.in - Ethereum 2.0 beacon chain explorer - %v", time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "",
		},
		ShowSyncingMessage: services.IsSyncing(),
		Active:             "index",
		Data:               indexPageData,
	}

	err = indexTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Fatalf("Error executing template for %v route: %v", r.URL.String(), err)
	}
}
