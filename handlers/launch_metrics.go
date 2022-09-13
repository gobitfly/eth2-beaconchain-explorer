package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/utils"
	"net/http"
	"sort"
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
		COALESCE(e.finalized, false) as finalized
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

	type blockType struct {
		Epoch  uint64
		Slot   uint64
		Status string `json:"status"`
		Active bool   `json:"active"`
	}
	type epochType struct {
		Epoch          uint64         `json:"epoch"`
		Finalized      bool           `json:"finalized"`
		Particicpation float64        `json:"participation"`
		Slots          [32]*blockType `json:"slots"`
	}

	epochMap := map[uint64]*epochType{}

	res := struct {
		Epochs []*epochType
		// Peers  []peer
	}{}

	for _, b := range blks {
		if b.Globalparticipationrate == 1 && !b.Finalized {
			b.Globalparticipationrate = 0
		}
		_, exists := epochMap[b.Epoch]
		if !exists {
			r := epochType{
				Epoch:          b.Epoch,
				Finalized:      b.Finalized,
				Particicpation: b.Globalparticipationrate,
				Slots:          [32]*blockType{},
			}
			epochMap[b.Epoch] = &r
		}

		slotIndex := b.Slot - (b.Epoch * utils.Config.Chain.Config.SlotsPerEpoch)

		epochMap[b.Epoch].Slots[slotIndex] = &blockType{
			Epoch:  b.Epoch,
			Slot:   b.Slot,
			Status: b.Status,
			Active: b.Slot == currentSlot,
		}
	}

	for _, epoch := range epochMap {
		for i := 0; i < 32; i++ {
			if epoch.Slots[i] == nil {
				status := "scheduled"
				slot := epoch.Epoch*utils.Config.Chain.Config.SlotsPerEpoch + uint64(i)
				if slot < currentSlot-3 {
					status = "missed"
				}
				epoch.Slots[i] = &blockType{
					Epoch:  epoch.Epoch,
					Slot:   slot,
					Status: status,
					Active: slot == currentSlot,
				}
			}
		}
	}

	for _, epoch := range epochMap {
		for _, slot := range epoch.Slots {
			slot.Active = slot.Slot == currentSlot

			if slot.Status != "proposed" {
				if slot.Slot >= currentSlot {
					slot.Status = "scheduled"
				} else {
					slot.Status = "missed"
				}
			}
		}
		res.Epochs = append(res.Epochs, epoch)
	}

	sort.Slice(res.Epochs, func(i, j int) bool {
		return res.Epochs[i].Epoch > res.Epochs[j].Epoch
	})

	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}
