package requestctx

import (
	"context"
	"io"
	"log/slog"
	"testing"
)

func TestRequestIDFromContext(t *testing.T) {
	ctx := WithRequestID(context.Background(), "request-id")

	if got := RequestIDFromContext(ctx); got != "request-id" {
		t.Fatalf("expected request ID %q, got %q", "request-id", got)
	}
}

func TestRequestIDFromContextMissingOrInvalid(t *testing.T) {
	if got := RequestIDFromContext(nil); got != "" {
		t.Fatalf("expected empty request ID from nil context, got %q", got)
	}

	ctx := context.WithValue(context.Background(), requestIDContextKey{}, 123)
	if got := RequestIDFromContext(ctx); got != "" {
		t.Fatalf("expected empty request ID for invalid value, got %q", got)
	}
}

func TestLoggerFromContext(t *testing.T) {
	base := slog.New(slog.NewTextHandler(io.Discard, nil))
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	ctx := WithLogger(context.Background(), logger)
	if got := LoggerFromContext(ctx, base); got != logger {
		t.Fatalf("expected logger from context")
	}
}

func TestLoggerFromContextFallsBackToBase(t *testing.T) {
	base := slog.New(slog.NewTextHandler(io.Discard, nil))

	if got := LoggerFromContext(nil, base); got != base {
		t.Fatalf("expected base logger from nil context")
	}

	ctx := context.WithValue(context.Background(), loggerContextKey{}, "not a logger")
	if got := LoggerFromContext(ctx, base); got != base {
		t.Fatalf("expected base logger for invalid value")
	}
}
