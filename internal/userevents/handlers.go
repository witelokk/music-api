package userevents

import (
	"context"
	"errors"
	"log/slog"

	"github.com/witelokk/music-api/internal/auth"
	openapi "github.com/witelokk/music-api/internal/openapi"
	"github.com/witelokk/music-api/internal/requestctx"
)

func HandleRecordUserEvent(ctx context.Context, service *UserEventsService, logger *slog.Logger, request openapi.RecordUserEventRequestObject) (openapi.RecordUserEventResponseObject, error) {
	reqLogger := requestctx.LoggerFromContext(ctx, logger)

	if request.Body == nil {
		return openapi.RecordUserEvent400JSONResponse(openapi.Error{Error: "invalid request body"}), nil
	}

	userID := auth.UserIDFromContext(ctx)
	if userID == "" {
		return openapi.RecordUserEvent500JSONResponse(openapi.Error{Error: "failed to record user event"}), nil
	}

	if !request.Body.EventType.Valid() {
		return openapi.RecordUserEvent400JSONResponse(openapi.Error{Error: "invalid event_type"}), nil
	}

	if request.Body.PositionSeconds != nil && *request.Body.PositionSeconds < 0 {
		return openapi.RecordUserEvent400JSONResponse(openapi.Error{Error: "position_seconds must be greater than or equal to 0"}), nil
	}

	if request.Body.DurationSeconds != nil && *request.Body.DurationSeconds < 1 {
		return openapi.RecordUserEvent400JSONResponse(openapi.Error{Error: "duration_seconds must be greater than or equal to 1"}), nil
	}

	if request.Body.PercentPlayed != nil && (*request.Body.PercentPlayed < 0 || *request.Body.PercentPlayed > 100) {
		return openapi.RecordUserEvent400JSONResponse(openapi.Error{Error: "percent_played must be between 0 and 100"}), nil
	}

	songID := request.Body.SongId.String()
	clientEventID := request.Body.ClientEventId.String()
	var contextType *string
	if request.Body.ContextType != nil {
		if !request.Body.ContextType.Valid() {
			return openapi.RecordUserEvent400JSONResponse(openapi.Error{Error: "invalid context_type"}), nil
		}
		value := string(*request.Body.ContextType)
		contextType = &value
	}

	event := UserEvent{
		UserID:          userID,
		EventType:       EventType(request.Body.EventType),
		SongID:          &songID,
		PositionSeconds: request.Body.PositionSeconds,
		DurationSeconds: request.Body.DurationSeconds,
		Source:          request.Body.Source,
		ContextType:     contextType,
		ClientEventID:   &clientEventID,
	}

	if request.Body.PercentPlayed != nil {
		percentPlayed := float64(*request.Body.PercentPlayed)
		event.PercentPlayed = &percentPlayed
	}

	if request.Body.ContextId != nil {
		contextID := request.Body.ContextId.String()
		event.ContextID = &contextID
	}

	if err := service.RecordEvent(ctx, event); err != nil {
		if errors.Is(err, ErrSongNotFound) {
			return openapi.RecordUserEvent400JSONResponse(openapi.Error{Error: "song not found"}), nil
		}
		if isValidationError(err) {
			return openapi.RecordUserEvent400JSONResponse(openapi.Error{Error: err.Error()}), nil
		}

		reqLogger.Error("failed to record user event",
			slog.String("user_id", userID),
			slog.String("song_id", songID),
			slog.String("event_type", string(request.Body.EventType)),
			slog.String("error", err.Error()),
		)
		return openapi.RecordUserEvent500JSONResponse(openapi.Error{Error: "failed to record user event"}), nil
	}

	return openapi.RecordUserEvent204Response{}, nil
}

func isValidationError(err error) bool {
	return errors.Is(err, ErrInvalidEventType) ||
		errors.Is(err, ErrSongIDRequired) ||
		errors.Is(err, ErrClientEventIDRequired) ||
		errors.Is(err, ErrInvalidPercentPlayed) ||
		errors.Is(err, ErrInvalidPositionSeconds) ||
		errors.Is(err, ErrInvalidDurationSeconds) ||
		errors.Is(err, ErrPositionRequired) ||
		errors.Is(err, ErrInvalidContext)
}
