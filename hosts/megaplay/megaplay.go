package megaplay

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"

	"github.com/dhilzyi/hianime-cli/internal/core"
)

const (
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:150.0) Gecko/20100101 Firefox/150.0"
	baseURL   = "https://megaplay.buzz"
)

var reDataId = regexp.MustCompile(`data-id="(.*)"`)

type sourceResponse struct {
	Sources struct {
		File string
	}
	Tracks []Track
	Intro  Timestamp
	Outro  Timestamp
	Server int
}

type Track struct {
	File    string
	Label   string
	Kind    string
	Default bool
}

type Timestamp struct {
	Start int
	End   int
}

func GetStreamData(rawURL string) (core.StreamData, error) {
	id, err := getAnimeId(rawURL, "https://anikototv.to")
	if err != nil {
		return core.StreamData{}, err
	}
	sources, err := fetchFromAPI(id)
	if err != nil {
		return core.StreamData{}, err
	}
	var tracks []core.Track
	for _, track := range sources.Tracks {
		tracks = append(tracks, core.Track{
			Url:      track.File,
			Name:     track.Label,
			Type:     track.Kind,
			Language: track.Label,
		})
	}
	chapters := []core.Timestamp{
		{
			Start: sources.Intro.Start,
			End:   sources.Intro.End,
			Name:  "Intro",
		},
		{
			Start: sources.Outro.Start,
			End:   sources.Outro.End,
			Name:  "Outro",
		},
	}
	streamdata := core.StreamData{
		Url: sources.Sources.File,
		Headers: map[string]string{
			"accept":     "*/*",
			"origin":     baseURL,
			"referer":    baseURL + "/",
			"user-agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:136.0) Gecko/20100101 Firefox/136.0",
		},
		Tracks:   tracks,
		Chapters: chapters,
	}
	return streamdata, nil
}

func getAnimeId(rawURL, referer string) (int, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Add("Referer", referer)
	req.Header.Add("User-Agent", userAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode >= 400 {
		return 0, fmt.Errorf("bad status fetch url '%s': %d", req.URL, resp.StatusCode)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	id, err := reAnimeId(string(body))
	if err != nil {
		return 0, err
	}

	return id, nil
}

func reAnimeId(rawHTML string) (int, error) {
	match := reDataId.FindStringSubmatch(rawHTML)
	if len(match) < 2 {
		return 0, fmt.Errorf("could not find match for data-id")
	}
	fmt.Println(match)

	id, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, err
	}

	return id, nil
}

func fetchFromAPI(id int) (sourceResponse, error) {
	apiURL := fmt.Sprintf("%s/stream/getSources?id=%d&id=%d", baseURL, id, id)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return sourceResponse{}, err
	}
	req.Header.Add("Referer", baseURL+"/")
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return sourceResponse{}, err
	}
	if resp.StatusCode >= 400 {
		return sourceResponse{}, fmt.Errorf("bad status fetch url '%s': %d", req.URL, resp.StatusCode)
	}
	defer resp.Body.Close()

	var sources sourceResponse
	if err := json.NewDecoder(resp.Body).Decode(&sources); err != nil {
		return sourceResponse{}, err
	}

	return sources, nil
}
