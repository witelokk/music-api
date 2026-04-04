package artists

import "time"

type Artist struct {
	ID        string
	Name      string
	AvatarURL *string
	CoverURL  *string
	Popular   []PopularSong
	Releases  []ArtistRelease
}

type PopularSong struct {
	ID              string
	Name            string
	CoverURL        *string
	DurationSeconds int
	StreamURL       string
}

type ArtistRelease struct {
	ID        string
	Name      string
	CoverURL  *string
	Type      int
	ReleaseAt time.Time
}
