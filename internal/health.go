package internal

import (
	"context"

	openapi "github.com/witelokk/music-api/internal/openapi"
)

func HandleGetHealth(ctx context.Context, req openapi.GetHealthRequestObject) (openapi.GetHealthResponseObject, error) {
	_ = ctx
	_ = req

	return openapi.GetHealth204Response{}, nil
}
