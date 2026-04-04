package songs

type Song struct {
	ID              string
	Name            string
	CoverURL        *string
	DurationSeconds int
	StreamURL       string
}

