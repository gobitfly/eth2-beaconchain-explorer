package ethClients

import (
	"encoding/json"
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

type gitApiResponse struct {
	Url       string `json:"url"`
	AssetsUrl string `json:"assets_url"`
	UploadUrl string `json:"upload_url"`
	HtmlUrl   string `json:"html_url"`
	ID        uint64 `json:"id"`
	Author    struct {
		Login        string `json:"login"`
		ID           uint64 `json:"id"`
		NodeID       string `json:"node_id"`
		AvatarURL    string `json:"avatar_url"`
		GravatarID   string `json:"gravatar_id"`
		Url          string `json:"url"`
		HtmlUrl      string `json:"html_url"`
		FollowersUrl string `json:"followers_url"`
		FollowingUrl string `json:"following_url"`
		GistsUrl     string `json:"gists_url"`
		StarredUrl   string `json:"starred_url"`
		SubUrl       string `json:"subscriptions_url"`
		OrgUrl       string `json:"organizations_url"`
		ReposUrl     string `json:"repos_url"`
		EventsUrl    string `json:"events_url"`
		RxEventUrl   string `json:"received_events_url"`
		Type         string `json:"type"`
		SiteAdmin    bool   `json:"site_admin"`
	} `json:"author"`
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

// Init starts a go routine to update the ETH Clients Info
func Init() {
	go update()
}

func fetchClientData(repo string) *gitApiResponse {
	gitApi := new(gitApiResponse)

	resp, err := http.Get("https://api.github.com/repos" + repo + "/releases/latest")

	if err != nil {
		logger.Errorf("error retrieving ETH Client Data: %v", err)
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&gitApi)

	if err != nil {
		logger.Errorf("error decoding ETH Clients json response to struct: %v", err)
	}

	return gitApi
}

func prepareEthClientData(repo string, curTime time.Time) (string, string) {
	var year, month, day int64
	var err error
	client := fetchClientData(repo)
	date := strings.Split(client.PublishedDate, "T")
	if len(date) > 0 {
		dateDays := strings.Split(date[0], "-")
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
		rTime := time.Date(int(year), time.Month(int(month)), int(day), 0, 0, 0, 0, time.UTC)
		timeDiff := curTime.Sub(rTime).Hours() / 24
		if timeDiff < 1 {
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
	ethClients.Geth.ClientReleaseVersion, ethClients.Geth.ClientReleaseDate = prepareEthClientData("/ethereum/go-ethereum", curTime)
	ethClients.Nethermind.ClientReleaseVersion, ethClients.Nethermind.ClientReleaseDate = prepareEthClientData("/NethermindEth/nethermind", curTime)
	ethClients.OpenEthereum.ClientReleaseVersion, ethClients.OpenEthereum.ClientReleaseDate = prepareEthClientData("/openethereum/openethereum", curTime)
	ethClients.Besu.ClientReleaseVersion, ethClients.Besu.ClientReleaseDate = prepareEthClientData("/hyperledger/besu", curTime)

	ethClients.Teku.ClientReleaseVersion, ethClients.Teku.ClientReleaseDate = prepareEthClientData("/ConsenSys/teku", curTime)
	ethClients.Prysm.ClientReleaseVersion, ethClients.Prysm.ClientReleaseDate = prepareEthClientData("/prysmaticlabs/prysm", curTime)
	ethClients.Nimbus.ClientReleaseVersion, ethClients.Nimbus.ClientReleaseDate = prepareEthClientData("/status-im/nimbus-eth2", curTime)
	ethClients.Lighthouse.ClientReleaseVersion, ethClients.Lighthouse.ClientReleaseDate = prepareEthClientData("/sigp/lighthouse", curTime)

	ethClients.LastUpdate = curTime
}

func update() {
	for true {
		updateEthClient()
		time.Sleep(time.Minute * 20)
	}
}

// GetEthClientData returns a pointer of EthClientServicesPageData
func GetEthClientData() *types.EthClientServicesPageData {
	ethClientsMux.Lock()
	defer ethClientsMux.Unlock()
	return ethClients
}
