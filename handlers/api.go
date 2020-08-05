package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/utils"
	"fmt"
	"net/http"
	"time"
)

func ApiHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	lastEpoch, err := db.GetLatestEpoch()

	if err != nil {
		http.Error(w, "Internal server error: could not retrieve latest epoch from the db", 503)
		return
	}

	epochTime := utils.EpochToTime(lastEpoch)
	if epochTime.Before(time.Now().Add(time.Minute * -13)) {
		http.Error(w, "Internal server error: last epoch in db is more than 13 minutes old", 503)
		return
	}

	fmt.Fprintf(w, "OK. Last epoch is from %v ago", time.Since(epochTime))
}
