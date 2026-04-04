package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/mail"
)

func NewSendVerificationEmailHandler(service *AuthService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()

		var req SendVerificationEmailRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Email == "" {
			writeJSONError(w, http.StatusBadRequest, "email is required")
			return
		}

		if _, err := mail.ParseAddress(req.Email); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid email")
			return
		}

		if err := service.SendVerificationEmail(r.Context(), req.Email); err != nil {
			if errors.Is(err, ErrVerificationCodeRecentlySent) {
				writeJSONError(w, http.StatusTooManyRequests, "verification code recently sent")
				return
			}

			logger.Error("failed to send verification email",
				slog.String("email", req.Email),
				slog.String("error", err.Error()),
			)

			writeJSONError(w, http.StatusInternalServerError, "failed to send verification email")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func NewCreateUserHandler(service *AuthService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()

		var req CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Name == "" {
			writeJSONError(w, http.StatusBadRequest, "name is required")
			return
		}

		if req.Email == "" {
			writeJSONError(w, http.StatusBadRequest, "email is required")
			return
		}

		if _, err := mail.ParseAddress(req.Email); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid email")
			return
		}

		if req.VerificationCode == "" {
			writeJSONError(w, http.StatusBadRequest, "verification_code is required")
			return
		}

		user := &User{
			Name:  req.Name,
			Email: req.Email,
		}

		if err := service.CreateUser(r.Context(), user, req.VerificationCode); err != nil {
			switch {
			case errors.Is(err, ErrInvalidVerificationCode):
				writeJSONError(w, http.StatusBadRequest, "invalid verification code")
				return
			case errors.Is(err, ErrExpiredVerificationCode):
				writeJSONError(w, http.StatusBadRequest, "verification code expired")
				return
			case errors.Is(err, ErrUserAlreadyExists):
				writeJSONError(w, http.StatusConflict, "user already exists")
				return
			default:
				logger.Error("failed to create user",
					slog.String("email", req.Email),
					slog.String("error", err.Error()),
				)
				writeJSONError(w, http.StatusInternalServerError, "failed to create user")
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(user)
	}
}

func NewGenerateTokensHandler(service *AuthService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()

		var req GetTokensRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		switch req.GrantType {
		case "code":
			if req.Email == "" {
				writeJSONError(w, http.StatusBadRequest, "email is required")
				return
			}
			if req.VerificationCode == "" {
				writeJSONError(w, http.StatusBadRequest, "verification code is required")
				return
			}

			tokens, err := service.GetTokensWithVerificationCode(r.Context(), req.Email, req.VerificationCode)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidVerificationCode):
					writeJSONError(w, http.StatusBadRequest, "invalid verification code")
					return
				case errors.Is(err, ErrExpiredVerificationCode):
					writeJSONError(w, http.StatusBadRequest, "verification code expired")
					return
				default:
					logger.Error("failed to get tokens with verification code",
						slog.String("email", req.Email),
						slog.String("error", err.Error()),
					)
					writeJSONError(w, http.StatusInternalServerError, "failed to get tokens")
					return
				}
			}

			writeTokensResponse(w, tokens)

		case "refresh_token":
			if req.RefreshToken == "" {
				writeJSONError(w, http.StatusBadRequest, "refresh_token is required")
				return
			}

			tokens, err := service.GetTokensWithRefreshToken(r.Context(), req.RefreshToken)
			if err != nil {
				switch {
				case errors.Is(err, ErrInvalidRefreshToken):
					writeJSONError(w, http.StatusBadRequest, "invalid refresh token")
					return
				case errors.Is(err, ErrExpiredRefreshToken):
					writeJSONError(w, http.StatusBadRequest, "refresh token expired")
					return
				default:
					logger.Error("failed to get tokens with refresh token",
						slog.String("error", err.Error()),
					)
					writeJSONError(w, http.StatusInternalServerError, "failed to get tokens")
					return
				}
			}

			writeTokensResponse(w, tokens)

		default:
			writeJSONError(w, http.StatusBadRequest, "unsupported grant_type")
		}
	}
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func writeTokensResponse(w http.ResponseWriter, tokens *Tokens) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(TokensResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}
