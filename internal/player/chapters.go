package player

import (
	"fmt"
	"os"
	"strings"

	"github.com/dhilzyi/hianime-cli/internal/config"
	"github.com/dhilzyi/hianime-cli/internal/core"
)

// NOTE: For building temp file write intro and outro in mpv so user can know the timestamps and skip easily.
func createChapters(data []core.Timestamp, episodeData core.Episode) string {
	f, err := os.CreateTemp("", "hianime_chapters_*.txt")
	if err != nil {
		fmt.Println("Error while creating temporary file: " + err.Error())
		return ""
	}

	contents := ";FFMETADATA1\n"

	writePart := func(start, end int, title string) {
		contents += "[CHAPTER]\n"
		contents += "TIMEBASE=1/1\n"
		contents += fmt.Sprintf("START=%d\n", start)
		contents += fmt.Sprintf("END=%d\n", end)
		contents += fmt.Sprintf("title=%s\n\n", title)
	}

	for _, timestamp := range data {
		if timestamp.Start >= 0 && timestamp.End > timestamp.Start {
			writePart(timestamp.Start, timestamp.End, timestamp.Name)
		}
	}

	f.WriteString(contents)
	return f.Name()
}

func buildProvidedSubs(cfg config.Config, tracks []core.Track) []string {
	var added []string
	englishCandidate := []string{
		"english",
		"eng",
	}

	for _, track := range tracks {
		if track.Type == "thumbnails" || strings.Contains(track.Type, "thumbnail") {
			continue
		}
		if cfg.EnglishOnly {
			isEnglish := false
			langLower := strings.ToLower(track.Language)

			for _, keyword := range englishCandidate {
				if strings.Contains(langLower, keyword) {
					isEnglish = true
					break
				}
			}
			if !isEnglish {
				continue
			}
		}

		added = append(added, "--sub-file="+track.Url)
	}
	return added
}
