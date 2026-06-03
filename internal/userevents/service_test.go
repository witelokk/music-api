package userevents

import (
	"context"
	"errors"
	"testing"
)

type fakeRecordRepo struct {
	calls int
	err   error
}

func (r *fakeRecordRepo) Record(ctx context.Context, event UserEvent) error {
	r.calls++
	return r.err
}

type fakeFeedRefreshQueue struct {
	calls  int
	userID string
	reason string
	err    error
}

func (q *fakeFeedRefreshQueue) Enqueue(ctx context.Context, userID, reason string) error {
	q.calls++
	q.userID = userID
	q.reason = reason
	return q.err
}

func TestRecordEvent_RequiresClientEventID(t *testing.T) {
	repo := &fakeRecordRepo{}
	service := NewUserEventsService(repo)
	songID := "song-id"

	err := service.RecordEvent(context.Background(), UserEvent{
		UserID:    "user-id",
		EventType: EventTypeSongPlay,
		SongID:    &songID,
	})

	if !errors.Is(err, ErrClientEventIDRequired) {
		t.Fatalf("expected ErrClientEventIDRequired, got %v", err)
	}
	if repo.calls != 0 {
		t.Fatalf("expected repository not to be called, got %d calls", repo.calls)
	}
}

func TestRecordEvent_RequiresPositionForSkip(t *testing.T) {
	repo := &fakeRecordRepo{}
	service := NewUserEventsService(repo)
	songID := "song-id"
	clientEventID := "client-event-id"

	err := service.RecordEvent(context.Background(), UserEvent{
		UserID:        "user-id",
		EventType:     EventTypeSongSkip,
		SongID:        &songID,
		ClientEventID: &clientEventID,
	})

	if !errors.Is(err, ErrPositionRequired) {
		t.Fatalf("expected ErrPositionRequired, got %v", err)
	}
	if repo.calls != 0 {
		t.Fatalf("expected repository not to be called, got %d calls", repo.calls)
	}
}

func TestRecordEvent_AllowsValidPlaybackEvent(t *testing.T) {
	repo := &fakeRecordRepo{}
	service := NewUserEventsService(repo)
	songID := "song-id"
	clientEventID := "client-event-id"

	err := service.RecordEvent(context.Background(), UserEvent{
		UserID:        "user-id",
		EventType:     EventTypeSongPlay,
		SongID:        &songID,
		ClientEventID: &clientEventID,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.calls != 1 {
		t.Fatalf("expected repository to be called once, got %d calls", repo.calls)
	}
}

func TestRecordEvent_EnqueuesFeedRefreshAfterRecordingEvent(t *testing.T) {
	repo := &fakeRecordRepo{}
	queue := &fakeFeedRefreshQueue{}
	service := NewUserEventsServiceWithFeedRefreshQueue(repo, queue)
	songID := "song-id"
	clientEventID := "client-event-id"

	err := service.RecordEvent(context.Background(), UserEvent{
		UserID:        "user-id",
		EventType:     EventTypeSongPlay,
		SongID:        &songID,
		ClientEventID: &clientEventID,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if queue.calls != 1 {
		t.Fatalf("expected feed refresh to be enqueued once, got %d calls", queue.calls)
	}
	if queue.userID != "user-id" {
		t.Fatalf("expected feed refresh user user-id, got %q", queue.userID)
	}
	if queue.reason != "user_event" {
		t.Fatalf("expected feed refresh reason user_event, got %q", queue.reason)
	}
}

func TestRecordEvent_ValidationDoesNotRecord(t *testing.T) {
	validSongID := "song-id"
	validClientEventID := "client-event-id"
	zeroID := zeroUUID
	negativePosition := -1
	zeroDuration := 0
	tooHighPercent := 101.0
	contextID := "context-id"

	tests := []struct {
		name  string
		event UserEvent
		want  error
	}{
		{
			name: "invalid event type",
			event: UserEvent{
				UserID:        "user-id",
				EventType:     EventTypeSongFavorite,
				SongID:        &validSongID,
				ClientEventID: &validClientEventID,
			},
			want: ErrInvalidEventType,
		},
		{
			name: "missing song id",
			event: UserEvent{
				UserID:        "user-id",
				EventType:     EventTypeSongPlay,
				ClientEventID: &validClientEventID,
			},
			want: ErrSongIDRequired,
		},
		{
			name: "zero song id",
			event: UserEvent{
				UserID:        "user-id",
				EventType:     EventTypeSongPlay,
				SongID:        &zeroID,
				ClientEventID: &validClientEventID,
			},
			want: ErrSongIDRequired,
		},
		{
			name: "zero client event id",
			event: UserEvent{
				UserID:        "user-id",
				EventType:     EventTypeSongPlay,
				SongID:        &validSongID,
				ClientEventID: &zeroID,
			},
			want: ErrClientEventIDRequired,
		},
		{
			name: "invalid percent played",
			event: UserEvent{
				UserID:        "user-id",
				EventType:     EventTypeSongPlay,
				SongID:        &validSongID,
				ClientEventID: &validClientEventID,
				PercentPlayed: &tooHighPercent,
			},
			want: ErrInvalidPercentPlayed,
		},
		{
			name: "invalid position seconds",
			event: UserEvent{
				UserID:          "user-id",
				EventType:       EventTypeSongPlay,
				SongID:          &validSongID,
				ClientEventID:   &validClientEventID,
				PositionSeconds: &negativePosition,
			},
			want: ErrInvalidPositionSeconds,
		},
		{
			name: "invalid duration seconds",
			event: UserEvent{
				UserID:          "user-id",
				EventType:       EventTypeSongPlay,
				SongID:          &validSongID,
				ClientEventID:   &validClientEventID,
				DurationSeconds: &zeroDuration,
			},
			want: ErrInvalidDurationSeconds,
		},
		{
			name: "context id without context type",
			event: UserEvent{
				UserID:        "user-id",
				EventType:     EventTypeSongPlay,
				SongID:        &validSongID,
				ClientEventID: &validClientEventID,
				ContextID:     &contextID,
			},
			want: ErrInvalidContext,
		},
		{
			name: "entity context without context id",
			event: UserEvent{
				UserID:        "user-id",
				EventType:     EventTypeSongPlay,
				SongID:        &validSongID,
				ClientEventID: &validClientEventID,
				ContextType:   stringPtr("playlist"),
			},
			want: ErrInvalidContext,
		},
		{
			name: "invalid context type",
			event: UserEvent{
				UserID:        "user-id",
				EventType:     EventTypeSongPlay,
				SongID:        &validSongID,
				ClientEventID: &validClientEventID,
				ContextType:   stringPtr("unknown"),
			},
			want: ErrInvalidContext,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRecordRepo{}
			service := NewUserEventsService(repo)

			err := service.RecordEvent(context.Background(), tt.event)

			if !errors.Is(err, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, err)
			}
			if repo.calls != 0 {
				t.Fatalf("expected repository not to be called, got %d calls", repo.calls)
			}
		})
	}
}

func TestRecordEvent_AllowsValidContexts(t *testing.T) {
	validSongID := "song-id"
	validClientEventID := "client-event-id"
	contextID := "context-id"

	tests := []struct {
		name  string
		event UserEvent
	}{
		{
			name: "queue without context id",
			event: UserEvent{
				UserID:        "user-id",
				EventType:     EventTypeSongPlay,
				SongID:        &validSongID,
				ClientEventID: &validClientEventID,
				ContextType:   stringPtr("queue"),
			},
		},
		{
			name: "playlist with context id",
			event: UserEvent{
				UserID:        "user-id",
				EventType:     EventTypeSongPlay,
				SongID:        &validSongID,
				ClientEventID: &validClientEventID,
				ContextType:   stringPtr("playlist"),
				ContextID:     &contextID,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRecordRepo{}
			service := NewUserEventsService(repo)

			err := service.RecordEvent(context.Background(), tt.event)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if repo.calls != 1 {
				t.Fatalf("expected repository to be called once, got %d calls", repo.calls)
			}
		})
	}
}

func stringPtr(value string) *string {
	return &value
}
