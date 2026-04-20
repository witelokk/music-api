package favorites

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	AddFavorite(ctx context.Context, userID, songID string) error
	RemoveFavorite(ctx context.Context, userID, songID string) error
	GetFavoriteSongs(ctx context.Context, userID string) ([]FavoriteSong, error)
}

type PostgresFavoritesRepository struct {
	pool *pgxpool.Pool
}

type FavoriteSong struct {
	ID              string
	Name            string
	CoverMediaID    *string
	DurationSeconds int
	StreamMediaID   string
	Artists         []ArtistSummary
}

type ArtistSummary struct {
	ID            string
	Name          string
	AvatarMediaID *string
}

func NewPostgresFavoritesRepository(pool *pgxpool.Pool) *PostgresFavoritesRepository {
	return &PostgresFavoritesRepository{pool: pool}
}

func (r *PostgresFavoritesRepository) AddFavorite(ctx context.Context, userID, songID string) error {
	const query = `
		INSERT INTO favorites (user_id, song_id, added_at)
		SELECT $1, s.id, NOW()
		FROM songs s
		WHERE s.id = $2
		ON CONFLICT (user_id, song_id) DO NOTHING
	`

	cmd, err := r.pool.Exec(ctx, query, userID, songID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return r.classifyFavoriteMutationError(ctx, songID)
	}
	return nil
}

func (r *PostgresFavoritesRepository) RemoveFavorite(ctx context.Context, userID, songID string) error {
	const query = `
		DELETE FROM favorites
		WHERE user_id = $1 AND song_id = $2
	`

	cmd, err := r.pool.Exec(ctx, query, userID, songID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return r.classifyFavoriteMutationError(ctx, songID)
	}
	return nil
}

func (r *PostgresFavoritesRepository) GetFavoriteSongs(ctx context.Context, userID string) ([]FavoriteSong, error) {
	const query = `
		SELECT s.id,
		       s.name,
		       s.cover_media_id,
		       s.duration,
		       s.stream_media_id,
		       a.id AS artist_id,
		       a.name AS artist_name,
		       a.avatar_media_id
		FROM favorites f
		JOIN songs s ON s.id = f.song_id
		LEFT JOIN song_artists sa ON sa.song_id = s.id
		LEFT JOIN artists a ON a.id = sa.artist_id
		WHERE f.user_id = $1
		ORDER BY f.added_at DESC, a.name
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		result      []FavoriteSong
		currentID   string
		currentSong FavoriteSong
	)

	for rows.Next() {
		var (
			id            string
			name          string
			coverMediaID  *string
			duration      int
			streamMediaID string
			artistID      *string
			artistName    *string
			artistAvatar  *string
		)
		if err := rows.Scan(&id, &name, &coverMediaID, &duration, &streamMediaID, &artistID, &artistName, &artistAvatar); err != nil {
			return nil, err
		}

		if id != currentID {
			if currentID != "" {
				result = append(result, currentSong)
			}
			currentID = id
			currentSong = FavoriteSong{
				ID:              id,
				Name:            name,
				CoverMediaID:    coverMediaID,
				DurationSeconds: duration,
				StreamMediaID:   streamMediaID,
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
		result = append(result, currentSong)
	}

	return result, rows.Err()
}

func (r *PostgresFavoritesRepository) classifyFavoriteMutationError(ctx context.Context, songID string) error {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM songs
			WHERE id = $1
		)
	`

	var exists bool
	if err := r.pool.QueryRow(ctx, query, songID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrSongNotFound
	}

	return nil
}
