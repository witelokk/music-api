package home_feed

import (
	"context"
	"errors"
	"time"

	"github.com/witelokk/music-api/internal/favorites"
	"github.com/witelokk/music-api/internal/followings"
	"github.com/witelokk/music-api/internal/playlists"
	"github.com/witelokk/music-api/internal/releases"
)

type ItemType string

const (
	ItemTypeRelease   ItemType = "release"
	ItemTypePlaylist  ItemType = "playlist"
	ItemTypeArtist    ItemType = "artist"
	ItemTypeFavorites ItemType = "favorites"
)

type Item struct {
	Type     ItemType
	Release  *releases.Release
	Playlist *playlists.PlaylistSummary
	Artist   *followings.FollowedArtist
}

type Section struct {
	Titles map[string]string
	Items  []Item
}

type Layout struct {
	FavoriteSongs   []favorites.FavoriteSong
	Playlists       []playlists.PlaylistSummary
	FollowedArtists []followings.FollowedArtist
	Sections        []Section
}

type Service struct {
	favoritesRepo     favorites.Repository
	playlistsRepo     playlists.PlaylistsRepository
	followingsRepo    followings.FollowingsRepository
	releasesRepo      releases.ReleasesRepository
	snapshots         FeedSnapshotReader
	feedRefreshQueue  FeedRefreshQueue
	snapshotFreshness time.Duration
}

func NewService(
	favoritesRepo favorites.Repository,
	playlistsRepo playlists.PlaylistsRepository,
	followingsRepo followings.FollowingsRepository,
	releasesRepo releases.ReleasesRepository,
) *Service {
	return &Service{
		favoritesRepo:  favoritesRepo,
		playlistsRepo:  playlistsRepo,
		followingsRepo: followingsRepo,
		releasesRepo:   releasesRepo,
	}
}

func NewCachedService(
	favoritesRepo favorites.Repository,
	playlistsRepo playlists.PlaylistsRepository,
	followingsRepo followings.FollowingsRepository,
	releasesRepo releases.ReleasesRepository,
	snapshots FeedSnapshotReader,
	feedRefreshQueue FeedRefreshQueue,
	snapshotFreshness time.Duration,
) *Service {
	service := NewService(favoritesRepo, playlistsRepo, followingsRepo, releasesRepo)
	service.snapshots = snapshots
	service.feedRefreshQueue = feedRefreshQueue
	service.snapshotFreshness = snapshotFreshness
	return service
}

func (s *Service) GetHomeFeed(ctx context.Context, userID string, now time.Time) (*Layout, error) {
	if s.snapshots != nil {
		snapshot, err := s.snapshots.Get(ctx, userID)
		if err == nil {
			if snapshot.IsFresh(now, s.snapshotFreshness) {
				return s.withLiveUserCollections(ctx, userID, snapshot.Layout)
			}
			s.enqueueRefresh(ctx, userID, "stale_home_feed")
			if snapshot.Layout != nil {
				return s.withLiveUserCollections(ctx, userID, snapshot.Layout)
			}
		} else if !errors.Is(err, ErrFeedSnapshotNotFound) {
			return nil, err
		} else {
			s.enqueueRefresh(ctx, userID, "missing_home_feed")
		}
	}

	return s.buildHomeFeed(ctx, userID, now)
}

func (s *Service) buildHomeFeed(ctx context.Context, userID string, now time.Time) (*Layout, error) {
	return buildHomeFeed(ctx, userID, now, s.favoritesRepo, s.playlistsRepo, s.followingsRepo, s.releasesRepo)
}

func (s *Service) withLiveUserCollections(ctx context.Context, userID string, layout *Layout) (*Layout, error) {
	if layout == nil {
		return nil, nil
	}

	favoriteSongs, err := s.favoritesRepo.GetFavoriteSongs(ctx, userID)
	if err != nil {
		return nil, err
	}

	playlistsRows, err := s.playlistsRepo.GetPlaylists(ctx, userID)
	if err != nil {
		return nil, err
	}

	followedArtistsRows, err := s.followingsRepo.GetFollowedArtists(ctx, userID)
	if err != nil {
		return nil, err
	}

	liveLayout := *layout
	liveLayout.FavoriteSongs = favoriteSongs
	liveLayout.Playlists = playlistsRows
	liveLayout.FollowedArtists = followedArtistsRows
	liveLayout.Sections = appendLiveSections(favoriteSongs, playlistsRows, followedArtistsRows, layout.Sections)
	return &liveLayout, nil
}

func buildHomeFeed(
	ctx context.Context,
	userID string,
	now time.Time,
	favoritesRepo favorites.Repository,
	playlistsRepo playlists.PlaylistsRepository,
	followingsRepo followings.FollowingsRepository,
	releasesRepo releases.ReleasesRepository,
) (*Layout, error) {
	now = now.UTC()

	favoriteSongs, err := favoritesRepo.GetFavoriteSongs(ctx, userID)
	if err != nil {
		return nil, err
	}

	playlistsRows, err := playlistsRepo.GetPlaylists(ctx, userID)
	if err != nil {
		return nil, err
	}

	followedArtistsRows, err := followingsRepo.GetFollowedArtists(ctx, userID)
	if err != nil {
		return nil, err
	}

	releaseSections, err := NewSeededReleasesSectionProvider(
		releasesRepo,
		"fallback_discovery",
		map[string]string{"en": "Discover Releases", "ru": "Откройте релизы"},
	).Sections(ctx, FeedRequest{
		UserID: userID,
		Now:    now,
		Limit:  50,
	})
	if err != nil {
		return nil, err
	}

	return &Layout{
		FavoriteSongs:   favoriteSongs,
		Playlists:       playlistsRows,
		FollowedArtists: followedArtistsRows,
		Sections:        appendLiveSections(favoriteSongs, playlistsRows, followedArtistsRows, releaseSections),
	}, nil
}

func appendLiveSections(
	favoriteSongs []favorites.FavoriteSong,
	playlistsRows []playlists.PlaylistSummary,
	followedArtistsRows []followings.FollowedArtist,
	generatedSections []Section,
) []Section {
	sections := make([]Section, 0, len(generatedSections)+2)
	if section, ok := livePlaylistsSection(favoriteSongs, playlistsRows); ok {
		sections = append(sections, section)
	}
	if section, ok := liveFollowedArtistsSection(followedArtistsRows); ok {
		sections = append(sections, section)
	}
	sections = append(sections, generatedSections...)
	return sections
}

func livePlaylistsSection(favoriteSongs []favorites.FavoriteSong, rows []playlists.PlaylistSummary) (Section, bool) {
	limit := defaultSectionItemLimit
	items := make([]Item, 0, limit)
	items = append(items, Item{Type: ItemTypeFavorites})
	playlistLimit := limit - len(items)
	if len(rows) < playlistLimit {
		playlistLimit = len(rows)
	}
	for i := 0; i < playlistLimit; i++ {
		items = append(items, Item{
			Type:     ItemTypePlaylist,
			Playlist: &rows[i],
		})
	}

	return Section{
		Titles: map[string]string{"en": "Playlists", "ru": "Плейлисты"},
		Items:  items,
	}, true
}

func liveFollowedArtistsSection(rows []followings.FollowedArtist) (Section, bool) {
	if len(rows) == 0 {
		return Section{}, false
	}

	limit := defaultSectionItemLimit
	if len(rows) < limit {
		limit = len(rows)
	}

	items := make([]Item, 0, limit)
	for i := 0; i < limit; i++ {
		items = append(items, Item{
			Type:   ItemTypeArtist,
			Artist: &rows[i],
		})
	}

	return Section{
		Titles: map[string]string{"en": "Followed Artists", "ru": "Отслеживаемые исполнители"},
		Items:  items,
	}, true
}

func (s *Service) enqueueRefresh(ctx context.Context, userID, reason string) {
	if s.feedRefreshQueue == nil {
		return
	}
	_ = s.feedRefreshQueue.Enqueue(ctx, userID, reason)
}

func (s FeedSnapshot) IsFresh(now time.Time, freshness time.Duration) bool {
	if s.Layout == nil {
		return false
	}
	if s.StaleAt != nil && !s.StaleAt.After(now) {
		return false
	}
	if freshness <= 0 {
		freshness = time.Hour
	}
	return s.GeneratedAt.Add(freshness).After(now)
}
