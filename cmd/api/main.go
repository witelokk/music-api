package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/witelokk/music-api/internal"
	"github.com/witelokk/music-api/internal/auth"
)

func main() {
	config := internal.MustLoadConfig()
	logger := internal.NewLogger(config.Logger.Type, config.Logger.Level)

	db, err := pgxpool.New(context.Background(), config.DatabaseURL)
	if err != nil {
		logger.Error("Failed to connect to the database:", slog.String("error", err.Error()))
	}
	defer db.Close()

	redis := redis.NewClient(&redis.Options{
		Addr: config.RedisURL,
	})
	defer redis.Close()

	userRespository := auth.NewPostgresUserRepository(db)
	verificationCodeRepository := auth.NewRedisVerificationCodeRepository(redis)
	refreshTokenRespository := auth.NewRedisRefreshTokenRepository(redis)
	emailSender := auth.NewMailgunEmailSender(
		config.Mailgun.APIKey,
		config.Mailgun.Domain,
		config.Mailgun.From,
		auth.MailGunRegion(config.Mailgun.Region),
	)

	authService := auth.NewAuthService(
		userRespository,
		verificationCodeRepository,
		refreshTokenRespository,
		emailSender,
		auth.AuthServiceParams{
			JWTSecret:                   config.Auth.JWTSecret,
			AccessTokenTTL:              config.Auth.AccessTokenTTL,
			RefreshTokenTTL:             config.Auth.RefreshTokenTTL,
			VerificationCodeTTL:         config.Auth.VerificationCodeTTL,
			NewVerificationCodeInterval: config.Auth.NewVerificationCodeInterval,
		},
	)

	router := http.NewServeMux()
	internal.AddRoutes(router, authService, logger)

	server := &http.Server{
		Addr:         config.HttpServer.Host + ":" + config.HttpServer.Port,
		Handler:      router,
		ReadTimeout:  config.HttpServer.Timeouts.Read,
		WriteTimeout: config.HttpServer.Timeouts.Write,
		IdleTimeout:  config.HttpServer.Timeouts.Idle,
	}

	logger.Info("Starting server", slog.String("address", server.Addr))
	if err := server.ListenAndServe(); err != nil {
		logger.Error("Failed to start the server:", slog.String("error", err.Error()))
	}
}
