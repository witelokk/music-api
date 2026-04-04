package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/witelokk/music-api/internal"
	"github.com/witelokk/music-api/internal/artists"
	"github.com/witelokk/music-api/internal/auth"
	"github.com/witelokk/music-api/internal/releases"
	"github.com/witelokk/music-api/internal/songs"
	openapi "github.com/witelokk/music-api/internal/openapi"
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
	songsRepository := songs.NewPostgresRepository(db)
	artistsRepository := artists.NewPostgresRepository(db)
	releasesRepository := releases.NewPostgresRepository(db)
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

	songsService := songs.NewService(songsRepository)
	artistsService := artists.NewService(artistsRepository)
	releasesService := releases.NewService(releasesRepository)

	serverImpl := internal.NewServer(authService, songsService, artistsService, releasesService, logger)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /openapi.yml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "openapi.yml")
	})
	strictHandler := openapi.NewStrictHandler(
		serverImpl,
		[]openapi.StrictMiddlewareFunc{
			auth.NewJWTMiddleware(config.Auth.JWTSecret, logger),
		},
	)
	handler := openapi.HandlerFromMux(strictHandler, mux)

	corsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		handler.ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:         config.HttpServer.Host + ":" + config.HttpServer.Port,
		Handler:      corsHandler,
		ReadTimeout:  config.HttpServer.Timeouts.Read,
		WriteTimeout: config.HttpServer.Timeouts.Write,
		IdleTimeout:  config.HttpServer.Timeouts.Idle,
	}

	logger.Info("Starting server", slog.String("address", server.Addr))
	if err := server.ListenAndServe(); err != nil {
		logger.Error("Failed to start the server:", slog.String("error", err.Error()))
	}
}
