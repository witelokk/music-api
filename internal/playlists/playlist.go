package playlists

import "time"

type PlaylistSummary struct {
	ID         string
	Name       string
	CoverURL   *string
	SongsCount int
}

type Playlist struct {
	ID         string
	UserID     string
	Name       string
	CreatedAt  time.Time
	CoverURL   *string
	SongsCount int
	Songs      []PlaylistSong
}

type PlaylistSong struct {
	ID              string
	Name            string
	CoverURL        *string
	DurationSeconds int
	StreamURL       string
	IsFavorite      bool
	Artists         []PlaylistArtist
}

type PlaylistArtist struct {
	ID        string
	Name      string
	AvatarURL *string
}
