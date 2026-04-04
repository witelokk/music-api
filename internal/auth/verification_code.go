package auth

import "time"

type VerificationCode struct {
	Code      string
	Email     string
	ExpiresAt time.Time
}
