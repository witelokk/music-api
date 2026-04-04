package requestctx

import (
	"context"
	"log/slog"
)

type requestIDContextKey struct{}

type loggerContextKey struct{}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDContextKey{}, requestID)
}

func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	val := ctx.Value(requestIDContextKey{})
	if val == nil {
		return ""
	}

	id, ok := val.(string)
	if !ok {
		return ""
	}

	return id
}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, logger)
}

func LoggerFromContext(ctx context.Context, base *slog.Logger) *slog.Logger {
	if ctx == nil {
		return base
	}

	val := ctx.Value(loggerContextKey{})
	if val == nil {
		return base
	}

	l, ok := val.(*slog.Logger)
	if !ok || l == nil {
		return base
	}

	return l
}

