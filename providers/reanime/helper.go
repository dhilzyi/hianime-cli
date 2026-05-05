package reanime

import (
	"fmt"
	"net/url"
	"strings"
)

func getAnimeIDFromURL(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	path := strings.Trim(parsed.Path, "/")
	path = strings.TrimPrefix(path, "watch/")

	build := fmt.Sprintf("%s://%s/anime/%s", parsed.Scheme, parsed.Host, path)

	return build, err
}
