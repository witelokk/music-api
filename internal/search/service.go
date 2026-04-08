package search

import (
	"context"
	"sort"
)

type Service struct {
	repo SearchRepository
}

func NewService(repo SearchRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Search(
	ctx context.Context,
	query string,
	searchType *ResultType,
	page, limit int,
	userID string,
) (*Results, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	var items []ResultItem

	addSongs := searchType == nil || *searchType == ResultTypeSong
	addArtists := searchType == nil || *searchType == ResultTypeArtist
	addReleases := searchType == nil || *searchType == ResultTypeRelease
	addPlaylists := searchType == nil || *searchType == ResultTypePlaylist

	if addSongs {
		songs, err := s.repo.SearchSongs(ctx, query, userID)
		if err != nil {
			return nil, err
		}
		for _, song := range songs {
			snapshot := song
			item := ResultItem{
				Type: ResultTypeSong,
				Song: &snapshot,
				Name: snapshot.Name,
			}
			items = append(items, item)
		}
	}

	if addArtists {
		artists, err := s.repo.SearchArtists(ctx, query)
		if err != nil {
			return nil, err
		}
		for _, artist := range artists {
			snapshot := artist
			item := ResultItem{
				Type:   ResultTypeArtist,
				Artist: &snapshot,
				Name:   artist.Name,
			}
			items = append(items, item)
		}
	}

	if addReleases {
		releases, err := s.repo.SearchReleases(ctx, query)
		if err != nil {
			return nil, err
		}
		for _, release := range releases {
			snapshot := release
			item := ResultItem{
				Type:    ResultTypeRelease,
				Release: &snapshot,
				Name:    release.Name,
			}
			items = append(items, item)
		}
	}

	if addPlaylists {
		playlists, err := s.repo.SearchPlaylists(ctx, query, userID)
		if err != nil {
			return nil, err
		}
		for _, playlist := range playlists {
			snapshot := playlist
			item := ResultItem{
				Type:     ResultTypePlaylist,
				Playlist: &snapshot,
				Name:     playlist.Name,
			}
			items = append(items, item)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Name == items[j].Name {
			return items[i].Type < items[j].Type
		}
		return items[i].Name < items[j].Name
	})

	total := len(items)
	offset := (page - 1) * limit
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}

	paged := items[offset:end]

	return &Results{
		Query: query,
		Page:  page,
		Limit: limit,
		Total: total,
		Items: paged,
	}, nil
}
