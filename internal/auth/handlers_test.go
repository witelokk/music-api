package auth

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
	openapi "github.com/witelokk/music-api/internal/openapi"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

func TestHandleSendVerificationEmail_ValidationErrors(t *testing.T) {
	svc, _, _, _, _ := newTestAuthService()
	logger := newTestLogger()
	ctx := context.Background()

	tests := []struct {
		name      string
		req       openapi.SendVerificationEmailRequestObject
		wantError string
	}{
		{
			name:      "nil body",
			req:       openapi.SendVerificationEmailRequestObject{},
			wantError: "invalid request body",
		},
		{
			name: "missing email",
			req: openapi.SendVerificationEmailRequestObject{
				Body: &openapi.SendVerificationEmailJSONRequestBody{},
			},
			wantError: "email is required",
		},
		{
			name: "invalid email",
			req: openapi.SendVerificationEmailRequestObject{
				Body: &openapi.SendVerificationEmailJSONRequestBody{
					Email: openapi_types.Email("not-an-email"),
				},
			},
			wantError: "invalid email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := HandleSendVerificationEmail(ctx, svc, logger, tt.req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			errResp, ok := resp.(openapi.SendVerificationEmail400JSONResponse)
			if !ok {
				t.Fatalf("expected 400 response, got %T", resp)
			}
			if errResp.Error != tt.wantError {
				t.Fatalf("expected error %q, got %q", tt.wantError, errResp.Error)
			}
		})
	}
}

func TestHandleSendVerificationEmail_Success(t *testing.T) {
	svc, _, _, _, _ := newTestAuthService()
	logger := newTestLogger()
	ctx := context.Background()

	req := openapi.SendVerificationEmailRequestObject{
		Body: &openapi.SendVerificationEmailJSONRequestBody{
			Email: openapi_types.Email("user@example.com"),
		},
	}

	resp, err := HandleSendVerificationEmail(ctx, svc, logger, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := resp.(openapi.SendVerificationEmail204Response); !ok {
		t.Fatalf("expected 204 response, got %T", resp)
	}
}

func TestHandleCreateUser_ValidationErrors(t *testing.T) {
	svc, _, _, _, _ := newTestAuthService()
	logger := newTestLogger()
	ctx := context.Background()

	base := openapi.CreateUserJSONRequestBody{
		Name:  "User",
		Email: openapi_types.Email("user@example.com"),
		Code:  "1234",
	}

	tests := []struct {
		name      string
		modify    func(b *openapi.CreateUserJSONRequestBody)
		wantError string
	}{
		{
			name:      "nil body",
			modify:    nil,
			wantError: "invalid request body",
		},
		{
			name: "missing name",
			modify: func(b *openapi.CreateUserJSONRequestBody) {
				b.Name = ""
			},
			wantError: "name is required",
		},
		{
			name: "missing email",
			modify: func(b *openapi.CreateUserJSONRequestBody) {
				b.Email = ""
			},
			wantError: "email is required",
		},
		{
			name: "invalid email",
			modify: func(b *openapi.CreateUserJSONRequestBody) {
				b.Email = openapi_types.Email("not-an-email")
			},
			wantError: "invalid email",
		},
		{
			name: "missing code",
			modify: func(b *openapi.CreateUserJSONRequestBody) {
				b.Code = ""
			},
			wantError: "verification_code is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req openapi.CreateUserRequestObject
			if tt.modify == nil {
				// nil body case
				req = openapi.CreateUserRequestObject{}
			} else {
				body := base
				tt.modify(&body)
				req = openapi.CreateUserRequestObject{
					Body: &body,
				}
			}

			resp, err := HandleCreateUser(ctx, svc, logger, req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			errResp, ok := resp.(openapi.CreateUser400JSONResponse)
			if !ok {
				t.Fatalf("expected 400 response, got %T", resp)
			}
			if errResp.Error != tt.wantError {
				t.Fatalf("expected error %q, got %q", tt.wantError, errResp.Error)
			}
		})
	}
}

func TestHandleCreateUser_Success(t *testing.T) {
	svc, _, codeRepo, _, _ := newTestAuthService()
	logger := newTestLogger()
	ctx := context.Background()

	email := "user@example.com"
	code := &VerificationCode{
		Code:      "abcd",
		Email:     email,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	codeRepo.codesByEmail[email] = []*VerificationCode{code}

	body := openapi.CreateUserJSONRequestBody{
		Name:  "User",
		Email: openapi_types.Email(email),
		Code:  "abcd",
	}

	req := openapi.CreateUserRequestObject{
		Body: &body,
	}

	resp, err := HandleCreateUser(ctx, svc, logger, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	createdResp, ok := resp.(openapi.CreateUser201JSONResponse)
	if !ok {
		t.Fatalf("expected 201 response, got %T", resp)
	}
	if string(createdResp.Email) != email {
		t.Fatalf("expected email %q, got %q", email, string(createdResp.Email))
	}
	if createdResp.Name != "User" {
		t.Fatalf("expected name %q, got %q", "User", createdResp.Name)
	}
}

func TestHandleGenerateTokens_ValidationErrors(t *testing.T) {
	svc, _, _, _, _ := newTestAuthService()
	logger := newTestLogger()
	ctx := context.Background()

	tests := []struct {
		name      string
		req       openapi.GenerateTokensRequestObject
		wantError string
	}{
		{
			name:      "nil body",
			req:       openapi.GenerateTokensRequestObject{},
			wantError: "invalid request body",
		},
		{
			name: "unsupported grant type",
			req: openapi.GenerateTokensRequestObject{
				Body: &openapi.GenerateTokensJSONRequestBody{
					GrantType: openapi.GetTokensRequestGrantType("unknown"),
				},
			},
			wantError: "unsupported grant_type",
		},
		{
			name: "code grant missing email",
			req: openapi.GenerateTokensRequestObject{
				Body: &openapi.GenerateTokensJSONRequestBody{
					GrantType: openapi.Code,
					Code:      ptr("1234"),
				},
			},
			wantError: "email is required",
		},
		{
			name: "code grant missing code",
			req: openapi.GenerateTokensRequestObject{
				Body: &openapi.GenerateTokensJSONRequestBody{
					GrantType: openapi.Code,
					Email:     ptrEmail("user@example.com"),
				},
			},
			wantError: "verification code is required",
		},
		{
			name: "refresh grant missing token",
			req: openapi.GenerateTokensRequestObject{
				Body: &openapi.GenerateTokensJSONRequestBody{
					GrantType: openapi.RefreshToken,
				},
			},
			wantError: "refresh_token is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := HandleGenerateTokens(ctx, svc, logger, tt.req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			errResp, ok := resp.(openapi.GenerateTokens400JSONResponse)
			if !ok {
				t.Fatalf("expected 400 response, got %T", resp)
			}
			if errResp.Error != tt.wantError {
				t.Fatalf("expected error %q, got %q", tt.wantError, errResp.Error)
			}
		})
	}
}

func TestHandleGenerateTokens_CodeGrant_Success(t *testing.T) {
	svc, userRepo, codeRepo, refreshRepo, _ := newTestAuthService()
	logger := newTestLogger()
	ctx := context.Background()

	email := "user@example.com"
	userRepo.users[email] = &User{
		ID:    "user-id",
		Name:  "User",
		Email: email,
	}

	code := &VerificationCode{
		Code:      "abcd",
		Email:     email,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	codeRepo.codesByEmail[email] = []*VerificationCode{code}

	body := openapi.GenerateTokensJSONRequestBody{
		GrantType: openapi.Code,
		Email:     ptrEmail(email),
		Code:      ptr("abcd"),
	}
	req := openapi.GenerateTokensRequestObject{
		Body: &body,
	}

	resp, err := HandleGenerateTokens(ctx, svc, logger, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	okResp, ok := resp.(openapi.GenerateTokens200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}
	if okResp.AccessToken == "" || okResp.RefreshToken == "" {
		t.Fatalf("expected non-empty tokens, got %#v", okResp)
	}
	if _, exists := refreshRepo.tokens[okResp.RefreshToken]; !exists {
		t.Fatalf("expected refresh token to be stored")
	}
}

func TestHandleGenerateTokens_RefreshGrant_Success(t *testing.T) {
	svc, _, _, refreshRepo, _ := newTestAuthService()
	logger := newTestLogger()
	ctx := context.Background()

	existing := &RefreshToken{
		Token:     "refresh-token",
		UserID:    "user-id",
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	refreshRepo.tokens[existing.Token] = existing

	body := openapi.GenerateTokensJSONRequestBody{
		GrantType:    openapi.RefreshToken,
		RefreshToken: ptr("refresh-token"),
	}
	req := openapi.GenerateTokensRequestObject{
		Body: &body,
	}

	resp, err := HandleGenerateTokens(ctx, svc, logger, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	okResp, ok := resp.(openapi.GenerateTokens200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}
	if okResp.AccessToken == "" || okResp.RefreshToken == "" {
		t.Fatalf("expected non-empty tokens, got %#v", okResp)
	}
	if _, exists := refreshRepo.tokens[okResp.RefreshToken]; !exists {
		t.Fatalf("expected new refresh token to be stored")
	}
}

func TestHandleGetCurrentUser_UnauthorizedWhenNoUserInContext(t *testing.T) {
	svc, _, _, _, _ := newTestAuthService()
	logger := newTestLogger()
	ctx := context.Background()

	req := openapi.GetCurrentUserRequestObject{}

	resp, err := HandleGetCurrentUser(ctx, svc, logger, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errResp, ok := resp.(openapi.GetCurrentUser401JSONResponse)
	if !ok {
		t.Fatalf("expected 401 response, got %T", resp)
	}
	if errResp.Error != "unauthorized" {
		t.Fatalf("expected error %q, got %q", "unauthorized", errResp.Error)
	}
}

func TestHandleGetCurrentUser_Success(t *testing.T) {
	svc, userRepo, _, _, _ := newTestAuthService()
	logger := newTestLogger()
	ctx := context.Background()

	user := &User{
		ID:    "user-id",
		Name:  "User",
		Email: "user@example.com",
	}
	userRepo.users[user.Email] = user
	ctx = WithUserID(ctx, user.ID)

	req := openapi.GetCurrentUserRequestObject{}

	resp, err := HandleGetCurrentUser(ctx, svc, logger, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	okResp, ok := resp.(openapi.GetCurrentUser200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}
	if okResp.Name != user.Name {
		t.Fatalf("expected name %q, got %q", user.Name, okResp.Name)
	}
	if string(okResp.Email) != user.Email {
		t.Fatalf("expected email %q, got %q", user.Email, string(okResp.Email))
	}
}

func ptr(s string) *string {
	return &s
}

func ptrEmail(s string) *openapi_types.Email {
	e := openapi_types.Email(s)
	return &e
}
