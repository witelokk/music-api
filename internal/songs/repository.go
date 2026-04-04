package songs

import (
	"context"
	"errors"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SongsRepository interface {
	GetSongWithFavorite(ctx context.Context, id, userID string) (*Song, bool, error)
}

type PostgresSongsRepository struct {
	pool     *pgxpool.Pool
	initOnce sync.Once
	initErr  error
}

func NewPostgresSongsRepository(pool *pgxpool.Pool) *PostgresSongsRepository {
	return &PostgresSongsRepository{pool: pool}
}

func (r *PostgresSongsRepository) ensureTable(ctx context.Context) error {
	r.initOnce.Do(func() {
		_, err := r.pool.Exec(ctx, `
			CREATE TABLE IF NOT EXISTS songs (
				id UUID PRIMARY KEY,
				name TEXT NOT NULL,
				cover_url TEXT,
				duration INT NOT NULL,
				stream_url TEXT NOT NULL,
				streams_count BIGINT NOT NULL DEFAULT 0
			);

			CREATE TABLE IF NOT EXISTS song_artists (
				song_id UUID NOT NULL,
				artist_id UUID NOT NULL,
				PRIMARY KEY (song_id, artist_id)
			)
		`)
		r.initErr = err
	})
	return r.initErr
}

func (r *PostgresSongsRepository) GetSongWithFavorite(ctx context.Context, id, userID string) (*Song, bool, error) {
	if err := r.ensureTable(ctx); err != nil {
		return nil, false, err
	}

	const songQuery = `
		SELECT s.id, s.name, s.cover_url, s.duration, s.stream_url,
		       EXISTS (
		         SELECT 1
		         FROM favorites f
		         WHERE f.user_id = $2 AND f.song_id = s.id
		       ) AS is_favorite
		FROM songs s
		WHERE s.id = $1
	`

	const artistsQuery = `
		SELECT a.id, a.name, a.avatar_url
		FROM song_artists sa
		JOIN artists a ON a.id = sa.artist_id
		WHERE sa.song_id = $1
		ORDER BY a.name
	`

	var (
		name       string
		coverURL   *string
		duration   int
		streamURL  string
		isFavorite bool
	)

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, false, err
	}
	defer tx.Rollback(ctx)

	if err := tx.
		QueryRow(ctx, songQuery, id, userID).
		Scan(&id, &name, &coverURL, &duration, &streamURL, &isFavorite); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, ErrSongNotFound
		}
		return nil, false, err
	}

	rows, err := tx.Query(ctx, artistsQuery, id)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	var artists []ArtistSummary
	for rows.Next() {
		var (
			artistID   string
			artistName string
			avatarURL  *string
		)
		if err := rows.Scan(&artistID, &artistName, &avatarURL); err != nil {
			return nil, false, err
		}
		artists = append(artists, ArtistSummary{
			ID:        artistID,
			Name:      artistName,
			AvatarURL: avatarURL,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}

	return &Song{
		ID:              id,
		Name:            name,
		CoverURL:        coverURL,
		DurationSeconds: duration,
		StreamURL:       streamURL,
		Artists:         artists,
	}, isFavorite, nil
}
