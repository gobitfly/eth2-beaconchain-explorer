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

	resp.Donors = services.GetFeed()
	resp.IsLive = services.IsFeedOn()

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
