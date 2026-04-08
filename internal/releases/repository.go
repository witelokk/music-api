package releases

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReleasesRepository interface {
	GetReleaseByID(ctx context.Context, id string) (*Release, error)
}

type PostgresReleasesRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresReleasesRepository(pool *pgxpool.Pool) *PostgresReleasesRepository {
	return &PostgresReleasesRepository{pool: pool}
}

func (r *PostgresReleasesRepository) GetReleaseByID(ctx context.Context, id string) (*Release, error) {
	const releaseQuery = `
		SELECT id, name, cover_url, type, release_at
		FROM releases
		WHERE id = $1
	`

	const songsQuery = `
		SELECT s.id, s.name, s.cover_url, s.duration, s.stream_url
		FROM release_songs rs
		JOIN songs s ON s.id = rs.song_id
		WHERE rs.release_id = $1
		ORDER BY s.name
	`

	const artistsQuery = `
		SELECT DISTINCT a.id, a.name, a.avatar_url
		FROM release_songs rs
		JOIN song_artists sa ON sa.song_id = rs.song_id
		JOIN artists a ON a.id = sa.artist_id
		WHERE rs.release_id = $1
		ORDER BY a.name
	`

	var (
		name      string
		coverURL  *string
		releaseAt time.Time
		typ       int
	)

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if err := tx.
		QueryRow(ctx, releaseQuery, id).
		Scan(&id, &name, &coverURL, &typ, &releaseAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrReleaseNotFound
		}
		return nil, err
	}

	songRows, err := tx.Query(ctx, songsQuery, id)
	if err != nil {
		return nil, err
	}
	defer songRows.Close()

	var songs []ReleaseSong
	for songRows.Next() {
		var (
			songID    string
			songName  string
			songCover *string
			duration  int
			streamURL string
		)
		if err := songRows.Scan(&songID, &songName, &songCover, &duration, &streamURL); err != nil {
			return nil, err
		}
		songs = append(songs, ReleaseSong{
			ID:              songID,
			Name:            songName,
			CoverURL:        songCover,
			DurationSeconds: duration,
			StreamURL:       streamURL,
		})
	}
	if err := songRows.Err(); err != nil {
		return nil, err
	}

	artistRows, err := tx.Query(ctx, artistsQuery, id)
	if err != nil {
		return nil, err
	}
	defer artistRows.Close()

	var artists []ReleaseArtist
	for artistRows.Next() {
		var (
			artistID   string
			artistName string
			avatarURL  *string
		)
		if err := artistRows.Scan(&artistID, &artistName, &avatarURL); err != nil {
			return nil, err
		}
		artists = append(artists, ReleaseArtist{
			ID:        artistID,
			Name:      artistName,
			AvatarURL: avatarURL,
		})
	}
	if err := artistRows.Err(); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &Release{
		ID:        id,
		Name:      name,
		CoverURL:  coverURL,
		Type:      typ,
		ReleaseAt: releaseAt,
		Songs:     songs,
		Artists:   artists,
	}, nil
}
