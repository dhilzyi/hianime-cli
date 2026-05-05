package kuudere

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func fetchRawHtml(inputUrl string, client *http.Client) (string, error) {
	req, err := http.NewRequest("GET", inputUrl, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	htmlRaw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(htmlRaw), nil
}

func extractSeriesUrl(rawUrl, seriesUrl string) (string, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s://%s%s", parsedUrl.Scheme, parsedUrl.Host, seriesUrl), nil
}
