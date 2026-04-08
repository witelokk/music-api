package media

import (
	"context"
	"errors"
	"log/slog"

	"github.com/witelokk/music-api/internal/openapi"
	"github.com/witelokk/music-api/internal/requestctx"
)

func GetMedia(
	ctx context.Context,
	mediaService *MediaService,
	request openapi.GetMediaRequestObject,
) (openapi.GetMediaResponseObject, error) {
	reqLogger := requestctx.LoggerFromContext(ctx, nil)

	if mediaService.storage == nil {
		return openapi.GetMedia500JSONResponse(openapi.Error{Error: "media service not configured"}), nil
	}

	objectName := request.Id.String()

	reader, size, err := mediaService.GetObjectStream(ctx, objectName)
	if err != nil {
		if errors.Is(err, ErrMediaNotFound) {
			return openapi.GetMedia404JSONResponse(openapi.Error{Error: "media not found"}), nil
		}

		reqLogger.Error("failed to get media",
			slog.String("id", objectName),
			slog.String("error", err.Error()),
		)
		return openapi.GetMedia500JSONResponse(openapi.Error{Error: "failed to fetch media"}), nil
	}

	return openapi.GetMedia200ApplicationoctetStreamResponse{
		Body:          reader,
		ContentLength: size,
	}, nil
}
