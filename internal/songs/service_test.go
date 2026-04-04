package songs

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

type fakeSongsRepo struct {
	song *Song
	err  error
}

func (r *fakeSongsRepo) GetSongByID(ctx context.Context, id string) (*Song, error) {
	return r.song, r.err
}

func TestService_GetSong_Success(t *testing.T) {
	want := &Song{
		ID:              "song-id",
		Name:            "Test Song",
		DurationSeconds: 180,
		StreamURL:       "https://example.com/stream",
	}

	repo := &fakeSongsRepo{song: want}
	svc := NewService(repo)

	got, err := svc.GetSong(context.Background(), "song-id")
	if err != nil {
		t.Fatalf("GetSong() error = %v, want nil", err)
	}
	if got != want {
		t.Fatalf("GetSong() = %#v, want %#v", got, want)
	}
}

func TestService_GetSong_NotFound(t *testing.T) {
	repo := &fakeSongsRepo{err: pgx.ErrNoRows}
	svc := NewService(repo)

	got, err := svc.GetSong(context.Background(), "missing-id")
	if got != nil {
		t.Fatalf("expected nil song, got %#v", got)
	}
	if !errors.Is(err, ErrSongNotFound) {
		t.Fatalf("expected ErrSongNotFound, got %v", err)
	}
}

func TestService_GetSong_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	repo := &fakeSongsRepo{err: repoErr}
	svc := NewService(repo)

	got, err := svc.GetSong(context.Background(), "song-id")
	if got != nil {
		t.Fatalf("expected nil song, got %#v", got)
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

