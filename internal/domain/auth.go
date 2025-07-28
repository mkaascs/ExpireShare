package domain

type Token struct {
	UserId       int64
	RefreshToken string
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
