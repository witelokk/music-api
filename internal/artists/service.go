package artists

import (
	"context"
	"errors"
)

type ArtistsService struct {
	repo ArtistsRepository
}

func NewService(repo ArtistsRepository) *ArtistsService {
	return &ArtistsService{repo: repo}
}

var ErrArtistNotFound = errors.New("artist not found")

func (s *ArtistsService) GetArtistWithStats(ctx context.Context, id, userID string) (artist *Artist, followers int, isFollowing bool, err error) {
	return s.repo.GetArtistWithStats(ctx, id, userID)
}
