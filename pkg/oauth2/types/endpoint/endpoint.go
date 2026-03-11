package endpoint

type AuthStyle int

const (
	AuthStyleAutoDetect AuthStyle = iota
	AuthStyleInParams
	AuthStyleInHeader
)

type Endpoint struct {
	AuthURL   string
	TokenURL  string
	AuthStyle AuthStyle
}
