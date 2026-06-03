CREATE TABLE IF NOT EXISTS home_feed_snapshots (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    payload JSONB NOT NULL,
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    stale_at TIMESTAMPTZ,
    version TEXT NOT NULL DEFAULT 'v1'
);

CREATE INDEX IF NOT EXISTS idx_home_feed_snapshots_generated
ON home_feed_snapshots (generated_at DESC);
