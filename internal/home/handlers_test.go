package home

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/witelokk/music-api/internal/auth"
	openapi "github.com/witelokk/music-api/internal/openapi"
	"github.com/witelokk/music-api/internal/followings"
	"github.com/witelokk/music-api/internal/playlists"
	"github.com/witelokk/music-api/internal/releases"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

func TestHandleGetHomeFeed_NoUserID(t *testing.T) {
	logger := newTestLogger()

	resp, err := HandleGetHomeFeed(context.Background(), nil, logger, openapi.GetHomeFeedRequestObject{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := resp.(openapi.GetHomeFeed500JSONResponse); !ok {
		t.Fatalf("expected 500 response, got %T", resp)
	}
}

func TestHandleGetHomeFeed_OK(t *testing.T) {
	logger := newTestLogger()
	playlistsRepo := &fakePlaylistsRepo{
		playlists: []playlists.PlaylistSummary{
			{ID: "00000000-0000-0000-0000-000000000001", Name: "Playlist 1"},
		},
	}
	followingsRepo := &fakeFollowingsRepo{
		artists: []followings.FollowedArtist{
			{ID: "00000000-0000-0000-0000-000000000002", Name: "Artist 1"},
		},
	}
	releasesRepo := &fakeReleasesRepo{
		releases: []releases.Release{
			{
				ID:        "00000000-0000-0000-0000-000000000003",
				Name:      "Release 1",
				ReleaseAt: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	service := NewService(playlistsRepo, followingsRepo, releasesRepo)

	ctx := auth.WithUserID(context.Background(), "user-id")

	resp, err := HandleGetHomeFeed(ctx, service, logger, openapi.GetHomeFeedRequestObject{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	okResp, ok := resp.(openapi.GetHomeFeed200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}

	if okResp.Playlists.Count != 1 {
		t.Fatalf("expected 1 playlist, got %d", okResp.Playlists.Count)
	}
	if okResp.FollowedArtists.Count != 1 {
		t.Fatalf("expected 1 followed artist, got %d", okResp.FollowedArtists.Count)
	}
	if len(okResp.Sections) == 0 {
		t.Fatalf("expected at least 1 section, got %d", len(okResp.Sections))
	}
	if okResp.Sections[0].Titles["en"] == "" || okResp.Sections[0].Titles["ru"] == "" {
		t.Fatalf("expected localized titles, got %+v", okResp.Sections[0].Titles)
	}
	if okResp.Sections[0].Releases.Count != 1 {
		t.Fatalf("expected 1 release in first section, got %d", okResp.Sections[0].Releases.Count)
	}
	firstRelease := okResp.Sections[0].Releases.Releases[0]
	if firstRelease.Name != "Release 1" {
		t.Fatalf("expected release name %q, got %q", "Release 1", firstRelease.Name)
	}
	if firstRelease.ReleasedAt != "2024-03-15" {
		t.Fatalf("expected release date %q, got %q", "2024-03-15", firstRelease.ReleasedAt)
	}
}
