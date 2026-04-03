package main

import (
	"log/slog"
	"net/http"

	"github.com/witelokk/music-api/internal"
)

func main() {
	config := internal.MustLoadConfig()
	logger := internal.NewLogger(config.Logger.Type, config.Logger.Level)

	router := http.NewServeMux()
	internal.AddRoutes(router, config, logger)

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
