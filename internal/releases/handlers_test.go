package releases

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	openapi "github.com/witelokk/music-api/internal/openapi"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

type fakeReleasesService struct {
	release *Release
	err     error
}

func (s *fakeReleasesService) GetRelease(ctx context.Context, id string) (*Release, error) {
	return s.release, s.err
}

func TestHandleGetRelease_NotFound(t *testing.T) {
	logger := newTestLogger()
	id := uuid.New()
	req := openapi.GetReleaseRequestObject{Id: id}

	svcWrapper := &ReleasesService{repo: &fakeReleasesRepo{err: ErrReleaseNotFound}}

	resp, err := HandleGetRelease(context.Background(), svcWrapper, logger, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errResp, ok := resp.(openapi.GetRelease404JSONResponse)
	if !ok {
		t.Fatalf("expected 404 response, got %T", resp)
	}
	if errResp.Error != "release not found" {
		t.Fatalf("expected error %q, got %q", "release not found", errResp.Error)
	}
}

func TestHandleGetRelease_Success(t *testing.T) {
	logger := newTestLogger()
	id := uuid.New()

	rel := &Release{
		ID:   id.String(),
		Name: "Test Release",
		Type: 1,
		Songs: []ReleaseSong{
			{
				ID:              uuid.New().String(),
				Name:            "Test Song",
				DurationSeconds: 180,
				StreamMediaID:   "stream-id",
				Artists: []ReleaseArtist{
					{
						ID:   uuid.New().String(),
						Name: "Artist 1",
					},
				},
			},
		},
		ReleaseAt: time.Now(),
	}

	svc := &ReleasesService{repo: &fakeReleasesRepo{release: rel}}

	req := openapi.GetReleaseRequestObject{Id: id}

	resp, err := HandleGetRelease(context.Background(), svc, logger, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	okResp, ok := resp.(openapi.GetRelease200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}
	if okResp.Name != rel.Name {
		t.Fatalf("expected name %q, got %q", rel.Name, okResp.Name)
	}

	if len(okResp.Songs.Songs) != 1 {
		t.Fatalf("expected 1 song, got %d", len(okResp.Songs.Songs))
	}
	if okResp.Songs.Songs[0].StreamUrl != "/media/stream-id" {
		t.Fatalf("expected stream url %q, got %q", "/media/stream-id", okResp.Songs.Songs[0].StreamUrl)
	}
	if len(okResp.Songs.Songs[0].Artists) != 1 {
		t.Fatalf("expected 1 artist for song, got %d", len(okResp.Songs.Songs[0].Artists))
	}
	if okResp.Songs.Songs[0].Artists[0].Name != "Artist 1" {
		t.Fatalf("expected artist name %q, got %q", "Artist 1", okResp.Songs.Songs[0].Artists[0].Name)
	}
}
