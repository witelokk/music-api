package releases

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReleasesRepository interface {
	GetReleaseByID(ctx context.Context, userID, id string) (*Release, error)
	GetRandomReleases(ctx context.Context, seed string, limit int) ([]Release, error)
	GetRecentReleases(ctx context.Context, limit int) ([]Release, error)
	GetReleasesByFollowedArtistSeeds(ctx context.Context, artistIDs []string, seed string, limit int) ([]Release, error)
	GetReleasesByFavoriteSongSeeds(ctx context.Context, songIDs []string, seed string, limit int) ([]Release, error)
}

type PostgresReleasesRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresReleasesRepository(pool *pgxpool.Pool) *PostgresReleasesRepository {
	return &PostgresReleasesRepository{pool: pool}
}

func (r *PostgresReleasesRepository) GetReleaseByID(ctx context.Context, userID, id string) (*Release, error) {
	const releaseQuery = `
		SELECT id, name, cover_media_id, type, release_at
		FROM releases
		WHERE id = $1
	`

	const songsQuery = `
		SELECT s.id,
		       s.name,
		       s.cover_media_id,
		       s.duration,
		       s.stream_media_id,
		       EXISTS (
		         SELECT 1
		         FROM favorites f
		         WHERE f.user_id = $2 AND f.song_id = s.id
		       ) AS is_favorite
		FROM release_songs rs
		JOIN songs s ON s.id = rs.song_id
		WHERE rs.release_id = $1
		ORDER BY s.name
	`

	const songArtistsQuery = `
		SELECT rs.song_id, a.id, a.name, a.avatar_media_id
		FROM release_songs rs
		JOIN song_artists sa ON sa.song_id = rs.song_id
		JOIN artists a ON a.id = sa.artist_id
		WHERE rs.release_id = $1
		ORDER BY rs.song_id, a.name
	`

	const artistsQuery = `
		SELECT a.id, a.name, a.avatar_media_id
		FROM release_artists ra
		JOIN artists a ON a.id = ra.artist_id
		WHERE ra.release_id = $1
		ORDER BY a.name
	`

	var (
		name         string
		coverMediaID *string
		releaseAt    time.Time
		typ          int
	)

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if err := tx.
		QueryRow(ctx, releaseQuery, id).
		Scan(&id, &name, &coverMediaID, &typ, &releaseAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrReleaseNotFound
		}
		return nil, err
	}

	songRows, err := tx.Query(ctx, songsQuery, id, userID)
	if err != nil {
		return nil, err
	}
	defer songRows.Close()

	var songs []ReleaseSong
	for songRows.Next() {
		var (
			songID        string
			songName      string
			songCoverID   *string
			duration      int
			streamMediaID string
			isFavorite    bool
		)
		if err := songRows.Scan(&songID, &songName, &songCoverID, &duration, &streamMediaID, &isFavorite); err != nil {
			return nil, err
		}
		songs = append(songs, ReleaseSong{
			ID:              songID,
			Name:            songName,
			CoverMediaID:    songCoverID,
			DurationSeconds: duration,
			StreamMediaID:   streamMediaID,
			IsFavorite:      isFavorite,
		})
	}
	if err := songRows.Err(); err != nil {
		return nil, err
	}

	artistRowsBySong, err := tx.Query(ctx, songArtistsQuery, id)
	if err != nil {
		return nil, err
	}
	defer artistRowsBySong.Close()

	songArtists := make(map[string][]ReleaseArtist)
	for artistRowsBySong.Next() {
		var (
			songID        string
			artistID      string
			name          string
			avatarMediaID *string
		)
		if err := artistRowsBySong.Scan(&songID, &artistID, &name, &avatarMediaID); err != nil {
			return nil, err
		}
		songArtists[songID] = append(songArtists[songID], ReleaseArtist{
			ID:            artistID,
			Name:          name,
			AvatarMediaID: avatarMediaID,
		})
	}
	if err := artistRowsBySong.Err(); err != nil {
		return nil, err
	}

	for i := range songs {
		if artists, ok := songArtists[songs[i].ID]; ok {
			songs[i].Artists = artists
		}
	}

	artistRows, err := tx.Query(ctx, artistsQuery, id)
	if err != nil {
		return nil, err
	}
	defer artistRows.Close()

	var artists []ReleaseArtist
	for artistRows.Next() {
		var (
			artistID      string
			artistName    string
			avatarMediaID *string
		)
		if err := artistRows.Scan(&artistID, &artistName, &avatarMediaID); err != nil {
			return nil, err
		}
		artists = append(artists, ReleaseArtist{
			ID:            artistID,
			Name:          artistName,
			AvatarMediaID: avatarMediaID,
		})
	}
	if err := artistRows.Err(); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &Release{
		ID:           id,
		Name:         name,
		CoverMediaID: coverMediaID,
		Type:         typ,
		ReleaseAt:    releaseAt,
		Songs:        songs,
		Artists:      artists,
	}, nil
}

func (r *PostgresReleasesRepository) GetRandomReleases(ctx context.Context, seed string, limit int) ([]Release, error) {
	if limit <= 0 {
		limit = 50
	}

	const query = `
		SELECT r.id, r.name, r.cover_media_id, r.type, r.release_at
		FROM releases r
		ORDER BY md5(r.id::text || $1)
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, seed, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Release
	for rows.Next() {
		var (
			id           string
			name         string
			coverMediaID *string
			typ          int
			releaseAt    time.Time
		)
		if err := rows.Scan(&id, &name, &coverMediaID, &typ, &releaseAt); err != nil {
			return nil, err
		}
		result = append(result, Release{
			ID:           id,
			Name:         name,
			CoverMediaID: coverMediaID,
			Type:         typ,
			ReleaseAt:    releaseAt,
			Songs:        nil,
			Artists:      nil,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return result, nil
	}

	return r.attachArtists(ctx, result)
}

func (r *PostgresReleasesRepository) GetRecentReleases(ctx context.Context, limit int) ([]Release, error) {
	if limit <= 0 {
		limit = 50
	}

	const query = `
		SELECT r.id, r.name, r.cover_media_id, r.type, r.release_at
		FROM releases r
		ORDER BY r.release_at DESC, r.id
		LIMIT $1
	`

	return r.getReleaseSummaries(ctx, query, limit)
}

func (r *PostgresReleasesRepository) GetReleasesByFollowedArtistSeeds(ctx context.Context, artistIDs []string, seed string, limit int) ([]Release, error) {
	if len(artistIDs) == 0 {
		return nil, nil
	}
	if limit <= 0 {
		limit = 50
	}

	const query = `
		WITH seed_artists AS (
			SELECT unnest($1::uuid[]) AS artist_id
		),
		seed_genres AS (
			SELECT DISTINCT sg.genre_id
			FROM seed_artists seed
			JOIN song_artists sa ON sa.artist_id = seed.artist_id
			JOIN song_genres sg ON sg.song_id = sa.song_id
		),
		seed_collaborators AS (
			SELECT DISTINCT sa2.artist_id
			FROM seed_artists seed
			JOIN song_artists sa1 ON sa1.artist_id = seed.artist_id
			JOIN song_artists sa2 ON sa2.song_id = sa1.song_id
			WHERE sa2.artist_id <> seed.artist_id
		),
		candidate_releases AS (
			SELECT r.id,
			       COUNT(DISTINCT sg.genre_id) + COUNT(DISTINCT sc.artist_id) AS score
			FROM releases r
			JOIN release_songs rs ON rs.release_id = r.id
			LEFT JOIN song_genres sg ON sg.song_id = rs.song_id
			LEFT JOIN song_artists sa ON sa.song_id = rs.song_id
			LEFT JOIN seed_collaborators sc ON sc.artist_id = sa.artist_id
			WHERE sg.genre_id IN (SELECT genre_id FROM seed_genres)
			   OR sc.artist_id IS NOT NULL
			GROUP BY r.id
		)
		SELECT r.id, r.name, r.cover_media_id, r.type, r.release_at
		FROM candidate_releases c
		JOIN releases r ON r.id = c.id
		ORDER BY c.score DESC, md5(r.id::text || $2), r.release_at DESC
		LIMIT $3
	`

	return r.getReleaseSummaries(ctx, query, artistIDs, seed, limit)
}

func (r *PostgresReleasesRepository) GetReleasesByFavoriteSongSeeds(ctx context.Context, songIDs []string, seed string, limit int) ([]Release, error) {
	if len(songIDs) == 0 {
		return nil, nil
	}
	if limit <= 0 {
		limit = 50
	}

	const query = `
		WITH seed_songs AS (
			SELECT unnest($1::uuid[]) AS song_id
		),
		seed_genres AS (
			SELECT DISTINCT sg.genre_id
			FROM seed_songs seed
			JOIN song_genres sg ON sg.song_id = seed.song_id
		),
		seed_artists AS (
			SELECT DISTINCT sa.artist_id
			FROM seed_songs seed
			JOIN song_artists sa ON sa.song_id = seed.song_id
		),
		seed_collaborators AS (
			SELECT DISTINCT sa2.artist_id
			FROM seed_artists seed
			JOIN song_artists sa1 ON sa1.artist_id = seed.artist_id
			JOIN song_artists sa2 ON sa2.song_id = sa1.song_id
			WHERE sa2.artist_id <> seed.artist_id
		),
		candidate_releases AS (
			SELECT r.id,
			       COUNT(DISTINCT sg.genre_id) + COUNT(DISTINCT sa.artist_id) + COUNT(DISTINCT sc.artist_id) AS score
			FROM releases r
			JOIN release_songs rs ON rs.release_id = r.id
			LEFT JOIN song_genres sg ON sg.song_id = rs.song_id
			LEFT JOIN song_artists sa ON sa.song_id = rs.song_id
			LEFT JOIN seed_collaborators sc ON sc.artist_id = sa.artist_id
			WHERE rs.song_id NOT IN (SELECT song_id FROM seed_songs)
			  AND (
			    sg.genre_id IN (SELECT genre_id FROM seed_genres)
			    OR sa.artist_id IN (SELECT artist_id FROM seed_artists)
			    OR sc.artist_id IS NOT NULL
			  )
			GROUP BY r.id
		)
		SELECT r.id, r.name, r.cover_media_id, r.type, r.release_at
		FROM candidate_releases c
		JOIN releases r ON r.id = c.id
		ORDER BY c.score DESC, md5(r.id::text || $2), r.release_at DESC
		LIMIT $3
	`

	return r.getReleaseSummaries(ctx, query, songIDs, seed, limit)
}

func (r *PostgresReleasesRepository) getReleaseSummaries(ctx context.Context, query string, args ...any) ([]Release, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Release
	for rows.Next() {
		var (
			id           string
			name         string
			coverMediaID *string
			typ          int
			releaseAt    time.Time
		)
		if err := rows.Scan(&id, &name, &coverMediaID, &typ, &releaseAt); err != nil {
			return nil, err
		}
		result = append(result, Release{
			ID:           id,
			Name:         name,
			CoverMediaID: coverMediaID,
			Type:         typ,
			ReleaseAt:    releaseAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return result, nil
	}

	return r.attachArtists(ctx, result)
}

func (r *PostgresReleasesRepository) attachArtists(ctx context.Context, result []Release) ([]Release, error) {
	ids := make([]string, 0, len(result))
	for _, rel := range result {
		ids = append(ids, rel.ID)
	}

	const artistsQuery = `
		SELECT ra.release_id, a.id, a.name, a.avatar_media_id
		FROM release_artists ra
		JOIN artists a ON a.id = ra.artist_id
		WHERE ra.release_id = ANY($1::uuid[])
		ORDER BY a.name
	`

	artistRows, err := r.pool.Query(ctx, artistsQuery, ids)
	if err != nil {
		return nil, err
	}
	defer artistRows.Close()

	artistsByRelease := make(map[string][]ReleaseArtist)
	for artistRows.Next() {
		var (
			releaseID     string
			artistID      string
			name          string
			avatarMediaID *string
		)
		if err := artistRows.Scan(&releaseID, &artistID, &name, &avatarMediaID); err != nil {
			return nil, err
		}
		artistsByRelease[releaseID] = append(artistsByRelease[releaseID], ReleaseArtist{
			ID:            artistID,
			Name:          name,
			AvatarMediaID: avatarMediaID,
		})
	}
	if err := artistRows.Err(); err != nil {
		return nil, err
	}

	for i, rel := range result {
		if artists, ok := artistsByRelease[rel.ID]; ok {
			result[i].Artists = artists
		}
	}

	return result, nil
}
