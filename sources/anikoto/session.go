package anikoto

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/dhilzyi/hianime-cli/internal/core"
)

type session struct {
	http *http.Client
}

func (s session) getEpisodes(animeId int) (map[int]episode, error) {
	episodesRawHtml, err := s.fetchEpisodes(animeId)
	if err != nil {
		return nil, err
	}
	episodesHtml, err := toValidHtml(episodesRawHtml)
	if err != nil {
		return nil, err
	}
	episodes, err := parseEpisodes(episodesHtml)
	if err != nil {
		return nil, err
	}
	return episodes, nil
}

func (s session) fetchEpisodes(animeId int) (string, error) {
	ajaxEpsUrl := fmt.Sprintf("%s/ajax/episode/list/%d?vrf=", baseURL, animeId)
	req, err := http.NewRequest("GET", ajaxEpsUrl, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("Referer", baseURL+"/")
	req.Header.Add("User-Agent", userAgent)

	resp, err := s.http.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("bad status code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	var ajaxResp ajaxResponse
	if err := json.NewDecoder(resp.Body).Decode(&ajaxResp); err != nil {
		return "", err
	}
	if ajaxResp.Status != 200 {
		return "", fmt.Errorf("bad ajax status: %d", ajaxResp.Status)
	}

	return ajaxResp.Result, nil
}

func (s session) getSeriesData(anikotoURL string) (series, error) {
	req, err := http.NewRequest("GET", anikotoURL, nil)
	if err != nil {
		return series{}, err
	}
	req.Header.Add("Referer", baseURL+"/")
	req.Header.Add("User-Agent", userAgent)

	resp, err := s.http.Do(req)
	if err != nil {
		return series{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return series{}, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return series{}, err
	}
	watchMainEl := doc.Find("#watch-main")
	animeIdRaw, exists := watchMainEl.Attr("data-id")
	if !exists && animeIdRaw == "" {
		return series{}, fmt.Errorf("could not find animeid for url: %s", resp.Request.URL)
	}
	animeId, err := strconv.Atoi(animeIdRaw)
	if err != nil {
		return series{}, err
	}
	seriesUrl, exists := watchMainEl.Attr("data-url")
	if !exists {
		seriesUrl = anikotoURL
	}
	h1Ele := doc.Find("h1.title")
	engTitle := h1Ele.Text()
	romajiTitle, _ := h1Ele.Attr("data-jp")

	return series{
		SeriesData: core.SeriesData{
			Titles: core.Title{
				EnglishTitle: engTitle,
				RomajiTitle:  romajiTitle,
			},
			SeriesUrl: seriesUrl,
		},
		animeID: animeId,
	}, nil
}

func (s session) getServers(serverDataIds string) (map[string]server, error) {
	serverListUrl := fmt.Sprintf("%s/ajax/server/list?servers=%s", baseURL, serverDataIds)
	req, err := http.NewRequest("GET", serverListUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("Referer", baseURL+"/")
	req.Header.Add("User-Agent", userAgent)

	resp, err := s.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	var ajaxResp ajaxResponse
	if err := json.NewDecoder(resp.Body).Decode(&ajaxResp); err != nil {
		return nil, err
	}
	cleanedHtml, err := toValidHtml(ajaxResp.Result)
	if err != nil {
		return nil, err
	}
	servers, err := parseServers(cleanedHtml)
	if err != nil {
		return nil, err
	}

	return servers, nil
}

func (s session) getServerUrl(dataLinkId string) (string, error) {
	serverGetUrl := fmt.Sprintf("%s/ajax/server?get=%s", baseURL, dataLinkId)
	req, err := http.NewRequest("GET", serverGetUrl, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("Referer", baseURL+"/")
	req.Header.Add("User-Agent", userAgent)

	resp, err := s.http.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("bad status code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	var ajaxResp ajaxServerResponse
	if err := json.NewDecoder(resp.Body).Decode(&ajaxResp); err != nil {
		return "", err
	}

	return ajaxResp.Result.Url, nil
}
