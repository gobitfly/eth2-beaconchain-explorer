package types

type ApiResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}
