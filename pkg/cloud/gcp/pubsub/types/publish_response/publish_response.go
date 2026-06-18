package publish_response

// Response is the result of a topics:publish call, containing the server-assigned
// ids of the published messages, in the same order as the request.
type Response struct {
	MessageIds []string `json:"messageIds,omitempty"`
}
