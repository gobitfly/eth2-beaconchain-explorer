package ethclients

import (
	"encoding/json"
	"errors"
	"eth2-exporter/types"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New().WithField("module", "ethClients")

type gitAPIResponse struct {
	URL           string        `json:"url"`
	AssetsURL     string        `json:"assets_url"`
	UploadURL     string        `json:"upload_url"`
	HTMLURL       string        `json:"html_url"`
	ID            uint64        `json:"id"`
	Author        interface{}   `json:"author"`
	NodeID        string        `json:"node_id"`
	TagName       string        `json:"tag_name"`
	Target        string        `json:"target_commitish"`
	Name          string        `json:"name"`
	Draft         bool          `json:"draft"`
	PreRelease    bool          `json:"prerelease"`
	CreatedDate   string        `json:"created_at"`
	PublishedDate string        `json:"published_at"`
	Assets        []interface{} `json:"assets"`
	Tarball       string        `json:"tarball_url"`
	ZipBall       string        `json:"zipball_url"`
	Body          string        `json:"body"`
}

var ethClients = new(types.EthClientServicesPageData)
var ethClientsMux = &sync.RWMutex{}
var bannerClients = ""
var bannerClientsMux = &sync.RWMutex{}

// Init starts a go routine to update the ETH Clients Info
func Init() {
	go update()
}

func fetchClientData(repo string) *gitAPIResponse {
	gitAPI := new(gitAPIResponse)
	resp, err := http.Get("https://api.github.com/repos" + repo + "/releases/latest")

	if err != nil {
		logger.Errorf("error retrieving ETH Client Data: %v", err)
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&gitAPI)

	if err != nil {
		logger.Errorf("error decoding ETH Clients json response to struct: %v", err)
	}

	return gitAPI
}

func getRepoTime(date string) (time.Time, error) {
	var year, month, day int64
	var err error
	dateDays := strings.Split(date, "-")
	if len(dateDays) < 3 {
		return time.Now(), errors.New("Invalid date string " + date)
	}
	year, err = strconv.ParseInt(dateDays[0], 10, 0)
	if err != nil {
		logger.Errorf("error parsing year: %v", err)
	}
	month, err = strconv.ParseInt(dateDays[1], 10, 0)
	if err != nil {
		logger.Errorf("error parsing month: %v", err)
	}
	day, err = strconv.ParseInt(dateDays[2], 10, 0)
	if err != nil {
		logger.Errorf("error parsing day: %v", err)
	}
	return time.Date(int(year), time.Month(int(month)), int(day), 0, 0, 0, 0, time.UTC), nil
}

func prepareEthClientData(repo string, name string, curTime time.Time) (string, string) {

	client := fetchClientData(repo)
	date := strings.Split(client.PublishedDate, "T")

	if len(date) > 0 {
		rTime, err := getRepoTime(date[0])
		if err != nil {
			return client.Name, "N/A"
		}
		timeDiff := curTime.Sub(rTime).Hours() / 24
		if timeDiff < 2 {
			bannerClients += name + " " + client.Name + " | "
			return client.Name, "Recently"
		}

		return client.Name, fmt.Sprintf("%.0f days ago", timeDiff)
	}
	return client.Name, ""
}

func updateEthClient() {
	curTime := time.Now()

	if curTime.Sub(ethClients.LastUpdate) < time.Hour { // LastUpdate is initialized at January 1, year 1 so no need to check for nil
		return
	}

	logger.Println("Updating ETH Clients Information")
	ethClientsMux.Lock()
	defer ethClientsMux.Unlock()
	bannerClientsMux.Lock()
	defer bannerClientsMux.Unlock()
	bannerClients = ""
	ethClients.Geth.ClientReleaseVersion, ethClients.Geth.ClientReleaseDate = prepareEthClientData("/ethereum/go-ethereum", "Go-Ethereum", curTime)
	ethClients.Nethermind.ClientReleaseVersion, ethClients.Nethermind.ClientReleaseDate = prepareEthClientData("/NethermindEth/nethermind", "Nethermind", curTime)
	ethClients.OpenEthereum.ClientReleaseVersion, ethClients.OpenEthereum.ClientReleaseDate = prepareEthClientData("/openethereum/openethereum", "OpenEthereum", curTime)
	ethClients.Besu.ClientReleaseVersion, ethClients.Besu.ClientReleaseDate = prepareEthClientData("/hyperledger/besu", "Besu", curTime)

	ethClients.Teku.ClientReleaseVersion, ethClients.Teku.ClientReleaseDate = prepareEthClientData("/ConsenSys/teku", "Teku", curTime)
	ethClients.Prysm.ClientReleaseVersion, ethClients.Prysm.ClientReleaseDate = prepareEthClientData("/prysmaticlabs/prysm", "Prysm", curTime)
	ethClients.Nimbus.ClientReleaseVersion, ethClients.Nimbus.ClientReleaseDate = prepareEthClientData("/status-im/nimbus-eth2", "Nimbus-ETH2", curTime)
	ethClients.Lighthouse.ClientReleaseVersion, ethClients.Lighthouse.ClientReleaseDate = prepareEthClientData("/sigp/lighthouse", "Lighthouse", curTime)

	ethClients.LastUpdate = curTime
}

func update() {
	for true {
		updateEthClient()            // sending 8 requests to github per call
		time.Sleep(time.Minute * 20) // git api rate-limit 60 per hour : 60/8 = 7.5 minutes minimum
	}
}

// GetEthClientData returns a pointer of EthClientServicesPageData
func GetEthClientData() *types.EthClientServicesPageData {
	ethClientsMux.Lock()
	defer ethClientsMux.Unlock()
	return ethClients
}

// GetBannerClients returns a string of latest updates of ETH clients
func GetBannerClients() string {
	bannerClientsMux.Lock()
	defer bannerClientsMux.Unlock()
	return bannerClients
}
