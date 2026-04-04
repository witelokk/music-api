package artists

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetArtistByID(ctx context.Context, id string) (*Artist, error)
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
			CREATE TABLE IF NOT EXISTS artists (
				id UUID PRIMARY KEY,
				name TEXT NOT NULL,
				avatar_url TEXT,
				cover_url TEXT
			)
		`)
		r.initErr = err
	})
	return r.initErr
}

func (r *PostgresRepository) GetArtistByID(ctx context.Context, id string) (*Artist, error) {
	if err := r.ensureTable(ctx); err != nil {
		return nil, err
	}

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
