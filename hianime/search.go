package hianime

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func Search(query string) ([]SearchElements, error) {
	if strings.Contains(query, " ") {
		words := strings.Fields(query)
		query = ""
		for i := range words {
			if i == len(words)-1 {
				query += words[i]
			} else {
				query += fmt.Sprintf(`%s+`, words[i])
			}
		}
	}

	searchUrl := "https://hianime.to/search?keyword=" + query

	res, err := http.Get(searchUrl)
	if err != nil {
		return []SearchElements{}, fmt.Errorf("Error while fetching search feature: %w", err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return []SearchElements{}, fmt.Errorf("Error while processing to document go query: %w", err)
	}

	results := []SearchElements{}
	doc.Find(".flw-item").Each(func(i int, s *goquery.Selection) {
		linkElement := s.Find(".film-poster")
		if linkElement.Length() == 0 {
			return
		}
		url, exists := linkElement.Find("a").Attr("href")
		if !exists {
			fmt.Println("Couldn't found href.")
		}

		filmDetailElement := s.Find(".film-detail")
		if filmDetailElement == nil {
			return
		}

		h3Element := s.Find("h3")
		englishName := strings.TrimSpace(h3Element.Text())
		japaneseName, _ := h3Element.Find("a").Attr("data-jname")

		typeSeries := s.Find(".fdi-item").First().Text()
		duration := s.Find(".fdi-duration").Text()

		numEps := s.Find(".tick-item.tick-eps")
		numEpsInt, _ := strconv.Atoi(strings.TrimSpace(numEps.Text()))

		results = append(results, SearchElements{
			Url:            BaseUrl + url,
			EnglishName:    englishName,
			JapaneseName:   japaneseName,
			Type:           typeSeries,
			Duration:       duration,
			NumberEpisodes: int16(numEpsInt),
		})
	})
	// searchHtml, err := doc.Html()
	// os.WriteFile("search.html", []byte(searchHtml), 0644)

	return results, nil
}
