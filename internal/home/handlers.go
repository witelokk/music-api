package home

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/witelokk/music-api/internal/auth"
	"github.com/witelokk/music-api/internal/mediaurl"
	openapi "github.com/witelokk/music-api/internal/openapi"
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

	layout, err := service.GetHomeScreenLayout(ctx, userID, time.Now())
	if err != nil {
		reqLogger.Error("failed to build home screen layout",
			slog.String("user_id", userID),
			slog.String("error", err.Error()),
		)
		return openapi.GetHomeFeed500JSONResponse(openapi.Error{Error: "failed to fetch home screen layout"}), nil
	}

	respPlaylists := make([]openapi.PlaylistSummary, 0, len(layout.Playlists))
	for _, p := range layout.Playlists {
		summary := openapi.PlaylistSummary{
			Id:         uuid.MustParse(p.ID),
			Name:       p.Name,
			SongsCount: p.SongsCount,
		}
		if p.CoverMediaID != nil && *p.CoverMediaID != "" {
			coverURL := mediaurl.Build(*p.CoverMediaID)
			summary.CoverUrl = &coverURL
		}
		respPlaylists = append(respPlaylists, summary)
	}

	var (
		artistSummaries = make([]openapi.ArtistSummary, 0, len(layout.FollowedArtists))
		artistNames     = make([]string, 0, len(layout.FollowedArtists))
	)
	for _, a := range layout.FollowedArtists {
		summary := openapi.ArtistSummary{
			Id:   uuid.MustParse(a.ID),
			Name: a.Name,
		}
		if a.AvatarMediaID != nil && *a.AvatarMediaID != "" {
			avatarURL := mediaurl.Build(*a.AvatarMediaID)
			summary.AvatarUrl = &avatarURL
		}
		artistSummaries = append(artistSummaries, summary)
		artistNames = append(artistNames, a.Name)
	}

	sections := make([]openapi.HomeScreenSection, 0, len(layout.Sections))
	for _, sec := range layout.Sections {
		releases := make([]openapi.ReleaseSummary, 0, len(sec.Releases))
		for _, rel := range sec.Releases {
			releaseSummary := openapi.ReleaseSummary{
				Id:         openapi_types.UUID(uuid.MustParse(rel.ID)),
				Name:       rel.Name,
				Type:       releasesapi.MapReleaseType(rel.Type),
				ReleasedAt: rel.ReleaseAt.Format("2006-01-02"),
			}
			if rel.CoverMediaID != nil && *rel.CoverMediaID != "" {
				coverURL := mediaurl.Build(*rel.CoverMediaID)
				releaseSummary.CoverUrl = &coverURL
			}
			releases = append(releases, releaseSummary)
		}

		sections = append(sections, openapi.HomeScreenSection{
			Titles: sec.Titles,
			Releases: openapi.ReleaseSummaryList{
				Count:    len(releases),
				Releases: releases,
			},
		})
	}

	return openapi.GetHomeFeed200JSONResponse(openapi.HomeScreenLayout{
		Playlists: openapi.PlaylistsSummary{
			Count:     len(respPlaylists),
			Playlists: respPlaylists,
		},
		FollowedArtists: openapi.ArtistList{
			Count:   len(artistSummaries),
			Artists: artistSummaries,
			Names:   strings.Join(artistNames, ", "),
		},
		Sections: sections,
	}), nil
}
