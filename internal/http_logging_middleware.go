package internal

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/witelokk/music-api/internal/requestctx"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
	body   bytes.Buffer
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
	if n > 0 {
		_, _ = r.body.Write(b[:n])
	}
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

			reqLogger.Info("request received",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			)

			isDebug := reqLogger.Enabled(ctx, slog.LevelDebug)

			var requestBody []byte
			if isDebug && r.Body != nil {
				bodyBytes, err := io.ReadAll(r.Body)
				if err == nil {
					requestBody = bodyBytes
					r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				} else {
					reqLogger.Debug("failed to read request body",
						slog.String("error", err.Error()),
					)
				}
			}

			recorder := &statusRecorder{ResponseWriter: w}

			next.ServeHTTP(recorder, r.WithContext(ctx))

			status := recorder.status
			if status == 0 {
				status = http.StatusOK
			}

			if isDebug {
				const maxBodyLogBytes = 2048

				headers := make(map[string][]string, len(r.Header))
				for k, v := range r.Header {
					if strings.EqualFold(k, "Authorization") {
						headers[k] = []string{"[REDACTED]"}
						continue
					}
					headers[k] = v
				}

				reqBodyStr := ""
				if len(requestBody) > 0 {
					if len(requestBody) > maxBodyLogBytes {
						reqBodyStr = string(requestBody[:maxBodyLogBytes]) + " [truncated]"
					} else {
						reqBodyStr = string(requestBody)
					}
				}

				respBytes := recorder.body.Bytes()
				respBodyStr := ""
				if len(respBytes) > 0 {
					if len(respBytes) > maxBodyLogBytes {
						respBodyStr = string(respBytes[:maxBodyLogBytes]) + " [truncated]"
					} else {
						respBodyStr = string(respBytes)
					}
				}

				reqLogger.Debug("http request",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.Any("headers", headers),
					slog.String("body", reqBodyStr),
				)

				reqLogger.Debug("http response",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.Int("status", status),
					slog.String("body", respBodyStr),
				)
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
