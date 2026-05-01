package kuudere

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

// defaultHeadersTransport acts as middleware to inject headers into every request
type defaultHeadersTransport struct {
	baseTransport http.RoundTripper
	headers       map[string]string
}

// RoundTrip intercepts the request before it is sent
func (t *defaultHeadersTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqClone := req.Clone(req.Context())

	for key, value := range t.headers {
		if reqClone.Header.Get(key) == "" {
			reqClone.Header.Set(key, value)
		}
	}

	return t.baseTransport.RoundTrip(reqClone)
}

func newSession(defaultDomain string) (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	defaultHeaders := map[string]string{
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Accept-Language": "en-US,en;q=0.5",
		"Referer":         defaultDomain,
	}

	transport := &defaultHeadersTransport{
		baseTransport: http.DefaultTransport,
		headers:       defaultHeaders,
	}

	return &http.Client{
		Jar:       jar,
		Transport: transport,
		Timeout:   15 * time.Second,
	}, nil
}

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

func extractSeriesUrl(rawUrl, seriesUrl string) (string, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s://%s%s", parsedUrl.Scheme, parsedUrl.Host, seriesUrl), nil
}
