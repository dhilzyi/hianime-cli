package animenosub

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"

	"github.com/dhilzyi/hianime-cli/internal/core"
)

func getStreamLink(vidmolyUrl string) (core.StreamData, error) {
	resp, err := http.Get(vidmolyUrl)
	if err != nil {
		return core.StreamData{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return core.StreamData{}, err
	}

	re := regexp.MustCompile(`file:\s*\'(.*?)\'`)
	result := re.FindStringSubmatch(string(body))
	var streamUrl string
	if len(result) > 1 {
		streamUrl = result[1]
	} else {
		return core.StreamData{}, fmt.Errorf("could not find the stream url link")

	}

	header := make(map[string]string)
	parsedUrl, err := url.Parse(vidmolyUrl)
	if err != nil {
		return core.StreamData{}, err
	}

	header["Referer"] = fmt.Sprintf("%s://%s/", parsedUrl.Scheme, parsedUrl.Host)

	streaminstance := core.StreamData{
		Url:     streamUrl,
		Headers: header,
	}

	return streaminstance, nil
}
