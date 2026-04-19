package playlists

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/witelokk/music-api/internal/auth"
	openapi "github.com/witelokk/music-api/internal/openapi"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

func TestHandleGetPlaylists_Empty(t *testing.T) {
	logger := newTestLogger()
	repo := &fakePlaylistsRepo{
		playlists: []PlaylistSummary{},
	}
	svc := NewPlaylistsService(repo)

	req := openapi.GetPlaylistsRequestObject{}
	ctx := auth.WithUserID(context.Background(), "user-id")

	resp, err := HandleGetPlaylists(ctx, svc, logger, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	okResp, ok := resp.(openapi.GetPlaylists200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}
	if okResp.Count != 0 || len(okResp.Playlists) != 0 {
		t.Fatalf("expected empty playlists list, got count=%d playlists=%d", okResp.Count, len(okResp.Playlists))
	}
}

func TestHandleCreatePlaylist_BadBody(t *testing.T) {
	logger := newTestLogger()
	repo := &fakePlaylistsRepo{}
	svc := NewPlaylistsService(repo)

	ctx := auth.WithUserID(context.Background(), "user-id")

	resp, err := HandleCreatePlaylist(ctx, svc, logger, openapi.CreatePlaylistRequestObject{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := resp.(openapi.CreatePlaylist400JSONResponse); !ok {
		t.Fatalf("expected 400 response, got %T", resp)
	}
}

func TestHandleCreatePlaylist_Success(t *testing.T) {
	logger := newTestLogger()
	id := uuid.New().String()
	repo := &fakePlaylistsRepo{
		createID: id,
	}
	svc := NewPlaylistsService(repo)

	ctx := auth.WithUserID(context.Background(), "user-id")
	body := openapi.CreatePlaylistJSONRequestBody{
		Name: "My playlist",
	}
	req := openapi.CreatePlaylistRequestObject{
		Body: &body,
	}

	resp, err := HandleCreatePlaylist(ctx, svc, logger, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	okResp, ok := resp.(openapi.CreatePlaylist201JSONResponse)
	if !ok {
		t.Fatalf("expected 201 response, got %T", resp)
	}
	if okResp.Id.String() != id {
		t.Fatalf("expected id %s, got %s", id, okResp.Id.String())
	}
}

func TestHandleGetPlaylist_NotFound(t *testing.T) {
	logger := newTestLogger()
	repo := &fakePlaylistsRepo{
		playlistErr: pgx.ErrNoRows,
	}
	svc := NewPlaylistsService(repo)

	ctx := auth.WithUserID(context.Background(), "user-id")
	req := openapi.GetPlaylistRequestObject{
		Id: openapi_types.UUID(uuid.New()),
	}

	resp, err := HandleGetPlaylist(ctx, svc, logger, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := resp.(openapi.GetPlaylist404JSONResponse); !ok {
		t.Fatalf("expected 404 response, got %T", resp)
	}
}

func TestHandleGetPlaylist_Success(t *testing.T) {
	logger := newTestLogger()
	playlistID := uuid.New()
	songID := uuid.New()
	artistID := uuid.New()

	repo := &fakePlaylistsRepo{
		playlist: &Playlist{
			ID:         playlistID.String(),
			UserID:     "user-id",
			Name:       "My playlist",
			SongsCount: 1,
		},
		songs: []PlaylistSong{
			{
				ID:              songID.String(),
				Name:            "Song",
				DurationSeconds: 120,
				StreamMediaID:   "stream-id",
				IsFavorite:      true,
				Artists: []PlaylistArtist{
					{
						ID:   artistID.String(),
						Name: "Artist",
					},
				},
			},
		},
	}
	svc := NewPlaylistsService(repo)

	ctx := auth.WithUserID(context.Background(), "user-id")
	req := openapi.GetPlaylistRequestObject{
		Id: openapi_types.UUID(playlistID),
	}

	resp, err := HandleGetPlaylist(ctx, svc, logger, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	okResp, ok := resp.(openapi.GetPlaylist200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}
	if okResp.SongsCount != 1 {
		t.Fatalf("expected SongsCount=1, got %d", okResp.SongsCount)
	}
	if len(okResp.Songs.Songs) != 1 {
		t.Fatalf("expected 1 song, got %d", len(okResp.Songs.Songs))
	}
	if okResp.Songs.Songs[0].StreamUrl != "/media/stream-id" {
		t.Fatalf("expected stream url %q, got %q", "/media/stream-id", okResp.Songs.Songs[0].StreamUrl)
	}
	if len(okResp.Songs.Songs[0].Artists) != 1 {
		t.Fatalf("expected 1 artist, got %d", len(okResp.Songs.Songs[0].Artists))
	}
}

func TestHandleGetPlaylistSongs_Success(t *testing.T) {
	logger := newTestLogger()
	playlistID := uuid.New()
	songID := uuid.New()

	repo := &fakePlaylistsRepo{
		songs: []PlaylistSong{
			{
				ID:              songID.String(),
				Name:            "Song",
				DurationSeconds: 120,
				StreamMediaID:   "stream-id",
				IsFavorite:      true,
			},
		},
	}
	svc := NewPlaylistsService(repo)

	ctx := auth.WithUserID(context.Background(), "user-id")
	req := openapi.GetPlaylistSongsRequestObject{
		Id: openapi_types.UUID(playlistID),
	}

	resp, err := HandleGetPlaylistSongs(ctx, svc, logger, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	okResp, ok := resp.(openapi.GetPlaylistSongs200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}
	if okResp.Count != 1 || len(okResp.Songs) != 1 {
		t.Fatalf("expected 1 song, got count=%d songs=%d", okResp.Count, len(okResp.Songs))
	}
	if okResp.Songs[0].StreamUrl != "/media/stream-id" {
		t.Fatalf("expected stream url %q, got %q", "/media/stream-id", okResp.Songs[0].StreamUrl)
	}
}

func TestHandleAddSongToPlaylist_BadBody(t *testing.T) {
	logger := newTestLogger()
	repo := &fakePlaylistsRepo{}
	svc := NewPlaylistsService(repo)

	ctx := auth.WithUserID(context.Background(), "user-id")

	resp, err := HandleAddSongToPlaylist(ctx, svc, logger, openapi.AddSongToPlaylistRequestObject{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := resp.(openapi.AddSongToPlaylist400JSONResponse); !ok {
		t.Fatalf("expected 400 response, got %T", resp)
	}
}

func TestHandleRemoveSongFromPlaylist_Success(t *testing.T) {
	logger := newTestLogger()
	repo := &fakePlaylistsRepo{}
	svc := NewPlaylistsService(repo)

	ctx := auth.WithUserID(context.Background(), "user-id")

	resp, err := HandleRemoveSongFromPlaylist(ctx, svc, logger, openapi.RemoveSongFromPlaylistRequestObject{
		Id:     openapi_types.UUID(uuid.New()),
		SongId: openapi_types.UUID(uuid.New()),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := resp.(openapi.RemoveSongFromPlaylist204Response); !ok {
		t.Fatalf("expected 204 response, got %T", resp)
	}
}
