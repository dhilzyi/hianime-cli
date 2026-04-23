package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dhilzyi/hianime-cli/internal/core"
	"github.com/dhilzyi/hianime-cli/internal/state"
	"github.com/dhilzyi/hianime-cli/providers/animenosub"
)

type FetchResult struct {
	Provider   core.Provider
	SeriesData core.SeriesData
	Episodes   []core.Episode
}

type ResolveParams struct {
	URL       string
	AnilistID int
}

func getProvider(rawURL string) core.Provider {
	if strings.Contains(rawURL, "animenosub") {
		return animenosub.New(rawURL)
	}
	return nil
}

func classifyInput(input string) (InputType, int) {
	if input == "q" {
		return InputCommand, 0
	}

	if i, err := strconv.Atoi(input); err == nil {
		return InputHistoryIndex, i
	}

	return InputURL, 0
}

func handleURLInput(input string) (core.Provider, []core.Episode, core.SeriesData, error) {
	provider := getProvider(input)
	if provider == nil {
		return nil, nil, core.SeriesData{}, fmt.Errorf("no provider found")
	}

	episodes, series, err := provider.GetEpisodes()
	if err != nil {
		return nil, nil, core.SeriesData{}, err
	}

	return provider, episodes, *series, nil
}

func fromCache(entry *CacheEntry, url string) *FetchResult {
	return &FetchResult{
		Provider:   getProvider(url),
		SeriesData: entry.SeriesData,
		Episodes:   entry.Episodes,
	}
}

func ResolveInput(p ResolveParams, cache *Cache) (*FetchResult, error) {
	if p.AnilistID != 0 {
		if entry, ok := cache.byID[p.AnilistID]; ok {
			fmt.Println("Info: cache hit by anilistId")
			return fromCache(entry, p.URL), nil
		}
	}

	slug, err := normalizeAnimeNoSubURL(p.URL)
	if err != nil {
		return nil, err
	}
	if entry, ok := cache.bySlug[slug]; ok {
		fmt.Println("Info: cache hit by unique URL path")

		return fromCache(entry, p.URL), nil
	}

	provider, episodes, series, err := handleURLInput(p.URL)
	if err != nil {
		return nil, err
	}

	cache.bySlug[slug] = &CacheEntry{
		SeriesData: series,
		Episodes:   episodes,
	}

	return &FetchResult{
		Provider:   provider,
		SeriesData: series,
		Episodes:   episodes,
	}, nil
}

func getHistoryByIndex(history []state.History, index int) (*state.History, error) {
	if index <= 0 || index > len(history) {
		return &state.History{}, fmt.Errorf("invalid index")
	}
	return &history[index-1], nil
}

func findOrCreateHistory(histories []state.History, seriesdata core.SeriesData) (*state.History, error) {
	for i := range histories {
		hist := &histories[i]

		if seriesdata.AnilistID != 0 && hist.AnilistID == seriesdata.AnilistID {
			return hist, nil
		}

		if seriesdata.Titles.EnglishTitle != "" &&
			hist.EnglishName == seriesdata.Titles.EnglishTitle {
			return hist, nil
		}

		if seriesdata.Titles.RomajiTitle != "" &&
			hist.JapaneseName == seriesdata.Titles.RomajiTitle {
			return hist, nil
		}
	}

	newHistory := &state.History{
		Url:          seriesdata.SeriesUrl,
		JapaneseName: seriesdata.Titles.RomajiTitle,
		EnglishName:  seriesdata.Titles.EnglishTitle,
		AnilistID:    seriesdata.AnilistID,
		LastEpisode:  1,
		Episode:      make(map[int]state.EpisodeProgress),
	}

	fmt.Println("Info: Create new history complete")
	return newHistory, nil
}
