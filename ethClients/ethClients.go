package ethclients

import (
	"encoding/json"
	"eth2-exporter/types"
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
		return gitAPI
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&gitAPI)

	if err != nil {
		logger.Errorf("error decoding ETH Clients json response to struct: %v", err)
	}

	return gitAPI
}

func fetchClientNetworkShare() []ethernodesAPIStruct {
	var ethernodesAPI []ethernodesAPIStruct
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

func ymdTodmy(date string) string {
	dateDays := strings.Split(date, "-")
	if len(dateDays) < 3 {
		logger.Errorf("error wrong date string %s", date)
		return ""
	}
	return fmt.Sprintf("%s-%s-%s", dateDays[2], dateDays[1], dateDays[0])
}

func prepareEthClientData(repo string, name string, curTime time.Time) (string, string) {

	client := fetchClientData(repo)
	date := strings.Split(client.PublishedDate, "T")

	if len(date) == 2 {
		rTime, err := getRepoTime(date[0], date[1])
		if err != nil {
			logger.Errorf("error parsing git repo. time: %v", err)
			return client.Name, "GitHub"
		}
		timeDiff := (curTime.Sub(rTime).Hours() / 24.0)
		if timeDiff < 2.0 { // show banner if update was less than 2 days ago
			bannerClients += fmt.Sprintf("<a href=\"/ethClients#ethClientsServices\" class=\"text-primary mr-2\">%s %s</a>\n", name, client.Name)
			return client.Name, "Recently"
		}

		if timeDiff > 30 {
			return client.Name, fmt.Sprintf("On %s", ymdTodmy(date[0]))
		}

		return client.Name, fmt.Sprintf("%.0f days ago", timeDiff) // can sub. -0.5 to round down the days but github is rounding up
	}
	return client.Name, "GitHub" // If API limit is exceeded
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
	if curTime.Sub(ethClients.LastUpdate) < time.Hour { // LastUpdate is initialized at January 1, year 1 so no need to check for nil
		return
	}

	logger.Println("Updating ETH Clients Information")
	ethClientsMux.Lock()
	defer ethClientsMux.Unlock()
	bannerClientsMux.Lock()
	defer bannerClientsMux.Unlock()
	bannerClients = ""
	updateEthClientNetShare()
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
		updateEthClient()
		time.Sleep(time.Minute * 5)
	}
}

// GetEthClientData returns a pointer of EthClientServicesPageData
func GetEthClientData() *types.EthClientServicesPageData {
	ethClientsMux.Lock()
	defer ethClientsMux.Unlock()
	return ethClients
}

// GetBannerClients returns a string of latest updates of ETH clients
func GetBannerClients() *template.HTML {
	bannerClientsMux.Lock()
	defer bannerClientsMux.Unlock()
	if bannerClients == "" {
		return nil
	}
	temp := template.HTML(fmt.Sprintf(`<i class="fab fa-github mr-2" aria-hidden="true"></i>
									   <span class="mr-2">Latest Client Releases:</span>
									   %s`, bannerClients))
	return &temp
}
