package releases

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/witelokk/music-api/internal/mediaurl"
	openapi "github.com/witelokk/music-api/internal/openapi"
	"github.com/witelokk/music-api/internal/requestctx"
)

func HandleGetRelease(
	ctx context.Context,
	releasesService *ReleasesService,
	logger *slog.Logger,
	req openapi.GetReleaseRequestObject,
) (openapi.GetReleaseResponseObject, error) {
	reqLogger := requestctx.LoggerFromContext(ctx, logger)

	id := req.Id

	release, err := releasesService.GetRelease(ctx, id.String())
	if err != nil {
		if errors.Is(err, ErrReleaseNotFound) {
			return openapi.GetRelease404JSONResponse(openapi.Error{Error: "release not found"}), nil
		}

		reqLogger.Error("failed to get release",
			slog.String("id", id.String()),
			slog.String("error", err.Error()),
		)
		return openapi.GetRelease500JSONResponse(openapi.Error{Error: "failed to fetch release"}), nil
	}

	songs := make([]openapi.Song, 0, len(release.Songs))
	for _, s := range release.Songs {
		artistSummaries := make([]openapi.ArtistSummary, 0, len(s.Artists))
		for _, a := range s.Artists {
			summary := openapi.ArtistSummary{
				Id:   uuid.MustParse(a.ID),
				Name: a.Name,
			}
			if a.AvatarMediaID != nil && *a.AvatarMediaID != "" {
				avatarURL := mediaurl.Build(*a.AvatarMediaID)
				summary.AvatarUrl = &avatarURL
			}
			artistSummaries = append(artistSummaries, summary)
		}

		song := openapi.Song{
			Id:              uuid.MustParse(s.ID),
			Name:            s.Name,
			DurationSeconds: s.DurationSeconds,
			StreamUrl:       mediaurl.Build(s.StreamMediaID),
			IsFavorite:      false,
			Artists:         artistSummaries,
		}
		if s.CoverMediaID != nil && *s.CoverMediaID != "" {
			coverURL := mediaurl.Build(*s.CoverMediaID)
			song.CoverUrl = &coverURL
		}
		songs = append(songs, song)
	}

	artistSummaries := make([]openapi.ArtistSummary, 0, len(release.Artists))
	artistNames := make([]string, 0, len(release.Artists))
	for _, a := range release.Artists {
		summary := openapi.ArtistSummary{
			Id:   uuid.MustParse(a.ID),
			Name: a.Name,
		}
		if a.AvatarMediaID != nil && *a.AvatarMediaID != "" {
			avatarURL := mediaurl.Build(*a.AvatarMediaID)
			summary.AvatarUrl = &avatarURL
		}
		artistSummaries = append(artistSummaries, summary)
		artistNames = append(artistNames, a.Name)
	}

	resp := openapi.Release{
		Id:         uuid.MustParse(release.ID),
		Name:       release.Name,
		Type:       MapReleaseType(release.Type),
		ReleasedAt: release.ReleaseAt.Format("2006-01-02"),
		Songs: openapi.SongList{
			Count: len(songs),
			Songs: songs,
		},
		Artists: openapi.ArtistList{
			Count:   len(artistSummaries),
			Artists: artistSummaries,
			Names:   strings.Join(artistNames, ", "),
		},
	}

	if release.CoverMediaID != nil && *release.CoverMediaID != "" {
		coverURL := mediaurl.Build(*release.CoverMediaID)
		resp.CoverUrl = &coverURL
	}

	return openapi.GetRelease200JSONResponse(resp), nil
}
