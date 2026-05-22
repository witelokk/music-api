DROP INDEX IF EXISTS idx_user_events_song_play_song;

ALTER TABLE songs ADD COLUMN IF NOT EXISTS streams_count BIGINT NOT NULL DEFAULT 0;
