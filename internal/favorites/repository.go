package favorites

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	AddFavorite(ctx context.Context, userID, songID string) error
	RemoveFavorite(ctx context.Context, userID, songID string) error
	GetFavoriteSongs(ctx context.Context, userID string) ([]FavoriteSong, error)
}

type PostgresFavoritesRepository struct {
	pool     *pgxpool.Pool
	initOnce sync.Once
	initErr  error
}

type FavoriteSong struct {
	ID              string
	Name            string
	CoverURL        *string
	DurationSeconds int
	StreamURL       string
}

func NewPostgresFavoritesRepository(pool *pgxpool.Pool) *PostgresFavoritesRepository {
	return &PostgresFavoritesRepository{pool: pool}
}

func (r *PostgresFavoritesRepository) ensureTable(ctx context.Context) error {
	r.initOnce.Do(func() {
		_, err := r.pool.Exec(ctx, `
			CREATE TABLE IF NOT EXISTS favorites (
				user_id UUID NOT NULL,
				song_id UUID NOT NULL,
				added_at TIMESTAMP NOT NULL DEFAULT NOW(),
				PRIMARY KEY (user_id, song_id)
			)
		`)
		r.initErr = err
	})
	return r.initErr
}

func (r *PostgresFavoritesRepository) AddFavorite(ctx context.Context, userID, songID string) error {
	if err := r.ensureTable(ctx); err != nil {
		return err
	}

	const query = `
		INSERT INTO favorites (user_id, song_id, added_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id, song_id) DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query, userID, songID)
	return err
}

func (r *PostgresFavoritesRepository) RemoveFavorite(ctx context.Context, userID, songID string) error {
	if err := r.ensureTable(ctx); err != nil {
		return err
	}

	const query = `
		DELETE FROM favorites
		WHERE user_id = $1 AND song_id = $2
	`

	_, err := r.pool.Exec(ctx, query, userID, songID)
	return err
}

func (r *PostgresFavoritesRepository) GetFavoriteSongs(ctx context.Context, userID string) ([]FavoriteSong, error) {
	if err := r.ensureTable(ctx); err != nil {
		return nil, err
	}

	const query = `
		SELECT s.id, s.name, s.cover_url, s.duration, s.stream_url
		FROM favorites f
		JOIN songs s ON s.id = f.song_id
		WHERE f.user_id = $1
		ORDER BY f.added_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []FavoriteSong
	for rows.Next() {
		var (
			id       string
			name     string
			coverURL *string
			duration int
			stream   string
		)
		if err := rows.Scan(&id, &name, &coverURL, &duration, &stream); err != nil {
			return nil, err
		}
		result = append(result, FavoriteSong{
			ID:              id,
			Name:            name,
			CoverURL:        coverURL,
			DurationSeconds: duration,
			StreamURL:       stream,
		})
	}

	return result, rows.Err()
}
