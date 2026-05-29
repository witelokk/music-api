package media

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

type fakeStorage struct {
	gotObjectName string
	reader        io.ReadCloser
	size          int64
	mime          string
	err           error
}

func (s *fakeStorage) GetObjectStream(ctx context.Context, objectName string) (io.ReadCloser, int64, string, error) {
	s.gotObjectName = objectName
	return s.reader, s.size, s.mime, s.err
}

func TestMediaServiceGetObjectStreamDelegatesToStorage(t *testing.T) {
	reader := io.NopCloser(strings.NewReader("media"))
	storage := &fakeStorage{reader: reader, size: 5, mime: "image/jpeg"}
	service := NewMediaService(storage)

	gotReader, gotSize, gotMime, err := service.GetObjectStream(context.Background(), "object-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if storage.gotObjectName != "object-id" {
		t.Fatalf("expected object name %q, got %q", "object-id", storage.gotObjectName)
	}
	if gotReader != reader {
		t.Fatalf("expected storage reader to be returned")
	}
	if gotSize != 5 {
		t.Fatalf("expected size 5, got %d", gotSize)
	}
	if gotMime != "image/jpeg" {
		t.Fatalf("expected mime %q, got %q", "image/jpeg", gotMime)
	}
}

func TestMediaServiceGetObjectStreamReturnsStorageError(t *testing.T) {
	wantErr := errors.New("storage failed")
	service := NewMediaService(&fakeStorage{err: wantErr})

	_, _, _, err := service.GetObjectStream(context.Background(), "object-id")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected storage error, got %v", err)
	}
}
