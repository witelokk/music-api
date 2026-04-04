package favorites

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/witelokk/music-api/internal/auth"
	openapi "github.com/witelokk/music-api/internal/openapi"
	"github.com/witelokk/music-api/internal/requestctx"
)

func HandleGetFavorites(
	ctx context.Context,
	favoritesService *FavoritesService,
	logger *slog.Logger,
	req openapi.GetFavoritesRequestObject,
) (openapi.GetFavoritesResponseObject, error) {
	_ = req
	reqLogger := requestctx.LoggerFromContext(ctx, logger)

	userID := auth.UserIDFromContext(ctx)
	if userID == "" {
		return openapi.GetFavorites500JSONResponse(openapi.Error{Error: "failed to fetch favorite songs"}), nil
	}

	songsRows, err := favoritesService.GetFavoriteSongs(ctx, userID)
	if err != nil {
		reqLogger.Error("failed to get favorite song ids",
			slog.String("user_id", userID),
			slog.String("error", err.Error()),
		)
		return openapi.GetFavorites500JSONResponse(openapi.Error{Error: "failed to fetch favorite songs"}), nil
	}

	var songsList []openapi.Song
	for _, song := range songsRows {
		respSong := openapi.Song{
			Id:              uuid.MustParse(song.ID),
			Name:            song.Name,
			DurationSeconds: song.DurationSeconds,
			StreamUrl:       song.StreamURL,
			IsFavorite:      true,
			Artists:         []openapi.ArtistSummary{},
		}
		if song.CoverURL != nil {
			respSong.CoverUrl = song.CoverURL
		}

		songsList = append(songsList, respSong)
	}

	return openapi.GetFavorites200JSONResponse(openapi.SongList{
		Count: len(songsList),
		Songs: songsList,
	}), nil
}

func HandleAddFavorite(
	ctx context.Context,
	favoritesService *FavoritesService,
	logger *slog.Logger,
	req openapi.AddFavoriteRequestObject,
) (openapi.AddFavoriteResponseObject, error) {
	reqLogger := requestctx.LoggerFromContext(ctx, logger)

	if req.Body == nil {
		return openapi.AddFavorite400JSONResponse(openapi.Error{Error: "invalid request body"}), nil
	}

	userID := auth.UserIDFromContext(ctx)
	if userID == "" {
		return openapi.AddFavorite500JSONResponse(openapi.Error{Error: "failed to add song to favorites"}), nil
	}

	songID := req.Body.SongId.String()
	if songID == "" {
		return openapi.AddFavorite400JSONResponse(openapi.Error{Error: "song_id is required"}), nil
	}

	if err := favoritesService.AddFavorite(ctx, userID, songID); err != nil {
		reqLogger.Error("failed to add favorite",
			slog.String("user_id", userID),
			slog.String("song_id", songID),
			slog.String("error", err.Error()),
		)
		return openapi.AddFavorite500JSONResponse(openapi.Error{Error: "failed to add song to favorites"}), nil
	}

	return openapi.AddFavorite204Response{}, nil
}

func HandleRemoveFavorite(
	ctx context.Context,
	favoritesService *FavoritesService,
	logger *slog.Logger,
	req openapi.RemoveFavoriteRequestObject,
) (openapi.RemoveFavoriteResponseObject, error) {
	reqLogger := requestctx.LoggerFromContext(ctx, logger)

	if req.Body == nil {
		return openapi.RemoveFavorite400JSONResponse(openapi.Error{Error: "invalid request body"}), nil
	}

	userID := auth.UserIDFromContext(ctx)
	if userID == "" {
		return openapi.RemoveFavorite500JSONResponse(openapi.Error{Error: "failed to remove song from favorites"}), nil
	}

	songID := req.Body.SongId.String()
	if songID == "" {
		return openapi.RemoveFavorite400JSONResponse(openapi.Error{Error: "song_id is required"}), nil
	}

	if err := favoritesService.RemoveFavorite(ctx, userID, songID); err != nil {
		reqLogger.Error("failed to remove favorite",
			slog.String("user_id", userID),
			slog.String("song_id", songID),
			slog.String("error", err.Error()),
		)
		return openapi.RemoveFavorite500JSONResponse(openapi.Error{Error: "failed to remove song from favorites"}), nil
	}

	return openapi.RemoveFavorite204Response{}, nil
}
