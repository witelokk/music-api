package favorites

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"

	"github.com/witelokk/music-api/internal/auth"
	openapi "github.com/witelokk/music-api/internal/openapi"
)

type fakeFavoritesRepo struct {
	songIDs []string
	err     error
}

func (r *fakeFavoritesRepo) AddFavorite(ctx context.Context, userID, songID string) error {
	return r.err
}

func (r *fakeFavoritesRepo) RemoveFavorite(ctx context.Context, userID, songID string) error {
	return r.err
}

func (r *fakeFavoritesRepo) GetFavoritesByUser(ctx context.Context, userID string) ([]string, error) {
	return r.songIDs, r.err
}
func (r *fakeFavoritesRepo) GetFavoriteSongs(ctx context.Context, userID string) ([]FavoriteSong, error) {
	return []FavoriteSong{}, r.err
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

func TestHandleGetFavorites_Empty(t *testing.T) {
	logger := newTestLogger()
	repo := &fakeFavoritesRepo{songIDs: []string{}}
	svc := NewFavoritesService(repo)

	req := openapi.GetFavoritesRequestObject{}

	ctx := auth.WithUserID(context.Background(), "user-id")

	resp, err := HandleGetFavorites(ctx, svc, logger, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	okResp, ok := resp.(openapi.GetFavorites200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}
	if okResp.Count != 0 || len(okResp.Songs) != 0 {
		t.Fatalf("expected empty favorites list, got count=%d songs=%d", okResp.Count, len(okResp.Songs))
	}

	raw, err := json.Marshal(okResp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}
	if string(raw) != "{\"count\":0,\"songs\":[]}" {
		t.Fatalf("unexpected json payload: %s", raw)
	}
}

func TestHandleAddFavorite_BadBody(t *testing.T) {
	logger := newTestLogger()
	svc := NewFavoritesService(&fakeFavoritesRepo{})

	ctx := auth.WithUserID(context.Background(), "user-id")

	resp, err := HandleAddFavorite(ctx, svc, logger, openapi.AddFavoriteRequestObject{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := resp.(openapi.AddFavorite400JSONResponse); !ok {
		t.Fatalf("expected 400 response, got %T", resp)
	}
}

func TestHandleRemoveFavorite_BadBody(t *testing.T) {
	logger := newTestLogger()
	svc := NewFavoritesService(&fakeFavoritesRepo{})

	ctx := auth.WithUserID(context.Background(), "user-id")

	resp, err := HandleRemoveFavorite(ctx, svc, logger, openapi.RemoveFavoriteRequestObject{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := resp.(openapi.RemoveFavorite400JSONResponse); !ok {
		t.Fatalf("expected 400 response, got %T", resp)
	}
}
