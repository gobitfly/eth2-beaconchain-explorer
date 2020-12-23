package exporter

import (
	"eth2-exporter/db"
	"eth2-exporter/rpc"
	"eth2-exporter/utils"
	"time"
)

func networkLivenessUpdater(client rpc.Client) {
	var prevHeadEpoch uint64
	err := db.DB.Get(&prevHeadEpoch, "SELECT COALESCE(MAX(headepoch), 0) FROM network_liveness")
	if err != nil {
		logger.Fatal(err)
	}

	epochDuration := time.Second * time.Duration(utils.Config.Chain.SecondsPerSlot*utils.Config.Chain.SlotsPerEpoch)
	slotDuration := time.Second * time.Duration(utils.Config.Chain.SecondsPerSlot)

	for {
		head, err := client.GetChainHead()
		if err != nil {
			logger.Errorf("error getting chainhead when exporting networkliveness: %v", err)
			time.Sleep(slotDuration)
			continue
		}

		if prevHeadEpoch == head.HeadEpoch {
			time.Sleep(slotDuration)
			continue
		}

		// wait for node to be synced
		if time.Now().Add(-epochDuration).After(utils.EpochToTime(head.HeadEpoch)) {
			time.Sleep(slotDuration)
			continue
		}

		_, err = db.DB.Exec(`
			INSERT INTO network_liveness (ts, headepoch, finalizedepoch, justifiedepoch, previousjustifiedepoch)
			VALUES (NOW(), $1, $2, $3, $4)`,
			head.HeadEpoch, head.FinalizedEpoch, head.JustifiedEpoch, head.PreviousJustifiedEpoch)
		if err != nil {
			logger.Errorf("error saving networkliveness: %v", err)
		} else {
			logger.Printf("updated networkliveness for epoch %v", head.HeadEpoch)
			prevHeadEpoch = head.HeadEpoch
		}

		time.Sleep(slotDuration)
	}
}
