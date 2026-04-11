package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/witelokk/music-api/internal/auth/idtoken"
)

var ErrVerificationCodeRecentlySent = errors.New("verification code recently sent")
var ErrInvalidVerificationCode = errors.New("invalid verification code")
var ErrExpiredVerificationCode = errors.New("verification code expired")
var ErrInvalidRefreshToken = errors.New("invalid refresh token")
var ErrExpiredRefreshToken = errors.New("refresh token expired")
var ErrUserAlreadyExists = errors.New("user already exists")
var ErrInvalidAccessToken = errors.New("invalid access token")
var ErrInvalidGoogleIDToken = errors.New("invalid Google ID token")
var ErrInvalidAppleIDToken = errors.New("invalid Apple ID token")

type AuthServiceParams struct {
	JWTSecret                   string
	AccessTokenTTL              time.Duration
	RefreshTokenTTL             time.Duration
	VerificationCodeTTL         time.Duration
	NewVerificationCodeInterval time.Duration
	GoogleIdTokenVerifier       *idtoken.Validator
	AppleIdTokenVerifier        *idtoken.Validator
}

type AuthService struct {
	userRepository             UserRepository
	verificationCodeRepository VerificationCodeRepository
	refreshTokenRepository     RefreshTokenRepository
	emailSender                EmailSender
	params                     AuthServiceParams
}

func NewAuthService(
	userRepository UserRepository,
	verificationCodeRepository VerificationCodeRepository,
	refreshTokenRepository RefreshTokenRepository,
	emailSender EmailSender,
	params AuthServiceParams,
) *AuthService {
	return &AuthService{
		userRepository:             userRepository,
		verificationCodeRepository: verificationCodeRepository,
		refreshTokenRepository:     refreshTokenRepository,
		emailSender:                emailSender,
		params:                     params,
	}
}

func (s *AuthService) SendVerificationEmail(
	ctx context.Context,
	Email string,
) error {
	if err := s.ensureCanIssueNewCode(ctx, Email); err != nil {
		return err
	}

	code, err := s.createAndStoreVerificationCode(ctx, Email)
	if err != nil {
		return err
	}

	if err := s.sendVerificationEmail(ctx, Email, code); err != nil {
		return err
	}

	return nil
}

func (s *AuthService) CreateUser(ctx context.Context, user *User, verificationCode string) error {
	_, err := s.validateVerificationCode(ctx, user.Email, verificationCode)
	if err != nil {
		return err
	}

	if err := s.createUserWithVerifiedEmail(ctx, user); err != nil {
		return err
	}

	return nil
}

func (s *AuthService) GetTokensWithVerificationCode(ctx context.Context, email, verificationCode string) (*Tokens, error) {
	_, err := s.validateVerificationCode(ctx, email, verificationCode)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	tokens, err := s.generateTokensForUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *AuthService) GetTokensWithRefreshToken(ctx context.Context, refreshToken string) (*Tokens, error) {
	stored, err := s.refreshTokenRepository.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}
	if stored == nil {
		return nil, ErrInvalidRefreshToken
	}

	if time.Now().After(stored.ExpiresAt) {
		return nil, ErrExpiredRefreshToken
	}

	if err := s.refreshTokenRepository.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to invalidate refresh token: %w", err)
	}

	return s.generateTokensForUser(ctx, stored.UserID)
}

func (s *AuthService) GetTokensWithGoogleIDToken(ctx context.Context, idToken string) (*Tokens, error) {
	payload, err := s.params.GoogleIdTokenVerifier.Validate(ctx, idToken)
	if err != nil {
		return nil, errors.Join(ErrInvalidGoogleIDToken, err)
	}

	user, err := s.userRepository.GetUserByEmail(ctx, payload["email"].(string))
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		user = &User{
			Name:  payload["name"].(string),
			Email: payload["email"].(string),
		}

		if err := s.createUserWithVerifiedEmail(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	return s.generateTokensForUser(ctx, user.ID)
}

func (s *AuthService) GetTokensWithAppleIDToken(ctx context.Context, idToken string) (*Tokens, error) {
	payload, err := s.params.AppleIdTokenVerifier.Validate(ctx, idToken)
	if err != nil {
		return nil, errors.Join(ErrInvalidAppleIDToken, err)
	}

	email, ok := payload["email"].(string)
	if !ok || email == "" {
		return nil, ErrInvalidAppleIDToken
	}

	name, _ := payload["name"].(string)

	user, err := s.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		user = &User{
			Name:  name,
			Email: email,
		}

		if err := s.createUserWithVerifiedEmail(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	return s.generateTokensForUser(ctx, user.ID)
}

func (s *AuthService) GetCurrentUser(ctx context.Context) (*User, error) {
	userID := UserIDFromContext(ctx)
	if userID == "" {
		return nil, ErrInvalidAccessToken
	}

	user, err := s.userRepository.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (s *AuthService) validateVerificationCode(ctx context.Context, email, verificationCode string) (*VerificationCode, error) {
	codes, err := s.verificationCodeRepository.GetCodesByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get verification codes: %w", err)
	}

	var matched *VerificationCode
	for _, c := range codes {
		if c.Code == verificationCode {
			matched = c
			break
		}
	}

	if matched == nil {
		return nil, ErrInvalidVerificationCode
	}

	now := time.Now()
	if now.After(matched.ExpiresAt) {
		_ = s.verificationCodeRepository.DeleteCode(ctx, email, verificationCode)
		return nil, ErrExpiredVerificationCode
	}

	return matched, nil
}

func (s *AuthService) createUserWithVerifiedEmail(ctx context.Context, user *User) error {
	existing, err := s.userRepository.GetUserByEmail(ctx, user.Email)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if existing != nil {
		return ErrUserAlreadyExists
	}

	createdUser, err := s.userRepository.CreateUser(ctx, user.Name, user.Email)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	if createdUser != nil && user != nil {
		*user = *createdUser
	}

	return nil
}

func (s *AuthService) ensureCanIssueNewCode(ctx context.Context, email string) error {
	existingCodes, err := s.verificationCodeRepository.GetCodesByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to get verification codes: %w", err)
	}

	if len(existingCodes) == 0 {
		return nil
	}

	latest := existingCodes[0]
	for _, c := range existingCodes[1:] {
		if c.ExpiresAt.After(latest.ExpiresAt) {
			latest = c
		}
	}

	issuedAt := latest.ExpiresAt.Add(-s.params.VerificationCodeTTL)
	if time.Since(issuedAt) < s.params.NewVerificationCodeInterval {
		return ErrVerificationCodeRecentlySent
	}

	return nil
}

func (s *AuthService) createAndStoreVerificationCode(ctx context.Context, email string) (*VerificationCode, error) {
	code := &VerificationCode{
		Code:      generateVerificationCode(),
		Email:     email,
		ExpiresAt: time.Now().Add(s.params.VerificationCodeTTL),
	}

	if err := s.verificationCodeRepository.SaveCode(ctx, code); err != nil {
		return nil, fmt.Errorf("failed to save verification code: %w", err)
	}

	return code, nil
}

func (s *AuthService) sendVerificationEmail(ctx context.Context, email string, code *VerificationCode) error {
	to := []string{email}
	subject := "Account Verification"
	text := fmt.Sprintf(`Your Music code: <strong>%s</strong>
                         <br><br>
                         The code will expire in %d minutes`, code.Code, int(s.params.VerificationCodeTTL.Minutes()))

	if err := s.emailSender.SendEmail(ctx, to, subject, text); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

func (s *AuthService) generateTokensForUser(ctx context.Context, userID string) (*Tokens, error) {
	now := time.Now()

	accessTTL := s.params.AccessTokenTTL
	if accessTTL == 0 {
		accessTTL = 15 * time.Minute
	}

	refreshTTL := s.params.RefreshTokenTTL
	if refreshTTL == 0 {
		refreshTTL = 30 * 24 * time.Hour
	}

	accessToken, err := s.signJWT(jwt.MapClaims{
		"sub": userID,
		"iat": now.Unix(),
		"exp": now.Add(accessTTL).Unix(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := generateRefreshTokenString()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	if s.refreshTokenRepository != nil {
		rt := &RefreshToken{
			Token:     refreshToken,
			UserID:    userID,
			ExpiresAt: now.Add(refreshTTL),
		}
		if err := s.refreshTokenRepository.SaveRefreshToken(ctx, rt); err != nil {
			return nil, fmt.Errorf("failed to store refresh token: %w", err)
		}
	}

	return &Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) signJWT(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.params.JWTSecret))
}

// creates a random 4-digit code as string
func generateVerificationCode() string {
	var b [2]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "0000"
	}
	n := int(b[0])<<8 | int(b[1])
	return fmt.Sprintf("%04d", n%10000)
}

func generateRefreshTokenString() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
