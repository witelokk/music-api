package home

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/witelokk/music-api/internal/auth"
	openapi "github.com/witelokk/music-api/internal/openapi"
	"github.com/witelokk/music-api/internal/requestctx"
)

func HandleGetHomeScreenLayout(
	ctx context.Context,
	service *Service,
	logger *slog.Logger,
	req openapi.GetHomeScreenLayoutRequestObject,
) (openapi.GetHomeScreenLayoutResponseObject, error) {
	_ = req
	reqLogger := requestctx.LoggerFromContext(ctx, logger)

	userID := auth.UserIDFromContext(ctx)
	if userID == "" {
		return openapi.GetHomeScreenLayout500JSONResponse(openapi.Error{Error: "failed to fetch home screen layout"}), nil
	}

	layout, err := service.GetHomeScreenLayout(ctx, userID, time.Now())
	if err != nil {
		reqLogger.Error("failed to build home screen layout",
			slog.String("user_id", userID),
			slog.String("error", err.Error()),
		)
		return openapi.GetHomeScreenLayout500JSONResponse(openapi.Error{Error: "failed to fetch home screen layout"}), nil
	}

	respPlaylists := make([]openapi.PlaylistSummary, 0, len(layout.Playlists))
	for _, p := range layout.Playlists {
		summary := openapi.PlaylistSummary{
			Id:         uuid.MustParse(p.ID),
			Name:       p.Name,
			SongsCount: p.SongsCount,
		}
		if p.CoverURL != nil {
			summary.CoverUrl = p.CoverURL
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
		if a.AvatarURL != nil {
			summary.AvatarUrl = a.AvatarURL
		}
		artistSummaries = append(artistSummaries, summary)
		artistNames = append(artistNames, a.Name)
	}

	sections := make([]openapi.HomeScreenSection, 0, len(layout.Sections))
	for _, sec := range layout.Sections {
		releases := make([]openapi.Release, 0, len(sec.Releases))
		for _, rel := range sec.Releases {
			artistSummaries := make([]openapi.ArtistSummary, 0, len(rel.Artists))
			artistNames := make([]string, 0, len(rel.Artists))
			for _, a := range rel.Artists {
				summary := openapi.ArtistSummary{
					Id:   openapi_types.UUID(uuid.MustParse(a.ID)),
					Name: a.Name,
				}
				if a.AvatarURL != nil {
					summary.AvatarUrl = a.AvatarURL
				}
				artistSummaries = append(artistSummaries, summary)
				artistNames = append(artistNames, a.Name)
			}

			releases = append(releases, openapi.Release{
				Id:         openapi_types.UUID(uuid.MustParse(rel.ID)),
				Name:       rel.Name,
				CoverUrl:   rel.CoverURL,
				Type:       strconv.Itoa(rel.Type),
				ReleasedAt: rel.ReleaseAt.Format("2006-01-02"),
				Artists: openapi.ArtistList{
					Artists: artistSummaries,
					Count:   len(artistSummaries),
					Names:   strings.Join(artistNames, ", "),
				},
				Songs: openapi.SongList{
					Count: 0,
					Songs: make([]openapi.Song, 0),
				},
			})
		}

		sections = append(sections, openapi.HomeScreenSection{
			Title:   sec.Title,
			TitleRu: sec.TitleRu,
			Releases: openapi.ReleaseList{
				Count:    len(releases),
				Releases: releases,
			},
		})
	}

	return openapi.GetHomeScreenLayout200JSONResponse(openapi.HomeScreenLayout{
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
