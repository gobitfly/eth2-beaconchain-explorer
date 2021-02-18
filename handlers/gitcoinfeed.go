package handlers

import (
	"encoding/json"
	"eth2-exporter/services"
	"net/http"
)

func GitcoinFeed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	feed := services.GetFeed()

	err := json.NewEncoder(w).Encode(feed)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
