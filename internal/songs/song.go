package songs

type Song struct {
	ID              string
	Name            string
	CoverMediaID    *string
	DurationSeconds int
	StreamMediaID   string
	Artists         []ArtistSummary
}

type ArtistSummary struct {
	ID            string
	Name          string
	AvatarMediaID *string
}
