package animenosub

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

func decodeBase64(input string) (string, error) {
	decodeByte, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return "", err
	}

	return string(decodeByte), nil
}

func getIframeSrc(rawhtml string) (string, error) {
	re := regexp.MustCompile(`src\s*=\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(rawhtml)
	if len(matches) > 1 {
		src := matches[1]
		return src, nil
	}

	return "", fmt.Errorf("Error: No src found")
}

func getPageType(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "invalid"
	}

	path := strings.Trim(u.Path, "/")

	if strings.HasPrefix(path, "anime/") {
		return "series"
	}

	if strings.Contains(path, "episode") {
		return "episode"
	}

	return "unknown"
}
