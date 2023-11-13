package ethclients

import (
	"encoding/json"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New().WithField("module", "ethClients")

type ethernodesAPIStruct struct {
	Client string `json:"client"`
	Value  int    `json:"value"`
}
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

type clientUpdateInfo struct {
	Name string
	Date time.Time
}

type EthClients struct {
	ClientReleaseVersion string
	ClientReleaseDate    template.HTML
	NetworkShare         string
	IsUserSubscribed     bool
}

type EthClientServicesPageData struct {
	LastUpdate          time.Time
	Geth                EthClients
	Nethermind          EthClients
	Besu                EthClients
	Teku                EthClients
	Prysm               EthClients
	Nimbus              EthClients
	Lighthouse          EthClients
	Erigon              EthClients
	RocketpoolSmartnode EthClients
	MevBoost            EthClients
	Lodestar            EthClients
	Banner              string
	CsrfField           template.HTML
}

var ethClients = EthClientServicesPageData{}
var ethClientsMux = &sync.RWMutex{}
var bannerClients = []clientUpdateInfo{}
var bannerClientsMux = &sync.RWMutex{}

var httpClient = &http.Client{Timeout: time.Second * 10}

// Init starts a go routine to update the ETH Clients Info
func Init() {
	go update()
}

func fetchClientData(repo string) *gitAPIResponse {
	var gitAPI = new(gitAPIResponse)
	resp, err := httpClient.Get("https://api.github.com/repos" + repo + "/releases/latest")
	// resp, err := http.Get("http://localhost:5000/repos" + repo)

	if err != nil {
		logger.Errorf("error retrieving ETH Client Data: %v", err)
		return nil
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Errorf("error retrieving ETH Client Data, status code: %v", resp.StatusCode)
		return nil
	}

	err = json.NewDecoder(resp.Body).Decode(&gitAPI)

	if err != nil {
		logger.Errorf("error decoding ETH Clients json response to struct: %v", err)
		return nil
	}

	return gitAPI
}

var ethernodesAPI []ethernodesAPIStruct

func fetchClientNetworkShare() []ethernodesAPIStruct {
	resp, err := http.Get("https://ethernodes.org/api/clients")

	if err != nil {
		logger.Errorf("error retrieving ETH Clients Network Share Data: %v", err)
		return ethernodesAPI
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&ethernodesAPI)

	if err != nil {
		logger.Errorf("error decoding ETH Clients Network Share json response to struct: %v", err)
	}

	return ethernodesAPI
}

func getRepoTime(date string, dTime string) (time.Time, error) {
	var year, month, day, hour, min int64
	var err error
	dateDays := strings.Split(date, "-")
	dateTimes := strings.Split(dTime, ":")
	if len(dateDays) < 3 || len(dateTimes) < 3 {
		return time.Now(), fmt.Errorf("invalid date string %s %s", date, dTime)
	}
	year, err = strconv.ParseInt(dateDays[0], 10, 0)
	if err != nil {
		return time.Now(), fmt.Errorf("error parsing year: %v", err)
	}
	month, err = strconv.ParseInt(dateDays[1], 10, 0)
	if err != nil {
		return time.Now(), fmt.Errorf("error parsing month: %v", err)
	}
	day, err = strconv.ParseInt(dateDays[2], 10, 0)
	if err != nil {
		return time.Now(), fmt.Errorf("error parsing day: %v", err)
	}
	hour, err = strconv.ParseInt(dateTimes[0], 10, 0)
	if err != nil {
		return time.Now(), fmt.Errorf("error parsing hour: %v", err)
	}
	min, err = strconv.ParseInt(dateTimes[1], 10, 0)
	if err != nil {
		return time.Now(), fmt.Errorf("error parsing min: %v", err)
	}
	return time.Date(int(year), time.Month(int(month)), int(day), int(hour), int(min), 0, 0, time.UTC), nil
}

func prepareEthClientData(repo string, name string, curTime time.Time) (string, template.HTML) {
	client := fetchClientData(repo)
	time.Sleep(time.Millisecond * 250) // consider github rate limit

	if client == nil {
		return "Github", "searching"
	}
	date := strings.Split(client.PublishedDate, "T")

	if len(date) == 2 {
		rTime, err := getRepoTime(date[0], date[1])
		if err != nil {
			logger.Errorf("error parsing git repo. time: %v", err)
			return client.Name, "GitHub" // client.Name is client version from github api
		}
		timeDiff := (curTime.Sub(rTime).Hours() / 24.0)

		if timeDiff < 1 { // add recent releases for notification collector to be collected
			update := clientUpdateInfo{Name: name, Date: rTime}
			bannerClients = append(bannerClients, update)
		}
		return client.Name, utils.FormatTimestamp(rTime.Unix())
	}
	return "Github", "searching" // If API limit is exceeded
}

func updateEthClientNetShare() {
	nShare := fetchClientNetworkShare()
	total := 0
	for _, item := range nShare {
		total += item.Value
	}

	for _, item := range nShare {
		share := fmt.Sprintf("%.1f%%", (float64(item.Value)/float64(total))*100.0)
		switch item.Client {
		case "geth":
			ethClients.Geth.NetworkShare = share
		case "nethermind":
			ethClients.Nethermind.NetworkShare = share
		case "besu":
			ethClients.Besu.NetworkShare = share
		case "erigon":
			ethClients.Erigon.NetworkShare = share
		default:
			continue
		}
	}
}

func updateEthClient() {
	curTime := time.Now()
	// sending 8 requests to github per call
	// git api rate-limit 60 per hour : 60/8 = 7.5 minutes minimum
	if curTime.Sub(ethClients.LastUpdate) < time.Hour { // LastUpdate is initialized at January 1, year 1 so no need to check for nil
		return
	}

	logger.Println("Updating ETH Clients Information")
	ethClientsMux.Lock()
	defer ethClientsMux.Unlock()
	bannerClientsMux.Lock()
	defer bannerClientsMux.Unlock()
	bannerClients = []clientUpdateInfo{}
	updateEthClientNetShare()
	ethClients.Geth.ClientReleaseVersion, ethClients.Geth.ClientReleaseDate = prepareEthClientData("/ethereum/go-ethereum", "Geth", curTime)
	ethClients.Nethermind.ClientReleaseVersion, ethClients.Nethermind.ClientReleaseDate = prepareEthClientData("/NethermindEth/nethermind", "Nethermind", curTime)
	ethClients.Besu.ClientReleaseVersion, ethClients.Besu.ClientReleaseDate = prepareEthClientData("/hyperledger/besu", "Besu", curTime)
	ethClients.Erigon.ClientReleaseVersion, ethClients.Erigon.ClientReleaseDate = prepareEthClientData("/ledgerwatch/erigon", "Erigon", curTime)

	ethClients.Teku.ClientReleaseVersion, ethClients.Teku.ClientReleaseDate = prepareEthClientData("/ConsenSys/teku", "Teku", curTime)
	ethClients.Prysm.ClientReleaseVersion, ethClients.Prysm.ClientReleaseDate = prepareEthClientData("/prysmaticlabs/prysm", "Prysm", curTime)
	ethClients.Nimbus.ClientReleaseVersion, ethClients.Nimbus.ClientReleaseDate = prepareEthClientData("/status-im/nimbus-eth2", "Nimbus", curTime)
	ethClients.Lighthouse.ClientReleaseVersion, ethClients.Lighthouse.ClientReleaseDate = prepareEthClientData("/sigp/lighthouse", "Lighthouse", curTime)
	ethClients.Lodestar.ClientReleaseVersion, ethClients.Lodestar.ClientReleaseDate = prepareEthClientData("/chainsafe/lodestar", "Lodestar", curTime)

	ethClients.RocketpoolSmartnode.ClientReleaseVersion, ethClients.RocketpoolSmartnode.ClientReleaseDate = prepareEthClientData("/rocket-pool/smartnode-install", "Rocketpool", curTime)
	ethClients.MevBoost.ClientReleaseVersion, ethClients.MevBoost.ClientReleaseDate = prepareEthClientData("/flashbots/mev-boost", "MEV-Boost", curTime)

	ethClients.LastUpdate = curTime
}

func update() {
	for {
		updateEthClient()
		time.Sleep(time.Minute * 5)
	}
}

// GetEthClientData returns a EthClientServicesPageData
func GetEthClientData() EthClientServicesPageData {
	ethClientsMux.Lock()
	defer ethClientsMux.Unlock()
	return ethClients
}

// GetUpdatedClients returns a slice of latest updated clients or empty slice if no updates
func GetUpdatedClients() []clientUpdateInfo {
	bannerClientsMux.Lock()
	defer bannerClientsMux.Unlock()
	return bannerClients
	// return []string{"Prysm", "Teku"}
}
