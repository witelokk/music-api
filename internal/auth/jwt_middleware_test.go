package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	openapi "github.com/witelokk/music-api/internal/openapi"
)

func TestJWTMiddleware_MissingAuthorizationHeader(t *testing.T) {
	logger := newTestLogger()

	mw := NewJWTMiddleware("test-secret", logger)

	// Simulate a protected operation by setting BearerAuthScopes
	ctx := context.WithValue(context.Background(), openapi.BearerAuthScopes, []string{})

	called := false
	next := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		called = true
		return nil, nil
	}

	handler := mw(next, "GetCurrentUser")

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	w := httptest.NewRecorder()

	resp, err := handler(ctx, w, req, nil)
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got resp=%T err=%v", resp, err)
	}

	if called {
		t.Fatalf("expected next handler not to be called")
	}
	if resp != nil {
		t.Fatalf("expected nil response, got %T", resp)
	}
}

func TestJWTMiddleware_ValidToken_PopulatesContextAndCallsNext(t *testing.T) {
	svc, _, _, _, _ := newTestAuthService()
	logger := newTestLogger()

	token, err := svc.signJWT(jwt.MapClaims{
		"sub": "user-id",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(15 * time.Minute).Unix(),
	})
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	mw := NewJWTMiddleware("test-secret", logger)

	var gotUserID string
	next := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		gotUserID = UserIDFromContext(ctx)
		return nil, nil
	}

	handler := mw(next, "GetCurrentUser")

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	ctx := context.WithValue(context.Background(), openapi.BearerAuthScopes, []string{})

	_, err = handler(ctx, w, req, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotUserID != "user-id" {
		t.Fatalf("expected user ID %q in context, got %q", "user-id", gotUserID)
	}
}
