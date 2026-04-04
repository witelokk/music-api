package artists

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetArtistByID(ctx context.Context, id string) (*Artist, error)
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) GetArtistByID(ctx context.Context, id string) (*Artist, error) {
	const query = `
		SELECT id, name, avatar_url, cover_url
		FROM artists
		WHERE id = $1
	`

	var (
		name      string
		avatarURL *string
		coverURL  *string
	)

	if err := r.pool.
		QueryRow(ctx, query, id).
		Scan(&id, &name, &avatarURL, &coverURL); err != nil {
		return nil, err
	}

	return &Artist{
		ID:        id,
		Name:      name,
		AvatarURL: avatarURL,
		CoverURL:  coverURL,
	}, nil
}

