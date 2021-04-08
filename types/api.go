package types

type ApiResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

type WidgetResponse struct {
	Eff       interface{} `json:"efficiency"`
	Validator interface{} `json:"validator"`
	Epoch     int64       `json:"epoch"`
}
