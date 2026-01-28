package hianime

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var BaseUrl string = "https://hianime.to"

// This is where the hianime scrapper logic lives. Check types.go in this same directory to see all the struct types.

func GetSeriesData(series_url string) SeriesData {
	resp, err := http.Get(series_url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// series_html, err := doc.Html()
	// os.WriteFile("a.html", []byte(series_html), 0644)

	header := doc.Find("h2.film-name")

	jname, exists := header.Attr("data-jname")
	if !exists {
		jname, exists = header.Find("a").Attr("data-jname")
	}

	syncData := doc.Find("#syncData")

	rawJson := syncData.Text()
	var data SeriesData
	json.Unmarshal([]byte(rawJson), &data)

	data.EnglishName = strings.TrimSpace(header.Text())
	data.JapaneseName = strings.TrimSpace(jname)

	return data
}

func GetEpisodes(animeId string) []Episodes {
	apiUrl := fmt.Sprintf("%s/ajax/v2/episode/list/%s", BaseUrl, animeId)

	apiResp, err := http.Get(apiUrl)
	if err != nil {
		log.Fatal(err)
	}

	defer apiResp.Body.Close()
	var jsonResp AjaxResponse
	if err := json.NewDecoder(apiResp.Body).Decode(&jsonResp); err != nil {
		fmt.Errorf("Failed to decode JSON: " + err.Error())
	}

	apiDoc, err := goquery.NewDocumentFromReader(strings.NewReader(jsonResp.Html))
	if err != nil {
		log.Fatal(err)
	}

	var episodes []Episodes

	apiDoc.Find("a.ep-item").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			fmt.Println("Couldn't found href.")
		}

		dataId, exists := s.Attr("data-id")
		if !exists {
			fmt.Println("Couldn't found data-id.")
		}

		id_int, err := strconv.Atoi(dataId)
		if err != nil {
			fmt.Print("Failed to convert to integer: " + err.Error())
		}

		titleDiv := s.Find(".ep-name")
		englishTitle := html.UnescapeString(titleDiv.Text())

		japaneseTitle := ""
		rawJName, exists := titleDiv.Attr("data-jname")
		if exists {
			japaneseTitle = html.UnescapeString(rawJName)
		}

		episodeMap := Episodes{
			Number:        i + 1,
			EnglishTitle:  englishTitle,
			JapaneseTitle: japaneseTitle,
			Url:           BaseUrl + html.UnescapeString(href),
			Id:            id_int,
		}
		episodes = append(episodes, episodeMap)
	})

	// api_html, err := apiDoc.Html()
	//
	// os.WriteFile("onepiece.html", []byte(api_html), 0644)

	return episodes
}

func GetEpisodeServerId(episodeId int) []ServerList {
	serverUrl := fmt.Sprintf("%s/ajax/v2/episode/servers?episodeId=%d", BaseUrl, episodeId)

	serverResp, err := http.Get(serverUrl)
	if err != nil {
		fmt.Println("Error while requesting server urls: " + err.Error())
	}
	defer serverResp.Body.Close()

	var serverJson AjaxResponse
	if err := json.NewDecoder(serverResp.Body).Decode(&serverJson); err != nil {
		fmt.Println("Error while converting to json: " + err.Error())
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(serverJson.Html))
	if err != nil {
		fmt.Println("Failed fecthing json to doc: ", err.Error())
	}

	var serverLists []ServerList

	doc.Find(".server-item").Each(func(i int, s *goquery.Selection) {
		dataType, exists := s.Attr("data-type")
		if !exists {
			fmt.Println("Couldn't found 'data-type': " + err.Error())
			return
		}
		dataId, exists := s.Attr("data-id")
		if !exists {
			fmt.Println("Couldn't found data-id: " + err.Error())
			return
		}
		dataIdInt, err := strconv.Atoi(dataId)
		if err != nil {
			fmt.Println("Failed to convert 'dataId' to int: " + err.Error())
			return
		}

		name := s.Find("a").Text()

		// NOTE: Excluding the HD-3 servers for now
		if strings.Contains("HD-3", name) {
			return
		}

		instance := ServerList{
			Type:   dataType,
			Name:   name,
			DataId: dataIdInt,
		}

		serverLists = append(serverLists, instance)
	})

	return serverLists
}

func GetStreamData(serverId int) (StreamData, error) {
	serverUrl := fmt.Sprintf("%s/ajax/v2/episode/sources?id=%d", BaseUrl, serverId)

	resp, err := http.Get(serverUrl)
	if err != nil {
		fmt.Println("Failed to connect with server url: " + err.Error())
	}
	defer resp.Body.Close()

	var respJson MegacloudUrl
	if err := json.NewDecoder(resp.Body).Decode(&respJson); err != nil {
		fmt.Println("Failed to decode JSON: " + err.Error())
	}

	var url string
	if respJson.Type == "iframe" {
		url = respJson.Url
	}

	return ExtractMegacloud(url)
}

func GetNonce(html string) string {
	reStandard := regexp.MustCompile(`\b[a-zA-Z0-9]{48}\b`)
	nonce := reStandard.FindString(html)
	if nonce != "" {
		return nonce
	}

	reSplit := regexp.MustCompile(`x:\s*"(\w+)",\s*y:\s*"(\w+)",\s*z:\s*"(\w+)"`)
	matches := reSplit.FindStringSubmatch(html)

	if len(matches) == 4 {
		return matches[1] + matches[2] + matches[3]
	}

	return ""
}

func ExtractMegacloud(iframeUrl string) (StreamData, error) {
	parsedUrl, err := url.Parse(iframeUrl)
	if err != nil {
		return StreamData{}, fmt.Errorf("Failed to parse url: %w", err)
	}
	defaultDomain := fmt.Sprintf("%s://%s/", parsedUrl.Scheme, parsedUrl.Host)
	userAgent := "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Mobile Safari/537.36"

	req, err := http.NewRequest("GET", iframeUrl, nil)
	if err != nil {
		return StreamData{}, fmt.Errorf("Failed to fecth iframe link: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", defaultDomain)

	client := &http.Client{}

	maxAttempt := 3
	var fileId string
	var nonce string

	for i := range maxAttempt {
		if i >= maxAttempt {
			break
		}

		fmt.Printf("--> Attempt %d/%d to extract...\n", i+1, maxAttempt)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Failed to request with custom headers: " + err.Error())
			continue
		}

		defer resp.Body.Close()

		docMegacloud, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			fmt.Println("Failed to create new document: " + err.Error())
			continue
		}
		megacloudPlayer := docMegacloud.Find("#megacloud-player")
		id, exists := megacloudPlayer.Attr("data-id")
		if !exists {
			fmt.Println("Couldn't found 'fileId'.")
			continue
		} else {
			fileId = id
		}

		singleSelect := docMegacloud.Selection
		outerHtml, _ := goquery.OuterHtml(singleSelect)

		nonce = GetNonce(outerHtml)
		if nonce == "" {
			fmt.Println("Could not find nonce.")
			time.Sleep(1 * time.Second)
			continue
		} else {
			fmt.Println("\n--> Extract success.")
			break
		}
	}

	sourcesUrl := fmt.Sprintf("%sembed-2/v3/e-1/getSources?id=%s&_k=%s", defaultDomain, fileId, nonce)
	sourceReq, err := http.NewRequest("GET", sourcesUrl, nil)
	if err != nil {
		return StreamData{}, fmt.Errorf("Failed when requesting source url: %w", err)
	}

	extractor_headers := map[string]string{
		"Accept":           "*/*",
		"X-Requested-With": "application/json",
		"Referer":          iframeUrl,
		"User-Agent":       userAgent,
	}
	for key, value := range extractor_headers {
		sourceReq.Header.Set(key, value)
	}

	sourceResp, err := client.Do(sourceReq)
	if err != nil {
		return StreamData{}, fmt.Errorf("Failed to fetch source url: %w", err)
	}
	defer sourceResp.Body.Close()

	var sourceJson Sources

	// doc, err := goquery.NewDocumentFromReader(sourceResp.Body)
	// fmt.Println(doc.Text())

	if err := json.NewDecoder(sourceResp.Body).Decode(&sourceJson); err != nil {
		return StreamData{}, fmt.Errorf("Failed to convert to JSON: %w", err)
	}

	streamMap := StreamData{}

	//  NOTE: Still can't play server 'HD-3' (url=douvid.xyz), because it was returning EXT encrypted, and impossible for mpv to play.
	if !sourceJson.Encrypted || strings.Contains(sourceJson.Sources[0].File, ".m3u8") {
		streamMap = StreamData{
			Url:       sourceJson.Sources[0].File,
			UserAgent: userAgent,
			Referer:   defaultDomain,
			Origin:    defaultDomain,
			Tracks:    sourceJson.Tracks,
			Intro:     sourceJson.Intro,
			Outro:     sourceJson.Outro,
		}
	} else {
		return StreamData{}, fmt.Errorf("Files are encrypted. Try other servers.")
	}

	return streamMap, nil
}
