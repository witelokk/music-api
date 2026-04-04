package followings

import "context"

type FollowingsService struct {
	repo FollowingsRepository
}

func NewFollowingsService(repo FollowingsRepository) *FollowingsService {
	return &FollowingsService{repo: repo}
}

func (s *FollowingsService) Follow(ctx context.Context, userID, artistID string) error {
	return s.repo.Follow(ctx, userID, artistID)
}

func (s *FollowingsService) Unfollow(ctx context.Context, userID, artistID string) error {
	return s.repo.Unfollow(ctx, userID, artistID)
}

func (s *FollowingsService) GetFollowedArtists(ctx context.Context, userID string) ([]FollowedArtist, error) {
	return s.repo.GetFollowedArtists(ctx, userID)
}
