package home

import (
	"context"
	"testing"
	"time"

	"github.com/witelokk/music-api/internal/followings"
	"github.com/witelokk/music-api/internal/playlists"
	"github.com/witelokk/music-api/internal/releases"
)

type fakePlaylistsRepo struct {
	playlists []playlists.PlaylistSummary
	err       error
}

func (r *fakePlaylistsRepo) CreatePlaylist(ctx context.Context, userID, name string) (string, error) {
	return "", nil
}

func (r *fakePlaylistsRepo) UpdatePlaylist(ctx context.Context, userID, playlistID, name string) error {
	return nil
}

func (r *fakePlaylistsRepo) DeletePlaylist(ctx context.Context, userID, playlistID string) error {
	return nil
}

func (r *fakePlaylistsRepo) GetPlaylists(ctx context.Context, userID string) ([]playlists.PlaylistSummary, error) {
	return r.playlists, r.err
}

func (r *fakePlaylistsRepo) GetPlaylist(ctx context.Context, userID, playlistID string) (*playlists.Playlist, error) {
	return nil, nil
}

func (r *fakePlaylistsRepo) GetPlaylistSongs(ctx context.Context, userID, playlistID string) ([]playlists.PlaylistSong, error) {
	return nil, nil
}

func (r *fakePlaylistsRepo) AddSongToPlaylist(ctx context.Context, userID, playlistID, songID string) error {
	return nil
}

func (r *fakePlaylistsRepo) RemoveSongFromPlaylist(ctx context.Context, userID, playlistID, songID string) error {
	return nil
}

type fakeFollowingsRepo struct {
	artists []followings.FollowedArtist
	err     error
}

func (r *fakeFollowingsRepo) Follow(ctx context.Context, userID, artistID string) error {
	return nil
}

func (r *fakeFollowingsRepo) Unfollow(ctx context.Context, userID, artistID string) error {
	return nil
}

func (r *fakeFollowingsRepo) GetFollowedArtists(ctx context.Context, userID string) ([]followings.FollowedArtist, error) {
	return r.artists, r.err
}

type fakeReleasesRepo struct {
	releases []releases.Release
	err      error
}

func (r *fakeReleasesRepo) GetReleaseByID(ctx context.Context, id string) (*releases.Release, error) {
	return nil, nil
}

func (r *fakeReleasesRepo) GetRandomReleases(ctx context.Context, seed string, limit int) ([]releases.Release, error) {
	return r.releases, r.err
}

func TestService_GetHomeScreenLayout_UsesSeededRepos(t *testing.T) {
	playlistsRepo := &fakePlaylistsRepo{
		playlists: []playlists.PlaylistSummary{
			{ID: "p1", Name: "Playlist 1"},
		},
	}
	followingsRepo := &fakeFollowingsRepo{
		artists: []followings.FollowedArtist{
			{ID: "a1", Name: "Artist 1"},
		},
	}
	releasesRepo := &fakeReleasesRepo{
		releases: []releases.Release{
			{ID: "r1", Name: "Release 1"},
			{ID: "r2", Name: "Release 2"},
		},
	}

	service := NewService(playlistsRepo, followingsRepo, releasesRepo)

	now := time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)
	layout, err := service.GetHomeScreenLayout(context.Background(), "user-id", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(layout.Playlists) != 1 || layout.Playlists[0].ID != "p1" {
		t.Fatalf("unexpected playlists in layout: %+v", layout.Playlists)
	}
	if len(layout.FollowedArtists) != 1 || layout.FollowedArtists[0].ID != "a1" {
		t.Fatalf("unexpected followed artists in layout: %+v", layout.FollowedArtists)
	}
	if len(layout.Sections) == 0 {
		t.Fatalf("expected at least one section, got 0")
	}
	// Each section should have some releases drawn from the fake repo.
	for _, sec := range layout.Sections {
		if len(sec.Releases) == 0 {
			t.Fatalf("expected section %q to have releases", sec.Titles["en"])
		}
	}
}
