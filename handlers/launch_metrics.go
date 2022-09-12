package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/utils"
	"net/http"
	"time"
)

type sqlBlocks struct {
	Slot                    uint64
	Epoch                   uint64
	Status                  string
	Globalparticipationrate float64
	Finalized               bool
	Justified               bool
	Previousjustified       bool
}

// var currentEpoch uint64
// var currentSlot uint64

// LaunchMetricsData returns the metrics for the earliest epochs
func LaunchMetricsData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var blks []sqlBlocks = []sqlBlocks{}
	lookBack := services.LatestEpoch()
	if lookBack < 4 {
		lookBack = 0
	} else {
		lookBack = lookBack - 4
	}
	// latestEpoch := services.LatestEpoch()
	// lowEpoch := latestEpoch - 5
	// if latestEpoch < 5 {
	// 	lowEpoch = 0

	// } else {
	// 	lowEpoch = latestEpoch - 5
	// }

	// highEpoch := latestEpoch

	err := db.ReaderDb.Select(&blks, `
	SELECT
		b.slot,
		case
			when b.status = '0' then 'scheduled'
			when b.status = '1' then 'proposed'
			when b.status = '2' then 'missed'
			when b.status = '3' then 'orphaned'
			else 'unknown'
		end as status,
		b.epoch,
		COALESCE(e.globalparticipationrate, 0) as globalparticipationrate,
		e.finalized
	FROM blocks b
		left join epochs e on e.epoch = b.epoch
	WHERE b.epoch >= $1
	ORDER BY slot desc;
`, services.LatestEpoch()-4)
	if err != nil {
		logger.Errorf("error querying blocks table for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	currentSlot := utils.TimeToSlot(uint64(time.Now().Unix()))
	currentEpoch := utils.EpochOfSlot(currentSlot)

	type blockType struct {
		Epoch  uint64
		Slot   uint64
		Status string `json:"status"`
		Active bool   `json:"active"`
	}
	type epochType struct {
		Epoch             uint64      `json:"epoch"`
		Finalized         bool        `json:"finalized"`
		Justified         bool        `json:"justified"`
		PreviousJustified bool        `json:"previousjustified"`
		Particicpation    float64     `json:"participation"`
		Slots             []blockType `json:"slots"`
	}

	epochMap := map[uint64]*epochType{}

	res := struct {
		Epochs []*epochType
		// Peers  []peer
	}{}

	for _, b := range blks {
		active := false
		if b.Epoch == currentEpoch && b.Slot == currentSlot {
			active = true

			// set previous active current slots to false
			for _, epoch := range epochMap {
				for _, slot := range epoch.Slots {
					slot.Active = false
				}
			}
		}
		_, exists := epochMap[b.Epoch]
		if exists {
			epochMap[b.Epoch].Slots = append(epochMap[b.Epoch].Slots, blockType{b.Epoch, b.Slot, b.Status, active})
			if b.Globalparticipationrate > epochMap[b.Epoch].Particicpation {
				epochMap[b.Epoch].Particicpation = b.Globalparticipationrate
			}
			if b.Finalized {
				epochMap[b.Epoch].Finalized = b.Finalized
			}
			if b.Justified {
				epochMap[b.Epoch].Justified = b.Justified
			}
			if b.Previousjustified {
				epochMap[b.Epoch].PreviousJustified = b.Previousjustified
			}
		} else {
			status := b.Status
			if b.Epoch == 0 {
				status = "genesis"
			}
			r := epochType{
				Epoch:             b.Epoch,
				Finalized:         b.Finalized,
				Justified:         b.Justified,
				PreviousJustified: b.Previousjustified,
				Particicpation:    b.Globalparticipationrate,
				Slots:             []blockType{{b.Epoch, b.Slot, status, active}},
			}
			epochMap[b.Epoch] = &r
			res.Epochs = append(res.Epochs, &r)
		}
	}

	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}
