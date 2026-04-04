package releases

import "time"

type Release struct {
	ID        string
	Name      string
	CoverURL  *string
	Type      int
	ReleaseAt time.Time
}

