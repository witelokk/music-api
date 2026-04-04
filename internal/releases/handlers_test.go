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

	svcWrapper := &Service{repo: &fakeReleasesRepo{err: ErrReleaseNotFound}}

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
		ID:        id.String(),
		Name:      "Test Release",
		Type:      1,
		ReleaseAt: time.Now(),
	}

	svc := &Service{repo: &fakeReleasesRepo{release: rel}}

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
}

