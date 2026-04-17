package artists

import "time"

type Artist struct {
	ID            string
	Name          string
	AvatarMediaID *string
	CoverMediaID  *string
	Popular       []PopularSong
	Releases      []ArtistRelease
}

type PopularSong struct {
	ID              string
	Name            string
	CoverMediaID    *string
	DurationSeconds int
	StreamMediaID   string
	Artists         []ArtistSummary
}

type ArtistRelease struct {
	ID           string
	Name         string
	CoverMediaID *string
	Type         int
	ReleaseAt    time.Time
}

type ArtistSummary struct {
	ID            string
	Name          string
	AvatarMediaID *string
}
