package animenosub

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/dhilzyi/hianime-cli/internal/common"
	"github.com/dhilzyi/hianime-cli/internal/core"
	"github.com/dhilzyi/hianime-cli/internal/ui"
)

// original site: 'animenosub.to'

type AnimeNoSub struct {
	serverData  map[string]string
	episodeData map[string]core.Episode
	inputUrl    string
	baseUrl     string
}

type serverData struct {
	Type       string
	ServerName string
	Value      string
}

func New(rawUrl string) *AnimeNoSub {
	return &AnimeNoSub{
		serverData:  make(map[string]string),
		episodeData: make(map[string]core.Episode),
		inputUrl:    rawUrl,
	}
}

func (a *AnimeNoSub) Name() string {
	return "AnimeNoSub"
}

func (a *AnimeNoSub) GetSeriesData() (core.SeriesData, error) {
	var seriesData core.SeriesData
	return seriesData, nil
}

func (a *AnimeNoSub) GetSearchResults(rawInput string) ([]core.SearchResult, error) {
	keywordVal := common.StringToQueryFormat(rawInput)

	url := "https://animenosub.to"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	query := req.URL.Query()
	query.Add("s", keywordVal)
	req.URL.RawQuery = query.Encode()

	searchresults, err := fetchSearchResult(req)
	if err != nil {
		return nil, err
	}

	ui.PrintSearchResults(searchresults, nil)

	return searchresults, nil
}

func (a *AnimeNoSub) GetEpisodes() ([]core.Episode, *core.SeriesData, error) {
	pageType := getPageType(a.inputUrl)

	var epsList []core.Episode
	var seriesData *core.SeriesData
	var err error

	if pageType == "series" {
		epsList, seriesData, err = getEpsListFromSeriesPage(a.inputUrl)
		if err != nil {
			return nil, nil, err
		}
	} else if pageType == "episode" {
		epsList, seriesData, err = getEpsListFromEpisodePage(a.inputUrl)
		if err != nil {
			return nil, nil, err
		}
	}

	for i := range epsList {
		inst := epsList[i]
		a.episodeData[inst.Titles.GetPreferredTitle()] = inst
	}

	return epsList, seriesData, nil
}

func (a *AnimeNoSub) GetServers(selectedEpisode core.Episode) ([]core.Server, error) {
	rawServers, err := getServers(selectedEpisode.Url)
	if err != nil {
		return nil, err
	}

	var rawGroup []core.Server
	var subGroup []core.Server

	for _, server := range rawServers {
		// skipping the Nova server for now
		// since i don't know how to scrape them yet or probably won't
		if strings.Contains(server.ServerName, "Nova") {
			continue
		}
		s := core.Server{
			Name: server.ServerName,
			Type: server.Type,
		}

		a.serverData[server.ServerName] = server.Value

		switch server.Type {
		case "RAW":
			rawGroup = append(rawGroup, s)
		case "SUB":
			subGroup = append(subGroup, s)
		default:
			subGroup = append(subGroup, s)
		}
	}

	serversData := append(rawGroup, subGroup...)

	return serversData, nil
}

func (a *AnimeNoSub) GetStreamData(keyServer string) (core.StreamData, error) {
	value, exists := a.serverData[keyServer]
	if !exists {
		return core.StreamData{}, fmt.Errorf("error: server name is not in the data")
	}
	streamData, err := getStreamDataFromValue(value)
	if err != nil {
		return core.StreamData{}, err
	}

	return streamData, nil
}

func getEpsListFromSeriesPage(seriesPageUrl string) ([]core.Episode, *core.SeriesData, error) {
	resp, err := http.Get(seriesPageUrl)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	epsLister := doc.Find(".eplister")

	var seriesEpisode []core.Episode
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

		instance := core.Episode{
			Url: href,
			Titles: core.Title{
				RomajiTitle: strings.TrimSpace(epsTitle.Text()),
			},
			Number: epsNumInt,
		}
		seriesEpisode = append(seriesEpisode, instance)
	})

	titleSeries := doc.Find("h1.entry-title")
	seriesData := &core.SeriesData{
		SeriesUrl: seriesPageUrl,
		Titles:    core.Title{EnglishTitle: titleSeries.Text()},
	}
	sort.Slice(seriesEpisode, func(i, j int) bool {
		return seriesEpisode[i].Number < seriesEpisode[j].Number
	})

	return seriesEpisode, seriesData, nil
}

func getEpsListFromEpisodePage(episodeUrl string) ([]core.Episode, *core.SeriesData, error) {
	resp, err := http.Get(episodeUrl)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	epsListElement := doc.Find(".episodelist")

	var episodes []core.Episode
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
		spanElement := s.Find(".playinfo span")
		epsNumber, err := getEpsNumber(spanElement.Text())
		if err != nil {
			fmt.Println(err)
		}

		episodes = append(episodes, core.Episode{
			Url: epsUrl,
			Titles: core.Title{
				RomajiTitle: strings.TrimSpace(epsTitle),
			},
			Number: epsNumber,
		})
	})

	h2Element := doc.Find("div.det h2")

	aElement := h2Element.Find("a")
	seriesUrl, exists := aElement.Attr("href")
	if !exists {
		fmt.Println("Info: Failed to find series url")
	}

	seriesData := &core.SeriesData{
		SeriesUrl: seriesUrl,
		Titles:    core.Title{EnglishTitle: h2Element.Text()},
	}

	return episodes, seriesData, nil
}

func getServers(episodeUrl string) ([]serverData, error) {
	resp, err := http.Get(episodeUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	itemVideoNav := doc.Find(".item.video-nav")

	var serverList []serverData
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
		instance := serverData{
			Type:       typeServer,
			Value:      value,
			ServerName: strings.TrimSpace(s.Text()),
		}
		serverList = append(serverList, instance)
	})

	return serverList, nil
}

func getStreamDataFromValue(valueEncrypted string) (core.StreamData, error) {
	rawIframeElement, err := decodeBase64(valueEncrypted)
	if err != nil {
		return core.StreamData{}, err
	}
	iframeUrl, err := getIframeSrc(rawIframeElement)
	if err != nil {
		return core.StreamData{}, err
	}
	var streamdata core.StreamData
	if strings.Contains(iframeUrl, "vidmoly") {
		streamdata, err = getStreamLink(iframeUrl)
		if err != nil {
			return core.StreamData{}, err
		}

	} else {
		embedUrl, err := videosFromUrl(iframeUrl)
		if err != nil {
			return core.StreamData{}, err
		}

		streamdata, err = getEmbedData(embedUrl, iframeUrl)
		if err != nil {
			return core.StreamData{}, err
		}
	}

	return streamdata, nil
}

func fetchSearchResult(req *http.Request) ([]core.SearchResult, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	results, err := getResults(resp)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func getResults(resp *http.Response) ([]core.SearchResult, error) {
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	listUpdEle := doc.Find("div.listupd")

	var data []core.SearchResult
	listUpdEle.Find("a").Each(func(i int, s *goquery.Selection) {
		urlSeries, exists := s.Attr("href")
		if !exists {
			return
		}
		title := strings.TrimSpace(s.Find("div.tt h2").Text())
		typeAnime := s.Find(".typez").Text()

		ins := core.SearchResult{
			Titles: core.Title{EnglishTitle: title},
			Url:    urlSeries,
			Type:   typeAnime,
		}
		data = append(data, ins)
	})

	return data, nil
}
