package artists

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

type fakeArtistsRepo struct {
	artist *Artist
	err    error
}

func (r *fakeArtistsRepo) GetArtistByID(ctx context.Context, id string) (*Artist, error) {
	return r.artist, r.err
}

func TestService_GetArtist_Success(t *testing.T) {
	want := &Artist{
		ID:   "artist-id",
		Name: "Test Artist",
	}

	repo := &fakeArtistsRepo{artist: want}
	svc := NewService(repo)

	got, err := svc.GetArtist(context.Background(), "artist-id")
	if err != nil {
		t.Fatalf("GetArtist() error = %v, want nil", err)
	}
	if got != want {
		t.Fatalf("GetArtist() = %#v, want %#v", got, want)
	}
}

func TestService_GetArtist_NotFound(t *testing.T) {
	repo := &fakeArtistsRepo{err: pgx.ErrNoRows}
	svc := NewService(repo)

	got, err := svc.GetArtist(context.Background(), "missing-id")
	if got != nil {
		t.Fatalf("expected nil artist, got %#v", got)
	}
	if !errors.Is(err, ErrArtistNotFound) {
		t.Fatalf("expected ErrArtistNotFound, got %v", err)
	}
}

func TestService_GetArtist_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	repo := &fakeArtistsRepo{err: repoErr}
	svc := NewService(repo)

	got, err := svc.GetArtist(context.Background(), "artist-id")
	if got != nil {
		t.Fatalf("expected nil artist, got %#v", got)
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

