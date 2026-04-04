package releases

import (
	"context"
	"errors"
	"testing"
)

type fakeReleasesRepo struct {
	release *Release
	err     error
}

func (r *fakeReleasesRepo) GetReleaseByID(ctx context.Context, id string) (*Release, error) {
	return r.release, r.err
}

func TestService_GetRelease_Success(t *testing.T) {
	want := &Release{
		ID:   "release-id",
		Name: "Test Release",
	}

	repo := &fakeReleasesRepo{release: want}
	svc := NewService(repo)

	got, err := svc.GetRelease(context.Background(), "release-id")
	if err != nil {
		t.Fatalf("GetRelease() error = %v, want nil", err)
	}
	if got != want {
		t.Fatalf("GetRelease() = %#v, want %#v", got, want)
	}
}

func TestService_GetRelease_NotFound(t *testing.T) {
	repo := &fakeReleasesRepo{err: ErrReleaseNotFound}
	svc := NewService(repo)

	got, err := svc.GetRelease(context.Background(), "missing-id")
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

	got, err := svc.GetRelease(context.Background(), "release-id")
	if got != nil {
		t.Fatalf("expected nil release, got %#v", got)
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
