package handlers

import (
	"encoding/json"
	"eth2-exporter/services"
	"net/http"
)

func GitcoinFeed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	feed := services.GetFeed()
	var resp [][4]string = feed
	if len(feed) > 10 {
		resp = feed[:10]
	} else {
		resp = feed
	}

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
