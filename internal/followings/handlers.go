package followings

import (
	"context"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/witelokk/music-api/internal/auth"
	"github.com/witelokk/music-api/internal/mediaurl"
	openapi "github.com/witelokk/music-api/internal/openapi"
	"github.com/witelokk/music-api/internal/requestctx"
)

func HandleGetFollowings(
	ctx context.Context,
	followingsService *FollowingsService,
	logger *slog.Logger,
	req openapi.GetFollowingsRequestObject,
) (openapi.GetFollowingsResponseObject, error) {
	_ = req
	reqLogger := requestctx.LoggerFromContext(ctx, logger)

	userID := auth.UserIDFromContext(ctx)
	if userID == "" {
		return openapi.GetFollowings500JSONResponse(openapi.Error{Error: "failed to fetch followed artists"}), nil
	}

	artistsRows, err := followingsService.GetFollowedArtists(ctx, userID)
	if err != nil {
		reqLogger.Error("failed to get followed artist ids",
			slog.String("user_id", userID),
			slog.String("error", err.Error()),
		)
		return openapi.GetFollowings500JSONResponse(openapi.Error{Error: "failed to fetch followed artists"}), nil
	}

	var summaries []openapi.ArtistSummary
	var names []string

	for _, artist := range artistsRows {
		summary := openapi.ArtistSummary{
			Id:   uuid.MustParse(artist.ID),
			Name: artist.Name,
		}
		if artist.AvatarMediaID != nil && *artist.AvatarMediaID != "" {
			avatarURL := mediaurl.Build(*artist.AvatarMediaID)
			summary.AvatarUrl = &avatarURL
		}
		summaries = append(summaries, summary)
		names = append(names, artist.Name)
	}

	return openapi.GetFollowings200JSONResponse(openapi.ArtistList{
		Count:   len(summaries),
		Artists: summaries,
		Names:   strings.Join(names, ", "),
	}), nil
}

func HandleFollowArtist(
	ctx context.Context,
	followingsService *FollowingsService,
	logger *slog.Logger,
	req openapi.FollowArtistRequestObject,
) (openapi.FollowArtistResponseObject, error) {
	reqLogger := requestctx.LoggerFromContext(ctx, logger)

	if req.Body == nil {
		return openapi.FollowArtist400JSONResponse(openapi.Error{Error: "invalid request body"}), nil
	}

	userID := auth.UserIDFromContext(ctx)
	if userID == "" {
		return openapi.FollowArtist500JSONResponse(openapi.Error{Error: "failed to follow artist"}), nil
	}

	artistID := req.Body.ArtistId.String()
	if artistID == "" {
		return openapi.FollowArtist400JSONResponse(openapi.Error{Error: "artist_id is required"}), nil
	}

	if err := followingsService.Follow(ctx, userID, artistID); err != nil {
		reqLogger.Error("failed to follow artist",
			slog.String("user_id", userID),
			slog.String("artist_id", artistID),
			slog.String("error", err.Error()),
		)
		return openapi.FollowArtist500JSONResponse(openapi.Error{Error: "failed to follow artist"}), nil
	}

	return openapi.FollowArtist204Response{}, nil
}

func HandleUnfollowArtist(
	ctx context.Context,
	followingsService *FollowingsService,
	logger *slog.Logger,
	req openapi.UnfollowArtistRequestObject,
) (openapi.UnfollowArtistResponseObject, error) {
	reqLogger := requestctx.LoggerFromContext(ctx, logger)

	if req.Body == nil {
		return openapi.UnfollowArtist400JSONResponse(openapi.Error{Error: "invalid request body"}), nil
	}

	userID := auth.UserIDFromContext(ctx)
	if userID == "" {
		return openapi.UnfollowArtist500JSONResponse(openapi.Error{Error: "failed to unfollow artist"}), nil
	}

	artistID := req.Body.ArtistId.String()
	if artistID == "" {
		return openapi.UnfollowArtist400JSONResponse(openapi.Error{Error: "artist_id is required"}), nil
	}

	if err := followingsService.Unfollow(ctx, userID, artistID); err != nil {
		reqLogger.Error("failed to unfollow artist",
			slog.String("user_id", userID),
			slog.String("artist_id", artistID),
			slog.String("error", err.Error()),
		)
		return openapi.UnfollowArtist500JSONResponse(openapi.Error{Error: "failed to unfollow artist"}), nil
	}

	return openapi.UnfollowArtist204Response{}, nil
}
