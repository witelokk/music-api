package artists

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/witelokk/music-api/internal/requestctx"
	openapi "github.com/witelokk/music-api/internal/openapi"
)

func HandleGetArtist(
	ctx context.Context,
	service *Service,
	logger *slog.Logger,
	req openapi.GetArtistRequestObject,
) (openapi.GetArtistResponseObject, error) {
	reqLogger := requestctx.LoggerFromContext(ctx, logger)

	id := req.Id

	artist, err := service.GetArtist(ctx, id.String())
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

	resp := openapi.Artist{
		Id:           uuid.MustParse(artist.ID),
		Name:         artist.Name,
		Followers:    0,
		Following:    false,
		PopularSongs: openapi.SongList{Count: 0, Songs: []openapi.Song{}},
		Releases:     openapi.ReleaseList{Count: 0, Releases: []openapi.Release{}},
	}

	if artist.AvatarURL != nil {
		resp.AvatarUrl = artist.AvatarURL
	}
	if artist.CoverURL != nil {
		resp.CoverUrl = artist.CoverURL
	}

	return openapi.GetArtist200JSONResponse(resp), nil
}
