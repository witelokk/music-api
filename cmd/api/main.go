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
	"github.com/witelokk/music-api/internal/favorites"
	"github.com/witelokk/music-api/internal/followings"
	"github.com/witelokk/music-api/internal/releases"
	"github.com/witelokk/music-api/internal/songs"
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
	songsRepository := songs.NewPostgresSongsRepository(db)
	artistsRepository := artists.NewPostgresArtistsRepository(db)
	releasesRepository := releases.NewPostgresReleasesRepository(db)
	favoritesRepository := favorites.NewPostgresFavoritesRepository(db)
	followingsRepository := followings.NewPostgresFollowingsRepository(db)
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
	favoritesService := favorites.NewFavoritesService(favoritesRepository)
	followingsService := followings.NewFollowingsService(followingsRepository)

	serverImpl := internal.NewServer(authService, songsService, artistsService, releasesService, favoritesService, followingsService, logger)
	httpHandler := internal.NewHTTPHandler(
		serverImpl,
		internal.HTTPHandlerConfig{
			JWTSecret: config.Auth.JWTSecret,
		},
		logger,
	)

	server := &http.Server{
		Addr:         config.HttpServer.Host + ":" + config.HttpServer.Port,
		Handler:      httpHandler,
		ReadTimeout:  config.HttpServer.Timeouts.Read,
		WriteTimeout: config.HttpServer.Timeouts.Write,
		IdleTimeout:  config.HttpServer.Timeouts.Idle,
	}

	logger.Info("Starting server", slog.String("address", server.Addr))
	if err := server.ListenAndServe(); err != nil {
		logger.Error("Failed to start the server:", slog.String("error", err.Error()))
	}
}
