package search

import (
	"context"
	"errors"
	"testing"
)

type fakeSearchRepo struct {
	songs        []SongResult
	artists      []ArtistResult
	releases     []ReleaseResult
	playlists    []PlaylistResult
	songsErr     error
	artistsErr   error
	releasesErr  error
	playlistsErr error
}

func (r *fakeSearchRepo) SearchSongs(ctx context.Context, query, userID string) ([]SongResult, error) {
	return r.songs, r.songsErr
}

func (r *fakeSearchRepo) SearchArtists(ctx context.Context, query string) ([]ArtistResult, error) {
	return r.artists, r.artistsErr
}

func (r *fakeSearchRepo) SearchReleases(ctx context.Context, query string) ([]ReleaseResult, error) {
	return r.releases, r.releasesErr
}

func (r *fakeSearchRepo) SearchPlaylists(ctx context.Context, query, userID string) ([]PlaylistResult, error) {
	return r.playlists, r.playlistsErr
}

func TestService_Search_AllTypes(t *testing.T) {
	repo := &fakeSearchRepo{
		songs: []SongResult{
			{ID: "song-1", Name: "Alpha Song"},
		},
		artists: []ArtistResult{
			{ID: "artist-1", Name: "Beta Artist"},
		},
		releases: []ReleaseResult{
			{ID: "release-1", Name: "Gamma Release"},
		},
		playlists: []PlaylistResult{
			{ID: "playlist-1", Name: "Delta Playlist"},
		},
	}
	svc := NewService(repo)

	res, err := svc.Search(context.Background(), "test", nil, 1, 10, "user-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Total != 4 {
		t.Fatalf("expected total=4, got %d", res.Total)
	}
	if len(res.Items) != 4 {
		t.Fatalf("expected 4 items, got %d", len(res.Items))
	}
}

func TestService_Search_FilterType(t *testing.T) {
	repo := &fakeSearchRepo{
		songs: []SongResult{
			{ID: "song-1", Name: "Alpha Song"},
		},
		artists: []ArtistResult{
			{ID: "artist-1", Name: "Beta Artist"},
		},
	}
	svc := NewService(repo)

	typ := ResultTypeArtist
	res, err := svc.Search(context.Background(), "test", &typ, 1, 10, "user-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Total != 1 {
		t.Fatalf("expected total=1, got %d", res.Total)
	}
	if len(res.Items) != 1 || res.Items[0].Type != ResultTypeArtist {
		t.Fatalf("expected one artist result, got %+v", res.Items)
	}
}

func TestService_Search_Pagination(t *testing.T) {
	var songs []SongResult
	for i := 0; i < 5; i++ {
		songs = append(songs, SongResult{ID: "song", Name: string(rune('A' + i))})
	}
	repo := &fakeSearchRepo{
		songs: songs,
	}
	svc := NewService(repo)

	res, err := svc.Search(context.Background(), "test", nil, 2, 2, "user-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Page != 2 || res.Limit != 2 {
		t.Fatalf("unexpected page/limit: %+v", res)
	}
	if len(res.Items) != 2 {
		t.Fatalf("expected 2 items on second page, got %d", len(res.Items))
	}
}

func TestService_Search_RepoError(t *testing.T) {
	repoErr := errors.New("repo error")
	repo := &fakeSearchRepo{
		songsErr: repoErr,
	}
	svc := NewService(repo)

	_, err := svc.Search(context.Background(), "test", nil, 1, 10, "user-id")
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

