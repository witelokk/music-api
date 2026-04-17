CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS songs (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    cover_url TEXT,
    duration INT NOT NULL,
    stream_url TEXT NOT NULL,
    streams_count BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS artists (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    avatar_url TEXT,
    cover_url TEXT
);

CREATE TABLE IF NOT EXISTS releases (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    cover_url TEXT,
    type INT NOT NULL,
    release_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS song_artists (
    song_id UUID NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    artist_id UUID NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    PRIMARY KEY (song_id, artist_id)
);

CREATE TABLE IF NOT EXISTS release_songs (
    release_id UUID NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
    song_id UUID NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    PRIMARY KEY (release_id, song_id)
);

CREATE TABLE IF NOT EXISTS favorites (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    song_id UUID NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    added_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, song_id)
);

CREATE TABLE IF NOT EXISTS followings (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    artist_id UUID NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, artist_id)
);

CREATE TABLE IF NOT EXISTS playlists (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS playlist_songs (
    playlist_id UUID NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
    song_id UUID NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    added_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (playlist_id, song_id)
);
