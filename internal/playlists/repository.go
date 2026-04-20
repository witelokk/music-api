package playlists

import (
	"context"
	"errors"
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
	pool *pgxpool.Pool
}

func NewPostgresPlaylistsRepository(pool *pgxpool.Pool) *PostgresPlaylistsRepository {
	return &PostgresPlaylistsRepository{pool: pool}
}

func (r *PostgresPlaylistsRepository) CreatePlaylist(ctx context.Context, userID, name string) (string, error) {
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
		return ErrPlaylistNotFound
	}
	return nil
}

func (r *PostgresPlaylistsRepository) DeletePlaylist(ctx context.Context, userID, playlistID string) error {
	const query = `
		DELETE FROM playlists
		WHERE id = $1 AND user_id = $2
	`

	cmd, err := r.pool.Exec(ctx, query, playlistID, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrPlaylistNotFound
	}
	return nil
}

func (r *PostgresPlaylistsRepository) GetPlaylists(ctx context.Context, userID string) ([]PlaylistSummary, error) {
	const query = `
		SELECT p.id,
		       p.name,
		       MAX(s.cover_media_id) AS cover_media_id,
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
			id           string
			name         string
			coverMediaID *string
			songsCount   int
			createdAt    time.Time
		)
		if err := rows.Scan(&id, &name, &coverMediaID, &songsCount, &createdAt); err != nil {
			return nil, err
		}
		result = append(result, PlaylistSummary{
			ID:           id,
			Name:         name,
			CoverMediaID: coverMediaID,
			SongsCount:   songsCount,
		})
	}

	return result, rows.Err()
}

func (r *PostgresPlaylistsRepository) GetPlaylist(ctx context.Context, userID, playlistID string) (*Playlist, error) {
	const query = `
		SELECT p.id,
		       p.user_id,
		       p.name,
		       p.created_at,
		       MAX(s.cover_media_id) AS cover_media_id,
		       COUNT(ps.song_id) AS songs_count
		FROM playlists p
		LEFT JOIN playlist_songs ps ON ps.playlist_id = p.id
		LEFT JOIN songs s ON s.id = ps.song_id
		WHERE p.id = $1 AND p.user_id = $2
		GROUP BY p.id, p.user_id, p.name, p.created_at
	`

	var (
		id           string
		ownerID      string
		name         string
		createdAt    time.Time
		coverMediaID *string
		songsCount   int
	)

	if err := r.pool.QueryRow(ctx, query, playlistID, userID).
		Scan(&id, &ownerID, &name, &createdAt, &coverMediaID, &songsCount); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPlaylistNotFound
		}
		return nil, err
	}

	return &Playlist{
		ID:           id,
		UserID:       ownerID,
		Name:         name,
		CreatedAt:    createdAt,
		CoverMediaID: coverMediaID,
		SongsCount:   songsCount,
	}, nil
}

func (r *PostgresPlaylistsRepository) GetPlaylistSongs(ctx context.Context, userID, playlistID string) ([]PlaylistSong, error) {
	const query = `
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

		if id != currentSongID {
			if currentSongID != "" {
				result = append(result, currentSong)
			}
			currentSongID = id
			currentSong = PlaylistSong{
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
			currentSong.Artists = append(currentSong.Artists, PlaylistArtist{
				ID:            *artistID,
				Name:          *artistName,
				AvatarMediaID: artistAvatar,
			})
		}
	}

	if currentSongID != "" {
		result = append(result, currentSong)
	}

	return result, rows.Err()
}

func (r *PostgresPlaylistsRepository) AddSongToPlaylist(ctx context.Context, userID, playlistID, songID string) error {
	const query = `
		INSERT INTO playlist_songs (playlist_id, song_id, added_at)
		SELECT p.id, s.id, NOW()
		FROM playlists p
		JOIN songs s ON s.id = $3
		WHERE p.id = $1 AND p.user_id = $2
		ON CONFLICT (playlist_id, song_id) DO NOTHING
	`

	cmd, err := r.pool.Exec(ctx, query, playlistID, userID, songID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return r.classifyPlaylistSongMutationError(ctx, userID, playlistID, songID)
	}
	return nil
}

func (r *PostgresPlaylistsRepository) RemoveSongFromPlaylist(ctx context.Context, userID, playlistID, songID string) error {
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
		return r.classifyPlaylistSongMutationError(ctx, userID, playlistID, songID)
	}
	return nil
}

func (r *PostgresPlaylistsRepository) classifyPlaylistSongMutationError(
	ctx context.Context,
	userID, playlistID, songID string,
) error {
	playlistExists, err := r.playlistExists(ctx, userID, playlistID)
	if err != nil {
		return err
	}
	if !playlistExists {
		return ErrPlaylistNotFound
	}

	songExists, err := r.songExists(ctx, songID)
	if err != nil {
		return err
	}
	if !songExists {
		return ErrSongNotFound
	}

	return nil
}

func (r *PostgresPlaylistsRepository) playlistExists(ctx context.Context, userID, playlistID string) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM playlists
			WHERE id = $1 AND user_id = $2
		)
	`

	var exists bool
	if err := r.pool.QueryRow(ctx, query, playlistID, userID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *PostgresPlaylistsRepository) songExists(ctx context.Context, songID string) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM songs
			WHERE id = $1
		)
	`

	var exists bool
	if err := r.pool.QueryRow(ctx, query, songID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}
