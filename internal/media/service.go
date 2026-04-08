package media

import (
	"context"
	"errors"
	"io"
)

var ErrMediaNotFound = errors.New("media not found")

type Storage interface {
	GetObjectStream(ctx context.Context, objectName string) (io.ReadCloser, int64, error)
}

type MediaService struct {
	storage Storage
}

func NewMediaService(storage Storage) *MediaService {
	return &MediaService{
		storage: storage,
	}
}

func (s *MediaService) GetObjectStream(ctx context.Context, objectName string) (io.ReadCloser, int64, error) {
	return s.storage.GetObjectStream(ctx, objectName)
}
