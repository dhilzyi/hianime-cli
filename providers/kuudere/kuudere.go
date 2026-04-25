package kuudere

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"github.com/dhilzyi/hianime-cli/internal/core"
)

// original site: 'https://kuudere.ru/'
// others domains: 'https://kuudere.ru/', 'https://kuudere.to/', 'https://kuudere.lol/'

type Kuudere struct {
	inputUrl   string
	anilistID  int
	serverData map[string]string
}

func New(rawUrl string, inputAnilistID int) *Kuudere {
	return &Kuudere{
		inputUrl:   rawUrl,
		serverData: make(map[string]string),
		anilistID:  inputAnilistID,
	}
}

// func (k *Kuudere) GetSeriesData() (core.SeriesData, error) {
// 	seriesdata, err := getSeriesData(k.inputUrl)
// 	if err != nil {
// 		return core.SeriesData{}, err
// 	}
// 	k.anilistID = seriesdata.AnilistID
// 	return seriesdata, nil
// }

func (k *Kuudere) Name() string {
	return "kuudere"
}

func (k *Kuudere) GetEpisodes() ([]core.Episode, *core.SeriesData, error) {
	var seriesdata core.SeriesData
	var err error
	if k.anilistID == 0 {
		seriesdata, err = getSeriesData(k.inputUrl)
		if err != nil {
			return nil, nil, err
		}
		if seriesdata.AnilistID == 0 {
			return nil, nil, fmt.Errorf("anilistid is 0 value and cant not retrieve episodes data")
		}
		k.anilistID = seriesdata.AnilistID
	}
	seriesdata.AnilistID = k.anilistID

	episodes, err := getEpisodes(k.anilistID)
	if err != nil {
		return nil, nil, err
	}
	return episodes, &seriesdata, nil
}

func (k *Kuudere) GetServers(selectedEpisode core.Episode) ([]core.Server, error) {
	var servers []core.Server
	serverName := "Zerocloud"
	k.serverData[serverName] = selectedEpisode.Url
	servers = append(servers, core.Server{
		Name: serverName,
	})
	return servers, nil
}

func (k *Kuudere) GetStreamData(keyServer string) (core.StreamData, error) {
	url, ok := k.serverData[keyServer]
	if !ok {
		return core.StreamData{}, fmt.Errorf("url server for following episode does not exist")
	}
	streamdata, err := getStreamData(url)
	if err != nil {
		return core.StreamData{}, err
	}
	return streamdata, nil
}

func getSeriesData(rawUrl string) (core.SeriesData, error) {
	req, err := http.NewRequest("GET", rawUrl, nil)
	if err != nil {
		return core.SeriesData{}, err
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return core.SeriesData{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return core.SeriesData{}, fmt.Errorf("bad status %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	var kuudereResp seriesDataResponse
	if err := json.NewDecoder(resp.Body).Decode(&kuudereResp); err != nil {
		return core.SeriesData{}, err
	}

	seriesurl, err := extractSeriesUrl(rawUrl, kuudereResp.AnimeInfo.URL)
	if err != nil {
		return core.SeriesData{}, err
	}

	return core.SeriesData{
		AnilistID: kuudereResp.AnimeInfo.Anilist,
		SeriesUrl: seriesurl,
		Titles: core.Title{
			EnglishTitle: kuudereResp.AnimeInfo.English,
			RomajiTitle:  kuudereResp.AnimeInfo.Romaji,
			KanjiTitle:   kuudereResp.AnimeInfo.Native,
		},
	}, nil
}

func getEpisodes(anilistID int) ([]core.Episode, error) {
	episodesEndpoint := "https://zencloud.cc/videos/raw"
	req, err := http.NewRequest("GET", episodesEndpoint, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Add("anilist_id", fmt.Sprintf("%d", anilistID))
	query.Add("a", fmt.Sprintf("%d", 0))
	req.URL.RawQuery = query.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	var episodeResp episodesResponse
	if err := json.NewDecoder(resp.Body).Decode(&episodeResp); err != nil {
		return nil, err
	}

	if episodeResp.Status != "success" {
		return nil, fmt.Errorf("status not success")
	}

	var episodes []core.Episode
	for _, eps := range episodeResp.Data {
		episodes = append(episodes, core.Episode{
			Url:    eps.PlayerURL,
			Number: eps.Episode,
			Titles: core.Title{EnglishTitle: fmt.Sprintf("Episode %d", eps.Episode)},
		})
	}

	slices.Reverse(episodes)

	return episodes, nil
}
