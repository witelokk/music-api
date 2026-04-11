package internal

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/witelokk/music-api/internal/requestctx"
)

func NewHTTPRecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					reqLogger := requestctx.LoggerFromContext(r.Context(), logger)

					reqLogger.Error("panic recovered",
						slog.Any("panic", rec),
						slog.String("request_id", requestctx.RequestIDFromContext(r.Context())),
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
						slog.String("stack", string(debug.Stack())),
					)

					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
