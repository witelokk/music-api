package internal

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/witelokk/music-api/internal/auth"
	openapi "github.com/witelokk/music-api/internal/openapi"
)

type HTTPHandlerConfig struct {
	JWTSecret string
}

func NewHTTPHandler(
	serverImpl openapi.StrictServerInterface,
	cfg HTTPHandlerConfig,
	logger *slog.Logger,
) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /openapi.yml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "openapi.yml")
	})

	strictHandler := openapi.NewStrictHandlerWithOptions(
		serverImpl,
		[]openapi.StrictMiddlewareFunc{
			auth.NewJWTMiddleware(cfg.JWTSecret, logger),
		},
		openapi.StrictHTTPServerOptions{
			RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
				http.Error(w, err.Error(), http.StatusBadRequest)
			},
			ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
				if errors.Is(err, auth.ErrUnauthorized) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					_ = json.NewEncoder(w).Encode(openapi.Error{Error: "unauthorized"})
					return
				}

				http.Error(w, err.Error(), http.StatusInternalServerError)
			},
		},
	)

	handler := openapi.HandlerFromMux(strictHandler, mux)

	corsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		handler.ServeHTTP(w, r)
	})

	return NewHTTPLoggingMiddleware(logger)(corsHandler)
}
