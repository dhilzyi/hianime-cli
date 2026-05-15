package anikoto

import (
	"crypto/rand"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/dhilzyi/hianime-cli/internal/core"
)

type series struct {
	core.SeriesData
	animeID int
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

type ajaxResponse struct {
	Status int
	Result string
}
type ajaxServerResponse struct {
	Status int
	Result struct {
		Url      string
		SkipData struct {
			Intro []int
			Outro []int
		} `json:"skip_data"`
	}
}

func toValidHtml(rawHtml string) (string, error) {
	clean, err := strconv.Unquote(`"` + rawHtml + `"`)
	if err != nil {
		// fallback: manual replace
		clean = strings.NewReplacer(
			`\n`, "\n",
			`\t`, "\t",
			`\"`, `"`,
			`\\`, `\`,
		).Replace(rawHtml)
	}
	// if clean == rawHtml {
	// 	return "", fmt.Errorf("failed to clean html")
	// }

	return clean, nil
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
					Name: serverName + fmt.Sprintf("[%s]", typeVid),
					Type: typeVid,
					Key:  key,
				},
				DataLinkId: dataLinkId,
			}
		})
	})

	return servers, nil
}

func parseSearch(body io.ReadCloser) (*core.SearchPage, error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, err
	}

	var searches []core.SearchResult
	listEle := doc.Find("#list-items")
	listEle.Find("div.item").Each(func(i int, s *goquery.Selection) {
		aEle := s.Find("a.name")
		jpTitle, _ := aEle.Attr("data-jp")
		seriesURL, exists := aEle.Attr("href")
		if !exists {
			fmt.Printf("Could not find link series url: %s", aEle.Text())
			return
		}
		typeSeries := s.Find("div.right").Text()
		epsTotalRaw := s.Find("span.total > span").Text()
		epsTotal, err := strconv.Atoi(epsTotalRaw)
		if err != nil {
			epsTotal = 1
		}
		searches = append(searches, core.SearchResult{
			Url:            seriesURL,
			Type:           typeSeries,
			NumberEpisodes: epsTotal,
			Titles: core.Title{
				RomajiTitle:  jpTitle,
				EnglishTitle: aEle.Text(),
			},
		})
	})
	liActiveEle := doc.Find(".page-item.active")
	currPage := liActiveEle.Find("a").Text()
	prevPage := liActiveEle.Prev().Find("a").Text()
	nextPage := liActiveEle.Next().Find("a").Text()

	var hasPrev bool
	var hasNext bool
	currPageInt, err := strconv.Atoi(currPage)
	if err != nil {
		currPageInt = 1
	} else {
		hasPrev, hasNext = checkNextAndPrev(currPageInt, prevPage, nextPage)
	}

	return &core.SearchPage{
		Results: searches,
		HasPrev: hasPrev,
		HasNext: hasNext,
		Page:    currPageInt,
	}, nil
}

func checkNextAndPrev(curInt int, prev, next string) (bool, bool) {
	var hasPrev bool
	var hasNext bool

	prevInt, _ := strconv.Atoi(prev)
	nextInt, _ := strconv.Atoi(next)

	if curInt == (prevInt+1) && prevInt != 0 {
		hasPrev = true
	}
	if curInt == (nextInt-1) && nextInt != 0 {
		hasNext = true
	}

	return hasPrev, hasNext
}
