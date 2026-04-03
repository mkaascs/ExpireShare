package entities

const (
	_ TokenStatus = iota
	Valid
	Expired
	Invalid
	Revoked
)

type TokenStatus int

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
