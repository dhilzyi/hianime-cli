package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/dhilzyi/hianime-cli/internal/anilist"
	"github.com/dhilzyi/hianime-cli/internal/core"
	"github.com/dhilzyi/hianime-cli/internal/state"
	"github.com/dhilzyi/hianime-cli/providers/animenosub"
	"github.com/dhilzyi/hianime-cli/providers/kuudere"
)

type FetchResult struct {
	Provider   core.Provider
	SeriesData core.SeriesData
	Episodes   []core.Episode
}

type resolveParams struct {
	URL       string
	AnilistID int
}

type ProviderType int

const (
	UnknownProvider ProviderType = iota
	AnimeNoSub
	Kuudere
)

func getProvider(p resolveParams) (core.Provider, ProviderType) {
	if strings.Contains(p.URL, "animenosub") {
		return animenosub.New(p.URL), AnimeNoSub
	} else if strings.Contains(p.URL, "kuudere") || p.AnilistID != 0 {
		return kuudere.New(p.URL, p.AnilistID), Kuudere
	}
	return nil, UnknownProvider
}

func classifyInput(input string) (InputType, int) {
	if input == "q" {
		return InputCommand, 0
	}

	if i, err := strconv.Atoi(input); err == nil {
		if i >= 21 {
			return InputAnilistID, i
		} else {
			return InputHistoryIndex, i
		}
	}

	return InputURL, 0
}

func handleURLInput(p resolveParams, provider core.Provider) ([]core.Episode, core.SeriesData, error) {
	episodes, series, err := provider.GetEpisodes()
	if err != nil {
		return nil, core.SeriesData{}, err
	}

	return episodes, *series, nil
}

func fromCache(entry *CacheEntry, p resolveParams) *FetchResult {
	provider, _ := getProvider(p)
	return &FetchResult{
		Provider:   provider,
		SeriesData: entry.SeriesData,
		Episodes:   entry.Episodes,
	}
}

func ResolveInput(p resolveParams, cache *Cache) (*FetchResult, error) {
	if p.AnilistID != 0 {
		if entry, ok := cache.byID[p.AnilistID]; ok {
			fmt.Println("Info: cache hit by anilistId")
			return fromCache(entry, p), nil
		}
	}

	provider, providerType := getProvider(p)
	if provider == nil {
		return nil, fmt.Errorf("provider is not found")
	}

	var slug string
	var err error
	if providerType == AnimeNoSub && p.URL != "" {
		slug, err = normalizeAnimeNoSubURL(p.URL)
		if err != nil {
			return nil, err
		}
		if entry, ok := cache.bySlug[slug]; ok {
			fmt.Println("Info: cache hit by unique URL path")

			return fromCache(entry, p), nil
		}
	}

	episodes, series, err := handleURLInput(p, provider)
	if err != nil {
		return nil, err
	}

	entry := &CacheEntry{
		SeriesData: series,
		Episodes:   episodes,
	}

	if slug != "" {
		cache.bySlug[slug] = entry
	}
	if series.AnilistID != 0 {
		cache.byID[series.AnilistID] = entry
	} else if p.AnilistID != 0 {
		cache.byID[p.AnilistID] = entry
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

func findOrCreateHistory(histories []state.History, seriesdata core.SeriesData, providerName string) (*state.History, error) {
	for i := range histories {
		hist := &histories[i]

		if seriesdata.AnilistID != 0 && hist.AnilistID == seriesdata.AnilistID {
			fmt.Println("Info: history hit by anilistID")
			return hist, nil
		}

		if seriesdata.Titles.EnglishTitle != "" &&
			hist.EnglishName == seriesdata.Titles.EnglishTitle {
			fmt.Println("Info: history hit by english title")
			return hist, nil
		}

		if seriesdata.Titles.RomajiTitle != "" &&
			hist.JapaneseName == seriesdata.Titles.RomajiTitle {
			fmt.Println("Info: history hit by romaji title")
			return hist, nil
		}
	}

	needsMetadata := seriesdata.AnilistID == 0 ||
		(seriesdata.Titles.EnglishTitle == "" && seriesdata.Titles.RomajiTitle == "")

	if needsMetadata {
		if err := anilist.FillSeriesData(&seriesdata); err == nil {
			fmt.Println("Info: successfully filled missing metadata from Anilist")
		} else {
			log.Println("Warning: failed to fill metadata:", err)
		}
	}

	newHistory := &state.History{
		Url:          seriesdata.SeriesUrl,
		JapaneseName: seriesdata.Titles.RomajiTitle,
		EnglishName:  seriesdata.Titles.EnglishTitle,
		AnilistID:    seriesdata.AnilistID,
		LastEpisode:  1,
		Episode:      make(map[int]state.EpisodeProgress),
		Provider:     providerName,
	}

	fmt.Println("Info: Create new history complete")
	return newHistory, nil
}

func fillUpSeriesDataWithHistory(seriesdata *core.SeriesData, history state.History) error {
	seriesdata.Titles.EnglishTitle = history.EnglishName
	seriesdata.Titles.RomajiTitle = history.JapaneseName
	seriesdata.AnilistID = history.AnilistID
	seriesdata.SeriesUrl = history.Url

	fmt.Println("Info: filling up seriesdata by history successfully")
	return nil
}
