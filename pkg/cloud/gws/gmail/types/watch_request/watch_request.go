package watch_request

type WatchRequest struct {
	TopicName         string   `json:"topicName,omitzero"`
	LabelIds          []string `json:"labelIds,omitzero"`
	LabelFilterAction string   `json:"labelFilterAction,omitzero"`
}
