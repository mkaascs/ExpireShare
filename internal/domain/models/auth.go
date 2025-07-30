package models

import "time"

type Token struct {
	UserId       int64
	RefreshToken string
	ExpiresAt    time.Time
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
