package main

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/dhilzyi/hianime-cli/internal/core"
)

type CacheEntry struct {
	SeriesData core.SeriesData
	Episodes   []core.Episode
}

type Cache struct {
	bySlug map[string]*CacheEntry
	byID   map[int]*CacheEntry
}

func normalizeAnimeNoSubURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}

	path := strings.Trim(u.Path, "/")

	// remove "anime/" prefix if exists
	path = strings.TrimPrefix(path, "anime/")

	// remove episode suffix
	re := regexp.MustCompile(`-episode-\d+/?$`)
	path = re.ReplaceAllString(path, "")

	return path, nil
}
