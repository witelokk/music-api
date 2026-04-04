package internal

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/witelokk/music-api/internal/requestctx"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

func NewHTTPLoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.NewString()
			}

			w.Header().Set("X-Request-ID", requestID)

			ctx := requestctx.WithRequestID(r.Context(), requestID)
			reqLogger := logger.With(slog.String("request_id", requestID))
			ctx = requestctx.WithLogger(ctx, reqLogger)

			recorder := &statusRecorder{ResponseWriter: w}

			next.ServeHTTP(recorder, r.WithContext(ctx))

			status := recorder.status
			if status == 0 {
				status = http.StatusOK
			}

			reqLogger.Info("request completed",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", status),
				slog.Int("bytes", recorder.bytes),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}
