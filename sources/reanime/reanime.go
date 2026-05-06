package reanime

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/dhilzyi/hianime-cli/hosts/zencloudz"
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

func (r *Reanime) GetServers(selectedEpisode core.Episode) ([]core.Server, error) {
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
	serverURL, ok := r.serverData[keyServer]
	if !ok {
		return core.StreamData{}, fmt.Errorf("no server data value found in map for key: %s", keyServer)
	}
	serverURL = strings.Replace(serverURL, "flixcloud", "zencloudz", 1)
	streamdata, err := zerocloudz.GetStreamData(serverURL)
	if err != nil {
		return core.StreamData{}, err
	}

	return streamdata, nil
}

func (r *Reanime) ExtractProviderID() (string, error) {
	id, err := getAnimeIDFromURL(r.inputUrl)
	if err != nil {
		return "", err
	}

	return id, nil
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
			Key:  srv.ID,
		})
		data[srv.ID] = srv.DataLink
	}
	sort.SliceStable(servers, func(i, j int) bool {
		si, sj := servers[i], servers[j]

		// sub before dub
		typeRank := map[string]int{"sub": 0, "dub": 1}
		ti, tj := typeRank[si.Type], typeRank[sj.Type]
		if ti != tj {
			return ti < tj
		}

		// within same type: HD-2 before HD-1, everything else after
		nameRank := map[string]int{"HD-2": 0, "HD-1": 1}
		ri, rj := nameRank[si.Name], nameRank[sj.Name]
		// unknown names get rank 2
		if _, ok := nameRank[si.Name]; !ok {
			ri = 2
		}
		if _, ok := nameRank[sj.Name]; !ok {
			rj = 2
		}

		return ri < rj
	})

	return servers, data, nil
}
