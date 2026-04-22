package animenosub

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
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

func getEpsNumber(input string) (int, error) {
	re := regexp.MustCompile(`\s*Eps\s*(\d)`)
	matches := re.FindStringSubmatch(input)
	if len(matches) < 2 {
		return 0, fmt.Errorf("Error: No match for episode number regexp")
	}
	numberRaw := matches[1]
	numberInt, err := strconv.Atoi(numberRaw)
	if err != nil {
		return 0, err
	}

	return numberInt, nil
}
