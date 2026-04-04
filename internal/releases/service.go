package releases

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetRelease(ctx context.Context, id string) (*Release, error) {
	release, err := s.repo.GetReleaseByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrReleaseNotFound
		}
		return nil, err
	}

	return release, nil
}

var ErrReleaseNotFound = errors.New("release not found")
