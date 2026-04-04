package songs

import (
	"context"
	"errors"
)

type SongsService struct {
	repo SongsRepository
}

func NewService(repo SongsRepository) *SongsService {
	return &SongsService{repo: repo}
}

var ErrSongNotFound = errors.New("song not found")

func (s *SongsService) GetSongWithFavorite(ctx context.Context, id, userID string) (*Song, bool, error) {
	return s.repo.GetSongWithFavorite(ctx, id, userID)
}
