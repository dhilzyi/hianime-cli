package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dhilzyi/hianime-cli/internal/core"
	"github.com/dhilzyi/hianime-cli/internal/state"
	"github.com/dhilzyi/hianime-cli/providers/animenosub"
)

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

func handleURLInput(input string, history []state.History) (core.Provider, []core.Episode, state.History, core.SeriesData, error) {
	provider := getProvider(input)
	if provider == nil {
		return nil, nil, state.History{}, core.SeriesData{}, fmt.Errorf("no provider found")
	}

	episodes, series, err := provider.GetEpisodes()
	if err != nil {
		return nil, nil, state.History{}, core.SeriesData{}, err
	}

	newHistory := state.History{
		Url:          series.SeriesUrl,
		JapaneseName: series.Titles.RomajiTitle,
		EnglishName:  series.Titles.EnglishTitle,
		AnilistID:    series.AnilistID,
		LastEpisode:  1,
		Episode:      make(map[int]state.EpisodeProgress),
	}

	return provider, episodes, newHistory, *series, nil
}

func handleHistoryInput(index int, history []state.History) (core.Provider, []core.Episode, state.History, core.SeriesData, error) {
	if index <= 0 || index > len(history) {
		return nil, nil, state.History{}, core.SeriesData{}, fmt.Errorf("invalid index")
	}

	selected := history[index-1]

	provider := getProvider(selected.Url)
	if provider == nil {
		return nil, nil, state.History{}, core.SeriesData{}, fmt.Errorf("no provider found")
	}

	episodes, series, err := provider.GetEpisodes()
	if err != nil {
		return nil, nil, state.History{}, core.SeriesData{}, err
	}

	return provider, episodes, selected, *series, nil
}
