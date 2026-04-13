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
	CoverURL        *string
	DurationSeconds int
	StreamURL       string
	Artists         []ArtistSummary
}

type ArtistSummary struct {
	ID        string
	Name      string
	AvatarURL *string
}

func NewPostgresFavoritesRepository(pool *pgxpool.Pool) *PostgresFavoritesRepository {
	return &PostgresFavoritesRepository{pool: pool}
}

func (r *PostgresFavoritesRepository) AddFavorite(ctx context.Context, userID, songID string) error {
	const query = `
		INSERT INTO favorites (user_id, song_id, added_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id, song_id) DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query, userID, songID)
	return err
}

func (r *PostgresFavoritesRepository) RemoveFavorite(ctx context.Context, userID, songID string) error {
	const query = `
		DELETE FROM favorites
		WHERE user_id = $1 AND song_id = $2
	`

	_, err := r.pool.Exec(ctx, query, userID, songID)
	return err
}

func (r *PostgresFavoritesRepository) GetFavoriteSongs(ctx context.Context, userID string) ([]FavoriteSong, error) {
	const query = `
		SELECT s.id,
		       s.name,
		       s.cover_url,
		       s.duration,
		       s.stream_url,
		       a.id AS artist_id,
		       a.name AS artist_name,
		       a.avatar_url
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
		result     []FavoriteSong
		currentID  string
		currentSong FavoriteSong
	)

	for rows.Next() {
		var (
			id           string
			name         string
			coverURL     *string
			duration     int
			stream       string
			artistID     *string
			artistName   *string
			artistAvatar *string
		)
		if err := rows.Scan(&id, &name, &coverURL, &duration, &stream, &artistID, &artistName, &artistAvatar); err != nil {
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
				CoverURL:        coverURL,
				DurationSeconds: duration,
				StreamURL:       stream,
				Artists:         nil,
			}
		}

		if artistID != nil && artistName != nil {
			currentSong.Artists = append(currentSong.Artists, ArtistSummary{
				ID:        *artistID,
				Name:      *artistName,
				AvatarURL: artistAvatar,
			})
		}
	}

	if currentID != "" {
		result = append(result, currentSong)
	}

	return result, rows.Err()
}
