CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_song_artists_artist_song
ON song_artists (artist_id, song_id);

CREATE INDEX IF NOT EXISTS idx_release_songs_song_release
ON release_songs (song_id, release_id);

CREATE INDEX IF NOT EXISTS idx_followings_artist_user
ON followings (artist_id, user_id);

CREATE INDEX IF NOT EXISTS idx_favorites_user_added_song
ON favorites (user_id, added_at DESC, song_id);

CREATE INDEX IF NOT EXISTS idx_playlists_user_created
ON playlists (user_id, created_at DESC, id);

CREATE INDEX IF NOT EXISTS idx_playlist_songs_playlist_added_song
ON playlist_songs (playlist_id, added_at DESC, song_id);

CREATE INDEX IF NOT EXISTS idx_songs_name_trgm
ON songs USING gin (name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_artists_name_trgm
ON artists USING gin (name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_releases_name_trgm
ON releases USING gin (name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_playlists_name_trgm
ON playlists USING gin (name gin_trgm_ops);
