package releases

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/google/uuid"
	"github.com/witelokk/music-api/internal/requestctx"
	openapi "github.com/witelokk/music-api/internal/openapi"
)

func HandleGetRelease(
	ctx context.Context,
	service *Service,
	logger *slog.Logger,
	req openapi.GetReleaseRequestObject,
) (openapi.GetReleaseResponseObject, error) {
	reqLogger := requestctx.LoggerFromContext(ctx, logger)

	id := req.Id

	release, err := service.GetRelease(ctx, id.String())
	if err != nil {
		if errors.Is(err, ErrReleaseNotFound) {
			return openapi.GetRelease404JSONResponse(openapi.Error{Error: "release not found"}), nil
		}

		reqLogger.Error("failed to get release",
			slog.String("id", id.String()),
			slog.String("error", err.Error()),
		)
		return openapi.GetRelease500JSONResponse(openapi.Error{Error: "failed to fetch release"}), nil
	}

	resp := openapi.Release{
		Id:         uuid.MustParse(release.ID),
		Name:       release.Name,
		Type:       strconv.Itoa(release.Type),
		ReleasedAt: release.ReleaseAt.Format("2006-01-02"),
		Songs:      openapi.SongList{Count: 0, Songs: []openapi.Song{}},
		Artists: openapi.ArtistList{
			Count:   0,
			Artists: []openapi.ArtistSummary{},
			Names:   "",
		},
	}

	if release.CoverURL != nil {
		resp.CoverUrl = release.CoverURL
	}

	return openapi.GetRelease200JSONResponse(resp), nil
}
