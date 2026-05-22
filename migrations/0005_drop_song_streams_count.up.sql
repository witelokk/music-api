ALTER TABLE songs DROP COLUMN IF EXISTS streams_count;

CREATE INDEX IF NOT EXISTS idx_user_events_song_play_song
ON user_events (song_id, user_id, client_event_id)
WHERE event_type = 'song_play';
