package releases

import "time"

type Release struct {
	ID        string
	Name      string
	CoverURL  *string
	Type      int
	ReleaseAt time.Time
	Songs     []ReleaseSong
	Artists   []ReleaseArtist
}

type ReleaseSong struct {
	ID              string
	Name            string
	CoverURL        *string
	DurationSeconds int
	StreamURL       string
	Artists         []ReleaseArtist
}

type ReleaseArtist struct {
	ID        string
	Name      string
	AvatarURL *string
}
