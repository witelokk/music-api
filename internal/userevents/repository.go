package userevents

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserEventsRepository interface {
	Record(ctx context.Context, event UserEvent) error
}

type PostgresUserEventsRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresUserEventsRepository(pool *pgxpool.Pool) *PostgresUserEventsRepository {
	return &PostgresUserEventsRepository{pool: pool}
}

func (r *PostgresUserEventsRepository) Record(ctx context.Context, event UserEvent) error {
	const query = `
		INSERT INTO user_events (
			user_id,
			event_type,
			song_id,
			artist_id,
			playlist_id,
			release_id,
			position_seconds,
			duration_seconds,
			percent_played,
			source,
			context_type,
			context_id,
			client_event_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (user_id, client_event_id) DO NOTHING
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		event.UserID,
		string(event.EventType),
		event.SongID,
		event.ArtistID,
		event.PlaylistID,
		event.ReleaseID,
		event.PositionSeconds,
		event.DurationSeconds,
		event.PercentPlayed,
		event.Source,
		event.ContextType,
		event.ContextID,
		event.ClientEventID,
	)
	if err != nil {
		if isSongForeignKeyViolation(err) {
			return ErrSongNotFound
		}
		return err
	}
	return nil
}

func isSongForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) &&
		pgErr.Code == "23503" &&
		pgErr.ConstraintName == "user_events_song_id_fkey"
}
