package songs

type Song struct {
	ID              string
	Name            string
	CoverURL        *string
	DurationSeconds int
	StreamURL       string
	Artists         []ArtistSummary
}

type ArtistSummary struct {
	ID        string
	Name      string
	AvatarURL *string
}
