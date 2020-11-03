package kong

import (
	"io/ioutil"
	"net/http"
	"strings"
)

func post_to_kong_admin(api_url string, data string) (response string) {
	reqBody := strings.NewReader(data)
	resp, err := http.Post("http://0.0.0.0:8001"+api_url,
		"application/x-www-form-urlencoded", reqBody)
	if err != nil {
		print(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		print(err)
	}
	return string(body)
}

func SetupKong() {
	post_to_kong_admin("/services/", "url=http://0.0.0.0:3333&name=api")
	post_to_kong_admin("/services/api/routes/", "hosts[]=0.0.0.0")
	post_to_kong_admin("/services/api/routes/", "hosts[]=127.0.0.1")
	post_to_kong_admin("/services/api/routes/", "hosts[]=localhost")
	post_to_kong_admin("/services/api/routes/", "hosts[]=beaconcha.in")
	post_to_kong_admin("/services/api/plugins/", "name=key-auth")
}

func addConsumer(userID string, apiKey string) {
	post_to_kong_admin("/consumers/", "username="+userID) // 'custom_id' doesn't work, 'username' is used for storing the id
	post_to_kong_admin("/consumers/"+userID+"/key-auth/", "key="+apiKey)
}

func SubUserToSapphire(userID string, apiKey string) {
	addConsumer(userID, apiKey)
	post_to_kong_admin("/consumers/"+userID+"/plugins/",
		"name=rate-limiting&config.minute=100&config.day=100000&config.month=500000&config.policy=local")
}

func SubUserToEmerald(userID string, apiKey string) {
	addConsumer(userID, apiKey)
	post_to_kong_admin("/consumers/"+userID+"/plugins/",
		"name=rate-limiting&config.day=200000&config.month=1000000&config.policy=local")
}

func SubUserToDiamond(userID string, apiKey string) {
	addConsumer(userID, apiKey)
	post_to_kong_admin("/consumers/"+userID+"/plugins/",
		"name=rate-limiting&config.month=4000000&config.policy=local")
}

func SubUserToFree(userID string, apiKey string) {
	addConsumer(userID, apiKey)
	post_to_kong_admin("/consumers/"+userID+"/plugins/",
		"name=rate-limiting&config.minute=10&config.day=10000&config.month=30000&config.policy=local")
}

//usage
// func main() {
// 	setupKong()
// 	SubUserToSapphire("username", "apikey")

// }
