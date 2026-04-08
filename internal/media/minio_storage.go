package media

import (
	"context"
	"errors"
	"io"

	"github.com/minio/minio-go/v7"
)

type MinioStorage struct {
	client *minio.Client
	bucket string
}

func NewMinioStorage(client *minio.Client, bucket string) Storage {
	return &MinioStorage{
		client: client,
		bucket: bucket,
	}
}

func (s *MinioStorage) GetObjectStream(ctx context.Context, objectName string) (io.ReadCloser, int64, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, 0, err
	}

	info, err := obj.Stat()
	if err != nil {
		var minioErr minio.ErrorResponse
		if errors.As(err, &minioErr) && minioErr.Code == "NoSuchKey" {
			return nil, 0, ErrMediaNotFound
		}
		return nil, 0, err
	}

	return obj, info.Size, nil
}

