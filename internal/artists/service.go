package artists

import "context"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetArtist(ctx context.Context, id string) (*Artist, error) {
	return s.repo.GetArtistByID(ctx, id)
}

