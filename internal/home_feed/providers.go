package home_feed

import (
	"context"
	"time"

	"github.com/witelokk/music-api/internal/followings"
	"github.com/witelokk/music-api/internal/releases"
)

const defaultSectionItemLimit = 10
const maxRecommendedArtistsPerSection = 2

type RecentReleasesSectionProvider struct {
	repo releases.ReleasesRepository
}

var _ SectionProvider = (*RecentReleasesSectionProvider)(nil)

func NewRecentReleasesSectionProvider(repo releases.ReleasesRepository) *RecentReleasesSectionProvider {
	return &RecentReleasesSectionProvider{repo: repo}
}

func (p *RecentReleasesSectionProvider) ID() string {
	return "recent_releases"
}

func (p *RecentReleasesSectionProvider) Sections(ctx context.Context, req FeedRequest) ([]Section, error) {
	rows, err := p.repo.GetRecentReleases(ctx, sectionItemLimit(req))
	if err != nil {
		return nil, err
	}

	return releaseSections(map[string]string{"en": "Recent Releases", "ru": "Новые релизы"}, rows), nil
}

type FollowedArtistReleasesSectionProvider struct {
	repo releases.ReleasesRepository
}

var _ SectionProvider = (*FollowedArtistReleasesSectionProvider)(nil)

func NewFollowedArtistReleasesSectionProvider(repo releases.ReleasesRepository) *FollowedArtistReleasesSectionProvider {
	return &FollowedArtistReleasesSectionProvider{repo: repo}
}

func (p *FollowedArtistReleasesSectionProvider) ID() string {
	return "because_you_follow"
}

func (p *FollowedArtistReleasesSectionProvider) Sections(ctx context.Context, req FeedRequest) ([]Section, error) {
	if len(req.FollowedArtistSeeds) == 0 {
		return nil, nil
	}

	sections := make([]Section, 0, len(req.FollowedArtistSeeds))
	for _, artist := range req.FollowedArtistSeeds {
		rows, err := p.repo.GetReleasesByFollowedArtistSeeds(ctx, []string{artist.ID}, providerSeed(req, p.ID()+":"+artist.ID), sectionItemLimit(req))
		if err != nil {
			return nil, err
		}

		title := "Because you follow " + artist.Name
		sections = append(sections, releaseAndArtistSections(map[string]string{
			"en": title,
			"ru": "Потому что вы подписаны на " + artist.Name,
		}, rows, sectionItemLimit(req))...)
	}

	return sections, nil
}

type FavoriteSongSectionProvider struct {
	repo releases.ReleasesRepository
}

var _ SectionProvider = (*FavoriteSongSectionProvider)(nil)

func NewFavoriteSongSectionProvider(repo releases.ReleasesRepository) *FavoriteSongSectionProvider {
	return &FavoriteSongSectionProvider{repo: repo}
}

func (p *FavoriteSongSectionProvider) ID() string {
	return "because_you_liked"
}

func (p *FavoriteSongSectionProvider) Sections(ctx context.Context, req FeedRequest) ([]Section, error) {
	if len(req.FavoriteSongSeeds) == 0 {
		return nil, nil
	}

	sections := make([]Section, 0, len(req.FavoriteSongSeeds))
	for _, song := range req.FavoriteSongSeeds {
		rows, err := p.repo.GetReleasesByFavoriteSongSeeds(ctx, []string{song.ID}, providerSeed(req, p.ID()+":"+song.ID), sectionItemLimit(req))
		if err != nil {
			return nil, err
		}

		title := "Because you like " + song.Name
		sections = append(sections, releaseAndArtistSections(map[string]string{
			"en": title,
			"ru": "Потому что вам нравится " + song.Name,
		}, rows, sectionItemLimit(req))...)
	}

	return sections, nil
}

type SeededReleasesSectionProvider struct {
	repo   releases.ReleasesRepository
	id     string
	titles map[string]string
}

var _ SectionProvider = (*SeededReleasesSectionProvider)(nil)

func NewSeededReleasesSectionProvider(
	repo releases.ReleasesRepository,
	id string,
	titles map[string]string,
) *SeededReleasesSectionProvider {
	return &SeededReleasesSectionProvider{
		repo:   repo,
		id:     id,
		titles: titles,
	}
}

func (p *SeededReleasesSectionProvider) ID() string {
	return p.id
}

func (p *SeededReleasesSectionProvider) Sections(ctx context.Context, req FeedRequest) ([]Section, error) {
	now := req.Now
	if now.IsZero() {
		now = time.Now()
	}
	seed := providerSeed(req, p.ID())
	limit := sectionItemLimit(req)

	rows, err := p.repo.GetRandomReleases(ctx, seed, limit)
	if err != nil {
		return nil, err
	}

	return releaseSections(p.titles, rows), nil
}

func releaseSections(titles map[string]string, rows []releases.Release) []Section {
	limit := defaultSectionItemLimit
	if len(rows) < limit {
		limit = len(rows)
	}

	items := make([]Item, 0, limit)
	for i := 0; i < limit; i++ {
		items = append(items, Item{
			Type:    ItemTypeRelease,
			Release: &rows[i],
		})
	}

	return []Section{{
		Titles: titles,
		Items:  items,
	}}
}

func releaseAndArtistSections(titles map[string]string, rows []releases.Release, limit int) []Section {
	if limit <= 0 {
		limit = defaultSectionItemLimit
	}

	artists := make([]followings.FollowedArtist, 0, maxRecommendedArtistsPerSection)
	artistIDs := make(map[string]struct{})
	for i := range rows {
		for _, artist := range rows[i].Artists {
			if _, ok := artistIDs[artist.ID]; ok {
				continue
			}
			artistIDs[artist.ID] = struct{}{}
			artists = append(artists, followings.FollowedArtist{
				ID:            artist.ID,
				Name:          artist.Name,
				AvatarMediaID: artist.AvatarMediaID,
			})
			if len(artists) == maxRecommendedArtistsPerSection {
				break
			}
		}
		if len(artists) == maxRecommendedArtistsPerSection {
			break
		}
	}

	artistLimit := len(artists)
	if artistLimit > maxRecommendedArtistsPerSection {
		artistLimit = maxRecommendedArtistsPerSection
	}
	if artistLimit > limit {
		artistLimit = limit
	}

	releaseLimit := limit - artistLimit
	if len(rows) < releaseLimit {
		releaseLimit = len(rows)
	}

	items := make([]Item, 0, releaseLimit+artistLimit)
	for i := 0; i < releaseLimit; i++ {
		items = append(items, Item{
			Type:    ItemTypeRelease,
			Release: &rows[i],
		})
	}
	for i := 0; i < artistLimit; i++ {
		items = append(items, Item{
			Type:   ItemTypeArtist,
			Artist: &artists[i],
		})
	}

	return []Section{{
		Titles: titles,
		Items:  items,
	}}
}

func sectionItemLimit(req FeedRequest) int {
	if req.Limit > 0 {
		return req.Limit
	}
	return defaultSectionItemLimit
}

func providerSeed(req FeedRequest, providerID string) string {
	now := req.Now
	if now.IsZero() {
		now = time.Now()
	}
	return now.UTC().Format("2006-01-02") + ":" + req.UserID + ":" + providerID
}
