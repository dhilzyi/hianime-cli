package common

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/dhilzyi/hianime-cli/internal/core"
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

func GetPreferredTitle(titles core.Title) string {
	var finalTitle string
	if titles.RomajiTitle != "" {
		finalTitle = titles.RomajiTitle
	} else if titles.EnglishTitle != "" {
		finalTitle = titles.EnglishTitle
	} else if titles.KanjiTitle != "" {
		finalTitle = titles.KanjiTitle

	} else {
		finalTitle = "UNKNOWN"
	}

	return finalTitle
}
