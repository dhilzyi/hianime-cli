package main

import (
	"fmt"
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
	byProviderID map[string]*CacheEntry
	byAnilistID  map[int]*CacheEntry
}

func extractAnimeNoSubID(raw string) (string, error) {
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

func extractKuudereID(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	path := strings.Trim(u.Path, "/")
	splitted := strings.Split(path, "/")
	if len(splitted) < 2 {
		return "", fmt.Errorf("invalid url. cannot be normalize")
	}

	return splitted[1], nil
}
