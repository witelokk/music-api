package artists

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ArtistsRepository interface {
	GetArtistWithStats(ctx context.Context, id, userID string) (*Artist, int, bool, error)
}

type PostgresArtistsRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresArtistsRepository(pool *pgxpool.Pool) *PostgresArtistsRepository {
	return &PostgresArtistsRepository{pool: pool}
}

func (r *PostgresArtistsRepository) GetArtistWithStats(ctx context.Context, id, userID string) (*Artist, int, bool, error) {
	const artistQuery = `
		SELECT a.id,
		       a.name,
		       a.avatar_media_id,
		       a.cover_media_id,
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
		SELECT s.id, s.name, s.cover_media_id, s.duration, s.stream_media_id
		FROM song_artists sa
		JOIN songs s ON s.id = sa.song_id
		WHERE sa.artist_id = $1
		ORDER BY s.streams_count DESC
		LIMIT 5
	`

	const popularSongArtistsQuery = `
		SELECT top_songs.song_id, a.id, a.name, a.avatar_media_id
		FROM (
			SELECT s.id AS song_id
			FROM song_artists sa
			JOIN songs s ON s.id = sa.song_id
			WHERE sa.artist_id = $1
			ORDER BY s.streams_count DESC
			LIMIT 5
		) top_songs
		JOIN song_artists sa ON sa.song_id = top_songs.song_id
		JOIN artists a ON a.id = sa.artist_id
		ORDER BY top_songs.song_id, a.name
	`

	const releasesQuery = `
		SELECT DISTINCT r.id, r.name, r.cover_media_id, r.type, r.release_at
		FROM song_artists sa
		JOIN release_songs rs ON rs.song_id = sa.song_id
		JOIN releases r ON r.id = rs.release_id
		WHERE sa.artist_id = $1
		ORDER BY r.release_at DESC
	`

	var (
		name          string
		avatarMediaID *string
		coverMediaID  *string
		followers     int
		isFollowed    bool
	)

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, 0, false, err
	}
	defer tx.Rollback(ctx)

	if err := tx.
		QueryRow(ctx, artistQuery, id, userID).
		Scan(&id, &name, &avatarMediaID, &coverMediaID, &followers, &isFollowed); err != nil {
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
			songID        string
			songName      string
			coverMediaID  *string
			duration      int
			streamMediaID string
		)
		if err := popRows.Scan(&songID, &songName, &coverMediaID, &duration, &streamMediaID); err != nil {
			return nil, 0, false, err
		}
		popular = append(popular, PopularSong{
			ID:              songID,
			Name:            songName,
			CoverMediaID:    coverMediaID,
			DurationSeconds: duration,
			StreamMediaID:   streamMediaID,
		})
	}
	if err := popRows.Err(); err != nil {
		return nil, 0, false, err
	}

	popArtistRows, err := tx.Query(ctx, popularSongArtistsQuery, id)
	if err != nil {
		return nil, 0, false, err
	}
	defer popArtistRows.Close()

	songArtists := make(map[string][]ArtistSummary)
	for popArtistRows.Next() {
		var (
			songID        string
			artistID      string
			artistName    string
			avatarMediaID *string
		)
		if err := popArtistRows.Scan(&songID, &artistID, &artistName, &avatarMediaID); err != nil {
			return nil, 0, false, err
		}
		songArtists[songID] = append(songArtists[songID], ArtistSummary{
			ID:            artistID,
			Name:          artistName,
			AvatarMediaID: avatarMediaID,
		})
	}
	if err := popArtistRows.Err(); err != nil {
		return nil, 0, false, err
	}

	for i := range popular {
		popular[i].Artists = songArtists[popular[i].ID]
	}

	relRows, err := tx.Query(ctx, releasesQuery, id)
	if err != nil {
		return nil, 0, false, err
	}
	defer relRows.Close()

	var rels []ArtistRelease
	for relRows.Next() {
		var (
			releaseID      string
			releaseName    string
			releaseCoverID *string
			releaseType    int
			releaseAt      time.Time
		)
		if err := relRows.Scan(&releaseID, &releaseName, &releaseCoverID, &releaseType, &releaseAt); err != nil {
			return nil, 0, false, err
		}
		rels = append(rels, ArtistRelease{
			ID:           releaseID,
			Name:         releaseName,
			CoverMediaID: releaseCoverID,
			Type:         releaseType,
			ReleaseAt:    releaseAt,
		})
	}
	if err := relRows.Err(); err != nil {
		return nil, 0, false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, 0, false, err
	}

	return &Artist{
		ID:            id,
		Name:          name,
		AvatarMediaID: avatarMediaID,
		CoverMediaID:  coverMediaID,
		Popular:       popular,
		Releases:      rels,
	}, followers, isFollowed, nil
}
