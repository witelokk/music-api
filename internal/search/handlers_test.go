package search

import (
	"context"
	"io"
	"log/slog"
	"testing"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/google/uuid"
	"github.com/witelokk/music-api/internal/auth"
	openapi "github.com/witelokk/music-api/internal/openapi"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

func TestHandleSearch_MissingQuery(t *testing.T) {
	logger := newTestLogger()
	repo := &fakeSearchRepo{}
	svc := NewService(repo)

	req := openapi.SearchRequestObject{
		Params: openapi.SearchParams{},
	}

	resp, err := HandleSearch(context.Background(), svc, logger, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := resp.(openapi.Search400JSONResponse); !ok {
		t.Fatalf("expected 400 response, got %T", resp)
	}
}

func TestHandleSearch_Success(t *testing.T) {
	logger := newTestLogger()

	songID := uuid.New().String()
	artistID := uuid.New().String()
	releaseID := uuid.New().String()
	playlistID := uuid.New().String()

	repo := &fakeSearchRepo{
		songs: []SongResult{
			{
				ID:   songID,
				Name: "Song",
				Artists: []ArtistSummary{
					{ID: artistID, Name: "Artist"},
				},
			},
		},
		artists: []ArtistResult{
			{ID: artistID, Name: "Artist"},
		},
		releases: []ReleaseResult{
			{ID: releaseID, Name: "Release"},
		},
		playlists: []PlaylistResult{
			{ID: playlistID, Name: "Playlist"},
		},
	}
	svc := NewService(repo)

	ctx := auth.WithUserID(context.Background(), "user-id")

	req := openapi.SearchRequestObject{
		Params: openapi.SearchParams{
			Q: "Song",
		},
	}

	resp, err := HandleSearch(ctx, svc, logger, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	okResp, ok := resp.(openapi.Search200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}

	if okResp.Query != "Song" {
		t.Fatalf("expected Query=Song, got %s", okResp.Query)
	}
	if okResp.Total != len(okResp.Results) {
		t.Fatalf("expected Total=%d, got %d", len(okResp.Results), okResp.Total)
	}
	if len(okResp.Results) == 0 {
		t.Fatalf("expected non-empty results")
	}

	foundSong := false
	for _, item := range okResp.Results {
		if item.Song != nil && item.Song.Id == openapi_types.UUID(uuid.MustParse(songID)) {
			foundSong = true
		}
	}
	if !foundSong {
		t.Fatalf("expected song result with id %s", songID)
	}
}

