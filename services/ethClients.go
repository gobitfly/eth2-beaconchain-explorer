package services

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// var logger = logrus.New().WithField("module", "ethClients")
type EthClients struct {
	ClientReleaseVersion string
	ClientReleaseDate    uint64
	NetworkShare         string
	IsUserSubscribed     bool
}

type EthClientServicesPageData struct {
	LastUpdate   time.Time
	Geth         EthClients
	Nethermind   EthClients
	OpenEthereum EthClients
	Besu         EthClients
	Teku         EthClients
	Prysm        EthClients
	Nimbus       EthClients
	Lighthouse   EthClients
	Banner       string
	CsrfField    template.HTML
}

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

var ethClients = EthClientServicesPageData{}
var ethClientsMux = &sync.RWMutex{}
var bannerClients = []clientUpdateInfo{}
var bannerClientsMux = &sync.RWMutex{}

var updateMap = map[string]string{}

func fetchClientData(repo string) *gitAPIResponse {
	var gitAPI = new(gitAPIResponse)
	resp, err := http.Get("https://api.github.com/repos" + repo + "/releases/latest")
	// resp, err := http.Get("http://localhost:5000/repos" + repo)

	if err != nil {
		logger.Errorf("error retrieving ETH Client Data: %v", err)
		return nil
	}

	defer resp.Body.Close()

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
		return time.Now(), fmt.Errorf("Invalid date string %s %s", date, dTime)
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

func prepareEthClientData(repo string, name string, curTime time.Time) EthClients {
	resp := EthClients{ClientReleaseVersion: "Github"}

	client := fetchClientData(repo)

	if client == nil {
		return resp
	}
	date := strings.Split(client.PublishedDate, "T")

	if len(date) == 2 {
		rTime, err := getRepoTime(date[0], date[1])
		if err != nil {
			logger.Errorf("error parsing git repo. time: %v", err)
			resp.ClientReleaseVersion = client.Name // client.Name is client version from github api
			resp.ClientReleaseDate = uint64(time.Time{}.Unix())
			return resp
		}

		if curTime.Sub(rTime).Hours()/24.0 < 1.0 && updateMap[name] != client.Name {
			if utils.Config.Frontend.Enabled {
				var dbData []uint64
				err = db.DB.Select(&dbData,
					`select user_id
					 from users_subscriptions 
					 where event_filter = $1 AND event_name=$2 AND last_sent_ts <= NOW() - INTERVAL '2 DAY'
					`, strings.ToLower(name), types.EthClientUpdateEventName)
				if err != nil {
					logger.Errorf("error getting user id subscriptions, can't create frontend notifications: %v", err)
				} else {
					for _, uid := range dbData {
						db.AddUserNotification(uid, types.EthClientUpdateEventName, strings.ToLower(name))
					}
				}
			}

			update := clientUpdateInfo{Name: name, Date: rTime}

			bannerClientsMux.Lock()
			bannerClients = append(bannerClients, update)
			bannerClientsMux.Unlock()
			updateMap[name] = client.Name
		}

		resp.ClientReleaseVersion = client.Name // client.Name is client version from github api
		resp.ClientReleaseDate = uint64(rTime.Unix())
		return resp
	}

	return resp // If API limit is exceeded
}

func updateEthClientNetShare() {
	nShare := fetchClientNetworkShare()
	total := 0
	for _, item := range nShare {
		total += item.Value
	}

	for _, item := range nShare {
		switch item.Client {
		case "geth":
			ethClients.Geth.NetworkShare = fmt.Sprintf("%.1f%%", (float64(item.Value)/float64(total))*100.0)
		case "openethereum":
			ethClients.OpenEthereum.NetworkShare = fmt.Sprintf("%.1f%%", (float64(item.Value)/float64(total))*100.0)
		case "nethermind":
			ethClients.Nethermind.NetworkShare = fmt.Sprintf("%.1f%%", (float64(item.Value)/float64(total))*100.0)
		case "besu":
			ethClients.Besu.NetworkShare = fmt.Sprintf("%.1f%%", (float64(item.Value)/float64(total))*100.0)
		default:
			continue
		}
	}
}

func updateEthClient() {
	curTime := time.Now()
	// sending 8 requests to github per call
	// git api rate-limit 60 per hour : 60/8 = 7.5 minutes minimum
	ethClientsMux.Lock()
	defer ethClientsMux.Unlock()

	if curTime.Sub(ethClients.LastUpdate) < time.Hour { // LastUpdate is initialized at January 1, year 1 so no need to check for nil
		return
	}

	logger.Println("Updating ETH Clients Information")

	bannerClientsMux.Lock()
	bannerClients = []clientUpdateInfo{}
	bannerClientsMux.Unlock()

	ethClients.Geth = prepareEthClientData("/ethereum/go-ethereum", "Geth", curTime)
	ethClients.Nethermind = prepareEthClientData("/NethermindEth/nethermind", "Nethermind", curTime)
	ethClients.OpenEthereum = prepareEthClientData("/openethereum/openethereum", "OpenEthereum", curTime)
	ethClients.Besu = prepareEthClientData("/hyperledger/besu", "Besu", curTime)

	ethClients.Teku = prepareEthClientData("/ConsenSys/teku", "Teku", curTime)
	ethClients.Prysm = prepareEthClientData("/prysmaticlabs/prysm", "Prysm", curTime)
	ethClients.Nimbus = prepareEthClientData("/status-im/nimbus-eth2", "Nimbus", curTime)
	ethClients.Lighthouse = prepareEthClientData("/sigp/lighthouse", "Lighthouse", curTime)

	updateEthClientNetShare()

	ethClients.LastUpdate = curTime
}

func updateClients() {
	for true {
		updateEthClient()
		time.Sleep(time.Minute * 5)
	}
}

// GetEthClientData returns a pointer of EthClientServicesPageData
func GetEthClientData() EthClientServicesPageData {
	ethClientsMux.Lock()
	defer ethClientsMux.Unlock()
	return ethClients
}

// ClientsUpdated returns a boolean indicating if clients are updated
func ClientsUpdated() bool {
	bannerClientsMux.Lock()
	defer bannerClientsMux.Unlock()
	if len(bannerClients) == 0 {
		return false
	}
	return true
}

//GetUpdatedClients returns a slice of latest updated clients or empty slice if no updates
func GetUpdatedClients() []clientUpdateInfo {
	bannerClientsMux.Lock()
	defer bannerClientsMux.Unlock()
	return bannerClients
	// return []string{"Prysm", "Teku"}
}
