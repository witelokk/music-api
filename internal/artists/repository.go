package artists

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ArtistsRepository interface {
	GetArtistWithStats(ctx context.Context, id, userID string) (*Artist, int, bool, error)
}

type PostgresArtistsRepository struct {
	pool     *pgxpool.Pool
	initOnce sync.Once
	initErr  error
}

func NewPostgresArtistsRepository(pool *pgxpool.Pool) *PostgresArtistsRepository {
	return &PostgresArtistsRepository{pool: pool}
}

func (r *PostgresArtistsRepository) ensureTable(ctx context.Context) error {
	r.initOnce.Do(func() {
		_, err := r.pool.Exec(ctx, `
			CREATE TABLE IF NOT EXISTS artists (
				id UUID PRIMARY KEY,
				name TEXT NOT NULL,
				avatar_url TEXT,
				cover_url TEXT
			);

			CREATE TABLE IF NOT EXISTS song_artists (
				song_id UUID NOT NULL,
				artist_id UUID NOT NULL,
				PRIMARY KEY (song_id, artist_id)
			);

			CREATE TABLE IF NOT EXISTS release_songs (
				release_id UUID NOT NULL,
				song_id UUID NOT NULL,
				PRIMARY KEY (release_id, song_id)
			)
		`)
		r.initErr = err
	})
	return r.initErr
}

func (r *PostgresArtistsRepository) GetArtistWithStats(ctx context.Context, id, userID string) (*Artist, int, bool, error) {
	if err := r.ensureTable(ctx); err != nil {
		return nil, 0, false, err
	}

	const artistQuery = `
		SELECT a.id,
		       a.name,
		       a.avatar_url,
		       a.cover_url,
		       (SELECT COUNT(*) FROM followings f WHERE f.artist_id = a.id) AS followers,
		       EXISTS (
		         SELECT 1
		         FROM followings f2
		         WHERE f2.artist_id = a.id AND f2.user_id = $2
		       ) AS is_following
		FROM artists a
		WHERE a.id = $1
	`

	const popularSongsQuery = `
		SELECT s.id, s.name, s.cover_url, s.duration, s.stream_url
		FROM song_artists sa
		JOIN songs s ON s.id = sa.song_id
		WHERE sa.artist_id = $1
		ORDER BY s.streams_count DESC
		LIMIT 5
	`

	const releasesQuery = `
		SELECT DISTINCT r.id, r.name, r.cover_url, r.type, r.release_at
		FROM song_artists sa
		JOIN release_songs rs ON rs.song_id = sa.song_id
		JOIN releases r ON r.id = rs.release_id
		WHERE sa.artist_id = $1
		ORDER BY r.release_at DESC
	`

	var (
		name       string
		avatarURL  *string
		coverURL   *string
		followers  int
		isFollowed bool
	)

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, 0, false, err
	}
	defer tx.Rollback(ctx)

	if err := tx.
		QueryRow(ctx, artistQuery, id, userID).
		Scan(&id, &name, &avatarURL, &coverURL, &followers, &isFollowed); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, 0, false, ErrArtistNotFound
		}
		return nil, 0, false, err
	}

	popRows, err := tx.Query(ctx, popularSongsQuery, id)
	if err != nil {
		return nil, 0, false, err
	}
	defer popRows.Close()

	var popular []PopularSong
	for popRows.Next() {
		var (
			songID     string
			songName   string
			cover      *string
			duration   int
			streamURL  string
		)
		if err := popRows.Scan(&songID, &songName, &cover, &duration, &streamURL); err != nil {
			return nil, 0, false, err
		}
		popular = append(popular, PopularSong{
			ID:              songID,
			Name:            songName,
			CoverURL:        cover,
			DurationSeconds: duration,
			StreamURL:       streamURL,
		})
	}
	if err := popRows.Err(); err != nil {
		return nil, 0, false, err
	}

	relRows, err := tx.Query(ctx, releasesQuery, id)
	if err != nil {
		return nil, 0, false, err
	}
	defer relRows.Close()

	var rels []ArtistRelease
	for relRows.Next() {
		var (
			releaseID   string
			releaseName string
			releaseCov  *string
			releaseType int
			releaseAt   time.Time
		)
		if err := relRows.Scan(&releaseID, &releaseName, &releaseCov, &releaseType, &releaseAt); err != nil {
			return nil, 0, false, err
		}
		rels = append(rels, ArtistRelease{
			ID:        releaseID,
			Name:      releaseName,
			CoverURL:  releaseCov,
			Type:      releaseType,
			ReleaseAt: releaseAt,
		})
	}
	if err := relRows.Err(); err != nil {
		return nil, 0, false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, 0, false, err
	}

	return &Artist{
		ID:        id,
		Name:      name,
		AvatarURL: avatarURL,
		CoverURL:  coverURL,
		Popular:   popular,
		Releases:  rels,
	}, followers, isFollowed, nil
}
