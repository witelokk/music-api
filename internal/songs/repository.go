package songs

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetSongByID(ctx context.Context, id string) (*Song, error)
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) GetSongByID(ctx context.Context, id string) (*Song, error) {
	const query = `
		SELECT id, name, cover_url, duration, stream_url
		FROM songs
		WHERE id = $1
	`

	var (
		name       string
		coverURL   *string
		duration   int
		streamURL  string
	)

	if err := r.pool.
		QueryRow(ctx, query, id).
		Scan(&id, &name, &coverURL, &duration, &streamURL); err != nil {
		return nil, err
	}

	return &Song{
		ID:              id,
		Name:            name,
		CoverURL:        coverURL,
		DurationSeconds: duration,
		StreamURL:       streamURL,
	}, nil
}

