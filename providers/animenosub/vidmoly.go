package animenosub

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
)

func GetStreamLink(vidmolyUrl string) (StreamData, error) {
	resp, err := http.Get(vidmolyUrl)
	if err != nil {
		return StreamData{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return StreamData{}, err
	}

	re := regexp.MustCompile(`file:\s*\'(.*?)\'`)
	result := re.FindStringSubmatch(string(body))
	var streamUrl string
	if len(result) > 1 {
		streamUrl = result[1]
	} else {
		fmt.Println("Coulnd't find the url link")
	}

	header := http.Header{}
	parsedUrl, err := url.Parse(vidmolyUrl)
	if err != nil {
		return StreamData{}, err
	}

	header.Set("Referer", fmt.Sprintf("%s://%s/", parsedUrl.Scheme, parsedUrl.Host))

	streaminstance := StreamData{
		Url:    streamUrl,
		Header: header,
	}

	return streaminstance, nil
}
