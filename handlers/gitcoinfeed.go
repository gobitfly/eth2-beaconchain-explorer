package handlers

import (
	"encoding/json"
	"eth2-exporter/services"
	"net/http"
)

func GitcoinFeed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type feedResp struct {
		Donors [][4]string `json:"donors"`
		IsLive bool        `json:"isLive"`
	}

	resp := feedResp{}

	feed := services.GetFeed()
	resp.IsLive = services.IsFeedOn()
	
	if len(feed) > 10 {
		resp.Donors = feed[:10]
	} else {
		resp.Donors = feed
	}

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
