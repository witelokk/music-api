package auth

type SendVerificationEmailRequest struct {
	Email string `json:"email"`
}

type CreateUserRequest struct {
	Name             string `json:"name"`
	Email            string `json:"email"`
	VerificationCode string `json:"code"`
}

type GetTokensRequest struct {
	GrantType        string `json:"grant_type"`
	Email            string `json:"email,omitempty"`
	VerificationCode string `json:"code,omitempty"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	GoogleToken      string `json:"google_token,omitempty"`
}
