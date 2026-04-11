package google_id_token

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"slices"
	"time"

	"github.com/bartventer/httpcache"
	_ "github.com/bartventer/httpcache/store/memcache"
	"github.com/golang-jwt/jwt/v5"
)

const GOOGLE_CERTS_URL = "https://www.googleapis.com/oauth2/v3/certs"

type key struct {
	KeyType   string `json:"kty"`
	Algorithm string `json:"alg"`
	Use       string `json:"use"`
	KeyID     string `json:"kid"`
	N         string `json:"n"`
	E         string `json:"e"`
}

type googleIDTokenCerts struct {
	Keys []key `json:"keys"`
}

type Validator struct {
	audiences []string
	client    *http.Client
	keys      map[string]key
}

func NewValidator(audiences []string) *Validator {
	client := httpcache.NewClient("memcache://")

	return &Validator{
		audiences: audiences,
		client:    client,
	}
}

func newValidatorWithClient(audiences []string, client *http.Client) *Validator {
	return &Validator{
		audiences: audiences,
		client:    client,
	}
}

func (v *Validator) Validate(ctx context.Context, idToken string) (jwt.MapClaims, error) {
	if err := v.fetchCerts(ctx); err != nil {
		return nil, err
	}

	token, err := jwt.Parse(idToken, v.keyFunc)
	if err != nil {
		return nil, errors.New("failed to parse token: " + err.Error())
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	payload := token.Claims.(jwt.MapClaims)

	if err := v.validateIssuer(payload); err != nil {
		return nil, err
	}

	if err := validateExpiration(payload); err != nil {
		return nil, err
	}

	if err := v.validateAudience(payload); err != nil {
		return nil, err
	}

	return token.Claims.(jwt.MapClaims), nil
}

func (v *Validator) keyFunc(token *jwt.Token) (interface{}, error) {
	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, errors.New("missing or invalid key ID")
	}

	key, ok := v.keys[kid]
	if !ok {
		return nil, errors.New("missing or invalid key ID")
	}

	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, errors.New("invalid key modulus")
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, errors.New("invalid key exponent")
	}
	if len(eBytes) == 0 {
		return nil, errors.New("invalid key exponent")
	}

	var n big.Int
	n.SetBytes(nBytes)

	e := 0
	for _, b := range eBytes {
		e = e<<8 | int(b)
	}
	if e <= 0 {
		return nil, errors.New("invalid key exponent")
	}

	return &rsa.PublicKey{
		N: &n,
		E: e,
	}, nil
}

func (v *Validator) fetchCerts(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, GOOGLE_CERTS_URL, nil)
	if err != nil {
		return err
	}

	client := v.client
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	cacheStatus := resp.Header.Get(httpcache.CacheStatusHeader)
	if v.keys != nil && (cacheStatus == "HIT" || cacheStatus == "STALE") {
		return nil
	}

	var certs googleIDTokenCerts
	if err := json.NewDecoder(resp.Body).Decode(&certs); err != nil {
		return err
	}

	v.keys = make(map[string]key)
	for _, k := range certs.Keys {
		v.keys[k.KeyID] = k
	}
	return nil
}

func (v *Validator) validateIssuer(payload jwt.MapClaims) error {
	issuer, ok := payload["iss"].(string)
	if !ok || (issuer != "accounts.google.com" && issuer != "https://accounts.google.com") {
		return errors.New("invalid issuer")
	}
	return nil
}

func validateExpiration(payload jwt.MapClaims) error {
	expValue, ok := payload["exp"]
	if !ok {
		return errors.New("invalid expiration time")
	}

	expFloat, ok := expValue.(float64)
	if !ok {
		return errors.New("invalid expiration time")
	}

	expTime := time.Unix(int64(expFloat), 0)
	if time.Now().After(expTime) {
		return errors.New("token expired")
	}

	return nil
}

func (v *Validator) validateAudience(payload jwt.MapClaims) error {
	aud, ok := payload["aud"].(string)
	if !ok {
		return errors.New("invalid audience")
	}

	if !slices.Contains(v.audiences, aud) {
		return errors.New("invalid audience")
	}

	return nil
}
