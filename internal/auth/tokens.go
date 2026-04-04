package auth

import "time"

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

type RefreshToken struct {
	Token     string
	UserID    string
	ExpiresAt time.Time
}
