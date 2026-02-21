package duration

import "time"

type Duration struct {
	Seconds int `json:"seconds"`
	Nanos   int `json:"nanos"`
}

func New(duration *time.Duration) *Duration {
	return &Duration{
		Seconds: int(*duration / time.Second),
		Nanos:   int(*duration % time.Second),
	}
}
