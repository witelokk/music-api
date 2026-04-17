package releases

import (
	"context"
	"errors"
	"testing"
)

type fakeReleasesRepo struct {
	release *Release
	err     error
	userID  string
	id      string
}

func (r *fakeReleasesRepo) GetReleaseByID(ctx context.Context, userID, id string) (*Release, error) {
	r.userID = userID
	r.id = id
	return r.release, r.err
}

func (r *fakeReleasesRepo) GetRandomReleases(ctx context.Context, seed string, limit int) ([]Release, error) {
	return nil, nil
}

func TestService_GetRelease_Success(t *testing.T) {
	want := &Release{
		ID:   "release-id",
		Name: "Test Release",
	}

	repo := &fakeReleasesRepo{release: want}
	svc := NewService(repo)

	got, err := svc.GetRelease(context.Background(), "user-id", "release-id")
	if err != nil {
		t.Fatalf("GetRelease() error = %v, want nil", err)
	}
	if got != want {
		t.Fatalf("GetRelease() = %#v, want %#v", got, want)
	}
	if repo.userID != "user-id" || repo.id != "release-id" {
		t.Fatalf("GetRelease() called repo with userID=%q id=%q", repo.userID, repo.id)
	}
}

func TestService_GetRelease_NotFound(t *testing.T) {
	repo := &fakeReleasesRepo{err: ErrReleaseNotFound}
	svc := NewService(repo)

	got, err := svc.GetRelease(context.Background(), "user-id", "missing-id")
	if got != nil {
		t.Fatalf("expected nil release, got %#v", got)
	}
	if !errors.Is(err, ErrReleaseNotFound) {
		t.Fatalf("expected ErrReleaseNotFound, got %v", err)
	}
}

func TestService_GetRelease_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	repo := &fakeReleasesRepo{err: repoErr}
	svc := NewService(repo)

	got, err := svc.GetRelease(context.Background(), "user-id", "release-id")
	if got != nil {
		t.Fatalf("expected nil release, got %#v", got)
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
