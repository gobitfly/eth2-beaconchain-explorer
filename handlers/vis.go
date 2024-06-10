package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"
)

// Vis returns the visualizations using a go template
func Vis(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "vis.html")
	var visTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "stats", "/viz", "Visualizations", templateFiles)

	if handleTemplateError(w, r, "vis.go", "Vis", "", visTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
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

	// slot in postgres is limited to int
	if sinceSlot > math.MaxInt32 {
		logger.Warnf("error retrieving block tree data, slot too big: %v", err)
		http.Error(w, "Error: Invalid parameter since.", http.StatusBadRequest)
		return
	}

	var chartData []*types.VisChartData

	err = db.ReaderDb.Select(&chartData, "select slot, blockroot, parentroot, proposer from blocks where slot >= $1 and status in ('1', '2') order by slot desc limit 50;", sinceSlot)

	if err != nil {
		logger.Errorf("error retrieving block tree data: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// VisVotes shows the votes visualizations using a go template
func VisVotes(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "vis_votes.html")
	var visVotesTemplate = templates.GetTemplate(templateFiles...)

	var err error

	w.Header().Set("Content-Type", "text/html")

	since := time.Now().Add(time.Minute * -20).Unix()
	sinceSlot := utils.TimeToSlot(uint64(since - 120))

	var chartData []*types.VotesVisChartData

	rows, err := db.ReaderDb.Query(`select blocks.slot, 
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	for rows.Next() {
		data := &types.VotesVisChartData{}
		err := rows.Scan(&data.Slot, &data.BlockRoot, &data.ParentRoot, &data.Validators)
		if err != nil {
			logger.Errorf("error scanning votes tree data: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		chartData = append(chartData, data)
	}

	logger.Printf("returning %v entries since %v", len(chartData), sinceSlot)

	data := InitPageData(w, r, "vis", "/vis", "Votes", templateFiles)
	data.Data = &types.VisVotesPageData{ChartData: chartData}

	if handleTemplateError(w, r, "vis.go", "VisVotes", "", visVotesTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}
