package releases

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetReleaseByID(ctx context.Context, id string) (*Release, error)
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
			CREATE TABLE IF NOT EXISTS releases (
				id UUID PRIMARY KEY,
				name TEXT NOT NULL,
				cover_url TEXT,
				type INT NOT NULL,
				release_at TIMESTAMP NOT NULL
			)
		`)
		r.initErr = err
	})
	return r.initErr
}

func (r *PostgresRepository) GetReleaseByID(ctx context.Context, id string) (*Release, error) {
	if err := r.ensureTable(ctx); err != nil {
		return nil, err
	}

	const query = `
		SELECT id, name, cover_url, type, release_at
		FROM releases
		WHERE id = $1
	`

	var (
		name      string
		coverURL  *string
		releaseAt time.Time
		typ       int
	)

	if err := r.pool.
		QueryRow(ctx, query, id).
		Scan(&id, &name, &coverURL, &typ, &releaseAt); err != nil {
		return nil, err
	}

	return &Release{
		ID:        id,
		Name:      name,
		CoverURL:  coverURL,
		Type:      typ,
		ReleaseAt: releaseAt,
	}, nil
}
