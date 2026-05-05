package reanime

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dhilzyi/hianime-cli/internal/common"
	"github.com/dhilzyi/hianime-cli/internal/core"
)

// Original site: 'reanime.to'
// Kuudere is reborning to re-anime. But they have still save host video site

type Reanime struct {
	inputUrl   string
	serverData map[string]string
}

func New(rawUrl string) *Reanime {
	return &Reanime{
		inputUrl:   rawUrl,
		serverData: make(map[string]string),
	}
}

func (r *Reanime) Name() string {
	return "Reanime"
}

func (r *Reanime) GetSearchResults(rawInput string) ([]core.SearchResult, error) {
	return nil, nil
}
func (r *Reanime) GetEpisodes() ([]core.Episode, *core.SeriesData, error) {
	return nil, nil, nil
}
func (r *Reanime) GetServeres(selectedEpisode core.Episode) ([]core.Server, error) {
	return nil, nil
}
func (r *Reanime) GetStreamData(keyServer string) (core.StreamData, error) {
	return core.StreamData{}, nil
}

func getSeriesData(rawURL string) (*core.SeriesData, error) {
	client, err := common.NewSession()
	if err != nil {
		return nil, err
	}
	animeURL := rawURL
	if strings.Contains(rawURL, "watch") {
		animeURL, err = getAnimeIDFromURL(rawURL)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest("GET", animeURL+"/__data.json?x-appkit-invalidated=01", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	svelteRaw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	svelteCleaned, err := DecodeDevalue(svelteRaw)
	if err != nil {
		return nil, err
	}
	cleanedRaw, err := json.Marshal(svelteCleaned[1])
	if err != nil {
		return nil, err
	}
	var rawSeriesData seriesData
	if err := json.Unmarshal(cleanedRaw, &rawSeriesData); err != nil {
		return nil, err
	}
	fmt.Printf("%+v\n", rawSeriesData)

	return &core.SeriesData{
		SeriesUrl: animeURL,
		AnilistID: rawSeriesData.Anime.AnilistID,
		Titles: core.Title{
			EnglishTitle: rawSeriesData.Anime.Title.English,
			RomajiTitle:  rawSeriesData.Anime.Title.Romaji,
			KanjiTitle:   rawSeriesData.Anime.Title.Native,
		},
	}, nil
}
