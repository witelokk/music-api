package home

import (
	"context"
	"math/rand"
	"time"

	"github.com/witelokk/music-api/internal/followings"
	"github.com/witelokk/music-api/internal/playlists"
	"github.com/witelokk/music-api/internal/releases"
)

type Section struct {
	Title    string
	TitleRu  string
	Releases []releases.Release
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

func (s *Service) GetHomeScreenLayout(ctx context.Context, userID string, now time.Time) (*Layout, error) {
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

	sections := buildSections(seedStr, allReleases)

	return &Layout{
		Playlists:       playlistsRows,
		FollowedArtists: followedArtistsRows,
		Sections:        sections,
	}, nil
}

func buildSections(seed string, allReleases []releases.Release) []Section {
	sectionDefs := []Section{
		{Title: "Featured Releases", TitleRu: "Избранные релизы"},
		{Title: "Popular This Week", TitleRu: "Популярное на этой неделе"},
		{Title: "Discover New Music", TitleRu: "Откройте новую музыку"},
		{Title: "Recently Added", TitleRu: "Недавно добавленные"},
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

	sections := make([]Section, 0, sectionCount)
	for _, idx := range indexes {
		def := sectionDefs[idx]

		var sectionReleases []releases.Release
		if len(allReleases) > 0 {
			// Derive per-section seed from base seed + title.
			var sh int64
			for i := 0; i < len(def.Title); i++ {
				sh = sh*31 + int64(def.Title[i])
			}
			srnd := rand.New(rand.NewSource(sh))

			count := srnd.Intn(10) + 1
			if count > len(allReleases) {
				count = len(allReleases)
			}

			indexes := srnd.Perm(len(allReleases))[:count]
			sectionReleases = make([]releases.Release, 0, count)
			for _, i := range indexes {
				sectionReleases = append(sectionReleases, allReleases[i])
			}
		}

		sections = append(sections, Section{
			Title:    def.Title,
			TitleRu:  def.TitleRu,
			Releases: sectionReleases,
		})
	}

	return sections
}
