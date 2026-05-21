package mediaurl

import (
	"net/url"
	"strings"
)

const defaultBasePath = "/api/v1/media/"

var basePath = defaultBasePath

func SetBasePath(path string) {
	path = strings.TrimSpace(path)
	if path == "" {
		basePath = defaultBasePath
		return
	}

	if strings.Contains(path, "://") {
		if parsed, err := url.Parse(path); err == nil && parsed.Scheme != "" && parsed.Host != "" {
			parsed.Path = normalizePath(parsed.Path)
			parsed.RawPath = ""
			basePath = strings.TrimRight(parsed.String(), "/") + "/"
			return
		}
	}

	basePath = normalizePath(path)
}

func Build(mediaID string) string {
	return basePath + mediaID
}

func normalizePath(path string) string {
	path = "/" + strings.Trim(path, "/")
	if path == "/api/media" || strings.HasPrefix(path, "/api/media/") {
		path = strings.Replace(path, "/api/media", "/api/v1/media", 1)
	}
	return strings.TrimRight(path, "/") + "/"
}
