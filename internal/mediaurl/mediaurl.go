package mediaurl

import "strings"

var basePath = "/media/"

func SetBasePath(path string) {
	path = strings.TrimSpace(path)
	if path == "" {
		basePath = "/media/"
		return
	}

	path = strings.Trim(path, "/")
	basePath = path + "/"
}

func Build(mediaID string) string {
	return basePath + mediaID
}
