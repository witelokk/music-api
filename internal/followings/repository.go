package followings

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FollowingsRepository interface {
	Follow(ctx context.Context, userID, artistID string) error
	Unfollow(ctx context.Context, userID, artistID string) error
	GetFollowedArtists(ctx context.Context, userID string) ([]FollowedArtist, error)
}

type PostgresFollowingsRepository struct {
	pool *pgxpool.Pool
}

type FollowedArtist struct {
	ID            string
	Name          string
	AvatarMediaID *string
}

func NewPostgresFollowingsRepository(pool *pgxpool.Pool) *PostgresFollowingsRepository {
	return &PostgresFollowingsRepository{pool: pool}
}

func (r *PostgresFollowingsRepository) Follow(ctx context.Context, userID, artistID string) error {
	const query = `
		INSERT INTO followings (user_id, artist_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, artist_id) DO NOTHING
	`

	const event_query = `
		INSERT INTO user_events (user_id, event_type, artist_id, event_time)
		VALUES ($1, 'artist_follow', $2, NOW())
	`

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	cmd, err := tx.Exec(ctx, query, userID, artistID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return tx.Commit(ctx)
	}

	_, err = tx.Exec(ctx, event_query, userID, artistID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *PostgresFollowingsRepository) Unfollow(ctx context.Context, userID, artistID string) error {
	const query = `
		DELETE FROM followings
		WHERE user_id = $1 AND artist_id = $2
	`

	const event_query = `
		INSERT INTO user_events (user_id, event_type, artist_id, event_time)
		VALUES ($1, 'artist_unfollow', $2, NOW())
	`

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	cmd, err := tx.Exec(ctx, query, userID, artistID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return tx.Commit(ctx)
	}

	_, err = tx.Exec(ctx, event_query, userID, artistID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *PostgresFollowingsRepository) GetFollowedArtists(ctx context.Context, userID string) ([]FollowedArtist, error) {
	const query = `
		SELECT a.id, a.name, a.avatar_media_id
		FROM followings f
		JOIN artists a ON a.id = f.artist_id
		WHERE f.user_id = $1
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []FollowedArtist
	for rows.Next() {
		var (
			id            string
			name          string
			avatarMediaID *string
		)
		if err := rows.Scan(&id, &name, &avatarMediaID); err != nil {
			return nil, err
		}
		result = append(result, FollowedArtist{
			ID:            id,
			Name:          name,
			AvatarMediaID: avatarMediaID,
		})
	}

	return result, rows.Err()
}
