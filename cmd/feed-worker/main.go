package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/witelokk/music-api/internal"
	"github.com/witelokk/music-api/internal/favorites"
	"github.com/witelokk/music-api/internal/followings"
	"github.com/witelokk/music-api/internal/home_feed"
	"github.com/witelokk/music-api/internal/releases"
)

func main() {
	config := internal.MustLoadConfig()
	logger := internal.NewLogger(config.Logger.Type, config.Logger.Level)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := pgxpool.New(ctx, config.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to the database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: config.RedisURL,
	})
	defer redisClient.Close()

	favoritesRepository := favorites.NewPostgresFavoritesRepository(db)
	followingsRepository := followings.NewPostgresFollowingsRepository(db)
	releasesRepository := releases.NewPostgresReleasesRepository(db)
	feedComposer := home_feed.NewRecommendationFeedComposer(favoritesRepository, followingsRepository, releasesRepository)
	feedRefreshQueue := home_feed.NewRedisFeedRefreshQueue(redisClient, home_feed.RedisFeedRefreshQueueOptions{})
	feedSnapshots := home_feed.NewPostgresFeedSnapshotRepository(db)

	worker := home_feed.NewFeedWorker(
		feedRefreshQueue,
		feedComposer,
		feedSnapshots,
		home_feed.FeedWorkerOptions{
			ConsumerName: feedWorkerConsumerName(),
			Logger:       logger,
		},
	)

	logger.Info("starting home feed worker")
	if err := worker.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("home feed worker stopped", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func feedWorkerConsumerName() string {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		return "feed-worker"
	}
	return hostname
}
