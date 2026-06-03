package home_feed

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/witelokk/music-api/internal/favorites"
	"github.com/witelokk/music-api/internal/followings"
	"github.com/witelokk/music-api/internal/playlists"
	"github.com/witelokk/music-api/internal/releases"
)

type fakeFavoritesRepo struct {
	songs []favorites.FavoriteSong
	err   error
}

func (r *fakeFavoritesRepo) AddFavorite(ctx context.Context, userID, songID string) error {
	return nil
}

func (r *fakeFavoritesRepo) RemoveFavorite(ctx context.Context, userID, songID string) error {
	return nil
}

func (r *fakeFavoritesRepo) GetFavoriteSongs(ctx context.Context, userID string) ([]favorites.FavoriteSong, error) {
	return r.songs, r.err
}

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
	releases               []releases.Release
	recentReleases         []releases.Release
	followedArtistReleases []releases.Release
	favoriteSongReleases   []releases.Release
	err                    error
}

func (r *fakeReleasesRepo) GetReleaseByID(ctx context.Context, userID, id string) (*releases.Release, error) {
	return nil, nil
}

func (r *fakeReleasesRepo) GetRandomReleases(ctx context.Context, seed string, limit int) ([]releases.Release, error) {
	return r.releases, r.err
}

func (r *fakeReleasesRepo) GetRecentReleases(ctx context.Context, limit int) ([]releases.Release, error) {
	if r.recentReleases != nil {
		return r.recentReleases, r.err
	}
	return r.releases, r.err
}

func (r *fakeReleasesRepo) GetReleasesByFollowedArtistSeeds(ctx context.Context, artistIDs []string, seed string, limit int) ([]releases.Release, error) {
	return r.followedArtistReleases, r.err
}

func (r *fakeReleasesRepo) GetReleasesByFavoriteSongSeeds(ctx context.Context, songIDs []string, seed string, limit int) ([]releases.Release, error) {
	return r.favoriteSongReleases, r.err
}

type fakeFeedSnapshotReader struct {
	snapshot *FeedSnapshot
	err      error
}

func (r *fakeFeedSnapshotReader) Get(ctx context.Context, userID string) (*FeedSnapshot, error) {
	return r.snapshot, r.err
}

type fakeServiceFeedRefreshQueue struct {
	calls  int
	userID string
	reason string
}

func (q *fakeServiceFeedRefreshQueue) Enqueue(ctx context.Context, userID, reason string) error {
	q.calls++
	q.userID = userID
	q.reason = reason
	return nil
}

func TestService_GetHomeFeed_UsesSeededRepos(t *testing.T) {
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
	favoritesRepo := &fakeFavoritesRepo{
		songs: []favorites.FavoriteSong{
			{ID: "s1", Name: "Song 1", StreamMediaID: "stream-1", DurationSeconds: 180},
		},
	}

	service := NewService(favoritesRepo, playlistsRepo, followingsRepo, releasesRepo)

	now := time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)
	layout, err := service.GetHomeFeed(context.Background(), "user-id", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(layout.FavoriteSongs) != 1 || layout.FavoriteSongs[0].ID != "s1" {
		t.Fatalf("unexpected favorite songs in layout: %+v", layout.FavoriteSongs)
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
	if len(layout.Sections) < 3 {
		t.Fatalf("expected live playlist, live artist, and generated sections, got %d", len(layout.Sections))
	}
	if layout.Sections[0].Items[0].Type != ItemTypeFavorites {
		t.Fatalf("expected first section to start with favorites item, got %+v", layout.Sections[0].Items[0])
	}
	if layout.Sections[0].Items[1].Type != ItemTypePlaylist {
		t.Fatalf("expected first section to contain playlist item, got %+v", layout.Sections[0].Items[1])
	}
	if layout.Sections[1].Items[0].Type != ItemTypeArtist {
		t.Fatalf("expected second section to contain followed artist item, got %+v", layout.Sections[1].Items[0])
	}
}

func TestService_GetHomeFeed_ReturnsFreshSnapshot(t *testing.T) {
	now := time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)
	snapshotLayout := &Layout{
		Playlists: []playlists.PlaylistSummary{
			{ID: "snapshot-playlist", Name: "Snapshot Playlist"},
		},
		FollowedArtists: []followings.FollowedArtist{
			{ID: "snapshot-artist", Name: "Snapshot Artist"},
		},
		Sections: []Section{
			{Titles: map[string]string{"en": "Discover Releases"}, Items: []Item{{Type: ItemTypeRelease, Release: &releases.Release{ID: "r1", Name: "Release 1"}}}},
		},
	}
	queue := &fakeServiceFeedRefreshQueue{}
	service := NewCachedService(
		&fakeFavoritesRepo{songs: []favorites.FavoriteSong{{ID: "live-song", Name: "Live Song", StreamMediaID: "stream-1", DurationSeconds: 180}}},
		&fakePlaylistsRepo{playlists: []playlists.PlaylistSummary{{ID: "live-playlist", Name: "Live Playlist"}}},
		&fakeFollowingsRepo{artists: []followings.FollowedArtist{{ID: "live-artist", Name: "Live Artist"}}},
		&fakeReleasesRepo{err: errors.New("releases should not be fetched")},
		&fakeFeedSnapshotReader{snapshot: &FeedSnapshot{
			UserID:      "user-id",
			Layout:      snapshotLayout,
			GeneratedAt: now.Add(-time.Minute),
			Version:     "v1",
		}},
		queue,
		time.Hour,
	)

	layout, err := service.GetHomeFeed(context.Background(), "user-id", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(layout.FavoriteSongs) != 1 || layout.FavoriteSongs[0].ID != "live-song" {
		t.Fatalf("expected live favorite songs, got %+v", layout.FavoriteSongs)
	}
	if len(layout.Playlists) != 1 || layout.Playlists[0].ID != "live-playlist" {
		t.Fatalf("expected live playlists, got %+v", layout.Playlists)
	}
	if len(layout.FollowedArtists) != 1 || layout.FollowedArtists[0].ID != "live-artist" {
		t.Fatalf("expected live followed artists, got %+v", layout.FollowedArtists)
	}
	if len(layout.Sections) != 3 {
		t.Fatalf("expected two live sections plus generated snapshot section, got %d", len(layout.Sections))
	}
	if layout.Sections[0].Items[0].Type != ItemTypeFavorites {
		t.Fatalf("expected first section to start with favorites item, got %+v", layout.Sections[0].Items[0])
	}
	if layout.Sections[1].Items[0].Type != ItemTypeArtist {
		t.Fatalf("expected second section to contain followed artist item, got %+v", layout.Sections[1].Items[0])
	}
	if layout.Sections[2].Items[0].Release == nil || layout.Sections[2].Items[0].Release.ID != "r1" {
		t.Fatalf("expected release section to remain from snapshot, got %+v", layout.Sections[2])
	}
	if queue.calls != 0 {
		t.Fatalf("expected no refresh enqueue, got %d calls", queue.calls)
	}
}

func TestService_GetHomeFeed_ReturnsStaleSnapshotAndEnqueuesRefresh(t *testing.T) {
	now := time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)
	snapshotLayout := &Layout{
		Playlists: []playlists.PlaylistSummary{
			{ID: "stale-playlist", Name: "Stale Playlist"},
		},
		Sections: []Section{
			{Titles: map[string]string{"en": "Discover Releases"}, Items: []Item{{Type: ItemTypeRelease, Release: &releases.Release{ID: "r1", Name: "Release 1"}}}},
		},
	}
	queue := &fakeServiceFeedRefreshQueue{}
	service := NewCachedService(
		&fakeFavoritesRepo{songs: []favorites.FavoriteSong{{ID: "live-song", Name: "Live Song", StreamMediaID: "stream-1", DurationSeconds: 180}}},
		&fakePlaylistsRepo{playlists: []playlists.PlaylistSummary{{ID: "live-playlist", Name: "Live Playlist"}}},
		&fakeFollowingsRepo{},
		&fakeReleasesRepo{err: errors.New("releases should not be fetched")},
		&fakeFeedSnapshotReader{snapshot: &FeedSnapshot{
			UserID:      "user-id",
			Layout:      snapshotLayout,
			GeneratedAt: now.Add(-2 * time.Hour),
			Version:     "v1",
		}},
		queue,
		time.Hour,
	)

	layout, err := service.GetHomeFeed(context.Background(), "user-id", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(layout.FavoriteSongs) != 1 || layout.FavoriteSongs[0].ID != "live-song" {
		t.Fatalf("expected stale snapshot to include live favorite songs, got %+v", layout.FavoriteSongs)
	}
	if len(layout.Playlists) != 1 || layout.Playlists[0].ID != "live-playlist" {
		t.Fatalf("expected stale snapshot to include live playlists, got %+v", layout.Playlists)
	}
	if len(layout.Sections) != 2 {
		t.Fatalf("expected playlist live section plus generated snapshot section, got %d", len(layout.Sections))
	}
	if layout.Sections[0].Items[0].Type != ItemTypeFavorites {
		t.Fatalf("expected first section to start with favorites item, got %+v", layout.Sections[0].Items[0])
	}
	if layout.Sections[1].Items[0].Release == nil || layout.Sections[1].Items[0].Release.ID != "r1" {
		t.Fatalf("expected generated section to remain from stale snapshot, got %+v", layout.Sections[1])
	}
	if queue.calls != 1 {
		t.Fatalf("expected one refresh enqueue, got %d calls", queue.calls)
	}
	if queue.userID != "user-id" || queue.reason != "stale_home_feed" {
		t.Fatalf("unexpected enqueue call: user=%q reason=%q", queue.userID, queue.reason)
	}
}

func TestService_GetHomeFeed_AlwaysShowsFavoritesShortcutAndOmitsEmptyFollowedArtists(t *testing.T) {
	now := time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)
	service := NewService(
		&fakeFavoritesRepo{},
		&fakePlaylistsRepo{},
		&fakeFollowingsRepo{},
		&fakeReleasesRepo{releases: []releases.Release{{ID: "r1", Name: "Release 1"}}},
	)

	layout, err := service.GetHomeFeed(context.Background(), "user-id", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(layout.Sections) != 2 {
		t.Fatalf("expected favorites shortcut and generated sections, got %d sections: %+v", len(layout.Sections), layout.Sections)
	}
	if layout.Sections[0].Items[0].Type != ItemTypeFavorites {
		t.Fatalf("expected favorites shortcut, got %+v", layout.Sections[0])
	}
	if layout.Sections[1].Items[0].Type != ItemTypeRelease {
		t.Fatalf("expected generated release section, got %+v", layout.Sections[1])
	}
}

func TestLiveSections_LimitItemsToTen(t *testing.T) {
	playlistsRows := make([]playlists.PlaylistSummary, 12)
	for i := range playlistsRows {
		playlistsRows[i] = playlists.PlaylistSummary{ID: "playlist-id", Name: "Playlist"}
	}
	favoriteSongs := []favorites.FavoriteSong{{ID: "song-id", Name: "Song"}}

	playlistSection, ok := livePlaylistsSection(favoriteSongs, playlistsRows)
	if !ok {
		t.Fatalf("expected playlist section")
	}
	if len(playlistSection.Items) != 10 {
		t.Fatalf("expected 10 playlist section items, got %d", len(playlistSection.Items))
	}
	if playlistSection.Items[0].Type != ItemTypeFavorites {
		t.Fatalf("expected favorites item first, got %+v", playlistSection.Items[0])
	}

	artistsRows := make([]followings.FollowedArtist, 12)
	for i := range artistsRows {
		artistsRows[i] = followings.FollowedArtist{ID: "artist-id", Name: "Artist"}
	}

	artistsSection, ok := liveFollowedArtistsSection(artistsRows)
	if !ok {
		t.Fatalf("expected followed artists section")
	}
	if len(artistsSection.Items) != 10 {
		t.Fatalf("expected 10 followed artist section items, got %d", len(artistsSection.Items))
	}
}

func TestService_GetHomeFeed_FallsBackWhenSnapshotMissing(t *testing.T) {
	now := time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)
	queue := &fakeServiceFeedRefreshQueue{}
	service := NewCachedService(
		&fakeFavoritesRepo{songs: []favorites.FavoriteSong{{ID: "s1", Name: "Song 1", StreamMediaID: "stream-1", DurationSeconds: 180}}},
		&fakePlaylistsRepo{playlists: []playlists.PlaylistSummary{{ID: "p1", Name: "Playlist 1"}}},
		&fakeFollowingsRepo{},
		&fakeReleasesRepo{releases: []releases.Release{{ID: "r1", Name: "Release 1"}}},
		&fakeFeedSnapshotReader{err: ErrFeedSnapshotNotFound},
		queue,
		time.Hour,
	)

	layout, err := service.GetHomeFeed(context.Background(), "user-id", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(layout.Playlists) != 1 || layout.Playlists[0].ID != "p1" {
		t.Fatalf("expected fallback layout from repos, got %+v", layout.Playlists)
	}
	if queue.calls != 1 {
		t.Fatalf("expected one refresh enqueue, got %d calls", queue.calls)
	}
	if queue.userID != "user-id" || queue.reason != "missing_home_feed" {
		t.Fatalf("unexpected enqueue call: user=%q reason=%q", queue.userID, queue.reason)
	}
}
