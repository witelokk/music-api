package songs

import (
	"context"
	"errors"
	"testing"
)

type fakeSongsRepo struct {
	song *Song
	err  error
}

func (r *fakeSongsRepo) GetSongWithFavorite(ctx context.Context, id, userID string) (*Song, bool, error) {
	if r.err != nil {
		return nil, false, r.err
	}
	return r.song, true, nil
}

func TestService_GetSongWithFavorite_Success(t *testing.T) {
	want := &Song{
		ID:              "song-id",
		Name:            "Test Song",
		DurationSeconds: 180,
		StreamMediaID:   "stream-id",
	}

	repo := &fakeSongsRepo{song: want}
	svc := NewService(repo)

	got, isFavorite, err := svc.GetSongWithFavorite(context.Background(), "song-id", "user-id")
	if err != nil {
		t.Fatalf("GetSongWithFavorite() error = %v, want nil", err)
	}
	if got != want {
		t.Fatalf("GetSongWithFavorite() = %#v, want %#v", got, want)
	}
	if !isFavorite {
		t.Fatalf("expected isFavorite=true")
	}
}

func TestService_GetSongWithFavorite_NotFound(t *testing.T) {
	repo := &fakeSongsRepo{err: ErrSongNotFound}
	svc := NewService(repo)

	_, _, err := svc.GetSongWithFavorite(context.Background(), "missing-id", "user-id")
	if !errors.Is(err, ErrSongNotFound) {
		t.Fatalf("expected ErrSongNotFound, got %v", err)
	}
}

func TestService_GetSongWithFavorite_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	repo := &fakeSongsRepo{err: repoErr}
	svc := NewService(repo)

	_, _, err := svc.GetSongWithFavorite(context.Background(), "song-id", "user-id")
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
