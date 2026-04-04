package songs

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetSongByID(ctx context.Context, id string) (*Song, error)
}

type PostgresRepository struct {
	pool     *pgxpool.Pool
	initOnce sync.Once
	initErr  error
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) ensureTable(ctx context.Context) error {
	r.initOnce.Do(func() {
		_, err := r.pool.Exec(ctx, `
			CREATE TABLE IF NOT EXISTS songs (
				id UUID PRIMARY KEY,
				name TEXT NOT NULL,
				cover_url TEXT,
				duration INT NOT NULL,
				stream_url TEXT NOT NULL
			)
		`)
		r.initErr = err
	})
	return r.initErr
}

func (r *PostgresRepository) GetSongByID(ctx context.Context, id string) (*Song, error) {
	if err := r.ensureTable(ctx); err != nil {
		return nil, err
	}

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
