package watch_response

type WatchResponse struct {
	HistoryId  string `json:"historyId,omitzero"`
	Expiration string `json:"expiration,omitzero"`
}
