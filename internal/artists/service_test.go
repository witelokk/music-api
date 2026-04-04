package artists

import (
	"context"
	"errors"
	"testing"
)

type fakeArtistsRepo struct {
	artist *Artist
	err    error
}

func (r *fakeArtistsRepo) GetArtistWithStats(ctx context.Context, id, userID string) (*Artist, int, bool, error) {
	if r.err != nil {
		return nil, 0, false, r.err
	}
	return r.artist, 0, false, nil
}

func TestService_GetArtistWithStats_Success(t *testing.T) {
	want := &Artist{
		ID:   "artist-id",
		Name: "Test Artist",
	}

	repo := &fakeArtistsRepo{artist: want}
	svc := NewService(repo)

	got, followers, following, err := svc.GetArtistWithStats(context.Background(), "artist-id", "user-id")
	if err != nil {
		t.Fatalf("GetArtistWithStats() error = %v, want nil", err)
	}
	if got != want {
		t.Fatalf("GetArtistWithStats() artist = %#v, want %#v", got, want)
	}
	if followers != 0 || following {
		t.Fatalf("expected followers=0 and following=false, got followers=%d following=%v", followers, following)
	}
}

func TestService_GetArtistWithStats_NotFound(t *testing.T) {
	repo := &fakeArtistsRepo{err: ErrArtistNotFound}
	svc := NewService(repo)

	_, _, _, err := svc.GetArtistWithStats(context.Background(), "missing-id", "user-id")
	if !errors.Is(err, ErrArtistNotFound) {
		t.Fatalf("expected ErrArtistNotFound, got %v", err)
	}
}

func TestService_GetArtistWithStats_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	repo := &fakeArtistsRepo{err: repoErr}
	svc := NewService(repo)

	_, _, _, err := svc.GetArtistWithStats(context.Background(), "artist-id", "user-id")
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
