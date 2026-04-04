package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

func decodeErrorBody(t *testing.T, body []byte) string {
	t.Helper()
	var resp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}
	return resp.Error
}

func TestNewSendVerificationEmailHandler_MethodNotAllowed(t *testing.T) {
	svc, _, _, _, _ := newTestAuthService()
	logger := newTestLogger()

	req := httptest.NewRequest(http.MethodGet, "/verification-email", nil)
	rr := httptest.NewRecorder()

	handler := NewSendVerificationEmailHandler(svc, logger)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestNewSendVerificationEmailHandler_ValidationAndDomainErrors(t *testing.T) {
	svc, _, codeRepo, _, emailSender := newTestAuthService()
	logger := newTestLogger()

	tests := []struct {
		name           string
		body           string
		setup          func()
		wantStatus     int
		wantError      string
		expectEmailSent bool
	}{
		{
			name:       "invalid JSON",
			body:       "{",
			setup:      func() {},
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid request body",
		},
		{
			name:       "missing email",
			body:       `{}`,
			setup:      func() {},
			wantStatus: http.StatusBadRequest,
			wantError:  "email is required",
		},
		{
			name:       "invalid email",
			body:       `{"email":"not-an-email"}`,
			setup:      func() {},
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid email",
		},
		{
			name: "recently sent",
			body: `{"email":"user@example.com"}`,
			setup: func() {
				email := "user@example.com"
				codeRepo.codesByEmail[email] = []*VerificationCode{
					{
						Code:      "1234",
						Email:     email,
						ExpiresAt: time.Now().Add(svc.params.VerificationCodeTTL - 5*time.Minute),
					},
				}
			},
			wantStatus: http.StatusTooManyRequests,
			wantError:  "verification code recently sent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// reset state between cases
			codeRepo.codesByEmail = make(map[string][]*VerificationCode)
			emailSender.sent = nil

			tt.setup()

			req := httptest.NewRequest(http.MethodPost, "/verification-email", bytes.NewBufferString(tt.body))
			rr := httptest.NewRecorder()

			handler := NewSendVerificationEmailHandler(svc, logger)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d, body=%s", tt.wantStatus, rr.Code, rr.Body.String())
			}
			if rr.Code != http.StatusNoContent {
				errMsg := decodeErrorBody(t, rr.Body.Bytes())
				if errMsg != tt.wantError {
					t.Fatalf("expected error %q, got %q", tt.wantError, errMsg)
				}
			}
		})
	}
}

func TestNewSendVerificationEmailHandler_Success(t *testing.T) {
	svc, _, codeRepo, _, emailSender := newTestAuthService()
	logger := newTestLogger()

	email := "user@example.com"

	req := httptest.NewRequest(http.MethodPost, "/verification-email", bytes.NewBufferString(`{"email":"`+email+`"}`))
	rr := httptest.NewRecorder()

	handler := NewSendVerificationEmailHandler(svc, logger)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusNoContent, rr.Code, rr.Body.String())
	}

	if len(codeRepo.codesByEmail[email]) != 1 {
		t.Fatalf("expected 1 code saved, got %d", len(codeRepo.codesByEmail[email]))
	}
	if len(emailSender.sent) != 1 {
		t.Fatalf("expected 1 email sent, got %d", len(emailSender.sent))
	}
}

func TestNewCreateUserHandler_MethodNotAllowed(t *testing.T) {
	svc, _, _, _, _ := newTestAuthService()
	logger := newTestLogger()

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rr := httptest.NewRecorder()

	handler := NewCreateUserHandler(svc, logger)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestNewCreateUserHandler_ValidationErrors(t *testing.T) {
	svc, _, _, _, _ := newTestAuthService()
	logger := newTestLogger()

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantError  string
	}{
		{
			name:       "invalid JSON",
			body:       "{",
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid request body",
		},
		{
			name:       "missing name",
			body:       `{"email":"user@example.com","code":"1234"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "name is required",
		},
		{
			name:       "missing email",
			body:       `{"name":"Test","code":"1234"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "email is required",
		},
		{
			name:       "invalid email",
			body:       `{"name":"Test","email":"not-an-email","code":"1234"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid email",
		},
		{
			name:       "missing verification code",
			body:       `{"name":"Test","email":"user@example.com"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "verification_code is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(tt.body))
			rr := httptest.NewRecorder()

			handler := NewCreateUserHandler(svc, logger)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d, body=%s", tt.wantStatus, rr.Code, rr.Body.String())
			}

			errMsg := decodeErrorBody(t, rr.Body.Bytes())
			if errMsg != tt.wantError {
				t.Fatalf("expected error %q, got %q", tt.wantError, errMsg)
			}
		})
	}
}

func TestNewCreateUserHandler_DomainErrorsAndSuccess(t *testing.T) {
	svc, userRepo, codeRepo, _, _ := newTestAuthService()
	logger := newTestLogger()
	ctx := context.Background()

	email := "user@example.com"
	body := `{"name":"Test","email":"` + email + `","code":"abcd"}`

	t.Run("invalid verification code", func(t *testing.T) {
		codeRepo.codesByEmail[email] = []*VerificationCode{
			{Code: "other", Email: email, ExpiresAt: time.Now().Add(5 * time.Minute)},
		}

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
		rr := httptest.NewRecorder()

		handler := NewCreateUserHandler(svc, logger)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
		errMsg := decodeErrorBody(t, rr.Body.Bytes())
		if errMsg != "invalid verification code" {
			t.Fatalf("expected error %q, got %q", "invalid verification code", errMsg)
		}
	})

	t.Run("expired verification code", func(t *testing.T) {
		codeRepo.codesByEmail[email] = []*VerificationCode{
			{Code: "abcd", Email: email, ExpiresAt: time.Now().Add(-1 * time.Minute)},
		}

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
		rr := httptest.NewRecorder()

		handler := NewCreateUserHandler(svc, logger)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
		errMsg := decodeErrorBody(t, rr.Body.Bytes())
		if errMsg != "verification code expired" {
			t.Fatalf("expected error %q, got %q", "verification code expired", errMsg)
		}
	})

	t.Run("user already exists", func(t *testing.T) {
		codeRepo.codesByEmail[email] = []*VerificationCode{
			{Code: "abcd", Email: email, ExpiresAt: time.Now().Add(5 * time.Minute)},
		}
		userRepo.users[email] = &User{
			ID:    "existing",
			Name:  "Existing",
			Email: email,
		}

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
		rr := httptest.NewRecorder()

		handler := NewCreateUserHandler(svc, logger)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusConflict {
			t.Fatalf("expected status %d, got %d", http.StatusConflict, rr.Code)
		}
		errMsg := decodeErrorBody(t, rr.Body.Bytes())
		if errMsg != "user already exists" {
			t.Fatalf("expected error %q, got %q", "user already exists", errMsg)
		}
	})

	t.Run("success", func(t *testing.T) {
		codeRepo.codesByEmail[email] = []*VerificationCode{
			{Code: "abcd", Email: email, ExpiresAt: time.Now().Add(5 * time.Minute)},
		}
		userRepo.users = make(map[string]*User)

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
		rr := httptest.NewRecorder()

		handler := NewCreateUserHandler(svc, logger)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected status %d, got %d, body=%s", http.StatusCreated, rr.Code, rr.Body.String())
		}

		var resp User
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if resp.Email != email || resp.Name != "Test" {
			t.Fatalf("unexpected user in response: %#v", resp)
		}
	})

	// ensure context usage doesn't panic (sanity check)
	if _, err := svc.userRepository.GetUserByEmail(ctx, email); err != nil && !strings.Contains(err.Error(), "user") {
		// no-op, just ensuring fake repo compiles with context
	}
}

func TestNewGenerateTokensHandler_MethodNotAllowed(t *testing.T) {
	svc, _, _, _, _ := newTestAuthService()
	logger := newTestLogger()

	req := httptest.NewRequest(http.MethodGet, "/tokens", nil)
	rr := httptest.NewRecorder()

	handler := NewGenerateTokensHandler(svc, logger)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestNewGenerateTokensHandler_ValidationErrors(t *testing.T) {
	svc, _, _, _, _ := newTestAuthService()
	logger := newTestLogger()

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantError  string
	}{
		{
			name:       "invalid JSON",
			body:       "{",
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid request body",
		},
		{
			name:       "unsupported grant_type",
			body:       `{"grant_type":"unknown"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "unsupported grant_type",
		},
		{
			name:       "code grant missing email",
			body:       `{"grant_type":"code","code":"1234"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "email is required",
		},
		{
			name:       "code grant missing code",
			body:       `{"grant_type":"code","email":"user@example.com"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "verification code is required",
		},
		{
			name:       "refresh grant missing refresh_token",
			body:       `{"grant_type":"refresh_token"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "refresh_token is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/tokens", bytes.NewBufferString(tt.body))
			rr := httptest.NewRecorder()

			handler := NewGenerateTokensHandler(svc, logger)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d, body=%s", tt.wantStatus, rr.Code, rr.Body.String())
			}

			errMsg := decodeErrorBody(t, rr.Body.Bytes())
			if errMsg != tt.wantError {
				t.Fatalf("expected error %q, got %q", tt.wantError, errMsg)
			}
		})
	}
}

func TestNewGenerateTokensHandler_CodeGrant_DomainErrorsAndSuccess(t *testing.T) {
	svc, userRepo, codeRepo, _, _ := newTestAuthService()
	logger := newTestLogger()

	email := "user@example.com"
	body := `{"grant_type":"code","email":"` + email + `","code":"abcd"}`

	t.Run("invalid verification code", func(t *testing.T) {
		codeRepo.codesByEmail[email] = []*VerificationCode{
			{Code: "other", Email: email, ExpiresAt: time.Now().Add(5 * time.Minute)},
		}

		req := httptest.NewRequest(http.MethodPost, "/tokens", bytes.NewBufferString(body))
		rr := httptest.NewRecorder()

		handler := NewGenerateTokensHandler(svc, logger)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
		errMsg := decodeErrorBody(t, rr.Body.Bytes())
		if errMsg != "invalid verification code" {
			t.Fatalf("expected error %q, got %q", "invalid verification code", errMsg)
		}
	})

	t.Run("expired verification code", func(t *testing.T) {
		codeRepo.codesByEmail[email] = []*VerificationCode{
			{Code: "abcd", Email: email, ExpiresAt: time.Now().Add(-1 * time.Minute)},
		}

		req := httptest.NewRequest(http.MethodPost, "/tokens", bytes.NewBufferString(body))
		rr := httptest.NewRecorder()

		handler := NewGenerateTokensHandler(svc, logger)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
		errMsg := decodeErrorBody(t, rr.Body.Bytes())
		if errMsg != "verification code expired" {
			t.Fatalf("expected error %q, got %q", "verification code expired", errMsg)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		codeRepo.codesByEmail[email] = []*VerificationCode{
			{Code: "abcd", Email: email, ExpiresAt: time.Now().Add(5 * time.Minute)},
		}
		userRepo.users = make(map[string]*User)

		req := httptest.NewRequest(http.MethodPost, "/tokens", bytes.NewBufferString(body))
		rr := httptest.NewRecorder()

		handler := NewGenerateTokensHandler(svc, logger)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
		}
		errMsg := decodeErrorBody(t, rr.Body.Bytes())
		if errMsg != "failed to get tokens" {
			t.Fatalf("expected error %q, got %q", "failed to get tokens", errMsg)
		}
	})

	t.Run("success", func(t *testing.T) {
		codeRepo.codesByEmail[email] = []*VerificationCode{
			{Code: "abcd", Email: email, ExpiresAt: time.Now().Add(5 * time.Minute)},
		}
		userRepo.users[email] = &User{
			ID:    "user-id",
			Name:  "User",
			Email: email,
		}

		req := httptest.NewRequest(http.MethodPost, "/tokens", bytes.NewBufferString(body))
		rr := httptest.NewRecorder()

		handler := NewGenerateTokensHandler(svc, logger)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
		}

		var resp TokensResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if resp.AccessToken == "" || resp.RefreshToken == "" {
			t.Fatalf("expected non-empty tokens, got %#v", resp)
		}
	})
}

func TestNewGenerateTokensHandler_RefreshGrant_DomainErrorsAndSuccess(t *testing.T) {
	svc, _, _, refreshRepo, _ := newTestAuthService()
	logger := newTestLogger()

	body := func(token string) string {
		return `{"grant_type":"refresh_token","refresh_token":"` + token + `"}`
	}

	t.Run("invalid refresh token", func(t *testing.T) {
		refreshRepo.tokens = make(map[string]*RefreshToken)

		req := httptest.NewRequest(http.MethodPost, "/tokens", bytes.NewBufferString(body("unknown")))
		rr := httptest.NewRecorder()

		handler := NewGenerateTokensHandler(svc, logger)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
		errMsg := decodeErrorBody(t, rr.Body.Bytes())
		if errMsg != "invalid refresh token" {
			t.Fatalf("expected error %q, got %q", "invalid refresh token", errMsg)
		}
	})

	t.Run("expired refresh token", func(t *testing.T) {
		refreshRepo.tokens = map[string]*RefreshToken{
			"token": {
				Token:     "token",
				UserID:    "user-id",
				ExpiresAt: time.Now().Add(-1 * time.Minute),
			},
		}

		req := httptest.NewRequest(http.MethodPost, "/tokens", bytes.NewBufferString(body("token")))
		rr := httptest.NewRecorder()

		handler := NewGenerateTokensHandler(svc, logger)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
		errMsg := decodeErrorBody(t, rr.Body.Bytes())
		if errMsg != "refresh token expired" {
			t.Fatalf("expected error %q, got %q", "refresh token expired", errMsg)
		}
	})

	t.Run("success", func(t *testing.T) {
		refreshRepo.tokens = map[string]*RefreshToken{
			"token": {
				Token:     "token",
				UserID:    "user-id",
				ExpiresAt: time.Now().Add(5 * time.Minute),
			},
		}

		req := httptest.NewRequest(http.MethodPost, "/tokens", bytes.NewBufferString(body("token")))
		rr := httptest.NewRecorder()

		handler := NewGenerateTokensHandler(svc, logger)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
		}

		var resp TokensResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if resp.AccessToken == "" || resp.RefreshToken == "" {
			t.Fatalf("expected non-empty tokens, got %#v", resp)
		}
	})
}

