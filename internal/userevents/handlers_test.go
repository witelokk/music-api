package userevents

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/witelokk/music-api/internal/auth"
	openapi "github.com/witelokk/music-api/internal/openapi"
)

type fakeUserEventsRepo struct {
	event UserEvent
	calls int
	err   error
}

func (r *fakeUserEventsRepo) Record(ctx context.Context, event UserEvent) error {
	r.calls++
	r.event = event
	return r.err
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

func TestHandleRecordUserEvent_MapsClientEventID(t *testing.T) {
	logger := newTestLogger()
	repo := &fakeUserEventsRepo{}
	svc := NewUserEventsService(repo)

	songID := uuid.New()
	clientEventID := uuid.New()
	body := openapi.RecordUserEventJSONRequestBody{
		EventType:     openapi.SongPlay,
		SongId:        openapi_types.UUID(songID),
		ClientEventId: openapi_types.UUID(clientEventID),
	}

	ctx := auth.WithUserID(context.Background(), "user-id")
	resp, err := HandleRecordUserEvent(ctx, svc, logger, openapi.RecordUserEventRequestObject{
		Body: &body,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := resp.(openapi.RecordUserEvent204Response); !ok {
		t.Fatalf("expected 204 response, got %T", resp)
	}
	if repo.calls != 1 {
		t.Fatalf("expected repository to be called once, got %d", repo.calls)
	}
	if repo.event.ClientEventID == nil || *repo.event.ClientEventID != clientEventID.String() {
		t.Fatalf("expected client event id %s, got %v", clientEventID, repo.event.ClientEventID)
	}
}

func TestHandleRecordUserEvent_InvalidEventTypeDoesNotRecord(t *testing.T) {
	logger := newTestLogger()
	repo := &fakeUserEventsRepo{}
	svc := NewUserEventsService(repo)

	body := openapi.RecordUserEventJSONRequestBody{
		EventType:     openapi.CreateUserEventRequestEventType("unknown"),
		SongId:        openapi_types.UUID(uuid.New()),
		ClientEventId: openapi_types.UUID(uuid.New()),
	}

	ctx := auth.WithUserID(context.Background(), "user-id")
	resp, err := HandleRecordUserEvent(ctx, svc, logger, openapi.RecordUserEventRequestObject{
		Body: &body,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := resp.(openapi.RecordUserEvent400JSONResponse); !ok {
		t.Fatalf("expected 400 response, got %T", resp)
	}
	if repo.calls != 0 {
		t.Fatalf("expected repository not to be called, got %d calls", repo.calls)
	}
}
