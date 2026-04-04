package artists

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

var ErrArtistNotFound = errors.New("artist not found")

func (s *Service) GetArtist(ctx context.Context, id string) (*Artist, error) {
	artist, err := s.repo.GetArtistByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrArtistNotFound
		}
		return nil, err
	}

	return artist, nil
}
