package releases

import openapi "github.com/witelokk/music-api/internal/openapi"

func MapReleaseType(t int) openapi.ReleaseType {
	switch t {
	case 0:
		return openapi.Single
	case 1:
		return openapi.Ep
	case 2:
		return openapi.Album
	default:
		return openapi.ReleaseType("")
	}
}
