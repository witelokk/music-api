package playlists

import (
	"context"
	"errors"
)

var ErrPlaylistNotFound = errors.New("playlist not found")
var ErrSongNotFound = errors.New("song not found")

type PlaylistsService struct {
	repo PlaylistsRepository
}

func NewPlaylistsService(repo PlaylistsRepository) *PlaylistsService {
	return &PlaylistsService{repo: repo}
}

func (s *PlaylistsService) CreatePlaylist(ctx context.Context, userID, name string) (string, error) {
	return s.repo.CreatePlaylist(ctx, userID, name)
}

func (s *PlaylistsService) UpdatePlaylist(ctx context.Context, userID, playlistID, name string) error {
	if err := s.repo.UpdatePlaylist(ctx, userID, playlistID, name); err != nil {
		if errors.Is(err, ErrPlaylistNotFound) {
			return ErrPlaylistNotFound
		}
		return err
	}
	return nil
}

func (s *PlaylistsService) DeletePlaylist(ctx context.Context, userID, playlistID string) error {
	if err := s.repo.DeletePlaylist(ctx, userID, playlistID); err != nil {
		if errors.Is(err, ErrPlaylistNotFound) {
			return ErrPlaylistNotFound
		}
		return err
	}
	return nil
}

func (s *PlaylistsService) GetPlaylists(ctx context.Context, userID string) ([]PlaylistSummary, error) {
	return s.repo.GetPlaylists(ctx, userID)
}

func (s *PlaylistsService) GetPlaylist(ctx context.Context, userID, playlistID string) (*Playlist, error) {
	p, err := s.repo.GetPlaylist(ctx, userID, playlistID)
	if err != nil {
		if errors.Is(err, ErrPlaylistNotFound) {
			return nil, ErrPlaylistNotFound
		}
		return nil, err
	}
	return p, nil
}

func (s *PlaylistsService) GetPlaylistSongs(ctx context.Context, userID, playlistID string) ([]PlaylistSong, error) {
	if _, err := s.GetPlaylist(ctx, userID, playlistID); err != nil {
		return nil, err
	}

	songs, err := s.repo.GetPlaylistSongs(ctx, userID, playlistID)
	if err != nil {
		return nil, err
	}
	return songs, nil
}

func (s *PlaylistsService) GetPlaylistWithSongs(ctx context.Context, userID, playlistID string) (*Playlist, []PlaylistSong, error) {
	p, err := s.GetPlaylist(ctx, userID, playlistID)
	if err != nil {
		return nil, nil, err
	}
	songs, err := s.repo.GetPlaylistSongs(ctx, userID, playlistID)
	if err != nil {
		return nil, nil, err
	}
	return p, songs, nil
}

func (s *PlaylistsService) AddSongToPlaylist(ctx context.Context, userID, playlistID, songID string) error {
	if err := s.repo.AddSongToPlaylist(ctx, userID, playlistID, songID); err != nil {
		if errors.Is(err, ErrPlaylistNotFound) {
			return ErrPlaylistNotFound
		}
		if errors.Is(err, ErrSongNotFound) {
			return ErrSongNotFound
		}
		return err
	}
	return nil
}

func (s *PlaylistsService) RemoveSongFromPlaylist(ctx context.Context, userID, playlistID, songID string) error {
	if err := s.repo.RemoveSongFromPlaylist(ctx, userID, playlistID, songID); err != nil {
		if errors.Is(err, ErrPlaylistNotFound) {
			return ErrPlaylistNotFound
		}
		if errors.Is(err, ErrSongNotFound) {
			return ErrSongNotFound
		}
		return err
	}
	return nil
}
