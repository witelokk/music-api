package releases

import "context"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetRelease(ctx context.Context, id string) (*Release, error) {
	return s.repo.GetReleaseByID(ctx, id)
}

