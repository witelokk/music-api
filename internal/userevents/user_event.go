package userevents

type EventType string

const (
	EventTypeSongPlay           EventType = "song_play"
	EventTypeSongSkip           EventType = "song_skip"
	EventTypeSongComplete       EventType = "song_complete"
	EventTypeSongFavorite       EventType = "song_favorite"
	EventTypeSongUnfavorite     EventType = "song_unfavorite"
	EventTypePlaylistCreate     EventType = "playlist_create"
	EventTypePlaylistDelete     EventType = "playlist_delete"
	EventTypePlaylistAddSong    EventType = "playlist_add_song"
	EventTypePlaylistRemoveSong EventType = "playlist_remove_song"
	EventTypeArtistFollow       EventType = "artist_follow"
	EventTypeArtistUnfollow     EventType = "artist_unfollow"
)

type UserEvent struct {
	UserID    string
	EventType EventType

	SongID     *string
	ArtistID   *string
	PlaylistID *string
	ReleaseID  *string

	PositionSeconds *int
	DurationSeconds *int
	PercentPlayed   *float64

	Source      *string
	ContextType *string
	ContextID   *string

	ClientEventID *string
}
