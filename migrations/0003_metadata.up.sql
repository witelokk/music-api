CREATE TABLE genres (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE song_genres (
    song_id UUID NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    genre_id UUID NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
    PRIMARY KEY (song_id, genre_id)
);

CREATE TABLE song_languages (
    song_id UUID NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    language_code VARCHAR(2) NOT NULL,
    PRIMARY KEY (song_id, language_code)
);

