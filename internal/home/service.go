package home

import (
	"context"
	"math/rand"
	"time"

	"github.com/witelokk/music-api/internal/followings"
	"github.com/witelokk/music-api/internal/playlists"
	"github.com/witelokk/music-api/internal/releases"
)

type ItemType string

const (
	ItemTypeRelease  ItemType = "release"
	ItemTypePlaylist ItemType = "playlist"
	ItemTypeArtist   ItemType = "artist"
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
	Playlists       []playlists.PlaylistSummary
	FollowedArtists []followings.FollowedArtist
	Sections        []Section
}

type Service struct {
	playlistsRepo  playlists.PlaylistsRepository
	followingsRepo followings.FollowingsRepository
	releasesRepo   releases.ReleasesRepository
}

func NewService(
	playlistsRepo playlists.PlaylistsRepository,
	followingsRepo followings.FollowingsRepository,
	releasesRepo releases.ReleasesRepository,
) *Service {
	return &Service{
		playlistsRepo:  playlistsRepo,
		followingsRepo: followingsRepo,
		releasesRepo:   releasesRepo,
	}
}

func (s *Service) GetHomeFeed(ctx context.Context, userID string, now time.Time) (*Layout, error) {
	// Use date (UTC) as a stable seed so layout is the
	// same during the day but changes day-to-day.
	now = now.UTC()
	seedStr := now.Format("2006-01-02")

	playlistsRows, err := s.playlistsRepo.GetPlaylists(ctx, userID)
	if err != nil {
		return nil, err
	}

	followedArtistsRows, err := s.followingsRepo.GetFollowedArtists(ctx, userID)
	if err != nil {
		return nil, err
	}

	allReleases, err := s.releasesRepo.GetRandomReleases(ctx, seedStr, 50)
	if err != nil {
		return nil, err
	}

	sections := buildSections(seedStr, allReleases, playlistsRows, followedArtistsRows)

	return &Layout{
		Playlists:       playlistsRows,
		FollowedArtists: followedArtistsRows,
		Sections:        sections,
	}, nil
}

func buildSections(
	seed string,
	allReleases []releases.Release,
	allPlaylists []playlists.PlaylistSummary,
	allArtists []followings.FollowedArtist,
) []Section {
	sectionDefs := []Section{
		{Titles: map[string]string{"en": "Featured Releases", "ru": "Избранные релизы"}},
		{Titles: map[string]string{"en": "Popular This Week", "ru": "Популярное на этой неделе"}},
		{Titles: map[string]string{"en": "Discover New Music", "ru": "Откройте новую музыку"}},
		{Titles: map[string]string{"en": "Recently Added", "ru": "Недавно добавленные"}},
	}

	if len(sectionDefs) == 0 {
		return nil
	}

	// Derive a stable int64 seed from the date string.
	var h int64
	for i := 0; i < len(seed); i++ {
		h = h*31 + int64(seed[i])
	}
	rnd := rand.New(rand.NewSource(h))

	sectionCount := rnd.Intn(len(sectionDefs)) + 1
	indexes := rnd.Perm(len(sectionDefs))[:sectionCount]
	itemPool := make([]Item, 0, len(allReleases)+len(allPlaylists)+len(allArtists))
	for i := range allReleases {
		itemPool = append(itemPool, Item{
			Type:    ItemTypeRelease,
			Release: &allReleases[i],
		})
	}
	for i := range allPlaylists {
		itemPool = append(itemPool, Item{
			Type:     ItemTypePlaylist,
			Playlist: &allPlaylists[i],
		})
	}
	for i := range allArtists {
		itemPool = append(itemPool, Item{
			Type:   ItemTypeArtist,
			Artist: &allArtists[i],
		})
	}

	sections := make([]Section, 0, sectionCount)
	for _, idx := range indexes {
		def := sectionDefs[idx]

		var sectionItems []Item
		if len(itemPool) > 0 {
			// Derive per-section seed from base seed + title.
			var sh int64
			for i := 0; i < len(seed); i++ {
				sh = sh*31 + int64(seed[i])
			}
			for i := 0; i < len(def.Titles["en"]); i++ {
				sh = sh*31 + int64(def.Titles["en"][i])
			}
			srnd := rand.New(rand.NewSource(sh))

			count := srnd.Intn(10) + 1
			if count > len(itemPool) {
				count = len(itemPool)
			}

			indexes := srnd.Perm(len(itemPool))[:count]
			sectionItems = make([]Item, 0, count)
			for _, i := range indexes {
				sectionItems = append(sectionItems, itemPool[i])
			}
		}

		sections = append(sections, Section{
			Titles: def.Titles,
			Items:  sectionItems,
		})
	}

	return sections
}
