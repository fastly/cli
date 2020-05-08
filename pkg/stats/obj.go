package stats

// The structs in this file are similar to those in go-fastly, but
// intended for json use rather than mapstructure.

type statsResponse struct {
	Status string            `json:"status"`
	Msg    string            `json:"msg"`
	Meta   statsResponseMeta `json:"meta"`

	Data map[string][]statsResponseData `json:"data"`
}

type statsResponseMeta struct {
	From   string `json:"from"`
	To     string `json:"to"`
	By     string `json:"by"`
	Region string `json:"region"`
}

type statsResponseData map[string]interface{}

type realtimeResponse struct {
	Timestamp uint64                 `json:"timestamp"`
	Data      []realtimeResponseData `json:"data"`
}

type realtimeResponseData struct {
	Recorded   float64           `json:"recorded"`
	Aggregated statsResponseData `json:"aggregated"`
}
