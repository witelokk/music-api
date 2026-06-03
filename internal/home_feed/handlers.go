package home_feed

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/witelokk/music-api/internal/auth"
	"github.com/witelokk/music-api/internal/favorites"
	"github.com/witelokk/music-api/internal/mediaurl"
	openapi "github.com/witelokk/music-api/internal/openapi"
	"github.com/witelokk/music-api/internal/playlists"
	releasesapi "github.com/witelokk/music-api/internal/releases"
	"github.com/witelokk/music-api/internal/requestctx"
)

func HandleGetHomeFeed(
	ctx context.Context,
	service *Service,
	logger *slog.Logger,
	req openapi.GetHomeFeedRequestObject,
) (openapi.GetHomeFeedResponseObject, error) {
	_ = req
	reqLogger := requestctx.LoggerFromContext(ctx, logger)

	userID := auth.UserIDFromContext(ctx)
	if userID == "" {
		return openapi.GetHomeFeed500JSONResponse(openapi.Error{Error: "failed to fetch home screen layout"}), nil
	}

	layout, err := service.GetHomeFeed(ctx, userID, time.Now())
	if err != nil {
		reqLogger.Error("failed to build home screen layout",
			slog.String("user_id", userID),
			slog.String("error", err.Error()),
		)
		return openapi.GetHomeFeed500JSONResponse(openapi.Error{Error: "failed to fetch home screen layout"}), nil
	}

	sections := make([]openapi.HomeScreenSection, 0, len(layout.Sections))
	for _, sec := range layout.Sections {
		items := make([]openapi.HomeFeedItem, 0, len(sec.Items))
		for _, item := range sec.Items {
			switch item.Type {
			case ItemTypeRelease:
				if item.Release == nil {
					continue
				}
				releaseSummary := toOpenAPIReleaseSummary(*item.Release)
				items = append(items, openapi.HomeFeedItem{
					Type:    openapi.HomeFeedItemTypeRelease,
					Release: &releaseSummary,
				})
			case ItemTypePlaylist:
				if item.Playlist == nil {
					continue
				}
				playlistSummary := toOpenAPIPlaylistSummary(*item.Playlist)
				items = append(items, openapi.HomeFeedItem{
					Type:     openapi.HomeFeedItemTypePlaylist,
					Playlist: &playlistSummary,
				})
			case ItemTypeArtist:
				if item.Artist == nil {
					continue
				}
				artistSummary := toOpenAPIArtistSummary(item.Artist.ID, item.Artist.Name, item.Artist.AvatarMediaID)
				items = append(items, openapi.HomeFeedItem{
					Type:   openapi.HomeFeedItemTypeArtist,
					Artist: &artistSummary,
				})
			case ItemTypeFavorites:
				items = append(items, openapi.HomeFeedItem{
					Type: openapi.HomeFeedItemTypeFavorites,
				})
			}
		}

		sections = append(sections, openapi.HomeScreenSection{
			Titles: sec.Titles,
			Items:  items,
		})
	}

	return openapi.GetHomeFeed200JSONResponse(openapi.HomeFeed{
		Sections: sections,
	}), nil
}

func toOpenAPIFavoriteSong(song favorites.FavoriteSong) openapi.Song {
	artistSummaries := make([]openapi.ArtistSummary, 0, len(song.Artists))
	for _, a := range song.Artists {
		artistSummaries = append(artistSummaries, toOpenAPIArtistSummary(a.ID, a.Name, a.AvatarMediaID))
	}

	respSong := openapi.Song{
		Id:              uuid.MustParse(song.ID),
		Name:            song.Name,
		DurationSeconds: song.DurationSeconds,
		StreamUrl:       mediaurl.Build(song.StreamMediaID),
		IsFavorite:      true,
		Artists:         artistSummaries,
	}
	if song.CoverMediaID != nil && *song.CoverMediaID != "" {
		coverURL := mediaurl.Build(*song.CoverMediaID)
		respSong.CoverUrl = &coverURL
	}
	return respSong
}

func toOpenAPIPlaylistSummary(p playlists.PlaylistSummary) openapi.PlaylistSummary {
	summary := openapi.PlaylistSummary{
		Id:         uuid.MustParse(p.ID),
		Name:       p.Name,
		SongsCount: p.SongsCount,
	}
	if p.CoverMediaID != nil && *p.CoverMediaID != "" {
		coverURL := mediaurl.Build(*p.CoverMediaID)
		summary.CoverUrl = &coverURL
	}
	return summary
}

func toOpenAPIArtistSummary(id, name string, avatarMediaID *string) openapi.ArtistSummary {
	summary := openapi.ArtistSummary{
		Id:   uuid.MustParse(id),
		Name: name,
	}
	if avatarMediaID != nil && *avatarMediaID != "" {
		avatarURL := mediaurl.Build(*avatarMediaID)
		summary.AvatarUrl = &avatarURL
	}
	return summary
}

func toOpenAPIReleaseSummary(rel releasesapi.Release) openapi.ReleaseSummary {
	artists, artistNames := toOpenAPIReleaseArtistList(rel.Artists)
	summary := openapi.ReleaseSummary{
		Id:         uuid.MustParse(rel.ID),
		Name:       rel.Name,
		Type:       releasesapi.MapReleaseType(rel.Type),
		ReleasedAt: rel.ReleaseAt.Format("2006-01-02"),
		Artists: openapi.ArtistList{
			Count:   len(artists),
			Artists: artists,
			Names:   strings.Join(artistNames, ", "),
		},
	}
	if rel.CoverMediaID != nil && *rel.CoverMediaID != "" {
		coverURL := mediaurl.Build(*rel.CoverMediaID)
		summary.CoverUrl = &coverURL
	}
	return summary
}

func toOpenAPIReleaseArtistList(artists []releasesapi.ReleaseArtist) ([]openapi.ArtistSummary, []string) {
	summaries := make([]openapi.ArtistSummary, 0, len(artists))
	names := make([]string, 0, len(artists))
	for _, a := range artists {
		summaries = append(summaries, toOpenAPIArtistSummary(a.ID, a.Name, a.AvatarMediaID))
		names = append(names, a.Name)
	}
	return summaries, names
}
