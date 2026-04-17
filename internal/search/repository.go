package search

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SearchRepository interface {
	SearchSongs(ctx context.Context, query, userID string) ([]SongResult, error)
	SearchArtists(ctx context.Context, query string) ([]ArtistResult, error)
	SearchReleases(ctx context.Context, query string) ([]ReleaseResult, error)
	SearchPlaylists(ctx context.Context, query, userID string) ([]PlaylistResult, error)
}

type PostgresSearchRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresSearchRepository(pool *pgxpool.Pool) *PostgresSearchRepository {
	return &PostgresSearchRepository{pool: pool}
}

func (r *PostgresSearchRepository) SearchSongs(ctx context.Context, query, userID string) ([]SongResult, error) {
	const q = `
		SELECT s.id,
		       s.name,
		       s.cover_media_id,
		       s.duration,
		       s.stream_media_id,
		       EXISTS (
		         SELECT 1
		         FROM favorites f
		         WHERE f.user_id = $2 AND f.song_id = s.id
		       ) AS is_favorite,
		       a.id AS artist_id,
		       a.name AS artist_name,
		       a.avatar_media_id
		FROM songs s
		LEFT JOIN song_artists sa ON sa.song_id = s.id
		LEFT JOIN artists a ON a.id = sa.artist_id
		WHERE s.name ILIKE '%' || $1 || '%'
		ORDER BY s.name, a.name
	`

	rows, err := r.pool.Query(ctx, q, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		results     []SongResult
		currentID   string
		currentSong SongResult
	)

	for rows.Next() {
		var (
			id            string
			name          string
			coverMediaID  *string
			duration      int
			streamMediaID string
			isFavorite    bool
			artistID      *string
			artistName    *string
			artistAvatar  *string
		)

		if err := rows.Scan(&id, &name, &coverMediaID, &duration, &streamMediaID, &isFavorite, &artistID, &artistName, &artistAvatar); err != nil {
			return nil, err
		}

		if id != currentID {
			if currentID != "" {
				results = append(results, currentSong)
			}
			currentID = id
			currentSong = SongResult{
				ID:              id,
				Name:            name,
				CoverMediaID:    coverMediaID,
				DurationSeconds: duration,
				StreamMediaID:   streamMediaID,
				IsFavorite:      isFavorite,
				Artists:         nil,
			}
		}

		if artistID != nil && artistName != nil {
			currentSong.Artists = append(currentSong.Artists, ArtistSummary{
				ID:            *artistID,
				Name:          *artistName,
				AvatarMediaID: artistAvatar,
			})
		}
	}

	if currentID != "" {
		results = append(results, currentSong)
	}

	return results, rows.Err()
}

func (r *PostgresSearchRepository) SearchArtists(ctx context.Context, query string) ([]ArtistResult, error) {
	const q = `
		SELECT a.id, a.name, a.avatar_media_id
		FROM artists a
		WHERE a.name ILIKE '%' || $1 || '%'
		ORDER BY a.name
	`

	rows, err := r.pool.Query(ctx, q, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ArtistResult
	for rows.Next() {
		var (
			id            string
			name          string
			avatarMediaID *string
		)
		if err := rows.Scan(&id, &name, &avatarMediaID); err != nil {
			return nil, err
		}
		results = append(results, ArtistResult{
			ID:            id,
			Name:          name,
			AvatarMediaID: avatarMediaID,
		})
	}

	return results, rows.Err()
}

func (r *PostgresSearchRepository) SearchReleases(ctx context.Context, query string) ([]ReleaseResult, error) {
	const q = `
		SELECT r.id, r.name, r.cover_media_id, r.type, r.release_at
		FROM releases r
		WHERE r.name ILIKE '%' || $1 || '%'
		ORDER BY r.release_at DESC
	`

	rows, err := r.pool.Query(ctx, q, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ReleaseResult
	for rows.Next() {
		var (
			id           string
			name         string
			coverMediaID *string
			relType      int
			releaseAt    time.Time
		)
		if err := rows.Scan(&id, &name, &coverMediaID, &relType, &releaseAt); err != nil {
			return nil, err
		}
		results = append(results, ReleaseResult{
			ID:           id,
			Name:         name,
			CoverMediaID: coverMediaID,
			Type:         relType,
			ReleaseAt:    releaseAt.Format("2006-01-02"),
		})
	}

	return results, rows.Err()
}

func (r *PostgresSearchRepository) SearchPlaylists(ctx context.Context, query, userID string) ([]PlaylistResult, error) {
	if userID == "" {
		return nil, nil
	}

	const q = `
		SELECT p.id,
		       p.name,
		       MAX(s.cover_media_id) AS cover_media_id,
		       COUNT(ps.song_id) AS songs_count
		FROM playlists p
		LEFT JOIN playlist_songs ps ON ps.playlist_id = p.id
		LEFT JOIN songs s ON s.id = ps.song_id
		WHERE p.user_id = $2
		  AND p.name ILIKE '%' || $1 || '%'
		GROUP BY p.id, p.name
		ORDER BY p.name
	`

	rows, err := r.pool.Query(ctx, q, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []PlaylistResult
	for rows.Next() {
		var (
			id           string
			name         string
			coverMediaID *string
			songsCount   int
		)
		if err := rows.Scan(&id, &name, &coverMediaID, &songsCount); err != nil {
			return nil, err
		}
		results = append(results, PlaylistResult{
			ID:           id,
			Name:         name,
			CoverMediaID: coverMediaID,
			SongsCount:   songsCount,
		})
	}

	return results, rows.Err()
}
