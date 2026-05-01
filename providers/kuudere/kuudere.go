package kuudere

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/dhilzyi/hianime-cli/internal/common"
	"github.com/dhilzyi/hianime-cli/internal/core"
)

// original site: 'https://kuudere.ru/'
// others domains: 'https://kuudere.ru/', 'https://kuudere.to/', 'https://kuudere.lol/'

type Kuudere struct {
	inputUrl   string
	anilistID  int
	serverData map[string]string
	domains    []string
}

func New(rawUrl string, inputAnilistID int) *Kuudere {
	return &Kuudere{
		inputUrl:   rawUrl,
		serverData: make(map[string]string),
		anilistID:  inputAnilistID,
		domains:    []string{"to", "ru", "lol"},
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
	return "Kuudere"
}

func (k *Kuudere) GetSearchResults(rawInput string) ([]core.SearchResult, error) {
	keywordVal := common.StringToQueryFormat(rawInput)
	var searchResult []core.SearchResult
	for _, domain := range k.domains {
		url := fmt.Sprintf("https://kuudere.%s/search", domain)

		baseUrl, err := common.GetBaseURL(url)
		if err != nil {
			fmt.Printf("Error: failed to get base url: %s\n", url)
			continue
		}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("Error: failed to create new request for: %s\n", url)
			continue
		}
		query := req.URL.Query()
		query.Add("keyword", keywordVal)

		req.URL.RawQuery = query.Encode()

		searchResult, err = fetchQuerySearch(req, baseUrl)
		if err != nil {
			fmt.Printf("Error: failed to fetch search result: %s", url)
		} else {
			break
		}
	}

	return searchResult, nil
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
			return nil, nil, fmt.Errorf("anilistid value is 0 and cannot retrieve episodes data")
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

	var seriesdata core.SeriesData
	urlType := getUrlType(rawUrl)
	switch urlType {
	case AnimeType:
		var animeData animeDataResponse
		if err := json.NewDecoder(resp.Body).Decode(&animeData); err != nil {
			return core.SeriesData{}, err
		}
		if err := fillUpSeriesData(&seriesdata, animeData, rawUrl); err != nil {
			return core.SeriesData{}, err
		}
	case WatchType:
		var watchData watchDataResponse
		if err := json.NewDecoder(resp.Body).Decode(&watchData); err != nil {
			return core.SeriesData{}, err
		}
		seriesurl, err := extractSeriesUrl(rawUrl, watchData.AnimeMetadata.URL)
		if err != nil {
			fmt.Println("Warning: cannot be able to extract series url")
		}

		if err := fillUpSeriesData(&seriesdata, watchData, seriesurl); err != nil {
			return core.SeriesData{}, err
		}

	}

	return seriesdata, nil
}

func getEpisodes(anilistID int) ([]core.Episode, error) {
	episodesEndpoint := "https://zencloudz.cc/videos/raw"
	req, err := http.NewRequest("GET", episodesEndpoint, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Add("anilist_id", fmt.Sprintf("%d", anilistID))
	query.Add("a", fmt.Sprintf("%d", 0))
	req.URL.RawQuery = query.Encode()

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

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

	sort.Slice(episodes, func(i, j int) bool {
		return episodes[i].Number < episodes[j].Number
	})

	return episodes, nil
}

func getUrlType(rawUrl string) UrlType {
	if strings.Contains(rawUrl, "anime") {
		return AnimeType
	} else if strings.Contains(rawUrl, "watch") {
		return WatchType
	}

	return InvalidUrlType
}

func fillUpSeriesData(s *core.SeriesData, rawResponse any, sUrl string) error {
	s.SeriesUrl = sUrl

	switch data := rawResponse.(type) {
	case animeDataResponse:
		s.AnilistID = data.AnimeMetadata.AnilistID
		s.Titles.EnglishTitle = data.AnimeMetadata.English
		s.Titles.RomajiTitle = data.AnimeMetadata.Romaji
		s.Titles.KanjiTitle = data.AnimeMetadata.Native

	case watchDataResponse:
		s.AnilistID = data.AnimeMetadata.AnilistID
		s.Titles.EnglishTitle = data.AnimeMetadata.English
		s.Titles.RomajiTitle = data.AnimeMetadata.Romaji
		s.Titles.KanjiTitle = data.AnimeMetadata.Native

	default:
		return fmt.Errorf("fillUpSeriesData: unsupported response struct type")
	}

	return nil
}

func fetchQuerySearch(req *http.Request, baseUrl string) ([]core.SearchResult, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var searchData searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchData); err != nil {
		return nil, err
	}
	if !searchData.Success {
		return nil, fmt.Errorf("failed to fecth")
	}
	var realdata []core.SearchResult
	for _, data := range searchData.Documents {
		inst := core.SearchResult{
			Titles: core.Title{
				EnglishTitle: data.English,
				RomajiTitle:  data.Romaji,
				KanjiTitle:   data.Native,
			},
			Type:           data.Type,
			Duration:       data.Duration,
			NumberEpisodes: data.EpCount,
			Url:            fmt.Sprintf("%s/anime/%s", baseUrl, data.ID),
			Year:           data.Year,
		}
		realdata = append(realdata, inst)

	}

	return realdata, nil
}
