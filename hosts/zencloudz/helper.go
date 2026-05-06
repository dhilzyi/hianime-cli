package zerocloudz

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func fetchTokenApi(tokenReference, defaultDomain string, client *http.Client) (tokenApiResponse, error) {
	url := fmt.Sprintf("%sapi/m3u8/%s/", defaultDomain, tokenReference)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return tokenApiResponse{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return tokenApiResponse{}, err
	}
	defer resp.Body.Close()

	var tokenResp tokenApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return tokenApiResponse{}, err
	}

	return tokenResp, nil
}

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
