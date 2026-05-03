package common

import (
	"fmt"
	"net/url"
	"strings"
)

func StringToQueryFormat(rawInput string) string {
	rawSplitted := strings.Split(rawInput, " ")
	if len(rawSplitted) == 0 {
		return rawInput
	}

	return strings.Join(rawSplitted, "+")
}

func GetBaseURL(rawUrl string) (string, error) {
	parsedURL, err := url.Parse(rawUrl)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host), nil
}

func TruncatedString(raw string, maxChar int) string {
	if len(raw) <= maxChar {
		return raw
	}

	return raw[:maxChar] + "..."
}
