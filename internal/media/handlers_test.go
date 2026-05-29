package media

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/witelokk/music-api/internal/openapi"
	"github.com/witelokk/music-api/internal/requestctx"
)

func TestGetMediaReturnsConfiguredMediaResponse(t *testing.T) {
	id := uuid.New()
	service := NewMediaService(&fakeStorage{
		reader: io.NopCloser(strings.NewReader("image")),
		size:   5,
		mime:   "image/jpeg",
	})

	resp, err := GetMedia(context.Background(), service, openapi.GetMediaRequestObject{
		Id: openapi_types.UUID(id),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	okResp, ok := resp.(openapi.GetMedia200ImagejpegResponse)
	if !ok {
		t.Fatalf("expected image/jpeg response, got %T", resp)
	}
	if okResp.ContentLength != 5 {
		t.Fatalf("expected content length 5, got %d", okResp.ContentLength)
	}
}

func TestGetMediaReturnsNotFound(t *testing.T) {
	id := uuid.New()
	service := NewMediaService(&fakeStorage{err: ErrMediaNotFound})

	resp, err := GetMedia(context.Background(), service, openapi.GetMediaRequestObject{
		Id: openapi_types.UUID(id),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errResp, ok := resp.(openapi.GetMedia404JSONResponse)
	if !ok {
		t.Fatalf("expected 404 response, got %T", resp)
	}
	if errResp.Error != "media not found" {
		t.Fatalf("expected error %q, got %q", "media not found", errResp.Error)
	}
}

func TestGetMediaReturnsNotConfigured(t *testing.T) {
	resp, err := GetMedia(context.Background(), NewMediaService(nil), openapi.GetMediaRequestObject{
		Id: openapi_types.UUID(uuid.New()),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errResp, ok := resp.(openapi.GetMedia500JSONResponse)
	if !ok {
		t.Fatalf("expected 500 response, got %T", resp)
	}
	if errResp.Error != "media service not configured" {
		t.Fatalf("expected configuration error, got %q", errResp.Error)
	}
}

func TestGetMediaReturnsInternalError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx := requestctx.WithLogger(context.Background(), logger)
	service := NewMediaService(&fakeStorage{err: io.ErrUnexpectedEOF})

	resp, err := GetMedia(ctx, service, openapi.GetMediaRequestObject{
		Id: openapi_types.UUID(uuid.New()),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errResp, ok := resp.(openapi.GetMedia500JSONResponse)
	if !ok {
		t.Fatalf("expected 500 response, got %T", resp)
	}
	if errResp.Error != "failed to fetch media" {
		t.Fatalf("expected fetch error, got %q", errResp.Error)
	}
}
