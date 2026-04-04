package releases

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetReleaseByID(ctx context.Context, id string) (*Release, error)
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) GetReleaseByID(ctx context.Context, id string) (*Release, error) {
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

