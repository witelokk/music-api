CREATE TABLE IF NOT EXISTS release_artists (
    release_id UUID NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
    artist_id UUID NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    PRIMARY KEY (release_id, artist_id)
);

CREATE INDEX IF NOT EXISTS idx_release_artists_artist_release
ON release_artists (artist_id, release_id);
