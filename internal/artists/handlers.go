package artists

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/witelokk/music-api/internal/auth"
	"github.com/witelokk/music-api/internal/mediaurl"
	openapi "github.com/witelokk/music-api/internal/openapi"
	releasesapi "github.com/witelokk/music-api/internal/releases"
	"github.com/witelokk/music-api/internal/requestctx"
)

func HandleGetArtist(
	ctx context.Context,
	artistsService *ArtistsService,
	logger *slog.Logger,
	req openapi.GetArtistRequestObject,
) (openapi.GetArtistResponseObject, error) {
	reqLogger := requestctx.LoggerFromContext(ctx, logger)

	id := req.Id

	userID := auth.UserIDFromContext(ctx)

	artist, followers, following, err := artistsService.GetArtistWithStats(ctx, id.String(), userID)
	if err != nil {
		if errors.Is(err, ErrArtistNotFound) {
			return openapi.GetArtist404JSONResponse(openapi.Error{Error: "artist not found"}), nil
		}

		reqLogger.Error("failed to get artist",
			slog.String("id", id.String()),
			slog.String("error", err.Error()),
		)
		return openapi.GetArtist500JSONResponse(openapi.Error{Error: "failed to fetch artist"}), nil
	}

	artistID := uuid.MustParse(artist.ID)

	popularSongs := make([]openapi.Song, 0, len(artist.Popular))
	for _, s := range artist.Popular {
		songArtists := make([]openapi.ArtistSummary, 0, len(s.Artists))
		for _, a := range s.Artists {
			summary := openapi.ArtistSummary{
				Id:   uuid.MustParse(a.ID),
				Name: a.Name,
			}
			if a.AvatarMediaID != nil && *a.AvatarMediaID != "" {
				avatarURL := mediaurl.Build(*a.AvatarMediaID)
				summary.AvatarUrl = &avatarURL
			}
			songArtists = append(songArtists, summary)
		}

		song := openapi.Song{
			Id:              uuid.MustParse(s.ID),
			Name:            s.Name,
			DurationSeconds: s.DurationSeconds,
			StreamUrl:       mediaurl.Build(s.StreamMediaID),
			IsFavorite:      s.IsFavorite,
			Artists:         songArtists,
		}
		if s.CoverMediaID != nil && *s.CoverMediaID != "" {
			coverURL := mediaurl.Build(*s.CoverMediaID)
			song.CoverUrl = &coverURL
		}
		popularSongs = append(popularSongs, song)
	}

	releases := make([]openapi.ReleaseSummary, 0, len(artist.Releases))
	for _, r := range artist.Releases {
		rel := openapi.ReleaseSummary{
			Id:         uuid.MustParse(r.ID),
			Name:       r.Name,
			Type:       releasesapi.MapReleaseType(r.Type),
			ReleasedAt: r.ReleaseAt.Format("2006-01-02"),
		}
		if r.CoverMediaID != nil && *r.CoverMediaID != "" {
			coverURL := mediaurl.Build(*r.CoverMediaID)
			rel.CoverUrl = &coverURL
		}
		releases = append(releases, rel)
	}

	resp := openapi.Artist{
		Id:        artistID,
		Name:      artist.Name,
		Followers: followers,
		Following: following,
		PopularSongs: openapi.SongList{
			Count: len(popularSongs),
			Songs: popularSongs,
		},
		Releases: openapi.ReleaseSummaryList{
			Count:    len(releases),
			Releases: releases,
		},
	}

	if artist.AvatarMediaID != nil && *artist.AvatarMediaID != "" {
		avatarURL := mediaurl.Build(*artist.AvatarMediaID)
		resp.AvatarUrl = &avatarURL
	}
	if artist.CoverMediaID != nil && *artist.CoverMediaID != "" {
		coverURL := mediaurl.Build(*artist.CoverMediaID)
		resp.CoverUrl = &coverURL
	}

	return openapi.GetArtist200JSONResponse(resp), nil
}
