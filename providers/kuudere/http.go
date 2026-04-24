package kuudere

import (
	"net/http"
	"net/http/cookiejar"
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

func NewSession(defaultDomain string) (*http.Client, error) {
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
