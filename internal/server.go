package internal

import (
	"context"

	"log/slog"

	"github.com/witelokk/music-api/internal/artists"
	"github.com/witelokk/music-api/internal/auth"
	"github.com/witelokk/music-api/internal/favorites"
	"github.com/witelokk/music-api/internal/home"
	"github.com/witelokk/music-api/internal/followings"
	"github.com/witelokk/music-api/internal/media"
	openapi "github.com/witelokk/music-api/internal/openapi"
	"github.com/witelokk/music-api/internal/playlists"
	"github.com/witelokk/music-api/internal/search"
	"github.com/witelokk/music-api/internal/releases"
	"github.com/witelokk/music-api/internal/songs"
)

type Server struct {
	authService       *auth.AuthService
	homeService       *home.Service
	songsService      *songs.SongsService
	artistsService    *artists.ArtistsService
	releasesService   *releases.ReleasesService
	favoritesService  *favorites.FavoritesService
	followingsService *followings.FollowingsService
	mediaService      *media.MediaService
	playlistsService  *playlists.PlaylistsService
	searchService     *search.Service
	logger            *slog.Logger
}

func NewServer(
	authService *auth.AuthService,
	homeService *home.Service,
	songsService *songs.SongsService,
	artistsService *artists.ArtistsService,
	releasesService *releases.ReleasesService,
	favoritesService *favorites.FavoritesService,
	followingsService *followings.FollowingsService,
	mediaService *media.MediaService,
	playlistsService *playlists.PlaylistsService,
	searchService *search.Service,
	logger *slog.Logger,
) openapi.StrictServerInterface {
	return &Server{
		authService:       authService,
		homeService:       homeService,
		songsService:      songsService,
		artistsService:    artistsService,
		releasesService:   releasesService,
		favoritesService:  favoritesService,
		followingsService: followingsService,
		mediaService:      mediaService,
		playlistsService:  playlistsService,
		searchService:     searchService,
		logger:            logger,
	}
}

func (s *Server) SendVerificationEmail(ctx context.Context, req openapi.SendVerificationEmailRequestObject) (openapi.SendVerificationEmailResponseObject, error) {
	return auth.HandleSendVerificationEmail(ctx, s.authService, s.logger, req)
}

func (s *Server) CreateUser(ctx context.Context, req openapi.CreateUserRequestObject) (openapi.CreateUserResponseObject, error) {
	return auth.HandleCreateUser(ctx, s.authService, s.logger, req)
}

func (s *Server) GenerateTokens(ctx context.Context, req openapi.GenerateTokensRequestObject) (openapi.GenerateTokensResponseObject, error) {
	return auth.HandleGenerateTokens(ctx, s.authService, s.logger, req)
}

func (s *Server) GetCurrentUser(ctx context.Context, req openapi.GetCurrentUserRequestObject) (openapi.GetCurrentUserResponseObject, error) {
	return auth.HandleGetCurrentUser(ctx, s.authService, s.logger, req)
}

func (s *Server) GetSong(ctx context.Context, req openapi.GetSongRequestObject) (openapi.GetSongResponseObject, error) {
	return songs.HandleGetSong(ctx, s.songsService, s.logger, req)
}

func (s *Server) GetArtist(ctx context.Context, req openapi.GetArtistRequestObject) (openapi.GetArtistResponseObject, error) {
	return artists.HandleGetArtist(ctx, s.artistsService, s.logger, req)
}

func (s *Server) GetRelease(ctx context.Context, req openapi.GetReleaseRequestObject) (openapi.GetReleaseResponseObject, error) {
	return releases.HandleGetRelease(ctx, s.releasesService, s.logger, req)
}

func (s *Server) GetFavorites(ctx context.Context, req openapi.GetFavoritesRequestObject) (openapi.GetFavoritesResponseObject, error) {
	return favorites.HandleGetFavorites(ctx, s.favoritesService, s.logger, req)
}

func (s *Server) AddFavorite(ctx context.Context, req openapi.AddFavoriteRequestObject) (openapi.AddFavoriteResponseObject, error) {
	return favorites.HandleAddFavorite(ctx, s.favoritesService, s.logger, req)
}

func (s *Server) RemoveFavorite(ctx context.Context, req openapi.RemoveFavoriteRequestObject) (openapi.RemoveFavoriteResponseObject, error) {
	return favorites.HandleRemoveFavorite(ctx, s.favoritesService, s.logger, req)
}

func (s *Server) GetFollowings(ctx context.Context, req openapi.GetFollowingsRequestObject) (openapi.GetFollowingsResponseObject, error) {
	return followings.HandleGetFollowings(ctx, s.followingsService, s.logger, req)
}

func (s *Server) FollowArtist(ctx context.Context, req openapi.FollowArtistRequestObject) (openapi.FollowArtistResponseObject, error) {
	return followings.HandleFollowArtist(ctx, s.followingsService, s.logger, req)
}

func (s *Server) UnfollowArtist(ctx context.Context, req openapi.UnfollowArtistRequestObject) (openapi.UnfollowArtistResponseObject, error) {
	return followings.HandleUnfollowArtist(ctx, s.followingsService, s.logger, req)
}

func (s *Server) GetMedia(ctx context.Context, request openapi.GetMediaRequestObject) (openapi.GetMediaResponseObject, error) {
	return media.GetMedia(ctx, s.mediaService, request)
}

func (s *Server) Search(ctx context.Context, req openapi.SearchRequestObject) (openapi.SearchResponseObject, error) {
	return search.HandleSearch(ctx, s.searchService, s.logger, req)
}

func (s *Server) GetPlaylists(ctx context.Context, req openapi.GetPlaylistsRequestObject) (openapi.GetPlaylistsResponseObject, error) {
	return playlists.HandleGetPlaylists(ctx, s.playlistsService, s.logger, req)
}

func (s *Server) CreatePlaylist(ctx context.Context, req openapi.CreatePlaylistRequestObject) (openapi.CreatePlaylistResponseObject, error) {
	return playlists.HandleCreatePlaylist(ctx, s.playlistsService, s.logger, req)
}

func (s *Server) GetPlaylist(ctx context.Context, req openapi.GetPlaylistRequestObject) (openapi.GetPlaylistResponseObject, error) {
	return playlists.HandleGetPlaylist(ctx, s.playlistsService, s.logger, req)
}

func (s *Server) UpdatePlaylist(ctx context.Context, req openapi.UpdatePlaylistRequestObject) (openapi.UpdatePlaylistResponseObject, error) {
	return playlists.HandleUpdatePlaylist(ctx, s.playlistsService, s.logger, req)
}

func (s *Server) DeletePlaylist(ctx context.Context, req openapi.DeletePlaylistRequestObject) (openapi.DeletePlaylistResponseObject, error) {
	return playlists.HandleDeletePlaylist(ctx, s.playlistsService, s.logger, req)
}

func (s *Server) GetPlaylistSongs(ctx context.Context, req openapi.GetPlaylistSongsRequestObject) (openapi.GetPlaylistSongsResponseObject, error) {
	return playlists.HandleGetPlaylistSongs(ctx, s.playlistsService, s.logger, req)
}

func (s *Server) AddSongToPlaylist(ctx context.Context, req openapi.AddSongToPlaylistRequestObject) (openapi.AddSongToPlaylistResponseObject, error) {
	return playlists.HandleAddSongToPlaylist(ctx, s.playlistsService, s.logger, req)
}

func (s *Server) RemoveSongFromPlaylist(ctx context.Context, req openapi.RemoveSongFromPlaylistRequestObject) (openapi.RemoveSongFromPlaylistResponseObject, error) {
	return playlists.HandleRemoveSongFromPlaylist(ctx, s.playlistsService, s.logger, req)
}

func (s *Server) GetHomeScreenLayout(ctx context.Context, req openapi.GetHomeScreenLayoutRequestObject) (openapi.GetHomeScreenLayoutResponseObject, error) {
	return home.HandleGetHomeScreenLayout(ctx, s.homeService, s.logger, req)
}
