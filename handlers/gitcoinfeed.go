package handlers

import (
	"encoding/json"
	"net/http"
)

type gitcoinfeed struct {
	Meta      interface{} `json:"meta"`
	Addresses [][4]string `json:"addresses"`
}

func GitcoinFeed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	feed := fetchFeedData()
	if feed == nil {
		return
	}

	err := json.NewEncoder(w).Encode(feed.Addresses)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

func fetchFeedData() *gitcoinfeed {
	var api = new(gitcoinfeed)
	// resp, err := http.Get("https://api.github.com/repos" + repo + "/releases/latest")
	resp, err := http.Get("http://localhost:5000/addrs")

	if err != nil {
		logger.Errorf("error retrieving gitcoin feed Data: %v", err)
		return nil
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&api)

	if err != nil {
		logger.Errorf("error decoding gitcoin feed json response to struct: %v", err)
		return nil
	}

	return api
}
