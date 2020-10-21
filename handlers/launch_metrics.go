package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
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

func LaunchMetricsData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var blks []sqlBlocks = []sqlBlocks{}

	err := db.DB.Select(&blks, `
		select
			b.slot,
			case
				when b.status = '0' then 'scheduled'
				when b.status = '1' then 'proposed'
				when b.status = '2' then 'missed'
				when b.status = '3' then 'orphaned'
				else 'unknown'
			end as status,
			b.epoch,
			e.globalparticipationrate,
			case when nl.finalizedepoch >= b.epoch then true else false end as finalized,
			case when nl.justifiedepoch = b.epoch then true else false end as justified,
			case when nl.previousjustifiedepoch = b.epoch then true else false end as previousjustified
		from blocks b
			left join epochs e on e.epoch = b.epoch
			left join network_liveness nl on headepoch = b.epoch
		where b.epoch < 5
		order by slot asc`)
	if err != nil {
		logger.Errorf("error querying blocks table for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}

	currentSlot := utils.TimeToSlot(uint64(time.Now().Unix()))
	currentEpoch := utils.EpochOfSlot(currentSlot)

	// var states = []string{"missed", "proposed", "orphaned"}

	// if len(blks) <= 0 {
	// 	blks = append(blks, sqlBlocks{
	// 		Epoch:                   0,
	// 		Slot:                    0,
	// 		Finalized:               false,
	// 		Justified:               false,
	// 		Globalparticipationrate: 0,
	// 		Previousjustified:       false,
	// 		Status:                  "proposed",
	// 	})
	// 	blks = append(blks, sqlBlocks{
	// 		Epoch:                   0,
	// 		Slot:                    1,
	// 		Finalized:               false,
	// 		Justified:               false,
	// 		Globalparticipationrate: 0,
	// 		Previousjustified:       false,
	// 		Status:                  "scheduled",
	// 	})
	// 	currentEpoch = 0
	// 	currentSlot = 1
	// 	go func() {
	// 		for {
	// 			time.Sleep(time.Second * 3)

	// 			blks[len(blks)-1] = sqlBlocks{
	// 				Epoch:                   currentEpoch,
	// 				Slot:                    currentSlot,
	// 				Finalized:               rand.Intn(2) != 0,
	// 				Justified:               rand.Intn(2) != 0,
	// 				Globalparticipationrate: float64(rand.Float64()),
	// 				Previousjustified:       rand.Intn(2) != 0,
	// 				Status:                  states[rand.Intn(3)],
	// 			}

	// 			if currentSlot == 31 {
	// 				currentSlot = 0
	// 				currentEpoch += 1

	// 			} else {
	// 				currentSlot += 1
	// 			}

	// 			blks = append(blks, sqlBlocks{
	// 				Epoch:                   currentEpoch,
	// 				Slot:                    currentSlot,
	// 				Finalized:               false,
	// 				Justified:               false,
	// 				Globalparticipationrate: 0,
	// 				Previousjustified:       false,
	// 				Status:                  "scheduled",
	// 			})
	// 		}
	// 	}()
	// }

	type blockType struct {
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
			epochMap[b.Epoch].Slots = append(epochMap[b.Epoch].Slots, blockType{b.Status, active})
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
				Slots:             []blockType{{status, active}},
			}
			epochMap[b.Epoch] = &r
			res.Epochs = append(res.Epochs, &r)
		}
	}

	// peersMu.RLock()
	// res.Peers = peers
	// peersMu.RUnlock()

	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

// type peer struct {
// 	Address         string `json:"address"`
// 	Direction       string `json:"direction"`
// 	ConnectionState string `json:"connectionState"`
// 	PeerID          string `json:"peerId"`
// 	ENR             string `json:"enr"`
// }

// var peers = []peer{}
// var peersMu = &sync.RWMutex{}

// func init() {
// 	updatePeers()
// }

// func UpdatePeers() {
// 	for {
// 		localPeers, err := getPeers()
// 		if err != nil {
// 			logger.WithError(err).Error("error updating peers for launch-metrics")
// 			time.Sleep(time.Second * 12)
// 			continue
// 		}
// 		peersMu.Lock()
// 		peers = localPeers
// 		peersMu.Unlock()
// 		time.Sleep(time.Second * 12)
// 	}
// }

// func getPeers() ([]peer, error) {
// 	h := "http://" + utils.Config.Indexer.Node.Host + ":3500/eth/v1alpha1/node/peers"
// 	client := &http.Client{Timeout: time.Second * 10}
// 	resp, err := client.Get(h)
// 	if err != nil {
// 		return nil, err
// 	}
// 	data, err := ioutil.ReadAll(resp.Body)
// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("error-response: %v: %s", resp.StatusCode, data)
// 	}
// 	var res []peer
// 	if err := json.Unmarshal(data, &res); err != nil {
// 		return nil, fmt.Errorf("error unmarshaling json: %w", err)
// 	}
// 	return res, nil
// }
