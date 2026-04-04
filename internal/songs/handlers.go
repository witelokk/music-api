package songs

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/google/uuid"
	openapi "github.com/witelokk/music-api/internal/openapi"
)

func HandleGetSong(
	ctx context.Context,
	service *Service,
	logger *slog.Logger,
	req openapi.GetSongRequestObject,
) (openapi.GetSongResponseObject, error) {
	id := req.Id

	song, err := service.GetSong(ctx, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return openapi.GetSong404JSONResponse(openapi.Error{Error: "song not found"}), nil
		}

		logger.Error("failed to get song",
			slog.String("id", id.String()),
			slog.String("error", err.Error()),
		)
		return openapi.GetSong500JSONResponse(openapi.Error{Error: "failed to fetch song"}), nil
	}

	// For now, no favorites & artists logic is implemented.
	resp := openapi.Song{
		Id:              uuid.MustParse(song.ID),
		Name:            song.Name,
		DurationSeconds: song.DurationSeconds,
		StreamUrl:       song.StreamURL,
		IsFavorite:      false,
		Artists:         []openapi.ArtistSummary{},
	}
	if song.CoverURL != nil {
		resp.CoverUrl = song.CoverURL
	}

	return openapi.GetSong200JSONResponse(resp), nil
}

