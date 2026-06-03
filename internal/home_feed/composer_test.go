package home_feed

import (
	"context"
	"testing"
	"time"

	"github.com/witelokk/music-api/internal/favorites"
	"github.com/witelokk/music-api/internal/followings"
	"github.com/witelokk/music-api/internal/releases"
)

func TestRecommendationFeedComposer_ComposesBasicProviderSections(t *testing.T) {
	favoritesRepo := &fakeFavoritesRepo{
		songs: []favorites.FavoriteSong{
			{ID: "s1", Name: "Song 1"},
			{ID: "s2", Name: "Song 2"},
		},
	}
	followingsRepo := &fakeFollowingsRepo{
		artists: []followings.FollowedArtist{
			{ID: "a1", Name: "Artist 1"},
			{ID: "a2", Name: "Artist 2"},
		},
	}
	releasesRepo := &fakeReleasesRepo{
		recentReleases: []releases.Release{
			{ID: "recent-1", Name: "Recent Release 1"},
		},
		followedArtistReleases: []releases.Release{
			{
				ID:   "followed-1",
				Name: "Followed Release 1",
				Artists: []releases.ReleaseArtist{
					{ID: "artist-from-followed", Name: "Artist From Followed"},
				},
			},
		},
		favoriteSongReleases: []releases.Release{
			{
				ID:   "favorite-1",
				Name: "Favorite Release 1",
				Artists: []releases.ReleaseArtist{
					{ID: "artist-from-favorite", Name: "Artist From Favorite"},
				},
			},
		},
		releases: []releases.Release{
			{ID: "r1", Name: "Release 1"},
		},
	}
	composer := NewRecommendationFeedComposer(favoritesRepo, followingsRepo, releasesRepo)

	layout, err := composer.Compose(context.Background(), FeedRequest{
		UserID: "user-id",
		Now:    time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(layout.Playlists) != 0 {
		t.Fatalf("expected composer not to populate top-level playlists, got %+v", layout.Playlists)
	}
	if len(layout.FollowedArtists) != 0 {
		t.Fatalf("expected composer not to populate top-level artists, got %+v", layout.FollowedArtists)
	}
	if len(layout.Sections) != 8 {
		t.Fatalf("expected 8 provider sections, got %d", len(layout.Sections))
	}
	for _, section := range layout.Sections {
		assertSectionHasItemType(t, section, ItemTypeRelease)
	}
	if !hasSectionTitle(layout.Sections, "Because you follow Artist 1") && !hasSectionTitle(layout.Sections, "Because you follow Artist 2") {
		t.Fatalf("expected followed artist section, got %+v", layout.Sections)
	}
	if !hasSectionTitle(layout.Sections, "Because you like Song 1") && !hasSectionTitle(layout.Sections, "Because you like Song 2") {
		t.Fatalf("expected favorite song section, got %+v", layout.Sections)
	}
	if !hasAnySectionItemType(layout.Sections, ItemTypeArtist) {
		t.Fatalf("expected at least one recommendation section to include artist item, got %+v", layout.Sections)
	}
}

func assertSectionHasItemType(t *testing.T, section Section, itemType ItemType) {
	t.Helper()
	for _, item := range section.Items {
		if item.Type == itemType {
			return
		}
	}
	t.Fatalf("expected section %q to contain %q, got %+v", section.Titles["en"], itemType, section.Items)
}

func hasSectionTitle(sections []Section, title string) bool {
	for _, section := range sections {
		if section.Titles["en"] == title {
			return true
		}
	}
	return false
}

func hasAnySectionItemType(sections []Section, itemType ItemType) bool {
	for _, section := range sections {
		for _, item := range section.Items {
			if item.Type == itemType {
				return true
			}
		}
	}
	return false
}

func TestReleaseAndArtistSections_LimitsItemsAndKeepsArtistsFew(t *testing.T) {
	rows := make([]releases.Release, 12)
	for i := range rows {
		rows[i] = releases.Release{
			ID:   "release-" + string(rune('a'+i)),
			Name: "Release",
			Artists: []releases.ReleaseArtist{
				{ID: "artist-" + string(rune('a'+i)), Name: "Artist"},
			},
		}
	}

	sections := releaseAndArtistSections(map[string]string{"en": "Mixed"}, rows, 10)
	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	if len(sections[0].Items) != 10 {
		t.Fatalf("expected 10 items, got %d", len(sections[0].Items))
	}

	var releasesCount, artistsCount int
	for _, item := range sections[0].Items {
		switch item.Type {
		case ItemTypeRelease:
			releasesCount++
		case ItemTypeArtist:
			artistsCount++
		}
	}
	if releasesCount != 8 {
		t.Fatalf("expected 8 releases, got %d", releasesCount)
	}
	if artistsCount != 2 {
		t.Fatalf("expected 2 artists, got %d", artistsCount)
	}
}
