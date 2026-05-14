package anikoto

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/dhilzyi/hianime-cli/internal/core"
)

const (
	baseURL   = "https://anikototv.to"
	userAgent = `Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:150.0) Gecko/20100101 Firefox/150.0`
)

type Anikoto struct {
	episodes map[int]episode
}

type episode struct {
	core.Episode
	id           string
	serverDataId string
}

type server struct {
	core.Server
	DataLinkId string
}

func getEpisodes(anikotoURL string) error {
	return nil
}

func getAnimeID(anikotoURL string) error {
	resp, err := http.Get("GET")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)

	return nil
}

func parseEpisodes(cleanedHtml string) (map[int]episode, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(cleanedHtml))
	if err != nil {
		return nil, err
	}
	episodes := make(map[int]episode)
	doc.Find("li > a").Each(func(i int, s *goquery.Selection) {
		serverDataIds, exists := s.Attr("data-ids")
		if !exists {
			return
		}
		episodeDataId, _ := s.Attr("data-id")
		rawNumEp, exists := s.Attr("data-num")
		if !exists && rawNumEp == "" {
			rawNumEp = s.Find("b").Text()
		}
		numEp, err := strconv.Atoi(rawNumEp)
		if err != nil {
			fmt.Println("Could not convert episode number to integer")
			return
		}
		jpTitle, _ := s.Find("span").Attr("data-jp")
		episodes[numEp] = episode{
			id:           episodeDataId,
			serverDataId: serverDataIds,

			Episode: core.Episode{
				Titles: core.Title{
					RomajiTitle:  jpTitle,
					EnglishTitle: s.Find("span").Text(),
				},
				Number: numEp,
			},
		}
	})

	return episodes, nil
}

func getServers(serverDataIds string) error {
	serverListUrl := fmt.Sprintf("%s/ajax/server/list?servers=%s", baseURL, serverDataIds)
	req, err := http.NewRequest("GET", serverListUrl, nil)
	if err != nil {
		return err
	}
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("Referer", baseURL+"/")
	req.Header.Add("User-Agent", userAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	var ajaxResp ajaxResponse
	if err := json.NewDecoder(resp.Body).Decode(&ajaxResp); err != nil {
		return err
	}
	cleanedHtml, err := toValidHtml(ajaxResp.Result)
	if err != nil {
		return err
	}
	servers, err := parseServers(cleanedHtml)
	if err != nil {
		return err
	}
	fmt.Println(servers)
	return nil
}

func parseServers(cleanHtml string) (map[string]server, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(cleanHtml))
	if err != nil {
		return nil, err
	}
	servers := make(map[string]server)
	doc.Find("div[data-type]").Each(func(i int, s *goquery.Selection) {
		typeVid := strings.TrimSpace(s.Find("label").Text())
		s.Find("li").Each(func(i int, s2 *goquery.Selection) {
			serverName := strings.TrimSpace(s2.Text())
			dataLinkId, exists := s2.Attr("data-link-id")
			if !exists {
				fmt.Printf("Could not find data-link-id on 'li' element for: %s\n", serverName)
				return
			}
			key := serverName + rand.Text()

			servers[key] = server{
				Server: core.Server{
					Name: serverName + fmt.Sprintf("(%s)", typeVid),
					Type: typeVid,
					Key:  key,
				},
				DataLinkId: dataLinkId,
			}
		})
	})

	return servers, nil
}

func getServerUrl(dataLinkId string) error {
	serverGetUrl := fmt.Sprintf("%s/ajax/server?get=%s", baseURL, dataLinkId)
	req, err := http.NewRequest("GET", serverGetUrl, nil)
	if err != nil {
		return err
	}
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("Referer", baseURL+"/")
	req.Header.Add("User-Agent", userAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("bad status code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	var ajaxResp ajaxServerResponse
	if err := json.NewDecoder(resp.Body).Decode(&ajaxResp); err != nil {
		return err
	}

	fmt.Printf("%+v", ajaxResp)

	return nil
}
