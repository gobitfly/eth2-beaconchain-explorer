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

var visTemplate = template.Must(template.New("vis").ParseFiles("templates/layout.html", "templates/vis.html"))

func Vis(w http.ResponseWriter, r *http.Request) {

	var err error

	w.Header().Set("Content-Type", "text/html")

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fmt.Sprintf("Blocks - beaconcha.in - Ethereum 2.0 beacon chain explorer - %v", time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/blocks",
		},
		ShowSyncingMessage: services.IsSyncing(),
		Active:             "vis",
		Data:               nil,
	}

	err = visTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Fatalf("Error executing template for %v route: %v", r.URL.String(), err)
	}
}

func VisBlocks(w http.ResponseWriter, r *http.Request) {
	var err error

	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	since, err := strconv.ParseInt(q.Get("since"), 10, 64)
	if err != nil {
		since = time.Now().Add(time.Minute * -20).Unix()
	}

	sinceSlot := utils.TimeToSlot(uint64(since))

	var chartData []*types.VisChartData

	err = db.DB.Select(&chartData, "select slot, blockroot, parentroot from blocks where status = '1' and slot >= $1 order by slot desc limit 50;", sinceSlot)

	if err != nil {
		logger.Printf("Error retrieving block tree data: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	for _, d := range chartData {
		d.Number = d.Slot
		d.Timestamp = uint64(utils.SlotToTime(d.Slot).Unix())
		d.Hash = fmt.Sprintf("%x", d.BlockRoot)
		d.Parents = []string{fmt.Sprintf("%x", d.ParentRoot)}
		d.Difficulty = d.Slot
	}

	logger.Printf("Returning %v blocks since %v", len(chartData), sinceSlot)

	err = json.NewEncoder(w).Encode(chartData)
	if err != nil {
		logger.Fatalf("Error enconding json response for %v route: %v", r.URL.String(), err)
	}
}
