package playlists

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PlaylistsRepository interface {
	CreatePlaylist(ctx context.Context, userID, name string) (string, error)
	UpdatePlaylist(ctx context.Context, userID, playlistID, name string) error
	DeletePlaylist(ctx context.Context, userID, playlistID string) error
	GetPlaylists(ctx context.Context, userID string) ([]PlaylistSummary, error)
	GetPlaylist(ctx context.Context, userID, playlistID string) (*Playlist, error)
	GetPlaylistSongs(ctx context.Context, userID, playlistID string) ([]PlaylistSong, error)
	AddSongToPlaylist(ctx context.Context, userID, playlistID, songID string) error
	RemoveSongFromPlaylist(ctx context.Context, userID, playlistID, songID string) error
}

type PostgresPlaylistsRepository struct {
	pool     *pgxpool.Pool
	initOnce sync.Once
	initErr  error
}

func NewPostgresPlaylistsRepository(pool *pgxpool.Pool) *PostgresPlaylistsRepository {
	return &PostgresPlaylistsRepository{pool: pool}
}

func (r *PostgresPlaylistsRepository) ensureTables(ctx context.Context) error {
	r.initOnce.Do(func() {
		_, err := r.pool.Exec(ctx, `
			CREATE TABLE IF NOT EXISTS playlists (
				id UUID PRIMARY KEY,
				user_id UUID NOT NULL,
				name VARCHAR(255) NOT NULL,
				created_at TIMESTAMP NOT NULL DEFAULT NOW()
			);

			CREATE TABLE IF NOT EXISTS playlist_songs (
				playlist_id UUID NOT NULL,
				song_id UUID NOT NULL,
				added_at TIMESTAMP NOT NULL DEFAULT NOW(),
				PRIMARY KEY (playlist_id, song_id)
			)
		`)
		r.initErr = err
	})
	return r.initErr
}

func (r *PostgresPlaylistsRepository) CreatePlaylist(ctx context.Context, userID, name string) (string, error) {
	if err := r.ensureTables(ctx); err != nil {
		return "", err
	}

	const query = `
		INSERT INTO playlists (id, user_id, name, created_at)
		VALUES (gen_random_uuid(), $1, $2, NOW())
		RETURNING id
	`

	var id string
	if err := r.pool.QueryRow(ctx, query, userID, name).Scan(&id); err != nil {
		return "", err
	}
	return id, nil
}

func (r *PostgresPlaylistsRepository) UpdatePlaylist(ctx context.Context, userID, playlistID, name string) error {
	if err := r.ensureTables(ctx); err != nil {
		return err
	}

	const query = `
		UPDATE playlists
		SET name = $3
		WHERE id = $1 AND user_id = $2
	`

	cmd, err := r.pool.Exec(ctx, query, playlistID, userID, name)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *PostgresPlaylistsRepository) DeletePlaylist(ctx context.Context, userID, playlistID string) error {
	if err := r.ensureTables(ctx); err != nil {
		return err
	}

	const query = `
		DELETE FROM playlists
		WHERE id = $1 AND user_id = $2
	`

	cmd, err := r.pool.Exec(ctx, query, playlistID, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *PostgresPlaylistsRepository) GetPlaylists(ctx context.Context, userID string) ([]PlaylistSummary, error) {
	if err := r.ensureTables(ctx); err != nil {
		return nil, err
	}

	const query = `
		SELECT p.id,
		       p.name,
		       MAX(s.cover_url) AS cover_url,
		       COUNT(ps.song_id) AS songs_count,
		       p.created_at
		FROM playlists p
		LEFT JOIN playlist_songs ps ON ps.playlist_id = p.id
		LEFT JOIN songs s ON s.id = ps.song_id
		WHERE p.user_id = $1
		GROUP BY p.id, p.name, p.created_at
		ORDER BY p.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PlaylistSummary
	for rows.Next() {
		var (
			id         string
			name       string
			coverURL   *string
			songsCount int
			createdAt  time.Time
		)
		if err := rows.Scan(&id, &name, &coverURL, &songsCount, &createdAt); err != nil {
			return nil, err
		}
		result = append(result, PlaylistSummary{
			ID:         id,
			Name:       name,
			CoverURL:   coverURL,
			SongsCount: songsCount,
		})
	}

	return result, rows.Err()
}

func (r *PostgresPlaylistsRepository) GetPlaylist(ctx context.Context, userID, playlistID string) (*Playlist, error) {
	if err := r.ensureTables(ctx); err != nil {
		return nil, err
	}

	const query = `
		SELECT p.id,
		       p.user_id,
		       p.name,
		       p.created_at,
		       MAX(s.cover_url) AS cover_url,
		       COUNT(ps.song_id) AS songs_count
		FROM playlists p
		LEFT JOIN playlist_songs ps ON ps.playlist_id = p.id
		LEFT JOIN songs s ON s.id = ps.song_id
		WHERE p.id = $1 AND p.user_id = $2
		GROUP BY p.id, p.user_id, p.name, p.created_at
	`

	var (
		id         string
		ownerID    string
		name       string
		createdAt  time.Time
		coverURL   *string
		songsCount int
	)

	if err := r.pool.QueryRow(ctx, query, playlistID, userID).
		Scan(&id, &ownerID, &name, &createdAt, &coverURL, &songsCount); err != nil {
		if err == pgx.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return &Playlist{
		ID:         id,
		UserID:     ownerID,
		Name:       name,
		CreatedAt:  createdAt,
		CoverURL:   coverURL,
		SongsCount: songsCount,
	}, nil
}

func (r *PostgresPlaylistsRepository) GetPlaylistSongs(ctx context.Context, userID, playlistID string) ([]PlaylistSong, error) {
	if err := r.ensureTables(ctx); err != nil {
		return nil, err
	}

	const query = `
		SELECT s.id,
		       s.name,
		       s.cover_url,
		       s.duration,
		       s.stream_url,
		       EXISTS (
		         SELECT 1
		         FROM favorites f
		         WHERE f.user_id = $2 AND f.song_id = s.id
		       ) AS is_favorite,
		       a.id AS artist_id,
		       a.name AS artist_name,
		       a.avatar_url
		FROM playlist_songs ps
		JOIN playlists p ON p.id = ps.playlist_id
		JOIN songs s ON s.id = ps.song_id
		LEFT JOIN song_artists sa ON sa.song_id = s.id
		LEFT JOIN artists a ON a.id = sa.artist_id
		WHERE ps.playlist_id = $1 AND p.user_id = $2
		ORDER BY ps.added_at DESC, a.name
	`

	rows, err := r.pool.Query(ctx, query, playlistID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PlaylistSong
	var (
		currentSongID string
		currentSong   PlaylistSong
	)

	for rows.Next() {
		var (
			id          string
			name        string
			coverURL    *string
			duration    int
			streamURL   string
			isFavorite  bool
			artistID    *string
			artistName  *string
			artistAvatar *string
		)
		if err := rows.Scan(&id, &name, &coverURL, &duration, &streamURL, &isFavorite, &artistID, &artistName, &artistAvatar); err != nil {
			return nil, err
		}

		if id != currentSongID {
			if currentSongID != "" {
				result = append(result, currentSong)
			}
			currentSongID = id
			currentSong = PlaylistSong{
				ID:              id,
				Name:            name,
				CoverURL:        coverURL,
				DurationSeconds: duration,
				StreamURL:       streamURL,
				IsFavorite:      isFavorite,
				Artists:         nil,
			}
		}

		if artistID != nil && artistName != nil {
			currentSong.Artists = append(currentSong.Artists, PlaylistArtist{
				ID:        *artistID,
				Name:      *artistName,
				AvatarURL: artistAvatar,
			})
		}
	}

	if currentSongID != "" {
		result = append(result, currentSong)
	}

	return result, rows.Err()
}

func (r *PostgresPlaylistsRepository) AddSongToPlaylist(ctx context.Context, userID, playlistID, songID string) error {
	if err := r.ensureTables(ctx); err != nil {
		return err
	}

	const query = `
		INSERT INTO playlist_songs (playlist_id, song_id, added_at)
		SELECT p.id, $3, NOW()
		FROM playlists p
		WHERE p.id = $1 AND p.user_id = $2
		ON CONFLICT (playlist_id, song_id) DO NOTHING
	`

	cmd, err := r.pool.Exec(ctx, query, playlistID, userID, songID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *PostgresPlaylistsRepository) RemoveSongFromPlaylist(ctx context.Context, userID, playlistID, songID string) error {
	if err := r.ensureTables(ctx); err != nil {
		return err
	}

	const query = `
		DELETE FROM playlist_songs ps
		USING playlists p
		WHERE ps.playlist_id = p.id
		  AND ps.playlist_id = $1
		  AND ps.song_id = $2
		  AND p.user_id = $3
	`

	cmd, err := r.pool.Exec(ctx, query, playlistID, songID, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
