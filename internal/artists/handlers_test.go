package artists

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/witelokk/music-api/internal/auth"
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

func TestHandleGetArtist_Success_PreservesFavoriteState(t *testing.T) {
	logger := newTestLogger()
	artistID := uuid.New()
	songID := uuid.New()
	artist := &Artist{
		ID:   artistID.String(),
		Name: "Artist",
		Popular: []PopularSong{
			{
				ID:              songID.String(),
				Name:            "Popular Song",
				DurationSeconds: 180,
				StreamMediaID:   "stream-id",
				IsFavorite:      true,
				Artists: []ArtistSummary{
					{
						ID:   artistID.String(),
						Name: "Artist",
					},
				},
			},
		},
	}

	svcWrapper := &ArtistsService{repo: &fakeArtistsRepo{artist: artist}}
	req := openapi.GetArtistRequestObject{Id: artistID}

	resp, err := HandleGetArtist(auth.WithUserID(context.Background(), "user-id"), svcWrapper, logger, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	okResp, ok := resp.(openapi.GetArtist200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}
	if len(okResp.PopularSongs.Songs) != 1 {
		t.Fatalf("expected 1 popular song, got %d", len(okResp.PopularSongs.Songs))
	}
	if !okResp.PopularSongs.Songs[0].IsFavorite {
		t.Fatalf("expected popular song to be favorite")
	}
}
