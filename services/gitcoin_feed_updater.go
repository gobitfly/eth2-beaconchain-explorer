package services

import (
	"encoding/json"
	"net/http"
	"time"

	"sync"
)

type gitcoinfeed struct {
	Meta      interface{} `json:"meta"`
	Addresses [][4]string `json:"addresses"`
}

var feed *gitcoinfeed
var feedMux = &sync.RWMutex{}

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

func updateFeed() {
	feedMux.Lock()
	defer feedMux.Unlock()
	tempFeed := fetchFeedData()
	if tempFeed == nil { // don't delete the existing users
		logger.Infoln("Gitcoin feed: empty respons")
		return
	}
	feed = tempFeed
}

func InitGitCoinFeed() {
	logger.Infoln("Started GitcoinFeed service")
	go func() {
		for true {
			updateFeed()
			time.Sleep(time.Second * 10)
		}
	}()
}

func GetFeed() [][4]string {
	feedMux.Lock()
	defer feedMux.Unlock()

	if feed == nil {
		return [][4]string{}
	}

	return feed.Addresses
}
