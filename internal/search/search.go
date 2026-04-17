package search

type ResultType string

const (
	ResultTypeSong     ResultType = "song"
	ResultTypeArtist   ResultType = "artist"
	ResultTypeRelease  ResultType = "release"
	ResultTypePlaylist ResultType = "playlist"
)

type SongResult struct {
	ID              string
	Name            string
	CoverMediaID    *string
	DurationSeconds int
	StreamMediaID   string
	IsFavorite      bool
	Artists         []ArtistSummary
}

type ArtistSummary struct {
	ID            string
	Name          string
	AvatarMediaID *string
}

type ArtistResult struct {
	ID            string
	Name          string
	AvatarMediaID *string
}

type ReleaseResult struct {
	ID           string
	Name         string
	CoverMediaID *string
	Type         int
	ReleaseAt    string
}

type PlaylistResult struct {
	ID           string
	Name         string
	CoverMediaID *string
	SongsCount   int
}

type ResultItem struct {
	Type     ResultType
	Song     *SongResult
	Artist   *ArtistResult
	Release  *ReleaseResult
	Playlist *PlaylistResult
	Name     string
}

type Results struct {
	Query string
	Page  int
	Limit int
	Total int
	Items []ResultItem
}
