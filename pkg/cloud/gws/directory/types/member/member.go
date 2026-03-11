package member

type Member struct {
	Kind string `json:"kind,omitempty"`
	Etag string `json:"etag,omitempty"`
	Id   string `json:"id,omitempty"`

	Email            string `json:"email,omitempty"`
	Role             string `json:"role,omitempty"`
	Type             string `json:"type,omitempty"`
	Status           string `json:"status,omitempty"`
	DeliverySettings string `json:"delivery_settings,omitempty"`
}
