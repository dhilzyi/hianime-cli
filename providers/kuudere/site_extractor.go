package kuudere

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dhilzyi/hianime-cli/internal/core"
)

const baseUrl string = "https://kuudere.ru/"

type Kuudere struct {
	inputUrl string
}

type kuudereSeriesDataResponse struct {
	AnimeInfo struct {
		ID           string      `json:"id"`
		English      string      `json:"english"`
		Romaji       string      `json:"romaji"`
		Native       string      `json:"native"`
		AgeRating    string      `json:"ageRating"`
		MalScore     float64     `json:"malScore"`
		AverageScore int         `json:"averageScore"`
		Duration     int         `json:"duration"`
		Genres       []string    `json:"genres"`
		Cover        string      `json:"cover"`
		Banner       string      `json:"banner"`
		Season       string      `json:"season"`
		StartDate    string      `json:"startDate"`
		Status       string      `json:"status"`
		Synonyms     []string    `json:"synonyms"`
		Studios      []string    `json:"studios"`
		Type         string      `json:"type"`
		Year         int         `json:"year"`
		EpCount      int         `json:"epCount"`
		SubbedCount  int         `json:"subbedCount"`
		DubbedCount  int         `json:"dubbedCount"`
		Description  string      `json:"description"`
		Ep           int         `json:"ep"`
		UserLiked    bool        `json:"userLiked"`
		UserUnliked  bool        `json:"userUnliked"`
		Likes        int         `json:"likes"`
		Dislikes     int         `json:"dislikes"`
		InWatchlist  bool        `json:"inWatchlist"`
		Folder       interface{} `json:"folder"`
		Anilist      int         `json:"anilist"`
		URL          string      `json:"url"`
	} `json:"anime_info"`
}

type kuudereEpisodesResponse struct {
	Status string `json:"status"`
	Data   []struct {
		AccessID  string      `json:"access_id"`
		Audio     string      `json:"audio"`
		AnilistID int         `json:"anilist_id"`
		MalID     interface{} `json:"mal_id"`
		Episode   int         `json:"episode"`
		PlayerURL string      `json:"player_url"`
	} `json:"data"`
	Pagination struct {
		CurrentPage int         `json:"current_page"`
		PerPage     int         `json:"per_page"`
		TotalItems  int         `json:"total_items"`
		TotalPages  int         `json:"total_pages"`
		NextPage    interface{} `json:"next_page"`
		PrevPage    interface{} `json:"prev_page"`
	} `json:"pagination"`
}

func New(rawUrl string) *Kuudere {
	return &Kuudere{
		inputUrl: rawUrl,
	}
}

func (k *Kuudere) GetSeriesData() (core.SeriesData, error) {
	return core.SeriesData{}, nil
}

func (k *Kuudere) GetEpisodes() ([]core.Episode, *core.SeriesData, error) {
	return nil, nil, nil
}

func (k *Kuudere) GetServers(selectedEpisode core.Episode) ([]core.Server, error) {
	return nil, nil
}

func (k *Kuudere) GetStreamData(keyServer string) (core.StreamData, error) {
	return core.StreamData{}, nil
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

	var kuudereResp kuudereSeriesDataResponse
	if err := json.NewDecoder(resp.Body).Decode(&kuudereResp); err != nil {
		return core.SeriesData{}, err
	}

	return core.SeriesData{
		AnilistID: kuudereResp.AnimeInfo.Anilist,
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

	req.URL.RawQuery = query.Encode()
	fmt.Println(req.URL.String())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	var episodeResp kuudereEpisodesResponse
	if err := json.NewDecoder(resp.Body).Decode(&episodeResp); err != nil {
		return nil, err
	}

	if episodeResp.Status != "success" {
		return nil, fmt.Errorf("status not success")
	}

	var episodes []core.Episode
	for _, eps := range episodeResp.Data {
		episodes = append(episodes, core.Episode{
			Url: eps.PlayerURL,
		})
	}

	return nil, nil
}
