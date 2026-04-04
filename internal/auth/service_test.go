package auth

import (
	"context"
	"strings"
	"testing"
	"time"
)

type fakeUserRepository struct {
	users          map[string]*User
	getErr        error
	createErr     error
	lastCreatedID string
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{
		users: make(map[string]*User),
	}
}

func (r *fakeUserRepository) CreateUser(ctx context.Context, name, email string) (*User, error) {
	if r.createErr != nil {
		return nil, r.createErr
	}

	u := &User{
		ID:        "user-" + email,
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
	}
	r.users[email] = u
	r.lastCreatedID = u.ID

	return u, nil
}

func (r *fakeUserRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	u, ok := r.users[email]
	if !ok {
		return nil, nil
	}
	return u, nil
}

type fakeVerificationCodeRepository struct {
	codesByEmail map[string][]*VerificationCode
	saveErr      error
	getErr       error
	deleteErr    error

	lastDeletedEmail string
	lastDeletedCode  string
}

func newFakeVerificationCodeRepository() *fakeVerificationCodeRepository {
	return &fakeVerificationCodeRepository{
		codesByEmail: make(map[string][]*VerificationCode),
	}
}

func (r *fakeVerificationCodeRepository) SaveCode(ctx context.Context, code *VerificationCode) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.codesByEmail[code.Email] = append(r.codesByEmail[code.Email], code)
	return nil
}

func (r *fakeVerificationCodeRepository) GetCodesByEmail(ctx context.Context, email string) ([]*VerificationCode, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	return r.codesByEmail[email], nil
}

func (r *fakeVerificationCodeRepository) DeleteCode(ctx context.Context, email string, code string) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}

	r.lastDeletedEmail = email
	r.lastDeletedCode = code

	codes := r.codesByEmail[email]
	for i, c := range codes {
		if c.Code == code {
			r.codesByEmail[email] = append(codes[:i], codes[i+1:]...)
			break
		}
	}

	return nil
}

type fakeRefreshTokenRepository struct {
	tokens    map[string]*RefreshToken
	saveErr   error
	getErr    error
	deleteErr error

	lastDeletedToken string
}

func newFakeRefreshTokenRepository() *fakeRefreshTokenRepository {
	return &fakeRefreshTokenRepository{
		tokens: make(map[string]*RefreshToken),
	}
}

func (r *fakeRefreshTokenRepository) SaveRefreshToken(ctx context.Context, token *RefreshToken) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.tokens[token.Token] = token
	return nil
}

func (r *fakeRefreshTokenRepository) GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	rt, ok := r.tokens[token]
	if !ok {
		return nil, nil
	}
	return rt, nil
}

func (r *fakeRefreshTokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	delete(r.tokens, token)
	r.lastDeletedToken = token
	return nil
}

type fakeEmailSender struct {
	sent []struct {
		to      []string
		subject string
		text    string
	}
	err error
}

func (s *fakeEmailSender) SendEmail(ctx context.Context, to []string, subject, text string) error {
	if s.err != nil {
		return s.err
	}
	s.sent = append(s.sent, struct {
		to      []string
		subject string
		text    string
	}{
		to:      to,
		subject: subject,
		text:    text,
	})
	return nil
}

func newTestAuthService() (*AuthService, *fakeUserRepository, *fakeVerificationCodeRepository, *fakeRefreshTokenRepository, *fakeEmailSender) {
	userRepo := newFakeUserRepository()
	codeRepo := newFakeVerificationCodeRepository()
	refreshRepo := newFakeRefreshTokenRepository()
	emailSender := &fakeEmailSender{}

	service := NewAuthService(
		userRepo,
		codeRepo,
		refreshRepo,
		emailSender,
		AuthServiceParams{
			JWTSecret:                   "test-secret",
			AccessTokenTTL:              15 * time.Minute,
			RefreshTokenTTL:             30 * 24 * time.Hour,
			VerificationCodeTTL:         15 * time.Minute,
			NewVerificationCodeInterval: 10 * time.Minute,
		},
	)

	return service, userRepo, codeRepo, refreshRepo, emailSender
}

func TestAuthService_SendVerificationEmail_Success(t *testing.T) {
	svc, _, codeRepo, _, emailSender := newTestAuthService()
	ctx := context.Background()
	email := "user@example.com"

	if err := svc.SendVerificationEmail(ctx, email); err != nil {
		t.Fatalf("SendVerificationEmail() error = %v, want nil", err)
	}

	codes := codeRepo.codesByEmail[email]
	if len(codes) != 1 {
		t.Fatalf("expected 1 code saved, got %d", len(codes))
	}
	if codes[0].Code == "" {
		t.Fatalf("expected non-empty verification code")
	}
	if len(emailSender.sent) != 1 {
		t.Fatalf("expected 1 email sent, got %d", len(emailSender.sent))
	}
	if len(emailSender.sent[0].to) != 1 || emailSender.sent[0].to[0] != email {
		t.Fatalf("unexpected email recipients: %#v", emailSender.sent[0].to)
	}
}

func TestAuthService_SendVerificationEmail_RecentlySent(t *testing.T) {
	svc, _, codeRepo, _, emailSender := newTestAuthService()
	ctx := context.Background()
	email := "user@example.com"

	// simulate an existing code issued recently
	now := time.Now()
	codeRepo.codesByEmail[email] = []*VerificationCode{
		{
			Code:      "1234",
			Email:     email,
			ExpiresAt: now.Add(svc.params.VerificationCodeTTL - 5*time.Minute),
		},
	}

	err := svc.SendVerificationEmail(ctx, email)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != ErrVerificationCodeRecentlySent {
		t.Fatalf("expected ErrVerificationCodeRecentlySent, got %v", err)
	}
	if len(emailSender.sent) != 0 {
		t.Fatalf("expected no email sent, got %d", len(emailSender.sent))
	}
}

func TestAuthService_CreateUser_Success(t *testing.T) {
	svc, userRepo, codeRepo, _, _ := newTestAuthService()
	ctx := context.Background()
	email := "user@example.com"

	code := &VerificationCode{
		Code:      "abcd",
		Email:     email,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	codeRepo.codesByEmail[email] = []*VerificationCode{code}

	user := &User{
		Name:  "Test User",
		Email: email,
	}

	if err := svc.CreateUser(ctx, user, "abcd"); err != nil {
		t.Fatalf("CreateUser() error = %v, want nil", err)
	}

	stored, ok := userRepo.users[email]
	if !ok {
		t.Fatalf("expected user to be stored")
	}
	if user.ID == "" {
		t.Fatalf("expected user ID to be set")
	}
	if stored.ID != user.ID {
		t.Fatalf("expected stored user ID %q, got %q", user.ID, stored.ID)
	}
}

func TestAuthService_CreateUser_InvalidCode(t *testing.T) {
	svc, userRepo, codeRepo, _, _ := newTestAuthService()
	ctx := context.Background()
	email := "user@example.com"

	codeRepo.codesByEmail[email] = []*VerificationCode{
		{
			Code:      "other",
			Email:     email,
			ExpiresAt: time.Now().Add(5 * time.Minute),
		},
	}

	user := &User{
		Name:  "Test User",
		Email: email,
	}

	err := svc.CreateUser(ctx, user, "abcd")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != ErrInvalidVerificationCode {
		t.Fatalf("expected ErrInvalidVerificationCode, got %v", err)
	}
	if len(userRepo.users) != 0 {
		t.Fatalf("expected no users to be created, got %d", len(userRepo.users))
	}
}

func TestAuthService_CreateUser_ExpiredCode(t *testing.T) {
	svc, userRepo, codeRepo, _, _ := newTestAuthService()
	ctx := context.Background()
	email := "user@example.com"

	code := &VerificationCode{
		Code:      "abcd",
		Email:     email,
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	}
	codeRepo.codesByEmail[email] = []*VerificationCode{code}

	user := &User{
		Name:  "Test User",
		Email: email,
	}

	err := svc.CreateUser(ctx, user, "abcd")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != ErrExpiredVerificationCode {
		t.Fatalf("expected ErrExpiredVerificationCode, got %v", err)
	}
	if len(userRepo.users) != 0 {
		t.Fatalf("expected no users to be created, got %d", len(userRepo.users))
	}
	if codeRepo.lastDeletedEmail != email || codeRepo.lastDeletedCode != "abcd" {
		t.Fatalf("expected verification code to be deleted, got email=%q code=%q", codeRepo.lastDeletedEmail, codeRepo.lastDeletedCode)
	}
}

func TestAuthService_CreateUser_UserAlreadyExists(t *testing.T) {
	svc, userRepo, codeRepo, _, _ := newTestAuthService()
	ctx := context.Background()
	email := "user@example.com"

	code := &VerificationCode{
		Code:      "abcd",
		Email:     email,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	codeRepo.codesByEmail[email] = []*VerificationCode{code}

	// pre-existing user
	userRepo.users[email] = &User{
		ID:    "existing-id",
		Name:  "Existing",
		Email: email,
	}

	user := &User{
		Name:  "Test User",
		Email: email,
	}

	err := svc.CreateUser(ctx, user, "abcd")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != ErrUserAlreadyExists {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestAuthService_GetTokensWithVerificationCode_Success(t *testing.T) {
	svc, userRepo, codeRepo, refreshRepo, _ := newTestAuthService()
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

	tokens, err := svc.GetTokensWithVerificationCode(ctx, email, "abcd")
	if err != nil {
		t.Fatalf("GetTokensWithVerificationCode() error = %v, want nil", err)
	}
	if tokens == nil {
		t.Fatalf("expected tokens, got nil")
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("expected non-empty tokens, got %#v", tokens)
	}
	rt, ok := refreshRepo.tokens[tokens.RefreshToken]
	if !ok {
		t.Fatalf("expected refresh token to be stored")
	}
	if rt.UserID != "user-id" {
		t.Fatalf("expected refresh token user ID %q, got %q", "user-id", rt.UserID)
	}
}

func TestAuthService_GetTokensWithVerificationCode_UserNotFound(t *testing.T) {
	svc, _, codeRepo, _, _ := newTestAuthService()
	ctx := context.Background()
	email := "user@example.com"

	code := &VerificationCode{
		Code:      "abcd",
		Email:     email,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	codeRepo.codesByEmail[email] = []*VerificationCode{code}

	_, err := svc.GetTokensWithVerificationCode(ctx, email, "abcd")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "user not found") {
		t.Fatalf("expected error to contain %q, got %v", "user not found", err)
	}
}

func TestAuthService_GetTokensWithRefreshToken_Success(t *testing.T) {
	svc, _, _, refreshRepo, _ := newTestAuthService()
	ctx := context.Background()

	existing := &RefreshToken{
		Token:     "refresh-token",
		UserID:    "user-id",
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	refreshRepo.tokens[existing.Token] = existing

	tokens, err := svc.GetTokensWithRefreshToken(ctx, existing.Token)
	if err != nil {
		t.Fatalf("GetTokensWithRefreshToken() error = %v, want nil", err)
	}
	if tokens == nil {
		t.Fatalf("expected tokens, got nil")
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("expected non-empty tokens, got %#v", tokens)
	}
	if _, ok := refreshRepo.tokens["refresh-token"]; ok {
		t.Fatalf("expected old refresh token to be deleted")
	}
	if _, ok := refreshRepo.tokens[tokens.RefreshToken]; !ok {
		t.Fatalf("expected new refresh token to be stored")
	}
}

func TestAuthService_GetTokensWithRefreshToken_Invalid(t *testing.T) {
	svc, _, _, refreshRepo, _ := newTestAuthService()
	ctx := context.Background()

	// empty repo → nil token
	refreshRepo.tokens = make(map[string]*RefreshToken)

	_, err := svc.GetTokensWithRefreshToken(ctx, "unknown")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != ErrInvalidRefreshToken {
		t.Fatalf("expected ErrInvalidRefreshToken, got %v", err)
	}
}

func TestAuthService_GetTokensWithRefreshToken_Expired(t *testing.T) {
	svc, _, _, refreshRepo, _ := newTestAuthService()
	ctx := context.Background()

	expired := &RefreshToken{
		Token:     "refresh-token",
		UserID:    "user-id",
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	}
	refreshRepo.tokens[expired.Token] = expired

	_, err := svc.GetTokensWithRefreshToken(ctx, expired.Token)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != ErrExpiredRefreshToken {
		t.Fatalf("expected ErrExpiredRefreshToken, got %v", err)
	}
	if _, ok := refreshRepo.tokens[expired.Token]; !ok {
		t.Fatalf("expected expired refresh token to remain stored")
	}
}

