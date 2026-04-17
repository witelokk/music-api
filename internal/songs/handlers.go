package songs

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/witelokk/music-api/internal/auth"
	"github.com/witelokk/music-api/internal/mediaurl"
	openapi "github.com/witelokk/music-api/internal/openapi"
	"github.com/witelokk/music-api/internal/requestctx"
)

func HandleGetSong(
	ctx context.Context,
	songsService *SongsService,
	logger *slog.Logger,
	req openapi.GetSongRequestObject,
) (openapi.GetSongResponseObject, error) {
	reqLogger := requestctx.LoggerFromContext(ctx, logger)

	id := req.Id

	userID := auth.UserIDFromContext(ctx)

	song, isFavorite, err := songsService.GetSongWithFavorite(ctx, id.String(), userID)
	if err != nil {
		if errors.Is(err, ErrSongNotFound) {
			return openapi.GetSong404JSONResponse(openapi.Error{Error: "song not found"}), nil
		}

		reqLogger.Error("failed to get song",
			slog.String("id", id.String()),
			slog.String("error", err.Error()),
		)
		return openapi.GetSong500JSONResponse(openapi.Error{Error: "failed to fetch song"}), nil
	}

	artistSummaries := make([]openapi.ArtistSummary, 0, len(song.Artists))
	for _, a := range song.Artists {
		summary := openapi.ArtistSummary{
			Id:   uuid.MustParse(a.ID),
			Name: a.Name,
		}
		if a.AvatarMediaID != nil && *a.AvatarMediaID != "" {
			avatarURL := mediaurl.Build(*a.AvatarMediaID)
			summary.AvatarUrl = &avatarURL
		}
		artistSummaries = append(artistSummaries, summary)
	}

	resp := openapi.Song{
		Id:              uuid.MustParse(song.ID),
		Name:            song.Name,
		DurationSeconds: song.DurationSeconds,
		StreamUrl:       mediaurl.Build(song.StreamMediaID),
		IsFavorite:      isFavorite,
		Artists:         artistSummaries,
	}
	if song.CoverMediaID != nil && *song.CoverMediaID != "" {
		coverURL := mediaurl.Build(*song.CoverMediaID)
		resp.CoverUrl = &coverURL
	}

	return openapi.GetSong200JSONResponse(resp), nil
}
