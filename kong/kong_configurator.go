package kong

import (
	"io/ioutil"
	"net/http"
	"strings"
)

func postToKongAdmin(apiURL string, data string) (response string) {
	reqBody := strings.NewReader(data)
	resp, err := http.Post("http://0.0.0.0:8001"+apiURL,
		"application/x-www-form-urlencoded", reqBody)
	if err != nil {
		logger.Errorf("error posting to kong admin: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("error receiving kong admin response: %v", err)
	}
	return string(body)
}

// SetupKong limit any '/api' calls to the kong port 8000 from specified hosts
func SetupKong() {
	postToKongAdmin("/services/", "url=http://0.0.0.0:3333&name=api")
	postToKongAdmin("/services/api/routes/", "hosts[]=0.0.0.0")
	postToKongAdmin("/services/api/routes/", "hosts[]=127.0.0.1")
	postToKongAdmin("/services/api/routes/", "hosts[]=localhost")
	postToKongAdmin("/services/api/routes/", "hosts[]=beaconcha.in")
	postToKongAdmin("/services/api/plugins/", "name=key-auth")
}

func addConsumer(userID string, apiKey string) {
	postToKongAdmin("/consumers/", "username="+userID) // 'custom_id' doesn't work, 'username' is used for storing the id
	postToKongAdmin("/consumers/"+userID+"/key-auth/", "key="+apiKey)
}

// SubUserToSapphire subscribe a user to Sapphire plan
func SubUserToSapphire(userID string, apiKey string) {
	addConsumer(userID, apiKey)
	postToKongAdmin("/consumers/"+userID+"/plugins/",
		"name=rate-limiting&config.minute=100&config.day=100000&config.month=500000&config.policy=local")
}

// SubUserToEmerald subscribe a user to Emerald plan
func SubUserToEmerald(userID string, apiKey string) {
	addConsumer(userID, apiKey)
	postToKongAdmin("/consumers/"+userID+"/plugins/",
		"name=rate-limiting&config.day=200000&config.month=1000000&config.policy=local")
}

// SubUserToDiamond subscribe a user to Diamond plan
func SubUserToDiamond(userID string, apiKey string) {
	addConsumer(userID, apiKey)
	postToKongAdmin("/consumers/"+userID+"/plugins/",
		"name=rate-limiting&config.month=4000000&config.policy=local")
}

// SubUserToFree subscribe a user to Free plan
func SubUserToFree(userID string, apiKey string) {
	addConsumer(userID, apiKey)
	postToKongAdmin("/consumers/"+userID+"/plugins/",
		"name=rate-limiting&config.minute=10&config.day=10000&config.month=30000&config.policy=local")
}
