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

var visTemplate = template.Must(template.New("vis").ParseFiles("templates/layout.html", "templates/vis.html"))
var visVotesTemplate = template.Must(template.New("vis").ParseFiles("templates/layout.html", "templates/vis_votes.html"))

// Vis returns the visualizations using a go template
func Vis(w http.ResponseWriter, r *http.Request) {

	var err error

	w.Header().Set("Content-Type", "text/html")

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - Visualizations - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/blocks",
		},
		ShowSyncingMessage:    services.IsSyncing(),
		Active:                "vis",
		Data:                  nil,
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
	}

	err = visTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

// VisBlocks returns the visualizations in json
func VisBlocks(w http.ResponseWriter, r *http.Request) {
	var err error

	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	since, err := strconv.ParseInt(q.Get("since"), 10, 64)
	if err != nil {
		since = time.Now().Add(time.Minute * -20).Unix()
	}

	sinceSlot := utils.TimeToSlot(uint64(since - 120))

	var chartData []*types.VisChartData

	err = db.DB.Select(&chartData, "select slot, blockroot, parentroot, proposer from blocks where slot >= $1 and status in ('1', '2') order by slot desc limit 50;", sinceSlot)

	if err != nil {
		logger.Errorf("error retrieving block tree data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	lastMissedHash := ""
	for _, d := range chartData {
		d.Number = d.Slot
		d.Timestamp = uint64(utils.SlotToTime(d.Slot).Unix())
		d.Hash = fmt.Sprintf("0x%x", d.BlockRoot)
		d.Parents = []string{fmt.Sprintf("0x%x", d.ParentRoot)}
		if len(d.BlockRoot) == 1 {
			d.Hash += fmt.Sprintf("%v", d.Slot)
			d.Parents = []string{lastMissedHash}
			lastMissedHash = d.Hash
		}
		d.Difficulty = d.Slot
	}

	logger.Printf("returning %v blocks since %v", len(chartData), sinceSlot)

	err = json.NewEncoder(w).Encode(chartData)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

// VisVotes shows the votes visualizations using a go template
func VisVotes(w http.ResponseWriter, r *http.Request) {
	var err error

	w.Header().Set("Content-Type", "text/html")

	since := time.Now().Add(time.Minute * -20).Unix()
	sinceSlot := utils.TimeToSlot(uint64(since - 120))

	var chartData []*types.VotesVisChartData

	rows, err := db.DB.Query(`select blocks.slot, 
       											ENCODE(blocks.blockroot::bytea, 'hex') AS blockroot, 
       											ENCODE(blocks.parentroot::bytea, 'hex') AS parentroot,
												blocks_attestations.validators 
												from blocks 
													left join blocks_attestations on 
														blocks_attestations.beaconblockroot = blocks.blockroot 
												where blocks.slot >= $1 and blocks.status in ('1', '3') 
												order by blocks.slot desc LIMIT 10;`, sinceSlot)

	if err != nil {
		logger.Errorf("error retrieving votes tree data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	for rows.Next() {
		data := &types.VotesVisChartData{}
		err := rows.Scan(&data.Slot, &data.BlockRoot, &data.ParentRoot, &data.Validators)
		if err != nil {
			logger.Errorf("error scanning votes tree data: %v", err)
			http.Error(w, "Internal server error", 503)
			return
		}
		chartData = append(chartData, data)
	}

	logger.Printf("returning %v entries since %v", len(chartData), sinceSlot)

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("Blocks - beaconcha.in - Ethereum 2.0 beacon chain explorer - %v", time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/blocks",
		},
		ShowSyncingMessage:    services.IsSyncing(),
		Active:                "vis",
		Data:                  &types.VisVotesPageData{ChartData: chartData},
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
	}

	err = visVotesTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
