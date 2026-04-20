package favorites

import (
	"context"
	"errors"
)

var ErrSongNotFound = errors.New("song not found")

type FavoritesService struct {
	repo Repository
}

func NewFavoritesService(repo Repository) *FavoritesService {
	return &FavoritesService{repo: repo}
}

func (s *FavoritesService) AddFavorite(ctx context.Context, userID, songID string) error {
	return s.repo.AddFavorite(ctx, userID, songID)
}

func (s *FavoritesService) RemoveFavorite(ctx context.Context, userID, songID string) error {
	return s.repo.RemoveFavorite(ctx, userID, songID)
}

func (s *FavoritesService) GetFavoriteSongs(ctx context.Context, userID string) ([]FavoriteSong, error) {
	return s.repo.GetFavoriteSongs(ctx, userID)
}
