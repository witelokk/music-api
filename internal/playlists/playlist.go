package playlists

import "time"

type PlaylistSummary struct {
	ID           string
	Name         string
	CoverMediaID *string
	SongsCount   int
}

type Playlist struct {
	ID           string
	UserID       string
	Name         string
	CreatedAt    time.Time
	CoverMediaID *string
	SongsCount   int
	Songs        []PlaylistSong
}

type PlaylistSong struct {
	ID              string
	Name            string
	CoverMediaID    *string
	DurationSeconds int
	StreamMediaID   string
	IsFavorite      bool
	Artists         []PlaylistArtist
}

type PlaylistArtist struct {
	ID            string
	Name          string
	AvatarMediaID *string
}
