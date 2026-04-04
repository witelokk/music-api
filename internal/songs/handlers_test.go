package songs

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	openapi "github.com/witelokk/music-api/internal/openapi"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

func TestHandleGetSong_NotFound(t *testing.T) {
	logger := newTestLogger()
	id := uuid.New()
	req := openapi.GetSongRequestObject{Id: id}

	svc := &Service{repo: &fakeSongsRepo{err: ErrSongNotFound}}

	resp, err := HandleGetSong(context.Background(), svc, logger, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errResp, ok := resp.(openapi.GetSong404JSONResponse)
	if !ok {
		t.Fatalf("expected 404 response, got %T", resp)
	}
	if errResp.Error != "song not found" {
		t.Fatalf("expected error %q, got %q", "song not found", errResp.Error)
	}
}

func TestHandleGetSong_Success(t *testing.T) {
	logger := newTestLogger()
	id := uuid.New()

	song := &Song{
		ID:              id.String(),
		Name:            "Test Song",
		DurationSeconds: 120,
		StreamURL:       "https://example.com/stream",
	}

	svc := &Service{repo: &fakeSongsRepo{song: song}}

	req := openapi.GetSongRequestObject{Id: id}

	resp, err := HandleGetSong(context.Background(), svc, logger, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	okResp, ok := resp.(openapi.GetSong200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}
	if okResp.Name != song.Name {
		t.Fatalf("expected name %q, got %q", song.Name, okResp.Name)
	}
}

