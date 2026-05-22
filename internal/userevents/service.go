package userevents

import (
	"context"
	"errors"
)

const zeroUUID = "00000000-0000-0000-0000-000000000000"

var (
	ErrSongNotFound           = errors.New("song not found")
	ErrInvalidEventType       = errors.New("invalid event type")
	ErrSongIDRequired         = errors.New("song_id is required")
	ErrClientEventIDRequired  = errors.New("client_event_id is required")
	ErrInvalidPercentPlayed   = errors.New("percent_played must be between 0 and 100")
	ErrInvalidPositionSeconds = errors.New("position_seconds must be greater than or equal to 0")
	ErrInvalidDurationSeconds = errors.New("duration_seconds must be greater than or equal to 1")
	ErrPositionRequired       = errors.New("position_seconds is required")
	ErrInvalidContext         = errors.New("invalid context")
)

type UserEventsService struct {
	repository UserEventsRepository
}

func NewUserEventsService(repository UserEventsRepository) *UserEventsService {
	return &UserEventsService{repository: repository}
}

func (s *UserEventsService) RecordEvent(ctx context.Context, event UserEvent) error {
	if err := validateClientEvent(event); err != nil {
		return err
	}
	return s.repository.Record(ctx, event)
}

func validateClientEvent(event UserEvent) error {
	switch event.EventType {
	case EventTypeSongPlay, EventTypeSongSkip, EventTypeSongComplete:
	default:
		return ErrInvalidEventType
	}

	if isMissingID(event.SongID) {
		return ErrSongIDRequired
	}

	if isMissingID(event.ClientEventID) {
		return ErrClientEventIDRequired
	}

	if event.PercentPlayed != nil && (*event.PercentPlayed < 0 || *event.PercentPlayed > 100) {
		return ErrInvalidPercentPlayed
	}

	if event.PositionSeconds != nil && *event.PositionSeconds < 0 {
		return ErrInvalidPositionSeconds
	}

	if event.DurationSeconds != nil && *event.DurationSeconds < 1 {
		return ErrInvalidDurationSeconds
	}

	if event.EventType == EventTypeSongSkip && event.PositionSeconds == nil {
		return ErrPositionRequired
	}

	if event.ContextID != nil && !isMissingID(event.ContextID) && (event.ContextType == nil || *event.ContextType == "") {
		return ErrInvalidContext
	}

	if event.ContextType != nil {
		switch *event.ContextType {
		case "playlist", "release", "artist":
			if isMissingID(event.ContextID) {
				return ErrInvalidContext
			}
		case "queue", "recommendations":
		default:
			return ErrInvalidContext
		}
	}

	return nil
}

func isMissingID(id *string) bool {
	return id == nil || *id == "" || *id == zeroUUID
}
