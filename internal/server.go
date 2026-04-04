package internal

import (
	"context"

	"log/slog"

	"github.com/witelokk/music-api/internal/auth"
	openapi "github.com/witelokk/music-api/internal/openapi"
)

type Server struct {
	authService *auth.AuthService
	logger      *slog.Logger
}

func NewServer(authService *auth.AuthService, logger *slog.Logger) openapi.StrictServerInterface {
	return &Server{
		authService: authService,
		logger:      logger,
	}
}

func (s *Server) SendVerificationEmail(ctx context.Context, req openapi.SendVerificationEmailRequestObject) (openapi.SendVerificationEmailResponseObject, error) {
	return auth.HandleSendVerificationEmail(ctx, s.authService, s.logger, req)
}

func (s *Server) CreateUser(ctx context.Context, req openapi.CreateUserRequestObject) (openapi.CreateUserResponseObject, error) {
	return auth.HandleCreateUser(ctx, s.authService, s.logger, req)
}

func (s *Server) GenerateTokens(ctx context.Context, req openapi.GenerateTokensRequestObject) (openapi.GenerateTokensResponseObject, error) {
	return auth.HandleGenerateTokens(ctx, s.authService, s.logger, req)
}
