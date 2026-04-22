package animenosub

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type seriesData struct {
	url           string
	episodeNumber int
	episodeTitle  string
}

type server struct {
	Type       string
	ServerName string
	Value      string
}

var baseUrl string = "https://animenosub.to"

func GetSeriesData(url string) ([]seriesData, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	epsLister := doc.Find(".eplister")

	var seriesEpisode []seriesData
	epsLister.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exist := s.Attr("href")
		if !exist {
			fmt.Println("Info: Could not find href in 'a' element")
			return
		}
		epsNumRaw := s.Find(".epl-num")
		epsNumInt, err := strconv.Atoi(epsNumRaw.Text())
		if err != nil {
			fmt.Println("Error: Fail to epsNumRaw convert to int")
			return
		}
		epsTitle := s.Find(".epl-title")

		instance := seriesData{
			url:           href,
			episodeNumber: epsNumInt,
			episodeTitle:  strings.TrimSpace(epsTitle.Text()),
		}
		seriesEpisode = append(seriesEpisode, instance)
	})

	return seriesEpisode, nil
}

func GetSeriesDataFromEpisodeUrl(episodeUrl string) ([]seriesData, error) {

	resp, err := http.Get(episodeUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	epsListElement := doc.Find(".episodelist")

	var series []seriesData
	epsListElement.Find("a").Each(func(i int, s *goquery.Selection) {
		epsUrl, exists := s.Attr("href")
		if !exists {
			fmt.Println("Fail to find href attribute")
			return
		}
		imgElement := s.Find("img")
		epsTitle, exists := imgElement.Attr("title")
		if !exists {
			fmt.Println("Fail to find title attribute in img element")
		}

		series = append(series, seriesData{
			url:          epsUrl,
			episodeTitle: epsTitle,
		})

	})

	fmt.Println(series)

	return nil, nil
}

func GetServerLink(url string) ([]server, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	itemVideoNav := doc.Find(".item.video-nav")

	var serverList []server
	itemVideoNav.Find("option").Each(func(i int, s *goquery.Selection) {
		value, exist := s.Attr("value")
		if !exist {
			fmt.Println("Info: Could not find value in 'option' element")
			return
		}
		if value == "" {
			return
		}

		var typeServer string
		if strings.Contains(s.Text(), "SUB") {
			typeServer = "SUB"
		} else {
			typeServer = "RAW"
		}
		instance := server{
			Type:       typeServer,
			Value:      value,
			ServerName: strings.TrimSpace(s.Text()),
		}
		serverList = append(serverList, instance)
	})

	return serverList, nil
}

func decodeBase64(input string) (string, error) {
	decodeByte, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return "", err
	}

	return string(decodeByte), nil
}

func decodeB64URL(input string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(input)
}

func getIframeSrc(rawhtml string) (string, error) {
	re := regexp.MustCompile(`src\s*=\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(rawhtml)
	if len(matches) > 1 {
		src := matches[1]
		return src, nil
	}

	return "", fmt.Errorf("Error: No src found")
}
