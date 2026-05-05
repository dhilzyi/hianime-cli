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
	anilistID  int
	client     *http.Client
}

func New(rawUrl string) (*Reanime, error) {
	client, err := common.NewSession()
	if err != nil {
		return nil, err
	}
	return &Reanime{
		inputUrl:   rawUrl,
		serverData: make(map[string]string),
		client:     client,
	}, nil
}

func (r *Reanime) Name() string {
	return "Reanime"
}

func (r *Reanime) GetSearchResults(rawInput string) ([]core.SearchResult, error) {
	return nil, nil
}
func (r *Reanime) GetEpisodes() ([]core.Episode, *core.SeriesData, error) {
	seriesData, episodes, err := getSeriesData(r.client, r.inputUrl)
	if err != nil {
		return nil, nil, err
	}
	if seriesData.AnilistID == 0 {
		return nil, nil, fmt.Errorf("anilistid value is 0 and cannot retrieve episodes data")
	}
	r.anilistID = seriesData.AnilistID

	return episodes, seriesData, nil
}

func (r *Reanime) GetServeres(selectedEpisode core.Episode) ([]core.Server, error) {
	baseURL, err := common.GetBaseURL(r.inputUrl)
	if err != nil {
		return nil, err
	}
	serverURL := fmt.Sprintf("%s/api/flix/%d/%d", baseURL, r.anilistID, selectedEpisode.Number)

	servers, serversData, err := getServers(r.client, serverURL)
	if err != nil {
		return nil, err
	}
	r.serverData = serversData
	return servers, nil
}
func (r *Reanime) GetStreamData(keyServer string) (core.StreamData, error) {
	return core.StreamData{}, nil
}

func getSeriesData(client *http.Client, rawURL string) (*core.SeriesData, []core.Episode, error) {
	animeURL := rawURL
	var err error
	if strings.Contains(rawURL, "watch") {
		animeURL, err = getAnimeIDFromURL(rawURL)
		if err != nil {
			return nil, nil, err
		}
	}
	req, err := http.NewRequest("GET", animeURL+"/__data.json?x-appkit-invalidated=01", nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	svelteRaw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	svelteCleaned, err := DecodeDevalue(svelteRaw)
	if err != nil {
		return nil, nil, err
	}
	cleanedRaw, err := json.Marshal(svelteCleaned[1])
	if err != nil {
		return nil, nil, err
	}
	var rawSeriesData seriesDataResponse
	if err := json.Unmarshal(cleanedRaw, &rawSeriesData); err != nil {
		return nil, nil, err
	}

	var episodes []core.Episode
	for _, eps := range rawSeriesData.Episodes.Data {
		episodes = append(episodes, core.Episode{
			Number: eps.Number,
			Titles: core.Title{
				RomajiTitle:  eps.TitleJapanese,
				EnglishTitle: eps.Title,
			},
		})
	}
	return &core.SeriesData{
		SeriesUrl: animeURL,
		AnilistID: rawSeriesData.Anime.AnilistID,
		Titles: core.Title{
			EnglishTitle: rawSeriesData.Anime.Title.English,
			RomajiTitle:  rawSeriesData.Anime.Title.Romaji,
			KanjiTitle:   rawSeriesData.Anime.Title.Native,
		},
	}, episodes, nil
}

func getServers(client *http.Client, serverURL string) ([]core.Server, map[string]string, error) {
	req, err := http.NewRequest("GET", serverURL, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	var rawServers serverApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&rawServers); err != nil {
		return nil, nil, err
	}
	if !rawServers.Success || len(rawServers.Servers) == 0 {
		return nil, nil, fmt.Errorf("failed to retrieve servers data")
	}
	var servers []core.Server
	data := make(map[string]string)
	for _, srv := range rawServers.Servers {
		servers = append(servers, core.Server{
			Name: srv.ServerName,
			Type: srv.DataType,
		})
		data[srv.ID] = srv.DataLink
	}

	return servers, data, nil
}
