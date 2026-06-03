package home_feed

import (
	"context"
	"math/rand"
	"time"

	"github.com/witelokk/music-api/internal/favorites"
	"github.com/witelokk/music-api/internal/followings"
	"github.com/witelokk/music-api/internal/releases"
)

// FeedRequest carries the request-scoped inputs used to build a home feed.
type FeedRequest struct {
	UserID string
	Now    time.Time
	Limit  int
	Locale string

	FavoriteSongSeeds   []favorites.FavoriteSong
	FollowedArtistSeeds []followings.FollowedArtist
}

// FeedComposer composes section candidates into the final home feed layout.
type FeedComposer interface {
	Compose(ctx context.Context, req FeedRequest) (*Layout, error)
}

type RecommendationFeedComposer struct {
	favoritesRepo  favorites.Repository
	followingsRepo followings.FollowingsRepository
	providers      []SectionProvider
}

var _ FeedComposer = (*RecommendationFeedComposer)(nil)

func NewRecommendationFeedComposer(
	favoritesRepo favorites.Repository,
	followingsRepo followings.FollowingsRepository,
	releasesRepo releases.ReleasesRepository,
) *RecommendationFeedComposer {
	return &RecommendationFeedComposer{
		favoritesRepo:  favoritesRepo,
		followingsRepo: followingsRepo,
		providers: []SectionProvider{
			NewRecentReleasesSectionProvider(releasesRepo),
			NewFollowedArtistReleasesSectionProvider(releasesRepo),
			NewFavoriteSongSectionProvider(releasesRepo),
			NewSeededReleasesSectionProvider(releasesRepo, "daily_discovery", map[string]string{"en": "Daily Discovery", "ru": "Открытия дня"}),
			NewSeededReleasesSectionProvider(releasesRepo, "hidden_gems", map[string]string{"en": "Hidden Gems", "ru": "Скрытые находки"}),
			NewSeededReleasesSectionProvider(releasesRepo, "fresh_picks", map[string]string{"en": "Fresh Picks", "ru": "Свежая подборка"}),
		},
	}
}

func (c *RecommendationFeedComposer) Compose(ctx context.Context, req FeedRequest) (*Layout, error) {
	req = c.withSeedData(ctx, req)

	sections, err := composeSections(ctx, req, c.providers)
	if err != nil {
		return nil, err
	}
	shuffleSections(req, sections)

	return &Layout{
		Sections: sections,
	}, nil
}

func (c *RecommendationFeedComposer) withSeedData(ctx context.Context, req FeedRequest) FeedRequest {
	seed := stableSeed(req)
	rnd := rand.New(rand.NewSource(seed))

	if c.favoritesRepo != nil {
		if songs, err := c.favoritesRepo.GetFavoriteSongs(ctx, req.UserID); err == nil {
			req.FavoriteSongSeeds = sampleFavoriteSongs(songs, 3, rnd)
		}
	}

	if c.followingsRepo != nil {
		if artists, err := c.followingsRepo.GetFollowedArtists(ctx, req.UserID); err == nil {
			req.FollowedArtistSeeds = sampleFollowedArtists(artists, 3, rnd)
		}
	}

	return req
}

// SectionProvider produces candidate sections for one feed concern.
type SectionProvider interface {
	ID() string
	Sections(ctx context.Context, req FeedRequest) ([]Section, error)
}

func composeSections(ctx context.Context, req FeedRequest, providers []SectionProvider) ([]Section, error) {
	sections := make([]Section, 0, len(providers))
	for _, provider := range providers {
		providerSections, err := provider.Sections(ctx, req)
		if err != nil {
			return nil, err
		}
		for _, section := range providerSections {
			if len(section.Items) == 0 {
				continue
			}
			sections = append(sections, section)
		}
	}
	return sections, nil
}

func shuffleSections(req FeedRequest, sections []Section) {
	rnd := rand.New(rand.NewSource(stableSeed(req) + 17))
	rnd.Shuffle(len(sections), func(i, j int) {
		sections[i], sections[j] = sections[j], sections[i]
	})
}

func stableSeed(req FeedRequest) int64 {
	now := req.Now
	if now.IsZero() {
		now = time.Now()
	}
	seedStr := now.UTC().Format("2006-01-02") + req.UserID
	var seed int64
	for i := 0; i < len(seedStr); i++ {
		seed = seed*31 + int64(seedStr[i])
	}
	return seed
}

func sampleFavoriteSongs(rows []favorites.FavoriteSong, limit int, rnd *rand.Rand) []favorites.FavoriteSong {
	if len(rows) <= limit {
		return rows
	}
	indexes := rnd.Perm(len(rows))[:limit]
	result := make([]favorites.FavoriteSong, 0, limit)
	for _, idx := range indexes {
		result = append(result, rows[idx])
	}
	return result
}

func sampleFollowedArtists(rows []followings.FollowedArtist, limit int, rnd *rand.Rand) []followings.FollowedArtist {
	if len(rows) <= limit {
		return rows
	}
	indexes := rnd.Perm(len(rows))[:limit]
	result := make([]followings.FollowedArtist, 0, limit)
	for _, idx := range indexes {
		result = append(result, rows[idx])
	}
	return result
}
