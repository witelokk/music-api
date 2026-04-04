package artists

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/google/uuid"
	"github.com/witelokk/music-api/internal/auth"
	openapi "github.com/witelokk/music-api/internal/openapi"
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

	mainArtistSummary := openapi.ArtistSummary{
		Id:   artistID,
		Name: artist.Name,
	}
	if artist.AvatarURL != nil {
		mainArtistSummary.AvatarUrl = artist.AvatarURL
	}

	popularSongs := make([]openapi.Song, 0, len(artist.Popular))
	for _, s := range artist.Popular {
		song := openapi.Song{
			Id:              uuid.MustParse(s.ID),
			Name:            s.Name,
			DurationSeconds: s.DurationSeconds,
			StreamUrl:       s.StreamURL,
			IsFavorite:      false,
			Artists:         []openapi.ArtistSummary{mainArtistSummary},
		}
		if s.CoverURL != nil {
			song.CoverUrl = s.CoverURL
		}
		popularSongs = append(popularSongs, song)
	}

	releases := make([]openapi.Release, 0, len(artist.Releases))
	for _, r := range artist.Releases {
		rel := openapi.Release{
			Id:         uuid.MustParse(r.ID),
			Name:       r.Name,
			Type:       strconv.Itoa(r.Type),
			ReleasedAt: r.ReleaseAt.Format("2006-01-02"),
			Songs: openapi.SongList{
				Count: 0,
				Songs: []openapi.Song{},
			},
			Artists: openapi.ArtistList{
				Count:   1,
				Artists: []openapi.ArtistSummary{mainArtistSummary},
				Names:   artist.Name,
			},
		}
		if r.CoverURL != nil {
			rel.CoverUrl = r.CoverURL
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
		Releases: openapi.ReleaseList{
			Count:    len(releases),
			Releases: releases,
		},
	}

	if artist.AvatarURL != nil {
		resp.AvatarUrl = artist.AvatarURL
	}
	if artist.CoverURL != nil {
		resp.CoverUrl = artist.CoverURL
	}

	return openapi.GetArtist200JSONResponse(resp), nil
}
