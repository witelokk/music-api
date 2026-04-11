package idtoken

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/bartventer/httpcache"
	"github.com/golang-jwt/jwt/v5"
)

// helper to create a validator pre-populated with a key and
// a signed ID token string for testing.
func newTestValidatorAndToken(t *testing.T, allowedAudience, tokenAudience string) (*Validator, string) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	kid := "test-kid"

	// Stub transport returning a certs document with our public key.
	certs := googleIDTokenCerts{
		Keys: []key{
			{
				KeyID: kid,
				N:     base64.RawURLEncoding.EncodeToString(privateKey.N.Bytes()),
				E:     base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.E)).Bytes()),
			},
		},
	}

	body, err := json.Marshal(certs)
	if err != nil {
		t.Fatalf("failed to marshal certs: %v", err)
	}

	client := &http.Client{
		Transport: httpcache.NewTransport(
			"memcache://",
			httpcache.WithUpstream(&roundTripperStub{
				body:   body,
				header: http.Header{},
			}),
		),
	}

	v := newValidatorWithClient([]string{allowedAudience}, client)

	claims := jwt.MapClaims{
		"iss":   "https://accounts.google.com",
		"aud":   tokenAudience,
		"email": "user@example.com",
		"name":  "Test User",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iat":   time.Now().Add(-time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	return v, tokenString
}

type roundTripperStub struct {
	body   []byte
	header http.Header
	calls  int
}

func (r *roundTripperStub) RoundTrip(req *http.Request) (*http.Response, error) {
	r.calls++

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(r.body)),
		Header:     r.header.Clone(),
		Request:    req,
	}
	return resp, nil
}

func TestValidatorValidate_Success(t *testing.T) {
	ctx := context.Background()
	aud := "my-audience"

	v, tokenString := newTestValidatorAndToken(t, aud, aud)

	claims, err := v.Validate(ctx, tokenString)
	if err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}

	if claims["email"] != "user@example.com" {
		t.Fatalf("expected email %q, got %v", "user@example.com", claims["email"])
	}
	if claims["name"] != "Test User" {
		t.Fatalf("expected name %q, got %v", "Test User", claims["name"])
	}
	if claims["aud"] != aud {
		t.Fatalf("expected audience %q, got %v", aud, claims["aud"])
	}
}

func TestValidatorValidate_InvalidAudience(t *testing.T) {
	ctx := context.Background()

	allowedAudience := "expected-audience"
	tokenAudience := "other-audience"

	v, tokenString := newTestValidatorAndToken(t, allowedAudience, tokenAudience)

	_, err := v.Validate(ctx, tokenString)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if got, want := err.Error(), "invalid audience"; got != want {
		t.Fatalf("expected error %q, got %q", want, got)
	}
}

func TestValidatorValidate_InvalidSignature(t *testing.T) {
	ctx := context.Background()
	aud := "my-audience"

	// First create a validator and token to set up keys.
	v, _ := newTestValidatorAndToken(t, aud, aud)

	// Now create a second RSA key and sign a token with the same kid
	// but a different key so that signature verification fails.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	claims := jwt.MapClaims{
		"iss":   "https://accounts.google.com",
		"aud":   aud,
		"email": "user@example.com",
		"name":  "Test User",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iat":   time.Now().Add(-time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-kid"

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, err = v.Validate(ctx, tokenString)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if got, want := err.Error(), "failed to parse token: token signature is invalid: crypto/rsa: verification error"; got != want {
		t.Fatalf("expected error %q, got %q", want, got)
	}
}

func TestValidatorValidate_InvalidIssuer(t *testing.T) {
	ctx := context.Background()

	allowedAudience := "my-audience"

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	kid := "test-kid"
	certs := googleIDTokenCerts{
		Keys: []key{
			{
				KeyID: kid,
				N:     base64.RawURLEncoding.EncodeToString(privateKey.N.Bytes()),
				E:     base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.E)).Bytes()),
			},
		},
	}
	body, err := json.Marshal(certs)
	if err != nil {
		t.Fatalf("failed to marshal certs: %v", err)
	}

	client := &http.Client{
		Transport: &roundTripperStub{
			body:   body,
			header: http.Header{},
		},
	}

	v := newValidatorWithClient([]string{allowedAudience}, client)

	claims := jwt.MapClaims{
		"iss":   "https://evil.example.com",
		"aud":   allowedAudience,
		"email": "user@example.com",
		"name":  "Test User",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iat":   time.Now().Add(-time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, err = v.Validate(ctx, tokenString)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if got, want := err.Error(), "invalid issuer"; got != want {
		t.Fatalf("expected error %q, got %q", want, got)
	}
}

func TestValidatorValidate_ExpiredToken(t *testing.T) {
	ctx := context.Background()

	allowedAudience := "my-audience"

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	kid := "test-kid"
	certs := googleIDTokenCerts{
		Keys: []key{
			{
				KeyID: kid,
				N:     base64.RawURLEncoding.EncodeToString(privateKey.N.Bytes()),
				E:     base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.E)).Bytes()),
			},
		},
	}
	body, err := json.Marshal(certs)
	if err != nil {
		t.Fatalf("failed to marshal certs: %v", err)
	}

	client := &http.Client{
		Transport: &roundTripperStub{
			body:   body,
			header: http.Header{},
		},
	}

	v := newValidatorWithClient([]string{allowedAudience}, client)

	claims := jwt.MapClaims{
		"iss":   "https://accounts.google.com",
		"aud":   allowedAudience,
		"email": "user@example.com",
		"name":  "Test User",
		"exp":   time.Now().Add(-time.Hour).Unix(),
		"iat":   time.Now().Add(-2 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, err = v.Validate(ctx, tokenString)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if got, want := err.Error(), "failed to parse token: token has invalid claims: token is expired"; got != want {
		t.Fatalf("expected error %q, got %q", want, got)
	}
}

func TestValidatorFetchCerts_PopulatesKeysAndExpiry(t *testing.T) {
	certs := googleIDTokenCerts{
		Keys: []key{
			{
				KeyID: "kid-1",
				N:     base64.RawURLEncoding.EncodeToString([]byte("test-modulus")),
				E:     base64.RawURLEncoding.EncodeToString([]byte{0x01, 0x00, 0x01}),
			},
		},
	}

	body, err := json.Marshal(certs)
	if err != nil {
		t.Fatalf("failed to marshal certs: %v", err)
	}

	stub := &roundTripperStub{
		body: body,
		header: http.Header{
			"Cache-Control": []string{"public, max-age=600"},
		},
	}

	client := &http.Client{
		Transport: httpcache.NewTransport(
			"memcache://",
			httpcache.WithUpstream(stub),
		),
	}

	v := newValidatorWithClient([]string{"aud"}, client)

	if err := v.fetchCerts(context.Background()); err != nil {
		t.Fatalf("fetchCerts() error = %v, want nil", err)
	}

	if len(v.keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(v.keys))
	}

	if stub.calls != 1 {
		t.Fatalf("expected 1 HTTP call, got %d", stub.calls)
	}
}

func TestValidatorValidate_RefreshesCertsWhenNearExpiry(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	certs := googleIDTokenCerts{
		Keys: []key{
			{
				KeyID: "kid-1",
				N:     base64.RawURLEncoding.EncodeToString(privateKey.N.Bytes()),
				E:     base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.E)).Bytes()),
			},
		},
	}

	body, err := json.Marshal(certs)
	if err != nil {
		t.Fatalf("failed to marshal certs: %v", err)
	}

	stub := &roundTripperStub{
		body:   body,
		header: http.Header{},
	}

	client := &http.Client{
		Transport: httpcache.NewTransport(
			"memcache://",
			httpcache.WithUpstream(stub),
		),
	}

	v := newValidatorWithClient([]string{"my-audience"}, client)

	claims := jwt.MapClaims{
		"iss":   "https://accounts.google.com",
		"aud":   "my-audience",
		"email": "user@example.com",
		"name":  "Test User",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iat":   time.Now().Add(-time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "kid-1"

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	// First call should fetch certs.
	if _, err := v.Validate(context.Background(), tokenString); err != nil {
		t.Fatalf("first Validate() error = %v, want nil", err)
	}
	if stub.calls != 1 {
		t.Fatalf("expected 1 HTTP call after first validate, got %d", stub.calls)
	}

	// Second call should also succeed; in this setup the response is not
	// cacheable, so we still expect another upstream call.
	if _, err := v.Validate(context.Background(), tokenString); err != nil {
		t.Fatalf("second Validate() error = %v, want nil", err)
	}
	if stub.calls != 2 {
		t.Fatalf("expected 2 HTTP calls after second validate, got %d", stub.calls)
	}
}

func TestValidatorFetchCerts_UsesHttpcacheForCaching(t *testing.T) {
	certs := googleIDTokenCerts{
		Keys: []key{
			{
				KeyID: "kid-1",
				N:     "test-key",
			},
		},
	}

	body, err := json.Marshal(certs)
	if err != nil {
		t.Fatalf("failed to marshal certs: %v", err)
	}

	upstream := &roundTripperStub{
		body: body,
		header: http.Header{
			"Cache-Control": []string{"public, max-age=600"},
		},
	}

	client := &http.Client{
		Transport: httpcache.NewTransport(
			"memcache://",
			httpcache.WithUpstream(upstream),
		),
	}

	v := newValidatorWithClient([]string{"aud"}, client)

	if err := v.fetchCerts(context.Background()); err != nil {
		t.Fatalf("first fetchCerts() error = %v, want nil", err)
	}
	if len(v.keys) != 1 {
		t.Fatalf("expected 1 key after first fetch, got %d", len(v.keys))
	}
	if upstream.calls != 1 {
		t.Fatalf("expected 1 upstream call after first fetch, got %d", upstream.calls)
	}

	if err := v.fetchCerts(context.Background()); err != nil {
		t.Fatalf("second fetchCerts() error = %v, want nil", err)
	}
	if upstream.calls != 1 {
		t.Fatalf("expected cached response on second fetch (still 1 upstream call), got %d", upstream.calls)
	}
}
