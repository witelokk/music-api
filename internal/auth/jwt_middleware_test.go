package auth

import (
	"context"
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

	called := false
	next := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		called = true
		return nil, nil
	}

	handler := mw(next, "GetCurrentUser")

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	w := httptest.NewRecorder()

	resp, err := handler(context.Background(), w, req, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if called {
		t.Fatalf("expected next handler not to be called")
	}

	unauth, ok := resp.(openapi.GetCurrentUser401JSONResponse)
	if !ok {
		t.Fatalf("expected 401 response, got %T", resp)
	}
	if unauth.Error == "" {
		t.Fatalf("expected error message, got empty")
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

	_, err = handler(context.Background(), w, req, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotUserID != "user-id" {
		t.Fatalf("expected user ID %q in context, got %q", "user-id", gotUserID)
	}
}
