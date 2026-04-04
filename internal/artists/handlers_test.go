package artists

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

type fakeArtistsService struct {
	artist *Artist
	err    error
}

func (s *fakeArtistsService) GetArtist(ctx context.Context, id string) (*Artist, error) {
	return s.artist, s.err
}

func TestHandleGetArtist_NotFound(t *testing.T) {
	logger := newTestLogger()
	id := uuid.New()
	req := openapi.GetArtistRequestObject{Id: id}

	svcWrapper := &ArtistsService{repo: &fakeArtistsRepo{err: ErrArtistNotFound}}

	resp, err := HandleGetArtist(context.Background(), svcWrapper, logger, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errResp, ok := resp.(openapi.GetArtist404JSONResponse)
	if !ok {
		t.Fatalf("expected 404 response, got %T", resp)
	}
	if errResp.Error != "artist not found" {
		t.Fatalf("expected error %q, got %q", "artist not found", errResp.Error)
	}
}
