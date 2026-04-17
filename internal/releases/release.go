package releases

import "time"

type Release struct {
	ID           string
	Name         string
	CoverMediaID *string
	Type         int
	ReleaseAt    time.Time
	Songs        []ReleaseSong
	Artists      []ReleaseArtist
}

type ReleaseSong struct {
	ID              string
	Name            string
	CoverMediaID    *string
	DurationSeconds int
	StreamMediaID   string
	Artists         []ReleaseArtist
}

type ReleaseArtist struct {
	ID            string
	Name          string
	AvatarMediaID *string
}
