package home_feed

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrFeedSnapshotNotFound = errors.New("home feed snapshot not found")

type FeedSnapshot struct {
	UserID      string
	Layout      *Layout
	GeneratedAt time.Time
	StaleAt     *time.Time
	Version     string
}

type FeedSnapshotReader interface {
	Get(ctx context.Context, userID string) (*FeedSnapshot, error)
}

type FeedSnapshotWriter interface {
	Save(ctx context.Context, userID string, layout *Layout, generatedAt time.Time) error
}

type PostgresFeedSnapshotRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresFeedSnapshotRepository(pool *pgxpool.Pool) *PostgresFeedSnapshotRepository {
	return &PostgresFeedSnapshotRepository{pool: pool}
}

func (r *PostgresFeedSnapshotRepository) Get(ctx context.Context, userID string) (*FeedSnapshot, error) {
	const query = `
		SELECT user_id, payload, generated_at, stale_at, version
		FROM home_feed_snapshots
		WHERE user_id = $1
	`

	var (
		snapshot     FeedSnapshot
		payload      []byte
		generatedAt  time.Time
		staleAt      *time.Time
		version      string
		snapshotUser string
	)
	if err := r.pool.QueryRow(ctx, query, userID).Scan(&snapshotUser, &payload, &generatedAt, &staleAt, &version); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFeedSnapshotNotFound
		}
		return nil, err
	}

	var layout Layout
	if err := json.Unmarshal(payload, &layout); err != nil {
		return nil, err
	}

	snapshot.UserID = snapshotUser
	snapshot.Layout = &layout
	snapshot.GeneratedAt = generatedAt
	snapshot.StaleAt = staleAt
	snapshot.Version = version
	return &snapshot, nil
}

func (r *PostgresFeedSnapshotRepository) Save(ctx context.Context, userID string, layout *Layout, generatedAt time.Time) error {
	payload, err := json.Marshal(layout)
	if err != nil {
		return err
	}

	const query = `
		INSERT INTO home_feed_snapshots (user_id, payload, generated_at, version)
		VALUES ($1, $2, $3, 'v1')
		ON CONFLICT (user_id) DO UPDATE
		SET payload = EXCLUDED.payload,
		    generated_at = EXCLUDED.generated_at,
		    stale_at = NULL,
		    version = EXCLUDED.version
	`

	_, err = r.pool.Exec(ctx, query, userID, payload, generatedAt)
	return err
}
