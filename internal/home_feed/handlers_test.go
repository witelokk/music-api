package home_feed

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/witelokk/music-api/internal/auth"
	"github.com/witelokk/music-api/internal/favorites"
	"github.com/witelokk/music-api/internal/followings"
	openapi "github.com/witelokk/music-api/internal/openapi"
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
	favoritesRepo := &fakeFavoritesRepo{
		songs: []favorites.FavoriteSong{
			{
				ID:              "00000000-0000-0000-0000-000000000004",
				Name:            "Favorite Song 1",
				DurationSeconds: 180,
				StreamMediaID:   "stream-media-id",
			},
		},
	}
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
	service := NewService(favoritesRepo, playlistsRepo, followingsRepo, releasesRepo)

	ctx := auth.WithUserID(context.Background(), "user-id")

	resp, err := HandleGetHomeFeed(ctx, service, logger, openapi.GetHomeFeedRequestObject{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	okResp, ok := resp.(openapi.GetHomeFeed200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}

	if len(okResp.Sections) < 3 {
		t.Fatalf("expected at least 3 sections, got %d", len(okResp.Sections))
	}
	if okResp.Sections[0].Titles["en"] == "" || okResp.Sections[0].Titles["ru"] == "" {
		t.Fatalf("expected localized titles, got %+v", okResp.Sections[0].Titles)
	}
	if len(okResp.Sections[0].Items) < 2 {
		t.Fatalf("expected favorites and playlist items in first section, got %+v", okResp.Sections[0].Items)
	}
	if okResp.Sections[0].Items[0].Type != openapi.HomeFeedItemTypeFavorites {
		t.Fatalf("expected first section to start with favorites item, got %+v", okResp.Sections[0].Items[0])
	}
	if okResp.Sections[0].Items[1].Type != openapi.HomeFeedItemTypePlaylist {
		t.Fatalf("expected first section to include playlist item, got %+v", okResp.Sections[0].Items[1])
	}
	if len(okResp.Sections[1].Items) != 1 || okResp.Sections[1].Items[0].Type != openapi.HomeFeedItemTypeArtist {
		t.Fatalf("expected second section to contain followed artist item, got %+v", okResp.Sections[1].Items)
	}

	var foundRelease bool
	for _, section := range okResp.Sections {
		for _, item := range section.Items {
			if item.Type != openapi.HomeFeedItemTypeRelease || item.Release == nil {
				continue
			}
			foundRelease = true
			if item.Release.Name != "Release 1" {
				t.Fatalf("expected release name %q, got %q", "Release 1", item.Release.Name)
			}
			if item.Release.ReleasedAt != "2024-03-15" {
				t.Fatalf("expected release date %q, got %q", "2024-03-15", item.Release.ReleasedAt)
			}
		}
	}
	if !foundRelease {
		t.Fatalf("expected release item in first section, got %+v", okResp.Sections[0].Items)
	}
}
