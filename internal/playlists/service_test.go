package playlists

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

type fakePlaylistsRepo struct {
	createID  string
	createErr error

	updateErr error
	deleteErr error

	playlists    []PlaylistSummary
	playlistsErr error

	playlist    *Playlist
	playlistErr error

	songs    []PlaylistSong
	songsErr error

	addErr    error
	removeErr error
}

func (r *fakePlaylistsRepo) CreatePlaylist(ctx context.Context, userID, name string) (string, error) {
	return r.createID, r.createErr
}

func (r *fakePlaylistsRepo) UpdatePlaylist(ctx context.Context, userID, playlistID, name string) error {
	return r.updateErr
}

func (r *fakePlaylistsRepo) DeletePlaylist(ctx context.Context, userID, playlistID string) error {
	return r.deleteErr
}

func (r *fakePlaylistsRepo) GetPlaylists(ctx context.Context, userID string) ([]PlaylistSummary, error) {
	return r.playlists, r.playlistsErr
}

func (r *fakePlaylistsRepo) GetPlaylist(ctx context.Context, userID, playlistID string) (*Playlist, error) {
	return r.playlist, r.playlistErr
}

func (r *fakePlaylistsRepo) GetPlaylistSongs(ctx context.Context, userID, playlistID string) ([]PlaylistSong, error) {
	return r.songs, r.songsErr
}

func (r *fakePlaylistsRepo) AddSongToPlaylist(ctx context.Context, userID, playlistID, songID string) error {
	return r.addErr
}

func (r *fakePlaylistsRepo) RemoveSongFromPlaylist(ctx context.Context, userID, playlistID, songID string) error {
	return r.removeErr
}

func TestPlaylistsService_UpdatePlaylist_NotFound(t *testing.T) {
	repo := &fakePlaylistsRepo{
		updateErr: pgx.ErrNoRows,
	}
	svc := NewPlaylistsService(repo)

	err := svc.UpdatePlaylist(context.Background(), "user-id", "playlist-id", "name")
	if !errors.Is(err, ErrPlaylistNotFound) {
		t.Fatalf("expected ErrPlaylistNotFound, got %v", err)
	}
}

func TestPlaylistsService_DeletePlaylist_PropagatesError(t *testing.T) {
	wantErr := errors.New("delete failed")
	repo := &fakePlaylistsRepo{
		deleteErr: wantErr,
	}
	svc := NewPlaylistsService(repo)

	err := svc.DeletePlaylist(context.Background(), "user-id", "playlist-id")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func TestPlaylistsService_GetPlaylist_NotFound(t *testing.T) {
	repo := &fakePlaylistsRepo{
		playlistErr: pgx.ErrNoRows,
	}
	svc := NewPlaylistsService(repo)

	p, err := svc.GetPlaylist(context.Background(), "user-id", "playlist-id")
	if !errors.Is(err, ErrPlaylistNotFound) {
		t.Fatalf("expected ErrPlaylistNotFound, got %v", err)
	}
	if p != nil {
		t.Fatalf("expected nil playlist, got %+v", p)
	}
}

func TestPlaylistsService_GetPlaylistWithSongs_OK(t *testing.T) {
	repo := &fakePlaylistsRepo{
		playlist: &Playlist{
			ID:         "playlist-id",
			UserID:     "user-id",
			Name:       "My playlist",
			SongsCount: 1,
		},
		songs: []PlaylistSong{
			{
				ID:              "song-id",
				Name:            "Song",
				DurationSeconds: 120,
				StreamMediaID:   "stream-id",
				IsFavorite:      true,
			},
		},
	}
	svc := NewPlaylistsService(repo)

	pl, songs, err := svc.GetPlaylistWithSongs(context.Background(), "user-id", "playlist-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pl == nil || pl.ID != "playlist-id" {
		t.Fatalf("unexpected playlist: %+v", pl)
	}
	if len(songs) != 1 || songs[0].ID != "song-id" {
		t.Fatalf("unexpected songs: %+v", songs)
	}
}

func TestPlaylistsService_AddSongToPlaylist_NotFound(t *testing.T) {
	repo := &fakePlaylistsRepo{
		addErr: pgx.ErrNoRows,
	}
	svc := NewPlaylistsService(repo)

	err := svc.AddSongToPlaylist(context.Background(), "user-id", "playlist-id", "song-id")
	if !errors.Is(err, ErrPlaylistNotFound) {
		t.Fatalf("expected ErrPlaylistNotFound, got %v", err)
	}
}
