package playlists

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

func HandleGetPlaylists(
	ctx context.Context,
	playlistsService *PlaylistsService,
	logger *slog.Logger,
	req openapi.GetPlaylistsRequestObject,
) (openapi.GetPlaylistsResponseObject, error) {
	_ = req
	reqLogger := requestctx.LoggerFromContext(ctx, logger)

	userID := auth.UserIDFromContext(ctx)
	if userID == "" {
		return openapi.GetPlaylists500JSONResponse(openapi.Error{Error: "failed to fetch playlists"}), nil
	}

	playlists, err := playlistsService.GetPlaylists(ctx, userID)
	if err != nil {
		reqLogger.Error("failed to get playlists",
			slog.String("user_id", userID),
			slog.String("error", err.Error()),
		)
		return openapi.GetPlaylists500JSONResponse(openapi.Error{Error: "failed to fetch playlists"}), nil
	}

	respPlaylists := make([]openapi.PlaylistSummary, 0, len(playlists))
	for _, p := range playlists {
		summary := openapi.PlaylistSummary{
			Id:         uuid.MustParse(p.ID),
			Name:       p.Name,
			SongsCount: p.SongsCount,
		}
		if p.CoverMediaID != nil && *p.CoverMediaID != "" {
			coverURL := mediaurl.Build(*p.CoverMediaID)
			summary.CoverUrl = &coverURL
		}
		respPlaylists = append(respPlaylists, summary)
	}

	return openapi.GetPlaylists200JSONResponse(openapi.PlaylistsSummary{
		Count:     len(respPlaylists),
		Playlists: respPlaylists,
	}), nil
}

func HandleCreatePlaylist(
	ctx context.Context,
	service *PlaylistsService,
	logger *slog.Logger,
	req openapi.CreatePlaylistRequestObject,
) (openapi.CreatePlaylistResponseObject, error) {
	reqLogger := requestctx.LoggerFromContext(ctx, logger)

	if req.Body == nil || req.Body.Name == "" {
		return openapi.CreatePlaylist400JSONResponse(openapi.Error{Error: "invalid request body"}), nil
	}

	userID := auth.UserIDFromContext(ctx)
	if userID == "" {
		return openapi.CreatePlaylist500JSONResponse(openapi.Error{Error: "failed to create playlist"}), nil
	}

	id, err := service.CreatePlaylist(ctx, userID, req.Body.Name)
	if err != nil {
		reqLogger.Error("failed to create playlist",
			slog.String("user_id", userID),
			slog.String("error", err.Error()),
		)
		return openapi.CreatePlaylist500JSONResponse(openapi.Error{Error: "failed to create playlist"}), nil
	}

	parsedID, parseErr := uuid.Parse(id)
	if parseErr != nil {
		reqLogger.Error("failed to parse playlist id",
			slog.String("id", id),
			slog.String("error", parseErr.Error()),
		)
		return openapi.CreatePlaylist500JSONResponse(openapi.Error{Error: "failed to create playlist"}), nil
	}

	return openapi.CreatePlaylist201JSONResponse(openapi.CreatePlaylistResponse{
		Id: parsedID,
	}), nil
}

func HandleGetPlaylist(
	ctx context.Context,
	playlistsService *PlaylistsService,
	logger *slog.Logger,
	req openapi.GetPlaylistRequestObject,
) (openapi.GetPlaylistResponseObject, error) {
	reqLogger := requestctx.LoggerFromContext(ctx, logger)

	userID := auth.UserIDFromContext(ctx)
	if userID == "" {
		return openapi.GetPlaylist500JSONResponse(openapi.Error{Error: "failed to fetch playlist"}), nil
	}

	playlistID := req.Id.String()

	pl, songs, err := playlistsService.GetPlaylistWithSongs(ctx, userID, playlistID)
	if err != nil {
		if errors.Is(err, ErrPlaylistNotFound) {
			return openapi.GetPlaylist404JSONResponse(openapi.Error{Error: "playlist not found"}), nil
		}

		reqLogger.Error("failed to get playlist",
			slog.String("user_id", userID),
			slog.String("playlist_id", playlistID),
			slog.String("error", err.Error()),
		)
		return openapi.GetPlaylist500JSONResponse(openapi.Error{Error: "failed to fetch playlist"}), nil
	}

	respSongs := make([]openapi.Song, 0, len(songs))
	for _, sng := range songs {
		artistSummaries := make([]openapi.ArtistSummary, 0, len(sng.Artists))
		for _, a := range sng.Artists {
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
			Id:              uuid.MustParse(sng.ID),
			Name:            sng.Name,
			DurationSeconds: sng.DurationSeconds,
			StreamUrl:       mediaurl.Build(sng.StreamMediaID),
			IsFavorite:      sng.IsFavorite,
			Artists:         artistSummaries,
		}
		if sng.CoverMediaID != nil && *sng.CoverMediaID != "" {
			coverURL := mediaurl.Build(*sng.CoverMediaID)
			resp.CoverUrl = &coverURL
		}
		respSongs = append(respSongs, resp)
	}

	var coverURL *string
	if pl.CoverMediaID != nil && *pl.CoverMediaID != "" {
		value := mediaurl.Build(*pl.CoverMediaID)
		coverURL = &value
	}

	return openapi.GetPlaylist200JSONResponse(openapi.Playlist{
		Id:         uuid.MustParse(pl.ID),
		Name:       pl.Name,
		CoverUrl:   coverURL,
		SongsCount: pl.SongsCount,
		Songs: openapi.SongList{
			Count: len(respSongs),
			Songs: respSongs,
		},
	}), nil
}

func HandleUpdatePlaylist(
	ctx context.Context,
	playlistsService *PlaylistsService,
	logger *slog.Logger,
	req openapi.UpdatePlaylistRequestObject,
) (openapi.UpdatePlaylistResponseObject, error) {
	reqLogger := requestctx.LoggerFromContext(ctx, logger)

	if req.Body == nil || req.Body.Name == "" {
		return openapi.UpdatePlaylist400JSONResponse(openapi.Error{Error: "invalid request body"}), nil
	}

	userID := auth.UserIDFromContext(ctx)
	if userID == "" {
		return openapi.UpdatePlaylist500JSONResponse(openapi.Error{Error: "failed to update playlist"}), nil
	}

	playlistID := req.Id.String()

	if err := playlistsService.UpdatePlaylist(ctx, userID, playlistID, req.Body.Name); err != nil {
		if errors.Is(err, ErrPlaylistNotFound) {
			return openapi.UpdatePlaylist404JSONResponse(openapi.Error{Error: "playlist not found"}), nil
		}

		reqLogger.Error("failed to update playlist",
			slog.String("user_id", userID),
			slog.String("playlist_id", playlistID),
			slog.String("error", err.Error()),
		)
		return openapi.UpdatePlaylist500JSONResponse(openapi.Error{Error: "failed to update playlist"}), nil
	}

	return openapi.UpdatePlaylist204Response{}, nil
}

func HandleDeletePlaylist(
	ctx context.Context,
	playlistsService *PlaylistsService,
	logger *slog.Logger,
	req openapi.DeletePlaylistRequestObject,
) (openapi.DeletePlaylistResponseObject, error) {
	reqLogger := requestctx.LoggerFromContext(ctx, logger)

	userID := auth.UserIDFromContext(ctx)
	if userID == "" {
		return openapi.DeletePlaylist500JSONResponse(openapi.Error{Error: "failed to delete playlist"}), nil
	}

	playlistID := req.Id.String()

	if err := playlistsService.DeletePlaylist(ctx, userID, playlistID); err != nil {
		if errors.Is(err, ErrPlaylistNotFound) {
			return openapi.DeletePlaylist404JSONResponse(openapi.Error{Error: "playlist not found"}), nil
		}

		reqLogger.Error("failed to delete playlist",
			slog.String("user_id", userID),
			slog.String("playlist_id", playlistID),
			slog.String("error", err.Error()),
		)
		return openapi.DeletePlaylist500JSONResponse(openapi.Error{Error: "failed to delete playlist"}), nil
	}

	return openapi.DeletePlaylist204Response{}, nil
}

func HandleGetPlaylistSongs(
	ctx context.Context,
	playlistsService *PlaylistsService,
	logger *slog.Logger,
	req openapi.GetPlaylistSongsRequestObject,
) (openapi.GetPlaylistSongsResponseObject, error) {
	reqLogger := requestctx.LoggerFromContext(ctx, logger)

	userID := auth.UserIDFromContext(ctx)
	if userID == "" {
		return openapi.GetPlaylistSongs500JSONResponse(openapi.Error{Error: "failed to fetch playlist songs"}), nil
	}

	playlistID := req.Id.String()

	songs, err := playlistsService.GetPlaylistSongs(ctx, userID, playlistID)
	if err != nil {
		if errors.Is(err, ErrPlaylistNotFound) {
			return openapi.GetPlaylistSongs404JSONResponse(openapi.Error{Error: "playlist not found"}), nil
		}

		reqLogger.Error("failed to get playlist songs",
			slog.String("user_id", userID),
			slog.String("playlist_id", playlistID),
			slog.String("error", err.Error()),
		)
		return openapi.GetPlaylistSongs500JSONResponse(openapi.Error{Error: "failed to fetch playlist songs"}), nil
	}

	respSongs := make([]openapi.Song, 0, len(songs))
	for _, sng := range songs {
		artistSummaries := make([]openapi.ArtistSummary, 0, len(sng.Artists))
		for _, a := range sng.Artists {
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
			Id:              uuid.MustParse(sng.ID),
			Name:            sng.Name,
			DurationSeconds: sng.DurationSeconds,
			StreamUrl:       mediaurl.Build(sng.StreamMediaID),
			IsFavorite:      sng.IsFavorite,
			Artists:         artistSummaries,
		}
		if sng.CoverMediaID != nil && *sng.CoverMediaID != "" {
			coverURL := mediaurl.Build(*sng.CoverMediaID)
			resp.CoverUrl = &coverURL
		}
		respSongs = append(respSongs, resp)
	}

	return openapi.GetPlaylistSongs200JSONResponse(openapi.SongList{
		Count: len(respSongs),
		Songs: respSongs,
	}), nil
}

func HandleAddSongToPlaylist(
	ctx context.Context,
	playlistsService *PlaylistsService,
	logger *slog.Logger,
	req openapi.AddSongToPlaylistRequestObject,
) (openapi.AddSongToPlaylistResponseObject, error) {
	reqLogger := requestctx.LoggerFromContext(ctx, logger)

	if req.Body == nil {
		return openapi.AddSongToPlaylist400JSONResponse(openapi.Error{Error: "invalid request body"}), nil
	}

	userID := auth.UserIDFromContext(ctx)
	if userID == "" {
		return openapi.AddSongToPlaylist500JSONResponse(openapi.Error{Error: "failed to add song to playlist"}), nil
	}

	playlistID := req.Id.String()
	songID := req.Body.SongId.String()

	if songID == "" {
		return openapi.AddSongToPlaylist400JSONResponse(openapi.Error{Error: "song_id is required"}), nil
	}

	if err := playlistsService.AddSongToPlaylist(ctx, userID, playlistID, songID); err != nil {
		if errors.Is(err, ErrPlaylistNotFound) {
			return openapi.AddSongToPlaylist404JSONResponse(openapi.Error{Error: "playlist not found"}), nil
		}

		reqLogger.Error("failed to add song to playlist",
			slog.String("user_id", userID),
			slog.String("playlist_id", playlistID),
			slog.String("song_id", songID),
			slog.String("error", err.Error()),
		)
		return openapi.AddSongToPlaylist500JSONResponse(openapi.Error{Error: "failed to add song to playlist"}), nil
	}

	return openapi.AddSongToPlaylist204Response{}, nil
}

func HandleRemoveSongFromPlaylist(
	ctx context.Context,
	playlistsService *PlaylistsService,
	logger *slog.Logger,
	req openapi.RemoveSongFromPlaylistRequestObject,
) (openapi.RemoveSongFromPlaylistResponseObject, error) {
	reqLogger := requestctx.LoggerFromContext(ctx, logger)

	userID := auth.UserIDFromContext(ctx)
	if userID == "" {
		return openapi.RemoveSongFromPlaylist500JSONResponse(openapi.Error{Error: "failed to remove song from playlist"}), nil
	}

	playlistID := req.Id.String()
	songID := req.SongId.String()

	if err := playlistsService.RemoveSongFromPlaylist(ctx, userID, playlistID, songID); err != nil {
		if errors.Is(err, ErrPlaylistNotFound) {
			return openapi.RemoveSongFromPlaylist404JSONResponse(openapi.Error{Error: "playlist not found"}), nil
		}

		reqLogger.Error("failed to remove song from playlist",
			slog.String("user_id", userID),
			slog.String("playlist_id", playlistID),
			slog.String("song_id", songID),
			slog.String("error", err.Error()),
		)
		return openapi.RemoveSongFromPlaylist500JSONResponse(openapi.Error{Error: "failed to remove song from playlist"}), nil
	}

	return openapi.RemoveSongFromPlaylist204Response{}, nil
}
