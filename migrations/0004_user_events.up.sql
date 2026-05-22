CREATE TABLE IF NOT EXISTS user_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    event_type TEXT NOT NULL CHECK (
        event_type IN (
            'song_play',
            'song_skip',
            'song_complete',
            'song_favorite',
            'song_unfavorite',
            'playlist_add_song',
            'playlist_remove_song',
            'artist_follow',
            'artist_unfollow'
        )
    ),

    song_id UUID REFERENCES songs(id) ON DELETE CASCADE,
    artist_id UUID REFERENCES artists(id) ON DELETE CASCADE,
    playlist_id UUID REFERENCES playlists(id) ON DELETE CASCADE,
    release_id UUID REFERENCES releases(id) ON DELETE CASCADE,

    position_seconds INT CHECK (position_seconds IS NULL OR position_seconds >= 0),
    duration_seconds INT CHECK (duration_seconds IS NULL OR duration_seconds > 0),
    percent_played NUMERIC(5, 2) CHECK (percent_played IS NULL OR (percent_played >= 0 AND percent_played <= 100)),

    source TEXT,
    context_type TEXT,
    context_id UUID,
    client_event_id UUID,

    event_time TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT user_events_client_event_unique UNIQUE (user_id, client_event_id),
    CONSTRAINT user_events_song_event_has_song CHECK (
        event_type NOT IN ('song_play', 'song_skip', 'song_complete', 'song_favorite', 'song_unfavorite')
        OR song_id IS NOT NULL
    ),
    CONSTRAINT user_events_skip_has_position CHECK (
        event_type != 'song_skip'
        OR position_seconds IS NOT NULL
    ),
    CONSTRAINT user_events_client_playback_event_has_client_event_id CHECK (
        event_type NOT IN ('song_play', 'song_skip', 'song_complete')
        OR client_event_id IS NOT NULL
    ),
    CONSTRAINT user_events_artist_event_has_artist CHECK (
        event_type NOT IN ('artist_follow', 'artist_unfollow')
        OR artist_id IS NOT NULL
    ),
    CONSTRAINT user_events_playlist_song_event_has_playlist_and_song CHECK (
        event_type NOT IN ('playlist_add_song', 'playlist_remove_song')
        OR (playlist_id IS NOT NULL AND song_id IS NOT NULL)
    ),
    CONSTRAINT user_events_context_id_has_type CHECK (
        context_id IS NULL OR context_type IS NOT NULL
    ),
    CONSTRAINT user_events_context_type_valid CHECK (
        context_type IS NULL
        OR context_type IN ('playlist', 'release', 'artist', 'queue', 'recommendations')
    ),
    CONSTRAINT user_events_context_entity_has_id CHECK (
        context_type NOT IN ('playlist', 'release', 'artist')
        OR context_id IS NOT NULL
    )
);
