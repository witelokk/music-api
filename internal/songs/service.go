package songs

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

var ErrSongNotFound = errors.New("song not found")

func (s *Service) GetSong(ctx context.Context, id string) (*Song, error) {
	song, err := s.repo.GetSongByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSongNotFound
		}
		return nil, err
	}

	return song, nil
}
