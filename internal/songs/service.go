package songs

import "context"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetSong(ctx context.Context, id string) (*Song, error) {
	return s.repo.GetSongByID(ctx, id)
}

