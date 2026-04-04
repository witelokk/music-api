package internal

import (
	"log/slog"
	"net/http"

	"github.com/witelokk/music-api/internal/auth"
)

func AddRoutes(
	router *http.ServeMux,
	authService *auth.AuthService,
	logger *slog.Logger,
) {
	router.HandleFunc("POST /verification-email", auth.NewSendVerificationEmailHandler(authService, logger))
	router.HandleFunc("POST /users", auth.NewCreateUserHandler(authService, logger))
	router.HandleFunc("POST /tokens", auth.NewGenerateTokensHandler(authService, logger))
}
