package songs

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SongsRepository interface {
	GetSongWithFavorite(ctx context.Context, id, userID string) (*Song, bool, error)
}

type PostgresSongsRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresSongsRepository(pool *pgxpool.Pool) *PostgresSongsRepository {
	return &PostgresSongsRepository{pool: pool}
}

func (r *PostgresSongsRepository) GetSongWithFavorite(ctx context.Context, id, userID string) (*Song, bool, error) {
	const songQuery = `
		SELECT s.id, s.name, s.cover_media_id, s.duration, s.stream_media_id,
		       EXISTS (
		         SELECT 1
		         FROM favorites f
		         WHERE f.user_id = $2 AND f.song_id = s.id
		       ) AS is_favorite
		FROM songs s
		WHERE s.id = $1
	`

	const artistsQuery = `
		SELECT a.id, a.name, a.avatar_media_id
		FROM song_artists sa
		JOIN artists a ON a.id = sa.artist_id
		WHERE sa.song_id = $1
		ORDER BY a.name
	`

	var (
		name          string
		coverMediaID  *string
		duration      int
		streamMediaID string
		isFavorite    bool
	)

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, false, err
	}
	defer tx.Rollback(ctx)

	if err := tx.
		QueryRow(ctx, songQuery, id, userID).
		Scan(&id, &name, &coverMediaID, &duration, &streamMediaID, &isFavorite); err != nil {
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
			artistID      string
			artistName    string
			avatarMediaID *string
		)
		if err := rows.Scan(&artistID, &artistName, &avatarMediaID); err != nil {
			return nil, false, err
		}
		artists = append(artists, ArtistSummary{
			ID:            artistID,
			Name:          artistName,
			AvatarMediaID: avatarMediaID,
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
		CoverMediaID:    coverMediaID,
		DurationSeconds: duration,
		StreamMediaID:   streamMediaID,
		Artists:         artists,
	}, isFavorite, nil
}
