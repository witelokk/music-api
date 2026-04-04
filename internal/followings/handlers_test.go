package followings

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/witelokk/music-api/internal/auth"
	openapi "github.com/witelokk/music-api/internal/openapi"
)

type fakeFollowingsRepo struct {
	ids []string
	err error
}

func (r *fakeFollowingsRepo) Follow(ctx context.Context, userID, artistID string) error {
	return r.err
}

func (r *fakeFollowingsRepo) Unfollow(ctx context.Context, userID, artistID string) error {
	return r.err
}

func (r *fakeFollowingsRepo) GetFollowedArtists(ctx context.Context, userID string) ([]FollowedArtist, error) {
	return []FollowedArtist{}, r.err
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

func TestHandleGetFollowings_Empty(t *testing.T) {
	logger := newTestLogger()
	repo := &fakeFollowingsRepo{ids: []string{}}
	svc := NewFollowingsService(repo)

	req := openapi.GetFollowingsRequestObject{}
	ctx := auth.WithUserID(context.Background(), "user-id")

	resp, err := HandleGetFollowings(ctx, svc, logger, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	okResp, ok := resp.(openapi.GetFollowings200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}
	if okResp.Count != 0 || len(okResp.Artists) != 0 {
		t.Fatalf("expected empty followings list, got count=%d artists=%d", okResp.Count, len(okResp.Artists))
	}
}

func TestHandleFollowArtist_BadBody(t *testing.T) {
	logger := newTestLogger()
	svc := NewFollowingsService(&fakeFollowingsRepo{})

	ctx := auth.WithUserID(context.Background(), "user-id")
	resp, err := HandleFollowArtist(ctx, svc, logger, openapi.FollowArtistRequestObject{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := resp.(openapi.FollowArtist400JSONResponse); !ok {
		t.Fatalf("expected 400 response, got %T", resp)
	}
}

func TestHandleUnfollowArtist_BadBody(t *testing.T) {
	logger := newTestLogger()
	svc := NewFollowingsService(&fakeFollowingsRepo{})

	ctx := auth.WithUserID(context.Background(), "user-id")
	resp, err := HandleUnfollowArtist(ctx, svc, logger, openapi.UnfollowArtistRequestObject{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := resp.(openapi.UnfollowArtist400JSONResponse); !ok {
		t.Fatalf("expected 400 response, got %T", resp)
	}
}
