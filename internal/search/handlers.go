package search

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/witelokk/music-api/internal/auth"
	openapi "github.com/witelokk/music-api/internal/openapi"
	"github.com/witelokk/music-api/internal/requestctx"
)

func HandleSearch(
	ctx context.Context,
	service *Service,
	logger *slog.Logger,
	req openapi.SearchRequestObject,
) (openapi.SearchResponseObject, error) {
	reqLogger := requestctx.LoggerFromContext(ctx, logger)

	params := req.Params
	if params.Q == "" {
		return openapi.Search400JSONResponse(openapi.Error{Error: "query parameter q is required"}), nil
	}

	var (
		page  = 1
		limit = 20
	)
	if params.Page != nil {
		page = *params.Page
	}
	if params.Limit != nil {
		limit = *params.Limit
	}

	var resultType *ResultType
	if params.Type != nil {
		switch *params.Type {
		case openapi.SearchParamsTypeSong:
			t := ResultTypeSong
			resultType = &t
		case openapi.SearchParamsTypeArtist:
			t := ResultTypeArtist
			resultType = &t
		case openapi.SearchParamsTypeRelease:
			t := ResultTypeRelease
			resultType = &t
		case openapi.SearchParamsTypePlaylist:
			t := ResultTypePlaylist
			resultType = &t
		default:
			return openapi.Search400JSONResponse(openapi.Error{Error: "invalid type parameter"}), nil
		}
	}

	userID := auth.UserIDFromContext(ctx)

	results, err := service.Search(ctx, params.Q, resultType, page, limit, userID)
	if err != nil {
		reqLogger.Error("failed to perform search",
			slog.String("query", params.Q),
			slog.String("error", err.Error()),
		)
		return openapi.Search500JSONResponse(openapi.Error{Error: "failed to perform search"}), nil
	}

	items := make([]openapi.SearchResultItem, 0, len(results.Items))
	for _, item := range results.Items {
		var apiItem openapi.SearchResultItem

		switch item.Type {
		case ResultTypeSong:
			apiItem.Type = openapi.SearchResultItemTypeSong
			if item.Song != nil {
				artists := make([]openapi.ArtistSummary, 0, len(item.Song.Artists))
				for _, a := range item.Song.Artists {
					artists = append(artists, openapi.ArtistSummary{
						Id:        uuidMustParse(a.ID),
						Name:      a.Name,
						AvatarUrl: a.AvatarURL,
					})
				}
				apiItem.Song = &openapi.Song{
					Id:              uuidMustParse(item.Song.ID),
					Name:            item.Song.Name,
					CoverUrl:        item.Song.CoverURL,
					DurationSeconds: item.Song.DurationSeconds,
					StreamUrl:       item.Song.StreamURL,
					IsFavorite:      item.Song.IsFavorite,
					Artists:         artists,
				}
			}
		case ResultTypeArtist:
			apiItem.Type = openapi.SearchResultItemTypeArtist
			if item.Artist != nil {
				apiItem.Artist = &openapi.ArtistSummary{
					Id:        uuidMustParse(item.Artist.ID),
					Name:      item.Artist.Name,
					AvatarUrl: item.Artist.AvatarURL,
				}
			}
		case ResultTypeRelease:
			apiItem.Type = openapi.SearchResultItemTypeRelease
			if item.Release != nil {
				apiItem.Release = &openapi.ReleaseSummary{
					Id:         uuidMustParse(item.Release.ID),
					Name:       item.Release.Name,
					CoverUrl:   item.Release.CoverURL,
					Type:       intToString(item.Release.Type),
					ReleasedAt: item.Release.ReleaseAt,
				}
			}
		case ResultTypePlaylist:
			apiItem.Type = openapi.SearchResultItemTypePlaylist
			if item.Playlist != nil {
				apiItem.Playlist = &openapi.PlaylistSummary{
					Id:         uuidMustParse(item.Playlist.ID),
					Name:       item.Playlist.Name,
					CoverUrl:   item.Playlist.CoverURL,
					SongsCount: item.Playlist.SongsCount,
				}
			}
		default:
			continue
		}

		items = append(items, apiItem)
	}

	resp := openapi.SearchResponse{
		Query:   results.Query,
		Page:    results.Page,
		Limit:   results.Limit,
		Total:   results.Total,
		Results: items,
	}

	return openapi.Search200JSONResponse(resp), nil
}

func uuidMustParse(id string) openapi_types.UUID {
	return openapi_types.UUID(uuid.MustParse(id))
}

func intToString(v int) string {
	return strconv.Itoa(v)
}
