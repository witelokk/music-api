package releases

import (
	"context"
	"errors"
)

type ReleasesService struct {
	repo ReleasesRepository
}

func NewService(repo ReleasesRepository) *ReleasesService {
	return &ReleasesService{repo: repo}
}

func (s *ReleasesService) GetRelease(ctx context.Context, id string) (*Release, error) {
	return s.repo.GetReleaseByID(ctx, id)
}

var ErrReleaseNotFound = errors.New("release not found")
