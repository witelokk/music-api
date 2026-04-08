package followings

import (
	"context"

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
	ID        string
	Name      string
	AvatarURL *string
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

	_, err := r.pool.Exec(ctx, query, userID, artistID)
	return err
}

func (r *PostgresFollowingsRepository) Unfollow(ctx context.Context, userID, artistID string) error {
	const query = `
		DELETE FROM followings
		WHERE user_id = $1 AND artist_id = $2
	`

	_, err := r.pool.Exec(ctx, query, userID, artistID)
	return err
}

func (r *PostgresFollowingsRepository) GetFollowedArtists(ctx context.Context, userID string) ([]FollowedArtist, error) {
	const query = `
		SELECT a.id, a.name, a.avatar_url
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
			id        string
			name      string
			avatarURL *string
		)
		if err := rows.Scan(&id, &name, &avatarURL); err != nil {
			return nil, err
		}
		result = append(result, FollowedArtist{
			ID:        id,
			Name:      name,
			AvatarURL: avatarURL,
		})
	}

	return result, rows.Err()
}
